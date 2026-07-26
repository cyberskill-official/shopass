package gw

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Deps struct {
	WAFConfig WAFConfig
	Redis     RedisClient
	JWKS      JWKSCache
	Upstreams Upstreams
}

// Upstreams are private Compose-network URLs. The gateway is the only place
// where caller identity is verified and translated into trusted headers.
type Upstreams struct {
	Auth  string
	Track string
	Price string
	Deal  string
	Notif string
	Bill  string
	BFF   string

	// Handler fields are deliberately injectable for unit tests. Production
	// leaves them nil and uses the private URLs above.
	AuthHandler  http.Handler
	TrackHandler http.Handler
	PriceHandler http.Handler
	DealHandler  http.Handler
	NotifHandler http.Handler
	BillHandler  http.Handler
	BFFHandler   http.Handler
}

func NewHandler(deps Deps) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})
	mux.Handle("/v1/", upstreamREST(deps.Upstreams))
	mux.Handle("/graphql", handlerOrProxy(deps.Upstreams.BFFHandler, deps.Upstreams.BFF))

	return chain(
		sanitizeClientIdentity(),
		requestID(),
		waf(deps.WAFConfig),
		jwtVerify(deps.JWKS),
		rateLimit(deps.Redis),
	)(mux)
}

func chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		h := final
		for i := len(middlewares) - 1; i >= 0; i-- {
			h = middlewares[i](h)
		}
		return h
	}
}

func upstreamREST(upstreams Upstreams) http.Handler {
	auth := handlerOrProxy(upstreams.AuthHandler, upstreams.Auth)
	track := handlerOrProxy(upstreams.TrackHandler, upstreams.Track)
	deal := handlerOrProxy(upstreams.DealHandler, upstreams.Deal)
	notif := handlerOrProxy(upstreams.NotifHandler, upstreams.Notif)
	bill := handlerOrProxy(upstreams.BillHandler, upstreams.Bill)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.HasPrefix(path, "/v1/auth/"):
			auth.ServeHTTP(w, r)
		case path == "/v1/track", path == "/v1/tracked-products":
			track.ServeHTTP(w, r)
		case strings.HasPrefix(path, "/v1/products/") && strings.HasSuffix(path, "/browser-snapshot"):
			track.ServeHTTP(w, r)
		case path == "/v1/alerts", strings.HasPrefix(path, "/v1/alerts/"):
			track.ServeHTTP(w, r)
		case path == "/v1/devices":
			notif.ServeHTTP(w, r)
		case strings.HasPrefix(path, "/v1/products/") && strings.HasSuffix(path, "/chart"):
			deal.ServeHTTP(w, r)
		case path == "/v1/billing/checkout", strings.HasPrefix(path, "/v1/billing/ipn/"):
			bill.ServeHTTP(w, r)
		default:
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		}
	})
}

func handlerOrProxy(handler http.Handler, rawURL string) http.Handler {
	if handler != nil {
		return handler
	}
	return proxyOrUnavailable(rawURL)
}

func proxyOrUnavailable(rawURL string) http.Handler {
	target, err := url.Parse(rawURL)
	if err != nil || target.Scheme == "" || target.Host == "" {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "upstream unavailable"})
		})
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	original := proxy.Director
	proxy.Director = func(r *http.Request) {
		original(r)
		// These values can only have been set by jwtVerify after the inbound
		// headers were stripped by sanitizeClientIdentity.
		r.Header.Del("X-Forwarded-User")
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, _ error) {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "upstream unavailable"})
	}
	return proxy
}

func sanitizeClientIdentity() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, header := range []string{
				"X-User-Id",
				"X-User-Locale",
				"X-User-Tier",
				"X-Forwarded-User",
			} {
				r.Header.Del(header)
			}
			next.ServeHTTP(w, r)
		})
	}
}
