---
fr_id: TASK-NOTIF-005
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt

Audit độc lập, tái diễn từ file TASK-NOTIF-005 hiện tại. task đặc tả dispatcher APNs cho iOS, đối xứng với TASK-NOTIF-002 (FCM Android/Web). 12 mệnh đề §1 BCP-14 (10 MUST + 2 SHOULD/ranh giới), testable. Kiểm khớp rubric chuyên biệt: lấy device_token thiết bị iOS qua user_channel_token platform='ios' (§1 #4, ClaimIOSPushBatch WHERE t.platform='ios') - KHÔNG nhặt token android/web; HTTP/2 token-based JWT .p8 (§1 #1, #2); 410 Unregistered -> verified=false (§1 #8); backoff 500/503 tôn trọng Retry-After (§1 #9). Đạt 10/10.

## §2 - Findings

Không còn khiếm khuyết tồn dư. Kiểm độc lập:
- ClaimIOSPushBatch lọc platform='ios' (§1 #4) - ranh giới với FCM (NOTIF-002 nhặt android/web) sạch, không gửi trùng thiết bị.
- HTTP/2 /3/device/{token} token-based, KHÔNG certificate legacy (§1 #1); JWT ES256 .p8 cache <60 phút (§1 #2); có TestSend_UsesHTTP2DeviceEndpoint_TokenAuth.
- Nhiều kết nối HTTP/2 song song + multiplex (§1 #5, DEC-NOTIF-51); có TestPool_OpensParallelConnsAndRoundRobins.
- 410 Unregistered (và 400 BadDeviceToken) -> verified=false + MarkFailed (§1 #8), tái dùng InvalidateToken của NOTIF-002; có TestDispatch_410_InvalidatesTokenAndFails.
- 500/503 backoff + Retry-After giữ queued (§1 #9) không drop; có TestSend_500_TriggersRetry + RespectsRetryAfter.
- FOR UPDATE SKIP LOCKED + idempotent (§1 #10).
- Ranh giới cứng với NOTIF-001 (§1 #12): không routing/render.
- Đối xứng đúng anh em FCM: cùng vòng đời notification, cùng pattern claim/mark.
- Typography prose plain ASCII + tiếng Việt có dấu; không tự cấm; sentinel có mặt.

## §3 - Bảng truy vết (từ file hiện tại)

| §1 mệnh đề | AC | Test/Artefact |
|---|---|---|
| #1 HTTP/2 /3/device token-based | #1 | client.go Send + TestSend_UsesHTTP2DeviceEndpoint_TokenAuth |
| #2 JWT ES256 .p8 cache <60p | #1,#2 | jwt.go ProviderToken.Bearer |
| #3 header apns-topic/push-type/priority | #1 | client.go header set |
| #4 token verified + platform=ios | #10 | ClaimIOSPushBatch WHERE platform='ios' |
| #5 nhiều kết nối HTTP/2 song song | #9 | pool.go + TestPool_OpensParallelConnsAndRoundRobins |
| #6 aps payload <4KB | - | payload.go + failure mode |
| #7 classify phản hồi | #3,#4,#5,#6 | client.go classify |
| #8 410 -> verified=false | #4,#5 | InvalidateToken + TestDispatch_410_InvalidatesTokenAndFails |
| #9 500/503 backoff + Retry-After | #6,#7,#8 | backoff.go + 2 retry tests |
| #10 claim an toàn + idempotent | #10,#11 | ClaimIOSPushBatch FOR UPDATE SKIP LOCKED |
| #11 error không panic + log apns-id | - | Send trả ResultRetry+err; SendOutcome.APNsID |
| #12 ranh giới + OTel | #12 | apns_send_total{result} |

## §4 - Kết luận

Mọi mệnh đề normative có code/test backing; không mệnh đề mồ côi. Bốn fact §3.6 (HTTP/2 nhiều kết nối song song, 410 -> verified=false, backoff 500/503, token-based JWT .p8) encode chính xác. Phần platform='ios' giữ ranh giới với FCM (android/web). Đối xứng đúng anh em TASK-NOTIF-002. Không cần sửa. Score = 10/10. Verdict: PASS. Sẵn sàng build.

---

*Hết audit TASK-NOTIF-005.*
