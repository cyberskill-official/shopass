CREATE TABLE tracked_product (
  id               BIGSERIAL   PRIMARY KEY,
  platform_id      SMALLINT    NOT NULL REFERENCES platform(id),
  platform_item_id TEXT        NOT NULL,
  shop_id          TEXT,
  title            TEXT,
  category_id      BIGINT,
  canonical_key    TEXT,                                  -- NULL cho tới khi FR-PRICE-005 so khớp
  first_seen       TIMESTAMPTZ NOT NULL DEFAULT now(),    -- mốc cold-start, bất biến
  UNIQUE (platform_id, platform_item_id)
);

CREATE INDEX idx_tp_canonical ON tracked_product (canonical_key);
