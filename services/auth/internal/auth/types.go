package auth

import "errors"

var (
	ErrNoIdentifier = errors.New("must provide at least one identifier (email or phone)")
	ErrEmailTaken   = errors.New("email already taken")
	ErrWeakPassword = errors.New("password too weak (minimum 8 characters)")
)

type AppUser struct {
	ID             int64
	Email          string
	Phone          string
	Locale         string
	Status         string
	PwdHash        string
	ReferralCodeID *int64
	Tier           string
}

type RegisterInput struct {
	Email    string
	Phone    string
	Password string
}

type TokenPair struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
	Type    string `json:"token_type"`
	Expires int    `json:"expires_in"` // seconds
}

var (
	ErrInvalidRefresh       = errors.New("invalid refresh token")
	ErrRefreshReuseDetected = errors.New("refresh token reuse detected")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrAccountNotActive     = errors.New("account is not active")
	ErrInvalidExtRef        = errors.New("invalid external user reference")
	ErrExtRefNotAnonymized  = errors.New("external user reference must be anonymized")
)

type PlatformAccount struct {
	ID         int64
	UserID     int64
	PlatformID int16
	ExtUserRef string
}
