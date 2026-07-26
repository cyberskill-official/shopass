CREATE TABLE IF NOT EXISTS market_trend_daily (
  category_id  INT NOT NULL,
  platform_id  SMALLINT NOT NULL,
  day          DATE NOT NULL,
  sku_count    INT NOT NULL,
  p25          BIGINT,
  median       BIGINT,
  p75          BIGINT,
  suppressed   BOOLEAN NOT NULL DEFAULT false,
  PRIMARY KEY (category_id, platform_id, day)
);
