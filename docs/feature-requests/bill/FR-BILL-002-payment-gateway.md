---
id: FR-BILL-002
title: "Tích hợp cổng thanh toán (MoMo / ZaloPay / VNPay / VietQR) - tạo yêu cầu thanh toán theo gateway, ưu tiên QR rẻ hơn thẻ (~1,5-2,5%/giao dịch); adapter chung + ký request, không lưu thông tin thẻ"
module: BILL
priority: MUST
status: done
verify: T
phase: P2
milestone: P2 - slice 2
slice: 2
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-BILL-001, FR-BILL-003, FR-INFRA-003, FR-INFRA-001, FR-COMPLY-005]
depends_on: [FR-BILL-001]
blocks: [FR-AFFIL-005, FR-BILL-003, FR-TRUST-005]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §4.1 (phí thanh toán MoMo/ZaloPay/VNPay: QR rẻ hơn thẻ ~1,5-2,5%/giao dịch)"
  - "docs/... §4.3 (payment rails: MoMo 31tr+ user, ZaloPay, VNPay, bank QR VietQR)"
source_decisions:
  - "DEC-BILL-06: một interface PaymentGateway chung (CreatePayment) + adapter per-cổng (momo/zalopay/vnpay/vietqr); router chọn gateway theo yêu cầu user"
  - "DEC-BILL-07: ưu tiên VietQR/QR (phí ~1,5-2,5%) làm mặc định gợi ý; thẻ chỉ khi user chọn (phí cao hơn)"
  - "DEC-BILL-08: số tiền gửi gateway là BIGINT VND (đồng nhất plan_catalog DEC-BILL-02); KHÔNG float"
  - "DEC-BILL-09: secret/khóa ký mỗi cổng đọc từ Vault (FR-INFRA-003); KHÔNG cleartext; SănDeal KHÔNG lưu thông tin thẻ (PCI scope tối thiểu)"
  - "DEC-BILL-10: mỗi CreatePayment sinh một order_ref idempotency duy nhất; gọi lại cùng order_ref không tạo hai giao dịch (chống double-charge khi user bấm lại)"

language: "Go 1.22 (bill-svc); PostgreSQL 16; HTTP client tới cổng thanh toán"
service: shopass/services/bill/
new_files:
  - services/bill/internal/pay/gateway.go
  - services/bill/internal/pay/momo.go
  - services/bill/internal/pay/zalopay.go
  - services/bill/internal/pay/vnpay.go
  - services/bill/internal/pay/vietqr.go
  - services/bill/internal/pay/sign.go
  - services/bill/internal/api/checkout.go
  - services/bill/internal/pay/gateway_test.go
  - services/bill/internal/pay/sign_test.go
  - services/bill/internal/api/checkout_test.go
modified_files:
  - services/bill/internal/api/router.go            # đăng ký POST /v1/billing/checkout
