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
	RotateRefreshToken(ctx context.Context, oldHash, replacementHash string, replacementExpiresAt time.Time) (RefreshRotationStatus, error)

	UpsertPlatformAccount(ctx context.Context, pa PlatformAccount) error
	ListPlatformAccountsByUser(ctx context.Context, userID int64) ([]PlatformAccount, error)
	DeletePlatformAccount(ctx context.Context, userID int64, platformID int16) error

	// Lifecycle & Reset
	FindByIdentifier(ctx context.Context, identifier string) (AppUser, bool)
	SaveReset(ctx context.Context, userID int64, hash string, expiresAt time.Time) error
	FindResetByHash(ctx context.Context, hash string) (PasswordResetRow, bool)
	UpdatePassword(ctx context.Context, userID int64, newHash string) error
	MarkResetUsed(ctx context.Context, id int64) error
	RevokeAllRefresh(ctx context.Context, userID int64) error
	SetStatus(ctx context.Context, userID int64, status string) error
	AnonymizePII(ctx context.Context, userID int64) error
	DeletePlatformAccounts(ctx context.Context, userID int64) error
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
	// COALESCE phone/pwd_hash so a social-only account (TASK-AUTH-004: pwd_hash NULL,
	// no phone yet) reads back cleanly into the string fields of AppUser.
	query := `
		SELECT id, email, COALESCE(phone, ''), locale, status, COALESCE(pwd_hash, ''), referral_code_id
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

// RefreshRotationStatus is the authoritative outcome of the transactional
// compare-and-set performed by RotateRefreshToken.
type RefreshRotationStatus uint8

const (
	// RefreshRotationInvalid means the token does not exist, was revoked, or
	// expired before it could be claimed.
	RefreshRotationInvalid RefreshRotationStatus = iota
	// RefreshRotationSucceeded means the old token was marked used and its
	// replacement was inserted in the same transaction.
	RefreshRotationSucceeded
	// RefreshRotationReuseDetected means a previously-used token was presented;
	// its entire family has been revoked in the same transaction.
	RefreshRotationReuseDetected
)

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
	// Kept for compatibility with older callers. A refresh rotation must use
	// RotateRefreshToken instead, because it also writes the replacement in the
	// same transaction.
	result, err := r.db.ExecContext(ctx, `
		UPDATE refresh_token
		SET used_at = now()
		WHERE id = $1
		  AND used_at IS NULL
		  AND revoked_at IS NULL
		  AND expires_at > now()
	`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *pgRepo) InsertRefreshToken(ctx context.Context, userID int64, hash, familyID string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO refresh_token (user_id, token_hash, family_id, expires_at)
		VALUES ($1, $2, $3, $4)
	`, userID, hash, familyID, expiresAt)
	return err
}

// RotateRefreshToken atomically claims a one-time refresh token and inserts its
// replacement. The conditional UPDATE is the concurrency boundary: exactly one
// caller can transition used_at from NULL. If another caller presents that old
// token afterwards, it revokes the complete family before returning.
func (r *pgRepo) RotateRefreshToken(ctx context.Context, oldHash, replacementHash string, replacementExpiresAt time.Time) (RefreshRotationStatus, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return RefreshRotationInvalid, err
	}
	defer func() { _ = tx.Rollback() }()

	var userID int64
	var familyID string
	err = tx.QueryRowContext(ctx, `
		UPDATE refresh_token
		SET used_at = now()
		WHERE token_hash = $1
		  AND used_at IS NULL
		  AND revoked_at IS NULL
		  AND expires_at > now()
		RETURNING user_id, family_id
	`, oldHash).Scan(&userID, &familyID)
	if err == nil {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO refresh_token (user_id, token_hash, family_id, expires_at)
			VALUES ($1, $2, $3, $4)
		`, userID, replacementHash, familyID, replacementExpiresAt); err != nil {
			return RefreshRotationInvalid, err
		}
		if err := tx.Commit(); err != nil {
			return RefreshRotationInvalid, err
		}
		return RefreshRotationSucceeded, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return RefreshRotationInvalid, err
	}

	// The conditional UPDATE did not claim a row. Lock and inspect the current
	// state after any competing transaction has committed, so reuse handling is
	// deterministic rather than based on a stale pre-read.
	var row RefreshTokenRow
	err = tx.QueryRowContext(ctx, `
		SELECT id, user_id, token_hash, family_id, expires_at, revoked_at, used_at
		FROM refresh_token
		WHERE token_hash = $1
		FOR UPDATE
	`, oldHash).Scan(&row.ID, &row.UserID, &row.TokenHash, &row.FamilyID, &row.ExpiresAt, &row.RevokedAt, &row.UsedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return RefreshRotationInvalid, nil
	}
	if err != nil {
		return RefreshRotationInvalid, err
	}

	// Preserve existing semantics: a family already revoked or an expired token
	// is simply invalid. Only a live, previously-used token signals theft/reuse.
	if row.RevokedAt != nil || !row.ExpiresAt.After(time.Now()) {
		return RefreshRotationInvalid, nil
	}
	if row.UsedAt == nil {
		// This should be unreachable because the conditional UPDATE and this
		// SELECT run in one transaction. Fail closed if storage state is unusual.
		return RefreshRotationInvalid, nil
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE refresh_token
		SET revoked_at = now()
		WHERE family_id = $1 AND revoked_at IS NULL
	`, row.FamilyID); err != nil {
		return RefreshRotationInvalid, err
	}
	if err := tx.Commit(); err != nil {
		return RefreshRotationInvalid, err
	}
	return RefreshRotationReuseDetected, nil
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
