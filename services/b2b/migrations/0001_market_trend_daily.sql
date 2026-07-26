-- TASK-B2B-001: anonymized market trend cells (k-anonymity gate).
-- Schema intentionally omits identity keys (DEC-B2B-06).
CREATE TABLE IF NOT EXISTS market_trend_daily (
  category_id      BIGINT      NOT NULL,
  platform_id      SMALLINT    NOT NULL,
  day              DATE        NOT NULL,
  median_p         BIGINT,                 -- NULL when suppressed
  p25_p            BIGINT,
  p75_p            BIGINT,
  avg_discount_pct NUMERIC(5,2),           -- 0..100
  sku_count        INTEGER     NOT NULL CHECK (sku_count >= 0),
  suppressed       BOOLEAN     NOT NULL DEFAULT false,
  computed_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (category_id, platform_id, day),
  CHECK (suppressed OR (p25_p <= median_p AND median_p <= p75_p))
);

CREATE INDEX IF NOT EXISTS idx_mtd_published ON market_trend_daily (category_id, platform_id, day)
  WHERE suppressed = false;
