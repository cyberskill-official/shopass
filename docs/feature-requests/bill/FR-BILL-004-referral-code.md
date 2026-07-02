---
id: FR-BILL-004
title: "`referral_code` (code unique, uses) + attribution + hook chống abuse - mỗi user một mã, gắn người được giới thiệu, chặn tự giới thiệu/farming qua hook anti-fraud"
module: BILL
priority: SHOULD
status: done
verify: T
phase: P2
milestone: P2 - slice 2
slice: 2
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-BILL-001, FR-AUTH-001, FR-TRUST-004, FR-INFRA-002, FR-MOBILE-003]
depends_on: [FR-BILL-001]
blocks: [FR-MOBILE-003, FR-TRUST-004]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.4 (data model: referral_code code unique + uses; app_user.referral_code_id)"
  - "docs/... §5.3 (gian lận: referral abuse, fake account farming -> delay payout, velocity, đồ thị quan hệ)"
source_decisions:
  - "DEC-BILL-16: mỗi user có đúng một referral_code của riêng mình (code TEXT UNIQUE, uses đếm lượt dùng thành công)"
  - "DEC-BILL-17: attribution gắn người được giới thiệu qua app_user.referral_code_id (FR-AUTH-001) tại lúc đăng ký; bất biến sau khi gắn"
  - "DEC-BILL-18: CẤM tự giới thiệu (user dùng mã của chính mình) và CẤM gắn lại (đã có referrer thì không đổi)"
  - "DEC-BILL-19: referral KHÔNG tự trả thưởng ngay; phát một sự kiện cho FR-TRUST-004 (anti-fraud) + delay reward tới khi vượt cửa kiểm tra (best practice §5.3)"
  - "DEC-BILL-20: tăng uses là idempotent theo cặp (referrer, referee) - một người được giới thiệu chỉ tính một lần"

language: "PostgreSQL 16; service Go 1.22 (bill-svc)"
service: shopass/services/bill/
new_files:
  - services/bill/migrations/0004_referral_code.sql
  - services/bill/internal/referral/code.go
  - services/bill/internal/referral/attribute.go
  - services/bill/internal/referral/repo.go
  - services/bill/internal/referral/attribute_test.go
  - services/bill/internal/referral/code_test.go
