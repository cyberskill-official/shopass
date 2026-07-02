-- services/auth/migrations/0007_social_identity.up.sql
-- Links an OAuth/OIDC identity provider subject to an app_user (FR-AUTH-004).
CREATE TABLE social_identity (
  id         BIGSERIAL   PRIMARY KEY,
  user_id    BIGINT      NOT NULL REFERENCES app_user(id),
  provider   TEXT        NOT NULL CHECK (provider IN ('google','facebook','zalo')),
  subject    TEXT        NOT NULL,          -- 'sub' from the provider (stable id)
  created_at TIMESTAMPTZ DEFAULT now(),
  UNIQUE (provider, subject)                -- §1 #5: one provider identity -> one app_user
);
CREATE INDEX idx_si_user ON social_identity (user_id);
