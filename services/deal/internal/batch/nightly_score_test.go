package batch

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type mockNotif struct {
	items []NotifItem
}

func (m *mockNotif) Enqueue(ctx context.Context, item NotifItem) error {
	m.items = append(m.items, item)
	return nil
}

func (m *mockNotif) Count() int {
	return len(m.items)
}

func (m *mockNotif) Last() NotifItem {
	if len(m.items) == 0 {
		return NotifItem{}
	}
	return m.items[len(m.items)-1]
}

type testDeps struct {
	pool  *pgxpool.Pool
	notif *mockNotif
}

func setupBatch(t *testing.T) (*Batch, *testDeps) {
	// Require TEST_DB_URL to be set, or use a default one for local testing
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/shopass_deal_test?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)

	// Clean up tables
	_, err = pool.Exec(ctx, `TRUNCATE TABLE bottom_alert_log, alert_rule, price_forecast, tracked_product CASCADE`)
	require.NoError(t, err)

	// Seed FK prerequisites (platform + app_user) not covered by TRUNCATE.
	_, err = pool.Exec(ctx, `INSERT INTO platform (id, code, country) VALUES (1, 'shopee', 'VN') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO app_user (id) VALUES (999) ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	notif := &mockNotif{}
	b := New(pool, slog.Default(), notif)

	return b, &testDeps{pool: pool, notif: notif}
}

var (
	productID = int64(100)
	userID    = int64(999)
	today     = time.Now()
	ctx       = context.Background()
)

func seedMatureForecast(t *testing.T, deps *testDeps, pID int64, p float64) {
	_, err := deps.pool.Exec(ctx, `
		INSERT INTO tracked_product (id, platform_id, platform_item_id, first_seen)
		VALUES ($1::bigint, 1, 'item-' || $1::bigint::text, now() - INTERVAL '100 days')
		ON CONFLICT DO NOTHING`, pID)
	require.NoError(t, err)

	_, err = deps.pool.Exec(ctx, `
		INSERT INTO price_forecast (product_id, run_date, horizon_day, yhat, yhat_lower, yhat_upper, p_bottom_14d, model_kind, scored_at)
		VALUES ($1, current_date, 14, 1000, 900, 1100, $2, 'lgbm', now())`, pID, p)
	require.NoError(t, err)
}

func seedImmatureForecast(t *testing.T, deps *testDeps, pID int64, p float64) {
	_, err := deps.pool.Exec(ctx, `
		INSERT INTO tracked_product (id, platform_id, platform_item_id, first_seen)
		VALUES ($1::bigint, 1, 'item-' || $1::bigint::text, now() - INTERVAL '30 days')
		ON CONFLICT DO NOTHING`, pID)
	require.NoError(t, err)

	_, err = deps.pool.Exec(ctx, `
		INSERT INTO price_forecast (product_id, run_date, horizon_day, yhat, yhat_lower, yhat_upper, p_bottom_14d, model_kind, scored_at)
		VALUES ($1, current_date, 14, 1000, 900, 1100, $2, 'lgbm', now())`, pID, p)
	require.NoError(t, err)
}

func seedRule(t *testing.T, deps *testDeps, uID, pID int64, ruleType string) {
	_, err := deps.pool.Exec(ctx, `
		INSERT INTO alert_rule (user_id, product_id, rule_type, active)
		VALUES ($1, $2, $3, true)`, uID, pID, ruleType)
	require.NoError(t, err)
}

func seedBottomRule(t *testing.T, deps *testDeps, uID, pID int64) {
	seedRule(t, deps, uID, pID, "bottom_predicted")
}

func countAlertLog(t *testing.T, deps *testDeps, uID, pID int64) int {
	var count int
	err := deps.pool.QueryRow(ctx, `SELECT count(*) FROM bottom_alert_log WHERE user_id = $1 AND product_id = $2`, uID, pID).Scan(&count)
	require.NoError(t, err)
	return count
}

func TestNightly_FiresAboveThreshold(t *testing.T) {
	b, deps := setupBatch(t)
	seedMatureForecast(t, deps, productID, 0.71) // trên ngưỡng
	seedBottomRule(t, deps, userID, productID)   // có luật khớp
	require.NoError(t, b.RunNightlyScore(ctx, today))
	require.Equal(t, 1, deps.notif.Count()) // 0.71 -> bắn

	// 0.69 không bắn; 0.70 (đúng biên) cũng không bắn (strict >).
	for _, p := range []float64{0.69, 0.70} {
		b2, d2 := setupBatch(t)
		seedMatureForecast(t, d2, productID, p)
		seedBottomRule(t, d2, userID, productID)
		require.NoError(t, b2.RunNightlyScore(ctx, today))
		require.Equal(t, 0, d2.notif.Count(), "p=%.2f không được bắn", p)
	}
}

func TestNightly_RespectsMaturityGate(t *testing.T) {
	b, deps := setupBatch(t)
	seedImmatureForecast(t, deps, productID, 0.95) // SKU 30 ngày, P rất cao
	seedBottomRule(t, deps, userID, productID)
	require.NoError(t, b.RunNightlyScore(ctx, today))
	require.Equal(t, 0, deps.notif.Count()) // chưa MATURE -> không chấm, không bắn
}

func TestNightly_DedupePerDay(t *testing.T) {
	b, deps := setupBatch(t)
	seedMatureForecast(t, deps, productID, 0.80)
	seedBottomRule(t, deps, userID, productID)
	require.NoError(t, b.RunNightlyScore(ctx, today))
	require.NoError(t, b.RunNightlyScore(ctx, today)) // chạy lại cùng ngày
	require.Equal(t, 1, deps.notif.Count())            // chỉ 1 alert
	require.Equal(t, 1, countAlertLog(t, deps, userID, productID))
}

func TestNightly_Cooldown_NoRepeatWhileHigh(t *testing.T) {
	b, deps := setupBatch(t)
	seedMatureForecast(t, deps, productID, 0.85)
	seedBottomRule(t, deps, userID, productID)
	require.NoError(t, b.RunNightlyScore(ctx, today))            // ngày 0: bắn
	require.NoError(t, b.RunNightlyScore(ctx, today.AddDate(0, 0, 3))) // ngày 3, P vẫn cao
	require.Equal(t, 1, deps.notif.Count()) // trong cooldown 7 ngày -> không bắn lại
	
	// sau cooldown (ngày 8) thì cạnh lên lại -> bắn.
	require.NoError(t, b.RunNightlyScore(ctx, today.AddDate(0, 0, 8)))
	require.Equal(t, 2, deps.notif.Count())
}

func TestNightly_MatchesAlertRuleType(t *testing.T) {
	b, deps := setupBatch(t)
	seedMatureForecast(t, deps, productID, 0.90)
	seedRule(t, deps, userID, productID, "price_below") // sai type (valid enum, khac bottom_predicted)
	require.NoError(t, b.RunNightlyScore(ctx, today))
	require.Equal(t, 0, deps.notif.Count()) // chỉ 'bottom_predicted' mới khớp
}

func TestNightly_EnqueuesNotification(t *testing.T) {
	b, deps := setupBatch(t)
	seedMatureForecast(t, deps, productID, 0.78)
	seedBottomRule(t, deps, userID, productID)
	require.NoError(t, b.RunNightlyScore(ctx, today))
	item := deps.notif.Last()
	require.Equal(t, "bottom_predicted", item.Reason)
	require.Equal(t, userID, item.UserID)
	require.InDelta(t, 0.78, item.Payload["p_bottom_14d"], 1e-6)
}
