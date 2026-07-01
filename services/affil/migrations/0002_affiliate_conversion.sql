CREATE TABLE affiliate_conversion (
  id           BIGSERIAL   PRIMARY KEY,
  click_id     BIGINT      NOT NULL UNIQUE REFERENCES affiliate_click(id),
  order_value  BIGINT      NOT NULL CHECK (order_value >= 0),
  commission   BIGINT      NOT NULL CHECK (commission  >= 0),
  status       TEXT        NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending','confirmed','rejected')),
  confirmed_at TIMESTAMPTZ
);
