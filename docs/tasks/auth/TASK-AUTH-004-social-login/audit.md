---
fr_id: TASK-AUTH-004
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập từ nội dung hiện tại. TASK-AUTH-004 đặc tả social login (Google OIDC + adapter Facebook/Zalo): Authorization Code + PKCE, state chống CSRF, nonce chống replay, verify id_token (JWKS provider + aud/iss/nonce), merge CHỈ khi email_verified, đường token thống nhất (cùng TokenPair TASK-AUTH-002). 12 mệnh đề §1 (priority SHOULD nhưng bảo mật MUST chặt: clause #2-#6 dùng MUST), testable. social_identity khớp DATA-MODEL.md (provider CHECK google|facebook|zalo, UNIQUE(provider, subject)). Đạt 10/10, PASS.

## §2 - Findings

- ISS-001 (frontmatter A): id/module/folder khớp; priority SHOULD hợp lệ; key đủ; depends_on=[TASK-AUTH-002]. Pass.
- ISS-002 (security D): kiểm phòng account takeover - resolveUser chỉ LinkSocial khi c.EmailVerified (§1 #6, DEC-AUTH-19); state take-once + nonce verify; client secret qua TASK-INFRA-003 (disallowed_tools). social_identity UNIQUE(provider, subject) khớp DATA-MODEL. Khớp source §3.1. Pass.
- ISS-003 (normative B): SHOULD ở priority nhưng clause bảo mật #2/#3/#4/#6 là MUST (đúng: nếu ship thì ship chuẩn). Có >=1 MUST. Tiêu chí rõ. Pass.
- ISS-004 (AC/test E,F): 12 AC; test TestCallback_BadState_Rejected/NonceMismatch_Rejected, TestResolve_VerifiedEmail_LinksExisting/UnverifiedEmail_NoMerge/ReturningUser_NoDuplicate/MultiProvider_OneUser. Pass.
- ISS-005 (typography O): mũi tên unicode chỉ trong comment Go §5 (code block); prose ASCII thuần; không banned word. Pass.
- ISS-006 (§6-§11, M, N): đủ khung; failure-modes 9 dòng; sentinel có; self-contained, không TBD. Pass.

## §3 - Traceability §1 -> AC -> artefact

| §1 clause | AC | Artefact |
|---|---|---|
| #1 Google + adapter | #12 | oauth_google.go + stub |
| #2 PKCE + state + nonce | #2 | StartOAuth + authURL |
| #3 kiểm state | #3 | OAuthCallback + TestCallback_BadState_Rejected |
| #4 verify id_token | #4,#5 | ExchangeAndVerify + TestCallback_NonceMismatch_Rejected |
| #5 social_identity UNIQUE | #1 | 0007_social_identity.up.sql |
| #6 merge verified-only | #7,#8 | resolveUser + TestResolve_Unverified/VerifiedEmail |
| #7 tạo user mới | #6 | createUserWithSocial |
| #8 TokenPair thống nhất | #12 | IssueTokenPair |
| #9 secret qua INFRA-003 | #11 | §1 #9 + grep |
| #10 multi-provider | #10 | TestResolve_MultiProvider_OneUser |
| #11 state/verifier take-once | - | tmp.Take 5min TTL |
| #12 metric (SHOULD) | - | §1 #12 |

## §4 - Kết luận

Bất biến bảo mật (PKCE/state/nonce, verify id_token, merge verified-only) kiểm bằng test; schema khớp DATA-MODEL; đường token thống nhất. Không defect. Score = 10/10. Verdict: PASS.

---

*Hết audit TASK-AUTH-004.*
