package main

import (
	"context"
	"log/slog"

	"shopass/services/auth/internal/auth"
)

// auditResetNotifier records that a reset was requested without logging the
// token or PII. Live email/ZNS delivery is R23 (SMTP/Zalo creds).
type auditResetNotifier struct {
	log *slog.Logger
}

func (n *auditResetNotifier) SendReset(_ context.Context, u auth.AppUser, _ string) error {
	if n == nil || n.log == nil {
		return nil
	}
	n.log.Info("password_reset_issued", "user_id", u.ID, "delivery", "pending_r23_smtp")
	return nil
}
