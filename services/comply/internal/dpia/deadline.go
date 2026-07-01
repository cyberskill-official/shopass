package dpia

import "time"

const (
	FilingWindow = 60 * 24 * time.Hour
	ReviewCycle  = 6 * 30 * 24 * time.Hour
)

// Status suy tu deadline, KHONG nhap tay.
func Status(a ProcessingActivity, d DPIA, now time.Time) string {
	filingDue := a.StartedAt.Add(FilingWindow)
	if d.FiledAt == nil {
		if now.After(filingDue) {
			return "overdue"
		}
		return "draft"
	}
	base := *d.FiledAt
	if d.LastReviewedAt != nil {
		base = *d.LastReviewedAt
	}
	if now.After(base.Add(ReviewCycle)) {
		return "review_overdue"
	}
	return "submitted"
}
