CREATE TABLE cart_snapshot (
  id           BIGSERIAL   PRIMARY KEY,
  user_id      BIGINT      NOT NULL REFERENCES app_user(id),
  platform_id  SMALLINT    NOT NULL REFERENCES platform(id),
  snapshot_ref UUID        NOT NULL,
  captured_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (user_id, snapshot_ref)        -- idempotent retry (DEC-CART-11)
);

CREATE TABLE cart_item (
  id               BIGSERIAL PRIMARY KEY,
  cart_snapshot_id BIGINT    NOT NULL REFERENCES cart_snapshot(id) ON DELETE CASCADE,
  product_id       BIGINT    REFERENCES tracked_product(id),  -- NULL nếu chưa track (DEC-CART-12)
  platform_item_id TEXT,                                      -- giữ ref thô khi product_id NULL
  shop_id          TEXT,
  qty              INTEGER   NOT NULL CHECK (qty > 0),
  unit_price       BIGINT    NOT NULL CHECK (unit_price > 0), -- VND
  -- phải có ít nhất một cách định danh SKU
  CONSTRAINT item_identified CHECK (product_id IS NOT NULL OR platform_item_id IS NOT NULL)
);
CREATE INDEX idx_ci_snapshot ON cart_item (cart_snapshot_id);
