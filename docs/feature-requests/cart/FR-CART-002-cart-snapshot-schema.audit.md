---
fr_id: FR-CART-002
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

FR-CART-002 đặc tả schema `cart_snapshot` + `cart_item` và API nhận giỏ ở mức triển khai được. 12 mệnh đề §1 normative neo vào AC §4 và test §5. Ba ranh giới niềm tin/an toàn được test khẳng định: không cột credential trong schema, user_id từ JWT chứ không từ payload, và payload chứa cookie/token không lọt vào DB. Tiền tệ BIGINT, ghi transaction + idempotent theo snapshot_ref, và cho ghi SKU chưa track giữ optimizer có đầu vào. Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - Rò rỉ credential qua schema giỏ (đã chốt)
Rủi ro: schema lỡ có cột cookie/token -> rò rỉ là thảm họa PDPL, phá định vị niềm tin. Giải: DEC-CART-09 + §1 #4/#6 KHÔNG cột credential + DTO không trường credential; AC #2 + #5 + test RejectsCredentialFields khẳng định DB không chứa giá trị đó.

### ISS-002 - Giả mạo chủ sở hữu giỏ
user_id từ payload extension (chạy trên máy user) có thể bị giả mạo. Giải: DEC-CART-10 + §1 #5 lấy từ JWT; AC #4 + test UserIDFromJWT_NotPayload khẳng định payload user_id=99999 bị bỏ qua.

### ISS-003 - Chặn ghi khi SKU chưa track
FK cứng làm mất khả năng tối ưu giỏ có SKU mới. Giải: DEC-CART-12 + §1 #9 product_id NULL + platform_item_id + CHECK item_identified; AC #8 + test UntrackedSku_StillWrites.

### ISS-004 - Retry tạo giỏ trùng
Extension retry gửi cùng giỏ hai lần. Giải: DEC-CART-11 + §1 #8 UNIQUE(user_id, snapshot_ref) + ghi transaction; AC #6 + #7 + test Idempotent.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 cart_snapshot | #1 | `0011_cart_snapshot.up.sql` |
| #2 cart_item | #1 | bảng cart_item |
| #3 BIGINT qty/price | #3 | CHECK + types int64 |
| #4 không cột credential | #2 | schema + DTO |
| #5 user_id từ JWT | #4 | `snapshot.go` handler |
| #6 từ chối field credential | #5 | DTO + test |
| #7 transaction | #6 | `InsertSnapshot` |
| #8 idempotent | #7 | UNIQUE(user_id, snapshot_ref) |
| #9 ghi SKU chưa track | #8 | CHECK item_identified |
| #10 scope user đọc | #9 | `GetSnapshot` 404 |
| #11 bất biến | #10 | không update path |

## §4 - Kết luận

Mọi mệnh đề normative có SQL/code/test backing; ba ranh giới niềm tin (không credential, JWT chủ sở hữu, DB sạch credential) kiểm chứng bằng test. Không mệnh đề "mồ côi". **Score = 10/10. Verdict: PASS.** Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit FR-CART-002.*
