package auth

import (
	"context"
)

func (r *pgRepo) UpsertPlatformAccount(ctx context.Context, pa PlatformAccount) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO platform_account (user_id, platform_id, ext_user_ref)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, platform_id)
		 DO UPDATE SET ext_user_ref = EXCLUDED.ext_user_ref, linked_at = now()`,
		pa.UserID, pa.PlatformID, pa.ExtUserRef)
	return err
}

func (r *pgRepo) ListPlatformAccountsByUser(ctx context.Context, userID int64) ([]PlatformAccount, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, user_id, platform_id, ext_user_ref FROM platform_account WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PlatformAccount
	for rows.Next() {
		var pa PlatformAccount
		if err := rows.Scan(&pa.ID, &pa.UserID, &pa.PlatformID, &pa.ExtUserRef); err != nil {
			return nil, err
		}
		result = append(result, pa)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *pgRepo) DeletePlatformAccount(ctx context.Context, userID int64, platformID int16) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM platform_account WHERE user_id=$1 AND platform_id=$2`, userID, platformID)
	return err
}
