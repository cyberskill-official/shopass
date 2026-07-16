---
fr_id: TASK-DEAL-004
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập, suy lại từ file hiện tại. TASK-DEAL-004 đặc tả baseline Prophet dự đoán đáy giá: đọc `price_daily` (ds=day, y=close_p), dựng regressor nhịp khuyến mãi VN (double-date, payday window, flash), fit Prophet yearly+monthly, dự báo 14 ngày, suy `p_bottom_14d` từ khoảng tin cậy, fallback category prior khi cold-start, lưu `price_forecast`. Feature engineering khớp §3.5(2): double-date d==m, payday, baseline Prophet -> nâng cấp LightGBM, cold-start category priors, serving "alert nếu P(bottom within 14d) > 0.7". Là task sở hữu DDL bảng `price_forecast`. 13 mệnh đề §1 có AC §4 và test §5. Đạt 10/10.

## §2 - Findings (đã kiểm trong lượt này)

### ISS-001 - Coherence bảng price_forecast chéo DEAL-004/005/006 (đã xác nhận)
TASK-DEAL-004 sở hữu `price_forecast` ở §3 migration. Cột `scored_at TIMESTAMPTZ NOT NULL DEFAULT now()` hiện diện (để TASK-DEAL-006 lọc tươi `scored_at >= now()-36h`). CHECK `model_kind IN ('prophet','category_prior','lgbm')` đã bao 'lgbm' để TASK-DEAL-005 ghi cùng bảng. Ánh xạ cột tài liệu hóa ở §7 (`expected_min_14d`->yhat, `as_of_date`->run_date, `horizon_days`->horizon_day). Coherent ba chiều, không cần sửa.

### ISS-002 - Fidelity feature §3.5(2) (đã xác nhận)
`is_double_date` = `d.day == d.month` (1.1, 2.2, ... 12.12); `is_payday_window` quanh ngày 1 và 15; flash_sale=0 cho ngày tương lai chưa biết. Khớp danh sách feature §3.5(2). Test `test_is_double_date` (2.2/12.12 true, 3.4 false), `test_is_payday_window`. Không cần sửa.

### ISS-003 - Tách train/serve + tất định (đã xác nhận)
§1 #10 lưu price_forecast để batch/alert đọc, không tính lúc alert (DEC-DEAL-34). §1 #11 cố định `mcmc_samples=0`, `uncertainty_samples`, seed numpy cho test lặp lại. `p_bottom_14d` đơn điệu theo khoảng tin cậy (§1 #9), test `test_p_bottom_monotonic_with_interval`. yhat làm tròn BIGINT VND (§1 #12). Không cần sửa.

### ISS-004 - Typography (đã xác nhận sạch)
Quét toàn file: không mũi tên unicode, em-dash, en-dash, curly quote, ellipsis, emoji trong prose. Không cần sửa.

## §3 - Traceability §1 -> AC -> artefact (dựng từ file hiện tại)

| §1 clause | §4 AC | Test / artefact §5 / §3 |
|---|---|---|
| #1 đọc price_daily ds/y | AC #1 | `forecast_bottom` input |
| #2 khung regressor 3 cờ | AC #2 | `build_regressor_frame` |
| #3 is_double_date / days_to | AC #3 | `test_is_double_date`, `test_days_to_next_double_date` |
| #4 is_payday_window | AC #4 | `test_is_payday_window` |
| #5 cấu hình Prophet | AC #5 | `build_baseline` |
| #6 MATURE fit Prophet | AC #6 | `test_prophet_forecast_shape` |
| #7 cold-start category prior | AC #7 | `test_cold_start_uses_prior` |
| #8 suy p_bottom_14d | AC #8 | `_p_bottom_14d` |
| #9 đơn điệu theo CI | AC #9 | `test_p_bottom_monotonic_with_interval` |
| #10 lưu 14 dòng price_forecast | AC #10 | `0001_price_forecast.sql` |
| #11 tất định seed | AC #11 | `mcmc_samples=0` config |
| #12 yhat BIGINT VND | AC #12 | CHECK migration |
| #13 OTel metric | AC #13 | `ml_forecast_run_total` |

## §4 - Kết luận

Mỗi mệnh đề có AC và test đối ứng; feature double-date/payday tái tạo đúng §3.5(2); bảng price_forecast coherent với DEAL-005/006 (scored_at + 'lgbm'); tách train/serve và tất định seed đạt. Prose sạch ASCII, không cần sửa. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập TASK-DEAL-004.*
