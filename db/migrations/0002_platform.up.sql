CREATE TABLE platform (
  id         SMALLINT     PRIMARY KEY,
  code       TEXT         UNIQUE NOT NULL
               CHECK (code IN ('shopee','tiktok','lazada')),
  country    TEXT         NOT NULL
               CHECK (country ~ '^[A-Z]{2}$'),     -- ISO-3166 alpha-2
  base_url   TEXT,
  created_at TIMESTAMPTZ  DEFAULT now()
);

-- Canonical platform rows (VN first). Every product/price/affiliate row FKs
-- platform(id), so the lookup must be seeded. ids match the adapter constants
-- (shopee=1, tiktok=2, lazada=3).
INSERT INTO platform (id, code, country, base_url) VALUES
  (1, 'shopee', 'VN', 'https://shopee.vn'),
  (2, 'tiktok', 'VN', 'https://www.tiktok.com'),
  (3, 'lazada', 'VN', 'https://www.lazada.vn')
ON CONFLICT (id) DO NOTHING;
