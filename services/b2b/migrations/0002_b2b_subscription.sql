CREATE TABLE IF NOT EXISTS b2b_subscription (
  id             BIGSERIAL PRIMARY KEY,
  org_id         BIGINT NOT NULL,
  tier           TEXT NOT NULL CHECK (tier IN ('starter','pro','enterprise')),
  status         TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','canceled','expired')),
  max_categories INT NOT NULL DEFAULT 3,
  history_days   INT NOT NULL DEFAULT 30,
  can_export     BOOLEAN NOT NULL DEFAULT false,
  expires_at     TIMESTAMPTZ NOT NULL,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_b2b_sub_org ON b2b_subscription (org_id) WHERE status = 'active';
