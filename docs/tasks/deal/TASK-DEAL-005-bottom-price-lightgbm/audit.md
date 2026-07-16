---
fr_id: TASK-DEAL-005
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập, suy lại từ file hiện tại. TASK-DEAL-005 nâng cấp dự đoán đáy lên LightGBM cho SKU >=180 ngày: feature store thống nhất train/serve (1 dòng/(product_id, as_of_date)), nhãn `future_min_price_14d` dựng bằng forward-join CHỈ lúc train, cổng 180 ngày (dưới ngưỡng fallback Prophet TASK-DEAL-004 rồi category prior TASK-DEAL-002), giữ chung hợp đồng tín hiệu `p_bottom_14d`. Feature set khớp §3.5(2) (day_of_month, trailing_min_30/60/90, price_vs_median90, volatility_30d, category_seasonality, flash_sale_flag, platform_id; target future_min_price_14d; >=180d history). 14 mệnh đề §1 có AC §4 và test §5. Priority SHOULD hợp lý (nâng cấp, không chặn MVP). Đạt 10/10.

## §2 - Findings (đã kiểm trong lượt này)

### ISS-001 - Chống train/serve skew + rò nhãn (đã xác nhận)
§1 #6 train và serve đọc CÙNG `feature_store`/`build_features` (DEC-DEAL-42); §1 #5 serve không bao giờ gọi `future_min_price_14d` (DEC-DEAL-43). `labels.py` tách khỏi đường serve; test `test_label_no_leak_at_serve` (cửa sổ tương lai chưa đủ 14 ngày trả None) + `test_features_use_no_future`. Hai lớp lỗi nguy hiểm nhất được khép. Không cần sửa.

### ISS-002 - Không tạo bảng price_forecast thứ hai (đã xác nhận)
§1 #11 + §7 ghi rõ INSERT vào bảng `price_forecast` do TASK-DEAL-004 sở hữu với `model_kind='lgbm'` (đã nằm trong CHECK của DEAL-004), KHÔNG tạo bảng thứ hai. Ánh xạ cột `expected_min_14d`->yhat, `as_of_date`->run_date, `horizon_days`->horizon_day. Migration §3 chỉ tạo `feature_store` (0002), không đụng price_forecast. Coherent.

### ISS-003 - Cổng 180 ngày + hợp đồng tín hiệu chung (đã xác nhận)
§1 #7 chỉ chạy LightGBM khi >=180 ngày, dưới ngưỡng fallback Prophet (DEC-DEAL-40); test `test_eligibility_180d_gate` (179->prophet, 180->lgbm). §1 #10 phát cùng schema tín hiệu TASK-DEAL-004 (`p_bottom_14d`, `expected_min_14d`, `horizon_days=14`, chỉ khác model_kind) để TASK-DEAL-006 bất khả tri theo mô hình; test `test_p_bottom_contract_matches_prophet`. Không cần sửa.

### ISS-004 - Tất định + xử lý NULL feature + typography (đã xác nhận)
§1 #12 `random_state` cố định; §1 #14 điền feature thiếu theo quy ước rõ ràng thay vì NaN ngầm. Prose sạch: không mũi tên unicode, em-dash, en-dash, curly quote, ellipsis, emoji. Không cần sửa.

## §3 - Traceability §1 -> AC -> artefact (dựng từ file hiện tại)

| §1 clause | §4 AC | Test / artefact §5 / §3 |
|---|---|---|
| #1 bảng feature_store 11 cột | AC #1 | `0002_feature_store.sql` |
| #2 feature <= as_of | AC #2 | `test_features_use_no_future` |
| #3 đặc tả từng feature | AC #3 | `build_features` |
| #4 nhãn forward-join 14d | AC #4 | `test_label_future_min_14d` |
| #5 serve không tính nhãn | AC #5 | `test_label_no_leak_at_serve` |
| #6 cùng nguồn feature | AC #6 | `feature_store`/`build_features` |
| #7 cổng 180d fallback | AC #7 | `test_eligibility_180d_gate` |
| #8 train LGBMRegressor seed | AC #8 | `test_lgbm_train_predict_shape` |
| #9 suy p_bottom | AC #9 | `p_bottom` |
| #10 cùng hợp đồng tín hiệu | AC #10 | `test_p_bottom_contract_matches_prophet` |
| #11 ghi price_forecast lgbm | AC #11 | §7 ánh xạ cột |
| #12 tái lập seed | AC #12 | `random_state=SEED` |
| #13 OTel metric | AC #13 | `lgbm_predict_total` |
| #14 xử lý NULL feature | AC #13 | §1 #14 + failure mode |

## §4 - Kết luận

Mỗi mệnh đề có AC và test đối ứng; feature store khép train/serve skew, forward-join khép rò nhãn, cổng 180d và hợp đồng tín hiệu chung tái tạo đúng §3.5(2); ghi vào bảng price_forecast của DEAL-004 không nhân đôi. Prose sạch ASCII, không cần sửa. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập TASK-DEAL-005.*
