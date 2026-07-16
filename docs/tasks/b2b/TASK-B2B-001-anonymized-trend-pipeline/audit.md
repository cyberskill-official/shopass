---
fr_id: TASK-B2B-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

TASK-B2B-001 đặc tả pipeline dữ liệu xu hướng thị trường ẩn danh ở mức triển khai được. 11 mệnh đề §1 normative, mỗi mệnh đề có AC tương ứng và test trong §5. Hai lớp phòng vệ bảo mật được đặc tả rõ và độc lập: chỉ đọc dữ liệu đã tổng hợp theo ngày (DEC-B2B-01) và cổng k-anonymity K_MIN = 50 (DEC-B2B-02). Lược đồ đầu ra cố ý không chứa khóa định danh (DEC-B2B-06). Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - Đường rò rỉ qua bảng user-level (đã chốt)
Ban đầu pipeline có thể vô tình join sang wishlist/cart để làm giàu dữ liệu. Rủi ro: tạo kênh dẫn dữ liệu cá nhân ra sản phẩm bán-được. Giải: §1 #2 + DEC-B2B-01 ép chỉ đọc `price_daily` + `tracked_product`; AC #9 (test grep truy vấn) + disallowed_tools.

### ISS-002 - Ngưỡng k-anonymity và biên
Bản đầu chưa định lượng "đủ ẩn danh". Nếu ô chỉ 1-2 SKU thì median ô gần như là giá một sản phẩm. Giải: §1 #4 + DEC-B2B-02 đặt K_MIN = 50; AC #2/#3/#4 kiểm biên 49/50; test `TestKAnon_BelowThreshold_Suppressed` + `TestKAnon_AtThreshold_Published`.

### ISS-003 - Phân biệt "chưa tính" với "cố ý không phát hành"
Nếu suppress bằng cách bỏ hẳn dòng thì audit không chứng minh được cổng đã chạy. Giải: §1 #8 + DEC-B2B-05 ghi `suppressed=true` + `sku_count` thật, chỉ số NULL; `QueryCells` lọc mặc định; test `TestQueryCells_SkipsSuppressed`.

### ISS-004 - Tái lập dữ liệu bán-được
Dữ liệu B2B bán ra phải tái lập để chứng minh đúng khi khách chất vấn. Giải: §1 #6 + DEC-B2B-04 batch UPSERT idempotent; AC #6 + test `TestJob_Idempotent`.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 schema không khóa định danh | #1,#10 | `0001_market_trend_daily.sql` + `TestAggregate_NoIdentifierColumns` |
| #2 chỉ đọc price_daily | #9 | truy vấn aggregate.go (join price_daily + tracked_product) |
| #3 chỉ số ô | #7,#8 | `aggregate.go` percentile_cont |
| #4 cổng k-anonymity | #2,#3,#4 | `kanon.go::applyKAnon` + 2 test biên |
| #5 không phát hành cá thể | #10 | lược đồ + DEC-B2B-03 |
| #6 batch idempotent | #6 | `job.go` + `TestJob_Idempotent` |
| #7 repo funcs | #5 | `repo.go` QueryCells/UpsertCells |
| #8 suppress ghi audit | #3,#5 | `TestQueryCells_SkipsSuppressed` |
| #9 bất biến phân vị | #7,#8 | CHECK + `TestAggregate_PercentileOrder` |

## §4 - Kết luận

Toàn bộ mệnh đề normative có code/SQL/test backing. Không có mệnh đề "mồ côi". Hai lớp bảo mật (tổng hợp-only + k-anonymity) được kiểm bằng test biên và test cấu trúc lược đồ. Score = 10/10. Verdict: PASS. Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit TASK-B2B-001.*
