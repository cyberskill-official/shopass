CREATE TABLE affiliate_click (
  id          BIGSERIAL   PRIMARY KEY,
  user_id     BIGINT      NOT NULL REFERENCES app_user(id),
  platform_id SMALLINT    NOT NULL REFERENCES platform(id),
  product_id  BIGINT      REFERENCES tracked_product(id),
  sub_id      TEXT        NOT NULL UNIQUE,
  network     TEXT        NOT NULL,
  clicked_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_click_user_time ON affiliate_click (user_id, clicked_at DESC);
