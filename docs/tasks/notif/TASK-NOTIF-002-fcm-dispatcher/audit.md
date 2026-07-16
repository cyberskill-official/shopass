---
fr_id: TASK-NOTIF-002
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt

Audit độc lập, tái diễn từ file TASK-NOTIF-002 hiện tại. task đặc tả dispatcher FCM Web/Android. 12 mệnh đề §1 BCP-14 (10 MUST + 2 ranh giới/SHOULD), testable. Kiểm khớp rubric chuyên biệt: FCM HTTP v1 (không legacy, §1 #1); lấy token verified=true và platform IN ('android','web') (§1 #2) - token platform='ios' thuộc APNs, KHÔNG bị FCM nhặt; token bucket 600k/phút/project (§1 #4); 429 RESOURCE_EXHAUSTED backoff + tôn trọng Retry-After, không drop (§1 #5); UNREGISTERED -> verified=false (§1 #7). Đạt 10/10.

## §2 - Findings

Không còn khiếm khuyết tồn dư. Kiểm độc lập:
- HTTP v1 messages:send, KHÔNG legacy /fcm/send (§1 #1); có TestSend_UsesHTTPv1Endpoint_NotLegacy khóa cứng path.
- ClaimPushBatch lọc platform IN ('android','web') (§1 #2) - ranh giới với APNs (NOTIF-005) sạch.
- Token bucket 600k/phút (§1 #4, DEC-NOTIF-11) vẫn chủ động trước 429; có TestQuota_BlocksOverLimitThenRefills.
- 429 backoff + Retry-After giữ queued (§1 #5) không drop; có TestSend_429_TriggersRetry + RespectsRetryAfter.
- UNREGISTERED/INVALID_ARGUMENT -> verified=false + MarkFailed (§1 #7); có TestDispatch_Unregistered_InvalidatesTokenAndFails.
- FOR UPDATE SKIP LOCKED (§1 #8) chống gửi trùng khi scale ngang; MarkSent/MarkFailed điều kiện status='queued' (§1 #9).
- Ranh giới cứng với NOTIF-001 (§1 #12): không routing/render.
- §10 failure-modes 8 hàng không tầm thường (OAuth hết hạn, gửi lại dòng sent, thundering herd).
- Typography prose plain ASCII + tiếng Việt có dấu; không tự cấm; sentinel có mặt.

## §3 - Bảng truy vết (từ file hiện tại)

| §1 mệnh đề | AC | Test/Artefact |
|---|---|---|
| #1 HTTP v1 không legacy | #1 | client.go Send + TestSend_UsesHTTPv1Endpoint_NotLegacy |
| #2 token verified + platform android/web | #8 | ClaimPushBatch JOIN verified=true |
| #3 JSON message HTTP v1 | #1 | message.go + §8 payload |
| #4 token bucket 600k/phút | #10 | quota.go Bucket + TestQuota_BlocksOverLimitThenRefills |
| #5 429 backoff + Retry-After | #3,#4,#5 | backoff.go nextDelay + 2 test 429 |
| #6 classify phản hồi | #2,#3,#6,#7 | client.go classify |
| #7 UNREGISTERED -> verified=false | #6,#7 | InvalidateToken + TestDispatch_Unregistered |
| #8 FOR UPDATE SKIP LOCKED | #9 | ClaimPushBatch |
| #9 idempotent không gửi lại | #11 | MarkSent/MarkFailed điều kiện queued |
| #11 OTel | #12 | fcm_send_total{result} |
| #12 ranh giới NOTIF-001 | - | chỉ gửi, không routing/render |

## §4 - Kết luận

Mọi mệnh đề normative có code/test backing; không mệnh đề mồ côi. Ba rủi ro load-bearing §3.6 (API v1 thay legacy, 429 không drop mà backoff + Retry-After, token chết bị tắt) đều được đặc tả + test khóa. Phần platform android/web giữ ranh giới với APNs. Không cần sửa. Score = 10/10. Verdict: PASS. Sẵn sàng build.

---

*Hết audit TASK-NOTIF-002.*
