---
fr_id: FR-AUTH-005
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập từ nội dung hiện tại. FR-AUTH-005 đặc tả vòng đời tài khoản: reset mật khẩu (token một-lần TTL ngắn lưu hash, phản hồi đồng nhất no-enumeration, thu hồi mọi refresh khi đổi mật khẩu), status active/suspended/deleted với thu hồi token, xóa thỏa DSAR PDPL (ẩn danh hóa PII tombstone + gỡ platform_account + thu hồi token + cửa sổ ân hạn). 12+ mệnh đề §1, testable. password_reset khớp DATA-MODEL.md (token_hash UNIQUE, expires_at, used_at). Phối FR-COMPLY-003. Đạt 10/10, PASS.

## §2 - Findings

- ISS-001 (frontmatter A): id/module/folder khớp; key đủ; depends_on=[FR-AUTH-001, FR-COMPLY-003]. Pass.
- ISS-002 (cấu trúc §1): §1 đánh số có một clause `8` và một clause `8b` (hai clause khác biệt cùng nhóm số 8: ân hạn + phối COMPLY-003). Không phải trùng nội dung, cả hai testable và truy được; AC #8 phủ grace window. Ghi nhận là bất thường đánh số nhẹ, KHÔNG phải defect cần sửa surgical (mọi clause vẫn có AC). Pass.
- ISS-003 (security/PDPL D,P): reset no-enumeration (RequestReset luôn trả nil, §1 #3); đổi mật khẩu/suspend/delete đều RevokeAllRefresh (§1 #4/#5, cùng access TTL ngắn AUTH-002); DeleteAccount ẩn danh hóa PII thật (không chỉ đặt cờ, §1 #7/#9) - đúng quyền xóa PDPL Luật 91/2025. Pass.
- ISS-004 (AC/test E,F): 12 AC; test TestRequestReset_NoEnumeration, TestConfirmReset_OneTime/RevokesSessions, TestSuspended_CannotLogin_TokensRevoked, TestDelete_ErasesPII_AndLinks/Idempotent. Pass.
- ISS-005 (typography O): mũi tên/dấu unicode chỉ trong comment Go §3 (code block); prose ASCII thuần; không banned word. Pass.
- ISS-006 (§6-§11, M, N): đủ khung; failure-modes 9 dòng; sentinel có; self-contained, không TBD. Pass.

## §3 - Traceability §1 -> AC -> artefact

| §1 clause | AC | Artefact |
|---|---|---|
| #1 status active/suspended/deleted | #6,#7,#8 | lifecycle.go + login gate |
| #2 reset token một-lần | #2,#3,#4 | reset.go ConfirmReset |
| #3 no-enumeration | #1 | RequestReset + TestRequestReset_NoEnumeration |
| #4 revoke khi đổi mật khẩu | #5 | RevokeAllRefresh + TestConfirmReset_RevokesSessions |
| #5 revoke khi đổi status | #6,#7 | SetStatus + TestSuspended_CannotLogin_TokensRevoked |
| #6 login từ chối non-active | #6,#7 | ErrAccountNotActive |
| #7 DeleteAccount DSAR | #9,#10,#11 | erasure.go + TestDelete_ErasesPII_AndLinks |
| #8 grace window | - | SetStatus deleted + purge job |
| #8b phối COMPLY-003 | - | §7 |
| #9 PII không truy định danh | #9 | AnonymizePII tombstone |
| #10 no-log token/PII | - | §6 audit log |
| #12 idempotent | #12 | TestDelete_Idempotent |

## §4 - Kết luận

Bất biến bảo mật + PDPL (no-enumeration, revoke-on-status, erase PII thật) kiểm bằng test; schema khớp DATA-MODEL; đánh số 8/8b bất thường nhẹ nhưng mọi clause có AC. Không defect cần sửa. Score = 10/10. Verdict: PASS.

---

*Hết audit FR-AUTH-005.*
