---
fr_id: TASK-TRACK-002
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt

Audit độc lập, tái diễn từ file TASK-TRACK-002 hiện tại. task đặc tả schema + API CRUD wishlist / wishlist_item với target_price BIGINT VND nullable và kiểm chủ sở hữu mọi route. 12 mệnh đề §1 BCP-14 (11 MUST + 1 SHOULD), testable. Schema khớp DATA-MODEL: wishlist_item có FK wishlist_id ON DELETE CASCADE + product_id -> tracked_product, target_price BIGINT CHECK > 0, UNIQUE(wishlist_id, product_id). Quyết định bảo mật 404-không-403 chống IDOR trên khóa BIGSERIAL tuần tự. Đạt 10/10.

## §2 - Findings

Không còn khiếm khuyết tồn dư. Kiểm độc lập:
- target_price BIGINT VND (§1 #3, DEC-TRACK-14) - KHÔNG float; CHECK + JSON int64 có test TestTargetPrice_Int64InJSON.
- Kiểm chủ sở hữu trả 404 (§1 #9, DEC-TRACK-12) thay vì 403 đóng lỗ lộ tồn tại + liệt kê id; có TestCrossUser_404.
- Thêm item idempotent ON CONFLICT DO UPDATE (§1 #7) - bấm lại cập nhật giá, không nhân đôi.
- FK product_id -> tracked_product (§1 #2) chặn id rác, map 23503 -> 400.
- §10 failure-modes 9 hàng không tầm thường (race, user_id giả từ body, 403 vô tình lộ).
- Typography prose plain ASCII + tiếng Việt có dấu; không tự cấm; sentinel có mặt.

## §3 - Bảng truy vết (từ file hiện tại)

| §1 mệnh đề | AC | Test/Artefact |
|---|---|---|
| #1 bảng wishlist | #1 | 0002_wishlist.sql |
| #2 wishlist_item + FK CASCADE | #7 | TestDeleteWishlist_Cascade |
| #3 target_price BIGINT nullable | #3,#9 | CHECK + TestTargetPrice_Int64InJSON |
| #4 UNIQUE(wishlist_id, product_id) | #10 | UNIQUE + ON CONFLICT |
| #5 POST tạo wishlist | #1 | HandleCreate |
| #6 GET scope user | #2 | TestList_ScopedToUser |
| #7 add item idempotent + FK | #4,#5 | TestAddItem_Idempotent + TestAddItem_UnknownProduct_FK |
| #8 DELETE item/wishlist | #7,#8 | CASCADE |
| #9 kiểm chủ sở hữu 404 | #6 | TestCrossUser_404 |
| #11 JSON int64 | #9 | TestTargetPrice_Int64InJSON |
| #12 OTel | #12 | wishlist_ops_total |

## §4 - Kết luận

Mọi mệnh đề normative có code/SQL/test backing; không mệnh đề mồ côi. Lỗ IDOR (rủi ro lớn nhất, dùng PDPL) được đóng bằng 404 + test cross-user. Không cần sửa. Score = 10/10. Verdict: PASS. Sẵn sàng build.

---

*Hết audit TASK-TRACK-002.*
