---
id: NFR-NOTIF-001
title: "NOTIF scale đỉnh 00:00 - flatten the curve để không bucket phút nào vượt 600k tin/phút/project, không bao giờ ăn 429, ưu tiên kênh rẻ để chặn chi phí"
module: NOTIF
category: scalability
priority: MUST
verification: T
phase: P1
slo: "Không bucket phút nào vượt 600.000 tin/phút/project lúc đỉnh 00:00 (traffic tăng >2 lần trong 30s-2 phút đầu giờ); 0 lần ăn 429 RESOURCE_EXHAUSTED; ưu tiên kênh push > email > sms để chặn chi phí"
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-NOTIF-002, FR-NOTIF-003, FR-NOTIF-004]
source: "docs/... §3.6 (giới hạn FCM 600.000 tin/phút/project; đỉnh 00:00 traffic tăng >2x trong 30s-2 phút đầu mỗi giờ; flatten the curve; cost model push > email > sms)"
---

## §1 - Statement (BCP-14 normative)

1. Lúc đỉnh 00:00 (và mọi mốc giờ), sau khi scheduler (FR-NOTIF-004) san phẳng, KHÔNG bucket phút nào **MUST** vượt `FCM_RATE_LIMIT_PER_MIN = 600.000` tin/phút/project. Bất biến `notif_scheduler_max_bucket_size <= 600.000` là điều kiện sống còn của module.
2. Hệ **MUST** đạt 0 lần FCM trả `429 RESOURCE_EXHAUSTED` ở điều kiện đỉnh đã san phẳng; `fcm_429_total` ở cửa sổ 00:00 **MUST** = 0. Một lần 429 nghĩa là bucket đã vượt trần hoặc spread hỏng.
3. Khi traffic tăng >2 lần trong 30s-2 phút đầu mỗi giờ (đo được ở §3.6), hệ **MUST** hấp thụ đột biến đó bằng jitter `[-90s, +180s]` + spread overflow (FR-NOTIF-004) mà không drop alert nào và không đẩy bucket nào vượt trần.
4. Dưới tải đỉnh, hệ **MUST** ưu tiên kênh theo cost model push > email > sms (FR-NOTIF-001 routing): chọn kênh rẻ nhất còn khả dụng cho mỗi alert, để chi phí gửi không bùng nổ khi lượng tin tăng vọt. SMS chỉ dùng khi push và email đều không khả dụng cho alert tới hạn.
5. Fan-out (FR-NOTIF-003) **MUST** sustain throughput phát đủ để rút hết lô đỉnh đã san phẳng đúng theo `scheduled_at`: ở 1.500.000 alert/đỉnh rải qua ~3 phút (600k+600k+300k), pipeline phát kịp từng bucket trong phút của nó, không tồn dư cuốn sang phút sau.
6. FCM dispatcher (FR-NOTIF-002) **MUST** xử lý 429 có kỷ luật nếu trần bị chạm sát: backoff theo `Retry-After`, đẩy phần dư vào DLQ thay vì retry mù, sao cho một sự cố sát trần không thành bão lỗi dây chuyền quanh nửa đêm.

## §2 - Vì sao ràng buộc này

Đỉnh 00:00 là khoảnh khắc nguy hiểm nhất của toàn module thông báo. §3.6 đo được traffic tăng hơn 2 lần chỉ trong 30 giây tới 2 phút đầu mỗi giờ, và nửa đêm là đỉnh của đỉnh: flash sale Shopee/TikTok/Lazada cùng mở lúc 00:00, hàng loạt rule giá cùng chạm ngưỡng, batch dự đoán đáy giá cùng bắn. Nếu để mọi alert đập thẳng vào FCM trong một phút, một bucket phút duy nhất dễ vượt trần 600.000 và FCM trả 429; phần dư retry dồn sang phút sau làm phút đó cũng vỡ, cuộn lại thành bão lỗi kéo dài cả chục phút đúng giờ vàng. Hệ quả là đúng lúc deal nóng nhất thì người dùng không nhận được cảnh báo, hoặc nhận trễ tới mức deal đã hết - phá thẳng giá trị cốt lõi của SănDeal là báo kịp lúc. Mặt chi phí cũng vậy: nếu không ưu tiên kênh rẻ, một đêm cao điểm có thể đẩy hóa đơn SMS lên gấp nhiều lần. Đây là ràng buộc nền tảng quyết định cả trải nghiệm (alert tới đúng giờ) lẫn biên lợi nhuận (chi phí gửi có trần).

## §3 - Đo lường (measurement)

- Gauge `notif_scheduler_max_bucket_size` - bucket phút lớn nhất sau spread; **MUST** `<= 600.000` mọi thời điểm (do FR-NOTIF-004 phát).
- Counter `fcm_429_total` - số lần FCM trả 429 RESOURCE_EXHAUSTED; target 0 ở cửa sổ đỉnh 00:00.
- Histogram `notification_dispatch_latency_ms` - độ trễ từ `scheduled_at` tới lúc FCM nhận; đo fan-out có phát kịp trong phút của bucket.
- Counter `notification_routing_total{chosen_channel}` - channel mix push/email/sms; tỷ trọng push cao = chi phí có trần; sms tăng bất thường = báo động chi phí.
- Gauge peak traffic ratio `notif_peak_traffic_ratio` = lượng tin 30s-2 phút đầu giờ / baseline cùng cửa sổ - xác nhận đột biến >2x được hấp thụ chứ không tràn trần.

## §4 - Verification

- Load test (T): seed một lô midnight-burst >2x baseline (vd 1.500.000 alert cùng `event_time = 00:00:00`), chạy qua `ScheduleAlerts` (FR-NOTIF-004) rồi mô phỏng phát; assert KHÔNG bucket phút nào > 600.000 và `fcm_429_total = 0` suốt cửa sổ.
- Surge test (T): bơm traffic tăng đúng >2x trong 30s-2 phút đầu giờ; xác nhận jitter + spread hấp thụ, không drop alert (đếm `len(sched)` = đầu vào), không bucket nào vượt trần.
- Throughput test (T): fan-out (FR-NOTIF-003) rút lô đỉnh đã san phẳng; đo `notification_dispatch_latency_ms` p95 và xác nhận mỗi bucket được phát trong phút của nó, không tồn dư cuốn sang phút kế.
- Channel-priority test (T): với alert có nhiều kênh khả dụng, xác nhận routing chọn push trước email trước sms; `notification_routing_total{chosen_channel="sms"}` chỉ tăng khi push và email đều không khả dụng.
- 429 handling test (T): ép FCM trả 429 sát trần; xác nhận FR-NOTIF-002 backoff theo `Retry-After` và đẩy DLQ, không retry mù làm vỡ phút kế.

## §5 - Xử lý khi vi phạm

- Một bucket phút > 600.000 HOẶC quan sát thấy 429 ở đỉnh -> sev-2; kiểm scheduler spread có lặp tới hội tụ (FR-NOTIF-004 #4,#5) và quota headroom còn đủ; đây là van chính, vỡ là alert giờ vàng hỏng.
- `notification_dispatch_latency_ms` ở đỉnh quá cao (fan-out không rút kịp bucket trong phút của nó) -> sev-3; scale worker fan-out (FR-NOTIF-003) + kiểm DLQ/backpressure.
- Cost spike vì channel mix sai (`notification_routing_total{chosen_channel="sms"}` tăng vọt khi push/email đáng lẽ khả dụng) -> sev-3; kiểm routing push > email > sms (FR-NOTIF-001) và token/health của push.

---

*Hết NFR-NOTIF-001.*
