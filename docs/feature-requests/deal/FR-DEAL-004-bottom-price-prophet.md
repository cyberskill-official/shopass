---
id: FR-DEAL-004
title: "Dự đoán đáy giá - Prophet baseline với regressor double-date/payday/flash, đọc price_daily, fallback category prior khi cold-start, suy P(đáy trong 14 ngày) và lưu price_forecast cho batch/alert đọc"
module: DEAL
priority: MUST
status: done
verify: T
phase: P2
milestone: P2 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-PRICE-002, FR-DEAL-002, FR-DEAL-005, FR-DEAL-006]
depends_on: [FR-PRICE-002, FR-DEAL-002]
blocks: [FR-DEAL-005, FR-DEAL-006]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.5 (dự đoán đáy giá AI, Prophet baseline, double-date/payday regressors)"
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §5.1 (cold-start category priors)"
source_decisions:
  - "DEC-DEAL-30: Prophet là baseline dự đoán đáy giá - seasonality yearly + monthly cộng regressor ngoài (double-date, payday window, flash sale); LightGBM nâng cấp ở FR-DEAL-005"
  - "DEC-DEAL-31: đọc continuous aggregate price_daily (FR-PRICE-002) làm chuỗi đầu vào (cột ds=day, y=close_p), không đọc raw price_snapshot"
  - "DEC-DEAL-32: cold-start dùng category_prior (FR-DEAL-002) khi SKU chưa qua cổng MATURE - không fit Prophet trên chuỗi quá ngắn, trả forecast phẳng quanh prior"
  - "DEC-DEAL-33: suy P(đáy trong 14 ngày) từ dự báo + khoảng tin cậy yhat_lower - so cận dưới 14 ngày tới với đáy quan sát gần đây"
  - "DEC-DEAL-34: lưu forecast vào bảng price_forecast (1 dòng/horizon_day) để batch chấm điểm và alert (FR-DEAL-006) đọc, tách huấn luyện khỏi phục vụ"

language: "Python 3.12 (Prophet, pandas, numpy) - ml-svc; forecast persisted to Postgres for deal-svc"
service: shopass/services/ml/
new_files:
  - services/ml/bottom/regressors.py
  - services/ml/bottom/features.py
  - services/ml/bottom/prophet_baseline.py
  - services/ml/migrations/0001_price_forecast.sql
  - services/ml/tests/test_regressors.py
  - services/ml/tests/test_prophet_baseline.py
modified_files:
  - services/ml/bottom/__init__.py            # export build_baseline, forecast_bottom
