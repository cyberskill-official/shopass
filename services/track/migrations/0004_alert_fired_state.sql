-- services/track/migrations/0004_alert_fired_state.sql
CREATE TABLE alert_fired_state (
  alert_rule_id      BIGINT      PRIMARY KEY REFERENCES alert_rule(id) ON DELETE CASCADE,
  last_condition_met BOOLEAN     NOT NULL DEFAULT false,
  last_fired_at      TIMESTAMPTZ
);
