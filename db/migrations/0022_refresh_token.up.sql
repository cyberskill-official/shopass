CREATE TABLE refresh_token (
  id          BIGSERIAL   PRIMARY KEY,
  user_id     BIGINT      NOT NULL REFERENCES app_user(id),
  token_hash  TEXT        NOT NULL,            -- hash của refresh token, KHÔNG cleartext
  family_id   UUID        NOT NULL,            -- nhóm token cùng một chuỗi rotation
  expires_at  TIMESTAMPTZ NOT NULL,
  revoked_at  TIMESTAMPTZ,                     -- NULL = còn hiệu lực
  used_at     TIMESTAMPTZ,                     -- đánh dấu đã xoay (dùng một lần)
  created_at  TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_rt_user ON refresh_token (user_id);
CREATE UNIQUE INDEX idx_rt_hash ON refresh_token (token_hash);
