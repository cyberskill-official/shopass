package coin

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

// UpsertTask inserts or updates a task for a user on a platform.
func (r *Repo) UpsertTask(ctx context.Context, userID int64, platformID int16, t CoinTask) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_activity_coin_task (user_id, platform_id, task_type, due_date, done)
         VALUES ($1,$2,$3,$4,$5)
         ON CONFLICT (user_id, platform_id, task_type, due_date) DO UPDATE SET done = EXCLUDED.done`,
		userID, platformID, t.TaskType, t.DueDate, t.Done)
	return err
}

// ListPending returns tasks that are not done for the specific user and day.
func (r *Repo) ListPending(ctx context.Context, userID int64, day time.Time) ([]CoinTask, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT task_type, due_date, done 
         FROM user_activity_coin_task 
         WHERE user_id = $1 AND due_date = $2 AND done = false`,
		userID, day)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []CoinTask
	for rows.Next() {
		var t CoinTask
		if err := rows.Scan(&t.TaskType, &t.DueDate, &t.Done); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}
