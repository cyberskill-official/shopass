---
fr_id: FR-COMPLY-003
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập từ nội dung hiện tại. FR-COMPLY-003 đặc tả DSAR (access/rectify/erase/portability) theo Luật 91/2025: xác minh danh tính (user_id từ token), chống rò cross-user bằng property test, erase hỗn hợp (hard-delete dữ liệu thuần + soft-anonymize kế toán + GIỮ consent_record), portability JSON máy đọc được, SLA suy từ requested_at. 12 mệnh đề §1 (11 MUST + 1 SHOULD metric), testable. dsar_request khớp DATA-MODEL.md (kind CHECK access|rectify|erase|portability). Phụ thuộc trực tiếp của AUTH-005. PDPL P đạt. Đạt 10/10, PASS.

## §2 - Findings

- ISS-001 (PDPL accuracy P): kiểm so source §5.5 - bốn quyền chủ thể dữ liệu (truy cập/sửa/xóa/di chuyển), chế tài 3 tỷ. Erase giữ consent_record làm chứng cứ (đúng yêu cầu tái lập PDPL) + soft-anonymize dữ liệu kế toán. Chính xác. Pass.
- ISS-002 (frontmatter A): id/module/folder khớp; key đủ; depends_on=[FR-COMPLY-001], blocks=[FR-AUTH-005, FR-B2B-001]. Pass.
- ISS-003 (contract D): dsar_request DDL khớp DATA-MODEL owner (kind CHECK, sla_due_at, completed_at, note); idx_dsar_open partial index. Export/Erase gọi service khác qua interface inject. Pass.
- ISS-004 (AC/test E,F + cổng cross-user): 12 AC; property test TestExport_OnlyOwnData (rò cross-user = 0), TestExport_PortabilityIsJSON, TestErase_KeepsConsentLog/Idempotent, TestDSAR_Overdue. Pass.
- ISS-005 (typography O): de-accent comment Go §3 (code block); prose ASCII thuần; không banned word. Pass.
- ISS-006 (§6-§11, M, N): đủ khung; failure-modes 8 dòng; sentinel có; self-contained, không TBD. Pass.

## §3 - Traceability §1 -> AC -> artefact

| §1 clause | AC | Artefact |
|---|---|---|
| #1 dsar_request | #1,#8 | 0005_dsar_request.sql CHECK kind |
| #2 xác minh danh tính | #11 | user_id từ token + tầng auth |
| #3 access | #3 | Export summary |
| #4 portability JSON | #3 | ExportBundle + TestExport_PortabilityIsJSON |
| #5 erase hỗn hợp + giữ consent | #5,#6 | erase.go + TestErase_KeepsConsentLog |
| #6 rectify | #12 | luồng rectify + unique AUTH-001 |
| #7 SLA suy từ deadline | #9,#10 | sla_due_at + idx_dsar_open |
| #8 hàm Create/Export/Erase/Overdue | #2,#9 | request.go/export.go/erase.go |
| #9 chống rò cross-user | #4 | property test TestExport_OnlyOwnData |
| #10 metric (SHOULD) | - | §1 #10 |
| #11 ghi completed_at | #10 | erase dấu vết |
| #12 erase idempotent | #7 | TestErase_Idempotent |

## §4 - Kết luận

Bốn quyền DSAR có code/test; cổng rò cross-user = 0 có property test; erase cân bằng quyền xóa vs nghĩa vụ lưu trữ; schema khớp DATA-MODEL. Không defect. Score = 10/10. Verdict: PASS.

---

*Hết audit FR-COMPLY-003.*
