package report

import (
	"errors"
	"time"
)

var (
	ErrPaymentRequired = errors.New("b2b: subscription required")
	ErrForbiddenScope  = errors.New("b2b: scope exceeds tier")
	ErrNoExport        = errors.New("b2b: export not allowed")
)

type Subscription struct {
	OrgID         int64
	Tier          string
	Status        string
	MaxCategories int
	HistoryDays   int
	CanExport     bool
	ExpiresAt     time.Time
}

func AssertActive(sub Subscription, now time.Time) error {
	if sub.Status != "active" || !sub.ExpiresAt.After(now) {
		return ErrPaymentRequired
	}
	return nil
}

func AssertScope(sub Subscription, categories int, historyDays int) error {
	if categories > sub.MaxCategories || historyDays > sub.HistoryDays {
		return ErrForbiddenScope
	}
	return nil
}

func AssertExport(sub Subscription) error {
	if !sub.CanExport {
		return ErrNoExport
	}
	return nil
}
