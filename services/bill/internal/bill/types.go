package bill

import "time"

type Subscription struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	PlanID    int16     `db:"plan_id"`
	StartedAt time.Time `db:"started_at"`
	RenewsAt  time.Time `db:"renews_at"`
	Status    string    `db:"status"`
}

type MetricsClient interface {
	StatusChange(from, to string)
}
