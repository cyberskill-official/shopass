CREATE TABLE IF NOT EXISTS api_key (
  id           BIGSERIAL PRIMARY KEY,
  org_id       BIGINT NOT NULL,
  prefix       TEXT NOT NULL UNIQUE,
  secret_hash  TEXT NOT NULL,
  tier         TEXT NOT NULL,
  revoked_at   TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS api_usage (
  api_key_id BIGINT NOT NULL REFERENCES api_key(id),
  day        DATE NOT NULL,
  hits       INT NOT NULL DEFAULT 0,
  PRIMARY KEY (api_key_id, day)
);
