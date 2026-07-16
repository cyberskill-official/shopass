// Command authsvc serves the auth HTTP endpoints: password register/login/refresh
// (TASK-AUTH-001/002), JWKS for the gateway, and social login (TASK-AUTH-004). It
// sits behind the API gateway. Social login is enabled only when a provider
// client id is configured; the client secret comes from the environment / secret
// manager, never a code literal (TASK-AUTH-004 §1 #9).
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

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

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Error("db open", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	repo := auth.NewRepo(db)

	// Access-token signing key. In production, load from the secrets manager
	// (TASK-INFRA-003); for a standalone run, generate one at boot.
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Error("keygen", "err", err)
		os.Exit(1)
	}
	tokens := auth.NewTokenService(repo, iss, aud, 15*time.Minute)
	tokens.AddSigningKey("auth-key-1", priv)

	regsvc := auth.NewService(repo, auth.Argon2Params{
		Time: 1, Memory: 64 * 1024, Parallelism: 2, SaltLen: 16, KeyLen: 32,
	})

	// Social login is wired only when a Google client id is present.
	providers := map[string]auth.OAuthProvider{}
	if cid := os.Getenv("GOOGLE_CLIENT_ID"); cid != "" {
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

	log.Info("authsvc listening", "addr", addr, "social", len(providers) > 0)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Error("serve", "err", err)
		os.Exit(1)
	}
}
