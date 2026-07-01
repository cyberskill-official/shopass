package breach

import "time"

type Status string

type BreachIncident struct {
	ID                  int64
	Summary             string
	Severity            string
	Status              Status
	OccurredAt          *time.Time
	AcknowledgedAt      time.Time
	TriagedAt           *time.Time
	NotifiedAuthorityAt *time.Time
	NotifiedSubjectsAt  *time.Time
	ClosedAt            *time.Time
	SourceRef           *string
	CreatedAt           time.Time
}

type BreachInput struct {
	Summary    string
	Severity   string
	OccurredAt *time.Time
	SourceRef  *string
}
