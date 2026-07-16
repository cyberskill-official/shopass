---
nfr_id: NFR-NOTIF-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: nfr-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Tái thẩm độc lập từ file NFR hiện tại. NFR-NOTIF-001 đặt SLO định lượng cho đỉnh 00:00: không bucket phút nào vượt `FCM_RATE_LIMIT_PER_MIN = 600.000` tin/phút/project; `fcm_429_total = 0` ở cửa sổ 00:00; hấp thụ đột biến >2x trong 30s-2 phút đầu giờ bằng jitter [-90s,+180s] + spread overflow; channel mix push > email > sms. 6 mệnh đề §1 đều đo được qua metric cụ thể và có verify (load/surge/throughput/channel-priority/429 test). Số khớp §3.6 nguồn (600k/phút/project, 429 RESOURCE_EXHAUSTED, >2x, jitter [-90s,+180s], cost model push>email>sms). related_tasks (TASK-NOTIF-002/003/004) đều resolve. Đạt 10/10.

## §2 - Findings (đã kiểm)

Kiểm frontmatter: id=NFR-NOTIF-001 khớp tên file; category=scalability, priority=MUST, verification=T, phase=P1; slo định lượng đa-điều-kiện; source §3.6. Đạt.

Kiểm số nguồn §3.6: line 261 FCM 600.000 messages/phút/project, vượt -> HTTP 429 RESOURCE_EXHAUSTED, traffic tăng >2 lần trong 30s-2 phút đầu mỗi giờ; line 250 jitter random(-90s, +180s); line 267 ưu tiên push > email > SMS. §1 #1/#2/#3/#4 dùng đúng các con số này. Không lệch.

Kiểm §1: 6 clause BCP-14. #1 bất biến `notif_scheduler_max_bucket_size <= 600.000`. #3 jitter + spread overflow của TASK-NOTIF-004 hấp thụ surge không drop. #5 fan-out rút kịp 1.500.000 alert/đỉnh rải qua ~3 phút (600k+600k+300k - cộng đúng). #6 dispatcher backoff theo Retry-After + DLQ. Đo được.

Kiểm §3 đo lường: gauge `notif_scheduler_max_bucket_size`, counter `fcm_429_total`, histogram `notification_dispatch_latency_ms`, counter `notification_routing_total{chosen_channel}`, gauge `notif_peak_traffic_ratio`. Cụ thể.

Kiểm §4 verification: load test seed 1.500.000 alert cùng 00:00:00 -> assert không bucket > 600.000 và 429=0; surge test >2x; throughput test (mỗi bucket phát trong phút của nó); channel-priority test (sms chỉ khi push+email không khả dụng); 429 handling test (backoff + DLQ). Mỗi clause khóa có test.

Kiểm §5: bucket > 600k hoặc 429 ở đỉnh sev-2, dispatch latency cao sev-3, cost spike channel mix sai sev-3. Hợp lý.

Kiểm khớp pseudocode §3.6 (scheduleMidnightAlerts): jitter, groupBy bucket 60s, spreadAcrossNextMinutes khi vượt trần - khớp §1 #1/#3. Đạt.

Kiểm typo: prose ASCII thuần, tiếng Việt đủ dấu, không từ cấm; backtick metric/ký hiệu hợp lệ. Không sửa gì.

## §3 - Kết luận

SLO đo được, verify được, gắn cơ chế TASK-NOTIF-004 (san phẳng) / 003 (fan-out) / 002 (429) và số §3.6 (trần 600k, đột biến >2x, cost model kênh). Không tìm thấy defect cần sửa. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập NFR-NOTIF-001.*
