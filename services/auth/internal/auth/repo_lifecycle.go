package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type PasswordResetRow struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
}

func (r *pgRepo) FindByIdentifier(ctx context.Context, identifier string) (AppUser, bool) {
	// identifier is email
	u, err := r.FindByEmail(ctx, identifier)
	if err != nil || u.ID == 0 {
		return AppUser{}, false
	}
	return u, true
}

func (r *pgRepo) SaveReset(ctx context.Context, userID int64, hash string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO password_reset (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, hash, expiresAt)
	return err
}

func (r *pgRepo) FindResetByHash(ctx context.Context, hash string) (PasswordResetRow, bool) {
	var row PasswordResetRow
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token_hash, expires_at, used_at
		FROM password_reset WHERE token_hash = $1
	`, hash).Scan(&row.ID, &row.UserID, &row.TokenHash, &row.ExpiresAt, &row.UsedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return row, false
	}
	return row, err == nil
}

func (r *pgRepo) UpdatePassword(ctx context.Context, userID int64, newHash string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE app_user SET pwd_hash = $1 WHERE id = $2`, newHash, userID)
	return err
}

func (r *pgRepo) MarkResetUsed(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE password_reset SET used_at = now() WHERE id = $1`, id)
	return err
}

func (r *pgRepo) RevokeAllRefresh(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_token SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	return err
}

func (r *pgRepo) SetStatus(ctx context.Context, userID int64, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE app_user SET status = $1 WHERE id = $2`, status, userID)
	return err
}

func (r *pgRepo) AnonymizePII(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE app_user SET email = 'deleted_' || id || '@tombstone.local', phone = NULL WHERE id = $1`, userID)
	return err
}

func (r *pgRepo) DeletePlatformAccounts(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM platform_account WHERE user_id = $1`, userID)
	return err
}
