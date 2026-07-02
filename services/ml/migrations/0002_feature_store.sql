-- services/ml/migrations/0002_feature_store.sql
-- 1 dòng/(product_id, as_of_date): đủ 11 đặc trưng + cột nhãn nullable.
-- Đặc trưng chỉ nhìn quá khứ (<= as_of_date); nhãn nhìn 14 ngày tới, CHỈ điền lúc train.
CREATE TABLE feature_store (
  product_id               BIGINT  NOT NULL REFERENCES tracked_product(id),
  as_of_date               DATE    NOT NULL,
  day_of_month             SMALLINT NOT NULL,                 -- 1..31
  is_double_date           BOOLEAN NOT NULL,                  -- d == m
  days_to_next_double_date SMALLINT NOT NULL,
  is_payday_window         BOOLEAN NOT NULL,
  trailing_min_30          BIGINT  NOT NULL,                  -- VND
  trailing_min_60          BIGINT  NOT NULL,
  trailing_min_90          BIGINT  NOT NULL,
  price_vs_median90        REAL    NOT NULL,                  -- close_p / median90
  volatility_30d           REAL    NOT NULL,                  -- std log-return 30d
  category_seasonality     REAL    NOT NULL,                  -- chỉ số mùa vụ gộp theo category
  flash_sale_flag          BOOLEAN NOT NULL,
  platform_id              SMALLINT NOT NULL,
  future_min_price_14d     BIGINT,                            -- NHÃN: nullable, chỉ điền lúc train
  PRIMARY KEY (product_id, as_of_date)
);

CREATE INDEX idx_fs_asof ON feature_store (as_of_date);
