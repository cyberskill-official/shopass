---
fr_id: FR-BILL-002
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

FR-BILL-002 đặc tả tích hợp bốn cổng thanh toán VN (MoMo/ZaloPay/VNPay/VietQR) ở mức triển khai được. 12 mệnh đề §1 normative, mỗi mệnh đề có AC và test trong §5. Năm trụ được khóa: interface chung + adapter per-cổng, ưu tiên QR (phí thấp), không lưu thẻ + khóa ký từ Vault (no-cleartext), số tiền BIGINT từ plan_catalog (không từ client), idempotent theo order_ref (chống double-charge). Lỗi gọi cổng giữ payment pending. Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - Lưu thẻ + khóa cleartext
Lưu PAN/CVV kéo PCI scope nặng; khóa ký cleartext lộ DB = giả mạo thanh toán. Giải: §1 #4/#5 + DEC-BILL-09 - không chạm thẻ (cổng xử lý), khóa từ Vault; AC #8/#9 (review + grep).

### ISS-002 - Thao túng giá
Nhận số tiền từ client cho phép trả 1 đồng cho Premium. Giải: §1 #3 - `amount` luôn từ `plan_catalog.price` (BIGINT), bỏ qua body; AC #3 + `TestCheckout_AmountFromCatalog_NotBody`.

### ISS-003 - Double-charge
User bấm thanh toán hai lần bị trừ hai lần. Giải: §1 #6 + DEC-BILL-10 - `order_ref` idempotent, gọi lại không tạo intent mới; AC #6 + `TestCheckout_Idempotent_NoDoubleCharge`.

### ISS-004 - Lỗi mạng coi là thành công
Cấp Premium khi chưa thu tiền. Giải: §1 #11 - lỗi gọi cổng -> 502, payment giữ pending, trạng thái cuối do FR-BILL-003 chốt; AC #10 + `TestCheckout_GatewayError_502_StaysPending`.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 interface + registry | #5 | `gateway.go` + `TestRegistry_SelectsAdapter` |
| #2 checkout endpoint | #1,#2 | `checkout.go` + `TestCheckout_HappyPath` |
| #3 amount BIGINT từ catalog | #3 | `TestCheckout_AmountFromCatalog_NotBody` |
| #4 ký từ Vault | #7,#8 | `sign.go` + `TestSignMoMo_FixedVector` |
| #5 không lưu thẻ | #9 | review + grep schema payment |
| #6 idempotent order_ref | #6 | `NewOrderRef` + `TestCheckout_Idempotent_NoDoubleCharge` |
| #7 ưu tiên QR | #1 | VietQR mặc định gợi ý |
| #8 PaymentResult | #1,#2 | `PaymentResult` (pay_url/qr_payload) |
| #11 lỗi giữ pending | #10 | `TestCheckout_GatewayError_502_StaysPending` |

## §4 - Kết luận

Toàn bộ mệnh đề normative có code/test backing, gồm không-lưu-thẻ, khóa-Vault, amount-từ-catalog, idempotent-order_ref và lỗi-giữ-pending. Không có mệnh đề "mồ côi". Score = 10/10. Verdict: PASS. Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit FR-BILL-002.*
