package consent

import (
	"errors"
	"net/netip"
	"time"
)

var (
	ErrUnknownPurpose = errors.New("unknown purpose")
)

type Purpose string

const (
	PurposeCartRead     Purpose = "cart_read"
	PurposeTracking     Purpose = "price_tracking"
	PurposeMarketing    Purpose = "marketing_notification"
	PurposeAnalyticsB2B Purpose = "analytics_b2b"
)

func validPurpose(p Purpose) bool {
	switch p {
	case PurposeCartRead, PurposeTracking, PurposeMarketing, PurposeAnalyticsB2B:
		return true
	}
	return false
}

type ConsentRecord struct {
	ID            int64       `db:"id"`
	UserID        int64       `db:"user_id"`
	PurposeKey    string      `db:"purpose_key"`
	PolicyVersion int32       `db:"policy_version"`
	Granted       bool        `db:"granted"`
	Source        string      `db:"source"`
	TS            time.Time   `db:"ts"`
	IP            *netip.Addr `db:"ip"`
	UserAgent     *string     `db:"user_agent"`
}

type ReqMeta struct {
	IP        *netip.Addr
	UserAgent *string
}
