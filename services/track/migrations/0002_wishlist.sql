-- services/track/migrations/0002_wishlist.sql
CREATE TABLE wishlist (
  id         BIGSERIAL   PRIMARY KEY,
  user_id    BIGINT      NOT NULL REFERENCES app_user(id),
  name       TEXT        NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_wishlist_user ON wishlist (user_id);

CREATE TABLE wishlist_item (
  id           BIGSERIAL   PRIMARY KEY,
  wishlist_id  BIGINT      NOT NULL REFERENCES wishlist(id) ON DELETE CASCADE,
  product_id   BIGINT      NOT NULL REFERENCES tracked_product(id),
  target_price BIGINT      CHECK (target_price IS NULL OR target_price > 0), -- VND
  added_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (wishlist_id, product_id)
);
