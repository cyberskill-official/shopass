package report

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrPaymentRequired = errors.New("b2b: subscription required")
	ErrNoExport        = errors.New("b2b: export not allowed")
)

// ErrScopeExceeded is a 403 — never silently truncate (DEC-B2B-13).
type ErrScopeExceeded struct {
	Field string
	Limit int
}

func (e ErrScopeExceeded) Error() string {
	return fmt.Sprintf("b2b: scope exceeded on %s (limit %d)", e.Field, e.Limit)
}

type Entitlement struct {
	Tier          string
	MaxCategories int
	HistoryDays   int
	CanExport     bool
}

type Subscription struct {
	ID            int64
	OrgName       string
	Tier          string
	Status        string
	MaxCategories int
	HistoryDays   int
	CanExport     bool
	ExpiresAt     time.Time
}

type ReportScope struct {
	CategoryIDs []int64   `json:"category_ids"`
	PlatformIDs []int16   `json:"platform_ids"`
	From        time.Time `json:"from"`
	To          time.Time `json:"to"`
}

func EntitlementFrom(sub Subscription) Entitlement {
	return Entitlement{
		Tier:          sub.Tier,
		MaxCategories: sub.MaxCategories,
		HistoryDays:   sub.HistoryDays,
		CanExport:     sub.CanExport,
	}
}

func AssertActive(sub Subscription, now time.Time) error {
	if sub.Status != "active" || !sub.ExpiresAt.After(now) {
		return ErrPaymentRequired
	}
	return nil
}

// CheckScope clamps by entitlement; returns ErrScopeExceeded when over limit.
func CheckScope(e Entitlement, s ReportScope) error {
	if len(s.CategoryIDs) > e.MaxCategories {
		return ErrScopeExceeded{Field: "categories", Limit: e.MaxCategories}
	}
	if s.To.Sub(s.From) > time.Duration(e.HistoryDays)*24*time.Hour {
		return ErrScopeExceeded{Field: "history_days", Limit: e.HistoryDays}
	}
	return nil
}

func AssertExport(e Entitlement) error {
	if !e.CanExport {
		return ErrNoExport
	}
	return nil
}

// AssertScope keeps the older Subscription-based helper used by thin tests.
func AssertScope(sub Subscription, categories int, historyDays int) error {
	e := EntitlementFrom(sub)
	s := ReportScope{
		CategoryIDs: make([]int64, categories),
		From:        time.Unix(0, 0).UTC(),
		To:          time.Unix(0, 0).UTC().Add(time.Duration(historyDays) * 24 * time.Hour),
	}
	return CheckScope(e, s)
}