allowed_tools:
  - file_read: services/ml/**
  - file_write: services/ml/**
  - bash: cd services/ml && python -m pytest tests/
disallowed_tools:
  - đọc raw price_snapshot thay vì continuous aggregate price_daily (vi phạm DEC-DEAL-31, đốt I/O và lệch độ phân giải)
  - fit Prophet trên SKU chưa MATURE thay vì fallback category prior (vi phạm DEC-DEAL-32, overfit chuỗi mỏng)
  - chạy không cố định seed/uncertainty_samples (phá tính tất định của test, §1 #11)
  - tính trực tiếp lúc alert thay vì đọc price_forecast đã lưu (vi phạm DEC-DEAL-34, ghép cứng train vào serve)

effort_hours: 10
sub_tasks:
  - "1.0h: regressors.py - is_double_date(d==m), days_to_next_double_date, is_payday_window quanh ngày 1 và 15"
  - "1.5h: features.py - dựng khung regressor cho mọi ngày của lịch sử + 14 ngày tương lai (double_date, payday, flash_sale)"
  - "2.0h: prophet_baseline.py - build_baseline(): Prophet yearly+monthly + add_regressor 3 cờ, fit, predict 14 ngày"
  - "1.0h: prophet_baseline.py - forecast_bottom(): suy p_bottom_14d từ yhat_lower và đáy gần đây (DEC-DEAL-33)"
  - "0.5h: prophet_baseline.py - nhánh cold-start gọi category_prior, trả forecast phẳng quanh prior (DEC-DEAL-32)"
  - "0.5h: 0001_price_forecast.sql - bảng price_forecast (product_id, run_date, horizon_day, yhat, yhat_lower/upper, p_bottom_14d, model_kind)"
  - "1.5h: test_regressors.py - is_double_date (2.2/12.12 true, 3.4 false), days_to_next_double_date, is_payday_window"
  - "1.5h: test_prophet_baseline.py - shape 14 dòng, cold-start dùng prior, p_bottom đơn điệu theo khoảng tin cậy"
  - "0.5h: OTel metric ml_forecast_run_total{model_kind} + ml_forecast_p_bottom (histogram)"

risk_if_skipped: "Dự đoán đáy giá là tính năng phân biệt giai đoạn 2 của SănDeal: thay vì chỉ nói giá hiện tại đắt hay rẻ so với quá khứ (FR-DEAL-001), nó nói thẳng cho người dùng có nên chờ hay không - 'khả năng cao còn giảm trước 12.12, đừng mua vội'. Không có baseline Prophet thì không có nguồn dự báo nào để FR-DEAL-006 bắn cảnh báo đáy, và FR-DEAL-005 (LightGBM) cũng mất mốc so sánh để chứng minh nó tốt hơn. Bỏ regressor double-date/payday thì mô hình mù với chính nhịp khuyến mãi đặc thù VN - nơi các đáy giá thật sự rơi vào (1.1, 2.2, ... 12.12, kỳ lương). Bỏ nhánh cold-start (category prior) thì SKU mới hoặc Prophet fit trên chuỗi vài tuần cho dự báo nhiễu, hoặc service vỡ vì không đủ điểm. Không lưu price_forecast mà tính trực tiếp lúc alert thì batch đêm và đường phục vụ bị ghép cứng vào huấn luyện, không scale tới hàng triệu SKU."
---

## §1 - Mô tả (BCP-14 normative)

Service ML **MUST** cung cấp một bộ dự đoán đáy giá baseline dựa trên Facebook/Meta Prophet: đọc chuỗi giá ngày từ `price_daily` (FR-PRICE-002), dựng khung regressor cho nhịp khuyến mãi VN, fit Prophet với mùa vụ năm + tháng, dự báo 14 ngày tới, suy ra xác suất chạm đáy trong 14 ngày, và lưu kết quả vào bảng `price_forecast` để batch đêm và alert (FR-DEAL-006) đọc. Đây là BASELINE; FR-DEAL-005 nâng cấp lên LightGBM cho SKU có `>= 180` ngày lịch sử. Hợp đồng:

1. **MUST** đọc chuỗi đầu vào từ continuous aggregate `price_daily` (DEC-DEAL-31): map `ds = day`, `y = close_p`, sắp theo `ds` tăng dần - KHÔNG đọc raw `price_snapshot` (sai độ phân giải và đốt I/O).
2. **MUST** dựng khung regressor cho từng ngày của lịch sử VÀ 14 ngày tương lai gồm ba cờ: `is_double_date` (ngày mà `day == month`, ví dụ 1.1, 2.2, ... 12.12), `is_payday_window` (cửa sổ quanh ngày 1 và ngày 15 hằng tháng), `flash_sale` (cờ flash sale của ngày đó nếu có, mặc định 0 cho ngày tương lai chưa biết).
3. **MUST** cung cấp hàm thuần `is_double_date(date) -> bool` trả `True` đúng khi `date.day == date.month`, và `days_to_next_double_date(date) -> int` trả số ngày tới ngày double-date kế tiếp (0 khi chính ngày đó là double-date).
4. **MUST** cung cấp `is_payday_window(date) -> bool` trả `True` khi ngày nằm trong cửa sổ lương: quanh ngày 1 (ngày cuối tháng trước tới ngày 2) và quanh ngày 15 (ngày 14 tới 16), là khoảng sức mua tăng.
5. **MUST** cấu hình Prophet với `seasonality_mode='multiplicative'`, bật `yearly_seasonality=True`, thêm mùa vụ tháng (`add_seasonality(name='monthly', period=30.5, fourier_order=5)`), và `add_regressor` cho cả ba cờ ở §1 #2 (DEC-DEAL-30).
6. **MUST** khi SKU đã qua cổng MATURE (`>= 90` ngày, theo FR-DEAL-002): fit Prophet trên toàn bộ chuỗi `price_daily` và predict 14 ngày tới, sinh `yhat`, `yhat_lower`, `yhat_upper` cho mỗi `horizon_day` từ 1 tới 14.
7. **MUST** khi SKU CHƯA qua cổng MATURE (cold-start, DEC-DEAL-32): KHÔNG fit Prophet trên chuỗi mỏng. Thay vào đó gọi `category_prior` (FR-DEAL-002 `PriorFor`) và trả forecast phẳng quanh `prior.median_price` với khoảng tin cậy nới rộng; `model_kind = 'category_prior'`.
8. **MUST** suy `p_bottom_14d` (xác suất giá chạm đáy trong 14 ngày tới) từ dự báo và khoảng tin cậy (DEC-DEAL-33): so cận dưới dự báo `min(yhat_lower[1..14])` với đáy quan sát gần đây (`trailing_min` 30/60/90 ngày); xác suất cao khi cận dưới tương lai xuống sát hoặc dưới đáy gần đây.
9. **MUST** đảm bảo `p_bottom_14d` đơn điệu không giảm theo độ rộng khoảng tin cậy ở phía dưới: khoảng tin cậy càng nới xuống thấp (cận dưới càng thấp so với đáy gần đây) thì `p_bottom_14d` càng lớn, kẹp trong `[0.0, 1.0]`.
10. **MUST** lưu forecast vào bảng `price_forecast` (DEC-DEAL-34): mỗi lần chạy ghi 14 dòng (1 dòng/`horizon_day`) cho `(product_id, run_date)` kèm `yhat`, `yhat_lower`, `yhat_upper`, `p_bottom_14d` (lặp cùng giá trị trên 14 dòng), `model_kind` (`'prophet'` hoặc `'category_prior'`). Batch đêm và FR-DEAL-006 chỉ đọc bảng này, không gọi lại huấn luyện.
11. **MUST** tất định với seed và `uncertainty_samples` cố định: cùng chuỗi đầu vào, cùng `run_date`, cùng cấu hình cho cùng `yhat`/`yhat_lower`/`p_bottom_14d` (đặt `mcmc_samples=0`, cố định `uncertainty_samples`, seed numpy) để test lặp lại được.
12. **MUST** trả giá ở đơn vị VND số nguyên khi ghi xuống `price_forecast` (làm tròn `yhat` về `BIGINT`), đồng bộ DEC-PRICE-05 của FR-PRICE-002 (không lưu giá dạng thập phân float).
13. **SHOULD** phát OTel metric: `ml_forecast_run_total{model_kind}` (counter), `ml_forecast_p_bottom` (histogram phân bố `p_bottom_14d`), `ml_forecast_duration_ms` (histogram thời gian fit + predict).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao Prophet làm baseline (DEC-DEAL-30)?** Prophet xử lý mùa vụ và sự kiện kiểu ngày lễ với rất ít tinh chỉnh: khai báo yearly + monthly seasonality và thêm vài regressor nhị phân là đủ bắt được chu kỳ. Nó chịu được lỗ hổng dữ liệu, cho sẵn khoảng tin cậy `yhat_lower/upper` mà ta cần để suy xác suất đáy, và là một mốc trung thực để FR-DEAL-005 (LightGBM) phải vượt qua. Bắt đầu bằng baseline đơn giản, có khoảng tin cậy, dễ giải thích là lựa chọn đúng cho giai đoạn này.

**Vì sao regressor double-date và payday (§1 #2)?** Lịch khuyến mãi VN không trải đều: đáy giá thật sự dồn vào các ngày double-date (1.1, 2.2, ... 12.12 nơi ngày trùng tháng) và quanh kỳ lương (đầu tháng, giữa tháng) khi sức mua bật lên. Một mô hình mùa vụ thuần năm/tháng sẽ bỏ lỡ các đỉnh sự kiện rời rạc này. Đưa chúng vào dưới dạng regressor cho Prophet biết "ngày 12.12 sắp tới thường là điểm giá thấp", đúng tín hiệu người dùng cần để được khuyên chờ.

**Vì sao category prior cho cold-start (DEC-DEAL-32)?** Một SKU mới hoặc chuỗi vài tuần không đủ để Prophet ước lượng mùa vụ năm - fit lên nó cho dự báo nhiễu và khoảng tin cậy vô nghĩa. FR-DEAL-002 đã có sẵn `category_prior` (chỉ gộp từ SKU MATURE). Khi SKU chưa qua cổng MATURE, ta mượn median của category làm dự báo phẳng tạm, nới khoảng tin cậy để phản ánh sự thiếu chắc chắn, thay vì giả vờ có một mô hình tốt. Đây cũng là cùng prior mà FR-DEAL-002 dùng cho fallback verdict.

**Vì sao suy P(đáy trong 14 ngày) từ khoảng tin cậy (DEC-DEAL-33)?** Người dùng không cần một con số `yhat` chính xác, họ cần một câu trả lời "nên chờ không". Cận dưới dự báo `yhat_lower` cho ta kịch bản giá thấp hợp lý trong 14 ngày tới. So nó với đáy quan sát gần đây cho biết khả năng giá phá đáy. Đổi dự báo điểm thành một xác suất hành động được là cách biến mô hình thành lời khuyên cụ thể, đồng thời cho FR-DEAL-006 một ngưỡng (`> 0,7`) để quyết định bắn cảnh báo.

**Vì sao lưu price_forecast thay vì tính lúc alert (DEC-DEAL-34)?** Fit Prophet cho hàng triệu SKU là việc nặng, chạy theo mẻ ban đêm. Alert và batch chấm điểm phải nhẹ và nhanh. Tách huấn luyện (ghi bảng) khỏi phục vụ (đọc bảng) cho phép đường alert chỉ là một truy vấn `price_forecast`, không kéo theo Prophet. Đây là cùng nguyên tắc decouple mà `price_daily` áp cho biểu đồ: tính sẵn, đọc nhanh.

---

## §3 - Hợp đồng API / DDL

### Regressors (Python)

```python
# services/ml/bottom/regressors.py
from datetime import date, timedelta


def is_double_date(d: date) -> bool:
    """True đúng khi ngày trùng tháng (1.1, 2.2, ... 12.12)."""
    return d.day == d.month


def days_to_next_double_date(d: date) -> int:
    """Số ngày tới ngày double-date kế tiếp; 0 nếu chính ngày đó là double-date."""
    probe = d
    for _ in range(366):
        if is_double_date(probe):
            return (probe - d).days
        probe += timedelta(days=1)
    return 0  # không xảy ra (luôn có double-date trong 1 năm)


def is_payday_window(d: date) -> bool:
    """True trong cửa sổ lương: quanh ngày 1 (cuối tháng trước -> ngày 2) và ngày 15 (14 -> 16)."""
    if d.day in (1, 2) or d.day in (14, 15, 16):
        return True
    nxt = d + timedelta(days=1)
    return nxt.day == 1  # ngày cuối tháng (liền trước mùng 1)
```

### Features (Python)

```python
# services/ml/bottom/features.py
import pandas as pd
from .regressors import is_double_date, is_payday_window

HORIZON_DAYS = 14


def build_regressor_frame(history: pd.DataFrame, horizon: int = HORIZON_DAYS) -> pd.DataFrame:
    """history: cột ds (datetime ngày), y (close_p), flash_sale (0/1).
    Trả khung ds + 3 cờ regressor, nối thêm `horizon` ngày tương lai (flash_sale=0).
    """
    future_ds = pd.date_range(
        history["ds"].max() + pd.Timedelta(days=1), periods=horizon, freq="D"
    )
    future = pd.DataFrame({"ds": future_ds, "flash_sale": 0})
    frame = pd.concat([history[["ds", "flash_sale"]], future], ignore_index=True)
    d = frame["ds"].dt.date
    frame["is_double_date"] = d.map(is_double_date).astype(int)
    frame["is_payday_window"] = d.map(is_payday_window).astype(int)
    return frame
```

### Prophet baseline (Python)

```python
# services/ml/bottom/prophet_baseline.py
import numpy as np
import pandas as pd
from prophet import Prophet
from .features import build_regressor_frame, HORIZON_DAYS

REGRESSORS = ("is_double_date", "is_payday_window", "flash_sale")


def build_baseline(seed: int = 42) -> Prophet:
    """Prophet baseline: mùa vụ năm + tháng, regressor double-date/payday/flash (DEC-DEAL-30)."""
    np.random.seed(seed)
    m = Prophet(
        seasonality_mode="multiplicative",
        yearly_seasonality=True,
        weekly_seasonality=False,
        daily_seasonality=False,
        mcmc_samples=0,
        uncertainty_samples=1000,
    )
    m.add_seasonality(name="monthly", period=30.5, fourier_order=5)
    for r in REGRESSORS:
        m.add_regressor(r)
    return m


def forecast_bottom(history: pd.DataFrame, prior_median: int | None = None,
                    seed: int = 42) -> pd.DataFrame:
    """Trả DataFrame 14 dòng: horizon_day, yhat, yhat_lower, yhat_upper, p_bottom_14d, model_kind.
    Nếu history chưa MATURE (prior_median được truyền) -> forecast phẳng quanh prior (DEC-DEAL-32)."""
    if prior_median is not None:
        return _cold_start_forecast(prior_median)

    frame = build_regressor_frame(history)
    m = build_baseline(seed)
    train = history.merge(frame, on=["ds", "flash_sale"], how="left")
    m.fit(train[["ds", "y", *REGRESSORS]])
    fcst = m.predict(frame).tail(HORIZON_DAYS).reset_index(drop=True)

    trailing_min = int(history["y"].tail(90).min())
    p_bottom = _p_bottom_14d(fcst["yhat_lower"].to_numpy(), trailing_min)
    out = pd.DataFrame({
        "horizon_day": range(1, HORIZON_DAYS + 1),
        "yhat": fcst["yhat"].round().astype("int64"),
        "yhat_lower": fcst["yhat_lower"].round().astype("int64"),
        "yhat_upper": fcst["yhat_upper"].round().astype("int64"),
        "p_bottom_14d": p_bottom,
        "model_kind": "prophet",
    })
    return out


def _p_bottom_14d(yhat_lower: np.ndarray, trailing_min: int) -> float:
    """Đơn điệu theo khoảng tin cậy phía dưới (DEC-DEAL-33): cận dưới 14 ngày càng thấp
    so với đáy gần đây thì xác suất phá đáy càng cao. Kẹp [0, 1]."""
    floor14 = float(yhat_lower.min())
    if trailing_min <= 0:
        return 0.0
    drop = (trailing_min - floor14) / trailing_min  # >0 khi cận dưới xuống dưới đáy cũ
    return float(min(1.0, max(0.0, 0.5 + drop)))


def _cold_start_forecast(prior_median: int) -> pd.DataFrame:
    band = int(round(prior_median * 0.12))  # khoảng tin cậy nới rộng cho cold-start
    return pd.DataFrame({
        "horizon_day": range(1, HORIZON_DAYS + 1),
        "yhat": prior_median,
        "yhat_lower": prior_median - band,
        "yhat_upper": prior_median + band,
        "p_bottom_14d": 0.0,        # chưa đủ cơ sở -> không hứa đáy
        "model_kind": "category_prior",
    })
```

### Migration (SQL)

```sql
-- services/ml/migrations/0001_price_forecast.sql
CREATE TABLE price_forecast (
  product_id    BIGINT      NOT NULL REFERENCES tracked_product(id),
  run_date      DATE        NOT NULL,                  -- ngày chạy batch dự báo (Asia/Ho_Chi_Minh)
  horizon_day   SMALLINT    NOT NULL CHECK (horizon_day BETWEEN 1 AND 14),
  yhat          BIGINT      NOT NULL,            -- VND, làm tròn (DEC-PRICE-05); cũng là expected_min_14d mà FR-DEAL-005 tham chiếu
  yhat_lower    BIGINT      NOT NULL,
  yhat_upper    BIGINT      NOT NULL,
  p_bottom_14d  REAL        NOT NULL CHECK (p_bottom_14d BETWEEN 0 AND 1),
  model_kind    TEXT        NOT NULL CHECK (model_kind IN ('prophet', 'category_prior', 'lgbm')), -- 'lgbm' do FR-DEAL-005 ghi cùng bảng
  scored_at     TIMESTAMPTZ NOT NULL DEFAULT now(),    -- mốc sinh dòng dự báo; FR-DEAL-006 lọc tươi (scored_at >= now()-36h)
  PRIMARY KEY (product_id, run_date, horizon_day)
);

CREATE INDEX idx_forecast_latest ON price_forecast (product_id, run_date DESC);
```

---

## §4 - Acceptance criteria

1. Bộ dự báo đọc chuỗi từ `price_daily` map `ds=day, y=close_p` (kiểm qua frame đầu vào của `forecast_bottom`), không chạm `price_snapshot`.
2. `build_regressor_frame` trả khung phủ toàn lịch sử + 14 ngày tương lai, có cột `is_double_date`, `is_payday_window`, `flash_sale` (0/1).
3. `is_double_date(d)` trả `True` đúng khi `d.day == d.month`; `days_to_next_double_date(d)` trả `0` khi d là double-date, dương ngược lại.
4. `is_payday_window(d)` trả `True` trong cửa sổ quanh ngày 1 và ngày 15 (gồm ngày cuối tháng), `False` ngoài đó.
5. Prophet cấu hình `seasonality_mode='multiplicative'`, `yearly_seasonality=True`, có `monthly` seasonality và `add_regressor` cho cả ba cờ.
6. SKU MATURE -> fit Prophet, `forecast_bottom` trả đúng 14 dòng `horizon_day` 1..14 với `yhat`/`yhat_lower`/`yhat_upper`, `model_kind='prophet'`.
7. SKU cold-start (truyền `prior_median`) -> KHÔNG fit Prophet, trả forecast phẳng quanh prior, `model_kind='category_prior'`.
8. `p_bottom_14d` suy từ `min(yhat_lower)` và `trailing_min` gần đây, nằm trong `[0,1]`.
9. `p_bottom_14d` đơn điệu không giảm khi cận dưới dự báo hạ thấp so với đáy gần đây (test khoảng tin cậy rộng hơn -> xác suất không nhỏ hơn).
10. Mỗi lần chạy ghi 14 dòng vào `price_forecast` cho `(product_id, run_date)`; alert đọc bảng này, không gọi lại fit.
11. Cùng chuỗi + seed + `run_date` cho cùng `yhat`/`p_bottom_14d` (tất định, `mcmc_samples=0`).
12. `yhat` ghi xuống `price_forecast` là `BIGINT` VND (đã làm tròn), thỏa CHECK của migration.
13. Metric `ml_forecast_run_total{model_kind}` tăng theo nhãn; `ml_forecast_p_bottom` ghi nhận phân bố xác suất.

---

## §5 - Kiểm thử (verification)

```python
# services/ml/tests/test_regressors.py
from datetime import date
from bottom.regressors import is_double_date, days_to_next_double_date, is_payday_window


def test_is_double_date():
    assert is_double_date(date(2026, 2, 2))    # 2.2
    assert is_double_date(date(2026, 12, 12))  # 12.12
    assert not is_double_date(date(2026, 4, 3))  # 3.4 không trùng


def test_days_to_next_double_date():
    assert days_to_next_double_date(date(2026, 12, 12)) == 0      # chính ngày
    assert days_to_next_double_date(date(2026, 12, 1)) == 11      # tới 12.12
    assert days_to_next_double_date(date(2026, 12, 13)) > 0       # vòng sang 1.1 năm sau


def test_is_payday_window():
    assert is_payday_window(date(2026, 6, 1))    # đầu tháng
    assert is_payday_window(date(2026, 6, 15))   # giữa tháng
    assert is_payday_window(date(2026, 6, 30))   # cuối tháng (liền mùng 1)
    assert not is_payday_window(date(2026, 6, 8))  # ngoài cửa sổ
```

```python
# services/ml/tests/test_prophet_baseline.py
import numpy as np
import pandas as pd
from bottom.prophet_baseline import forecast_bottom, _p_bottom_14d, HORIZON_DAYS


def _series(days: int, base: int = 100_000) -> pd.DataFrame:
    ds = pd.date_range("2025-01-01", periods=days, freq="D")
    y = base + (np.sin(np.arange(days) / 7.0) * 5_000)  # nhịp nhẹ quanh base
    return pd.DataFrame({"ds": ds, "y": y, "flash_sale": 0})


def test_prophet_forecast_shape():
    out = forecast_bottom(_series(180))  # MATURE -> fit Prophet
    assert len(out) == HORIZON_DAYS
    assert list(out["horizon_day"]) == list(range(1, 15))
    assert (out["model_kind"] == "prophet").all()
    assert out["yhat"].dtype == np.int64


def test_cold_start_uses_prior():
    out = forecast_bottom(_series(10), prior_median=88_000)  # cold-start
    assert (out["model_kind"] == "category_prior").all()
    assert (out["yhat"] == 88_000).all()
    assert (out["p_bottom_14d"] == 0.0).all()  # không hứa đáy khi thiếu dữ liệu


def test_p_bottom_monotonic_with_interval():
    trailing_min = 80_000
    narrow = _p_bottom_14d(np.full(14, 79_000.0), trailing_min)  # cận dưới sát đáy
    wide = _p_bottom_14d(np.full(14, 70_000.0), trailing_min)    # cận dưới phá đáy sâu hơn
    assert 0.0 <= narrow <= wide <= 1.0  # khoảng nới xuống thấp -> xác suất không giảm
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `regressors.py` (3 hàm lịch thuần) -> `features.py` (`build_regressor_frame`) -> `prophet_baseline.py` (`build_baseline` + `forecast_bottom` + `_p_bottom_14d` + nhánh cold-start) -> migration `0001_price_forecast.sql` -> tests. `forecast_bottom` nhận chuỗi `price_daily` đã nạp sẵn (ds/y/flash_sale) từ deal-svc hoặc job; tầng gọi quyết định MATURE hay cold-start qua cổng FR-DEAL-002 (`IsFeatureReady`) rồi truyền `prior_median` khi cần. Batch đêm: với mỗi SKU, nạp `price_daily`, gọi `forecast_bottom`, ghi 14 dòng vào `price_forecast` (UPSERT theo PK `(product_id, run_date, horizon_day)`). FR-DEAL-006 chỉ đọc `price_forecast` mới nhất. Gói `regressors`/`features` là hàm thuần trên ngày/pandas - test nhanh, không cần Prophet; chỉ `test_prophet_forecast_shape` mới chạm Prophet.

---

## §7 - Phụ thuộc

- **FR-PRICE-002** - `price_daily` (continuous aggregate: `day`, `close_p`) là chuỗi đầu vào (ds/y). `yhat` ghi BIGINT VND đồng bộ DEC-PRICE-05.
- **FR-DEAL-002** - cổng `IsFeatureReady` quyết định MATURE vs cold-start; `PriorFor(category_id)` cấp `prior_median` cho nhánh cold-start.
- **FR-DEAL-005 (downstream, blocks)** - LightGBM nâng cấp cho SKU `>= 180` ngày; ghi cùng bảng `price_forecast` với `model_kind = 'lgbm'` (đã có trong CHECK). FR-DEAL-005 gọi `expected_min_14d` chính là cột `yhat` và `as_of_date` chính là `run_date` của bảng này; `horizon_days` của FR-DEAL-005 ứng với cột `horizon_day` (1..14) ở đây - không phải cột riêng.
- **FR-DEAL-006 (downstream, blocks)** - batch chấm điểm đêm đọc `price_forecast`, lọc dòng tươi qua cột `scored_at` (`scored_at >= now() - INTERVAL '36 hours'`) rồi bắn cảnh báo khi `p_bottom_14d > 0,7`. Bảng này là hợp đồng đọc-chỉ của FR-DEAL-006.
- Thư viện: Prophet, pandas, numpy (ml-svc, Python 3.12); driver Postgres để ghi `price_forecast`.

---

## §8 - Payload ví dụ

### Một dòng price_forecast (yhat tụt dần về 12.12, p_bottom ~0,78)

```json
{
  "product_id": 90112,
  "run_date": "2026-12-01",
  "horizon_day": 11,
  "yhat": 92000,
  "yhat_lower": 78000,
  "yhat_upper": 106000,
  "p_bottom_14d": 0.78,
  "model_kind": "prophet"
}
```

Chạy ngày 1.12, horizon_day 11 rơi vào 12.12 (double-date): `yhat` hạ về `92000` (thấp hơn mặt bằng), `yhat_lower = 78000` xuống dưới đáy 90 ngày gần đây nên `p_bottom_14d = 0,78` (vượt ngưỡng 0,7 của FR-DEAL-006 -> khuyên chờ tới 12.12).

### Truy vấn alert đọc forecast mới nhất (FR-DEAL-006)

```sql
SELECT product_id, p_bottom_14d, model_kind
FROM price_forecast
WHERE product_id = 90112
  AND run_date = (SELECT max(run_date) FROM price_forecast WHERE product_id = 90112)
  AND horizon_day = 1;   -- p_bottom_14d lặp trên 14 dòng, đọc 1 dòng là đủ
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Đưa `days_to_next_double_date` thành regressor liên tục (ngoài cờ nhị phân) - thử ở FR-DEAL-005 nơi LightGBM nhận đặc trưng số.
- Hiệu chỉnh `p_bottom_14d` thành xác suất có hiệu chuẩn (calibrated) bằng backtest - tách sang FR-DEAL-005/006 khi có nhãn đáy thực.
- Hệ số nới khoảng tin cậy cold-start (0,12) có nên co giãn theo độ phân tán giá của category - tinh chỉnh khi có dữ liệu thực.
- Lịch nghỉ lễ VN (Tết) như một regressor riêng ngoài double-date/payday - cân nhắc thêm `holidays` của Prophet ở slice sau.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Thiếu lịch sử (SKU chưa MATURE) | cổng IsFeatureReady (FR-DEAL-002) | Prophet fit chuỗi mỏng -> nhiễu | Nhánh cold-start trả prior phẳng (DEC-DEAL-32) |
| Prophet overfit chuỗi thưa/ngắn | tuổi lịch sử + test cold-start | dự báo và khoảng tin cậy vô nghĩa | Chỉ fit khi MATURE; dưới ngưỡng dùng prior |
| Rò rỉ regressor double-date (lệch ngày tương lai) | test is_double_date theo ds | mô hình thấy sai ngày sự kiện | Dựng frame future từ ds, map cờ theo ngày thực |
| flash_sale tương lai chưa biết bị đoán bừa | review features | dự báo dựa cờ giả định | Đặt flash_sale=0 cho ngày tương lai (§1 #2) |
| Seed/uncertainty không cố định | test tất định | yhat/p_bottom đổi giữa các lần chạy | mcmc_samples=0, uncertainty_samples cố định, seed numpy |
| Đọc nhầm raw price_snapshot | review nguồn nạp | sai độ phân giải, đốt I/O | Bắt buộc đọc price_daily (DEC-DEAL-31) |
| p_bottom vượt [0,1] | CHECK price_forecast + kẹp | alert sai ngưỡng | Kẹp trong _p_bottom_14d và CHECK constraint |
| Ghi forecast cũ đè/thiếu run_date | PK (product_id, run_date, horizon_day) | alert đọc dữ liệu lỗi thời | UPSERT theo PK; alert đọc max(run_date) |
| category NULL/prior mỏng ở cold-start | PriorFor trả ok=false (FR-DEAL-002) | không có prior để dựa | Bỏ qua SKU lần này, chờ đủ lịch sử |
| yhat lưu dạng float thập phân | review ghi DB | sai số tiền tệ | Làm tròn về BIGINT VND (DEC-PRICE-05) |

---

## §11 - Ghi chú

- Đây là BASELINE Prophet: mục tiêu là một mốc trung thực, có khoảng tin cậy, dễ giải thích để FR-DEAL-005 (LightGBM) phải vượt qua, không phải mô hình cuối.
- Regressor double-date/payday là phần "đặc thù VN" của baseline: đáy giá thật dồn vào 1.1, 2.2, ... 12.12 và quanh kỳ lương, nên mô hình phải nhìn thấy chúng.
- Nhánh cold-start mượn cùng `category_prior` của FR-DEAL-002 - một nguồn prior duy nhất phục vụ cả fallback verdict lẫn dự báo đáy, tránh hai định nghĩa lệch nhau.
- Tách huấn luyện (ghi `price_forecast`) khỏi phục vụ (đọc bảng) là điều kiện để scale tới hàng triệu SKU: batch đêm nặng, alert nhẹ.
- `p_bottom_14d` cố ý suy từ khoảng tin cậy `yhat_lower` thay vì điểm `yhat`: người dùng cần biết "có khả năng còn giảm" hơn là một con số giá chính xác.
- Tính tất định (seed + uncertainty cố định) là yêu cầu để test lặp lại và để hai lần chạy cùng đầu vào không cho hai lời khuyên khác nhau.

---

*Hết FR-DEAL-004. Status: ready_to_implement (mục tiêu audit 10/10).*