modified_files: []
allowed_tools:
  - file_read: services/bill/**
  - file_write: services/bill/**
  - bash: cd services/bill && go test ./...
disallowed_tools:
  - cho user tự giới thiệu (dùng mã của chính mình) (vi phạm DEC-BILL-18)
  - cho gắn lại referrer khi đã có (vi phạm DEC-BILL-18, DEC-BILL-17)
  - tự trả thưởng referral ngay không qua hook anti-fraud + delay (vi phạm DEC-BILL-19)
  - đếm uses trùng cho cùng cặp referrer-referee (vi phạm DEC-BILL-20)

effort_hours: 5
sub_tasks:
  - "0.5h: 0004_referral_code.sql - bảng referral_code (user_id, code UNIQUE, uses) + FK; index code"
  - "1.0h: code.go - sinh mã ngắn dễ chia sẻ, không trùng, không gây nhầm (loại ký tự O/0/I/1)"
  - "1.5h: attribute.go - gắn referrer lúc đăng ký: chặn tự giới thiệu, chặn gắn lại, idempotent uses, phát sự kiện anti-fraud"
  - "0.5h: repo.go - CreateCodeForUser, FindByCode, IncrementUses, SetReferrer"
  - "1.0h: attribute_test.go - tự giới thiệu bị chặn; gắn lại bị chặn; uses idempotent; mã lạ bị từ chối"
  - "0.5h: code_test.go - mã duy nhất, không chứa ký tự gây nhầm, độ dài hợp lý"
  - "0.5h: OTel metric referral_attributed_total + referral_self_blocked_total + phát sự kiện referral.attributed"

risk_if_skipped: "referral là đòn bẩy tăng trưởng viral chi phí thấp (§5.7) - CAC qua viral/referral thấp hơn nhiều quảng cáo. Nhưng referral là mục tiêu gian lận hàng đầu (§5.3): nếu cho tự giới thiệu thì một người tạo mã rồi tự dùng để farm thưởng; nếu cho gắn lại referrer thì attribution bị thao túng; nếu trả thưởng ngay không qua kiểm tra thì fake-account farming rút thưởng hàng loạt trước khi bị phát hiện - lỗ trực tiếp. Nếu đếm uses trùng thì thống kê tăng trưởng sai và thưởng nhân đôi. Tài liệu nguồn nêu rõ best practice là delay payout + velocity + đồ thị quan hệ (§5.3); FR này phải phát sự kiện cho anti-fraud (FR-TRUST-004) thay vì tự thưởng. Làm sai biến tính năng tăng trưởng thành lỗ hổng farming."
---

## §1 - Mô tả (BCP-14 normative)

Service BILL **MUST** cấp cho mỗi user một `referral_code` duy nhất, gắn attribution người được giới thiệu lúc đăng ký (bất biến, chặn tự giới thiệu/gắn lại), đếm `uses` idempotent, và phát sự kiện cho anti-fraud thay vì tự trả thưởng. Hợp đồng:

1. **MUST** định nghĩa bảng `referral_code (id, user_id, code, uses, created_at)` với `user_id` REFERENCES `app_user(id)`, `code TEXT NOT NULL UNIQUE`, `uses INTEGER NOT NULL DEFAULT 0`.
2. **MUST** đảm bảo mỗi user có đúng một `referral_code` của riêng mình (DEC-BILL-16): `UNIQUE (user_id)` trên `referral_code` (một mã/người).
3. **MUST** sinh `code` ngắn, dễ chia sẻ, không trùng, và tránh ký tự gây nhầm (loại `O/0/I/1` để tránh đọc/gõ sai khi truyền miệng).
4. **MUST** gắn attribution lúc đăng ký (DEC-BILL-17): khi người mới đăng ký kèm một mã giới thiệu hợp lệ, đặt `app_user.referral_code_id` (FR-AUTH-001) trỏ tới `referral_code` của người giới thiệu. Gắn xong là bất biến.
5. **MUST** CẤM tự giới thiệu (DEC-BILL-18): nếu mã giới thiệu thuộc về chính user đang đăng ký (cùng `user_id`) -> từ chối gắn, không tăng `uses`, trả lỗi xác định `ErrSelfReferral`.
6. **MUST** CẤM gắn lại (DEC-BILL-18): user đã có `referral_code_id` (đã được giới thiệu) -> không đổi sang referrer khác; trả `ErrAlreadyAttributed`.
7. **MUST** từ chối mã không tồn tại: mã giới thiệu không khớp `referral_code` nào -> `ErrUnknownCode`, không gắn, không tăng `uses`.
8. **MUST** tăng `uses` idempotent theo cặp (referrer, referee) (DEC-BILL-20): một người được giới thiệu chỉ làm tăng `uses` của referrer đúng một lần (dựa attribution bất biến ở §1 #4/#6).
9. **MUST NOT** tự trả thưởng referral ngay (DEC-BILL-19): thay vào đó phát một sự kiện `referral.attributed {referrer_id, referee_id, at}` cho FR-TRUST-004 (anti-fraud) + cơ chế delay reward; thưởng chỉ được cấp sau khi vượt cửa kiểm tra (velocity, đồ thị quan hệ, §5.3).
10. **MUST** expose hàm:
    - `CreateCodeForUser(ctx, userID int64) (string, error)` - tạo mã duy nhất cho user (idempotent: gọi lại trả mã đã có).
    - `Attribute(ctx, refereeID int64, code string) error` - gắn referrer lúc đăng ký, áp mọi ràng buộc §1 #4-#9.
    - `FindByCode(ctx, code string) (ReferralCode, bool, error)`.
11. **MUST** coi attribution là một phần của luồng đăng ký (FR-AUTH-001) nhưng KHÔNG chặn đăng ký nếu mã lỗi: mã giới thiệu sai/tự giới thiệu -> user vẫn đăng ký được, chỉ không gắn attribution (đăng ký không phụ thuộc tính hợp lệ của mã giới thiệu).
12. **SHOULD** phát OTel + sự kiện: `referral_attributed_total` (counter), `referral_self_blocked_total` (counter), `referral_unknown_code_total` (counter); sự kiện `referral.attributed` lên bus cho FR-TRUST-004.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao một mã mỗi user + tránh ký tự gây nhầm (DEC-BILL-16, §1 #3)?** Mã giới thiệu được chia sẻ qua tin nhắn, đọc miệng, in trên ảnh. `O` và `0`, `I` và `1` dễ đọc/gõ sai, làm mã hỏng và người giới thiệu mất công. Một mã ngắn không chứa ký tự dễ nhầm tăng tỷ lệ dùng thành công. Một mã mỗi user giữ attribution rõ ràng (mỗi mã trỏ đúng một người giới thiệu).

**Vì sao attribution bất biến (DEC-BILL-17, §1 #6)?** Người giới thiệu một user là một sự thật xảy ra một lần (lúc đăng ký). Cho gắn lại sau đó mở cửa thao túng: user đổi referrer để chuyển thưởng, hoặc bị lừa gắn lại. Bất biến sau khi gắn giữ attribution trung thực và là cơ sở để đếm `uses` không trùng.

**Vì sao cấm tự giới thiệu (DEC-BILL-18)?** Đây là cách farming đơn giản nhất: tạo mã của mình rồi tự dùng để nhận thưởng. So `user_id` của chủ mã với người đăng ký và từ chối khi trùng đóng lỗ hổng này ngay tại nguồn - rẻ hơn nhiều việc phát hiện sau qua anti-fraud.

**Vì sao không tự trả thưởng ngay (DEC-BILL-19)?** Tài liệu nguồn (§5.3) nêu rõ delay payout là best practice ngành chống farming. Trả thưởng tức thì cho phép kẻ tạo hàng loạt tài khoản giả rút thưởng trước khi bị phát hiện. Phát một sự kiện cho anti-fraud (FR-TRUST-004) và chỉ cấp thưởng sau khi qua velocity/đồ thị quan hệ chuyển rủi ro từ "trả rồi mới phát hiện" sang "kiểm rồi mới trả".

**Vì sao uses idempotent theo cặp (DEC-BILL-20)?** Nếu một người được giới thiệu làm tăng `uses` nhiều lần (đăng ký lại, race), thống kê tăng trưởng phình và thưởng nhân đôi. Dựa attribution bất biến (mỗi referee gắn đúng một referrer một lần), việc tăng `uses` cũng chỉ xảy ra một lần cho cặp đó.

**Vì sao mã lỗi không chặn đăng ký (§1 #11)?** Đăng ký là hành động cốt lõi; nó không nên thất bại vì một mã giới thiệu gõ sai. Tách bạch: đăng ký luôn thành công nếu thông tin tài khoản hợp lệ (FR-AUTH-001); attribution là phần thêm, mã lỗi chỉ làm bỏ qua gắn, không cản người dùng vào hệ thống.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/bill/migrations/0004_referral_code.sql
CREATE TABLE referral_code (
  id         BIGSERIAL   PRIMARY KEY,
  user_id    BIGINT      NOT NULL UNIQUE REFERENCES app_user(id), -- một mã/người (§1 #2)
  code       TEXT        NOT NULL UNIQUE,                          -- §1 #1
  uses       INTEGER     NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_referral_code_lookup ON referral_code (code);
-- app_user.referral_code_id (FR-AUTH-001) trỏ tới referral_code(id) của người giới thiệu;
-- FK đó được thêm khi cả hai bảng tồn tại.
```

### Sinh mã (Go)

```go
// services/bill/internal/referral/code.go
// alphabet loại ký tự gây nhầm (O/0/I/1) để truyền miệng/gõ không sai (§1 #3).
const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func NewCode(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = alphabet[randIndex(len(alphabet))]
    }
    return "SD" + string(b) // tiền tố thương hiệu: ví dụ "SD7KQ2M9"
}
```

### Attribution (Go)

```go
// services/bill/internal/referral/attribute.go

// Attribute gắn referrer cho người mới đăng ký, áp mọi ràng buộc (§1 #4-#9).
func (s *Service) Attribute(ctx context.Context, refereeID int64, code string) error {
    rc, ok, err := s.repo.FindByCode(ctx, code)
    if err != nil { return err }
    if !ok { return ErrUnknownCode }                 // §1 #7
    if rc.UserID == refereeID { return ErrSelfReferral } // §1 #5 - tự giới thiệu

    already, err := s.repo.HasReferrer(ctx, refereeID)
    if err != nil { return err }
    if already { return ErrAlreadyAttributed }       // §1 #6 - gắn lại

    if err := s.repo.SetReferrer(ctx, refereeID, rc.ID); err != nil {
        return err
    }
    s.repo.IncrementUses(ctx, rc.ID)                 // idempotent theo cặp (§1 #8)
    s.bus.Publish(ctx, ReferralAttributed{           // §1 #9 - không tự thưởng
        ReferrerID: rc.UserID, RefereeID: refereeID, At: time.Now(),
    })
    metrics.Attributed()
    return nil
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> `referral_code` tồn tại với `code` UNIQUE + `user_id` UNIQUE.
2. `CreateCodeForUser(u)` -> mã duy nhất, bắt đầu `SD`, không chứa `O/0/I/1`; gọi lại cho cùng user trả mã đã có (idempotent).
3. Hai user khác nhau -> hai mã khác nhau (UNIQUE code).
4. `Attribute(referee, code_of_referrer)` hợp lệ -> `app_user(referee).referral_code_id` trỏ tới mã referrer; `uses` của referrer tăng 1.
5. `Attribute(user, code_của_chính_user)` -> `ErrSelfReferral`, `uses` KHÔNG tăng.
6. `Attribute` khi referee đã có referrer -> `ErrAlreadyAttributed`, không đổi.
7. `Attribute` với mã không tồn tại -> `ErrUnknownCode`, không gắn, không tăng `uses`.
8. Gọi `Attribute` cùng cặp (referrer, referee) hai lần -> `uses` chỉ tăng một lần (lần hai trả `ErrAlreadyAttributed`).
9. `Attribute` thành công phát sự kiện `referral.attributed` (kiểm bus mock) - KHÔNG tự cộng thưởng/tiền.
10. Đăng ký với mã giới thiệu sai (tự giới thiệu/mã lạ) -> user VẪN đăng ký được (đăng ký không phụ thuộc mã), chỉ không gắn attribution.
11. `FindByCode(code)` trả đúng `referral_code`; mã không tồn tại -> `found=false`.
12. Metric `referral_self_blocked_total` tăng khi chặn tự giới thiệu; `referral_attributed_total` tăng khi gắn thành công.

---

## §5 - Kiểm thử (verification)

```go
// services/bill/internal/referral/attribute_test.go
func TestAttribute_Valid(t *testing.T) {
    s, referrer, referee := setupTwoUsers(t)
    code, _ := s.repo.CodeOf(ctx, referrer)
    require.NoError(t, s.Attribute(ctx, referee, code))
    require.Equal(t, 1, s.repo.UsesOf(ctx, code))
    require.True(t, s.repo.HasReferrerBool(ctx, referee))
}

func TestAttribute_SelfReferral_Blocked(t *testing.T) {
    s, u, _ := setupTwoUsers(t)
    code, _ := s.repo.CodeOf(ctx, u)
    require.ErrorIs(t, s.Attribute(ctx, u, code), ErrSelfReferral) // dùng mã của chính mình
    require.Equal(t, 0, s.repo.UsesOf(ctx, code)) // uses không tăng
}

func TestAttribute_AlreadyAttributed_Blocked(t *testing.T) {
    s, referrer, referee := setupTwoUsers(t)
    code, _ := s.repo.CodeOf(ctx, referrer)
    require.NoError(t, s.Attribute(ctx, referee, code))
    require.ErrorIs(t, s.Attribute(ctx, referee, code), ErrAlreadyAttributed) // gắn lại
    require.Equal(t, 1, s.repo.UsesOf(ctx, code)) // không tăng lần hai
}

func TestAttribute_UnknownCode(t *testing.T) {
    s, _, referee := setupTwoUsers(t)
    require.ErrorIs(t, s.Attribute(ctx, referee, "SDNOPE99"), ErrUnknownCode)
}

func TestAttribute_PublishesEvent_NoDirectReward(t *testing.T) {
    s, referrer, referee := setupTwoUsers(t)
    code, _ := s.repo.CodeOf(ctx, referrer)
    s.Attribute(ctx, referee, code)
    require.Equal(t, 1, s.bus.CountOf("referral.attributed")) // phát sự kiện
    require.Equal(t, 0, s.rewardsGranted())                   // KHÔNG tự thưởng (§1 #9)
}
```

```go
// services/bill/internal/referral/code_test.go
func TestNewCode_NoConfusingChars(t *testing.T) {
    seen := map[string]bool{}
    for i := 0; i < 1000; i++ {
        c := NewCode(6)
        require.False(t, seen[c]); seen[c] = true
        require.NotContains(t, c[2:], "O")
        require.NotContains(t, c[2:], "0")
        require.NotContains(t, c[2:], "I")
        require.NotContains(t, c[2:], "1")
    }
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration `0004_referral_code.sql` (code UNIQUE + user_id UNIQUE) -> `code.go` (sinh mã từ alphabet loại ký tự nhầm) -> `repo.go` (`CreateCodeForUser`, `FindByCode`, `IncrementUses`, `SetReferrer`) -> `attribute.go` (gắn referrer + áp ràng buộc + phát sự kiện) -> tests. `Attribute` được gọi từ luồng đăng ký FR-AUTH-001 khi có mã giới thiệu, nhưng lỗi của nó không lăn ngược đăng ký (§1 #11). Sự kiện `referral.attributed` lên bus (cùng hạ tầng fan-out FR-NOTIF-003 hoặc bus nội bộ) cho FR-TRUST-004 tiêu thụ. FK `app_user.referral_code_id -> referral_code(id)` thêm khi cả hai bảng tồn tại.

---

## §7 - Phụ thuộc

- **FR-BILL-001** - module BILL + migration nền; referral sống cùng bill-svc.
- **FR-AUTH-001** - `app_user.referral_code_id` (cột để gắn attribution); luồng đăng ký gọi `Attribute`.
- **FR-INFRA-002** - `app_user` + quy ước migration.
- **FR-TRUST-004 (downstream)** - tiêu thụ sự kiện `referral.attributed`, áp velocity/đồ thị quan hệ, quyết định trả thưởng (delay payout).
- **FR-MOBILE-003 (downstream, P3)** - share-on-sale + referral dùng mã này cho virality.
- Lib: `pgx`, `crypto/rand` (sinh mã).

---

## §8 - Payload ví dụ

### Đăng ký kèm mã giới thiệu (luồng FR-AUTH-001 gọi Attribute)

```go
userID, _ := authSvc.Register(ctx, in) // đăng ký thành công trước
if in.ReferralCode != "" {
    if err := referral.Attribute(ctx, userID, in.ReferralCode); err != nil {
        log.Info("referral not attributed", "err", err) // không lăn ngược đăng ký (§1 #11)
    }
}
```

### Sự kiện phát cho anti-fraud (không tự thưởng)

```json
{
  "type": "referral.attributed",
  "referrer_id": 7,
  "referee_id": 4096,
  "at": "2026-06-28T10:15:00+07:00"
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Cơ chế thưởng cụ thể (giảm giá Premium, ngày dùng thử) khi referral qua cửa anti-fraud - định ở FR-TRUST-004 + FR-BILL-005; FR này chỉ attribution + sự kiện.
- Mã giới thiệu nhiều tầng (referee giới thiệu tiếp được chia) - rủi ro pyramid; cân nhắc kỹ, hoãn.
- Hết hạn mã / mã chiến dịch riêng - thêm cột `expires_at`/`campaign` khi cần; schema tối thiểu hiện tại đủ.
- Giới hạn số referral/người/ngày - đặt ở FR-TRUST-004 (velocity), không ở schema.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Tự giới thiệu farm thưởng | self-referral test | farming thưởng | So user_id chủ mã vs referee (§1 #5) |
| Gắn lại đổi referrer | already-attributed test | thao túng attribution | Bất biến sau khi gắn (§1 #6) |
| Tự trả thưởng ngay | publish-event test | fake-account rút thưởng | Phát sự kiện + delay (DEC-BILL-19) |
| uses tăng trùng | idempotent test | thống kê sai, thưởng nhân đôi | Idempotent theo cặp (§1 #8) |
| Mã gây nhầm O/0/I/1 | code test | mã hỏng khi gõ/đọc | Loại ký tự nhầm (§1 #3) |
| Mã lạ | unknown-code test | gắn rác | ErrUnknownCode (§1 #7) |
| Mã lỗi chặn đăng ký | register-not-blocked test | mất user vì gõ sai mã | Mã lỗi không lăn ngược đăng ký (§1 #11) |
| Hai mã/người | UNIQUE(user_id) | attribution mơ hồ | Một mã/người (§1 #2) |
| Race tăng uses | idempotent + attribution bất biến | đếm trùng | Dựa attribution một-lần (§1 #8) |

---

## §11 - Ghi chú

- referral là đòn bẩy viral chi phí thấp (§5.7) nhưng là mục tiêu gian lận hàng đầu (§5.3) - thiết kế phải phòng farming từ đầu.
- Cấm tự giới thiệu + bất biến attribution đóng hai lỗ hổng farming phổ biến nhất ngay tại nguồn.
- Không tự trả thưởng: phát sự kiện cho anti-fraud (FR-TRUST-004) + delay payout chuyển từ "trả rồi phát hiện" sang "kiểm rồi trả".
- uses idempotent theo cặp giữ thống kê tăng trưởng đúng và thưởng không nhân đôi.
- Mã loại ký tự gây nhầm (O/0/I/1) tăng tỷ lệ dùng thành công khi truyền miệng/in ảnh.
- Mã lỗi không chặn đăng ký: attribution là phần thêm, không phải điều kiện vào hệ thống.
- Đây là nguồn dữ liệu cho FR-TRUST-004 (anti-fraud) và virality FR-MOBILE-003.

---

*Hết FR-BILL-004. Status: ready_to_implement (mục tiêu audit 10/10).*
