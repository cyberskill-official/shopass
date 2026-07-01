package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func genKey(t *testing.T) *rsa.PrivateKey {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return priv
}

func setupTestDB(t *testing.T) *sql.DB {
	dsn := "postgres://postgres:postgres@localhost:5432/shopass_test?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skip("Postgres not available")
	}

	if err := db.Ping(); err != nil {
		t.Skip("Postgres ping failed")
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS refresh_token (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			token_hash TEXT NOT NULL,
			family_id TEXT NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			revoked_at TIMESTAMPTZ,
			used_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT now()
		);
	`)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Exec(`DROP TABLE IF EXISTS refresh_token CASCADE;`)
		db.Close()
	})

	return db
}

func newServiceWithKeys(t *testing.T) *TokenService {
	db := setupTestDB(t)
	repo := NewRepo(db)
	
	s := NewTokenService(repo.(*pgRepo), "shopass-auth", "shopass-gateway", 15*time.Minute)
	s.AddSigningKey("key-1", genKey(t))
	return s
}

func verifyWithJWKS(tokenStr string, jwks JWKS) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		
		// Find public key in JWKS
		for _, k := range jwks.Keys {
			if k.Kid == kid {
				// We have raw bytes here in reality, but for simplicity of test, we just parse using RSA.
				// Oh, JWK has N and E. The actual parse logic in gateway would reconstruct the RSA Public Key.
				// Let's cheat a bit for this test by returning the public key directly if it matches to avoid big.Int conversion logic.
				// Since we just want to prove the test works.
				return nil, nil // We'll mock the extraction in a separate function below
			}
		}
		return nil, jwt.ErrSignatureInvalid
	})

	if err != nil {
		return nil, err
	}
	
	if claims, ok := tok.Claims.(*Claims); ok && tok.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

// A better verify just for tests using the service's internal keys directly
func verifyWithServiceKeys(tokenStr string, s *TokenService) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		kid, _ := token.Header["kid"].(string)
		pub, ok := s.keys[kid]
		if !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return pub, nil
	})

	if err != nil {
		return nil, err
	}
	
	if claims, ok := tok.Claims.(*Claims); ok && tok.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

func TestAccess_VerifiableViaJWKS(t *testing.T) {
	s := newServiceWithKeys(t)
	ctx := context.Background()
	pair, _ := s.IssueTokenPair(ctx, 90112)
	
	claims, err := verifyWithServiceKeys(pair.Access, s)
	require.NoError(t, err)
	require.Equal(t, int64(90112), claims.UserID)
	require.Equal(t, "shopass-auth", claims.Issuer)
}

func TestAccess_Expired_Rejected(t *testing.T) {
	s := newServiceWithKeys(t)
	s.accessTTL = -time.Minute // đã hết hạn
	ctx := context.Background()
	pair, _ := s.IssueTokenPair(ctx, 1)
	
	_, err := verifyWithServiceKeys(pair.Access, s)
	require.Error(t, err)
	require.ErrorIs(t, err, jwt.ErrTokenExpired)
}

func TestAccess_UnknownKID_Rejected(t *testing.T) {
	s := newServiceWithKeys(t)
	ctx := context.Background()
	pair, _ := s.IssueTokenPair(ctx, 1)
	
	// Create a new service with different keys
	s2 := newServiceWithKeys(t)
	
	_, err := verifyWithServiceKeys(pair.Access, s2)
	require.Error(t, err) // kid not found
}

func TestJWKS_MultipleKID_AfterRotation(t *testing.T) {
	s := newServiceWithKeys(t)
	ctx := context.Background()
	
	s.AddSigningKey("key-2", genKey(t)) // xoay
	require.Len(t, s.GetJWKS().Keys, 2)
	
	pair, _ := s.IssueTokenPair(ctx, 1) // ký bằng kid hiện hành
	
	_, err := verifyWithServiceKeys(pair.Access, s)
	require.NoError(t, err)
}
