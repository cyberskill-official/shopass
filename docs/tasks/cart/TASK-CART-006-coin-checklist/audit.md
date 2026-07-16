---
fr_id: TASK-CART-006
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

TASK-CART-006 đặc tả checklist xu chỉ-nhắc ở mức triển khai được. 12 mệnh đề §1 normative neo vào AC §4 và test §5. Ranh giới chống-ban-High được giữ tuyệt đối và test bằng grep: không auto-click, không dispatchEvent, không API hoàn thành nhiệm vụ, không thu thập credential - đúng §3.9c (tự động hóa xu rủi ro ban cao nhất) và giải pháp tài liệu (chỉ checklist nhắc). UI minh bạch "bạn tự thực hiện". Là SHOULD về ưu tiên nhưng MUST về ranh giới an toàn. Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - Auto-click farming kích hoạt ban High (đã chốt)
Rủi ro: tự hoàn thành nhiệm vụ xu -> §3.9c ban cao nhất, khóa tài khoản user. Giải: DEC-CART-30/31 + §1 #1/#2 chỉ nhắc, chỉ đọc; AC #1 + test grep không .click/dispatchEvent/complete-task.

### ISS-002 - Mô phỏng click / API hoàn thành
Dispatch event hay gọi API hoàn thành là vượt ranh giới chỉ-đọc. Giải: DEC-CART-31 + §1 #2; AC #1 + test coin-no-autoclick grep ba mẫu.

### ISS-003 - Thu thập credential khi đọc xu
Đọc trạng thái không được chạm cookie/token. Giải: DEC-CART-34 + §1 #6; AC #7 + test reader không document.cookie/token.

### ISS-004 - UI tạo ấn tượng tự động
Nút "tự động hoàn thành" lệch triết lý + rủi ro hiểu lầm. Giải: §1 #3/#7 chỉ link dẫn user tự bấm + ghi chú minh bạch; AC #3 + #4.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 chỉ nhắc | #1 | `coin-task-reader.ts` |
| #2 chỉ đọc trạng thái | #1,#2 | `readCoinTasks` |
| #3 dữ liệu nhắc + link | #3 | `coin-checklist.ts` |
| #4 bảng coin_task | #5 | `0012_coin_task.up.sql` |
| #5 reminder nhắc nhẹ | #8 | qua TASK-NOTIF-001 |
| #6 không credential | #7 | reader grep |
| #7 UI minh bạch | #4 | showNote |
| #8 scope user | #6,#10 | `ListPending` |
| #9 đọc lỗi lịch sự | #9 | fail-state |
| #10 local-first | - | đọc client |
| #12 test xanh | #11 | npm + go test |

## §4 - Kết luận

Mọi mệnh đề normative có code/SQL/test backing; ranh giới chống-ban-High (không auto-click/dispatch/API-hoàn-thành/credential) kiểm chứng bằng grep. SHOULD về ưu tiên, MUST về ranh giới an toàn - khi build phải giữ. Không mệnh đề "mồ côi". **Score = 10/10. Verdict: PASS.** Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit TASK-CART-006.*
