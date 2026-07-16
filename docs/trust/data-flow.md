# SănDeal - Luồng dữ liệu (data flow)

Trang sàn (DOM)
   │  content script đọc giá/giỏ (TASK-EXT-002)
   ▼
[minimize] allowlist + redact + validate  (TASK-EXT-003)  <-- CHẠY TRÊN CLIENT
   │  CHỈ {platform, productId, price, qty} (+voucher) đi tiếp; cookie/PII bị loại
   ▼  <===== ĐÂY là điểm DỮ LIỆU RỜI MÁY (đã sạch) =====
hàng đợi đồng bộ (TASK-EXT-005) -> đính JWT SănDeal (KHÔNG token sàn)
   ▼
Backend price-svc (lưu price_snapshot)

KHÔNG có mũi tên nào mang cookie/mật khẩu/token sàn rời máy.
