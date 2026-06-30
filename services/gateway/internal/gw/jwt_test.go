package gw

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

type mockJWKS struct {
	err error
}

func (m *mockJWKS) Verify(ctx context.Context, tokenString string) (*Claims, error) {
	if m.err != nil {
		return nil, m.err
	}
	if tokenString == "expired" {
		return nil, errors.New("expired")
	}
	if tokenString == "bad-aud" {
		return nil, errors.New("bad aud")
	}
	if tokenString == "valid-90112" {
		return &Claims{
			UserID: 90112,
			Locale: "vi",
			Tier:   "free",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				Issuer:    "sandeal.auth",
				Audience:  jwt.ClaimStrings{"sandeal.api"},
			},
		}, nil
	}
	if tokenString == "valid-1" {
		return &Claims{UserID: 1}, nil
	}
	if tokenString == "valid-2" {
		return &Claims{UserID: 2}, nil
	}
	return nil, errors.New("invalid signature")
}

func TestJWT_Expired_401(t *testing.T) {
	deps := Deps{JWKS: &mockJWKS{}}
	h := NewHandler(deps)

	req := httptest.NewRequest("GET", "/v1/track", nil)
	req.Header.Set("Authorization", "Bearer expired")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	require.Equal(t, 401, rr.Code)
}

func TestJWT_BadAudience_401(t *testing.T) {
	deps := Deps{JWKS: &mockJWKS{}}
	h := NewHandler(deps)

	req := httptest.NewRequest("GET", "/v1/track", nil)
	req.Header.Set("Authorization", "Bearer bad-aud")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	require.Equal(t, 401, rr.Code)
}

func TestJWT_Valid_PropagatesUserID(t *testing.T) {
	deps := Deps{JWKS: &mockJWKS{}}
	h := NewHandler(deps)

	req := httptest.NewRequest("GET", "/v1/track", nil)
	req.Header.Set("Authorization", "Bearer valid-90112")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	require.Equal(t, 200, rr.Code)
	require.Equal(t, "90112", rr.Header().Get("X-User-Id-Echo"))
}

func TestJWT_PublicRoute_PassesWithoutToken(t *testing.T) {
	deps := Deps{JWKS: &mockJWKS{}}
	h := NewHandler(deps)

	req := httptest.NewRequest("GET", "/v1/health", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	require.Equal(t, 200, rr.Code)
}
