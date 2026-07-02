package auth

import (
	"context"
	"database/sql"
	"errors"
)

// pgRepo implements SocialRepo (it already provides FindByEmail via repo.go).
var _ SocialRepo = (*pgRepo)(nil)

// FindBySocial returns the app_user linked to a provider subject, if any.
func (r *pgRepo) FindBySocial(ctx context.Context, provider, subject string) (int64, bool, error) {
	var uid int64
	err := r.db.QueryRowContext(ctx,
		`SELECT user_id FROM social_identity WHERE provider = $1 AND subject = $2`,
		provider, subject).Scan(&uid)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return uid, true, nil
}

// LinkSocial attaches a provider subject to an app_user. ON CONFLICT keeps the
// UNIQUE(provider, subject) guarantee (§1 #5) idempotent under retry.
func (r *pgRepo) LinkSocial(ctx context.Context, userID int64, provider, subject string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO social_identity (user_id, provider, subject)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (provider, subject) DO NOTHING`,
		userID, provider, subject)
	return err
}

// CreateSocialUser creates a social-only app_user (pwd_hash NULL, §1 #7). An
// empty email is stored as NULL so an unverified-email account cannot collide
// with an existing email (the takeover defense in resolveUser).
func (r *pgRepo) CreateSocialUser(ctx context.Context, email, locale string) (int64, error) {
	if locale == "" {
		locale = "vi-VN"
	}
	var id int64
	// A social-only account: pwd_hash NULL (§1 #7) and no phone yet. FindByEmail
	// COALESCEs those NULLs when reading the row back.
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO app_user (email, locale, status, pwd_hash)
		 VALUES (NULLIF($1, ''), $2, 'active', NULL)
		 RETURNING id`,
		email, locale).Scan(&id)
	return id, err
}
