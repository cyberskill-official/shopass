package coin

import "time"

// CoinTask represents a task for coins that the user needs to complete.
type CoinTask struct {
	TaskType string    `json:"task_type"`
	DueDate  time.Time `json:"due_date"`
	Done     bool      `json:"done"`
}
