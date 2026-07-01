CREATE TABLE plan_catalog (
  id             SMALLSERIAL PRIMARY KEY,
  tier           TEXT     NOT NULL UNIQUE
                   CHECK (tier IN ('free','premium_basic','premium_plus','premium_pro')),
  price          BIGINT   NOT NULL CHECK (price >= 0),
  billing_period TEXT     NOT NULL DEFAULT 'monthly',
  active         BOOLEAN  NOT NULL DEFAULT true
);

INSERT INTO plan_catalog (tier, price) VALUES
  ('free',          0),
  ('premium_basic', 29000),
  ('premium_plus',  49000),
  ('premium_pro',   79000)
ON CONFLICT (tier) DO NOTHING;
