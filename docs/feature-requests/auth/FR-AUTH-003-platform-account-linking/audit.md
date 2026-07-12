---
fr_id: FR-AUTH-003
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập từ nội dung hiện tại. FR-AUTH-003 mang cam kết niềm tin lõi hậu-Honey: liên kết tài khoản sàn qua `platform_account` chỉ lưu `ext_user_ref` ẩn danh, UNIQUE(user_id, platform_id), TUYỆT ĐỐI không cột token/cookie/session/password. 12 mệnh đề §1 (10 MUST + 1 MUST cấm cấu trúc + 1 SHOULD metric), testable. Schema khớp DATA-MODEL.md exactly (ext_user_ref TEXT CHECK length 1..128, UNIQUE, no token column). Bất biến D đạt: ext_user_ref anonymized, no session token column. Đạt 10/10, PASS.

## §2 - Findings

- ISS-001 (frontmatter A): id/module/folder khớp; key đủ; depends_on=[FR-AUTH-001], blocks=[]. Pass.
- ISS-002 (security D - bất biến quan trọng nhất): kiểm cấm-token-về-cấu-trúc - DDL cố ý không có cột token/cookie/session/password (§1 #2, DEC-AUTH-12/14); test introspection TestSchema_NoCredentialColumns biến cam kết thành bất biến máy kiểm. ext_user_ref CHECK length 1..128 + looksLikeRawCredential reject email/cookie/token thô. Khớp source §3.4/§3.2/§5.5. Pass.
- ISS-003 (normative B): clause #2 cấm cấu trúc, #3 ẩn danh không username thật, #4 UNIQUE + upsert, #11 cô lập user - tiêu chí rõ. Pass.
- ISS-004 (AC/test E,F): 12 AC; test TestSchema_NoCredentialColumns, TestLink_NewAndUpsert/MultiPlatform/RejectsRawCredential, TestList_IsolatedPerUser, TestUnlink_Idempotent. Pass.
- ISS-005 (typography O): mũi tên/dấu unicode chỉ trong comment Go/SQL §3 (code block); prose ASCII thuần; không banned word. Pass.
- ISS-006 (§6-§11, M, N): đủ khung; failure-modes 8 dòng; sentinel có; self-contained, không TBD. Pass.

## §3 - Traceability §1 -> AC -> artefact

| §1 clause | AC | Artefact |
|---|---|---|
| #1 bảng + UNIQUE | #1 | 0006_platform_account.up.sql |
| #2 cấm token cấu trúc | #2 | DDL no-token + TestSchema_NoCredentialColumns |
| #3 ext_user_ref ẩn danh | #8 | looksLikeRawCredential + TestLink_RejectsRawCredential |
| #4 UNIQUE + upsert | #4,#5 | Upsert ON CONFLICT + TestLink_NewAndUpsert |
| #5 Link/List/Unlink | #3,#10,#11 | linkacct.go |
| #6 validate ext_user_ref | #6,#7 | CHECK + ErrInvalidExtRef |
| #7 từ extension tối thiểu | - | §6 + DEC-AUTH-15 |
| #8 unlink idempotent | #11 | TestUnlink_Idempotent |
| #9 no-log tái định danh | #12 | review + test output |
| #10 bằng chứng TRUST-003/COMPLY-005 | #2 | §7 |
| #11 cô lập user | #9 | ListByUser + TestList_IsolatedPerUser |
| #12 metric (SHOULD) | - | §1 #12 |

## §4 - Kết luận

Bất biến cam kết niềm tin (no session token column, ext_user_ref anonymized) kiểm bằng test introspection; schema khớp DATA-MODEL exactly; là bằng chứng cho TRUST-003 + COMPLY-005. Không defect. Score = 10/10. Verdict: PASS.

---

*Hết audit FR-AUTH-003.*