allowed_tools:
  - file_read: services/bill/**
  - file_write: services/bill/**
  - bash: cd services/bill && go test ./...
disallowed_tools:
  - lưu thông tin thẻ (PAN/CVV) ở bất kỳ đâu (vi phạm DEC-BILL-09, PCI)
  - lưu khóa ký cổng cleartext trong DB/code (vi phạm DEC-BILL-09, no-cleartext §5.5)
  - gửi số tiền dạng float (vi phạm DEC-BILL-08)
  - cho phép double-charge khi gọi lại cùng order_ref (vi phạm DEC-BILL-10)

effort_hours: 10
sub_tasks:
  - "1.0h: gateway.go - interface PaymentGateway + struct PaymentRequest/PaymentResult + registry chọn theo code"
  - "1.0h: sign.go - ký request per-cổng (HMAC/checksum theo spec từng cổng) với khóa từ Vault"
  - "1.5h: momo.go + zalopay.go - adapter tạo yêu cầu thanh toán + parse phản hồi (deeplink/QR)"
  - "1.5h: vnpay.go + vietqr.go - adapter VNPay (redirect URL ký) + VietQR (sinh chuỗi QR)"
  - "1.5h: checkout.go - handler POST /v1/billing/checkout: chọn plan, sinh order_ref idempotent, gọi gateway, trả pay_url/qr"
  - "1.0h: gateway_test.go - chọn đúng adapter; QR là mặc định gợi ý; số tiền BIGINT; order_ref idempotent"
  - "1.0h: sign_test.go - ký đúng theo từng cổng (vector cố định); khóa từ Vault không cleartext"
  - "1.0h: checkout_test.go - happy path tạo thanh toán; gọi lại cùng order_ref không tạo 2; plan lạ 400"
  - "0.5h: OTel metric payment_intent_created_total{gateway} + payment_intent_amount_vnd"

risk_if_skipped: "Không có cổng thanh toán thì Premium (FR-BILL-001) không thu được tiền - dòng doanh thu Premium đứng hoàn toàn (§4.3). Nếu lưu thông tin thẻ thì SănDeal rơi vào phạm vi PCI nặng và rủi ro lộ thẻ - thảm họa pháp lý + niềm tin; mô hình đúng là không chạm thẻ, để cổng xử lý. Nếu lưu khóa ký cleartext thì lộ DB là lộ khả năng giả mạo yêu cầu thanh toán (vi phạm no-cleartext §5.5). Nếu gửi số tiền float thì lệch tiền với cổng. Nếu không idempotent theo order_ref thì user bấm thanh toán hai lần bị trừ tiền hai lần - khiếu nại và hoàn tiền tốn kém. Nếu không ưu tiên QR thì SănDeal chịu phí thẻ cao hơn (§4.1) ăn vào biên vốn đã mỏng của thị trường WTP thấp. Đây là cửa thu tiền của toàn hệ Premium."
---

## §1 - Mô tả (BCP-14 normative)

Service BILL **MUST** tích hợp bốn cổng thanh toán VN (MoMo, ZaloPay, VNPay, VietQR) qua một interface chung, tạo yêu cầu thanh toán theo cổng user chọn, ưu tiên QR (phí thấp), ký request bằng khóa từ Vault, và idempotent theo `order_ref`. SănDeal **MUST NOT** lưu thông tin thẻ. Hợp đồng:

1. **MUST** định nghĩa interface `PaymentGateway` với `CreatePayment(ctx, PaymentRequest) (PaymentResult, error)` và một registry chọn adapter theo `gateway` code - {`'momo'`,`'zalopay'`,`'vnpay'`,`'vietqr'`} (DEC-BILL-06).
2. **MUST** phục vụ `POST /v1/billing/checkout {plan_tier, gateway}`: tra `plan_catalog` (FR-BILL-001) lấy `price` (BIGINT VND), sinh một `order_ref` idempotency, gọi adapter, trả `PaymentResult` (pay_url hoặc qr_payload tùy cổng).
3. **MUST** gửi số tiền cho cổng dạng `BIGINT` VND (DEC-BILL-08) - KHÔNG float; lấy từ `plan_catalog.price`, KHÔNG nhận số tiền do client gửi (chống thao túng giá).
4. **MUST** ký mọi request gửi cổng theo đúng spec ký của cổng đó (HMAC/checksum), với khóa bí mật đọc từ Vault (FR-INFRA-003) (DEC-BILL-09); khóa KHÔNG nằm trong DB/code cleartext.
5. SănDeal **MUST NOT** lưu, log, hay truyền thông tin thẻ (PAN, CVV, ngày hết hạn) ở bất kỳ đâu (DEC-BILL-09): luồng thẻ (nếu có qua VNPay) do cổng xử lý trên trang cổng; SănDeal chỉ giữ `order_ref` + kết quả.
6. **MUST** sinh `order_ref` duy nhất cho mỗi checkout và idempotent theo nó (DEC-BILL-10): gọi `CreatePayment` lại với cùng `order_ref` (user bấm lại) **MUST NOT** tạo hai giao dịch ở cổng - trả lại intent đã tạo hoặc từ chối trùng.
7. **MUST** ưu tiên QR (VietQR/QR động) làm gợi ý mặc định ở tầng gợi ý (DEC-BILL-07) vì phí thấp hơn thẻ (~1,5-2,5%/giao dịch, §4.1); thẻ chỉ khi user chủ động chọn. (Đây là mặc định gợi ý, user vẫn chọn được cổng khác.)
8. `PaymentResult` **MUST** chứa: `order_ref`, `gateway`, `amount` (BIGINT VND), và một trong `pay_url` (MoMo/ZaloPay/VNPay deeplink/redirect) hoặc `qr_payload` (VietQR chuỗi QR) để client hiển thị.
9. **MUST** ghi một bản ghi `payment` trạng thái khởi tạo (`pending`) gắn `order_ref` + `subscription`/user trước khi trả về (để FR-BILL-003 đối soát khi IPN/webhook về). Tiền `amount` BIGINT VND.
10. **MUST** lấy `user_id` từ JWT (FR-INFRA-001); plan_tier không tồn tại trong `plan_catalog` -> `400`; `gateway` ngoài tập -> `400`.
11. **MUST** đặt timeout + xử lý lỗi gọi cổng: cổng không phản hồi/đỗ -> trả `502`/`503` và để `payment` ở `pending` (không đánh dấu thành công khi chưa chắc); KHÔNG coi lỗi mạng là thanh toán thành công.
12. **SHOULD** phát OTel: `payment_intent_created_total{gateway}` (counter), `payment_intent_amount_vnd` (histogram), `payment_gateway_error_total{gateway}` (counter).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao một interface chung + adapter per-cổng (DEC-BILL-06)?** Bốn cổng có API ký và định dạng phản hồi khác nhau (MoMo deeplink, ZaloPay order, VNPay redirect ký, VietQR chuỗi QR). Một interface `PaymentGateway` chung để phần checkout không phải biết chi tiết từng cổng; mỗi adapter giấu sự khác biệt. Thêm cổng mới là thêm một adapter, không sửa handler.

**Vì sao ưu tiên QR (DEC-BILL-07)?** Phí giao dịch ăn trực tiếp vào biên. QR/VietQR rẻ hơn thẻ (~1,5-2,5% so với thẻ cao hơn, §4.1). Ở thị trường VN với willingness-to-pay thấp và giá Premium nhẹ (29k-79k), mỗi điểm phần trăm phí là đáng kể. Gợi ý QR mặc định hướng người dùng tới phương thức rẻ mà vẫn để họ tự chọn - tối ưu chi phí không ép buộc.

**Vì sao không lưu thông tin thẻ (DEC-BILL-09)?** Lưu PAN/CVV kéo SănDeal vào phạm vi PCI-DSS nặng và biến DB thành mục tiêu giá trị cao. Mô hình đúng cho một startup là không bao giờ chạm thẻ: luồng thẻ xảy ra trên trang cổng (VNPay), SănDeal chỉ giữ `order_ref` và kết quả. Phạm vi PCI thu về tối thiểu, rủi ro lộ thẻ về gần không.

**Vì sao khóa ký từ Vault (DEC-BILL-09, §1 #4)?** Khóa ký là thứ chứng minh request đến từ SănDeal. Lộ khóa là lộ khả năng giả mạo yêu cầu thanh toán/hoàn tiền. Đọc khóa từ Vault (không để trong DB/code cleartext) tuân no-cleartext của PDPL (§5.5) và thu hẹp bề mặt lộ - đồng nhất cách FR-AFFIL-003 xử lý postback secret.

**Vì sao idempotent theo order_ref (DEC-BILL-10)?** Người dùng bấm "thanh toán" hai lần, hoặc retry khi mạng chậm, là chuyện thường. Không idempotent thì hai giao dịch tạo ra, user bị trừ tiền hai lần - khiếu nại và hoàn tiền tốn kém, mất niềm tin. Một `order_ref` duy nhất cho mỗi ý định thanh toán, gọi lại cùng ref không tạo giao dịch mới, đóng cửa double-charge.

**Vì sao lỗi mạng không phải thành công (§1 #11)?** Khi gọi cổng đỗ/timeout, ta KHÔNG biết thanh toán có thành công hay không. Coi lỗi là thất bại (giữ `pending`) an toàn hơn coi là thành công (cấp Premium khi chưa thu tiền). Trạng thái thật được FR-BILL-003 chốt qua IPN/webhook của cổng - nguồn sự thật về thanh toán là cổng, không phải kết quả lời gọi của ta.

---

## §3 - Hợp đồng API / DDL

### Interface + registry (Go)

```go
// services/bill/internal/pay/gateway.go
type PaymentRequest struct {
    OrderRef string // idempotency key (§1 #6)
    Amount   int64  // VND, BIGINT (§1 #3)
    UserID   int64
    PlanTier string
}
type PaymentResult struct {
    OrderRef  string `json:"order_ref"`
    Gateway   string `json:"gateway"`
    Amount    int64  `json:"amount"`           // VND
    PayURL    string `json:"pay_url,omitempty"`    // MoMo/ZaloPay/VNPay
    QRPayload string `json:"qr_payload,omitempty"` // VietQR
}

type PaymentGateway interface {
    Code() string
    CreatePayment(ctx context.Context, r PaymentRequest) (PaymentResult, error)
}

// Registry chọn adapter theo code (§1 #1).
func (reg *Registry) Get(code string) (PaymentGateway, bool) {
    g, ok := reg.byCode[code] // momo|zalopay|vnpay|vietqr
    return g, ok
}
```

### Ký request (Go)

```go
// services/bill/internal/pay/sign.go
// SignMoMo ký payload theo spec MoMo (HMAC-SHA256) với khóa từ Vault (§1 #4).
func SignMoMo(ctx context.Context, secrets SecretReader, raw string) (string, error) {
    key, err := secrets.Get(ctx, "bill/momo/secret_key") // Vault (FR-INFRA-003)
    if err != nil {
        return "", err
    }
    mac := hmac.New(sha256.New, []byte(key))
    mac.Write([]byte(raw))
    return hex.EncodeToString(mac.Sum(nil)), nil
}
```

### Checkout handler (Go)

```go
// services/bill/internal/api/checkout.go
func (h *Handler) HandleCheckout(w http.ResponseWriter, req *http.Request) {
    userID := auth.UserID(req.Context()) // FR-INFRA-001
    var body struct{ PlanTier, Gateway string }
    if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
        writeErr(w, 400, "invalid body"); return
    }
    plan, ok := h.plans.ByTier(req.Context(), body.PlanTier) // FR-BILL-001
    if !ok { writeErr(w, 400, "unknown plan"); return }
    gw, ok := h.gateways.Get(body.Gateway)
    if !ok { writeErr(w, 400, "unsupported gateway"); return }

    orderRef := pay.NewOrderRef(userID, plan.Tier) // duy nhất + idempotent (§1 #6)
    if existing, found := h.payments.ByOrderRef(req.Context(), orderRef); found {
        _ = json.NewEncoder(w).Encode(existing.Result()); return // không tạo lần hai
    }
    res, err := gw.CreatePayment(req.Context(), pay.PaymentRequest{
        OrderRef: orderRef, Amount: plan.Price, UserID: userID, PlanTier: plan.Tier, // BIGINT VND
    })
    if err != nil {
        metrics.GatewayError(body.Gateway)
        writeErr(w, 502, "payment gateway error"); return // KHÔNG coi là thành công (§1 #11)
    }
    h.payments.InsertPending(req.Context(), orderRef, userID, plan.Price, body.Gateway) // §1 #9
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    _ = json.NewEncoder(w).Encode(res)
    metrics.IntentCreated(body.Gateway)
}
```

---

## §4 - Acceptance criteria

1. `POST /v1/billing/checkout {plan_tier:"premium_basic", gateway:"vietqr"}` -> `200` với `PaymentResult` chứa `qr_payload`, `amount=29000`.
2. `gateway:"momo"` -> `PaymentResult` chứa `pay_url` (deeplink MoMo); `amount` đúng từ `plan_catalog`.
3. `amount` gửi cổng luôn lấy từ `plan_catalog.price` (BIGINT), KHÔNG từ body client (gửi kèm `amount` trong body bị bỏ qua).
4. `plan_tier` không có trong `plan_catalog` -> `400` "unknown plan".
5. `gateway` ngoài {momo,zalopay,vnpay,vietqr} -> `400` "unsupported gateway".
6. Gọi checkout hai lần cùng `order_ref` (cùng user+plan trong cửa sổ idempotent) -> chỉ một intent tạo ở cổng; lần hai trả lại intent cũ.
7. Ký request: với khóa cố định (test) + payload cố định -> chữ ký khớp vector kỳ vọng của từng cổng.
8. Khóa ký đọc từ Vault (mock SecretReader); review xác nhận không có khóa cleartext trong code/DB.
9. Không có cột/đường lưu thông tin thẻ (PAN/CVV) trong schema `payment` (review + grep).
10. Cổng trả lỗi/timeout (mock fail) -> `502`, `payment` giữ `pending` (KHÔNG đánh dấu thành công).
11. Một bản ghi `payment` `pending` được tạo với `order_ref` + `amount` BIGINT trước khi trả response.
12. Metric `payment_intent_created_total{gateway}` tăng khi tạo; `payment_gateway_error_total` tăng khi cổng lỗi.

---

## §5 - Kiểm thử (verification)

```go
// services/bill/internal/pay/gateway_test.go
func TestRegistry_SelectsAdapter(t *testing.T) {
    reg := newRegistry(t)
    for _, code := range []string{"momo", "zalopay", "vnpay", "vietqr"} {
        g, ok := reg.Get(code)
        require.True(t, ok)
        require.Equal(t, code, g.Code())
    }
    _, ok := reg.Get("paypal")
    require.False(t, ok)
}

func TestVietQR_ReturnsQRPayload(t *testing.T) {
    g, _ := newRegistry(t).Get("vietqr")
    res, err := g.CreatePayment(ctx, pay.PaymentRequest{OrderRef: "o1", Amount: 29000})
    require.NoError(t, err)
    require.NotEmpty(t, res.QRPayload)
    require.Equal(t, int64(29000), res.Amount)
}

// services/bill/internal/pay/sign_test.go
func TestSignMoMo_FixedVector(t *testing.T) {
    secrets := fakeVault{"bill/momo/secret_key": "test-key"}
    sig, err := SignMoMo(ctx, secrets, "amount=29000&orderId=o1")
    require.NoError(t, err)
    require.Equal(t, hmacHex("test-key", "amount=29000&orderId=o1"), sig)
}

// services/bill/internal/api/checkout_test.go
func TestCheckout_HappyPath(t *testing.T) {
    h := setupCheckout(t)
    rec := doPOST(t, h, "/v1/billing/checkout", `{"plan_tier":"premium_basic","gateway":"vietqr"}`)
    require.Equal(t, 200, rec.Code)
    var res pay.PaymentResult
    decode(t, rec, &res)
    require.Equal(t, int64(29000), res.Amount) // lấy từ plan_catalog
}

func TestCheckout_AmountFromCatalog_NotBody(t *testing.T) {
    h := setupCheckout(t)
    rec := doPOST(t, h, "/v1/billing/checkout",
        `{"plan_tier":"premium_basic","gateway":"vietqr","amount":1}`) // body bịa amount=1
    var res pay.PaymentResult
    decode(t, rec, &res)
    require.Equal(t, int64(29000), res.Amount) // bỏ qua amount client, dùng catalog
}

func TestCheckout_Idempotent_NoDoubleCharge(t *testing.T) {
    h := setupCheckout(t)
    body := `{"plan_tier":"premium_basic","gateway":"vietqr"}`
    doPOST(t, h, "/v1/billing/checkout", body)
    doPOST(t, h, "/v1/billing/checkout", body) // bấm lại
    require.Equal(t, 1, h.gatewayCalls()) // chỉ một intent ở cổng
}

func TestCheckout_GatewayError_502_StaysPending(t *testing.T) {
    h := setupCheckoutGatewayFail(t)
    rec := doPOST(t, h, "/v1/billing/checkout", `{"plan_tier":"premium_basic","gateway":"momo"}`)
    require.Equal(t, 502, rec.Code)
    require.NotContains(t, []string{"confirmed", "paid"}, h.lastPaymentStatus()) // không đánh dấu thành công
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `gateway.go` (interface + registry) -> `sign.go` (ký per-cổng, khóa Vault FR-INFRA-003) -> bốn adapter `momo.go`/`zalopay.go`/`vnpay.go`/`vietqr.go` (tạo yêu cầu + parse phản hồi) -> `checkout.go` (handler: tra plan, sinh order_ref idempotent, gọi gateway, ghi payment pending) -> đăng ký route sau JWT middleware (FR-INFRA-001) -> tests. Số tiền luôn lấy từ `plan_catalog` (FR-BILL-001), không từ client. HTTP client tới cổng có timeout; lỗi giữ `payment` pending. Trạng thái thanh toán cuối do FR-BILL-003 chốt qua IPN/webhook.

---

## §7 - Phụ thuộc

- **FR-BILL-001** - `plan_catalog` (giá BIGINT theo tier) + `subscription`; checkout lấy giá từ đây.
- **FR-INFRA-003 (Vault)** - đọc khóa ký mỗi cổng; khóa không nằm trong DB/code.
- **FR-INFRA-001 (gateway)** - gắn JWT + `user_id`; định tuyến `POST /v1/billing/checkout`.
- **FR-BILL-003 (downstream)** - nhận IPN/webhook cổng, đối soát `payment`, đẩy `subscription` sang active + `renews_at`.
- **FR-COMPLY-005** - audit no-cleartext: khóa ký từ Vault, không lưu thẻ.
- Lib: `crypto/hmac`, `crypto/sha256`, `net/http`, `encoding/json`.

---

## §8 - Payload ví dụ

### Checkout request (user chọn VietQR)

```http
POST /v1/billing/checkout HTTP/1.1
Authorization: Bearer <JWT-SanDeal>
Content-Type: application/json

{"plan_tier": "premium_basic", "gateway": "vietqr"}
```

### PaymentResult (VietQR - trả chuỗi QR)

```json
{
  "order_ref": "sd_bill_7_premium_basic_1719500000",
  "gateway": "vietqr",
  "amount": 29000,
  "qr_payload": "00020101021238540010A00000072701240006970436..."
}
```

### PaymentResult (MoMo - trả deeplink)

```json
{
  "order_ref": "sd_bill_7_premium_basic_1719500050",
  "gateway": "momo",
  "amount": 29000,
  "pay_url": "https://test-payment.momo.vn/gw_payment/...&signature=..."
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Phí chính xác từng cổng theo hợp đồng merchant (§10 tài liệu nguồn - cần xác minh) - lưu vào cấu hình khi có hợp đồng; FR này ưu tiên QR theo khoảng ~1,5-2,5%.
- Auto-renew (tạo charge định kỳ không cần user bấm lại) - cần token hóa của cổng; thêm khi cổng hỗ trợ và có consent rõ; hiện mỗi kỳ user xác nhận.
- Hoàn tiền (refund) qua API cổng - thêm khi có quy trình hủy/DSAR; FR này chỉ tạo thanh toán.
- Chọn cổng theo gợi ý thông minh (lịch sử user) - hiện QR mặc định gợi ý, user chọn.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Lưu thông tin thẻ | review + grep (§4 #9) | PCI scope nặng, lộ thẻ | Không chạm thẻ, cổng xử lý (DEC-BILL-09) |
| Khóa ký cleartext | review (§4 #8) | lộ DB = giả mạo thanh toán | Khóa từ Vault (DEC-BILL-09) |
| Số tiền từ client | amount-from-catalog test | thao túng giá | Lấy từ plan_catalog (§1 #3) |
| Double-charge khi bấm lại | idempotent test | trừ tiền hai lần | order_ref idempotent (§1 #6) |
| Coi lỗi mạng là thành công | gateway-error test | cấp Premium chưa thu tiền | Lỗi giữ pending (§1 #11) |
| Phí thẻ cao ăn biên | gợi ý QR mặc định | biên mỏng thêm | Ưu tiên QR (§1 #7) |
| plan/gateway lạ | 400 test | request rác | Validate allowlist (§1 #10) |
| Chữ ký sai spec cổng | sign vector test | cổng từ chối | Ký đúng spec từng cổng (§1 #4) |
| Trạng thái thật chưa chốt | payment pending | - | FR-BILL-003 chốt qua IPN |

---

## §11 - Ghi chú

- FR-BILL-002 là cửa thu tiền của Premium: không có nó, FR-BILL-001 không thành doanh thu.
- Interface chung + adapter per-cổng giấu khác biệt MoMo/ZaloPay/VNPay/VietQR; thêm cổng là thêm adapter.
- Ưu tiên QR (phí ~1,5-2,5% so với thẻ cao hơn) bảo vệ biên ở thị trường WTP thấp - mặc định gợi ý, không ép.
- Không lưu thẻ giữ phạm vi PCI tối thiểu; khóa ký từ Vault tuân no-cleartext (§5.5).
- Idempotent theo order_ref đóng cửa double-charge khi user bấm lại/retry.
- Lỗi gọi cổng giữ payment pending - nguồn sự thật về thanh toán là cổng (qua IPN FR-BILL-003), không phải kết quả lời gọi.
- Số tiền luôn từ plan_catalog, không từ client - chống thao túng giá.

---

*Hết FR-BILL-002. Status: ready_to_implement (mục tiêu audit 10/10).*
