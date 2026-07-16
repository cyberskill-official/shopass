---
id: NFR-PRICE-001
title: "PRICE time-series scale - price_snapshot tỷ-dòng phải đọc <500ms và nén để giữ unit economics"
module: PRICE
category: scalability
priority: MUST
verification: T
phase: P1
slo: "Query lịch sử 90 ngày 1 SKU p95 < 500ms ở quy mô >=1 tỷ dòng; storage time-series <= ~0,1-0,2 USD/user/tháng (biến phí, §4.1)"
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_tasks: [TASK-PRICE-002, TASK-PRICE-003, TASK-DEAL-001, TASK-SCRAPE-005]
source: "docs/... §3.8 (NFR khả năng mở rộng), §3.4 (lý do hypertable/partitioning), §4.1 (unit economics)"
---

## §1 - Statement (BCP-14 normative)

1. Ở quy mô >=1 tỷ dòng `price_snapshot` (hàng triệu SKU x snapshot delta-only), query lịch sử 90 ngày của 1 SKU qua continuous aggregate `price_daily` **MUST** đạt p95 < 500ms.
2. Query raw `price_snapshot` cho 1 SKU trong 7 ngày gần nhất **MUST** đạt p95 < 300ms (chạm tối đa ~1-2 chunk).
3. Storage time-series (sau nén) **MUST** giữ biến phí <= ~0,1-0,2 USD/user/tháng theo unit economics §4.1 - đạt qua delta-only (TASK-PRICE-002 #4) + nén columnar 30 ngày (#5).
4. Tỷ lệ nén trên chunk cũ hơn 30 ngày **SHOULD** đạt >= 8x (segmentby `product_id`, giá tương quan cao).
5. Ghi snapshot (delta-only path) **MUST** đạt throughput >= 5.000 INSERT-quyết-định/giây/node để theo kịp scraping farm 3 sàn lúc flash sale.
6. Retention + compression policy **MUST** chạy tự động; không có chunk raw nào tồn tại quá 18 tháng.

## §2 - Vì sao ràng buộc này

`price_snapshot` là bảng lớn nhất và là nguồn đọc của mọi tính năng lõi (sale ảo, biểu đồ, dự đoán đáy, so sánh chéo). Nếu query lịch sử quét toàn bảng -> biểu đồ và sale ảo vỡ NFR-INFRA-001 (<500ms), trải nghiệm hỏng. Nếu không nén/không delta-only -> storage bùng nổ, phá unit economics (§4.1) vốn là điều kiện sống còn của mô hình free-tier tài trợ bằng affiliate. Đây là ràng buộc nền tảng quyết định cả trải nghiệm và biên lợi nhuận.

## §3 - Đo lường (measurement)

- Histogram `price_query_duration_ms{source=raw|daily, range_bucket}` - đo p95 query.
- Gauge `price_snapshot_rows_total` + `price_snapshot_bytes_compressed` / `_uncompressed` - tỷ lệ nén.
- Counter `price_snapshot_written_total` vs `delta_skipped_total` - hiệu quả delta-only.
- Báo cáo chi phí storage/user hằng tháng (Grafana panel) đối chiếu ngưỡng §4.1.

## §4 - Verification

- Load test (T): seed >=1 tỷ dòng tổng hợp (sinh từ ~1 triệu SKU x 1.000 snapshot), đo p95 query 90 ngày qua `price_daily` < 500ms.
- Compression test (T): nén 1 chunk cũ, đo `_compressed/_uncompressed` >= 8x.
- Throughput test (T): bắn 5.000 InsertSnapshot/s với 90% no-change, xác nhận delta-only giữ ghi thực ở mức thấp.
- Reconciliation: tổng `delta_skipped + written` = tổng lần quét; tỷ lệ skip >= 80% ở SKU thường.

## §5 - Xử lý khi vi phạm

- Query 90 ngày p95 > 500ms -> sev-3; kiểm continuous aggregate có refresh kịp; xem xét index/chunk_interval.
- Tỷ lệ nén < 8x -> sev-3; xem lại segmentby; có thể skew dữ liệu.
- Storage/user vượt ngưỡng §4.1 -> sev-2; kiểm delta-only có chạy (delta_skipped thấp bất thường = bug `changed()`), kiểm retention policy.
- Throughput ghi không theo kịp flash sale (hàng đợi scraper đọng) -> sev-2; scale node price-svc + tinh chỉnh batch INSERT.

---

*Hết NFR-PRICE-001.*
