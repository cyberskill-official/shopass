---
id: TASK-BILL-003
title: "Bảng `payment` (gateway, amount, fee, paid_at, status) + webhook IPN + reconciliation - xác thực IPN, idempotent, khớp số tiền, kích hoạt subscription khi paid"
module: BILL
priority: MUST
status: done
verify: T
phase: P2
milestone: P2 - slice 2
slice: 2
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-BILL-001, TASK-BILL-002, TASK-INFRA-003, TASK-INFRA-004, TASK-COMPLY-005]
depends_on: [TASK-BILL-002]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.4 (data model: payment gateway/amount/fee/paid_at/status)"
  - "docs/... §4.1 (phí giao dịch QR ~1,5-2,5% - fee lưu để tính doanh thu ròng), §3.8 (observability)"
source_decisions:
  - "DEC-BILL-11: payment lưu gateway, amount (BIGINT VND), fee (BIGINT VND), paid_at, status; gắn order_ref (TASK-BILL-002) + subscription"
  - "DEC-BILL-12: IPN/webhook cổng xác thực chữ ký (khóa Vault TASK-INFRA-003) trước khi tin; chữ ký sai -> bỏ qua, không đổi payment"
  - "DEC-BILL-13: IPN idempotent theo order_ref + transaction_id cổng; cùng IPN gửi lại không đổi trạng thái hai lần (cổng retry)"
  - "DEC-BILL-14: khớp số tiền IPN với payment.amount (phải bằng); lệch -> đánh dấu 'mismatch' để điều tra, KHÔNG kích hoạt subscription"
  - "DEC-BILL-15: payment 'paid' (khớp) -> kích hoạt/gia hạn subscription (TASK-BILL-001 SetRenewsAt + status active); reconciliation định kỳ bù IPN bị mất"

language: "Go 1.22 (bill-svc); PostgreSQL 16; webhook nhận qua API Gateway (TASK-INFRA-001)"
service: shopass/services/bill/
new_files:
  - services/bill/migrations/0003_payment.sql
  - services/bill/internal/api/ipn.go
  - services/bill/internal/bill/reconcile.go
  - services/bill/internal/bill/payment_repo.go
  - services/bill/internal/api/ipn_test.go
  - services/bill/internal/bill/reconcile_test.go
modified_files:
  - services/bill/internal/api/router.go            # đăng ký POST /v1/billing/ipn/{gateway}
