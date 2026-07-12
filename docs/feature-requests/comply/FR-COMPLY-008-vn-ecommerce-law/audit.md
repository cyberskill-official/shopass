---
fr_id: FR-COMPLY-008
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập từ nội dung hiện tại. FR-COMPLY-008 đặc tả theo dõi nghĩa vụ TMĐT VN: đăng ký MOIT (NĐ 52/2013), trách nhiệm sàn + ngưỡng 100.000 giao dịch/năm foreign platform (NĐ 85/2021), disclosure livestream + affiliate (dự thảo Luật TMĐT 2025). Cờ must_register suy tự động từ bộ đếm; ngưỡng + checklist cấu hình versioned một chỗ; giá trị dự thảo 2025 đánh dấu "chờ luật chốt". 12 mệnh đề §1 (priority SHOULD, cấp P3, SHOULD/MAY đúng). Schema khớp DATA-MODEL.md (ecommerce_obligation source_law CHECK ND_52_2013|ND_85_2021|DRAFT_2025, yearly_transaction_count, compliance_threshold). PDPL/TMĐT P đạt. Đạt 10/10, PASS.

## §2 - Findings

- ISS-001 (legal accuracy P): kiểm so source §5.5 - NĐ 52/2013 + 85/2021, đăng ký MOIT, ngưỡng >100.000 giao dịch/năm foreign platform, dự thảo Luật TMĐT 2025 lần đầu quản livestream + affiliate. Hằng seed 100000 + source_law DRAFT_2025 đánh dấu tạm thời (trung thực với §10 nguồn). Chính xác. Pass.
- ISS-002 (frontmatter A): id/module/folder khớp; priority SHOULD + phase P3 hợp lệ; depends_on=[FR-COMPLY-001]. Pass.
- ISS-003 (contract D): 3 bảng DDL khớp DATA-MODEL owner; ecommerce_obligation status CHECK (FR mở rộng tập trạng thái sâu hơn DATA-MODEL "status TEXT" - nhất quán); compliance_threshold UNIQUE(key,version). Pass.
- ISS-004 (normative B): SHOULD/MAY đúng cấp P3; clause #3 cờ suy tự động (không nhập tay), #11 ngưỡng một chỗ, #12 đánh dấu tạm thời - tiêu chí rõ. Pass.
- ISS-005 (AC/test E,F): 12 AC; test TestThreshold_BelowDoesNotFlag/AboveFlagsRegister/FromVersionedConfig, TestObligation_DraftMarkedProvisional/MarkApproved/InvalidStatusRejected. Pass.
- ISS-006 (typography O + §6-§11, M, N): `NĐ`/`TMĐT` trong SQL string literal §3 (code block, scoped out); prose ASCII thuần; không banned word; failure-modes 8 dòng; sentinel có; self-contained. Pass.

## §3 - Traceability §1 -> AC -> artefact

| §1 clause | AC | Artefact |
|---|---|---|
| #1 ecommerce_obligation | #1,#7 | 0008_ecommerce_obligation.sql |
| #2 bộ đếm + ngưỡng versioned | #1,#9 | yearly_transaction_count + compliance_threshold |
| #3 cờ must_register tự suy | #2,#3,#4 | threshold.go + TestThreshold_Below/AboveFlags |
| #4 trạng thái MOIT | #5,#10 | MarkObligation + TestObligation_MarkApproved |
| #5 disclosure affiliate/livestream | #6 | seed 2 nghĩa vụ DRAFT_2025 |
| #6 source_law | #7 | cột source_law |
| #7 checklist versioned | #9,#11 | UNIQUE(obligation_key,version) |
| #8 hàm Threshold/Obligations/Mark/Outstanding | #2,#5 | obligation.go |
| #9 validate status | #8 | CHECK + TestObligation_InvalidStatusRejected |
| #10 metric (MAY) | #12 | ecom_threshold_exceeded |
| #11 ngưỡng một chỗ | #9 | TestThreshold_FromVersionedConfig |
| #12 đánh dấu tạm thời | #6,#11 | TestObligation_DraftMarkedProvisional |

## §4 - Kết luận

Cờ ngưỡng suy tự động + cấu hình versioned có test; giá trị dự thảo đánh dấu tạm thời trung thực; schema khớp DATA-MODEL. Không defect. Score = 10/10. Verdict: PASS.

---

*Hết audit FR-COMPLY-008.*
