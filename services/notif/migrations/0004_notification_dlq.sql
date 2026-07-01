-- Dead-letter queue for notifications (FR-NOTIF-003)
CREATE TABLE notification_dlq (
  id              BIGSERIAL   PRIMARY KEY,
  notification_id BIGINT      NOT NULL REFERENCES notification(id),
  channel         TEXT        NOT NULL,
  payload         JSONB,
  attempts        INTEGER     NOT NULL,
  last_error      TEXT        NOT NULL,
  reason          TEXT        NOT NULL,
  dead_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_dlq_channel ON notification_dlq (channel, dead_at DESC);