allowed_tools:
  - file_read: services/bill/**
  - file_write: services/bill/**
  - bash: cd services/bill && go test ./...
disallowed_tools:
  - tin IPN chưa xác thực chữ ký (vi phạm DEC-BILL-12 - giả mạo thanh toán)
  - kích hoạt subscription khi số tiền IPN lệch payment.amount (vi phạm DEC-BILL-14)
  - đổi trạng thái payment hai lần khi cổng gửi IPN lặp (vi phạm DEC-BILL-13)
  - lưu amount/fee dạng float (vi phạm DEC-BILL-11/DEC-PRICE-05)

effort_hours: 6
sub_tasks:
  - "0.5h: 0003_payment.sql - bảng payment (order_ref UNIQUE, gateway, amount, fee, status, paid_at, transaction_id) + CHECK"
  - "1.0h: payment_repo.go - InsertPending (từ TASK-BILL-002), MarkPaid, MarkFailed, MarkMismatch, ByOrderRef"
  - "1.5h: ipn.go - handler webhook: verify chữ ký -> idempotent check -> khớp amount -> MarkPaid + kích hoạt subscription"
  - "1.0h: reconcile.go - job định kỳ đối soát payment pending lâu với report cổng (bù IPN mất)"
  - "1.0h: ipn_test.go - chữ ký sai bỏ qua; amount khớp -> paid + sub active; amount lệch -> mismatch không active; IPN lặp idempotent"
  - "0.5h: reconcile_test.go - payment pending lâu được đối soát; paid bù kích hoạt subscription"
  - "0.5h: OTel metric payment_ipn_total{gateway,result} + payment_revenue_net_vnd (amount - fee)"

risk_if_skipped: "payment + reconciliation là nơi tiền Premium thực sự được chốt và đối soát - không có nó thì checkout (TASK-BILL-002) tạo intent rồi không bao giờ biết user đã trả hay chưa, subscription không được kích hoạt, và doanh thu không đối soát được với cổng. Nếu tin IPN chưa xác thực chữ ký thì kẻ xấu gửi IPN giả 'đã thanh toán' để được cấp Premium miễn phí (mất doanh thu + gian lận). Nếu kích hoạt subscription khi số tiền lệch thì user trả 1k được Premium 79k - lỗ trực tiếp. Nếu đổi trạng thái hai lần khi cổng retry IPN thì gia hạn nhầm/đếm doanh thu trùng. Nếu không lưu fee thì không tính được doanh thu ròng (phí QR ~1,5-2,5% §4.1) cho unit economics. Nếu không có reconciliation định kỳ thì IPN bị mất (mạng) làm user đã trả tiền không được kích hoạt - khiếu nại. Đây là mắt xích chốt tiền của toàn hệ Premium."
---

## §1 - Mô tả (BCP-14 normative)

Service BILL **MUST** định nghĩa bảng `payment`, nhận IPN/webhook từ cổng (có xác thực chữ ký, idempotent, khớp số tiền), kích hoạt subscription khi thanh toán hợp lệ, và đối soát định kỳ để bù IPN bị mất. Hợp đồng:

1. **MUST** định nghĩa bảng `payment (id, order_ref, subscription_id, gateway, amount, fee, status, transaction_id, paid_at, created_at)` với `order_ref TEXT NOT NULL UNIQUE` (idempotency từ TASK-BILL-002) và `subscription_id` REFERENCES `subscription(id)`.
2. **MUST** lưu `amount` và `fee` dạng `BIGINT` (VND, không thập phân) (DEC-BILL-11) - KHÔNG float; `fee` là phí cổng (~1,5-2,5% với QR, §4.1) để tính doanh thu ròng `amount - fee`.
3. **MUST** ràng buộc `payment.status` - {`'pending'`,`'paid'`,`'failed'`,`'mismatch'`} qua CHECK; `amount >= 0`, `fee >= 0` qua CHECK.
4. **MUST** phục vụ webhook `POST /v1/billing/ipn/{gateway}` nhận IPN của cổng mang tối thiểu `{order_ref, transaction_id, amount, status}`.
5. **MUST** xác thực chữ ký IPN bằng khóa của `{gateway}` đọc từ Vault (TASK-INFRA-003) trước khi tin (DEC-BILL-12); chữ ký sai/thiếu -> bỏ qua (trả `200`/`400` theo spec cổng để cổng ngừng retry), KHÔNG đổi `payment`.
6. **MUST** idempotent theo `order_ref` + `transaction_id` (DEC-BILL-13): cùng IPN gửi lại (cổng retry) **MUST NOT** đổi trạng thái payment hai lần hay gia hạn subscription hai lần. IPN trên payment đã `paid` là no-op.
7. **MUST** khớp `amount` trong IPN với `payment.amount` đã ghi (DEC-BILL-14): bằng -> tiếp tục; lệch -> đặt `status='mismatch'`, ghi log/alert, và **MUST NOT** kích hoạt subscription (số tiền không khớp là dấu hiệu lỗi/gian lận).
8. Khi IPN hợp lệ và khớp với `status='paid'` từ cổng, **MUST** đặt `payment.status='paid'`, `paid_at=now()`, lưu `transaction_id`, rồi kích hoạt subscription (DEC-BILL-15): gọi TASK-BILL-001 `SetRenewsAt` (đẩy kỳ kế) + `UpdateStatus` về `active`.
9. Khi IPN báo cổng `failed`/hủy, **MUST** đặt `payment.status='failed'` và KHÔNG kích hoạt subscription; subscription giữ trạng thái cũ (hoặc `past_due` nếu là kỳ gia hạn).
10. **MUST** có job reconciliation định kỳ (`reconcile.go`) quét `payment` `pending` quá ngưỡng thời gian, đối soát với report/truy vấn cổng, và chốt trạng thái (bù IPN bị mất) - user đã trả không bị kẹt `pending` mãi.
11. **MUST** lấy khóa xác thực từ Vault ở thời điểm verify; webhook KHÔNG dùng JWT user (cổng không có), chỉ chữ ký. Mọi IPN (kể cả bị từ chối) **SHOULD** ghi log để truy vết.
12. **SHOULD** phát OTel: `payment_ipn_total{gateway, result}` (`result - {paid, failed, mismatch, unauthorized, duplicate}`), `payment_revenue_net_vnd` (histogram `amount - fee`), `payment_pending_age_seconds` (gauge cho reconciliation).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao xác thực chữ ký IPN (DEC-BILL-12)?** IPN webhook là endpoint công khai. Nếu tin payload chưa xác thực, kẻ xấu gửi IPN giả "order_ref X đã thanh toán" để được cấp Premium miễn phí. Xác thực chữ ký bằng khóa chung với cổng đảm bảo chỉ cổng thật mới đổi được trạng thái payment - rào chống gian lận tài chính cốt lõi, đồng nhất cách TASK-AFFIL-003 xử lý postback.

**Vì sao khớp số tiền (DEC-BILL-14)?** Một IPN hợp lệ về chữ ký vẫn có thể mang số tiền không khớp đơn (lỗi cổng, hoặc tấn công tinh vi). Nếu kích hoạt subscription mà không khớp `amount`, user trả 1k được Premium 79k. Khớp `amount` với `payment.amount` đã ghi và từ chối (đánh `mismatch`) khi lệch đóng lỗ hổng này; `mismatch` được điều tra thủ công thay vì âm thầm cấp quyền.

**Vì sao idempotent theo order_ref + transaction_id (DEC-BILL-13)?** Cổng gửi lại IPN khi không nhận được phản hồi đúng. Nếu mỗi IPN đổi trạng thái/gia hạn, một thanh toán bị đếm nhiều lần và subscription gia hạn dư. `order_ref` UNIQUE + kiểm `transaction_id` làm IPN idempotent: lần đầu chốt, lần sau no-op.

**Vì sao lưu fee (DEC-BILL-11, §1 #2)?** Doanh thu thật là `amount - fee` (phí cổng). Phí QR ~1,5-2,5% (§4.1) là biến phí trực tiếp trên mỗi giao dịch. Lưu `fee` cho phép báo cáo doanh thu ròng đúng và theo dõi unit economics - không lưu thì chỉ biết doanh thu gộp, sai bức tranh biên.

**Vì sao reconciliation định kỳ (DEC-BILL-15, §1 #10)?** IPN có thể mất (mạng chập, webhook đỗ). Nếu chỉ dựa IPN, một user đã trả tiền mà IPN không tới sẽ kẹt `pending` mãi - đã mất tiền mà không có Premium, khiếu nại chính đáng. Job đối soát định kỳ quét `pending` lâu, hỏi cổng trạng thái thật, và chốt - lưới an toàn cho IPN mất. Nguồn sự thật cuối là cổng, IPN chỉ là thông báo nhanh.

**Vì sao tách kích hoạt subscription khỏi checkout (§1 #8)?** Checkout (TASK-BILL-002) chỉ tạo ý định thanh toán; tiền chưa chắc vào. Chỉ khi IPN xác nhận `paid` + khớp số tiền, ta mới kích hoạt subscription. Tách "tạo ý định" (BILL-002) khỏi "xác nhận đã trả -> cấp quyền" (BILL-003) đảm bảo Premium chỉ bật sau khi tiền thực sự vào, đúng thứ tự nhân quả.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/bill/migrations/0003_payment.sql
CREATE TABLE payment (
  id             BIGSERIAL   PRIMARY KEY,
  order_ref      TEXT        NOT NULL UNIQUE,        -- idempotency (TASK-BILL-002)
  subscription_id BIGINT     REFERENCES subscription(id),
  gateway        TEXT        NOT NULL,
  amount         BIGINT      NOT NULL CHECK (amount >= 0),  -- VND
  fee            BIGINT      NOT NULL DEFAULT 0 CHECK (fee >= 0), -- phí cổng VND (§1 #2)
  status         TEXT        NOT NULL DEFAULT 'pending'
                   CHECK (status IN ('pending','paid','failed','mismatch')),
  transaction_id TEXT,                               -- mã giao dịch cổng (idempotent)
  paid_at        TIMESTAMPTZ,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_payment_pending ON payment (created_at) WHERE status = 'pending'; -- reconciliation
```

### IPN handler (Go)

```go
// services/bill/internal/api/ipn.go
func (h *Handler) HandleIPN(w http.ResponseWriter, req *http.Request) {
    gateway := req.PathValue("gateway")
    body, _ := io.ReadAll(req.Body)
    sig := req.Header.Get("X-Signature")

    ok, err := pay.VerifyIPN(req.Context(), h.secrets, gateway, body, sig)
    if err != nil { writeErr(w, 500, "verify error"); return }
    if !ok {
        metrics.IPN(gateway, "unauthorized")
        writeErr(w, 400, "invalid signature"); return // KHÔNG đổi payment (§1 #5)
    }

    var ipn IPNPayload
    if err := json.Unmarshal(body, &ipn); err != nil { writeErr(w, 400, "bad payload"); return }

    p, found := h.payments.ByOrderRef(req.Context(), ipn.OrderRef)
    if !found { writeErr(w, 404, "unknown order_ref"); return }
    if p.Status == "paid" { // idempotent: IPN lặp (§1 #6)
        metrics.IPN(gateway, "duplicate"); w.WriteHeader(200); return
    }
    if ipn.Amount != p.Amount { // khớp số tiền (§1 #7)
        h.payments.MarkMismatch(req.Context(), p.ID, ipn.Amount)
        metrics.IPN(gateway, "mismatch")
        log.Warn("payment amount mismatch", "order_ref", ipn.OrderRef,
            "expected", p.Amount, "got", ipn.Amount)
        w.WriteHeader(200); return // KHÔNG kích hoạt subscription
    }
    switch ipn.Status {
    case "paid":
        h.payments.MarkPaid(req.Context(), p.ID, ipn.TransactionID)
        h.subs.Activate(req.Context(), p.SubscriptionID) // TASK-BILL-001 SetRenewsAt + active (§1 #8)
        metrics.IPN(gateway, "paid")
    default:
        h.payments.MarkFailed(req.Context(), p.ID)
        metrics.IPN(gateway, "failed")
    }
    w.WriteHeader(200)
}
```

### Reconciliation (Go)

```go
// services/bill/internal/bill/reconcile.go
// ReconcilePending quét payment pending quá ngưỡng, hỏi cổng trạng thái thật, chốt (§1 #10).
func (r *Reconciler) ReconcilePending(ctx context.Context, olderThan time.Duration) error {
    rows, err := r.payments.ListPending(ctx, time.Now().Add(-olderThan))
    if err != nil {
        return err
    }
    for _, p := range rows {
        st, txID, err := r.gateways.QueryStatus(ctx, p.Gateway, p.OrderRef) // hỏi cổng
        if err != nil { continue }
        if st == "paid" {
            r.payments.MarkPaid(ctx, p.ID, txID)
            r.subs.Activate(ctx, p.SubscriptionID) // bù kích hoạt IPN bị mất
        } else if st == "failed" {
            r.payments.MarkFailed(ctx, p.ID)
        }
    }
    return nil
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> `payment` tồn tại với `order_ref` UNIQUE, CHECK status/amount/fee.
2. IPN chữ ký sai -> KHÔNG đổi `payment` (status vẫn `pending`); metric `unauthorized`.
3. IPN hợp lệ, `amount` khớp, cổng `paid` -> `payment.status='paid'`, `paid_at` set, `transaction_id` lưu, subscription được kích hoạt (active + renews_at đẩy).
4. IPN hợp lệ nhưng `amount` lệch `payment.amount` -> `status='mismatch'`, subscription KHÔNG kích hoạt, log cảnh báo.
5. IPN báo cổng `failed` -> `status='failed'`, subscription không kích hoạt.
6. IPN lặp (cùng order_ref, payment đã `paid`) -> no-op, subscription không gia hạn lần hai; metric `duplicate`.
7. IPN với `order_ref` không tồn tại -> `404`, không tạo payment.
8. `amount`/`fee` lưu BIGINT; `INSERT` `amount < 0` hoặc `fee < 0` -> lỗi CHECK.
9. `INSERT payment` `status='refunded'` (ngoài tập) -> lỗi CHECK.
10. `ReconcilePending` trên payment `pending` lâu mà cổng báo `paid` -> chốt `paid` + kích hoạt subscription (bù IPN mất).
11. Doanh thu ròng tính đúng `amount - fee` (kiểm metric/report).
12. Khóa xác thực IPN đọc từ Vault (mock); review xác nhận không có khóa cleartext.

---

## §5 - Kiểm thử (verification)

```go
// services/bill/internal/api/ipn_test.go
func TestIPN_BadSignature_NoChange(t *testing.T) {
    h, repo := setupIPN(t, "o1", 29000) // payment pending
    rec := doSignedIPN(t, h, "momo", `{"order_ref":"o1","transaction_id":"t1","amount":29000,"status":"paid"}`, "WRONG")
    require.Equal(t, 400, rec.Code)
    require.Equal(t, "pending", repo.StatusByOrderRef("o1")) // không đổi
}

func TestIPN_Paid_ActivatesSubscription(t *testing.T) {
    h, repo := setupIPN(t, "o2", 29000)
    body := `{"order_ref":"o2","transaction_id":"t2","amount":29000,"status":"paid"}`
    doSignedIPN(t, h, "momo", body, sign(body))
    require.Equal(t, "paid", repo.StatusByOrderRef("o2"))
    require.True(t, repo.SubscriptionActive("o2")) // subscription kích hoạt
}

func TestIPN_AmountMismatch_NoActivate(t *testing.T) {
    h, repo := setupIPN(t, "o3", 79000) // payment ghi 79000
    body := `{"order_ref":"o3","transaction_id":"t3","amount":1000,"status":"paid"}` // IPN 1000
    doSignedIPN(t, h, "momo", body, sign(body))
    require.Equal(t, "mismatch", repo.StatusByOrderRef("o3"))
    require.False(t, repo.SubscriptionActive("o3")) // KHÔNG cấp Premium (§1 #7)
}

func TestIPN_Duplicate_Idempotent(t *testing.T) {
    h, repo := setupIPN(t, "o4", 29000)
    body := `{"order_ref":"o4","transaction_id":"t4","amount":29000,"status":"paid"}`
    doSignedIPN(t, h, "momo", body, sign(body))
    doSignedIPN(t, h, "momo", body, sign(body)) // cổng retry
    require.Equal(t, 1, repo.ActivateCount("o4")) // không gia hạn hai lần
}

func TestIPN_NegativeAmount_CheckRejects(t *testing.T) {
    repo := setupRepo(t)
    err := repo.InsertRaw(ctx, "o5", "momo", -1, 0)
    require.Error(t, err) // CHECK amount >= 0
}
```

```go
// services/bill/internal/bill/reconcile_test.go
func TestReconcile_PaidBackfill(t *testing.T) {
    r, repo := setupReconciler(t, "o6", 29000) // pending lâu, cổng báo paid
    require.NoError(t, r.ReconcilePending(ctx, time.Hour))
    require.Equal(t, "paid", repo.StatusByOrderRef("o6"))
    require.True(t, repo.SubscriptionActive("o6")) // bù kích hoạt IPN mất
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration `0003_payment.sql` (order_ref UNIQUE + CHECK + index pending) -> `payment_repo.go` (`InsertPending` từ TASK-BILL-002, `MarkPaid`/`MarkFailed`/`MarkMismatch`, `ByOrderRef`) -> `ipn.go` (verify chữ ký Vault -> idempotent check -> khớp amount -> MarkPaid + kích hoạt subscription TASK-BILL-001) -> `reconcile.go` (job định kỳ đối soát pending) -> đăng ký `POST /v1/billing/ipn/{gateway}` (`http.ServeMux` `PathValue`) -> tests. Webhook KHÔNG dùng JWT user; xác thực bằng chữ ký HMAC, khóa từ TASK-INFRA-003. `subs.Activate` gọi `SetRenewsAt` + `UpdateStatus(active)` của TASK-BILL-001.

---

## §7 - Phụ thuộc

- **TASK-BILL-002** - tạo `payment` `pending` + `order_ref`; IPN chốt trạng thái thanh toán đó.
- **TASK-BILL-001** - `subscription` + `SetRenewsAt`/`UpdateStatus`; kích hoạt khi payment `paid`.
- **TASK-INFRA-003 (Vault)** - khóa xác thực IPN mỗi cổng; không nằm trong DB/code.
- **TASK-INFRA-001 (gateway)** - định tuyến webhook IPN; bỏ qua JWT user, dùng chữ ký.
- **TASK-INFRA-004 (observability)** - alert khi `mismatch` / pending tồn lâu / IPN lỗi.
- **TASK-BILL-005 (downstream)** - feature gating bật khi subscription active (sau payment paid).
- **TASK-COMPLY-005** - audit no-cleartext: khóa IPN từ Vault.
- Lib: `crypto/hmac`, `crypto/sha256`, `net/http`, `encoding/json`.

---

## §8 - Payload ví dụ

### IPN từ cổng (MoMo, thanh toán thành công)

```http
POST /v1/billing/ipn/momo HTTP/1.1
Content-Type: application/json
X-Signature: 3bf4f1b2b0b822cd15d6c15b0f00a089f86d081884c7d659a2feaa0c55ad015a

{
  "order_ref": "sd_bill_7_premium_basic_1719500000",
  "transaction_id": "MOMO2026062812345",
  "amount": 29000,
  "status": "paid"
}
```

### payment sau khi chốt

```sql
SELECT order_ref, gateway, amount, fee, status, transaction_id FROM payment WHERE order_ref='sd_bill_7_premium_basic_1719500000';
--  ...basic_1719500000 | momo | 29000 | 580 | paid | MOMO2026062812345
-- doanh thu ròng = amount - fee = 29000 - 580 = 28420 VND
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- `fee` chính xác từng cổng (§10 tài liệu nguồn - cần xác minh theo hợp đồng merchant) - hiện ghi theo phản hồi cổng nếu có, hoặc tính theo tỷ lệ cấu hình ~1,5-2,5%.
- Chargeback/refund làm đảo payment `paid` -> hủy/hạ subscription - cần IPN `refunded`; map khi cổng hỗ trợ.
- Tần suất job reconciliation + ngưỡng pending - cấu hình; mặc định quét pending > 1 giờ.
- Đối soát tổng định kỳ (cuối ngày so tổng doanh thu với report cổng) - thêm báo cáo khi có dashboard tài chính.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Tin IPN chưa xác thực | bad-signature test | Premium miễn phí qua IPN giả | Verify chữ ký Vault trước (§1 #5) |
| Kích hoạt khi amount lệch | mismatch test | trả 1k được Premium 79k | Khớp amount, đánh mismatch (§1 #7) |
| IPN lặp đổi trạng thái 2 lần | duplicate test | gia hạn dư, đếm doanh thu trùng | Idempotent order_ref+tx (§1 #6) |
| amount/fee float | review + CHECK | sai số doanh thu ròng | BIGINT VND (§1 #2) |
| IPN mất, user kẹt pending | reconcile test | đã trả không có Premium | Reconciliation định kỳ bù (§1 #10) |
| status ngoài tập | DB CHECK | trạng thái rác | CHECK status IN (...) (§1 #3) |
| order_ref lạ | 404 test | IPN trỏ payment không tồn tại | Tra ByOrderRef trước (§1 #4) |
| Khóa IPN cleartext | review (§4 #12) | giả mạo IPN | Khóa từ Vault (§1 #11) |
| Không tính được biên | fee lưu + net metric | unit economics sai | Doanh thu ròng amount-fee (§1 #2) |

---

## §11 - Ghi chú

- payment + reconciliation là nơi tiền Premium được chốt: chỉ khi IPN xác nhận paid + khớp số tiền, subscription mới bật.
- Xác thực chữ ký IPN chặn IPN giả "đã thanh toán" - webhook công khai, không xác thực là mời gian lận.
- Khớp số tiền đóng lỗ hổng "trả ít được nhiều"; lệch -> mismatch điều tra, không âm thầm cấp quyền.
- Idempotent theo order_ref + transaction_id chịu được cổng retry mà không gia hạn/đếm trùng.
- Lưu fee cho phép tính doanh thu ròng (phí QR ~1,5-2,5% §4.1) - đúng bức tranh unit economics.
- Reconciliation định kỳ là lưới an toàn cho IPN mất: nguồn sự thật cuối là cổng, không phải thông báo nhanh.
- Tách kích hoạt subscription (BILL-003) khỏi tạo ý định (BILL-002): Premium chỉ bật sau khi tiền thực sự vào.

---

*Hết TASK-BILL-003. Status: ready_to_implement (mục tiêu audit 10/10).*
