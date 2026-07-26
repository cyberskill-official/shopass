CREATE TABLE IF NOT EXISTS cashback_entry (
  id             BIGSERIAL PRIMARY KEY,
  conversion_id  BIGINT NOT NULL UNIQUE,
  user_id        BIGINT NOT NULL REFERENCES app_user(id),
  commission     BIGINT NOT NULL CHECK (commission >= 0),
  user_share     BIGINT NOT NULL CHECK (user_share >= 0),
  kept_margin    BIGINT NOT NULL CHECK (kept_margin >= 0),
  status         TEXT NOT NULL DEFAULT 'held'
    CHECK (status IN ('held','released','clawed_back','paid')),
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  released_at    TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS payout_request (
  id          BIGSERIAL PRIMARY KEY,
  user_id     BIGINT NOT NULL REFERENCES app_user(id),
  amount      BIGINT NOT NULL CHECK (amount > 0),
  status      TEXT NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending','paid','failed')),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
