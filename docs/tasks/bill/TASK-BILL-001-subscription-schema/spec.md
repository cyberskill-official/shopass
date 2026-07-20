---
id: TASK-BILL-001
title: "Schema `subscription` + tier (Premium 29k/49k/79k VND/tháng) + vòng đời - started_at/renews_at, status; giá lưu BIGINT VND, một subscription active mỗi user"
module: BILL
priority: MUST
status: done
verify: T
phase: P2
milestone: P2 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-AUTH-001, TASK-BILL-002, TASK-BILL-003, TASK-BILL-004, TASK-BILL-005, TASK-INFRA-002]
depends_on: [TASK-AUTH-001]
blocks: [TASK-B2B-002, TASK-BILL-002, TASK-BILL-004, TASK-BILL-005]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.4 (data model: subscription tier/price/started_at/renews_at/status)"
  - "docs/... §4.3 (pricing: Premium 29k/49k/79k VND/tháng theo tier; free-tier mạnh, Premium nhẹ + gamified)"
source_decisions:
  - "DEC-BILL-01: subscription.tier - {'free','premium_basic','premium_plus','premium_pro'}; ba bậc Premium ứng giá 29k/49k/79k VND/tháng"
  - "DEC-BILL-02: giá lưu BIGINT (VND, không thập phân) đồng nhất DEC-PRICE-05; KHÔNG float"
  - "DEC-BILL-03: mỗi user có TỐI ĐA một subscription active tại một thời điểm (partial unique index trên status='active')"
  - "DEC-BILL-04: vòng đời status - {'active','past_due','canceled','expired'}; renews_at là mốc gia hạn kế (thanh toán thực do TASK-BILL-002/003 đẩy)"
  - "DEC-BILL-05: giá tier lưu ở bảng tham chiếu plan_catalog (tier, price, billing_period) - subscription tham chiếu plan, không hardcode giá rải rác"

language: "PostgreSQL 16; service Go 1.22 (bill-svc)"
service: shopass/services/bill/
new_files:
  - services/bill/migrations/0001_plan_catalog.sql
  - services/bill/migrations/0002_subscription.sql
  - services/bill/internal/bill/types.go
  - services/bill/internal/bill/repo.go
  - services/bill/internal/bill/lifecycle.go
  - services/bill/internal/bill/repo_test.go
  - services/bill/internal/bill/lifecycle_test.go
