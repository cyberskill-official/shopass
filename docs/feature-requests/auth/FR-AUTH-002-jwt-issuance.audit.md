---
fr_id: FR-AUTH-002
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập từ nội dung hiện tại. FR-AUTH-002 đặc tả phát hành JWT: access RS256 ngắn hạn (claims user_id/locale/tier + exp/iat/iss/aud) + refresh dài hạn lưu hash, JWKS đa-kid (xoay không downtime), refresh rotation một-lần + theft detection thu hồi family. 12 mệnh đề §1 (10 MUST + 2 SHOULD theft/tier), testable. Bảng refresh_token khớp DATA-MODEL.md (token_hash, family_id UUID, expires_at/revoked_at/used_at). Khớp ranh giới verify của FR-INFRA-001. Đạt 10/10, PASS.

## §2 - Findings

- ISS-001 (frontmatter A): id/module/folder khớp; key đủ; depends_on=[FR-AUTH-001, FR-INFRA-001]. Pass.
- ISS-002 (security D): RS256 (không HS256 chia sẻ, disallowed_tools), khóa ký qua FR-INFRA-003; refresh lưu hash KHÔNG cleartext (§1 #6 + idx_rt_hash unique); rotation một-lần qua used_at; reuse -> RevokeFamily. Khớp source §3.1/§3.8. Pass.
- ISS-003 (normative B): clause #3 claims + aud khớp gateway, #5 JWKS đa-kid, #7 rotation, #10 theft detection - tiêu chí rõ. Pass.
- ISS-004 (AC/test E,F): 12 AC; test TestAccess_VerifiableViaJWKS/Expired/UnknownKID, TestJWKS_MultipleKID_AfterRotation, TestRefresh_OneTimeUse/StoredAsHash/ReuseRevokesFamily. Pass.
- ISS-005 (typography O): mũi tên unicode chỉ trong comment Go §5 (code block); prose ASCII thuần; không banned word. Pass.
- ISS-006 (§6-§11, M, N): đủ khung; failure-modes 9 dòng; sentinel có; self-contained, không TBD. Pass.

## §3 - Traceability §1 -> AC -> artefact

| §1 clause | AC | Artefact |
|---|---|---|
| #1 TokenPair | #1 | login.go IssueTokenPair |
| #2 access RS256 + khóa qua INFRA-003 | #2 | token.go issueAccess + TestAccess_VerifiableViaJWKS |
| #3 claims | #3 | Claims struct |
| #4 TTL ngắn | #4 | accessTTL + TestAccess_Expired_Rejected |
| #5 JWKS đa-kid | #6 | jwks.go + TestJWKS_MultipleKID_AfterRotation |
| #6 refresh hash | #7 | 0005_refresh_token.up.sql + TestRefresh_StoredAsHash |
| #7 rotation một-lần | #8 | MarkUsed + TestRefresh_OneTimeUse |
| #8 thu hồi | #10,#11 | revoked_at |
| #9 IssueTokenPair/Refresh | #1,#8 | login.go/refresh.go |
| #10 theft detection (SHOULD) | #9,#12 | RevokeFamily + TestRefresh_ReuseRevokesFamily |
| #11 tier claims (SHOULD) | #3 | Claims.Tier |
| #12 ký bằng kid hiện hành | #5 | tok.Header[kid] |

## §4 - Kết luận

Bất biến bảo mật (RS256, refresh hash, rotation, theft detection) kiểm bằng test; schema khớp DATA-MODEL; aud khớp gateway INFRA-001. Không defect. Score = 10/10. Verdict: PASS.

---

*Hết audit FR-AUTH-002.*
