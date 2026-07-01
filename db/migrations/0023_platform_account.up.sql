CREATE TABLE platform_account (
  id           BIGSERIAL    PRIMARY KEY,
  user_id      BIGINT       NOT NULL REFERENCES app_user(id),
  platform_id  SMALLINT     NOT NULL REFERENCES platform(id),
  ext_user_ref TEXT         NOT NULL
                 CHECK (length(ext_user_ref) BETWEEN 1 AND 128),
  linked_at    TIMESTAMPTZ  DEFAULT now(),
  UNIQUE (user_id, platform_id)
);
CREATE INDEX idx_pa_user ON platform_account (user_id);
