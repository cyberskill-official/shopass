-- TASK-B2B-002: B2B insights subscription + entitlement.
CREATE TABLE IF NOT EXISTS b2b_subscription (
  id             BIGSERIAL   PRIMARY KEY,
  org_name       TEXT        NOT NULL,
  tier           TEXT        NOT NULL CHECK (tier IN ('basic','pro','enterprise')),
  max_categories INTEGER     NOT NULL CHECK (max_categories > 0),
  history_days   INTEGER     NOT NULL CHECK (history_days > 0),
  can_export     BOOLEAN     NOT NULL DEFAULT false,
  status         TEXT        NOT NULL DEFAULT 'active'
                   CHECK (status IN ('active','past_due','canceled')),
  started_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at     TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_b2bsub_active ON b2b_subscription (id)
  WHERE status = 'active';
