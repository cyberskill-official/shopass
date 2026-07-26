package adapters

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"shopass/services/comply/internal/dsar"
)

type Users struct {
	pool *pgxpool.Pool
}

func NewUsers(pool *pgxpool.Pool) *Users {
	return &Users{pool: pool}
}

func (u *Users) View(ctx context.Context, userID int64) (dsar.AccountView, error) {
	var view dsar.AccountView
	err := u.pool.QueryRow(ctx, `
		SELECT id, COALESCE(email::text, ''), locale
		FROM app_user
		WHERE id = $1
	`, userID).Scan(&view.UserID, &view.Email, &view.Locale)
	return view, err
}

func (u *Users) Anonymize(ctx context.Context, userID int64) error {
	_, err := u.pool.Exec(ctx, `
		UPDATE app_user
		SET email = 'deleted_' || id || '@tombstone.local',
		    phone = NULL,
		    display_name = NULL,
		    status = 'deleted'
		WHERE id = $1
	`, userID)
	return err
}
