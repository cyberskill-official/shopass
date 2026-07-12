---
fr_id: FR-NOTIF-004
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt

Audit độc lập, tái diễn từ file FR-NOTIF-004 hiện tại. FR đặc tả midnight-spike scheduler san phẳng đỉnh 00:00. 12 mệnh đề §1 BCP-14 (10 MUST + 2 SHOULD), testable. Kiểm khớp rubric chuyên biệt: jitter [-90s,+180s] bất đối xứng (§1 #1, JitterMin=-90s/JitterMax=+180s); gom bucket theo phút bằng floor(dispatch_at/60s) (§1 #2, minuteKey); spread overflow khi vượt trần (§1 #4); né mốc tròn :00/:15/:30/:45 (§1 #6, isRoundMark + nextStart+1s); FCM_RATE_LIMIT_PER_MIN=600.000 (§1 #3). Bất biến mỗi bucket <= 600k có test trực tiếp. Đạt 10/10.

## §2 - Findings

Không còn khiếm khuyết tồn dư. Kiểm độc lập:
- jitter bất đối xứng [-90s,+180s] (§1 #1, #2) đúng hướng; có TestJitter_WithinAsymmetricBounds kiểm cả hai cận. Hàng failure-mode "jitter đối xứng nhầm" chốt.
- Bucket floor(dispatch_at/60s) (§1 #2) - đúng đại lượng trần FCM tính theo phút.
- spread giữ lại đúng trần, đẩy phần dư sang phút kế, lặp tới hội tụ (§1 #4, #5); có TestBucket_NeverExceedsLimit + TestOverflow_SpreadsToNextMinute.
- Né mốc tròn: nextStart = đầu phút + 1s, isRoundMark định nghĩa rõ (§1 #6); có TestSpread_AvoidsRoundMarks.
- RNG tiêm được (§1 #8) làm ScheduleAlerts thuần; sortStable theo high-value + id (§1 #10) tất định; có TestScheduleAlerts_Deterministic.
- Lô lớn hội tụ không mất alert; có TestLargeBatch_Flattened (5x trần, lên giữ nguyên, >=5 phút).
- Ranh giới cứng: chỉ đặt scheduled_at, KHÔNG gọi nhà cung cấp (§1 #9, DEC-NOTIF-35).
- BatchSetScheduledAt một UPDATE hàng loạt (§1 #8) không N+1.
- Typography prose plain ASCII + tiếng Việt có dấu; không tự cấm; sentinel có mặt.

## §3 - Bảng truy vết (từ file hiện tại)

| §1 mệnh đề | AC | Test/Artefact |
|---|---|---|
| #1 jitter [-90s,+180s] | #1,#2 | applyJitter + JitterMin/Max + TestJitter_WithinAsymmetricBounds |
| #2 bucket theo phút | #3 | bucketByMinute + minuteKey |
| #3 hàng 600.000 | #4 | FCM_RATE_LIMIT_PER_MIN |
| #4 spread overflow | #5,#6 | spreadAcrossNextMinutes + TestOverflow_SpreadsToNextMinute |
| #5 bất biến bucket<=trần | #4 | TestBucket_NeverExceedsLimit |
| #6 né mốc tròn | #7 | isRoundMark + nextStart+1s + TestSpread_AvoidsRoundMarks |
| #7 ScheduleAlerts thuần | #8 | ScheduleAlerts |
| #8 rng tiêm + BatchSet | #8,#12 | RNG interface + BatchSetScheduledAt + TestScheduleAlerts_Deterministic |
| #9 không tự gửi | #11 | không import client FCM trong package |
| #10 thứ tự tất định | #9 | sortStable high-value + id |
| #11 ưu tiên high-value | #10 | sortStable giữ high-value phút sớm |
| #12 OTel | #12 | notif_scheduler_max_bucket_size (<=600k) |

## §4 - Kết luận

Mọi mệnh đề normative có code/test backing; pseudo-code §3.5(5) được encode đúng từng bước (jitter -> bucket -> spread). Bất biến sống còn "không bucket phút nào vượt 600.000" có test trực tiếp, lô lớn hội tụ không mất alert, kết quả tất định với seeded rng, ranh giới chỉ-đặt-scheduled_at với fan-out FR-NOTIF-003 rõ ràng. Không cần sửa. Score = 10/10. Verdict: PASS. Sẵn sàng build.

---

*Hết audit FR-NOTIF-004.*
