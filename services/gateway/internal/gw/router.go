package gw

import (
	"net/http"
)

type Deps struct {
	WAFConfig WAFConfig
	Redis     RedisClient
	JWKS      JWKSCache
}

func NewHandler(deps Deps) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/v1/", upstreamREST(deps))
	mux.Handle("/graphql", upstreamGraphQL(deps))
	mux.Handle("/ws", wsUpgrade(deps))

	return chain(
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

func upstreamREST(deps Deps) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Upstream", "rest")
		w.Header().Set("X-User-Id-Echo", r.Header.Get("X-User-Id"))
		w.Header().Set("X-Request-Id-Echo", r.Header.Get("X-Request-Id"))
		w.WriteHeader(http.StatusOK)
	})
}

func upstreamGraphQL(deps Deps) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Upstream", "graphql")
		w.WriteHeader(http.StatusOK)
	})
}

func wsUpgrade(deps Deps) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Upstream", "ws")
		w.WriteHeader(http.StatusOK)
	})
}
