---
fr_id: TASK-COMPLY-004
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập từ nội dung hiện tại. TASK-COMPLY-004 đặc tả quy trình thông báo vi phạm 72 giờ theo Luật 91/2025. Điểm pháp lý quan trọng nhất - đồng hồ đếm từ `acknowledged_at` (nhận biết/became aware) KHÔNG từ `occurred_at` - mã hóa đúng (authority_due_at = acknowledged_at + 72h). State machine tuần tự (detected->triaged->notified_authority->notified_subjects->closed) + ràng buộc Close chặn high/critical chưa notified_subjects. 12 mệnh đề §1 (11 MUST + 1 SHOULD metric), testable. breach_incident khớp DATA-MODEL.md. Nối observability TASK-INFRA-004. PDPL P đạt. Đạt 10/10, PASS.

## §2 - Findings

- ISS-001 (PDPL accuracy P): kiểm so source §5.5 - thông báo vi phạm trong 72 giờ. Mốc đếm từ acknowledged_at (đúng "became aware", không từ thời điểm sự cố vật lý) - đây là điểm dễ sai mà file làm đúng. Chế tài 3 tỷ. Chính xác. Pass.
- ISS-002 (frontmatter A): id/module/folder khớp; key đủ; depends_on=[TASK-COMPLY-001, TASK-INFRA-004]. Pass.
- ISS-003 (contract D): breach_incident DDL khớp DATA-MODEL owner (severity CHECK low|medium|high|critical, status CHECK 5 trạng thái, acknowledged_at NOT NULL, source_ref). State machine `order` map chặn nhảy/lùi. Pass.
- ISS-004 (AC/test E,F): 12 AC; test TestClock_OverdueAfter72h/WithinWindow, TestAdvance_SequentialOnly/NoBackward, TestClose_CriticalNeedsSubjectNotice/LowSeverityNoSubjectNotice. Pass.
- ISS-005 (typography O): de-accent comment Go/SQL §3 (code block); prose ASCII thuần; không banned word. Pass.
- ISS-006 (§6-§11, M, N): đủ khung; failure-modes 8 dòng; sentinel có; self-contained, không TBD. Pass.

## §3 - Traceability §1 -> AC -> artefact

| §1 clause | AC | Artefact |
|---|---|---|
| #1 breach_incident | #1 | 0006_breach_incident.sql |
| #2 đồng hồ từ acknowledged_at | #6,#7 | AuthorityWindow + TestClock_OverdueAfter72h |
| #3 state machine tuần tự | #3,#4,#5 | order map + TestAdvance_SequentialOnly |
| #4 dấu thời gian từng bước | #3 | transition set ts |
| #5 severity quyết thông báo chủ thể | #8,#10 | Close + TestClose_CriticalNeedsSubjectNotice |
| #6 cờ deadline tự suy | #6,#7 | DeadlineFlag |
| #7 source_ref observability | - | source_ref + §8 |
| #8 hàm Open/Advance/Close/Overdue | #2 | incident.go |
| #9 từ chối transition lạ | #4,#5 | ErrInvalidTransition + TestAdvance_NoBackward |
| #10 metric (SHOULD) | #12 | breach_overdue_total |
| #11 Close chặn chưa notified_subjects | #8 | ErrSubjectsNotNotified |
| #12 log audit transition | - | §1 #12 + §10 |

## §4 - Kết luận

Mốc 72 giờ tính đúng từ nhận biết; state machine + ràng buộc Close có test; schema khớp DATA-MODEL. Không defect. Score = 10/10. Verdict: PASS.

---

*Hết audit TASK-COMPLY-004.*
