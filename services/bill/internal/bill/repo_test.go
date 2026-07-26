package bill

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) *Repo {
	ctx := context.Background()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/shopass_test?sslmode=disable"
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skip("Database not available")
	}
	if err := pool.Ping(ctx); err != nil {
		t.Skip("Database ping failed")
	}
	t.Cleanup(func() { pool.Close() })

	_, _ = pool.Exec(ctx, `DELETE FROM payment`)
	_, _ = pool.Exec(ctx, `DELETE FROM subscription`)
	_, _ = pool.Exec(ctx, `DELETE FROM app_user`)

	return NewRepo(pool, nil)
}

func setupWithUser(t *testing.T) (*Repo, int64) {
	r := setup(t)
	ctx := context.Background()
	var uid int64
	email := fmt.Sprintf("test_bill_user_%s_%d@example.com", strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()), time.Now().UnixNano())
	err := r.pool.QueryRow(ctx, `INSERT INTO app_user (email) VALUES ($1) RETURNING id`, email).Scan(&uid)
	require.NoError(t, err)
	return r, uid
}

func planID(t *testing.T, r *Repo, tier string) int16 {
	var id int16
	err := r.pool.QueryRow(context.Background(), `SELECT id FROM plan_catalog WHERE tier=$1`, tier).Scan(&id)
	require.NoError(t, err)
	return id
}

func priceOf(t *testing.T, r *Repo, tier string) int64 {
	var price int64
	err := r.pool.QueryRow(context.Background(), `SELECT price FROM plan_catalog WHERE tier=$1`, tier).Scan(&price)
	require.NoError(t, err)
	return price
}

func TestCreate_Active(t *testing.T) {
	r, uid := setupWithUser(t)
	ctx := context.Background()
	id, err := r.CreateSubscription(ctx, uid, planID(t, r, "premium_basic"), time.Now().AddDate(0, 1, 0))
	require.NoError(t, err)
	sub, ok, _ := r.GetActive(ctx, uid)
	require.True(t, ok)
	require.Equal(t, id, sub.ID)
	require.Equal(t, "active", sub.Status)
}

func TestCreate_OneActivePerUser(t *testing.T) {
	r, uid := setupWithUser(t)
	ctx := context.Background()
	renews := time.Now().AddDate(0, 1, 0)
	_, err1 := r.CreateSubscription(ctx, uid, planID(t, r, "premium_basic"), renews)
	require.NoError(t, err1)
	_, err2 := r.CreateSubscription(ctx, uid, planID(t, r, "premium_plus"), renews)
	require.Error(t, err2) // partial unique WHERE status='active'
}

func TestPlanCatalog_Prices(t *testing.T) {
	r := setup(t)
	require.Equal(t, int64(29000), priceOf(t, r, "premium_basic"))
	require.Equal(t, int64(49000), priceOf(t, r, "premium_plus"))
	require.Equal(t, int64(79000), priceOf(t, r, "premium_pro"))
}

func TestSubscription_StatusCheck(t *testing.T) {
	r, uid := setupWithUser(t)
	ctx := context.Background()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO subscription (user_id, plan_id, renews_at, status)
		 VALUES ($1,$2, now()+interval '1 month', 'trialing')`, uid, planID(t, r, "free"))
	require.Error(t, err) // CHECK status IN (...)
}

func TestSubscription_RenewsAfterStart(t *testing.T) {
	r, uid := setupWithUser(t)
	ctx := context.Background()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO subscription (user_id, plan_id, started_at, renews_at)
		 VALUES ($1,$2, now(), now()-interval '1 day')`, uid, planID(t, r, "premium_basic"))
	require.Error(t, err) // CHECK renews_at > started_at
}
