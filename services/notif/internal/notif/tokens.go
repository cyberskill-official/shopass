package notif

import (
	"context"
	"fmt"
)

// UpsertToken registers or refreshes a verified channel address for a user.
// platform must be one of ios|android|web|email|sms; channel one of push|email|sms.
func (r *Repo) UpsertToken(ctx context.Context, userID int64, channel, platform, address string) error {
	if userID <= 0 {
		return fmt.Errorf("user_id required")
	}
	if address == "" {
		return fmt.Errorf("address required")
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_channel_token (user_id, channel, platform, address, verified, updated_at)
		VALUES ($1, $2, $3, $4, true, now())
		ON CONFLICT (user_id, channel, platform) DO UPDATE
		SET address = EXCLUDED.address,
		    verified = true,
		    updated_at = now()`,
		userID, channel, platform, address)
	return err
}

// DeleteToken removes a push token matching address for the user (best-effort unregister).
func (r *Repo) DeleteToken(ctx context.Context, userID int64, address string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM user_channel_token
		WHERE user_id = $1 AND channel = 'push' AND address = $2`,
		userID, address)
	return err
}
