---
fr_id: FR-INFRA-005
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập từ nội dung hiện tại. FR-INFRA-005 đặc tả lớp cấu hình per-country: CountryPolicy dữ liệu hóa (không if-country rải rác), loader fail-fast 7 nước, feature flag scope-country, mặc định hạn chế nhất (deny). 12 mệnh đề §1 (11 MUST + 1 SHOULD override), testable; có test VN-stack vs MY/PH-no-stack (khớp source §3.9/§5.7) và nước-lạ-hạn-chế. depends_on=[FR-INFRA-002] (đọc platform.country) hợp lý. Đạt 10/10, PASS.

## §2 - Findings

- ISS-001 (frontmatter A): id/module/folder khớp; key bắt buộc đủ; depends_on/blocks nhất quán §7. Pass.
- ISS-002 (contract D): code Go thật - CountryPolicy struct, restrictivePolicy mặc định an toàn, loader validate alpha-2 + dup + channel enum, Lookup bắt buộc country. countries.yaml seed VN(stack)/MY,PH(no-stack)/ID khớp luật nguồn. Không phải bảng DB (config YAML). Pass.
- ISS-003 (normative B): clause #4 stacking đúng luật thực, #5 deny-by-default, #7 truy vấn bắt buộc country, #8 country từ platform/locale không từ IP - tiêu chí rõ. Pass.
- ISS-004 (AC/test E,F): 12 AC; test TestPolicy_VNStacks_MYPHDoNot, TestPolicy_UnknownCountry_Restrictive, TestPolicy_AffiliateChannel_PerCountry, TestLoad_DuplicateCountry/BadChannel/NonAlpha2_Errors, TestLookup_PerCountry/MissingCountry_False. Pass.
- ISS-005 (typography O): mũi tên unicode chỉ trong comment Go/YAML §3 (code block); prose ASCII thuần; không banned word. Pass.
- ISS-006 (§6-§11, M, N): đủ khung; failure-modes 9 dòng; sentinel có; self-contained, không TBD. Pass.

## §3 - Traceability §1 -> AC -> artefact

| §1 clause | AC | Artefact |
|---|---|---|
| #1 CountryPolicy dữ liệu | #2,#3 | policy.go |
| #2 policy mang 3 trục | #2 | CountryPolicy fields |
| #3 loader 7 nước | #1 | loader.go + countries.yaml |
| #4 stacking đúng luật | #3 | yaml VN/MY/PH + TestPolicy_VNStacks_MYPHDoNot |
| #5 deny-by-default | #4 | restrictivePolicy + TestPolicy_UnknownCountry_Restrictive |
| #6 flag scope-country | #6 | flags.go + TestLookup_PerCountry |
| #7 truy vấn bắt buộc country | #7 | Lookup + TestLookup_MissingCountry_False |
| #8 country từ platform/locale | - | §6 + DEC-INFRA-25 |
| #9 affiliate channel enum | #5 | Channel + TestPolicy_AffiliateChannel_PerCountry |
| #10 loader validate | #8,#9,#10,#11 | Load + 3 loader tests |
| #11 override runtime (SHOULD) | - | §1 #11 |
| #12 nguồn sự thật per-country | #12 | §7 CART/AFFIL/COMPLY |

## §4 - Kết luận

Mọi mệnh đề có code/test backing; khác biệt stacking VN vs MY/PH khớp source §3.9/§5.7; loader validate kiểm bằng test. Không defect. Score = 10/10. Verdict: PASS.

---

*Hết audit FR-INFRA-005.*
