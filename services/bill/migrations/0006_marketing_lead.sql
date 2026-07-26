-- R39: Premium waitlist / marketing leads
CREATE TABLE IF NOT EXISTS marketing_lead (
  id         BIGSERIAL PRIMARY KEY,
  email      TEXT        NOT NULL,
  zalo      TEXT,
  source     TEXT        NOT NULL DEFAULT 'pricing',
  tier_interest TEXT     NOT NULL DEFAULT 'premium_basic'
               CHECK (tier_interest IN ('premium_basic','premium_plus','premium_pro')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (email, source)
);

CREATE INDEX IF NOT EXISTS idx_marketing_lead_created
  ON marketing_lead (created_at DESC);
