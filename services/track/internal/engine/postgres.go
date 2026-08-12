package engine

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"shopass/services/track/internal/track"
)

// UnknownDeal never reports SALE_XIN, so real_sale rules cannot false-fire
// until dealsvc is wired as a live DetectFakeSale client.
type UnknownDeal struct{}

func (UnknownDeal) DetectFakeSale(context.Context, int64, int64, *int64) (DealVerdict, error) {
	return Unknown, nil
}

// PGStateRepo persists rising-edge state in alert_fired_state.
type PGStateRepo struct {
	DB *sql.DB
}

func (r *PGStateRepo) LastConditionMet(ctx context.Context, ruleID int64) (bool, error) {
	var met bool
	err := r.DB.QueryRowContext(ctx, `
		SELECT last_condition_met FROM alert_fired_state WHERE alert_rule_id = $1
	`, ruleID).Scan(&met)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return met, err
}

func (r *PGStateRepo) Set(ctx context.Context, ruleID int64, met bool) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO alert_fired_state (alert_rule_id, last_condition_met, last_fired_at)
		VALUES ($1, $2, CASE WHEN $2 THEN now() ELSE NULL END)
		ON CONFLICT (alert_rule_id) DO UPDATE SET
			last_condition_met = EXCLUDED.last_condition_met,
			last_fired_at = CASE
				WHEN EXCLUDED.last_condition_met THEN now()
				ELSE alert_fired_state.last_fired_at
			END
	`, ruleID, met)
	return err
}

// PGMedian reads the shared price_snapshot hypertable for drop_pct rules.
type PGMedian struct {
	DB *sql.DB
}

func (m *PGMedian) Median7d(ctx context.Context, productID int64) (int64, error) {
	var median sql.NullFloat64
	err := m.DB.QueryRowContext(ctx, `
		SELECT percentile_cont(0.5) WITHIN GROUP (ORDER BY price)
		FROM price_snapshot
		WHERE product_id = $1 AND ts >= now() - interval '7 days'
	`, productID).Scan(&median)
	if err != nil {
		return 0, err
	}
	if !median.Valid {
		return 0, fmt.Errorf("no snapshots for product %d", productID)
	}
	return int64(median.Float64), nil
}

// SQLHandoff inserts an in-app alert row and best-effort enqueues notifsvc.
type SQLHandoff struct {
	DB       *sql.DB
	NotifURL string
	HTTP     *http.Client
	Log      *slog.Logger
}

func (h *SQLHandoff) CreateAndEnqueue(ctx context.Context, r track.AlertRule, payload map[string]any) (int64, error) {
	if payload == nil {
		payload = map[string]any{}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	var id int64
	err = h.DB.QueryRowContext(ctx, `
		INSERT INTO alert (alert_rule_id, fired_at, payload, status)
		VALUES ($1, now(), $2, 'pending')
		RETURNING id
	`, r.ID, raw).Scan(&id)
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(h.NotifURL) == "" {
		return id, nil
	}
	body, err := json.Marshal(map[string]any{
		"user_id":    r.UserID,
		"product_id": r.ProductID,
		"reason":     r.RuleType,
		"payload":    payload,
	})
	if err != nil {
		return id, nil
	}
	client := h.HTTP
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.NotifURL, bytes.NewReader(body))
	if err != nil {
		h.log().Warn("alert notif enqueue skipped", "alert_id", id, "err", err)
		return id, nil
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		h.log().Warn("alert notif enqueue failed", "alert_id", id, "err", err)
		return id, nil
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		h.log().Warn("alert notif enqueue rejected", "alert_id", id, "status", resp.StatusCode)
	}
	return id, nil
}

func (h *SQLHandoff) log() *slog.Logger {
	if h != nil && h.Log != nil {
		return h.Log
	}
	return slog.Default()
}
