CREATE TABLE IF NOT EXISTS device_fingerprint (
  device_hash TEXT NOT NULL,
  user_id     BIGINT NOT NULL REFERENCES app_user(id),
  last_seen   TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (device_hash, user_id)
);
CREATE INDEX IF NOT EXISTS idx_device_fp_user ON device_fingerprint (user_id);
