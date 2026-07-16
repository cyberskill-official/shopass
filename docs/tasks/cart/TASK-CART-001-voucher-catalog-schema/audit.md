---
fr_id: TASK-CART-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

TASK-CART-001 đặc tả schema + ingest `voucher_catalog` ở mức triển khai được. 12 mệnh đề §1 normative neo vào AC §4 và test §5. Ba loại voucher (shop/platform/freeship), tiền tệ BIGINT VND, ràng buộc `shop_id` theo loại, và `stack_group` (nhãn cho luật per-country TASK-CART-004) đều có CHECK + test. `ListActive` lọc cửa sổ hiệu lực + shop tách lưu trữ khỏi tính toán, cấp đúng tập ứng viên cho optimizer. Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - Sai số tiền tệ trong optimizer (đã chốt)
discount_value/min_spend/cap dạng float làm optimizer cộng nhiều voucher sai số, vượt cap. Giải: DEC-CART-02/03 + §1 #3/#4 BIGINT VND + percent nguyên; AC #5 + test BigintRoundTrip.

### ISS-002 - Voucher shop mồ côi
shop voucher thiếu shop_id thì không biết áp đâu. Giải: DEC-CART-06 + §1 #5 CHECK `shop_id_by_type`; AC #3 + test RejectShopVoucherWithoutShopID.

### ISS-003 - Voucher quá hạn lọt vào optimizer
Tính combo dựa voucher không dùng được lúc thanh toán. Giải: DEC-CART-05 + §1 #8 `ListActive` lọc `[valid_from, valid_to]` theo now; AC #6 + test FiltersByWindow.

### ISS-004 - stack_group nền cho luật per-country
Không có nhãn loại trừ thì TASK-CART-004 không quyết được voucher nào xung khắc. Giải: DEC-CART-04 + §1 #6 `stack_group`; AC #10; là input cho TASK-CART-004.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 schema | #1 | `0010_voucher_catalog.up.sql` |
| #2 type enum | #2 | CHECK type |
| #3 discount_type + percent | #2,#4 | CHECK discount_type/value |
| #4 BIGINT min_spend/cap | #5 | types.go int64 + test |
| #5 shop_id theo loại | #3 | CHECK shop_id_by_type |
| #6 stack_group | #10 | cột + TASK-CART-004 |
| #7 cửa sổ hiệu lực | #6 | valid_from/to |
| #8 ListActive lọc | #6,#7 | `repo.go::ListActive` |
| #9 ingest validate | #8 | `ingest.go::validate` |
| #10 idempotent | #9 | ON CONFLICT DO UPDATE |
| #11 index | - | `idx_vc_active` |
| #12 CHECK value | #4 | CHECK discount_value |

## §4 - Kết luận

Mọi mệnh đề normative có SQL/code/test backing; tiền tệ BIGINT, ràng buộc loại, lọc hiệu lực, và nhãn stacking đều kiểm chứng được. Không mệnh đề "mồ côi". **Score = 10/10. Verdict: PASS.** Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit TASK-CART-001.*
