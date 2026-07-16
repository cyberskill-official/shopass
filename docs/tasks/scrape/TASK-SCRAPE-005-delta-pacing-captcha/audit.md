---
fr_id: TASK-SCRAPE-005
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Re-derive từ file hiện tại. TASK-SCRAPE-005 đặc tả pacing + jitter + CAPTCHA handling + sink delta-only ở mức triển khai được: 12 mệnh đề §1 normative (đủ MUST), mỗi cái testable có AC §4 và test §5. Pacing ngẫu nhiên [min,max] + jitter + min_delay cứng per-platform (tách khỏi concurrency cap); CAPTCHA slider/puzzle/verify có ngân sách per ngày + giới hạn per target tránh leo thang. Điểm load-bearing: sink ghi qua `price.Repo.InsertSnapshot` delta-only của TASK-PRICE-002, KHÔNG tự viết `INSERT INTO price_snapshot`, không nhân bản logic delta - đã kiểm và xác nhận đúng. Score = 10/10.

## §2 - Findings

### ISS-001 - Arrow glyph trong comment code (đã sửa)
Phát hiện 1 ký tự mũi tên U+2192 trong comment Vietnamese của khối code (`sink.go::Write` dòng 161), vi phạm tiêu chí O. Đã sửa thành `->`. Quét lại: 0 mũi tên; không có em-dash/en-dash/nháy cong/ellipsis.

### ISS-002 - Cross-ref delta-only (xác nhận KHÔNG nhân bản, gọi InsertSnapshot)
Kiểm trọng tâm rubric D: §1 #6 + `sink.go::Write` gọi `s.price.InsertSnapshot(ctx, snap)`; grep file thấy chuỗi `INSERT INTO price_snapshot` chỉ xuất hiện 2 lần và đều trong prose mô tả điều sink KHÔNG được làm (dòng 68, 178), không phải câu SQL thực thi. `TestSink_DelegatesToInsertSnapshot` khóa `insertCalls==1`. Logic "chỉ ghi khi giá đổi" thuộc PRICE, không lặp ở đây. Đúng tuyệt đối.

### ISS-003 - written=false không phải lỗi + tín hiệu re-tier (xác nhận)
§1 #7 đếm `snapshot_skipped_total` khi `written=false` (AC9 `TestSink_NoChangeSkips`); §1 #8 trả `(written, flashSale)` cho TASK-SCRAPE-001 re-tier (AC10 `TestSink_ReturnsFlashForRetier`). §1 #12 idempotent retry qua `ON CONFLICT DO NOTHING` của PRICE (AC11). Khớp.

### ISS-004 - Đối chiếu §3.3/§3.9 (xác nhận)
§3.3: pacing ngẫu nhiên + jitter tránh rate-limit, puzzle slider/CAPTCHA khi nghi ngờ cần dịch vụ giải hoặc headless hành vi người. task §1 #1-#5 phủ đủ. Không khiếm khuyết.

## §3 - Traceability §1 -> AC -> artefact

| §1 mệnh đề | §4 AC | §5 test / §3 artefact |
|---|---|---|
| #1 pacing jitter | AC2 | `limiter.go::Wait` + `TestLimiter_HasJitter` |
| #2 min_delay cứng | AC1 | `TestLimiter_RespectsMinDelay` |
| #3 phối hợp concurrency cap | AC1 (độc lập với cap) | rationale + cap TASK-SCRAPE-001 |
| #4 Detect CAPTCHA | AC4,AC5 | `detect.go::Detect` + `TestDetect_Kinds` |
| #5 CAPTCHA có ngân sách | AC6 | `solver.go` + `ErrCaptchaBudget` |
| #6 ghi qua InsertSnapshot | AC8 | `sink.go::Write` + `TestSink_DelegatesToInsertSnapshot` |
| #7 written=false không lỗi | AC9 | `TestSink_NoChangeSkips` |
| #8 tín hiệu re-tier | AC10 | `Write` trả flashSale + `TestSink_ReturnsFlashForRetier` |
| #9 không block toàn pool | AC7 | lùi target |
| #10 OTel metric (SHOULD) | AC12 | counters/histogram |
| #11 không vỡ SLA hot | AC3 | `hot_min_ms` config |
| #12 idempotent retry | AC11 | ON CONFLICT (TASK-PRICE-002) |

## §4 - Kết luận

Mọi mệnh đề normative có AC + test/artefact; không mệnh đề mồ côi. Điểm nối tới delta-only của TASK-PRICE-002 đúng: sink chỉ gọi `InsertSnapshot`, không tự ghi, không nhân bản delta - khóa bằng test. Một typography defect đã sửa. Score = 10/10. Verdict: PASS.

---

*Audit độc lập TASK-SCRAPE-005 - hết.*
