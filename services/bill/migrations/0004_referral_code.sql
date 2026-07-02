-- services/bill/migrations/0004_referral_code.sql
CREATE TABLE referral_code (
  id         BIGSERIAL   PRIMARY KEY,
  user_id    BIGINT      NOT NULL UNIQUE REFERENCES app_user(id),
  code       TEXT        NOT NULL UNIQUE,
  uses       INTEGER     NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_referral_code_lookup ON referral_code (code);
-- Note: app_user.referral_code_id FK will be added later when app_user is managed.
