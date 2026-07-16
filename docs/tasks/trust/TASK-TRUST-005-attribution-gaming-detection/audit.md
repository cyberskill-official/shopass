---
fr_id: TASK-TRUST-005
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập từ nội dung hiện tại. TASK-TRUST-005 đặc tả delay payout + phát hiện gaming attribution. Trục: cửa sổ delay (hold-then-release, eligible_at = confirmed_at + delay_window) + ba điều kiện giải ngân (hết delay + không hold_reason + network-confirm). Ba dấu hiệu gaming: last-click thao túng, self-referral qua SameCluster (TASK-TRUST-004), cookie-stuffing. Nguyên tắc: chỉ delay/giữ + điều tra, KHÔNG tự từ chối; denied chỉ sau confirmed_fraud. Job release idempotent (lock + status). 12 mệnh đề §1 (10 MUST + 2 MUST NOT/MUST tiền). Schema khớp DATA-MODEL.md. Tiền payout_amount BIGINT VND (D đạt). Đạt 10/10, PASS.

## §2 - Findings

- ISS-001 (frontmatter A): id/module/folder khớp; phase P3; key đủ; depends_on=[TASK-AFFIL-001, TASK-AFFIL-003, TASK-BILL-002, TASK-TRUST-004], blocks=[TASK-AFFIL-005]. Pass.
- ISS-002 (contract D + money): payout_hold DDL khớp DATA-MODEL (conversion_id PK -> affiliate_conversion, beneficiary -> app_user, payout_amount BIGINT CHECK >=0 VND, confirmed_at/eligible_at, hold_reason, status CHECK holding|under_investigation|released|denied, released_at + CHECK released->released_at NOT NULL). Money BIGINT VND không float (§1 #9, DEC-PRICE-05). dueSQL FOR UPDATE SKIP LOCKED idempotent. Khớp source §5.3/§4.2. Pass.
- ISS-003 (normative B): clause #1 MUST NOT trả ngay, #3 risk cao kéo dài không từ chối, #4 ba điều kiện, #6 denied chỉ sau điều tra - tiêu chí rõ. Pass.
- ISS-004 (AC/test E,F): 12 AC; test TestDelay_NotReleasedBeforeWindow/ReleasedAfterWindow_AllConditions, TestGuard_LastClickManipulation/SelfReferral_SameCluster/CleanConversion_NoHold, TestRelease_HighRisk_HoldsNotDenies/Idempotent_NoDoublePay. Pass.
- ISS-005 (typography O): SQL/Go comment §3 (code block); prose ASCII thuần; không banned word. Pass.
- ISS-006 (§6-§11, M, N): đủ khung; failure-modes 8 dòng; sentinel có; self-contained, không TBD. Pass.

## §3 - Traceability §1 -> AC -> artefact

| §1 clause | AC | Artefact |
|---|---|---|
| #1 delay không trả ngay | #1,#2 | payout_hold.eligible_at + TestDelay_NotReleasedBeforeWindow |
| #2a last-click thao túng | #4 | attribution_guard.Inspect + TestGuard_LastClickManipulation |
| #2b self-referral | #5 | SameCluster + TestGuard_SelfReferral_SameCluster |
| #2c cookie-stuffing | #6 | guard UserInitiated check |
| #3 risk cao kéo dài | #7 | delay.go + TestRelease_HighRisk_HoldsNotDenies |
| #4 ba điều kiện release | #3,#10 | release.go dueSQL |
| #5 ghi hold_reason | #12 | hold_reason cụ thể |
| #6 vòng đời status | #6 | payout_hold.status CHECK |
| #7 self-referral qua đồ thị | #5 | SameCluster (TRUST-004) |
| #9 tiền BIGINT VND | #12 | payout_amount BIGINT |
| #10 cleared -> released | #8 | release sau cleared |
| #11 idempotent no-double-pay | #11 | lock + status + TestRelease_Idempotent |

## §4 - Kết luận

Delay + ba điều kiện là cổng đúng-đắn cho tiền ra; ba dấu hiệu gaming có test; không-tự-từ-chối nhất quán TRUST-004; money BIGINT VND; schema khớp DATA-MODEL. Không defect. Score = 10/10. Verdict: PASS.

---

*Hết audit TASK-TRUST-005.*
