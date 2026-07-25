// Command authsvc serves the auth HTTP endpoints: password register/login/refresh
// (FR-AUTH-001/002), JWKS for the gateway, and social login (FR-AUTH-004). It
// sits behind the API gateway. Social login is enabled only when a provider
// client id is configured; the client secret comes from the environment / secret
// manager, never a code literal (FR-AUTH-004 §1 #9).
package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"shopass/obs"
	"shopass/services/auth/internal/auth"
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	log := slog.Default()
	dbURL := env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/shopass?sslmode=disable")
	addr := env("AUTH_ADDR", ":8084")
	iss := env("AUTH_ISS", "shopass-auth")
	aud := env("AUTH_AUD", "shopass-gateway")
	appEnv := env("APP_ENV", "development")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Error("db open", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	repo := auth.NewRepo(db)

	// Development may generate an ephemeral key. Production must mount a
	// durable secret so access tokens survive a process restart.
	priv, err := loadSigningKey(appEnv)
	if err != nil {
		log.Error("load signing key", "err", err)
		os.Exit(1)
	}
	tokens := auth.NewTokenService(repo, iss, aud, 15*time.Minute)
	tokens.AddSigningKey(env("AUTH_KEY_ID", "auth-key-1"), priv)

	regsvc := auth.NewService(repo, auth.Argon2Params{
		Time: 3, Memory: 64 * 1024, Parallelism: 2, SaltLen: 16, KeyLen: 32,
	})

	// OAuth callback currently returns a token pair directly. Keep it disabled
	// in production until the web BFF owns the callback and can store refresh
	// tokens in the HttpOnly host-only cookie instead of exposing one to the
	// browser response.
	enableGoogleOAuth := strings.EqualFold(os.Getenv("ENABLE_GOOGLE_OAUTH"), "true")
	if appEnv == "production" && enableGoogleOAuth {
		log.Error("Google OAuth cannot be enabled in production until the BFF callback flow is implemented")
		os.Exit(1)
	}

	providers := map[string]auth.OAuthProvider{}
	if cid := os.Getenv("GOOGLE_CLIENT_ID"); enableGoogleOAuth && cid != "" {
		keys := auth.NewHTTPKeySet(env("GOOGLE_JWKS_URL", "https://www.googleapis.com/oauth2/v3/certs"))
		providers["google"] = auth.NewGoogleProvider(
			cid,
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			os.Getenv("GOOGLE_REDIRECT_URI"),
			keys,
		)
	}
	oauthsvc := auth.NewOAuthService(providers, auth.NewMemTmpStore(), repo.(auth.SocialRepo), tokens)

	h := &handlers{log: log, tokens: tokens, reg: regsvc, oauth: oauthsvc, socialEnabled: len(providers) > 0}
	mux := http.NewServeMux()
	h.routes(mux)

	server := &http.Server{
		Addr:              addr,
		Handler:           obs.HTTP("authsvc")(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Info("authsvc listening", "addr", addr, "social", len(providers) > 0)
	if err := server.ListenAndServe(); err != nil {
		log.Error("serve", "err", err)
		os.Exit(1)
	}
}
