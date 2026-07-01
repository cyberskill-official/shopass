package breach

import "time"

const AuthorityWindow = 72 * time.Hour

// DeadlineFlag suy tu dong ho, KHONG nhap tay.
func DeadlineFlag(b BreachIncident, now time.Time) string {
	due := b.AcknowledgedAt.Add(AuthorityWindow)
	if b.NotifiedAuthorityAt == nil {
		if now.After(due) {
			return "breach_overdue"
		}
		return "within_window"
	}
	return "notified"
}
