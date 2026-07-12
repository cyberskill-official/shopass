---
fr_id: FR-TRACK-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt

Audit độc lập, tái diễn từ file FR-TRACK-001 hiện tại (không dựa vào audit đồng-tác-giả cũ). FR đặc tả `POST /v1/track`: parse item_url theo sàn -> Upsert tracked_product (FR-PRICE-001) -> link user (user_tracked_product) -> enqueue scrape job priming. 12 mệnh đề §1 BCP-14 (11 MUST + 1 SHOULD), mỗi mệnh đề testable. Schema user_tracked_product khớp DATA-MODEL (PK user_id+product_id, cả hai FK). Ranh giới sạch: không INSERT thẳng tracked_product, không tự gửi scrape. Đạt 10/10.

## §2 - Findings

Không còn khiếm khuyết tồn dư. Các điểm đã kiểm độc lập:
- Parse url phía server (§1 #3, DEC-TRACK-01) đóng bề mặt giả mạo platform_item_id; client chỉ đưa link.
- Upsert idempotent (§1 #5) + UNIQUE(user_id, product_id) (§1 #6) chống nhân bản SKU lẫn link.
- §1 #10 "enqueue lỗi vẫn 201" tách hành động người dùng khỏi tối ưu trải nghiệm - có AC #9 + test backing.
- Tiền (nếu trả kèm) là BIGINT VND (§1 #11), đồng nhất DEC-PRICE-05.
- §10 failure-modes 9 hàng không tầm thường (race, short link, platform chưa seed).
- Typography prose plain ASCII + tiếng Việt có dấu; không tự cấm; sentinel "Hết FR-TRACK-001" có mặt.

## §3 - Bảng truy vết (từ file hiện tại)

| §1 mệnh đề | AC | Test/Artefact |
|---|---|---|
| #1 route + body JSON | #1 | track.go HandleTrack |
| #2 platform allowlist | #3 | IDByCode + TestTrack_UnsupportedPlatform_400 |
| #3 parse url server-side | #2,#4 | url_parser.go + 4 parser tests |
| #4 map code -> platform_id | #11 | platforms.IDByCode |
| #5 Upsert không nhân bản | #7 | price.Upsert (FR-PRICE-001) |
| #6 bảng nối + UNIQUE | #5,#6 | 0001_user_tracked_product.sql |
| #8 enqueue priming | #8 | scrapeQueue.EnqueuePriming + TestTrack_NewProduct_201 |
| #9 already_tracked | #5 | TestTrack_Idempotent_SameUser |
| #10 enqueue lỗi vẫn 201 | #9 | TestTrack_EnqueueFails_Still201 |
| #11 Content-Type + BIGINT | - | header set + int64 |
| #12 OTel metrics | #12 | track_new_product_total |

## §4 - Kết luận

Mọi mệnh đề normative có code/SQL/test backing; không có mệnh đề mồ côi. Hợp đồng khớp DATA-MODEL + §3.4 + §3.7 nguồn. Không cần sửa. Score = 10/10. Verdict: PASS. Sẵn sàng build (status ready_to_implement).

---

*Hết audit FR-TRACK-001.*
