---
fr_id: TASK-TRUST-002
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập từ nội dung hiện tại. TASK-TRUST-002 đặc tả chính sách tối thiểu hóa dữ liệu kiểm chứng được: liệt kê đúng tập {platform, productId, price, qty} (+ voucher), mục đích + cơ sở pháp lý PDPL từng trường, cam kết KHÔNG cookie/mật khẩu/token, local-first. Trục: chính sách neo vào allowlist TASK-EXT-003 qua policy-allowlist-parity.test.ts (hai chiều: mọi trường allowlist phải khai; không khai trường pipeline không gửi). collected-fields.ts là một nguồn cho ba bề mặt (policy/consent/disclosure). 12 mệnh đề §1 (11 MUST + 1 MUST NOT). Không bảng DB. PDPL P đạt. Đạt 10/10, PASS.

## §2 - Findings

- ISS-001 (frontmatter A): id/module/folder khớp; key đủ; depends_on=[TASK-EXT-003]. Pass.
- ISS-002 (PDPL P): mỗi trường gắn legalBasis PDPL (TASK-COMPLY-001); tối thiểu hóa là nguyên tắc PDPL Luật 91/2025 không chỉ tự nguyện (§1 #6); cấm tuyên bố tuyệt đối mâu thuẫn B2B aggregate ẩn danh (§1 #11) - trung thực có điều kiện. Chính xác. Pass.
- ISS-003 (contract D): collected-fields.ts struct CollectedField (field/purpose/legalBasis) + NEVER_COLLECTED; parity test import ALLOWED_ITEM_FIELDS. Không bảng DB (policy/test). Pass.
- ISS-004 (AC/test E,F): 12 AC; test policy-allowlist-parity (hai chiều) + NEVER_COLLECTED chứa cookie/mật khẩu/token + grep cam kết "KHÔNG". Pass.
- ISS-005 (typography O): dấu tiếng Việt trong TS string/markdown §3 (code block, scoped out); prose ASCII thuần; không banned word. Pass.
- ISS-006 (§6-§11, M, N): đủ khung; failure-modes 8 dòng; sentinel có; self-contained, không TBD. Pass.

## §3 - Traceability §1 -> AC -> artefact

| §1 clause | AC | Artefact |
|---|---|---|
| #1 đúng tập dữ liệu | #1,#7 | DATA-MINIMIZATION-POLICY.md + parity test |
| #2 mục đích từng trường | #1 | collected-fields.ts purpose |
| #3 không cookie/PII | #2 | NEVER_COLLECTED |
| #4 local-first | #3 | chính sách + data-flow |
| #5 data-flow điểm rời máy | #4 | data-flow.md |
| #6 cơ sở pháp lý PDPL | #5 | collected-fields.ts legalBasis |
| #7 nguồn sự thật chung | #6 | collected-fields.ts |
| #8 parity test | #7,#8 | policy-allowlist-parity.test.ts |
| #9 retention/DSAR | #9 | chính sách + TASK-COMPLY-003 |
| #10 ngôn ngữ phổ thông | #10 | rà đọc hiểu |
| #11 không tuyên bố tuyệt đối | #11 | review + B2B điều kiện |
| #12 tham chiếu chéo | #12 | DISCLOSURE.md + UI consent |

## §4 - Kết luận

Parity test khóa chính sách vào pipeline thật; một nguồn cho ba bề mặt; mỗi mục đích gắn cơ sở pháp lý PDPL. Không defect. Score = 10/10. Verdict: PASS.

---

*Hết audit TASK-TRUST-002.*
