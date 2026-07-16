---
fr_id: TASK-AUTH-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập từ nội dung hiện tại. TASK-AUTH-001 đặc tả schema bảo mật `app_user` + đăng ký: argon2id định dạng PHC (chuỗi bắt đầu `$argon2id$`, lưu time/memory/parallelism + salt + hash), salt ngẫu nhiên mỗi lần, ConstantTimeCompare, email CITEXT chuẩn hóa, no-cleartext. 12 mệnh đề §1 (10 MUST + 2 SHOULD tham số/log), testable. Migration 0004 thêm pwd_hash/referral_code_id khớp DATA-MODEL.md (mở rộng app_user). Bất biến bảo mật D đạt: argon2id PHC, NEVER cleartext. Đạt 10/10, PASS.

## §2 - Findings

- ISS-001 (frontmatter A): id/module/folder khớp; key bắt buộc đủ; depends_on=[TASK-INFRA-002], blocks 4 task. Pass.
- ISS-002 (security D): kiểm bất biến - Hash dùng rand salt (§1 #3), PHC tự mã hóa tham số nên Verify hash cũ vẫn chạy sau nâng mặc định (§1 #4); cấm md5/sha1/cleartext (disallowed_tools); subtle.ConstantTimeCompare chống timing. Khớp source §3.4/§3.8 (argon2id, KHÔNG cleartext). Pass.
- ISS-003 (normative B): clause #2 PHC, #5 email HOẶC phone, #6 chuẩn hóa + CITEXT, #7 độ mạnh >=8, #10 ErrEmailTaken->409 - tiêu chí rõ. Pass.
- ISS-004 (AC/test E,F): 12 AC; test TestHash_DifferentEachTime (+ prefix $argon2id$), TestVerify_CorrectAndWrong/OldParamsStillWork, TestRegister_NoIdentifier/WeakPassword/DuplicateEmail_CaseInsensitive/TrimsEmail. Pass.
- ISS-005 (typography O): mũi tên unicode chỉ trong comment Go §3 (code block); prose ASCII thuần; không banned word. Pass.
- ISS-006 (§6-§11, M, N): đủ khung; failure-modes 9 dòng; sentinel có; self-contained, không TBD. Pass.

## §3 - Traceability §1 -> AC -> artefact

| §1 clause | AC | Artefact |
|---|---|---|
| #1 migration cột bảo mật | #1 | 0004_app_user_secrets.up.sql |
| #2 argon2id PHC | #5 | password.go Hash + TestHash (prefix $argon2id$) |
| #3 salt ngẫu nhiên | #2 | rand.Read + TestHash_DifferentEachTime |
| #4 verify hash cũ | #4 | decodePHC + TestVerify_OldParamsStillWork |
| #5 email HOẶC phone | #6 | Register + TestRegister_NoIdentifier |
| #6 chuẩn hóa email CITEXT | #8,#9,#10 | normalizeEmail + TestRegister_DuplicateEmail/TrimsEmail |
| #7 độ mạnh mật khẩu | #7 | checkPasswordStrength + TestRegister_WeakPassword |
| #8 repo Insert/FindByEmail | #10 | repo.go |
| #9 no-log credential | #11 | review + test output |
| #10 ErrEmailTaken | #8 | isUniqueViolation |
| #11 status active | #12 | DATA-MODEL default |
| #12 tham số mặc định (SHOULD) | - | §1 #12 |

## §4 - Kết luận

Bất biến bảo mật (argon2id PHC, no-cleartext, salt, constant-time) kiểm bằng test; schema khớp DATA-MODEL; là điểm chứng minh no-cleartext cho COMPLY-005. Không defect. Score = 10/10. Verdict: PASS.

---

*Hết audit TASK-AUTH-001.*
