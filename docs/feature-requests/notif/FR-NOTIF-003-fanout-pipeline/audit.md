---
fr_id: FR-NOTIF-003
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt

Audit độc lập, tái diễn từ file FR-NOTIF-003 hiện tại. FR đặc tả fan-out pipeline: producer -> Kafka/Redis Streams -> fan-out worker -> per-channel dispatcher, at-least-once + idempotent, backoff + jitter + DLQ. 12 mệnh đề §1 BCP-14 (11 MUST + 1 SHOULD), testable. Kiểm khớp rubric chuyên biệt: định tuyến theo channel in {push,email,sms} - KHÔNG có giá trị apns riêng (§1 #4); push iOS phân theo user_channel_token.platform='ios' (FR-NOTIF-005), KHÔNG phải channel riêng; fan-out KHÔNG gọi nhà cung cấp. Đạt 10/10.

## §2 - Findings

Không còn khiếm khuyết tồn dư. Kiểm độc lập:
- Idempotent qua claim/lease CAS (UPDATE ... WHERE status='pending' RETURNING) (§1 #3) - chỉ worker thắng gửi; có TestFanout_AtLeastOnce_NoDoubleSend + ConcurrentClaim_SingleSend (8 goroutine).
- Định tuyến channel {push,email,sms}, không apns (§1 #4); push iOS phân theo platform - đúng rubric. Có TestFanout_RoutesByChannel.
- Permanent vs Transient (§1 #7): Permanent -> DLQ ngay; Transient -> retry tới max_attempts rồi DLQ. Có TestFanout_PermanentStraightToDLQ + MaxAttempts_ToDLQ.
- Backoff full jitter (§1 #5) chống thundering herd; có TestBackoff_HasJitter + GrowsAndCaps.
- Re-claim lease hết hạn (§1 #9) chống job kẹt ở queued; có TestFanout_ReclaimExpiredLease.
- Phân vùng theo user_id (§1 #11) giữ thứ tự per-user.
- Ranh giới cứng: fan-out chỉ định tuyến, gọi thật ở NOTIF-002/005/006/007 (§1 #4).
- §10 failure-modes 11 hàng không tầm thường (DLQ phình, double_claim, worker crash).
- Typography prose plain ASCII + tiếng Việt có dấu; không tự cấm; sentinel có mặt.

## §3 - Bảng truy vết (từ file hiện tại)

| §1 mệnh đề | AC | Test/Artefact |
|---|---|---|
| #1 đường ống | #2,#3 | producer.go + worker.go + dispatch.go |
| #2 at-least-once | #4 | TestFanout_AtLeastOnce_NoDoubleSend |
| #3 idempotent claim/lease CAS | #4,#5 | ClaimPending + TestFanout_ConcurrentClaim_SingleSend |
| #4 định tuyến channel (no apns) | #11 | Router + TestFanout_RoutesByChannel |
| #5 backoff + jitter | #12 | backoff.go + TestBackoff_HasJitter |
| #6 DLQ khi cần retry | #7 | 0004_notification_dlq.sql + TestFanout_MaxAttempts_ToDLQ |
| #7 Permanent vs Transient | #6,#8 | errClass + TestFanout_PermanentStraightToDLQ |
| #8 cập nhật status | #3,#7,#8 | worker.go MarkSent/dead |
| #9 re-claim lease hết hạn | #10 | ClaimPending WHERE lease_until<now + TestFanout_ReclaimExpiredLease |
| #10 cột attempts/lease + index | #1 | 0003_notification_lease.sql |
| #11 phân vùng user_id | - | producer.go partition key |
| #12 OTel | #4 | notif_fanout_* + double_claim_total |

## §4 - Kết luận

Mọi mệnh đề normative có code/SQL/test backing; không mệnh đề mồ côi. Bốn rủi ro cốt lõi của pipeline gửi-quy-mô (gửi đôi, retry vô hạn, thundering herd, job kẹt) đều có cơ chế + test cụ thể. channel chỉ {push,email,sms} không apns; push iOS phân theo platform - đúng đặc tả. Ranh giới với NOTIF-001 + dispatcher downstream sạch. Không cần sửa. Score = 10/10. Verdict: PASS. Sẵn sàng build.

---

*Hết audit FR-NOTIF-003.*
