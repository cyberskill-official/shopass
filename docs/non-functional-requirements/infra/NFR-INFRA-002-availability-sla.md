---
id: NFR-INFRA-002
title: "Availability - SLA 99,5% cho bề mặt người dùng lõi; ngân sách lỗi và đo uptime"
module: INFRA
category: availability
priority: MUST
verification: T
phase: P0
slo: "Uptime hằng tháng >= 99,5% cho bề mặt lõi (gateway + đọc giá/theo dõi/biểu đồ); ngân sách lỗi ~3h39m/tháng"
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-INFRA-001, FR-INFRA-004, FR-AUTH-002, FR-PRICE-003, FR-NOTIF-002]
source: "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.8 (NFR khả dụng: SLA 99,5%)"
---

## §1 - Statement (BCP-14 normative)

1. Bề mặt người dùng lõi (gateway, đọc lịch sử giá, đọc theo dõi/wishlist, biểu đồ) **MUST** đạt uptime hằng tháng >= 99,5%, tương đương ngân sách lỗi (error budget) khoảng 3 giờ 39 phút mỗi tháng.
2. Uptime **MUST** đo bằng tỉ lệ request thành công trên tổng request hợp lệ tại gateway (success-rate SLI), KHÔNG chỉ bằng ping cổng. Một cổng "ping được" nhưng trả 5xx vẫn là downtime với người dùng.
3. Đường đọc lõi **MUST** suy biến nhẹ nhàng (graceful degradation) khi một phụ thuộc không lõi hỏng: ví dụ continuous aggregate trễ thì biểu đồ hiện dữ liệu cũ kèm dấu hiệu, thay vì lỗi toàn trang.
4. Thành phần trên đường lõi **SHOULD** không có điểm chết đơn lẻ (single point of failure) ở mức triển khai production: gateway, DB, Redis chạy chế độ dự phòng/replica.
5. Bảo trì có kế hoạch **SHOULD** thực hiện theo cách không downtime (rolling deploy, migration forward-only tương thích ngược) để không tiêu ngân sách lỗi vào việc dự đoán được.
6. Khi ngân sách lỗi trong tháng cạn (đã dùng hết ~3h39m), nhóm **MUST** ưu tiên ổn định hơn tính năng mới cho tới khi sang chu kỳ mới (chính sách error-budget).

## §2 - Vì sao ràng buộc này

Người dùng dựa vào SănDeal đúng lúc quyết định mua: khi sàn flash sale, khi cần biết "sale này thật hay ảo". Nếu dịch vụ sập đúng lúc đó, giá trị bằng không và niềm tin (vốn là moat hậu-Honey, §5.4) sứt mẻ. 99,5% là mức thực tế cho một startup giai đoạn đầu: đủ tin cậy để người dùng quay lại, nhưng không đắt như "bốn số chín" vốn đòi hạ tầng và quy trình tốn kém chưa tương xứng giai đoạn này. Đo bằng success-rate, không phải ping, vì một cổng trả 5xx hàng loạt vẫn "sống" theo ping nhưng đã chết với người dùng. Graceful degradation giữ phần lõi sống khi một phụ thuộc phụ chập chờn, bảo vệ phần lớn ngân sách lỗi.

## §3 - Đo lường (measurement)

- SLI success-rate: `sum(rate(http_requests_total{status!~"5.."}[5m])) / sum(rate(http_requests_total[5m]))` tại gateway (FR-INFRA-004).
- Error-budget burn-rate alert: cảnh báo nhiều cửa sổ (ví dụ burn nhanh 1h và burn chậm 6h) để bắt cả sự cố cấp tính lẫn rò rỉ chậm.
- Uptime hằng tháng tổng hợp trên Grafana, đối chiếu ngưỡng 99,5% và phần ngân sách lỗi đã tiêu.
- Probe tổng hợp (synthetic) định kỳ trên vài hành trình lõi (đăng nhập, đọc biểu đồ) như tín hiệu bổ trợ cho SLI thực.

## §4 - Verification

- Đo production (T): tính SLI success-rate hằng tháng; xác nhận >= 99,5% trên bề mặt lõi.
- Chaos/degradation test (T): tắt một phụ thuộc không lõi (ví dụ làm chậm cagg refresh) và xác nhận biểu đồ suy biến nhẹ (dữ liệu cũ + dấu hiệu) thay vì lỗi toàn trang.
- Deploy test: rolling deploy + migration forward-only không gây gián đoạn đường đọc (đo success-rate quanh cửa sổ deploy).
- Đối chiếu: so SLI thực với probe tổng hợp; lệch lớn gợi ý lỗ hổng đo lường.

## §5 - Xử lý khi vi phạm

- Burn-rate nhanh (sự cố cấp tính) -> sev-2; kích hoạt điều tra qua trace/log (FR-INFRA-004); ưu tiên khôi phục đường đọc lõi trước.
- Ngân sách lỗi tháng cạn -> freeze tính năng, dồn vào ổn định tới chu kỳ mới (§1 #6).
- 5xx tăng nhưng ping vẫn xanh -> tin SLI success-rate, không tin ping; tìm service trả lỗi qua metric per-service.
- Downtime do bảo trì ngoài kế hoạch lặp lại -> chuyển sang rolling deploy + migration tương thích ngược (§1 #5) để ngừng tiêu ngân sách vào việc dự đoán được.

---

*Hết NFR-INFRA-002.*
