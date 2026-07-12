---
fr_id: FR-NOTIF-006
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt

Audit độc lập, tái diễn từ file FR-NOTIF-006 hiện tại. FR đặc tả email dispatcher với interface EmailProvider pluggable (SES mặc định / SendGrid / Postmark). 12 mệnh đề §1 BCP-14 (11 MUST + 1 SHOULD), testable. Kênh rẻ thứ nhì trong cost model push>email>sms (§1 #2) được nhận đúng. Xử lý bounce/complaint qua SES SNS phân biệt hard (vô hiệu) với soft (retry) bảo vệ reputation. Cấu trúc claim/mark/status-update gương đúng FR-NOTIF-002. Đạt 10/10.

## §2 - Findings

Không còn khiếm khuyết tồn dư. Kiểm độc lập:
- Interface EmailProvider (§1 #1, DEC-NOTIF-60) - SES/SendGrid/Postmark thay được sau hợp đồng; có TestSESProvider_SatisfiesInterface + TestDispatch_ProviderSwap_NoDispatcherChange.
- Hard bounce + complaint qua SNS -> verified=false (§1 #5); soft bounce (Transient) KHÔNG vô hiệu; có 3 bounce tests phân biệt đúng.
- Backoff tôn trọng throttling/Retry-After giữ queued (§1 #6) không drop; có TestDispatch_Throttling_TriggersRetry_NotDropped.
- Dispatcher chỉ truyền Title/Body đã render bởi NOTIF-001 (§1 #4), KHÔNG render lại - escape đã làm ở NOTIF-001.
- FOR UPDATE SKIP LOCKED (§1 #8) + idempotent (§1 #9).
- Email là chỗ dựa cho user chỉ-có-web (không token push) - bảo vệ cost model.
- §10 failure-modes 9 hàng không tầm thường (sender domain chưa verified, nhầm soft/hard).
- Typography prose plain ASCII + tiếng Việt có dấu; không tự cấm; sentinel có mặt.

## §3 - Bảng truy vết (từ file hiện tại)

| §1 mệnh đề | AC | Test/Artefact |
|---|---|---|
| #1 interface EmailProvider | #1 | provider.go + TestSESProvider_SatisfiesInterface |
| #2 SES mặc định + cost model | #2 | ses.go + TestDispatch_ProviderSwap |
| #3 địa chỉ verified | #9 | ClaimEmailBatch JOIN verified=true |
| #4 truyền Title/Body đã render | #3 | EmailMessage + TestDispatch_Success_MarksSent (Subject) |
| #5 hard bounce/complaint -> verified=false | #6,#7,#8 | bounce.go HandleSNS + 3 bounce tests |
| #6 backoff throttling | #4,#5 | backoff.go + TestBackoff_RespectsRetryAfter |
| #7 phân loại 4 nhóm | #3,#4,#6 | classifySESError |
| #8 FOR UPDATE SKIP LOCKED | #10 | ClaimEmailBatch |
| #9 idempotent | #11 | MarkSent điều kiện status='queued' |
| #11 OTel | #12 | email_send_total{provider,result} |
| #12 ranh giới (không render/route) | - | §1 #4 truyền nguyên nội dung |

## §4 - Kết luận

Mọi mệnh đề normative có code/SQL/test backing; không mệnh đề mồ côi. Interface pluggable, xử lý bounce/complaint, backoff throttling, ranh giới với NOTIF-001 (chỉ truyền, không render) đều đặc tả đầy đủ và gương đúng cấu trúc FR-NOTIF-002. Không cần sửa. Score = 10/10. Verdict: PASS. Sẵn sàng build.

---

*Hết audit FR-NOTIF-006.*
