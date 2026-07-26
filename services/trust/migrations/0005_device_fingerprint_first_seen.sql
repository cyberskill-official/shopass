-- TRUST-006: first_seen + hash lookup index (0004 shipped last_seen only).
ALTER TABLE device_fingerprint
  ADD COLUMN IF NOT EXISTS first_seen TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE INDEX IF NOT EXISTS idx_fp_hash ON device_fingerprint (device_hash);
