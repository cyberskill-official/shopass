package gw

import (
	"net/http"
)

type Deps struct {
	WAFConfig WAFConfig
	Redis     RedisClient
	JWKS      *JWKSCache
	REST      http.Handler
	GraphQL   http.Handler
	WS        http.Handler
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

func upstreamREST(deps Deps) http.Handler {
	if deps.REST != nil {
		return deps.REST
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func upstreamGraphQL(deps Deps) http.Handler {
	if deps.GraphQL != nil {
		return deps.GraphQL
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func wsUpgrade(deps Deps) http.Handler {
	if deps.WS != nil {
		return deps.WS
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Upgrade") != "websocket" {
			writeJSON(w, http.StatusUpgradeRequired, errBody("upgrade_required"))
			return
		}
		w.WriteHeader(http.StatusSwitchingProtocols)
	})
}

type Middleware func(http.Handler) http.Handler

func chain(mws ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(mws) - 1; i >= 0; i-- {
			next = mws[i](next)
		}
		return next
	}
}
