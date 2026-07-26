package gating

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type PgRuleSource struct {
	db *sql.DB
}

func NewPgRuleSource(db *sql.DB) *PgRuleSource {
	return &PgRuleSource{db: db}
}

func (s *PgRuleSource) Load(ctx context.Context) ([]Rule, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("gating rule source requires db")
	}

	rows, err := s.db.QueryContext(ctx, `
SELECT country_code, gate_key, allowed, COALESCE(value, ''), version
FROM country_rule
ORDER BY country_code, gate_key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []Rule
	for rows.Next() {
		var rule Rule
		if err := rows.Scan(&rule.Country, &rule.GateKey, &rule.Allowed, &rule.Value, &rule.Version); err != nil {
			return nil, err
		}
		if err := ValidateRule(rule); err != nil {
			return nil, fmt.Errorf("country_rule %s/%s: %w", rule.Country, rule.GateKey, err)
		}
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rules, nil
}
