package ecom

import "time"

type EcommerceObligation struct {
	ID            int64      `json:"id"`
	ObligationKey string     `json:"obligation_key"`
	DescriptionVi string     `json:"description_vi"`
	Status        string     `json:"status"`
	DueAt         *time.Time `json:"due_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	SourceLaw     string     `json:"source_law"`
	Version       int        `json:"version"`
	CreatedAt     time.Time  `json:"created_at"`
}

type ThresholdState struct {
	Year         int   `json:"year"`
	Count        int64 `json:"count"`
	Threshold    int64 `json:"threshold"`
	MustRegister bool  `json:"must_register"`
}
