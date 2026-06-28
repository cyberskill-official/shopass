---
fr_id: FR-NOTIF-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt

Audit độc lập, tái diễn từ file FR-NOTIF-001 hiện tại. FR đặc tả bảng notification, template engine render an toàn, và routing kênh theo cost model push>email>sms. 12 mệnh đề §1 BCP-14 (11 MUST + 1 SHOULD), testable. Kiểm khớp rubric chuyên biệt: notification.channel CHECK in {push,email,sms} - KHÔNG có 'apns' (DDL §3 line CHECK channel IN ('push','email','sms')); user_channel_token có cột platform CHECK in {ios,android,web,email,sms}, PK (user_id, channel, platform); ghi chú rõ FCM (NOTIF-002) nhặt platform IN ('android','web'), APNs (NOTIF-005) nhặt platform='ios'. Đạt 10/10.

## §2 - Findings

Không còn khiếm khuyết tồn dư. Kiểm độc lập:
- channel CHECK chỉ {push,email,sms}; không giá trị apns riêng - đúng yêu cầu rubric.
- user_channel_token PK (user_id, channel, platform) cho phép một user có token ios + android + web cùng lúc; platform tách push thành hai nhánh gửi.
- Routing cost model push>email>sms (§1 #7, DEC-NOTIF-03) qua channelRank; SMS chỉ khi high-value hoặc hết kênh khác (§1 #8) - bảo vệ unit economics.
- Template escape an toàn + lỗi-khi-thiếu-biến (§1 #5, #6) chống injection + placeholder lộ; có TestRender_EscapesPayload + MissingVar_Errors.
- Ranh giới cứng: chỉ ghi pending, KHÔNG gọi nhà cung cấp (§1 #9, DEC-NOTIF-05); tầng gửi thuộc NOTIF-002..007.
- Tiền int64 VND trong payload (§1 #11), format ở render.
- DATA-MODEL line 178 ghi PK(user_id,channel) không có platform - FR phát triển schema sâu hơn §3.4 dùng quy ước catalog, FR là nguồn thực thi; không phải lỗi FR.
- Typography prose plain ASCII + tiếng Việt có dấu; không tự cấm; sentinel có mặt.

## §3 - Bảng truy vết (từ file hiện tại)

| §1 mệnh đề | AC | Test/Artefact |
|---|---|---|
| #1 notification schema | #1 | 0001_notification.sql |
| #2 CHECK channel/status (no apns) | #2 | CHECK channel IN ('push','email','sms') |
| #3 user_channel_token + platform PK | #1 | 0002_user_channel_token.sql PK(user_id,channel,platform) |
| #4 template registry 4 loại | #3 | templates map |
| #5 escape an toàn | #5 | TestRender_EscapesPayload |
| #6 lỗi khi thiếu biến | #4 | TestRender_MissingVar_Errors |
| #7 ResolveChannel cost model | #6,#7 | routing.go + 5 routing tests |
| #8 SMS chỉ khi đáng | #8,#9,#10 | avail("sms") + 2 tests |
| #9 ghi pending không gửi | - | InsertNotification + ranh giới |
| #10 partial index dispatch | #12 | idx_notif_dispatch |
| #11 BIGINT VND | - | int64 + formatVND ở render |
| #12 OTel | #12 | notification_routing_total |

## §4 - Kết luận

Mọi mệnh đề normative có code/SQL/test backing; không mệnh đề mồ côi. Hai điểm rủi ro lớn nhất (cost model routing, escape template) có test. channel không 'apns', platform tách FCM/APNs - đúng đặc tả. Ranh giới với fan-out + dispatcher rõ ràng. Không cần sửa. Score = 10/10. Verdict: PASS. Sẵn sàng build.

---

*Hết audit FR-NOTIF-001.*
