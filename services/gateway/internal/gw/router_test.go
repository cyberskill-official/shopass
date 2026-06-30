package gw

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWAF_PathTraversal_400(t *testing.T) {
	deps := Deps{JWKS: &mockJWKS{}}
	h := NewHandler(deps)

	req := httptest.NewRequest("GET", "/v1/../etc/passwd", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Equal(t, 400, rr.Code)
}

func TestWAF_BodySize_413(t *testing.T) {
	deps := Deps{JWKS: &mockJWKS{}, WAFConfig: WAFConfig{MaxBodySize: 10}} // very small cap
	h := NewHandler(deps)

	req := httptest.NewRequest("POST", "/v1/auth/login", strings.NewReader("this is a very long body that exceeds 10 bytes"))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Equal(t, 413, rr.Code)
}

func TestRequestID_Generated(t *testing.T) {
	deps := Deps{JWKS: &mockJWKS{}}
	h := NewHandler(deps)

	req := httptest.NewRequest("GET", "/v1/health", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	require.Equal(t, 200, rr.Code)
	require.NotEmpty(t, rr.Header().Get("X-Request-Id-Echo"))
	require.NotEmpty(t, rr.Header().Get("X-Request-Id"))
}

func TestRouter_Routing(t *testing.T) {
	deps := Deps{JWKS: &mockJWKS{}}
	h := NewHandler(deps)

	tests := []struct {
		path     string
		upstream string
	}{
		{"/v1/health", "rest"},
		{"/graphql", "graphql"},
		{"/ws", "ws"},
	}

	for _, tc := range tests {
		req := httptest.NewRequest("GET", tc.path, nil)
		// For ws and graphql, if we want them to pass they need to be public or authenticated.
		// WSS spec says "verify JWT in handshake (query param or subprotocol)". For test simplicity, let's just use valid token.
		req.Header.Set("Authorization", "Bearer valid-1")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		require.Equal(t, 200, rr.Code)
		require.Equal(t, tc.upstream, rr.Header().Get("X-Upstream"))
	}
}
