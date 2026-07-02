-- services/bill/migrations/0005_plan_feature.sql
CREATE TABLE plan_feature (
  id          BIGSERIAL PRIMARY KEY,
  tier        TEXT     NOT NULL CHECK (tier IN ('free','premium_basic','premium_plus','premium_pro')),
  feature_key TEXT     NOT NULL,
  limit_value BIGINT   NOT NULL,
  UNIQUE (tier, feature_key)
);

INSERT INTO plan_feature (tier, feature_key, limit_value) VALUES
  ('free',          'price_tracking',   -1),
  ('free',          'fake_sale_detect', -1),
  ('free',          'price_chart',      -1),
  ('free',          'wishlist_items',   20),
  ('free',          'bottom_predict',    0),
  ('premium_basic', 'wishlist_items',  100),
  ('premium_basic', 'bottom_predict',   -1),
  ('premium_plus',  'wishlist_items',  500),
  ('premium_plus',  'bottom_predict',   -1),
  ('premium_pro',   'wishlist_items',   -1),
  ('premium_pro',   'bottom_predict',   -1)
ON CONFLICT (tier, feature_key) DO NOTHING;
