---
fr_id: TASK-COMPLY-005
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập từ nội dung hiện tại. TASK-COMPLY-005 cưỡng chế hai bất biến: (a) no-cleartext (mật khẩu chỉ argon2id, secrets chỉ Vault); (b) token-not-on-server (token phiên sàn không có ở backend). Hai lớp: quét tĩnh CI gate (4 rule: cleartext_password, platform_session_token, hardcoded_secret, weak_password_hash) + payload guard động tại biên (GuardPayload chặn cookie/token/session/authorization). 12 mệnh đề §1 (10 MUST + 1 SHOULD báo cáo + 1 MUST fixture hai chiều), testable. Allowlist có lý do xử lý false-positive. Nền cho TRUST-003. Đạt 10/10, PASS.

## §2 - Findings

- ISS-001 (frontmatter A): id/module/folder khớp; key đủ; depends_on=[TASK-INFRA-003], blocks=[TASK-TRUST-003]. Pass.
- ISS-002 (security/contract D): kiểm bất biến token-not-on-server - rule regex nhắm token SÀN cụ thể (shopee|tiktok|lazada)_(token|cookie|session) (§1 #4), KHÁC token JWT nội bộ (xử lý qua allowlist // audit:allow); GuardPayload chặn forbidden keys. Tái dùng argon2id (AUTH-001) + Vault (INFRA-003). Khớp source §3.8/§5.5. Pass.
- ISS-003 (normative B): clause #3 CI gate fail build, #5 payload guard động, #9 allowlist có lý do, #12 fixture hai chiều (true-pos + clean-pass) - tiêu chí rõ. Pass.
- ISS-004 (AC/test E,F): 12 AC; test TestScan_CatchesPlatformToken/CleartextPassword/CleanCodePasses/AllowlistSkipped, TestGuard_RejectsCookie/RejectsAuthorizationAnyCase/MinimalPayloadPasses. Pass.
- ISS-005 (typography O): de-accent comment + shell §3 (code block); prose ASCII thuần; không banned word. Pass.
- ISS-006 (§6-§11, M, N): đủ khung; failure-modes 8 dòng; sentinel có; self-contained, không TBD. Pass.

## §3 - Traceability §1 -> AC -> artefact

| §1 clause | AC | Artefact |
|---|---|---|
| #1 Scan tĩnh | #1-#5 | scan.go + rules.go |
| #2 tập quy tắc cấm | #1,#2,#3,#4 | bannedRules (4 rule) |
| #3 CI gate | #7 | no_cleartext_gate.sh + ci.yml |
| #4 token-not-on-server | #2 | rule platform_session_token + TestScan_CatchesPlatformToken |
| #5 payload guard động | #8,#9,#10 | payload_guard.go + 3 guard tests |
| #6 argon2id reuse | #1,#4 | rule cleartext/weak |
| #7 Vault reuse | #3 | rule hardcoded_secret |
| #8 Finding xác định | #11 | sort ổn định |
| #9 allowlist có lý do | #6 | // audit:allow + TestScan_AllowlistSkipped |
| #10 báo cáo (SHOULD) | #12 | §1 #10 |
| #11 allowlist đếm/liệt kê | #12 | §1 #11 |
| #12 fixture hai chiều | #1-#6 | scan_test true-pos + clean-pass |

## §4 - Kết luận

Hai lớp cưỡng chế (tĩnh + động) ánh xạ vào artefact + test; bất biến token-not-on-server thành ràng buộc máy kiểm; allowlist kiểm soát false-positive. Không defect. Score = 10/10. Verdict: PASS.

---

*Hết audit TASK-COMPLY-005.*
