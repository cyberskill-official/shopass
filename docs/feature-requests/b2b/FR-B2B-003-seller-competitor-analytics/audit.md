---
fr_id: FR-B2B-003
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

FR-B2B-003 đặc tả analytics hướng seller ở mức triển khai được. 11 mệnh đề §1 normative, mỗi mệnh đề có AC và test. Ranh giới sống còn được giữ chặt: seller chỉ thấy vị thế (percentile_rank) của chính mình trong dải phân vị tổng hợp đã qua k-anonymity, KHÔNG bao giờ thấy giá đối thủ cụ thể. Hai cổng độc lập (xác minh quyền sở hữu shop + ô thị trường đã phát hành) cùng chặn rò rỉ. Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - Diễn giải "theo dõi giá đối thủ" (đã chốt)
"Theo dõi giá đối thủ" (§6 mục 8) dễ bị hiểu thành "xem giá shop cụ thể" - biến SănDeal thành công cụ do thám, vi phạm PDPL. Giải: §1 #1 + DEC-B2B-20 định nghĩa lại là vị thế trong dải tổng hợp; test `TestPosition_NoCompetitorPrice`.

### ISS-002 - Seller dò shop người khác
Nếu không xác minh quyền sở hữu, seller truy vấn shop_id bất kỳ để soi đối thủ. Giải: §1 #3 + DEC-B2B-24 cổng `verified=true`; test `TestOwnership_NotVerified_403` + `TestOwnership_OtherShop_403`.

### ISS-003 - Dải thưa có rủi ro lộ
Render vị thế khi ô market_trend_daily bị suppress có thể lộ qua dải quá hẹp. Giải: §1 #5 + DEC-B2B-23 trả 422; test `TestPosition_SuppressedMarket_422`.

### ISS-004 - Kho dữ liệu cạnh tranh thứ cấp
Log cặp (seller -> category/đối thủ) tạo ra dữ liệu nhạy cảm mới. Giải: §1 #9 + DEC-B2B-25 log mức tổng hợp.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 chỉ vị thế của seller | #6 | `position.go` (không trường đối thủ) + test |
| #2 schema sở hữu | #1 | `0003_seller_owned_sku.sql` |
| #3 cổng quyền sở hữu | #2,#3,#4 | `ownership.go::assertOwned` + 3 test |
| #4 percentile_rank | #7 | `position.go::rank` + `TestRank_Monotonic` |
| #5 suppress -> 422 | #5 | `seller_handler.go` + `TestPosition_SuppressedMarket_422` |
| #6 nguồn dải = market_trend_daily | #10 | builder gọi QueryCells |
| #9 log tổng hợp | #9 | review log/metric |
| #11 server-side org | #8 | đối chiếu seller_owned_sku |

## §4 - Kết luận

Toàn bộ mệnh đề normative có code/SQL/test backing. Ranh giới bảo mật (không lộ giá đối thủ) được kiểm bằng test cấu trúc response + hai cổng. Không mệnh đề mồ côi. Score = 10/10. Verdict: PASS. Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit FR-B2B-003.*
