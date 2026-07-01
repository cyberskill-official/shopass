CREATE TABLE voucher_catalog (
  id             BIGSERIAL   PRIMARY KEY,
  platform_id    SMALLINT    NOT NULL REFERENCES platform(id),
  code           TEXT        NOT NULL,
  type           TEXT        NOT NULL CHECK (type IN ('shop','platform','freeship')),
  discount_type  TEXT        NOT NULL CHECK (discount_type IN ('amount','percent')),
  discount_value BIGINT      NOT NULL CHECK (discount_value > 0
                              AND (discount_type <> 'percent' OR discount_value <= 100)),
  min_spend      BIGINT      CHECK (min_spend IS NULL OR min_spend >= 0),  -- VND
  cap            BIGINT      CHECK (cap IS NULL OR cap > 0),               -- VND, tran giam
  shop_id        TEXT,
  valid_from     TIMESTAMPTZ NOT NULL,
  valid_to       TIMESTAMPTZ NOT NULL CHECK (valid_to >= valid_from),
  stack_group    TEXT,
  CONSTRAINT shop_id_by_type CHECK (
    (type = 'shop' AND shop_id IS NOT NULL) OR
    (type <> 'shop' AND shop_id IS NULL)
  ),
  UNIQUE (platform_id, code)
);

CREATE INDEX idx_vc_active ON voucher_catalog (platform_id, type, valid_to);
