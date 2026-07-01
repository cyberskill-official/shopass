CREATE TABLE consent_record (
  id             BIGSERIAL   PRIMARY KEY,
  user_id        BIGINT      NOT NULL REFERENCES app_user(id),
  purpose_key    TEXT        NOT NULL,
  policy_version INTEGER     NOT NULL,
  granted        BOOLEAN     NOT NULL,
  source         TEXT        NOT NULL CHECK (source IN ('web','extension','mobile')),
  ts             TIMESTAMPTZ NOT NULL DEFAULT now(),
  ip             INET,
  user_agent     TEXT,
  FOREIGN KEY (purpose_key, policy_version)
    REFERENCES consent_policy(purpose_key, version)
);

-- Tra trang thai hieu luc hien tai = dong moi nhat theo ts cho moi (user, purpose).
CREATE INDEX idx_consent_latest ON consent_record (user_id, purpose_key, ts DESC);
