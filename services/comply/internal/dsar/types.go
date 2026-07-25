package dsar

import (
	"time"

	"shopass/services/comply/internal/consent"
)

type DSARRequest struct {
	ID          int64
	UserID      int64
	Kind        string
	Status      string
	RequestedAt time.Time
	SLADueAt    time.Time
	CompletedAt *time.Time
	Note        *string
}

type AccountView struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Locale string `json:"locale"`
}

type ProductView struct {
	ID       int64  `json:"id"`
	Platform string `json:"platform"`
	Name     string `json:"name"`
}

// ConsentView is an alias for consent.ConsentRecord since we can use it directly
type ConsentView = consent.ConsentRecord

type ExportBundle struct {
	Account         AccountView   `json:"account"`
	TrackedProducts []ProductView `json:"tracked_products"`
	ConsentHistory  []ConsentView `json:"consent_history"`
	GeneratedAt     time.Time     `json:"generated_at"`
}

type EraseResult struct {
	WishlistDeleted    int    `json:"wishlist_deleted"`
	PaymentsAnonymized int    `json:"payments_anonymized"`
	ConsentLogRetained bool   `json:"consent_log_retained"`
	Status             string `json:"status"`
}
