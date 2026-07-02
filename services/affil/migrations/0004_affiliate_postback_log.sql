CREATE TABLE affiliate_postback_log (
  id          BIGSERIAL   PRIMARY KEY,
  network     TEXT        NOT NULL,
  raw_payload JSONB       NOT NULL,        -- bằng chứng tranh chấp (§1 #6)
  signature   TEXT,
  verified    BOOLEAN     NOT NULL,
  received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_postback_received ON affiliate_postback_log (received_at DESC);
