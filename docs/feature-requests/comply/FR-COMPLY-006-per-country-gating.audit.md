---
fr_id: FR-COMPLY-006
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập từ nội dung hiện tại. FR-COMPLY-006 đặc tả khung per-country gating: ma trận luật khai báo (country_rule versioned) thay if-country rải rác, cổng Allow/Value deny-by-default, ba GateKey (voucher_stacking, affiliate_channel, data_protection_regime). 12 mệnh đề §1 (11 MUST + 1 SHOULD metric), testable. Seed khớp source §5.7 (MY/PH bỏ stacking 2025). Tách lớp với INFRA-005 (vùng) + nối COMPLY-007 (regime). country_rule khớp DATA-MODEL.md (country CHECK len=2, gate_key, value JSONB, UNIQUE(country, gate_key, version)). Đạt 10/10, PASS.

## §2 - Findings

- ISS-001 (frontmatter A): id/module/folder khớp; phase P3 hợp lệ; key đủ; depends_on=[FR-INFRA-005, FR-CART-004], blocks=[FR-COMPLY-007]. Pass.
- ISS-002 (contract D): country_rule DDL khớp DATA-MODEL owner; CHECK char_length(country)=2; UNIQUE(country, gate_key, version). Seed VN(allow)/MY,PH(deny)/regime PDPL|PDP_ID|PDPA_TH. Pass.
- ISS-003 (normative B): clause #1 dữ liệu hóa, #5 deny-by-default, #6 không if-country, #11 regime trỏ adapter, #12 mở rộng không phá - tiêu chí rõ. Pass.
- ISS-004 (AC/test E,F): 12 AC; test TestGate_MYNoStacking/VNStacking/UnknownCountryRejected/DenyByDefault, TestRegistry_DataRegimePerCountry/NewVersionDoesNotBreakOthers. Pass.
- ISS-005 (typography O): de-accent comment Go/SQL §3 (code block); prose ASCII thuần; không banned word. Pass.
- ISS-006 (§6-§11, M, N): đủ khung; failure-modes 8 dòng; sentinel có; self-contained, không TBD. Pass.

## §3 - Traceability §1 -> AC -> artefact

| §1 clause | AC | Artefact |
|---|---|---|
| #1 country_rule versioned | #1,#10 | 0007_country_rule.sql + CHECK |
| #2 tập GateKey | #2,#6,#9 | GateVoucherStacking/AffiliateChannel/DataRegime |
| #3 seed 7 nước | #1,#3,#4 | INSERT 7 nước |
| #4 Allow/Value | #2,#6 | registry.go |
| #5 deny-by-default | #9 | Value Denied + TestGate_DenyByDefault |
| #6 không if-country | - | CART/AFFIL gọi Value (§8) |
| #7 tái dùng INFRA-005 | - | §6 + DEC-COMPLY-26 |
| #8 versioned + reload | #11 | repo reload |
| #9 validate country/gate | #5,#10 | ErrUnknownCountry + CHECK + TestGate_UnknownCountryRejected |
| #10 metric (SHOULD) | #12 | gating_denied_total |
| #11 regime trỏ adapter | #6,#7,#8 | data_protection_regime + TestRegistry_DataRegimePerCountry |
| #12 mở rộng không phá | #11 | TestRegistry_NewVersionDoesNotBreakOthers |

## §4 - Kết luận

Ma trận khai báo + deny-by-default có test; khác biệt MY/PH no-stacking khớp source; schema khớp DATA-MODEL. Không defect. Score = 10/10. Verdict: PASS.

---

*Hết audit FR-COMPLY-006.*
