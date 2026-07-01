CREATE TABLE breach_incident (
  id                    BIGSERIAL   PRIMARY KEY,
  summary               TEXT        NOT NULL,
  severity              TEXT        NOT NULL CHECK (severity IN ('low','medium','high','critical')),
  status                TEXT        NOT NULL DEFAULT 'detected'
                          CHECK (status IN ('detected','triaged','notified_authority','notified_subjects','closed')),
  occurred_at           TIMESTAMPTZ,
  acknowledged_at       TIMESTAMPTZ NOT NULL,
  triaged_at            TIMESTAMPTZ,
  notified_authority_at TIMESTAMPTZ,
  notified_subjects_at  TIMESTAMPTZ,
  closed_at             TIMESTAMPTZ,
  source_ref            TEXT,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_breach_open ON breach_incident (status, acknowledged_at)
  WHERE status <> 'closed';
