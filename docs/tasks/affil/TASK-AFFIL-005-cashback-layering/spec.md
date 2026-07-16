---
id: TASK-AFFIL-005
title: "Cashback layering trên affiliate - chia % cho user, hold đến khi affiliate confirm, delay payout chống gian lận"
module: AFFIL
priority: SHOULD
status: done
verify: T
phase: P3
milestone: P3 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-AFFIL-001, TASK-AFFIL-002, TASK-AFFIL-003, TASK-BILL-002, TASK-TRUST-005]
depends_on: [TASK-AFFIL-003, TASK-BILL-002, TASK-TRUST-005]
blocks: []
source_pages:
  - "docs/... SănDeal §6 mục 2 (cashback layering trên affiliate)"
  - "docs/... §5.3 (gian lận user, delay payout affiliate để điều tra)"
  - "docs/... §4.2 (affiliate reality-check, last-click)"
source_decisions:
  - "DEC-AFFIL-50: cashback chỉ tính trên affiliate_conversion đã confirmed (không trả trước trên click)"
  - "DEC-AFFIL-51: hold-then-release - cashback giữ ở trạng thái pending đến khi network confirm và qua cửa sổ điều tra"
  - "DEC-AFFIL-52: SănDeal giữ phần chênh (kept_margin = commission - user_share) làm doanh thu"
  - "DEC-AFFIL-53: payout chỉ qua số dư nội bộ / VietQR khi đạt ngưỡng tối thiểu, KHÔNG tự động mỗi giao dịch"

language: "Go 1.22 (affil-svc) + Postgres"
service: shopass/services/affil/
new_files:
  - services/affil/migrations/0005_cashback_ledger.sql
  - services/affil/internal/cashback/ledger.go
  - services/affil/internal/cashback/release.go
  - services/affil/internal/cashback/payout.go
  - services/affil/internal/cashback/ledger_test.go
  - services/affil/internal/cashback/release_test.go
modified_files:
  - services/affil/internal/affiliate/conversion.go    # phát sự kiện khi conversion confirmed
