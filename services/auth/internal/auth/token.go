package auth

import (
	"crypto/rsa"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64  `json:"user_id"`
	Locale string `json:"locale"`
	Tier   string `json:"tier"`
	jwt.RegisteredClaims
}

type TokenService struct {
	repo       Repo
	iss        string
	aud        string
	accessTTL  time.Duration
	currentKID string
	signingKey *rsa.PrivateKey
	keys       map[string]*rsa.PublicKey
}

func NewTokenService(repo Repo, iss, aud string, ttl time.Duration) *TokenService {
	return &TokenService{
		repo:      repo,
		iss:       iss,
		aud:       aud,
		accessTTL: ttl,
		keys:      make(map[string]*rsa.PublicKey),
	}
}

func (s *TokenService) AddSigningKey(kid string, priv *rsa.PrivateKey) {
	s.currentKID = kid
	s.signingKey = priv
	s.keys[kid] = &priv.PublicKey
}

func (s *TokenService) IssueAccess(u AppUser) (string, error) {
	now := time.Now()

	tier := u.Tier
	if tier == "" {
		tier = "free"
	}

	c := Claims{
		UserID: u.ID,
		Locale: u.Locale,
		Tier:   tier,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.iss,
			Audience:  jwt.ClaimStrings{s.aud},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, c)
	tok.Header["kid"] = s.currentKID
	return tok.SignedString(s.signingKey)
}
