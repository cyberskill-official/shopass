CREATE TABLE notification (
  id           BIGSERIAL   PRIMARY KEY,
  user_id      BIGINT      NOT NULL REFERENCES app_user(id),
  channel      TEXT        NOT NULL CHECK (channel IN ('push','email','sms')),
  template     TEXT        NOT NULL,
  payload      JSONB,
  scheduled_at TIMESTAMPTZ,
  sent_at      TIMESTAMPTZ,
  status       TEXT        NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending','queued','sent','failed')),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notif_dispatch ON notification (status, scheduled_at)
  WHERE status IN ('pending','queued');
