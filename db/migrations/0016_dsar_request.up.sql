CREATE TABLE dsar_request (
  id           BIGSERIAL   PRIMARY KEY,
  user_id      BIGINT      NOT NULL,
  kind         TEXT        NOT NULL CHECK (kind IN ('access','rectify','erase','portability')),
  status       TEXT        NOT NULL DEFAULT 'open',
  requested_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  sla_due_at   TIMESTAMPTZ NOT NULL,
  completed_at TIMESTAMPTZ,
  note         TEXT
);

CREATE INDEX idx_dsar_user ON dsar_request (user_id, requested_at DESC);
CREATE INDEX idx_dsar_open ON dsar_request (status, sla_due_at) WHERE status <> 'completed';