allowed_tools:
  - file_read: services/affil/**
  - file_write: services/affil/**
  - bash: cd services/affil && go test ./...
disallowed_tools:
  - trả cashback trước khi affiliate_conversion sang trạng thái confirmed (vi phạm DEC-AFFIL-50)
  - bỏ qua cửa sổ delay payout (vi phạm DEC-AFFIL-51 và §5.3)
  - tự động payout từng giao dịch nhỏ (vi phạm DEC-AFFIL-53)

effort_hours: 10
sub_tasks:
  - "0.5h: 0005_cashback_ledger.sql - bảng cashback_entry + payout_request"
  - "1.0h: ledger.go - tạo entry pending khi conversion confirmed, tính user_share + kept_margin"
  - "1.0h: release.go - chuyển pending -> available khi qua cửa sổ điều tra (TASK-TRUST-005 clear)"
  - "1.0h: payout.go - gom available -> payout_request khi >= ngưỡng, qua VietQR (TASK-BILL-002)"
  - "0.5h: hook sự kiện conversion.confirmed -> ledger"
  - "1.5h: ledger_test.go - confirmed tạo pending; chưa confirmed không tạo; idempotent"
  - "1.5h: release_test.go - chưa qua cửa sổ giữ pending; qua cửa sổ + sạch -> available; flagged -> giữ/hủy"
  - "0.5h: metric cashback_pending_total + released_total + clawback_total"
risk_if_skipped: "Cashback là đòn bẩy user-love rất cao (§6) và là kênh giữ retention -> affiliate volume. Nếu trả cashback trước khi confirm hoặc không delay payout, kẻ gian lận tạo đơn giả, rút cashback, rồi hủy đơn -> SănDeal mất tiền thật. Hold-then-release + delay payout (best practice ngành §5.3) là lằn ranh giữa một tính năng sinh lời và một lỗ hổng bị farm."
---

## §1 - Mô tả (BCP-14 normative)

Service AFFIL **SHOULD** chia lại một phần hoa hồng affiliate cho người dùng dưới dạng cashback, theo mô hình hold-then-release để chống gian lận. Hợp đồng:

1. **MUST** chỉ tạo cashback khi `affiliate_conversion.status = 'confirmed'` (DEC-AFFIL-50). Cashback KHÔNG được tính trên click hay trên conversion còn `pending`.
2. **MUST** định nghĩa bảng `cashback_entry (id, user_id, conversion_id, commission, user_share, kept_margin, status, created_at, available_at, paid_at)`. `status` thuộc `{pending, available, paid, clawed_back}`.
3. **MUST** tính `user_share = floor(commission * share_rate)` và `kept_margin = commission - user_share` (DEC-AFFIL-52). `share_rate` cấu hình theo tier (ví dụ free 30%, Premium 50%). Tất cả đơn vị VND, BIGINT.
4. **MUST** giữ `cashback_entry` ở `pending` ít nhất qua cửa sổ điều tra (`available_at = confirmed_at + investigation_window`, mặc định 7-14 ngày) (DEC-AFFIL-51). Chỉ sang `available` khi cửa sổ đã qua VÀ TASK-TRUST-005 không gắn cờ gian lận trên conversion/user.
5. **MUST** hỗ trợ clawback: nếu network thu hồi conversion (status đổi từ `confirmed` về `rejected`) hoặc TASK-TRUST-005 xác nhận gian lận, cashback_entry sang `clawed_back`; nếu đã `available` nhưng chưa `paid` thì trừ khỏi số dư.
6. **MUST** gom các `available` entry của một user thành `payout_request` chỉ khi tổng đạt ngưỡng tối thiểu (ví dụ 50.000 VND) (DEC-AFFIL-53); payout qua VietQR/số dư nội bộ theo TASK-BILL-002. KHÔNG tự động payout từng giao dịch nhỏ.
7. **MUST** làm việc tạo entry idempotent theo `conversion_id` (một conversion confirmed chỉ sinh tối đa một cashback_entry).
8. **SHOULD** phát metric: `cashback_pending_total`, `cashback_released_total`, `cashback_clawback_total`, `cashback_paid_total` (VND counter).
9. **MUST** ghi disclosure rõ cho user: cashback là pending đến ngày `available_at`, có thể bị hủy nếu đơn bị hủy/hoàn (minh bạch hậu-Honey, §4.2).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao chỉ tính trên confirmed (DEC-AFFIL-50)?** Affiliate là last-click: hoa hồng chỉ có thật khi network xác nhận đơn hợp lệ (qua cửa sổ đối soát). Trả cashback trên click hoặc pending là trả tiền chưa tồn tại - kẻ gian lận tạo đơn rồi hủy sẽ rút sạch.

**Vì sao hold-then-release (DEC-AFFIL-51)?** Đây là lằn ranh chống farm. Đơn affiliate có thể bị hủy/hoàn sau khi confirm. Giữ cashback pending qua cửa sổ điều tra cho phép clawback trước khi tiền rời hệ thống. Best practice ngành là delay payout để điều tra (§5.3).

**Vì sao SănDeal giữ kept_margin (DEC-AFFIL-52)?** Cashback layering chia lại một phần, giữ phần chênh làm doanh thu. Đây là mô hình kiếm tiền của tính năng (§6 mục 2): user-love cao, margin từ phần giữ lại.

**Vì sao ngưỡng payout tối thiểu (DEC-AFFIL-53)?** Payout từng giao dịch nhỏ qua VietQR tăng phí cố định và tạo bề mặt gian lận. Gom đến ngưỡng giảm phí và cho phép một lần điều tra cuối trước khi trả.

**Vì sao buộc qua TASK-TRUST-005 (§1 #4)?** TASK-TRUST-005 phát hiện gaming attribution + delay payout. Cashback release phải hỏi tín hiệu gian lận trước khi mở khóa, nếu không cashback thành kênh rút tiền cho fake-account farming (§5.3).

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/affil/migrations/0005_cashback_ledger.sql
CREATE TABLE cashback_entry (
  id            BIGSERIAL PRIMARY KEY,
  user_id       BIGINT NOT NULL REFERENCES app_user(id),
  conversion_id BIGINT NOT NULL REFERENCES affiliate_conversion(id),
  commission    BIGINT NOT NULL CHECK (commission >= 0),   -- VND
  user_share    BIGINT NOT NULL CHECK (user_share >= 0),
  kept_margin   BIGINT NOT NULL CHECK (kept_margin >= 0),
  status        TEXT   NOT NULL CHECK (status IN ('pending','available','paid','clawed_back')),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  available_at  TIMESTAMPTZ NOT NULL,
  paid_at       TIMESTAMPTZ,
  UNIQUE (conversion_id)                                   -- idempotent theo conversion
);
CREATE INDEX idx_cashback_release ON cashback_entry (status, available_at)
  WHERE status = 'pending';

CREATE TABLE payout_request (
  id          BIGSERIAL PRIMARY KEY,
  user_id     BIGINT NOT NULL REFERENCES app_user(id),
  amount      BIGINT NOT NULL CHECK (amount > 0),
  status      TEXT NOT NULL CHECK (status IN ('queued','sent','failed')),
  gateway_ref TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### Tạo entry khi conversion confirmed (Go)

```go
// services/affil/internal/cashback/ledger.go
func (l *Ledger) OnConversionConfirmed(ctx context.Context, c Conversion) error {
    rate := l.shareRate(c.UserTier)                 // phan tram nguyen: free 30, premium 50
    userShare := c.Commission * rate / 100          // BIGINT, chia nguyen (floor); KHONG float
    kept := c.Commission - userShare
    availableAt := c.ConfirmedAt.Add(l.investigationWindow) // 7-14 ngay
    _, err := l.pool.Exec(ctx,
        `INSERT INTO cashback_entry
           (user_id, conversion_id, commission, user_share, kept_margin, status, available_at)
         VALUES ($1,$2,$3,$4,$5,'pending',$6)
         ON CONFLICT (conversion_id) DO NOTHING`,      // §1 #7 idempotent
        c.UserID, c.ID, c.Commission, userShare, kept, availableAt)
    return err
}
```

### Release (pending -> available)

```go
// services/affil/internal/cashback/release.go
func (r *Releaser) ReleaseDue(ctx context.Context, now time.Time) (int, error) {
    rows, _ := r.due(ctx, now) // status=pending AND available_at <= now
    n := 0
    for _, e := range rows {
        if r.fraud.IsFlagged(ctx, e.UserID, e.ConversionID) { // TASK-TRUST-005
            continue // giu pending / cho dieu tra; co the sang clawed_back
        }
        r.setStatus(ctx, e.ID, "available")
        n++
    }
    return n, nil
}
```

---

## §4 - Acceptance criteria

1. Conversion confirmed -> tạo 1 `cashback_entry` status `pending`, `available_at = confirmed_at + window`.
2. Conversion còn `pending` -> KHÔNG tạo cashback_entry.
3. `user_share + kept_margin == commission` (không sai đồng nào).
4. Tạo entry 2 lần cho cùng conversion -> chỉ 1 row (UNIQUE conversion_id, ON CONFLICT DO NOTHING).
5. Trước `available_at` -> entry vẫn `pending`, không vào payout.
6. Sau `available_at` + user không bị flag -> sang `available`.
7. Sau `available_at` nhưng TASK-TRUST-005 flag -> KHÔNG sang `available` (giữ/hủy để điều tra).
8. Conversion bị network thu hồi (confirmed -> rejected) -> entry sang `clawed_back`; nếu chưa paid, trừ khỏi số dư.
9. Tổng `available` của user < ngưỡng -> KHÔNG tạo payout_request.
10. Tổng `available` >= ngưỡng -> gom thành 1 payout_request qua VietQR.
11. Commission < 0 hoặc user_share < 0 -> lỗi CHECK constraint.
12. Endpoint `GET /v1/cashback/summary` trả disclosure rõ cho user (§1 #9): nêu cashback `pending` đến ngày `available_at` và có thể bị hủy nếu đơn bị hủy/hoàn (kiểm field `note` + `next_available_at` trong response).

---

## §5 - Kiểm thử (verification)

```go
func TestCashback_ConfirmedCreatesPending(t *testing.T) {
    l, uid, cid := setup(t)
    err := l.OnConversionConfirmed(ctx, Conversion{ID: cid, UserID: uid, Commission: 100_000, UserTier: "premium", ConfirmedAt: t0})
    require.NoError(t, err)
    e := getEntry(t, l, cid)
    require.Equal(t, "pending", e.Status)
    require.Equal(t, int64(50_000), e.UserShare)   // 50% premium
    require.Equal(t, int64(50_000), e.KeptMargin)
    require.Equal(t, e.Commission, e.UserShare+e.KeptMargin)
}

func TestCashback_Idempotent(t *testing.T) {
    l, uid, cid := setup(t)
    c := Conversion{ID: cid, UserID: uid, Commission: 100_000, UserTier: "free", ConfirmedAt: t0}
    l.OnConversionConfirmed(ctx, c)
    l.OnConversionConfirmed(ctx, c)
    require.Equal(t, 1, countEntries(t, l, cid))
}

func TestRelease_HoldsBeforeWindow(t *testing.T) {
    l, uid, cid := setup(t)
    l.OnConversionConfirmed(ctx, Conversion{ID: cid, UserID: uid, Commission: 100_000, UserTier: "free", ConfirmedAt: t0})
    n, _ := NewReleaser(l).ReleaseDue(ctx, t0.Add(24*time.Hour)) // chua qua window
    require.Equal(t, 0, n)
    require.Equal(t, "pending", getEntry(t, l, cid).Status)
}

func TestRelease_FlaggedUserNotReleased(t *testing.T) {
    l, uid, cid := setupFlagged(t, uid) // TASK-TRUST-005 gan co
    l.OnConversionConfirmed(ctx, Conversion{ID: cid, UserID: uid, Commission: 100_000, UserTier: "free", ConfirmedAt: t0})
    n, _ := NewReleaser(l).ReleaseDue(ctx, t0.Add(30*24*time.Hour))
    require.Equal(t, 0, n) // giu lai du qua window
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration 0005 -> ledger.go (hook từ conversion.confirmed) -> release.go (cron gọi ReleaseDue mỗi giờ) -> payout.go (gom available >= ngưỡng -> VietQR qua TASK-BILL-002) -> tests. Release chạy như job định kỳ; clawback chạy khi nhận postback rejected từ TASK-AFFIL-003.

---

## §7 - Phụ thuộc

- **TASK-AFFIL-003** - nguồn conversion confirmed/rejected (postback network).
- **TASK-BILL-002** - kênh payout VietQR/số dư.
- **TASK-TRUST-005** - tín hiệu gian lận + delay payout trước khi release.
- **TASK-AFFIL-001** - bảng affiliate_conversion (FK).
- Crates: pgx, cron scheduler.

---

## §8 - Payload ví dụ

### User xem cashback

```http
GET /v1/cashback/summary HTTP/1.1
Authorization: Bearer <jwt>

-> 200 OK
{
  "pending":   { "count": 3, "amount_vnd": 142000, "next_available_at": "2026-07-11" },
  "available": { "count": 1, "amount_vnd": 60000 },
  "paid_total_vnd": 380000,
  "payout_threshold_vnd": 50000,
  "note": "Cashback pending co the bi huy neu don bi huy hoac hoan."
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Tăng share_rate theo streak/loyalty (FR-loyalty tương lai).
- Cashback cross-border khi mở SEA (currency theo nước).
- Cho user rút về MoMo/ZaloPay ngoài VietQR - slice sau.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Trả cashback trước confirm | test #2 | mất tiền | Chỉ tạo trên confirmed (DEC-AFFIL-50) |
| Đơn hủy sau confirm | postback rejected | clawback | §1 #5 sang clawed_back, trừ số dư |
| Farm fake-account rút cashback | TASK-TRUST-005 flag | giữ pending | §1 #4 không release khi flagged |
| Tạo trùng entry | UNIQUE conversion_id | DO NOTHING | Theo thiết kế (idempotent) |
| Payout giao dịch nhỏ | ngưỡng tối thiểu | gộp đến ngưỡng | DEC-AFFIL-53 |
| user_share + kept != commission | property test | sai số | floor + trừ, test #3 |
| commission âm | DB CHECK | từ chối | Sửa nguồn postback |
| Release trước window | idx status+available_at | giữ pending | §1 #4 |

---

## §11 - Ghi chú

- Cashback là tính năng user-love rất cao nhưng cũng là bề mặt gian lận cao nhất - hold-then-release là lằn ranh.
- Chỉ tính trên confirmed + clawback khi rejected giữ SănDeal không trả tiền không có thật.
- kept_margin là dòng doanh thu của tính năng; share_rate theo tier là đòn bẩy upsell Premium.
- Delay payout + hỏi TASK-TRUST-005 trước release là best practice chống farm (§5.3).
- Disclosure pending/có-thể-hủy giữ minh bạch hậu-Honey (§4.2).

---

*Hết TASK-AFFIL-005. Status: ready_to_implement (mục tiêu audit 10/10).*
