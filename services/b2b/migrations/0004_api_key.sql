-- TASK-B2B-004: public API keys + usage (hash-only secrets).
CREATE TABLE IF NOT EXISTS api_key (
  id            BIGSERIAL   PRIMARY KEY,
  prefix        TEXT        NOT NULL UNIQUE,
  secret_hash   TEXT        NOT NULL,
  org_name      TEXT        NOT NULL,
  tier          TEXT        NOT NULL CHECK (tier IN ('free','pro','enterprise')),
  rate_per_min  INTEGER     NOT NULL CHECK (rate_per_min > 0),
  monthly_quota INTEGER     NOT NULL CHECK (monthly_quota > 0),
  revoked       BOOLEAN     NOT NULL DEFAULT false,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS api_usage (
  id          BIGSERIAL   PRIMARY KEY,
  api_key_id  BIGINT      NOT NULL REFERENCES api_key(id),
  endpoint    TEXT        NOT NULL,
  ts          TIMESTAMPTZ NOT NULL DEFAULT now(),
  status_code SMALLINT    NOT NULL
  -- MUST NOT store response body / payload content (DEC-B2B-35)
);

CREATE INDEX IF NOT EXISTS idx_usage_key_month ON api_usage (api_key_id, ts);
