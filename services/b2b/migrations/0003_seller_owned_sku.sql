CREATE TABLE IF NOT EXISTS seller_owned_sku (
  seller_org_id BIGINT NOT NULL,
  product_id    BIGINT NOT NULL,
  verified      BOOLEAN NOT NULL DEFAULT false,
  PRIMARY KEY (seller_org_id, product_id)
);
