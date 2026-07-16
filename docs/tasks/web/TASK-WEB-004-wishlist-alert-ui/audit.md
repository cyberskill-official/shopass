---
fr_id: TASK-WEB-004
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt

Audit độc lập, tái diễn từ file TASK-WEB-004 hiện tại. task đặc tả UI quản lý wishlist + alert, tiêu thụ API REST TASK-TRACK-002/003 qua lib/api.ts. 12 mệnh đề §1 BCP-14 (11 MUST + 1 SHOULD), testable. Form alert phản chiếu chính xác ngữ nghĩa threshold theo rule_type của TASK-TRACK-003 (ô động), bộ chọn channel khớp enum {push,email,sms}, tiền giữ int64 VND. Lỗi IDOR-trùng-lặp (403/404 = "không tìm thấy") bảo vệ quyền riêng tư đúng PDPL. Đạt 10/10.

## §2 - Findings

Không còn khiếm khuyết tồn dư. Kiểm độc lập:
- Ô threshold động theo rule_type (§1 #6, DEC-WEB-17): price_below ô giá VND, drop_pct ô % [1,99], real_sale/bottom_predicted KHÔNG ô threshold; validateAlert mirror DEC-TRACK-22; có alert-form-validate test.
- channel allowlist {push,email,sms} push mặc định (§1 #8) khớp enum TASK-TRACK-003 + dispatcher có thật; có alert-channel test.
- target_price/threshold int VND (§1 #3, DEC-WEB-19) - không float; có wishlist-api test Number.isInteger.
- Validate client báo sớm nhưng server vẫn là kiểm tra cuối (§1 #7) - defense-in-depth.
- Lỗi 403/404 trùng lặp (§1 #9, DEC-WEB-20) - không lộ tồn tại tài nguyên người khác (IDOR đã chặn ở TRACK-002/003).
- Server là nguồn sự thật (§1 #1, DEC-WEB-16) - client không giữ danh sách làm nguồn chính.
- §10 failure-modes 9 hàng không tầm thường (bỏ qua validate client gọi API thẳng).
- Typography prose plain ASCII + tiếng Việt có dấu; không tự cấm; sentinel có mặt.

## §3 - Bảng truy vết (từ file hiện tại)

| §1 mệnh đề | AC | Test/Artefact |
|---|---|---|
| #1 tiêu thụ API TRACK | #1 | lib/wishlist/api.ts + lib/alerts/api.ts |
| #2 CRUD wishlist | #2 | wishlist-panel.tsx |
| #3 target_price int VND | #3 | addItem + wishlist-api test |
| #4 CRUD alert | #4 | alert-list.tsx |
| #5 rule_type enum | #5 | select 4 giá trị |
| #6 ô threshold động | #6 | needsThreshold + alert-form-validate |
| #7 validate client + server cuối | #7 | validateAlert |
| #8 channel allowlist | #8 | CHANNELS + alert-channel test |
| #9 403/404 trùng lặp | #9 | thông báo trùng lặp |
| #10 trong (app) + JWT | #10 | route group + apiFetch |
| #11 trạng thái rỗng/lỗi | #11 | empty-state |
| #12 tsc/test | #12 | npm test |

## §4 - Kết luận

Mọi mệnh đề normative có mã/test backing; form phản chiếu đúng hợp đồng TASK-TRACK-002/003 (rule_type enum, threshold theo loại, channel {push,email,sms}), tiền int64, lỗi IDOR-trùng-lặp. Validate client báo sớm nhưng server vẫn kiểm tra cuối. Không mệnh đề mồ côi. Không cần sửa. Score = 10/10. Verdict: PASS. Sẵn sàng build.

---

*Hết audit TASK-WEB-004.*
