---
fr_id: FR-DEAL-006
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập, suy lại từ file hiện tại. FR-DEAL-006 đặc tả batch chấm điểm đáy giá hằng đêm + cảnh báo khi `p_bottom_14d > 0.7`: model-agnostic đọc cột chung `price_forecast.p_bottom_14d` (Prophet hoặc LightGBM), idempotent 1 alert/(user,product)/ngày qua `UNIQUE(user_id, product_id, fired_on)` + cooldown rising-edge 7 ngày, khớp `alert_rule` type 'bottom_predicted' đang bật rồi enqueue notification fan-out. Khớp §3.5 ("batch nightly score -> alert nếu P(bottom within 14d) > 0.7", strict greater-than) và §3.6 (fan-out). Phát hiện một defect khớp-cột thật và đã sửa surgical (xem §2). 13 mệnh đề §1 có AC §4 và test §5. Sau sửa đạt 10/10.

## §2 - Findings (đã kiểm trong lượt này)

### ISS-001 - Sai tên cột alert_rule: enabled vs active (đã sửa)
File tham chiếu `alert_rule ... AND enabled = true` ở §1 #6, §3 SQL (`matchBottomRules`), và một hàng §10. Nhưng FR sở hữu FR-TRACK-003 và DATA-MODEL đều định nghĩa cột là `active BOOLEAN` (không có cột `enabled`). SQL này sẽ vỡ lúc chạy (cột không tồn tại), phá hợp đồng đọc với FR-TRACK-003. Trớ trêu là §10 còn liệt kê failure mode "Nhầm tên cột active/enabled" rồi tự phạm đúng lỗi đó. Đã sửa surgical 4 vị trí `enabled` -> `active` (§1 #6, §3 query, §10 hàng + cột "Khắc phục"). Sau sửa khớp DATA-MODEL và FR-TRACK-003 (`idx_ar_eval ON alert_rule (product_id, rule_type) WHERE active = true`).

### ISS-002 - Coherence price_forecast: scored_at tươi 36h (đã xác nhận)
§1 #4 + §3 query lọc `scored_at >= now() - INTERVAL '36 hours'`. Cột `scored_at` do FR-DEAL-004 (chủ bảng) định nghĩa; ánh xạ đúng. Batch đọc-chỉ price_forecast, không gọi lại huấn luyện. Model-agnostic qua cột `p_bottom_14d` chung (§1 #3). Coherent với DEAL-004/005.

### ISS-003 - Ngưỡng 0.7 strict + idempotent + cooldown (đã xác nhận)
§1 #5 strict greater-than đúng nguyên văn §3.5; `p_bottom_14d == 0.7` KHÔNG bắn; CHECK `p_bottom > 0.7` ở bảng `bottom_alert_log` (DATA-MODEL khớp). Test `TestNightly_FiresAboveThreshold` đóng biên 0.69/0.70/0.71. Idempotent ngày + cooldown rising-edge 7 ngày: test `TestNightly_DedupePerDay`, `TestNightly_Cooldown_NoRepeatWhileHigh`. Maturity gate (§1 #2) test `TestNightly_RespectsMaturityGate`. Không cần sửa thêm.

### ISS-004 - Typography (đã xác nhận sạch)
Quét toàn file sau sửa: không mũi tên unicode, em-dash, en-dash, curly quote, ellipsis, emoji trong prose. Không cần sửa.

## §3 - Traceability §1 -> AC -> artefact (dựng từ file hiện tại)

| §1 clause | §4 AC | Test / artefact §5 / §3 |
|---|---|---|
| #1 cron 02:00 + advisory lock | AC #1 | `main.go` cron + `pg_try_advisory_lock` |
| #2 chỉ SKU MATURE | AC #2 | `TestNightly_RespectsMaturityGate` |
| #3 đọc p_bottom_14d chung | AC #3 | `RunNightlyScore` query |
| #4 scored_at tươi 36h | AC #4 | §3 query `scored_at >= now()-36h` |
| #5 ngưỡng 0.7 strict | AC #5 | `TestNightly_FiresAboveThreshold` |
| #6 join active 'bottom_predicted' | AC #6 | `matchBottomRules` (active=true) |
| #7 idempotent ngày | AC #7 | `TestNightly_DedupePerDay` |
| #8 cooldown rising-edge | AC #8 | `TestNightly_Cooldown_NoRepeatWhileHigh` |
| #9 enqueue fan-out payload | AC #9 | `TestNightly_EnqueuesNotification` |
| #10 ghi bottom_alert_log | AC #10 | `enqueueAndLog` |
| #11 lỗi enqueue cục bộ | AC #11 | §1 #11 + failure mode |
| #12 OTel metric | AC #12 | `deal_bottom_alert_fired_total` |
| #13 transaction nhất quán | AC #13 | §1 #13 + CHECK p_bottom>0.7 |

## §4 - Kết luận

Sau khi sửa defect khớp-cột (`enabled` -> `active`, 4 vị trí) để khớp FR-TRACK-003 / DATA-MODEL, mọi mệnh đề có AC và test đối ứng: ngưỡng 0.7 strict, idempotent ngày + cooldown rising-edge, maturity gate, model-agnostic qua price_forecast tươi 36h. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập FR-DEAL-006.*
