---
id: NFR-INFRA-001
title: "API performance - p95 < 300ms cho đọc cache, biểu đồ giá < 500ms; ngưỡng hiệu năng nền của mọi bề mặt người dùng"
module: INFRA
category: performance
priority: MUST
verification: T
phase: P0
slo: "p95 API đọc cache < 300ms; p95 render dữ liệu biểu đồ giá < 500ms; đo tại gateway, tách theo route-class"
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-INFRA-001, FR-INFRA-004, FR-PRICE-003, FR-DEAL-003, FR-WEB-003]
source: "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.8 (NFR hiệu năng: p95 API < 300ms đọc cache; biểu đồ giá < 500ms)"
---

## §1 - Statement (BCP-14 normative)

1. Với API đọc phục vụ từ cache (ví dụ trạng thái theo dõi, wishlist, đọc cấu hình), p95 độ trễ end-to-end đo tại gateway **MUST** < 300ms.
2. Với truy vấn dữ liệu biểu đồ lịch sử giá (FR-PRICE-003 / FR-DEAL-003 đọc continuous aggregate `price_daily`), p95 **MUST** < 500ms, kể cả ở quy mô `price_snapshot` hàng tỷ dòng.
3. Ngưỡng p95 **MUST** đo tại biên gateway (FR-INFRA-001), tách theo route-class, để con số phản ánh trải nghiệm thực của client chứ không chỉ thời gian xử lý nội bộ một service.
4. API ghi (đăng ký, theo dõi, tạo alert) **SHOULD** đạt p95 < 600ms; đây là mục tiêu mềm vì đường ghi có thể chạm nhiều service.
5. Khi p95 của một route-class vượt ngưỡng kéo dài (ví dụ > 5 phút), hệ thống **MUST** phát cảnh báo (qua FR-INFRA-004) để điều tra trước khi người dùng than phiền diện rộng.
6. Đo lường p95 **MUST** dựa trên histogram thực tế (Prometheus), KHÔNG dựa trên trung bình; trung bình che đuôi chậm mà người dùng cảm nhận.

## §2 - Vì sao ràng buộc này

Tốc độ là một phần của sản phẩm so giá: người dùng mở biểu đồ giá hay danh sách theo dõi và mong thấy ngay. Nếu đọc cache mất hơn 300ms hay biểu đồ mất hơn 500ms, trải nghiệm thành ì, và một công cụ tra cứu nhanh mất lý do tồn tại. Ngưỡng này đặc biệt khó với biểu đồ vì `price_snapshot` là bảng lớn nhất hệ thống (hàng tỷ dòng); chỉ đạt được nhờ continuous aggregate `price_daily` (FR-PRICE-002) và hypertable chunking. Đo tại gateway và theo p95 (không phải trung bình) buộc ta nhìn đúng cái người dùng chịu: đuôi chậm, không phải con số trung bình đẹp che lấp.

## §3 - Đo lường (measurement)

- Histogram `http_request_duration_ms{service,route,status}` tại gateway và mỗi service (FR-INFRA-004 #4).
- Phân loại route-class: `read_cached`, `price_chart`, `write`, `auth` - mỗi class một panel p95 riêng trên Grafana.
- Đường ngưỡng 300ms và 500ms vẽ trực tiếp trên panel latency (FR-INFRA-004 #12) để vi phạm thấy ngay.
- Theo dõi đồng thời cache hit-rate (đọc cache chỉ < 300ms nếu hit cao); p95 đọc cache xấu thường đi cùng hit-rate tụt.

## §4 - Verification

- Load test (T): bắn tải đại diện vào `read_cached` và `price_chart` ở quy mô dữ liệu lớn (seed `price_snapshot` >=1 tỷ dòng theo NFR-PRICE-001), xác nhận p95 < 300ms / < 500ms.
- Regression gate (T): CI chạy benchmark route biểu đồ; p95 vượt 500ms làm fail gate hiệu năng.
- Quan sát production: panel p95 per-route-class; alert khi vượt ngưỡng > 5 phút.
- Đối chiếu: p95 đọc cache cao bất thường -> kiểm cache hit-rate và đường tới Redis trước khi nghi DB.

## §5 - Xử lý khi vi phạm

- p95 đọc cache > 300ms kéo dài -> sev-3; kiểm cache hit-rate, kết nối Redis, kích thước payload; thêm cache layer nếu cần.
- p95 biểu đồ > 500ms -> sev-3; xác nhận `price_daily` continuous aggregate refresh kịp (FR-PRICE-002), xem lại index/chunk_interval, tránh quét raw cho khoảng dài.
- p95 write > 600ms kéo dài -> sev-4; tìm chặng chậm qua trace (FR-INFRA-004); cân nhắc async hóa phần không cần đồng bộ.
- Hồi quy do deploy (p95 nhảy sau release) -> rollback release nghi ngờ; benchmark lại trước khi rollout lại.

---

*Hết NFR-INFRA-001.*
