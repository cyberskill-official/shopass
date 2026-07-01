package dpia

import "time"

type ProcessingActivity struct {
	ID               int64
	Name             string
	PurposeKey       string
	DataCategories   []string
	StartedAt        time.Time
	CrossBorder      bool
	RecipientCountry *string
	CreatedAt        time.Time
}

type DPIA struct {
	ID             int64
	ActivityID     int64
	Version        int
	RiskLevel      string
	MitigationVi   *string
	Status         string
	FiledAt        *time.Time
	LastReviewedAt *time.Time
	CreatedAt      time.Time
}

type TIA struct {
	ID               int64
	DPIAID           int64
	RecipientCountry string
	SafeguardVi      string
	CreatedAt        time.Time
}

type TIAInput struct {
	RecipientCountry string
	Safeguard        string
}

type DPIAInput struct {
	RiskLevel    string
	MitigationVi string
	TIA          *TIAInput
}

type ActivityStatus struct {
	Name      string
	StartedAt time.Time
	Version   int
	Status    string
}
