---
fr_id: TASK-PRICE-004
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Tôi tái thẩm TASK-PRICE-004 độc lập từ file `.md` hiện tại. Endpoint `GET /v1/compare?canonical_key=...` khớp shape §3.7 và đặc tả ở mức triển khai được: 12 mệnh đề §1 normative (11 MUST + 1 SHOULD), mỗi mệnh đề có AC §4 và test §5. Giá hiện tại lấy đúng nguồn qua `DISTINCT ON (product_id) ... ORDER BY product_id, ts DESC` trên `price_snapshot` (không dùng `price_daily`, giữ `ts` giây/phút), JOIN `tracked_product` theo `canonical_key` + `platform`. `is_cheapest` tính phía server, trả `ts` mỗi sàn cho độ tươi. Xử lý đúng việc bảng `platform` không có cột `name` (suy `platform_name` từ `code` qua `displayName`). Không có defect cần sửa. Score 10/10.

## §2 - Findings

Không phát hiện defect ở lần thẩm độc lập này. Các điểm tôi kiểm chính diện và xác nhận đạt:

- ISS-001 (kiểm, đạt): lấy giá hiện tại đúng. `DISTINCT ON (product_id)` trong LATERAL đã lọc theo `tp.id` là dư nhẹ nhưng vẫn là SQL hợp lệ và đúng kết quả (một dòng mới nhất mỗi product). `disallowed_tools` cấm `price_daily`; §1 #4 + AC #4 chốt (DEC-PRICE-41).
- ISS-002 (kiểm, đạt): bảng `platform` (TASK-INFRA-002) chỉ có `id/code/country/base_url`, KHÔNG có `name`. task né đúng: SELECT chỉ `pf.code`, suy tên qua `displayName(code)`. Không SELECT cột không tồn tại. Khớp DATA-MODEL.
- ISS-003 (kiểm, đạt): tiền tệ. `Price int64` ở cả `CompareRow` và `CompareItem`; JSON không phần thập phân (§1 #7, AC #7). `currency:"VND"` là chuỗi nhãn, không phải giá. Đồng nhất DEC-PRICE-05.
- ISS-004 (kiểm, đạt): biên xử lý đủ. Thiếu key -> 400 không chạm DB (§1 #2, AC #2); key lạ -> 404 không phải mảng rỗng (§1 #10, AC #10); 1 sàn vẫn trả `is_cheapest=true` (§1 #9, AC #9). `markCheapest` set mọi dòng = min, xử lý đúng bằng giá (§1 #6, AC #6).
- ISS-005 (kiểm, đạt): rủi ro over-merge nằm ở TASK-PRICE-005 (matching), không ở endpoint này; §10 ghi nhận và trỏ DEC-PRICE-23. Ranh giới depends_on TASK-PRICE-005 đúng.
- Typography: prose sạch ASCII; mũi tên chỉ trong code. Đạt rule O.

## §3 - Traceability §1 -> AC -> artefact

| §1 (mệnh đề) | §4 AC | §5 test / §3 artefact |
|---|---|---|
| #1 route sau JWT gateway | AC #1 | `router.go` đăng ký GET /v1/compare |
| #2 require canonical_key (400) | AC #2 | `compare.go::Compare` guard + `TestCompare_MissingKey_400` |
| #3 JOIN theo canonical_key | AC #3 | `compare_query.go` JOIN tracked_product + platform |
| #4 latest-per-product DISTINCT ON | AC #4 | `compare_query.go` DISTINCT ON (product_id) ts DESC |
| #5 hình dạng dòng | AC #5 | `CompareRow`/`CompareItem` + `TestCompare_PayloadShape` |
| #6 is_cheapest server-side | AC #6 | `markCheapest` + `TestCompare_ThreePlatforms_HighlightsCheapest` |
| #7 price BIGINT VND | AC #7 | `Price int64` |
| #8 ts mỗi sàn | AC #8 | `CompareItem.TS` (RFC3339) + `TestCompare_PayloadShape` |
| #9 single-platform | AC #9 | `TestCompare_SinglePlatform` |
| #10 key lạ -> 404 | AC #10 | nhánh `len(rows)==0` + `TestCompare_UnknownKey_Empty` |
| #11 p95 <500ms | AC #11 | index idx_tp_canonical + ts hypertable; `compare_query_duration_ms` |
| #12 OTel metric (SHOULD) | AC #12 | `compare_request_total{result}` |

## §4 - Kết luận

Mọi mệnh đề normative có code/SQL/test backing; không mệnh đề mồ côi. API shape khớp §3.7; lấy giá hiện tại đúng nguồn (`price_snapshot` latest-per-product); xử lý đúng việc thiếu cột `name` của `platform`; highlight và độ tươi xử lý phía server. Frontmatter đủ, không "TBD"/"TODO". Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập TASK-PRICE-004.*
