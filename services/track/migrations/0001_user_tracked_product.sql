-- services/track/migrations/0001_user_tracked_product.sql
CREATE TABLE user_tracked_product (
  user_id    BIGINT      NOT NULL REFERENCES app_user(id),
  product_id BIGINT      NOT NULL REFERENCES tracked_product(id),
  tracked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, product_id)
);

-- Liệt kê nhanh mọi SKU một user đang theo dõi (cho wishlist/alert/dashboard)
CREATE INDEX idx_utp_user ON user_tracked_product (user_id);
-- Đếm nhanh số người theo dõi một SKU (chia sẻ chi phí scraping, ưu tiên tần suất quét)
CREATE INDEX idx_utp_product ON user_tracked_product (product_id);
