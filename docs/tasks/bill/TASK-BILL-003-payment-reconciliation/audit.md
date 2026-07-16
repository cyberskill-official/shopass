---
fr_id: TASK-BILL-003
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

TASK-BILL-003 đặc tả bảng `payment` + IPN webhook + reconciliation ở mức triển khai được. 12 mệnh đề §1 normative, mỗi mệnh đề có AC và test trong §5. Năm trụ chốt tiền được khóa: xác thực chữ ký IPN (khóa Vault), khớp số tiền IPN với payment.amount (lệch -> mismatch, không cấp Premium), idempotent theo order_ref+transaction_id, kích hoạt subscription chỉ khi paid, reconciliation định kỳ bù IPN mất. Lưu fee để tính doanh thu ròng. Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - IPN giả mạo
Webhook công khai; tin IPN chưa xác thực = Premium miễn phí qua IPN giả. Giải: §1 #5 + DEC-BILL-12 - VerifyIPN bằng khóa Vault trước khi đổi payment; AC #2 + `TestIPN_BadSignature_NoChange`.

### ISS-002 - Trả ít được nhiều
IPN hợp lệ về chữ ký nhưng số tiền lệch vẫn cấp Premium. Giải: §1 #7 + DEC-BILL-14 - khớp `amount`, lệch -> `mismatch`, không kích hoạt subscription; AC #4 + `TestIPN_AmountMismatch_NoActivate`.

### ISS-003 - IPN lặp + IPN mất
Cổng retry gia hạn dư; IPN mất làm user kẹt pending. Giải: §1 #6 (idempotent order_ref+tx) + §1 #10 (reconciliation bù); AC #6/#10 + `TestIPN_Duplicate_Idempotent` + `TestReconcile_PaidBackfill`.

### ISS-004 - Biên sai do thiếu fee
Không lưu fee chỉ biết doanh thu gộp. Giải: §1 #2 - `fee` BIGINT, doanh thu ròng `amount - fee` (phí QR ~1,5-2,5% §4.1); AC #11.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 schema payment | #1 | `0003_payment.sql` |
| #2 amount/fee BIGINT | #8,#11 | CHECK + net metric |
| #3 status CHECK | #9 | CHECK status IN (...) |
| #5 verify chữ ký | #2 | `VerifyIPN` + `TestIPN_BadSignature_NoChange` |
| #6 idempotent | #6 | order_ref UNIQUE + `TestIPN_Duplicate_Idempotent` |
| #7 khớp amount | #4 | `TestIPN_AmountMismatch_NoActivate` |
| #8 kích hoạt khi paid | #3 | `subs.Activate` + `TestIPN_Paid_ActivatesSubscription` |
| #10 reconciliation | #10 | `reconcile.go` + `TestReconcile_PaidBackfill` |

## §4 - Kết luận

Toàn bộ mệnh đề normative có DDL/code/test backing, gồm verify-IPN, khớp-số-tiền, idempotent, kích-hoạt-khi-paid và reconciliation-bù. Không có mệnh đề "mồ côi". Score = 10/10. Verdict: PASS. Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit TASK-BILL-003.*
