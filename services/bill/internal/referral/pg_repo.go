package referral

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGRepo struct {
	pool *pgxpool.Pool
}

func NewPGRepo(pool *pgxpool.Pool) *PGRepo {
	return &PGRepo{pool: pool}
}

func (r *PGRepo) FindByCode(ctx context.Context, code string) (ReferralCode, bool, error) {
	var out ReferralCode
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, code, uses
		FROM referral_code
		WHERE code = $1
	`, code).Scan(&out.ID, &out.UserID, &out.Code, &out.Uses)
	if errors.Is(err, pgx.ErrNoRows) {
		return ReferralCode{}, false, nil
	}
	if err != nil {
		return ReferralCode{}, false, err
	}
	return out, true, nil
}

func (r *PGRepo) FindByUser(ctx context.Context, userID int64) (ReferralCode, bool, error) {
	var out ReferralCode
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, code, uses
		FROM referral_code
		WHERE user_id = $1
	`, userID).Scan(&out.ID, &out.UserID, &out.Code, &out.Uses)
	if errors.Is(err, pgx.ErrNoRows) {
		return ReferralCode{}, false, nil
	}
	if err != nil {
		return ReferralCode{}, false, err
	}
	return out, true, nil
}

func (r *PGRepo) HasReferrer(ctx context.Context, userID int64) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx, `
		SELECT referral_code_id IS NOT NULL
		FROM app_user
		WHERE id = $1
	`, userID).Scan(&ok)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return ok, err
}

func (r *PGRepo) SetReferrer(ctx context.Context, userID int64, codeID int64) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE app_user
		SET referral_code_id = $1
		WHERE id = $2 AND referral_code_id IS NULL
	`, codeID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrAlreadyAttributed
	}
	return nil
}

func (r *PGRepo) IncrementUses(ctx context.Context, codeID int64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE referral_code SET uses = uses + 1 WHERE id = $1
	`, codeID)
	return err
}

func (r *PGRepo) CreateCodeForUser(ctx context.Context, userID int64) (string, error) {
	if existing, ok, err := r.FindByUser(ctx, userID); err != nil {
		return "", err
	} else if ok {
		return existing.Code, nil
	}
	for attempt := 0; attempt < 8; attempt++ {
		code := NewCode(6)
		var out string
		err := r.pool.QueryRow(ctx, `
			INSERT INTO referral_code (user_id, code)
			VALUES ($1, $2)
			ON CONFLICT (user_id) DO UPDATE SET code = referral_code.code
			RETURNING code
		`, userID, code).Scan(&out)
		if err == nil {
			return out, nil
		}
		// Unique code collision — retry.
	}
	return "", errors.New("could not allocate referral code")
}

// Ensure Repo still matches — FindByUser is extra helper used by API.
var _ Repo = (*PGRepo)(nil)
