-- R26: model versioning + evaluation gate
-- reviewed: additive; nullable FK on price_forecast

CREATE TABLE IF NOT EXISTS model_run (
  id                BIGSERIAL PRIMARY KEY,
  version           TEXT        NOT NULL,
  model_kind        TEXT        NOT NULL CHECK (model_kind IN ('prophet', 'category_prior', 'lgbm')),
  training_window_start DATE,
  training_window_end   DATE,
  feature_set_hash  TEXT        NOT NULL DEFAULT '',
  backtest_mape     REAL,
  backtest_hit_rate REAL,
  gate_passed       BOOLEAN     NOT NULL DEFAULT false,
  gate_reason       TEXT        NOT NULL DEFAULT '',
  artifact_path     TEXT        NOT NULL DEFAULT '',
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_model_run_kind_created
  ON model_run (model_kind, created_at DESC);

ALTER TABLE price_forecast
  ADD COLUMN IF NOT EXISTS model_run_id BIGINT REFERENCES model_run(id);

CREATE INDEX IF NOT EXISTS idx_price_forecast_model_run
  ON price_forecast (model_run_id)
  WHERE model_run_id IS NOT NULL;
