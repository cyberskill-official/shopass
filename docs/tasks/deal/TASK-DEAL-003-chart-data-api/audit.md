---
fr_id: TASK-DEAL-003
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập, suy lại từ file hiện tại. TASK-DEAL-003 đặc tả `GET /v1/products/{id}/chart` trả feed biểu đồ giá có chú giải: thân daily đọc từ continuous aggregate `price_daily` (TASK-PRICE-002), overlay median90 + trailing_min + verdict sale ảo (gọi TASK-DEAL-001, không tự tính) + mốc ngày đôi, tôn trọng cổng độ chín TASK-DEAL-002, p95 <500ms. Khớp §3.7 (shape API) và §3.8 (NFR biểu đồ <500ms). 13 mệnh đề §1 có AC §4 và test §5. Mọi giá là int64 VND trong JSON (không float/string). Phân biệt rõ với TASK-PRICE-003 (chuỗi thô không chú giải). Đạt 10/10.

## §2 - Findings (đã kiểm trong lượt này)

### ISS-001 - Nguồn dữ liệu đúng tầng (đã xác nhận)
§1 #3 + #9 đọc `price_daily` (bảng tổng hợp nhỏ) cho thân, KHÔNG quét raw `price_snapshot` cho cả khoảng - giữ p95 <500ms (DEC-DEAL-20). `disallowed_tools` chặn quét raw. Failure mode "quét raw toàn khoảng" ghi ở §10. Khớp NFR §3.8.

### ISS-002 - Verdict đồng nhất một nguồn (đã xác nhận)
§1 #5 gọi đánh giá TASK-DEAL-001 thay vì tính lại verdict ở DEAL-003 hay client; AC #5 yêu cầu khớp nhãn thẻ sản phẩm. Loại rủi ro dải verdict trên biểu đồ lệch nhãn thẻ. Cổng độ chín TASK-DEAL-002 ép NEW -> verdict UNKNOWN (§1 #6, AC #6). Không cần sửa.

### ISS-003 - Allowlist range chặn quét vô giới hạn (đã xác nhận)
§1 #2 chỉ nhận `{7d,30d,90d,180d,1y}`, default 90d, ngoài allowlist trả 400 (DEC-DEAL-24). `rangeWindows` map hiện thực đúng. Test `range=5d` (ngoài allowlist) phủ ở AC #2. Phân biệt 404 (SKU không tồn tại) vs 200+daily rỗng (SKU có thật chưa snapshot) ở §1 #10.

### ISS-004 - Kiểu tiền tệ và typography (đã xác nhận)
Mọi giá (`min_p/max_p/close_p/median90/trailing_min`) là int64 VND trong DTO JSON (§1 #11), đồng bộ DEC-PRICE-05. Prose sạch: không mũi tên unicode, em-dash, en-dash, curly quote, ellipsis, emoji. Không cần sửa.

## §3 - Traceability §1 -> AC -> artefact (dựng từ file hiện tại)

| §1 clause | §4 AC | Test / artefact §5 / §3 |
|---|---|---|
| #1 route + 400 id sai | AC #1 | `HandleChart` ParseInt |
| #2 allowlist range | AC #2 | `TestChart_Default90d` + AC #2 |
| #3 thân từ price_daily | AC #3 | `QueryDaily` + AC #3 |
| #4 median90 + trailing_min | AC #4 | `TestChart_Annotations_MedianTrailingMin` |
| #5 verdict TASK-DEAL-001 | AC #5 | `h.deal.Verdict` |
| #6 maturity TASK-DEAL-002 | AC #6 | `TestChart_MaturityFlag_Warming` |
| #7 mốc ngày đôi | AC #7 | `TestChart_DoubleDateMarkers` |
| #8 hình dạng JSON | AC #8 | `ChartResponse` struct |
| #9 p95 <500ms | AC #9 | metric `chart_duration_ms` |
| #10 404 vs 200-rỗng | AC #10 | `TestChart_UnknownProduct_404` |
| #11 int64 VND | AC #11 | DTO json tags int64 |
| #12 JWT gateway | AC #12 | router sau middleware |
| #13 OTel metric | AC #13 | `chart_requests_total` |

## §4 - Kết luận

Mỗi mệnh đề có AC và test đối ứng; đọc price_daily giữ p95, một nguồn verdict đồng nhất, allowlist chặn quét raw, giá int64 VND end-to-end. Prose sạch ASCII, không cần sửa. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập TASK-DEAL-003.*
