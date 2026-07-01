CREATE TABLE dpia (
  id               BIGSERIAL   PRIMARY KEY,
  activity_id      BIGINT      NOT NULL REFERENCES processing_activity(id),
  version          INTEGER     NOT NULL,
  risk_level       TEXT        NOT NULL CHECK (risk_level IN ('low','medium','high')),
  mitigation_vi    TEXT,
  status           TEXT        NOT NULL DEFAULT 'draft',
  filed_at         TIMESTAMPTZ,
  last_reviewed_at TIMESTAMPTZ,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (activity_id, version)
);

CREATE TABLE tia (
  id                BIGSERIAL   PRIMARY KEY,
  dpia_id           BIGINT      NOT NULL REFERENCES dpia(id),
  recipient_country TEXT        NOT NULL,
  safeguard_vi      TEXT        NOT NULL,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_dpia_activity ON dpia (activity_id, version DESC);
