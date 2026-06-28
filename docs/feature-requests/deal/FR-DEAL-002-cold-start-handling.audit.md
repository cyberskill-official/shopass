---
fr_id: FR-DEAL-002
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập, suy lại từ file hiện tại. FR-DEAL-002 đặc tả chính sách cold-start cho sale ảo + biểu đồ: ba trạng thái trưởng thành NEW (<14 ngày) / WARMING (14-90) / MATURE (>=90), cổng `IsFeatureReady` 90 ngày trước khi công khai tính năng cho 1 SKU (đúng §5.1), và `category_prior` materialized view làm fallback chỉ gộp từ SKU MATURE. Mệnh đề khớp §3.5 (SKU <14 ngày trả UNKNOWN) và §5.1 (baseline ~90 ngày). 13 mệnh đề §1 có AC §4 và test §5. Sàn `min_sample_count` (30) chặn prior mỏng. Đạt 10/10.

## §2 - Findings (đã kiểm trong lượt này)

### ISS-001 - Ngưỡng trưởng thành nhất quán với FR-DEAL-001 (đã xác nhận)
NEW <14 ngày trả UNKNOWN khớp đúng biên 14 của FR-DEAL-001 (`len(hist)<14 -> UNKNOWN`); cổng MATURE 90 ngày khớp cửa sổ 90 ngày của sale ảo. Hằng số có tên `warmingDays=14`, `matureDays=90` (không số ma thuật). Không cần sửa.

### ISS-002 - Prior chỉ gộp SKU MATURE (đã xác nhận)
§3 migration `category_prior` có `WHERE ... now()-first_seen >= INTERVAL '90 days'` trong CTE `mature_sku`; §1 #7 + DEC-DEAL-14 loại SKU NEW/WARMING khỏi mẫu. Test `TestPrior_ExcludesImmature` xác nhận SKU NEW không kéo lệch median. Đây là chi tiết dễ sai nhất, đã đặc tả đúng.

### ISS-003 - Nguồn DDL và REFRESH CONCURRENTLY (đã xác nhận)
View tính từ `price_daily` (FR-PRICE-002) + `tracked_product` (FR-PRICE-001); `idx_category_prior_cat` unique tạo cùng migration để `REFRESH ... CONCURRENTLY` hợp lệ. Failure mode "thiếu unique index" ghi rõ ở §10. Không cần sửa.

### ISS-004 - Typography (đã xác nhận sạch)
Quét toàn file: không mũi tên unicode, em-dash, en-dash, curly quote, ellipsis hay emoji trong prose. Không cần sửa.

## §3 - Traceability §1 -> AC -> artefact (dựng từ file hiện tại)

| §1 clause | §4 AC | Test / artefact §5 / §3 |
|---|---|---|
| #1 daysOfHistory | AC #1 | `Maturity` input + §3 |
| #2 ánh xạ 3 trạng thái | AC #2 | `TestMaturity_Boundaries` (13/14/89/90) |
| #3 NEW -> UNKNOWN | AC #3 | handoff FR-DEAL-001, AC #3 |
| #4 WARMING low_confidence | AC #4 | §1 #4 + UI flag |
| #5 MATURE full | AC #5 | `TestIsFeatureReady_Gate90d` |
| #6 view category_prior | AC #6 | `0001_category_prior.sql` |
| #7 chỉ gộp MATURE | AC #7 | `TestPrior_ExcludesImmature` |
| #8 IsFeatureReady gate | AC #8 | `TestIsFeatureReady_Gate90d` |
| #9 prior fallback khi non | AC #9 | `TestPriorFallback_WhenNew` |
| #10 category_id NULL | AC #10 | §1 #10 + failure mode |
| #11 refresh theo lịch | AC #11 | scheduler REFRESH |
| #12 sàn min_sample_count | AC #12 | `PriorFor` ok=false khi <30 |
| #13 OTel metric | AC #13 | metric `deal_maturity_state_total` |

## §4 - Kết luận

Mỗi mệnh đề có AC và test đối ứng; cổng 90 ngày và loại-SKU-non-khỏi-prior tái tạo đúng §5.1 và DEC-DEAL-14. Sàn sample_count chặn prior mỏng. Prose sạch ASCII, không cần sửa. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập FR-DEAL-002.*
