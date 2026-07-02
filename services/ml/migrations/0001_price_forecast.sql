-- services/ml/migrations/0001_price_forecast.sql
CREATE TABLE price_forecast (
  product_id    BIGINT      NOT NULL REFERENCES tracked_product(id),
  run_date      DATE        NOT NULL,
  horizon_day   SMALLINT    NOT NULL CHECK (horizon_day BETWEEN 1 AND 14),
  yhat          BIGINT      NOT NULL,
  yhat_lower    BIGINT      NOT NULL,
  yhat_upper    BIGINT      NOT NULL,
  p_bottom_14d  REAL        NOT NULL CHECK (p_bottom_14d BETWEEN 0 AND 1),
  model_kind    TEXT        NOT NULL CHECK (model_kind IN ('prophet', 'category_prior', 'lgbm')),
  scored_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (product_id, run_date, horizon_day)
);

CREATE INDEX idx_forecast_latest ON price_forecast (product_id, run_date DESC);
