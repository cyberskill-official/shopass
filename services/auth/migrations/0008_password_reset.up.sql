CREATE TABLE password_reset (
  id          BIGSERIAL   PRIMARY KEY,
  user_id     BIGINT      NOT NULL REFERENCES app_user(id),
  token_hash  TEXT        NOT NULL UNIQUE,   -- hash token reset, KHÔNG cleartext
  expires_at  TIMESTAMPTZ NOT NULL,
  used_at     TIMESTAMPTZ,                   -- NULL = chưa dùng
  created_at  TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_pr_user ON password_reset (user_id);