modified_files: []
allowed_tools:
  - file_read: services/bill/**
  - file_write: services/bill/**
  - bash: cd services/bill && go test ./...
disallowed_tools:
  - lưu giá dạng float/numeric thập phân (vi phạm DEC-BILL-02)
  - cho một user có nhiều subscription active đồng thời (vi phạm DEC-BILL-03)
  - hardcode giá 29k/49k/79k rải rác thay vì plan_catalog (vi phạm DEC-BILL-05)
  - cho status ngoài tập định nghĩa (vi phạm DEC-BILL-04)

effort_hours: 6
sub_tasks:
  - "0.5h: 0001_plan_catalog.sql - bảng plan_catalog (tier, price BIGINT, billing_period) + seed free/3 bậc Premium"
  - "1.0h: 0002_subscription.sql - bảng subscription + FK user/plan + CHECK status + partial unique active + CHECK renews_at > started_at"
  - "1.0h: types.go + repo.go - CreateSubscription, GetActive, UpdateStatus, SetRenewsAt"
  - "1.0h: lifecycle.go - chuyển trạng thái hợp lệ (active->past_due->canceled/expired); chặn chuyển không hợp lệ"
  - "1.0h: repo_test.go - tạo sub, một active mỗi user (partial unique), giá từ plan_catalog, CHECK status"
  - "1.0h: lifecycle_test.go - active->past_due->active (thanh toán lại); canceled là cuối; chuyển lạ bị chặn"
  - "0.5h: OTel metric subscription_active_total{tier} (gauge) + subscription_status_change_total{from,to}"

risk_if_skipped: "subscription là sổ gốc của dòng doanh thu Premium - một trong ba dòng tiền (§4.1). Không có nó thì payment gateway (TASK-BILL-002) không có thực thể để gắn thanh toán, reconciliation (TASK-BILL-003) không có chỗ đối soát, và feature gating (TASK-BILL-005) không biết user ở tier nào. Nếu giá lưu float thì sai số trên báo cáo doanh thu và đối soát. Nếu một user có nhiều subscription active đồng thời thì tính tiền trùng và quyền lợi rối - một lỗi dữ liệu khó gỡ. Nếu hardcode 29k/49k/79k rải rác thì đổi giá phải sửa nhiều nơi, dễ lệch. Nếu status không kiểm soát thì subscription rơi vào trạng thái rác (vừa active vừa canceled). Đây là nền dữ liệu cho toàn bộ billing."
---

## §1 - Mô tả (BCP-14 normative)

Service BILL **MUST** định nghĩa `plan_catalog` (bậc giá Premium) và `subscription` (đăng ký của user), với giá BIGINT VND, một subscription active mỗi user, và vòng đời status có kiểm soát. Hợp đồng:

1. **MUST** định nghĩa bảng `plan_catalog (id, tier, price, billing_period, active)`: `tier` - {`'free'`,`'premium_basic'`,`'premium_plus'`,`'premium_pro'`} (CHECK); ba bậc Premium ứng giá 29.000 / 49.000 / 79.000 VND/tháng (DEC-BILL-01).
2. **MUST** lưu `plan_catalog.price` dạng `BIGINT` (VND, không thập phân) - KHÔNG float/numeric (DEC-BILL-02); `free` có `price = 0`.
3. **MUST** định nghĩa bảng `subscription (id, user_id, plan_id, started_at, renews_at, status)` với `user_id` REFERENCES `app_user(id)`, `plan_id` REFERENCES `plan_catalog(id)`.
4. **MUST** ràng buộc mỗi user có TỐI ĐA một subscription `status='active'` tại một thời điểm (DEC-BILL-03) qua partial unique index `UNIQUE (user_id) WHERE status='active'`.
5. **MUST** ràng buộc `subscription.status` - {`'active'`,`'past_due'`,`'canceled'`,`'expired'`} qua CHECK (DEC-BILL-04).
6. **MUST** ràng buộc `renews_at > started_at` qua CHECK (chu kỳ gia hạn phải ở tương lai so với điểm bắt đầu).
7. **MUST** không hardcode giá Premium trong code nghiệp vụ (DEC-BILL-05): `subscription` tham chiếu `plan_id`; giá đọc từ `plan_catalog`. Đổi giá là cập nhật `plan_catalog`, không sửa code rải rác.
8. **MUST** expose hàm repo:
- `CreateSubscription(ctx, userID, planID int64, renewsAt time.Time) (int64, error)` - tạo subscription `active`; lỗi nếu user đã có active (partial unique).
- `GetActive(ctx, userID int64) (Subscription, bool, error)` - lấy subscription active hiện tại (nếu có).
- `UpdateStatus(ctx, subID int64, to string) error` - chuyển trạng thái (qua `lifecycle.go`, chỉ chuyển hợp lệ).
- `SetRenewsAt(ctx, subID int64, t time.Time) error` - đẩy mốc gia hạn (do TASK-BILL-002/003 gọi sau thanh toán).
9. **MUST** thực thi chuyển trạng thái hợp lệ (`lifecycle.go`): `active -> past_due` (trễ hạn), `past_due -> active` (thanh toán lại), `active|past_due -> canceled` (user hủy), `past_due -> expired` (quá hạn ân hạn). Chuyển không hợp lệ (ví dụ `canceled -> active`) -> lỗi, không đổi.
10. **MUST** seed `plan_catalog` idempotent với 4 bậc (free + 3 Premium) qua `ON CONFLICT (tier) DO NOTHING`.
11. **MUST** đặt mọi cột thời gian kiểu `TIMESTAMPTZ`; `started_at` mặc định `now()` khi tạo.
12. **SHOULD** phát OTel: `subscription_active_total{tier}` (gauge), `subscription_status_change_total{from,to}` (counter), `subscription_created_total{tier}` (counter).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao giá ở plan_catalog, không hardcode (DEC-BILL-05)?** Giá Premium 29k/49k/79k là quyết định kinh doanh có thể đổi (khuyến mãi, điều chỉnh theo thị trường). Nếu rải literal `29000` khắp code, một lần đổi giá phải sửa nhiều nơi và dễ bỏ sót, gây tính sai. Một bảng `plan_catalog` là nguồn sự thật: subscription tham chiếu `plan_id`, giá đọc từ bảng. Đổi giá là cập nhật dữ liệu, không phải sửa code.

**Vì sao giá BIGINT VND (DEC-BILL-02)?** Giá VN là số nguyên đồng. Float gây sai số khi cộng dồn doanh thu qua hàng nghìn subscription để báo cáo và đối soát với payment gateway. BIGINT chính xác tuyệt đối, đồng nhất với `price_snapshot` (DEC-PRICE-05) và `affiliate_conversion` (DEC-AFFIL-04).

**Vì sao một subscription active mỗi user (DEC-BILL-03)?** Một người chỉ nên có một gói Premium đang hiệu lực. Nếu cho nhiều active đồng thời, ta tính tiền trùng và quyền lợi nhập nhằng (theo gói nào?). Partial unique index `WHERE status='active'` cho phép giữ lịch sử nhiều subscription cũ (canceled/expired) nhưng chỉ một active - vừa sạch vừa giữ vết.

**Vì sao vòng đời status có kiểm soát (DEC-BILL-04, §1 #9)?** Subscription đi qua các trạng thái có ý nghĩa kinh doanh: đang chạy, trễ hạn (chờ thanh toán lại), bị hủy, hết hạn. Cho phép chuyển tùy ý dễ tạo trạng thái vô lý (`canceled` rồi `active` lại mà không qua thanh toán). Một máy trạng thái với các chuyển hợp lệ giữ dữ liệu nhất quán và phản ánh đúng quy trình billing.

**Vì sao renews_at tách khỏi thanh toán thực (DEC-BILL-04)?** `renews_at` là mốc dự kiến gia hạn kế. Thanh toán thực (thành công/thất bại) do TASK-BILL-002/003 xử lý rồi đẩy `renews_at` tới và đặt status phù hợp. Tách "lịch gia hạn" (BILL-001) khỏi "thực thi thanh toán" (BILL-002/003) giữ schema này đơn giản và để phần tích hợp gateway ở đúng chỗ.

**Vì sao seed idempotent (§1 #10)?** `plan_catalog` là dữ liệu tham chiếu cố định (4 bậc). Seed qua `ON CONFLICT (tier) DO NOTHING` cho phép chạy migration/seed nhiều lần (mọi môi trường) mà không nhân đôi bậc giá - đồng nhất cách seed `platform` của TASK-INFRA-002.

---

## §3 - Hợp đồng API / DDL

### Migrations

```sql
-- services/bill/migrations/0001_plan_catalog.sql
CREATE TABLE plan_catalog (
  id             SMALLSERIAL PRIMARY KEY,
  tier           TEXT     NOT NULL UNIQUE
                   CHECK (tier IN ('free','premium_basic','premium_plus','premium_pro')),
  price          BIGINT   NOT NULL CHECK (price >= 0),   -- VND, không thập phân
  billing_period TEXT     NOT NULL DEFAULT 'monthly',
  active         BOOLEAN  NOT NULL DEFAULT true
);

INSERT INTO plan_catalog (tier, price) VALUES
  ('free',          0),
  ('premium_basic', 29000),
  ('premium_plus',  49000),
  ('premium_pro',   79000)
ON CONFLICT (tier) DO NOTHING;

-- services/bill/migrations/0002_subscription.sql
CREATE TABLE subscription (
  id         BIGSERIAL   PRIMARY KEY,
  user_id    BIGINT      NOT NULL REFERENCES app_user(id),
  plan_id    SMALLINT    NOT NULL REFERENCES plan_catalog(id),
  started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  renews_at  TIMESTAMPTZ NOT NULL,
  status     TEXT        NOT NULL DEFAULT 'active'
               CHECK (status IN ('active','past_due','canceled','expired')),
  CHECK (renews_at > started_at)
);

-- Một user tối đa một subscription active (DEC-BILL-03)
CREATE UNIQUE INDEX uq_sub_active_user
  ON subscription (user_id) WHERE status = 'active';
```

### Types (Go)

```go
// services/bill/internal/bill/types.go
type Subscription struct {
    ID        int64     `db:"id"`
    UserID    int64     `db:"user_id"`
    PlanID    int16     `db:"plan_id"`
    StartedAt time.Time `db:"started_at"`
    RenewsAt  time.Time `db:"renews_at"`
    Status    string    `db:"status"`
}
```

### Lifecycle - chuyển trạng thái hợp lệ (Go)

```go
// services/bill/internal/bill/lifecycle.go

// valid map từ trạng thái hiện tại sang tập trạng thái được phép chuyển tới (§1 #9).
var valid = map[string]map[string]bool{
    "active":   {"past_due": true, "canceled": true},
    "past_due": {"active": true, "canceled": true, "expired": true},
    "canceled": {}, // cuối
    "expired":  {}, // cuối
}

func CanTransition(from, to string) bool { return valid[from][to] }

func (r *Repo) UpdateStatus(ctx context.Context, subID int64, to string) error {
    cur, err := r.statusOf(ctx, subID)
    if err != nil {
        return err
    }
    if !CanTransition(cur, to) {
        return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, cur, to)
    }
    _, err = r.pool.Exec(ctx, `UPDATE subscription SET status=$1 WHERE id=$2`, to, subID)
    metrics.StatusChange(cur, to)
    return err
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> `plan_catalog` (4 bậc) và `subscription` tồn tại với FK đúng.
2. Seed -> `plan_catalog` có đúng `premium_basic=29000`, `premium_plus=49000`, `premium_pro=79000`, `free=0`.
3. Seed chạy lần hai -> vẫn 4 bậc (ON CONFLICT DO NOTHING).
4. `CreateSubscription(user, premium_basic_plan, renews)` -> subscription `active`, `started_at` đặt, `renews_at > started_at`.
5. Tạo subscription active thứ hai cho cùng user (khi đã có active) -> vi phạm partial unique (`uq_sub_active_user`).
6. `INSERT subscription` với `status='trialing'` (ngoài tập) -> lỗi CHECK.
7. `INSERT` với `renews_at <= started_at` -> lỗi CHECK.
8. `GetActive(user)` trả đúng subscription active; user không có active -> `found=false`.
9. `UpdateStatus(active -> past_due)` hợp lệ; `UpdateStatus(past_due -> active)` hợp lệ (thanh toán lại).
10. `UpdateStatus(canceled -> active)` -> `ErrInvalidTransition`, status không đổi.
11. Sau khi một subscription chuyển `canceled`, có thể tạo subscription active mới cho cùng user (partial unique chỉ chặn active).
12. Metric `subscription_status_change_total{from,to}` tăng đúng nhãn khi chuyển trạng thái.

---

## §5 - Kiểm thử (verification)

```go
// services/bill/internal/bill/repo_test.go
func TestCreate_Active(t *testing.T) {
    r, uid := setupWithUser(t)
    id, err := r.CreateSubscription(ctx, uid, planID(t, r, "premium_basic"), time.Now().AddDate(0,1,0))
    require.NoError(t, err)
    sub, ok, _ := r.GetActive(ctx, uid)
    require.True(t, ok)
    require.Equal(t, id, sub.ID)
    require.Equal(t, "active", sub.Status)
}

func TestCreate_OneActivePerUser(t *testing.T) {
    r, uid := setupWithUser(t)
    renews := time.Now().AddDate(0,1,0)
    _, err1 := r.CreateSubscription(ctx, uid, planID(t, r, "premium_basic"), renews)
    require.NoError(t, err1)
    _, err2 := r.CreateSubscription(ctx, uid, planID(t, r, "premium_plus"), renews)
    require.Error(t, err2) // partial unique WHERE status='active'
}

func TestPlanCatalog_Prices(t *testing.T) {
    r := setup(t)
    require.Equal(t, int64(29000), priceOf(t, r, "premium_basic"))
    require.Equal(t, int64(49000), priceOf(t, r, "premium_plus"))
    require.Equal(t, int64(79000), priceOf(t, r, "premium_pro"))
}

func TestSubscription_StatusCheck(t *testing.T) {
    r, uid := setupWithUser(t)
    _, err := r.pool.Exec(ctx,
        `INSERT INTO subscription (user_id, plan_id, renews_at, status)
         VALUES ($1,$2, now()+interval '1 month', 'trialing')`, uid, planID(t, r, "free"))
    require.Error(t, err) // CHECK status IN (...)
}

func TestSubscription_RenewsAfterStart(t *testing.T) {
    r, uid := setupWithUser(t)
    _, err := r.pool.Exec(ctx,
        `INSERT INTO subscription (user_id, plan_id, started_at, renews_at)
         VALUES ($1,$2, now(), now()-interval '1 day')`, uid, planID(t, r, "premium_basic"))
    require.Error(t, err) // CHECK renews_at > started_at
}
```

```go
// services/bill/internal/bill/lifecycle_test.go
func TestTransition_Valid(t *testing.T) {
    require.True(t, CanTransition("active", "past_due"))
    require.True(t, CanTransition("past_due", "active"))
    require.True(t, CanTransition("active", "canceled"))
}

func TestTransition_Invalid(t *testing.T) {
    require.False(t, CanTransition("canceled", "active")) // canceled là cuối
    require.False(t, CanTransition("expired", "active"))
}

func TestUpdateStatus_RejectsInvalid(t *testing.T) {
    r, uid := setupWithUser(t)
    id, _ := r.CreateSubscription(ctx, uid, planID(t, r, "premium_basic"), time.Now().AddDate(0,1,0))
    require.NoError(t, r.UpdateStatus(ctx, id, "canceled"))
    require.ErrorIs(t, r.UpdateStatus(ctx, id, "active"), ErrInvalidTransition) // canceled -> active cấm
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration `0001_plan_catalog.sql` (4 bậc + seed) -> `0002_subscription.sql` (bảng + partial unique active + CHECK) -> `types.go` + `repo.go` (`CreateSubscription`, `GetActive`, `SetRenewsAt`) -> `lifecycle.go` (máy trạng thái `CanTransition` + `UpdateStatus`) -> tests trên Postgres ephemeral (kiểm partial unique + CHECK thật, không mock). TASK-BILL-002 gọi `SetRenewsAt`/`UpdateStatus` sau thanh toán; TASK-BILL-005 đọc `GetActive` cho feature gating. Driver `pgx`.

---

## §7 - Phụ thuộc

- **TASK-AUTH-001** - `app_user` phải tồn tại trước (FK `user_id`).
- **TASK-INFRA-002** - bảng `app_user` cột lõi + quy ước đặt tên/migration.
- **TASK-BILL-002 (downstream)** - payment gateway gắn thanh toán vào subscription, đẩy `renews_at`, đổi status.
- **TASK-BILL-003 (downstream)** - reconciliation đối soát payment với subscription.
- **TASK-BILL-004 (downstream)** - referral_code gắn vào user khi tạo subscription (attribution).
- **TASK-BILL-005 (downstream)** - feature gating đọc `GetActive` để biết tier.
- Lib: `pgx`.

---

## §8 - Payload ví dụ

### Tạo subscription (nội bộ, do TASK-BILL-002 gọi sau thanh toán đầu)

```go
subID, err := billRepo.CreateSubscription(ctx, userID,
    planID, // premium_basic
    time.Now().AddDate(0, 1, 0)) // gia hạn sau 1 tháng
```

### Bậc giá trong plan_catalog

```sql
SELECT tier, price FROM plan_catalog ORDER BY price;
--  free          | 0
--  premium_basic | 29000
--  premium_plus  | 49000
--  premium_pro   | 79000
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Chu kỳ năm (billing_period='yearly') với giá ưu đãi - thêm bậc vào `plan_catalog` khi có chiến lược; schema đã sẵn `billing_period`.
- Proration khi nâng/hạ bậc giữa kỳ - thêm logic ở TASK-BILL-002 khi cần; task này chỉ giữ một active.
- Trial period (dùng thử miễn phí có thời hạn) - thêm trạng thái `'trialing'` vào CHECK + máy trạng thái khi làm gamified upgrade (TASK-BILL-005).
- Quyền lợi chi tiết per-tier (bảng `plan_feature`) - gắn vào TASK-BILL-005 (feature gating).

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Giá lưu float | review + types int64 | sai số doanh thu/đối soát | BIGINT VND (DEC-BILL-02) |
| Nhiều subscription active/user | partial unique test | tính tiền trùng, quyền lợi rối | UNIQUE WHERE active (§1 #4) |
| Hardcode giá rải rác | review | đổi giá lệch nhiều nơi | plan_catalog nguồn sự thật (§1 #7) |
| status ngoài tập | DB CHECK | trạng thái rác | CHECK status IN (...) (§1 #5) |
| renews_at <= started_at | DB CHECK | chu kỳ vô lý | CHECK renews_at > started_at (§1 #6) |
| Chuyển trạng thái vô lý | lifecycle test | canceled rồi active không qua thanh toán | Máy trạng thái CanTransition (§1 #9) |
| Seed nhân đôi bậc | ON CONFLICT | giữ 4 bậc | Idempotent seed (§1 #10) |
| FK user không tồn tại | lỗi pgx | từ chối tạo | Tạo app_user trước (TASK-AUTH-001) |
| Mất lịch sử subscription cũ | partial unique giữ vết | - | Chỉ chặn active, giữ canceled/expired (§1 #4) |

---

## §11 - Ghi chú

- `subscription` + `plan_catalog` là sổ gốc của dòng doanh thu Premium (§4.1).
- Giá ở `plan_catalog` (không hardcode) cho phép đổi giá bằng cập nhật dữ liệu, không sửa code rải rác.
- Giá BIGINT VND đồng nhất với price_snapshot và affiliate_conversion - tránh sai số float trên báo cáo doanh thu.
- Partial unique `WHERE status='active'` cho một active mỗi user mà vẫn giữ lịch sử subscription cũ.
- Máy trạng thái giữ subscription nhất quán: chỉ chuyển hợp lệ, không có trạng thái vô lý.
- `renews_at` là lịch gia hạn; thanh toán thực do TASK-BILL-002/003 đẩy - tách lịch khỏi thực thi.
- Đây là nền cho TASK-BILL-002 (thanh toán), TASK-BILL-003 (đối soát), TASK-BILL-005 (gating theo tier).

---

*Hết TASK-BILL-001. Status: ready_to_implement (mục tiêu audit 10/10).*
