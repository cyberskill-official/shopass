package referral

import "context"

type ReferralCode struct {
	ID     int64
	UserID int64
	Code   string
	Uses   int
}

type Repo interface {
	FindByCode(ctx context.Context, code string) (ReferralCode, bool, error)
	HasReferrer(ctx context.Context, userID int64) (bool, error)
	SetReferrer(ctx context.Context, userID int64, codeID int64) error
	IncrementUses(ctx context.Context, codeID int64) error
	CreateCodeForUser(ctx context.Context, userID int64) (string, error)
}
