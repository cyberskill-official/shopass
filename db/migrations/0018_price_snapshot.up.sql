CREATE TABLE price_snapshot (
  product_id   BIGINT      NOT NULL REFERENCES tracked_product(id),
  ts           TIMESTAMPTZ NOT NULL,
  price        BIGINT      NOT NULL CHECK (price > 0),
  list_price   BIGINT      CHECK (list_price IS NULL OR list_price >= price),
  stock        INTEGER,
  sold         INTEGER,
  flash_sale   BOOLEAN     NOT NULL DEFAULT false,
  PRIMARY KEY (product_id, ts)
);

SELECT create_hypertable('price_snapshot', 'ts',
  chunk_time_interval => INTERVAL '7 days');

CREATE INDEX idx_ps_flash ON price_snapshot (product_id, ts DESC)
  WHERE flash_sale = true;
