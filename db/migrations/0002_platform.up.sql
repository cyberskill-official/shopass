CREATE TABLE platform (
  id         SMALLINT     PRIMARY KEY,
  code       TEXT         UNIQUE NOT NULL
               CHECK (code IN ('shopee','tiktok','lazada')),
  country    TEXT         NOT NULL
               CHECK (country ~ '^[A-Z]{2}$'),
  base_url   TEXT,
  created_at TIMESTAMPTZ  DEFAULT now()
);
