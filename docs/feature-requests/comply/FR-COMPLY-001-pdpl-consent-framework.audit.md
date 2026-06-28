---
fr_id: FR-COMPLY-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập từ nội dung hiện tại (không tin .audit cũ). FR-COMPLY-001 đặc tả khung consent PDPL Luật 91/2025 + NĐ 356/2025. Bốn nguyên tắc (tự nguyện, cụ thể, đơn mục đích, tái lập được) mã hóa trực tiếp: consent_policy versioned bất biến + consent_record append-only + cổng `IsAllowed` (false-default). Im lặng != đồng thuận chứng minh bằng test. 12 mệnh đề §1 (11 MUST + 1 SHOULD metric), testable. Schema khớp DATA-MODEL.md (consent_policy UNIQUE(purpose_key,version); consent_record composite FK + source CHECK web|extension|mobile). PDPL P đạt. Đạt 10/10, PASS.

## §2 - Findings

- ISS-001 (PDPL accuracy P): kiểm độc lập so source §5.5 - Luật 91/2025/QH15, NĐ 356/2025 thay NĐ 13/2023, hiệu lực 01/01/2026 (file ghi đúng + chú thích đề bài nhầm 01/07/2026 ở §11, đính chính đúng), consent tự nguyện/cụ thể/đơn mục đích/tái lập, im lặng != đồng thuận, chế tài 5%/3 tỷ. Tất cả chính xác. Pass.
- ISS-002 (frontmatter A): id/module/folder khớp; key đủ; depends_on=[FR-INFRA-002], blocks 6 FR. Pass.
- ISS-003 (contract D): consent_policy + consent_record DDL khớp DATA-MODEL owner; composite FK (purpose_key, policy_version)->consent_policy; CHECK source. append-only ở tầng app (repo chỉ `append`). Pass.
- ISS-004 (AC/test E,F): 12 AC; test TestConsent_SilenceIsNotConsent/GrantThenAllowed/WithdrawAppendsRow_KeepsHistory/UnknownPurpose_Rejected/OldConsentKeepsOldVersion. Mỗi clause testable có test. Pass.
- ISS-005 (typography O): prose ASCII thuần; nội dung seed có dấu de-accent trong SQL string §3 (code block, scoped out); không banned word trong prose. Pass.
- ISS-006 (§6-§11, M, N): đủ khung; failure-modes 8 dòng; sentinel có; self-contained, không TBD. Pass.

## §3 - Traceability §1 -> AC -> artefact

| §1 clause | AC | Artefact |
|---|---|---|
| #1 consent_policy versioned | #1,#11 | 0001_consent_policy.sql UNIQUE |
| #2 consent_record append-only | #5,#7 | 0002_consent_record.sql + repo.append |
| #3 đơn mục đích | #8 | enum Purpose + TestConsent_UnknownPurpose_Rejected |
| #4 im lặng != đồng thuận | #3,#4 | IsAllowed false-default + TestConsent_SilenceIsNotConsent |
| #5 trường chứng minh | #2 | ip/user_agent/source + CHECK |
| #6 Grant/Withdraw/IsAllowed/History | #2,#5,#6,#7 | service.go |
| #7 từ chối purpose/version lạ | #8,#9 | validPurpose + FK |
| #8 seed purpose lõi | #1 | INSERT 4 purpose |
| #9 idempotent double-submit | - | dòng mới nhất theo ts |
| #10 metric + audit log (SHOULD) | #12 | OTel counter |
| #11 IsAllowed nguồn sự thật | - | §7 coverage |
| #12 giữ version cũ | #11 | TestConsent_OldConsentKeepsOldVersion |

## §4 - Kết luận

Bốn nguyên tắc PDPL ánh xạ 1-1 vào artefact + test; mọi fact PDPL chính xác; schema khớp DATA-MODEL. Không defect. Score = 10/10. Verdict: PASS.

---

*Hết audit FR-COMPLY-001.*
