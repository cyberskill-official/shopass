CREATE TABLE IF NOT EXISTS payout_hold (
  id              BIGSERIAL PRIMARY KEY,
  conversion_id   BIGINT NOT NULL UNIQUE,
  user_id         BIGINT NOT NULL REFERENCES app_user(id),
  amount          BIGINT NOT NULL CHECK (amount >= 0),
  status          TEXT NOT NULL DEFAULT 'held'
    CHECK (status IN ('held','under_investigation','eligible','released','denied')),
  hold_reason     TEXT,
  confirmed_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  eligible_at     TIMESTAMPTZ NOT NULL,
  released_at     TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_payout_hold_eligible ON payout_hold (status, eligible_at)
  WHERE status IN ('held','eligible');
