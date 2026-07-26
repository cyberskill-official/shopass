-- services/trust/migrations/0001_fraud_signal.sql
CREATE TABLE IF NOT EXISTS fraud_signal (
  id              BIGSERIAL PRIMARY KEY,
  subject_user_id BIGINT      NOT NULL REFERENCES app_user(id),
  kind            TEXT        NOT NULL,
  risk_score      SMALLINT    NOT NULL CHECK (risk_score BETWEEN 0 AND 100),
  reasons         JSONB       NOT NULL,
  status          TEXT        NOT NULL DEFAULT 'open'
                  CHECK (status IN ('open','investigating','confirmed_fraud','cleared')),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (subject_user_id, kind)
);
CREATE INDEX IF NOT EXISTS idx_fraud_open ON fraud_signal (status, risk_score DESC) WHERE status = 'open';
