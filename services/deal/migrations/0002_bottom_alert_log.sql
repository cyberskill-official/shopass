-- services/deal/migrations/0002_bottom_alert_log.sql
-- Nhật ký cảnh báo đáy đã bắn. Dùng cho dedupe ngày (UNIQUE) và cooldown liên ngày.
CREATE TABLE bottom_alert_log (
  user_id    BIGINT       NOT NULL,
  product_id BIGINT       NOT NULL REFERENCES tracked_product(id),
  fired_on   DATE         NOT NULL,                 -- ngày lịch theo Asia/Ho_Chi_Minh
  p_bottom   DOUBLE PRECISION NOT NULL CHECK (p_bottom > 0.7),
  fired_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
  UNIQUE (user_id, product_id, fired_on)            -- idempotent 1 alert/cặp/ngày (§1 #7)
);

-- Tra cooldown liên ngày: alert gần nhất của 1 cặp (user, product).
CREATE INDEX idx_bal_cooldown ON bottom_alert_log (user_id, product_id, fired_on DESC);
