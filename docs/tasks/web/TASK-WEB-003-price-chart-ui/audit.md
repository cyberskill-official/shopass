---
fr_id: TASK-WEB-003
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt

Audit độc lập, tái diễn từ file TASK-WEB-003 hiện tại. task đặc tả UI biểu đồ lịch sử giá. 12 mệnh đề §1 BCP-14 (11 MUST + 1 SHOULD), testable. Kiểm khớp rubric chuyên biệt: UI đọc price_daily QUA feed TASK-DEAL-003 (GET /v1/products/{id}/chart) - KHÔNG gọi raw price_snapshot (§1 #1, DEC-WEB-11), KHÔNG tự tính verdict/median ở client (§1 #5, DEC-WEB-13). Tôn trọng maturity (NEW không kết luận, §1 #6) và allowlist range {7d,30d,90d,180d,1y} (§1 #7) giữ đúng hợp đồng feed và lời hứa không kết luận ẩu. Đạt 10/10.

## §2 - Findings

Không còn khiếm khuyết tồn dư. Kiểm độc lập:
- Tiêu thụ DUY NHẤT feed TASK-DEAL-003 (§1 #1) - không gọi raw, không recompute; có AC #1 grep không raw.
- Vẽ đường median90 + trailing_min lấy thẳng từ annotations (§1 #3) - không tự tính.
- Verdict lấy thẳng annotations.verdict (§1 #5) - khớp nhãn thẻ sản phẩm; grep client không có hàm tính verdict; có verdict-badge test.
- Tôn trọng maturity NEW (§1 #6) ẩn badge + ghi chú đang thu thập; có test "NEW không hiện badge".
- Range allowlist (§1 #7) - range là nêm lỗi không gọi mạng; có range-selector test.
- p95 <500ms (§1 #8) - feed downsample, client chỉ vẽ.
- Tiền int64 VND (§1 #11) format vi-VN, không float.
- §10 failure-modes 9 hàng không tầm thường (daily rỗng, payload phình).
- Typography prose plain ASCII + tiếng Việt có dấu; không tự cấm; sentinel có mặt.

## §3 - Bảng truy vết (từ file hiện tại)

| §1 mệnh đề | AC | Test/Artefact |
|---|---|---|
| #1 chỉ tiêu thụ feed (price_daily) | #1 | fetch-chart.ts |
| #2 vẽ thân daily | #2 | price-chart.tsx |
| #3 đường median90/trailing_min | #3 | overlay từ annotations |
| #4 mốc ngày đôi | #4 | double-date markers |
| #5 verdict không suy diễn | #5 | verdict-badge.tsx |
| #6 tôn trọng maturity | #6 | maturity-notice.tsx + badge ẩn NEW |
| #7 allowlist range | #7 | RANGE_ALLOWLIST + validate |
| #8 p95 <500ms | - | feed downsample, client chỉ vẽ |
| #9 trạng thái rỗng/lỗi | #8 | empty + 404 handling |
| #10 trong (app) + JWT | #9 | route group + apiFetch |
| #11 tiền int64 | #10 | format vi-VN |
| #12 tsc/test | #11 | npm test |

## §4 - Kết luận

Mọi mệnh đề normative có mã/test backing; ranh giới client-chỉ-vẽ + nhãn verdict một-nguồn được kiểm chứng bằng test. UI đọc price_daily qua feed TASK-DEAL-003, không recompute verdict client-side - đúng đặc tả. Tôn trọng maturity + allowlist range khớp hợp đồng feed. Không cần sửa. Score = 10/10. Verdict: PASS. Sẵn sàng build.

---

*Hết audit TASK-WEB-003.*
