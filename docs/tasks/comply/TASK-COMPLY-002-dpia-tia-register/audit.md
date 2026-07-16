---
fr_id: TASK-COMPLY-002
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập từ nội dung hiện tại. TASK-COMPLY-002 đặc tả sổ DPIA/TIA: hạn nộp 60 ngày (filing_due_at = started_at + 60 days) + chu kỳ rà soát 6 tháng (review_due_at), trạng thái suy từ deadline không nhập tay, TIA bắt buộc khi cross_border. 12 mệnh đề §1 (11 MUST + 1 SHOULD gauge), testable; logic `Status` tách thuần hàm. Schema khớp DATA-MODEL.md (processing_activity CHECK cross_border->recipient_country, dpia UNIQUE(activity_id,version), tia). task mở rộng dpia với mitigation_vi/filed_at/last_reviewed_at - nhất quán ghi chú "task phát triển sâu hơn §3.4". PDPL P đạt. Đạt 10/10, PASS.

## §2 - Findings

- ISS-001 (PDPL accuracy P): kiểm so source §5.5 - DPIA nộp 60 ngày + cập nhật 6 tháng, TIA cho chuyển dữ liệu xuyên biên giới, chế tài 5% doanh thu. Hằng FilingWindow=60d, ReviewCycle ~6 tháng (file tự ghi chú dùng AddDate(0,6,0) nếu cần chính xác lịch). Chính xác. Pass.
- ISS-002 (frontmatter A): id/module/folder khớp; key đủ; depends_on=[TASK-COMPLY-001]. Pass.
- ISS-003 (contract D): DDL khớp DATA-MODEL owner; CHECK (NOT cross_border OR recipient_country IS NOT NULL); dpia CHECK risk_level low|medium|high; tia FK dpia. Pass.
- ISS-004 (AC/test E,F): 12 AC; test TestDeadline_FilingOverdueAfter60d/DraftWithin60d/ReviewOverdueAfter6m, TestRegister_CrossBorderRequiresTIA/WithTIA_CreatesTIA, TestReview_CreatesNewVersion. Pass.
- ISS-005 (typography O): de-accent comment trong Go/SQL §3 (code block); prose ASCII thuần; không banned word. Pass.
- ISS-006 (§6-§11, M, N): đủ khung; failure-modes 8 dòng; sentinel có; self-contained, không TBD. Pass.

## §3 - Traceability §1 -> AC -> artefact

| §1 clause | AC | Artefact |
|---|---|---|
| #1 processing_activity | #1 | 0003_processing_activity.sql |
| #2 dpia versioned | #2,#9 | 0004_dpia_register.sql + TestReview_CreatesNewVersion |
| #3 tia bắt buộc cross-border | #8 | tia + TestRegister_CrossBorderWithTIA_CreatesTIA |
| #4 hạn 60 ngày | #3,#4 | FilingWindow + TestDeadline_FilingOverdueAfter60d |
| #5 chu kỳ 6 tháng | #6 | ReviewCycle + TestDeadline_ReviewOverdueAfter6m |
| #6 status suy từ deadline | #4,#5,#6 | Status() thuần hàm |
| #7 chặn cross-border thiếu TIA | #7 | ErrTIARequired + TestRegister_CrossBorderRequiresTIA |
| #8 hàm register/review/file/overdue | #2,#9,#12 | register.go + repo.go |
| #9 due-soon | - | §1 #9 |
| #10 gauge (SHOULD) | - | §1 #10 |
| #11 risk_level CHECK | #10,#11 | CHECK + AC |
| #12 báo cáo cơ quan | #12 | Report() + query §8 |

## §4 - Kết luận

Hai mốc PDPL (60 ngày, 6 tháng) ánh xạ vào hằng + test; TIA cross-border bắt buộc kiểm bằng test; schema khớp DATA-MODEL. Không defect. Score = 10/10. Verdict: PASS.

---

*Hết audit TASK-COMPLY-002.*
