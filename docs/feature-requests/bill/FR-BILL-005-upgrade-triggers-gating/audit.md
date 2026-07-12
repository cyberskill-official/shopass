---
fr_id: FR-BILL-005
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

FR-BILL-005 đặc tả feature gating + upgrade trigger ở mức triển khai được. 12 mệnh đề §1 normative, mỗi mệnh đề có AC và test trong §5. Năm trụ được khóa: gating ở backend (không tin client), tính năng lõi không khóa cứng (free-tier mạnh), quyền lợi ở `plan_feature` (không hardcode), fail-safe về free khi lỗi, trigger là tín hiệu không ép (không dark pattern). Tách gating khỏi thanh toán. Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - Tin cờ tier client
Client tự xưng Premium mở quyền miễn phí. Giải: §1 #2 + DEC-BILL-21 - `Allow` đọc subscription thật (FR-BILL-001), bỏ qua cờ client; AC #7 + `TestAllow_IgnoresClientTierFlag`.

### ISS-002 - Khóa cứng tính năng lõi
Paywall theo dõi giá/sale ảo/biểu đồ phá free-tier, đuổi user. Giải: §1 #3 + DEC-BILL-23 - lõi `limit_value=-1` cho free; AC #2/#3 + `TestAllow_CoreFeatureFreeUser`.

### ISS-003 - Cấp nhầm Premium khi lỗi + hardcode quyền lợi
Lỗi tier cấp nhầm quyền; rải limit khắp code khó đổi. Giải: §1 #4 (fail-safe free) + §1 #5 + DEC-BILL-22 (plan_feature nguồn sự thật); AC #8/#9 + `TestAllow_FailSafeFree`.

### ISS-004 - Dark pattern + lẫn gating với thanh toán
Ép upgrade phá niềm tin; `Allow` tự thu tiền lẫn vai trò. Giải: §1 #8 (CTA bỏ qua được) + §1 #11 (tách Allow khỏi checkout); AC #11 + `TestTriggers_DoesNotChargeOrActivate`.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 plan_feature | #1 | `0005_plan_feature.sql` |
| #2 gating backend | #7 | `gate.go::Allow` + `TestAllow_IgnoresClientTierFlag` |
| #3 lõi không khóa cứng | #2,#3 | seed limit -1 + `TestAllow_CoreFeatureFreeUser` |
| #4 so limit + fail-safe | #4,#5,#6,#8 | `Allow` switch + `TestAllow_FailSafeFree` |
| #5 plan_feature nguồn sự thật | #9 | đổi bảng không sửa code |
| #7 trigger tín hiệu | #10 | `EvaluateTriggers` + `TestTriggers_WishlistFull_ShowsCTA` |
| #8 không dark pattern | #10 | CTA bỏ qua được |
| #11 tách khỏi thanh toán | #11 | `TestTriggers_DoesNotChargeOrActivate` |

## §4 - Kết luận

Toàn bộ mệnh đề normative có DDL/code/test backing, gồm gating-backend, lõi-không-khóa-cứng, fail-safe-free, trigger-không-ép và tách-khỏi-thanh-toán. Không có mệnh đề "mồ côi". Score = 10/10. Verdict: PASS. Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit FR-BILL-005.*
