---
fr_id: TASK-AFFIL-002
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

TASK-AFFIL-002 đặc tả `POST /v1/affiliate/link` ở mức triển khai được, và là điểm thực thi kỹ thuật của lằn ranh đạo đức trung tâm hậu-Honey. 12 mệnh đề §1 normative, mỗi mệnh đề có AC và test trong §5. Bốn bất biến compliance được khóa cứng: bắt buộc cờ `user_initiated` (không ghi click khi thiếu), bắt buộc `disclosure` + `target_url` hiển thị, server không chạm cookie sàn, cấm batch/prefetch. Link và ghi click nguyên tử. Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - Đường tạo link nền (Honey-style)
Nếu endpoint chấp nhận tạo link mà không có chủ ý user, SănDeal rơi đúng mô hình bị Chrome gỡ + network đình chỉ. Giải: §1 #1 + DEC-AFFIL-06 - bắt buộc `user_initiated:true`, từ chối 400 VÀ không ghi click khi thiếu; AC #2 + `TestLink_NotUserInitiated_Rejected_NoClick`; ràng buộc chéo NFR-AFFIL-001.

### ISS-002 - Giấu đích / thiếu disclosure
Hưởng hoa hồng mà không cho user biết đi đâu là kiểu hành vi mất niềm tin. Giải: §1 #6/#7/#10 + DEC-AFFIL-08 - response bắt buộc `target_url` (domain sàn) + `disclosure` không rỗng; AC #1/#6/#8 + `TestLink_HappyPath_HasDisclosureAndTarget`.

### ISS-003 - Server chèn cookie sàn
Tự set cookie affiliate trên domain sàn chính là cookie-stuffing. Giải: §1 #8 + DEC-AFFIL-09 - handler chỉ trả URL, cookie do trang sàn set khi user mở deep link; AC #11 (review + test không có HTTP client tới sàn).

### ISS-004 - Link không khớp ghi click + batch trá hình
Trả link mà không ghi (sổ cái lệch) hoặc prefetch nhiều link (tự động hóa nền). Giải: §1 #4 (nguyên tử, lỗi ghi -> không trả link) + §1 #9 (một bấm một link); AC #9/#10 + `TestLink_RecordClickFails_NoLink`.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 bắt buộc user_initiated | #2 | `link.go` guard + `TestLink_NotUserInitiated_Rejected_NoClick` |
| #3 product tồn tại | #3 | `products.Get` + `TestLink_ProductNotFound_404` |
| #4 link+click nguyên tử | #4,#9 | `RecordClick` + `TestLink_RecordClickFails_NoLink` |
| #5 deep link theo network | #5,#7 | `BuildDeepLink` + `TestLink_NoNetwork_503_NoClick` |
| #6 disclosure+target_url | #1,#6,#8 | `LinkResponse` + `TestLink_HappyPath_HasDisclosureAndTarget` |
| #8 không chạm cookie sàn | #11 | review + handler chỉ trả JSON |
| #9 cấm batch | #10 | grep router một route |
| #10 disclosure không rỗng | #8 | `Disclosure()` |

## §4 - Kết luận

Toàn bộ mệnh đề normative có code/test backing, gồm bốn bất biến compliance (user-initiated, disclosure+target, no-cookie-sàn, no-batch) và tính nguyên tử link-click. Không có mệnh đề "mồ côi". Score = 10/10. Verdict: PASS. Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit TASK-AFFIL-002.*
