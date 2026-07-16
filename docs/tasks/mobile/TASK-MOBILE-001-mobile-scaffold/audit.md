---
fr_id: TASK-MOBILE-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

TASK-MOBILE-001 đặc tả scaffold mobile + auth + push ở mức triển khai được. 12 mệnh đề §1 normative, mỗi mệnh đề có AC và test. Ranh giới bảo mật mobile được giữ chặt: refresh token trong Keychain/Keystore, access token chỉ trong bộ nhớ, không bao giờ mật khẩu trên đĩa. Auth dùng lại JWT của TASK-AUTH-002 (không phát minh lại). Push qua FCM cho cả hai nền tảng, device token gỡ liên kết khi đăng xuất. Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - Lưu token sai chỗ (đã chốt)
Cám dỗ lớn nhất trên RN là dùng AsyncStorage cho mọi thứ - nhưng đó là plaintext, đọc được khi thiết bị bị root. Giải: §1 #2 + DEC-MOBILE-02 ép refresh token vào Keychain/Keystore, access chỉ trong bộ nhớ; AC #2/#3 + test kiểm AsyncStorage không nhận token.

### ISS-002 - Bề mặt xác thực thứ hai
Phát hành auth riêng cho mobile tạo thêm một hệ thống phải kiểm toán. Giải: §1 #1 + DEC-MOBILE-02 dùng lại JWT TASK-AUTH-002; review + disallowed_tools.

### ISS-003 - Rò push sang phiên khác
Token không gỡ khi đăng xuất làm push của người dùng cũ tới thiết bị người dùng mới. Giải: §1 #9 + DEC-MOBILE-04 unregisterDevice khi logout; test `logout gỡ liên kết device token`.

### ISS-004 - Vòng lặp refresh + trải nghiệm 401
Refresh vô hạn treo app; bắt đăng nhập lại mỗi lần hết hạn thì tệ. Giải: §1 #5 + DEC-MOBILE-05 refresh + retry đúng một lần, hỏng thì logout; test `401 -> refresh -> retry` + `refresh hỏng -> logout`.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 JWT TASK-AUTH-002 | #4 | `authClient.ts` |
| #2 refresh trong secure storage | #2,#3 | `tokenStore.ts` + test AsyncStorage |
| #3 không lưu mật khẩu | #11 | review + test |
| #4 gọi qua gateway + Bearer | #4 | `httpClient.ts` |
| #5 auto-refresh 401 | #4,#5 | `httpClient.ts` + 2 test |
| #6 đăng FCM token | #6 | `registerDevice.ts` |
| #7 từ chối quyền không chặn app | #7 | `từ chối quyền push` test |
| #8 token xoay | #8 | onTokenRefresh |
| #9 gỡ token khi logout | #9,#10 | `unregisterDevice` + test |
| #12 logout xóa sạch | #10 | `clearTokens` + test |

## §4 - Kết luận

Toàn bộ mệnh đề normative có code/test backing. Ranh giới bảo mật (secure storage, không mật khẩu, gỡ token) được kiểm bằng test cụ thể. Không mệnh đề mồ côi. Score = 10/10. Verdict: PASS. Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit TASK-MOBILE-001.*
