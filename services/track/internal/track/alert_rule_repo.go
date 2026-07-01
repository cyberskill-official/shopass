package track

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/lib/pq" // using lib/pq for Array? Wait, pgx natively supports arrays or lib/pq.
	// We might just use pgx types but shopass probably uses lib/pq or standard sql.
)

type AlertRule struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"-"`
	ProductID int64     `json:"product_id"`
	RuleType  string    `json:"rule_type"`
	Threshold *int64    `json:"threshold"`
	Channel   []string  `json:"channel"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

type Alert struct {
	ID          int64           `json:"id"`
	AlertRuleID int64           `json:"alert_rule_id"`
	FiredAt     time.Time       `json:"fired_at"`
	Payload     json.RawMessage `json:"payload"`
	Status      string          `json:"status"`
}

type AlertRuleRepo interface {
	CreateRule(ctx context.Context, userID, productID int64, ruleType string, threshold *int64, channel []string) (AlertRule, error)
	ListRules(ctx context.Context, userID int64) ([]AlertRule, error)
	OwnsRule(ctx context.Context, userID, ruleID int64) (bool, error)
	ToggleActive(ctx context.Context, ruleID int64, active bool) error
	DeleteRule(ctx context.Context, ruleID int64) error
	ListAlerts(ctx context.Context, ruleID int64) ([]Alert, error)
}

type alertRuleRepoImpl struct {
	db *sql.DB
}

func NewAlertRuleRepo(db *sql.DB) AlertRuleRepo {
	return &alertRuleRepoImpl{db: db}
}

func (r *alertRuleRepoImpl) CreateRule(ctx context.Context, userID, productID int64, ruleType string, threshold *int64, channel []string) (AlertRule, error) {
	var rule AlertRule
	// We'll use standard lib/pq Array for TEXT[]
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO alert_rule (user_id, product_id, rule_type, threshold, channel)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, product_id, rule_type, threshold, channel, active, created_at
	`, userID, productID, ruleType, threshold, pq.Array(channel)).Scan(
		&rule.ID, &rule.UserID, &rule.ProductID, &rule.RuleType, &rule.Threshold, pq.Array(&rule.Channel), &rule.Active, &rule.CreatedAt,
	)
	return rule, err
}

func (r *alertRuleRepoImpl) ListRules(ctx context.Context, userID int64) ([]AlertRule, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, product_id, rule_type, threshold, channel, active, created_at
		FROM alert_rule
		WHERE user_id = $1
		ORDER BY id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []AlertRule
	for rows.Next() {
		var rule AlertRule
		if err := rows.Scan(&rule.ID, &rule.UserID, &rule.ProductID, &rule.RuleType, &rule.Threshold, pq.Array(&rule.Channel), &rule.Active, &rule.CreatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *alertRuleRepoImpl) OwnsRule(ctx context.Context, userID, ruleID int64) (bool, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM alert_rule WHERE id = $1 AND user_id = $2)
	`, ruleID, userID).Scan(&ok)
	return ok, err
}

func (r *alertRuleRepoImpl) ToggleActive(ctx context.Context, ruleID int64, active bool) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE alert_rule SET active = $1 WHERE id = $2
	`, active, ruleID)
	return err
}

func (r *alertRuleRepoImpl) DeleteRule(ctx context.Context, ruleID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM alert_rule WHERE id = $1
	`, ruleID)
	return err
}

func (r *alertRuleRepoImpl) ListAlerts(ctx context.Context, ruleID int64) ([]Alert, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, alert_rule_id, fired_at, payload, status
		FROM alert
		WHERE alert_rule_id = $1
		ORDER BY fired_at DESC
	`, ruleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var a Alert
		if err := rows.Scan(&a.ID, &a.AlertRuleID, &a.FiredAt, &a.Payload, &a.Status); err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}
