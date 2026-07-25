package bill

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestTransition_Valid(t *testing.T) {
	require.True(t, CanTransition("active", "past_due"))
	require.True(t, CanTransition("past_due", "active"))
	require.True(t, CanTransition("active", "canceled"))
}

func TestTransition_Invalid(t *testing.T) {
	require.False(t, CanTransition("canceled", "active")) // canceled là cuối
	require.False(t, CanTransition("expired", "active"))
}

func TestUpdateStatus_RejectsInvalid(t *testing.T) {
	pool, err := pgxpool.New(context.Background(), "postgres://postgres:postgres@localhost:5432/shopass_test?sslmode=disable")
	if err != nil {
		t.Skip("Database not available")
	}
	if err := pool.Ping(context.Background()); err != nil {
		t.Skip("Database ping failed")
	}
	defer pool.Close()
	r := NewRepo(pool, nil)
	uid := int64(1) // assume user 1 exists from repo_test.go setup

	// Ensure cleanup
	_, _ = pool.Exec(context.Background(), `DELETE FROM subscription`)

	id, err := r.CreateSubscription(context.Background(), uid, 2, time.Now().AddDate(0, 1, 0))
	require.NoError(t, err)

	require.NoError(t, r.UpdateStatus(context.Background(), id, "canceled"))
	require.ErrorIs(t, r.UpdateStatus(context.Background(), id, "active"), ErrInvalidTransition) // canceled -> active cấm
}
