-- services/track/migrations/0003_alert_rule.sql
CREATE TABLE alert_rule (
  id         BIGSERIAL   PRIMARY KEY,
  user_id    BIGINT      NOT NULL REFERENCES app_user(id),
  product_id BIGINT      NOT NULL REFERENCES tracked_product(id),
  rule_type  TEXT        NOT NULL
               CHECK (rule_type IN ('price_below','drop_pct','real_sale','bottom_predicted')),
  threshold  BIGINT,
  channel    TEXT[]      NOT NULL DEFAULT ARRAY['push'],
  active     BOOLEAN     NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ar_eval ON alert_rule (product_id, rule_type) WHERE active = true;
CREATE INDEX idx_ar_user ON alert_rule (user_id);

CREATE TABLE alert (
  id            BIGSERIAL   PRIMARY KEY,
  alert_rule_id BIGINT      NOT NULL REFERENCES alert_rule(id) ON DELETE CASCADE,
  fired_at      TIMESTAMPTZ NOT NULL,
  payload       JSONB,
  status        TEXT        NOT NULL DEFAULT 'pending'
);
CREATE INDEX idx_alert_rule_time ON alert (alert_rule_id, fired_at DESC);
