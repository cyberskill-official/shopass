package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

type Repo interface {
	InsertUser(ctx context.Context, u AppUser) (int64, error)
	FindByEmail(ctx context.Context, email string) (AppUser, error)
	FindRefreshByHash(ctx context.Context, hash string) (RefreshTokenRow, error)
	RevokeFamily(ctx context.Context, familyID string) error
	MarkUsed(ctx context.Context, id int64) error
	InsertRefreshToken(ctx context.Context, userID int64, hash, familyID string, expiresAt time.Time) error
}

type pgRepo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) Repo {
	return &pgRepo{db: db}
}

func (r *pgRepo) InsertUser(ctx context.Context, u AppUser) (int64, error) {
	query := `
		INSERT INTO app_user (email, phone, locale, status, pwd_hash)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	var id int64
	err := r.db.QueryRowContext(ctx, query, u.Email, u.Phone, u.Locale, u.Status, u.PwdHash).Scan(&id)
	return id, err
}

func (r *pgRepo) FindByEmail(ctx context.Context, email string) (AppUser, error) {
	query := `
		SELECT id, email, phone, locale, status, pwd_hash, referral_code_id
		FROM app_user
		WHERE email = $1
	`
	var u AppUser
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.Phone, &u.Locale, &u.Status, &u.PwdHash, &u.ReferralCodeID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return u, nil // Not found
	}
	return u, err
}

// Refresh token methods
type RefreshTokenRow struct {
	ID        int64
	UserID    int64
	TokenHash string
	FamilyID  string // Using string to map from UUID in postgres
	ExpiresAt time.Time
	RevokedAt *time.Time
	UsedAt    *time.Time
}

func (r *pgRepo) FindRefreshByHash(ctx context.Context, hash string) (RefreshTokenRow, error) {
	var row RefreshTokenRow
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token_hash, family_id, expires_at, revoked_at, used_at
		FROM refresh_token WHERE token_hash = $1
	`, hash).Scan(&row.ID, &row.UserID, &row.TokenHash, &row.FamilyID, &row.ExpiresAt, &row.RevokedAt, &row.UsedAt)
	return row, err
}

func (r *pgRepo) RevokeFamily(ctx context.Context, familyID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_token SET revoked_at = now() WHERE family_id = $1`, familyID)
	return err
}

func (r *pgRepo) MarkUsed(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_token SET used_at = now() WHERE id = $1`, id)
	return err
}

func (r *pgRepo) InsertRefreshToken(ctx context.Context, userID int64, hash, familyID string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO refresh_token (user_id, token_hash, family_id, expires_at)
		VALUES ($1, $2, $3, $4)
	`, userID, hash, familyID, expiresAt)
	return err
}

func isUniqueViolation(err error, constraint string) bool {
	if err == nil {
		return false
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if pqErr.Code == "23505" && pqErr.Constraint == constraint {
			return true
		}
	}
	return false
}
