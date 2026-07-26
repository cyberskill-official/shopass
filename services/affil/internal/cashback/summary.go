package cashback

import "time"

// UserSummary is the GET /v1/cashback/summary payload (§1 #9 disclosure).
type UserSummary struct {
	PendingCount       int        `json:"-"`
	PendingAmount      int64      `json:"-"`
	NextAvailableAt    *time.Time `json:"-"`
	AvailableCount     int        `json:"-"`
	AvailableAmount    int64      `json:"-"`
	PaidTotal          int64      `json:"-"`
	PayoutThresholdVND int64      `json:"-"`
	Note               string     `json:"-"`
}

// SummaryResponse is the JSON shape for the summary endpoint.
type SummaryResponse struct {
	Pending struct {
		Count           int     `json:"count"`
		AmountVND       int64   `json:"amount_vnd"`
		NextAvailableAt *string `json:"next_available_at"`
	} `json:"pending"`
	Available struct {
		Count     int   `json:"count"`
		AmountVND int64 `json:"amount_vnd"`
	} `json:"available"`
	PaidTotalVND       int64  `json:"paid_total_vnd"`
	PayoutThresholdVND int64  `json:"payout_threshold_vnd"`
	Note               string `json:"note"`
}

func (s UserSummary) ToResponse(threshold int64) SummaryResponse {
	var resp SummaryResponse
	resp.Pending.Count = s.PendingCount
	resp.Pending.AmountVND = s.PendingAmount
	if s.NextAvailableAt != nil {
		d := s.NextAvailableAt.UTC().Format("2006-01-02")
		resp.Pending.NextAvailableAt = &d
	}
	resp.Available.Count = s.AvailableCount
	resp.Available.AmountVND = s.AvailableAmount
	resp.PaidTotalVND = s.PaidTotal
	if threshold <= 0 {
		threshold = DefaultConfig().PayoutThreshold
	}
	resp.PayoutThresholdVND = threshold
	resp.Note = DisclosureNote
	if s.Note != "" {
		resp.Note = s.Note
	}
	return resp
}
