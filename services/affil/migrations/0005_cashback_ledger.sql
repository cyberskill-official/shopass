-- TASK-AFFIL-005: cashback hold-then-release ledger (BIGINT VND).
-- Status vocab matches spec: pending|available|paid|clawed_back.

CREATE TABLE IF NOT EXISTS cashback_entry (
  id             BIGSERIAL PRIMARY KEY,
  user_id        BIGINT NOT NULL REFERENCES app_user(id),
  conversion_id  BIGINT NOT NULL REFERENCES affiliate_conversion(id),
  commission     BIGINT NOT NULL CHECK (commission >= 0),
  user_share     BIGINT NOT NULL CHECK (user_share >= 0),
  kept_margin    BIGINT NOT NULL CHECK (kept_margin >= 0),
  status         TEXT NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending','available','paid','clawed_back')),
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  available_at   TIMESTAMPTZ NOT NULL,
  paid_at        TIMESTAMPTZ,
  UNIQUE (conversion_id)
);

CREATE INDEX IF NOT EXISTS idx_cashback_release
  ON cashback_entry (status, available_at)
  WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_cashback_user_status
  ON cashback_entry (user_id, status);

CREATE TABLE IF NOT EXISTS payout_request (
  id          BIGSERIAL PRIMARY KEY,
  user_id     BIGINT NOT NULL REFERENCES app_user(id),
  amount      BIGINT NOT NULL CHECK (amount > 0),
  status      TEXT NOT NULL DEFAULT 'queued'
    CHECK (status IN ('queued','sent','failed')),
  gateway_ref TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
