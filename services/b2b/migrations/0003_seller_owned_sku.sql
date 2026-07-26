-- TASK-B2B-003: verified seller ownership of SKUs/shops.
CREATE TABLE IF NOT EXISTS seller_owned_sku (
  id            BIGSERIAL   PRIMARY KEY,
  seller_org_id BIGINT      NOT NULL,
  shop_id       TEXT        NOT NULL,
  product_id    BIGINT      NOT NULL,
  verified      BOOLEAN     NOT NULL DEFAULT false,
  linked_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (seller_org_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_sos_verified ON seller_owned_sku (seller_org_id, shop_id)
  WHERE verified = true;
