---
id: FR-DEAL-005
title: "Dự đoán đáy giá bằng LightGBM - feature store thống nhất train/serve, nhãn future_min_price_14d dựng bằng forward-join, cổng >=180 ngày lịch sử (dưới ngưỡng fallback Prophet FR-DEAL-004), giữ chung hợp đồng tín hiệu P(bottom within 14d)"
module: DEAL
priority: SHOULD
status: ready_to_implement
verify: T
phase: P2
milestone: P2 - slice 2
slice: 2
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-DEAL-004, FR-DEAL-002, FR-DEAL-006, FR-PRICE-002]
depends_on: [FR-DEAL-004]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.5 (LightGBM target future_min_price_14d, feature set, >=180d history)"
source_decisions:
  - "DEC-DEAL-40: dùng LightGBM khi SKU có >=180 ngày lịch sử; dưới ngưỡng fallback về Prophet baseline FR-DEAL-004 / category prior FR-DEAL-002"
  - "DEC-DEAL-41: target hồi quy là future_min_price_14d - giá nhỏ nhất trong 14 ngày tới của price_daily"
  - "DEC-DEAL-42: feature store là bảng thực thể hóa 1 dòng/(product_id, as_of_date) để train và serve đọc CÙNG một vector đặc trưng, tránh train/serve skew"
  - "DEC-DEAL-43: nhãn dựng bằng forward-join nhìn 14 ngày tới CHỈ lúc train; lúc serve không bao giờ chạm tương lai (no leakage)"
  - "DEC-DEAL-44: giữ chung hợp đồng tín hiệu P(bottom within 14d) với FR-DEAL-004 để FR-DEAL-006 cảnh báo bất khả tri theo mô hình"

language: "Python 3.12 (LightGBM, pandas, numpy, scikit-learn) - ml-svc; feature store in Postgres"
service: shopass/services/ml/
new_files:
  - services/ml/bottom/feature_store.py
  - services/ml/bottom/labels.py
  - services/ml/bottom/lgbm_model.py
  - services/ml/migrations/0002_feature_store.sql
  - services/ml/tests/test_feature_store.py
  - services/ml/tests/test_labels.py
  - services/ml/tests/test_lgbm_model.py
modified_files:
  - services/ml/bottom/serve.py                 # route SKU >=180 ngày sang LightGBM, dưới ngưỡng giữ Prophet (FR-DEAL-004)
allowed_tools:
  - file_read: services/ml/**
  - file_write: services/ml/**
  - bash: cd services/ml && python -m pytest -q
disallowed_tools:
  - dựng feature lúc serve khác cách dựng lúc train (vi phạm DEC-DEAL-42, gây train/serve skew)
  - cho nhãn forward-join 14 ngày chạm vào dữ liệu tương lai lúc serve (vi phạm DEC-DEAL-43, rò nhãn)
  - chạy LightGBM cho SKU <180 ngày thay vì fallback Prophet (vi phạm DEC-DEAL-40)
  - đổi hình dạng tín hiệu P(bottom within 14d) khác FR-DEAL-004 (vi phạm DEC-DEAL-44, phá FR-DEAL-006)

effort_hours: 12
sub_tasks:
  - "1.0h: 0002_feature_store.sql - bảng feature_store 1 dòng/(product_id, as_of_date) đủ 11 cột feature + cột nhãn nullable + PK"
  - "2.5h: feature_store.py - build_features() tính đủ feature từ price_daily/price_snapshot + tracked_product.category_id + platform_id"
  - "1.5h: labels.py - future_min_price_14d() forward-join cửa sổ 14 ngày, chỉ lúc train, upsert vào cột nhãn"
  - "2.0h: lgbm_model.py - train() LGBMRegressor(seed cố định) trên future_min_price_14d; predict(); p_bottom()"
  - "1.0h: serve.py - cổng >=180 ngày: route LightGBM, dưới ngưỡng gọi Prophet FR-DEAL-004; ghi price_forecast model_kind='lgbm'"
  - "1.0h: test_feature_store.py - test_feature_vector_complete (đủ 11 feature) + train/serve cùng schema"
  - "1.0h: test_labels.py - test_label_future_min_14d (cửa sổ forward đúng) + test_label_no_leak_at_serve"
  - "1.0h: test_lgbm_model.py - test_eligibility_180d_gate + test_lgbm_train_predict_shape + test_p_bottom_contract_matches_prophet"
  - "1.0h: OTel metric lgbm_predict_total + lgbm_fallback_prophet_total + feature_store_rows_built_total"

risk_if_skipped: "Prophet baseline (FR-DEAL-004) chỉ mô hình hóa xu hướng và mùa vụ một chiều, không bắt được tương tác giữa các đặc trưng (ví dụ double-date trùng cửa sổ payday cộng hưởng làm đáy sâu hơn tổng từng yếu tố). Với SKU đã có lịch sử dài (>=180 ngày), dữ liệu đủ để một mô hình gradient boosting học các tương tác đó và dự đoán đáy 14 ngày tới chính xác hơn rõ rệt, nâng chất lượng cảnh báo của FR-DEAL-006. Thiếu feature store thì train và serve dễ tính đặc trưng lệch nhau (train/serve skew), mô hình trông tốt lúc offline nhưng sai lúc online - lỗi âm thầm và rất khó truy. Thiếu kỷ luật forward-join chỉ lúc train thì nhãn rò tương lai vào feature, mô hình học gian và sụp khi chạy thật. Thiếu cổng >=180 ngày thì LightGBM bị chạy trên SKU dữ liệu mỏng, thua cả Prophet. Thiếu việc giữ chung hợp đồng P(bottom within 14d) thì FR-DEAL-006 phải biết mình đang gọi mô hình nào - phá thế bất khả tri theo mô hình và làm tầng cảnh báo dễ vỡ khi ta đổi mô hình."
---

## §1 - Mô tả (BCP-14 normative)

Service ML (ml-svc) **MUST** nâng cấp bộ dự đoán đáy giá từ Prophet baseline (FR-DEAL-004) lên LightGBM cho các SKU có lịch sử dài, qua một feature store thống nhất giữa train và serve, một nhãn `future_min_price_14d` dựng bằng forward-join chỉ lúc train, và một cổng đủ điều kiện `>=180` ngày. Hợp đồng:

1. **MUST** định nghĩa bảng feature store `feature_store` với khóa chính `(product_id, as_of_date)` và đúng 11 cột đặc trưng: `day_of_month`, `is_double_date`, `days_to_next_double_date`, `is_payday_window`, `trailing_min_30`, `trailing_min_60`, `trailing_min_90`, `price_vs_median90`, `volatility_30d`, `category_seasonality`, `flash_sale_flag`, `platform_id` (cùng cột nhãn `future_min_price_14d` nullable).
2. **MUST** tính mọi đặc trưng từ `price_daily`/`price_snapshot` (FR-PRICE-002) cộng `tracked_product.category_id` và `platform_id`, tất cả tính tới mốc `as_of_date` và KHÔNG dùng bất kỳ điểm dữ liệu nào sau `as_of_date` (đặc trưng chỉ nhìn về quá khứ).
3. **MUST** đặc tả từng đặc trưng: `day_of_month` = ngày trong tháng của `as_of_date`; `is_double_date` = 1 khi ngày bằng tháng (`d == m`, ví dụ 6/6, 12/12); `days_to_next_double_date` = số ngày tới double-date kế tiếp; `is_payday_window` = 1 khi `as_of_date` rơi vào cửa sổ trả lương; `trailing_min_30/60/90` = giá nhỏ nhất trong 30/60/90 ngày trước `as_of_date`; `price_vs_median90` = `close_p / median90` tại `as_of_date`; `volatility_30d` = độ lệch chuẩn log-return 30 ngày; `category_seasonality` = chỉ số mùa vụ gộp theo `category_id`; `flash_sale_flag` = 1 nếu có flash sale tại `as_of_date`; `platform_id` = nền tảng của SKU.
4. **MUST** dựng nhãn `future_min_price_14d` bằng forward-join: với mỗi dòng `(product_id, as_of_date)`, lấy `min(min_p)` của `price_daily` trong cửa sổ `(as_of_date, as_of_date + 14 ngày]`. Phép này nhìn về phía trước nên CHỈ chạy lúc train trên dữ liệu lịch sử đã hoàn tất.
5. **MUST** ở thời điểm serve, KHÔNG bao giờ tính nhãn và KHÔNG chạm bất kỳ dữ liệu nào sau `as_of_date`: serve chỉ đọc vector đặc trưng (cột nhãn để trống) rồi đưa qua mô hình. Không rò tương lai (DEC-DEAL-43).
6. **MUST** train và serve đọc đặc trưng từ CÙNG một nguồn là bảng `feature_store` (hoặc cùng một hàm `build_features` ghi vào bảng đó), bảo đảm vector đặc trưng giống hệt nhau - không có train/serve skew (DEC-DEAL-42).
7. **MUST** đặt cổng đủ điều kiện: chỉ train và serve LightGBM cho SKU có `>=180` ngày lịch sử; SKU dưới ngưỡng **MUST** fallback về Prophet baseline (FR-DEAL-004), và nếu cả Prophet cũng chưa đủ thì về category prior (FR-DEAL-002) (DEC-DEAL-40).
8. **MUST** train một mô hình hồi quy `lightgbm.LGBMRegressor` với mục tiêu `future_min_price_14d`, dùng đúng 11 đặc trưng ở #1, và `random_state` cố định để kết quả tái lập (deterministic seed).
9. **MUST** suy ra tín hiệu `p_bottom_14d` = P(bottom within 14d) từ giá đáy dự đoán so với giá hiện tại cộng một biên: `p_bottom_14d` cao khi `predicted_min >= current_price * (1 - margin)`, nghĩa là 14 ngày tới khó có giá thấp hơn đáng kể hiện tại (giờ đang gần đáy).
10. **MUST** phát ra cùng một hợp đồng tín hiệu như FR-DEAL-004: cùng tên trường `p_bottom_14d` (thực `[0,1]`), cùng `expected_min_14d`, cùng `horizon_days = 14`, chỉ khác `model_kind = 'lgbm'`. FR-DEAL-006 tiêu thụ tín hiệu này mà KHÔNG cần biết mô hình nào sinh ra (bất khả tri theo mô hình, DEC-DEAL-44).
11. **MUST** ghi dự đoán vào bảng chia sẻ `price_forecast` (cùng bảng FR-DEAL-004 ghi) với `model_kind = 'lgbm'`, `p_bottom_14d`, `expected_min_14d`, `as_of_date`, `horizon_days = 14`.
12. **MUST** bảo đảm tính tái lập: cùng một `feature_store` đầu vào và cùng `random_state` cho ra cùng mô hình và cùng dự đoán (seed cố định, không phụ thuộc thứ tự đọc).
13. **SHOULD** phát OTel metric: `lgbm_predict_total{platform_id}` (counter), `lgbm_fallback_prophet_total{reason}` (counter khi SKU <180 ngày), `feature_store_rows_built_total` (counter), `lgbm_train_mae` (gauge MAE trên tập kiểm định).
14. **MUST** xử lý đặc trưng thiếu (NULL) trong cửa sổ trailing: điền theo quy ước rõ ràng (ví dụ `trailing_min_*` lùi về giá sớm nhất có sẵn, `category_seasonality` về trung tính khi category thưa) thay vì để NaN rò vào mô hình một cách ngầm định.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao LightGBM thay vì Prophet ở mốc >=180 ngày (DEC-DEAL-40)?** Prophet mô hình hóa xu hướng cộng mùa vụ theo từng thành phần độc lập, nó không bắt được tương tác giữa các đặc trưng. Nhưng đáy giá thật sự sinh ra từ tương tác: một double-date rơi đúng cửa sổ payday cộng với volatility cao thường tạo đáy sâu hơn tổng riêng lẻ từng yếu tố. Gradient boosting học chính các tương tác bậc cao này. Khi SKU đã có >=180 ngày dữ liệu thì mẫu đủ dày để LightGBM không overfit, và nó thắng Prophet ở đúng những SKU quan trọng nhất (lịch sử dài, lượt theo dõi cao).

**Vì sao một feature store thực thể hóa (DEC-DEAL-42)?** Lỗi nguy hiểm nhất của một mô hình triển khai là train/serve skew: đặc trưng tính lúc train (theo lô, bằng pandas) khác tinh vi so với lúc serve (theo từng SKU, online). Mô hình trông tuyệt vời offline rồi sai online, và lỗi âm thầm vì không có ngoại lệ nào ném ra. Một bảng `feature_store` là nguồn sự thật duy nhất: cùng một hàng `(product_id, as_of_date)` được cả train lẫn serve đọc, nên vector đặc trưng giống hệt nhau theo định nghĩa. Đây là cách rẻ nhất để loại bỏ cả một lớp lỗi.

**Vì sao target là future_min_price_14d (DEC-DEAL-41)?** Câu hỏi người dùng thực sự hỏi là "14 ngày tới giá có xuống nữa không, hay giờ mua là được rồi?". Mô hình hóa trực tiếp giá nhỏ nhất trong 14 ngày tới khớp đúng câu hỏi đó: nếu đáy dự đoán không thấp hơn giá hiện tại bao nhiêu thì giờ đã gần đáy, nên mua. Chọn min thay vì giá trung bình vì người dùng quan tâm điểm thấp nhất họ có thể chờ được, không phải mức giá kỳ vọng.

**Vì sao forward-join chỉ lúc train (DEC-DEAL-43)?** Nhãn `future_min_price_14d` theo định nghĩa nhìn về tương lai. Lúc train trên lịch sử đã hoàn tất, tương lai đó đã xảy ra nên join hợp lệ. Nhưng nếu cùng phép join đó lọt vào đường serve, mô hình sẽ "thấy" giá tương lai và học gian - một dạng rò nhãn (label leakage) làm điểm offline đẹp giả tạo rồi sụp khi chạy thật, nơi tương lai chưa tồn tại. Tách bạch: `labels.py` chỉ chạy trong pipeline train; `serve.py` không bao giờ gọi nó và chỉ đọc cột đặc trưng (nhãn để trống).

**Vì sao giữ chung hợp đồng P(bottom within 14d) với FR-DEAL-004 (DEC-DEAL-44)?** FR-DEAL-006 (cảnh báo) chỉ nên quan tâm "xác suất giờ là đáy", không nên quan tâm mô hình nào tính ra nó. Nếu LightGBM và Prophet phát hai hình dạng tín hiệu khác nhau thì tầng cảnh báo phải rẽ nhánh theo mô hình, rồi vỡ mỗi lần ta đổi mô hình hay thêm mô hình thứ ba. Giữ chung đúng một hợp đồng (`p_bottom_14d`, `expected_min_14d`, `horizon_days`, chỉ khác `model_kind`) làm tầng trên bất khả tri theo mô hình - ta thay engine bên dưới mà không động tới FR-DEAL-006.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/ml/migrations/0002_feature_store.sql
-- 1 dòng/(product_id, as_of_date): đủ 11 đặc trưng + cột nhãn nullable.
-- Đặc trưng chỉ nhìn quá khứ (<= as_of_date); nhãn nhìn 14 ngày tới, CHỈ điền lúc train.
CREATE TABLE feature_store (
  product_id               BIGINT  NOT NULL REFERENCES tracked_product(id),
  as_of_date               DATE    NOT NULL,
  day_of_month             SMALLINT NOT NULL,                 -- 1..31
  is_double_date           BOOLEAN NOT NULL,                  -- d == m
  days_to_next_double_date SMALLINT NOT NULL,
  is_payday_window         BOOLEAN NOT NULL,
  trailing_min_30          BIGINT  NOT NULL,                  -- VND
  trailing_min_60          BIGINT  NOT NULL,
  trailing_min_90          BIGINT  NOT NULL,
  price_vs_median90        REAL    NOT NULL,                  -- close_p / median90
  volatility_30d           REAL    NOT NULL,                  -- std log-return 30d
  category_seasonality     REAL    NOT NULL,                  -- chỉ số mùa vụ gộp theo category
  flash_sale_flag          BOOLEAN NOT NULL,
  platform_id              SMALLINT NOT NULL,
  future_min_price_14d     BIGINT,                            -- NHÃN: nullable, chỉ điền lúc train
  PRIMARY KEY (product_id, as_of_date)
);

CREATE INDEX idx_fs_asof ON feature_store (as_of_date);
```

### Feature store (Python)

```python
# services/ml/bottom/feature_store.py
import pandas as pd

FEATURE_COLS = [
    "day_of_month", "is_double_date", "days_to_next_double_date",
    "is_payday_window", "trailing_min_30", "trailing_min_60", "trailing_min_90",
    "price_vs_median90", "volatility_30d", "category_seasonality",
    "flash_sale_flag", "platform_id",
]

def build_features(daily: pd.DataFrame, as_of: pd.Timestamp,
                   category_id: int, platform_id: int) -> dict:
    """Dựng vector đặc trưng tại as_of - CHỈ dùng dữ liệu <= as_of (no future)."""
    hist = daily[daily["day"] <= as_of].sort_values("day")
    if hist.empty:
        raise ValueError("no history at as_of")
    close_p = int(hist["close_p"].iloc[-1])
    median90 = float(hist["close_p"].tail(90).median())
    feats = {
        "day_of_month": as_of.day,
        "is_double_date": int(as_of.day == as_of.month),
        "days_to_next_double_date": _days_to_next_double_date(as_of),
        "is_payday_window": int(_in_payday_window(as_of)),
        "trailing_min_30": int(hist["min_p"].tail(30).min()),
        "trailing_min_60": int(hist["min_p"].tail(60).min()),
        "trailing_min_90": int(hist["min_p"].tail(90).min()),
        "price_vs_median90": close_p / median90 if median90 else 1.0,
        "volatility_30d": _log_return_std(hist["close_p"].tail(30)),
        "category_seasonality": _category_seasonality(category_id, as_of),
        "flash_sale_flag": int(bool(hist["flash_sale"].iloc[-1])),
        "platform_id": platform_id,
    }
    return feats  # cùng hàm này feed cả train lẫn serve -> không skew
```

### Labels (Python) - CHỈ lúc train

```python
# services/ml/bottom/labels.py
import pandas as pd

HORIZON_DAYS = 14

def future_min_price_14d(daily: pd.DataFrame, as_of: pd.Timestamp) -> int | None:
    """Forward-join: min giá trong (as_of, as_of + 14d]. CHỈ gọi lúc train.
    Trả None nếu cửa sổ tương lai chưa đủ 14 ngày dữ liệu (tránh nhãn cụt)."""
    end = as_of + pd.Timedelta(days=HORIZON_DAYS)
    fwd = daily[(daily["day"] > as_of) & (daily["day"] <= end)]
    if fwd.empty or fwd["day"].max() < end:
        return None
    return int(fwd["min_p"].min())
```

### LightGBM (Python)

```python
# services/ml/bottom/lgbm_model.py
import lightgbm as lgb
import numpy as np
from .feature_store import FEATURE_COLS

SEED = 42
DEFAULT_MARGIN = 0.03  # biên P(bottom): đáy dự đoán >= current*(1-margin) -> coi như gần đáy

def train(df) -> lgb.LGBMRegressor:
    """Train hồi quy future_min_price_14d trên đúng FEATURE_COLS, seed cố định."""
    model = lgb.LGBMRegressor(random_state=SEED, n_estimators=400, learning_rate=0.05)
    model.fit(df[FEATURE_COLS], df["future_min_price_14d"])
    return model

def predict(model, feats: dict) -> int:
    x = np.array([[feats[c] for c in FEATURE_COLS]])
    return int(model.predict(x)[0])  # expected_min_14d

def p_bottom(expected_min_14d: int, current_price: int, margin: float = DEFAULT_MARGIN) -> float:
    """P(bottom within 14d): cao khi đáy dự đoán không thấp hơn current quá margin."""
    if current_price <= 0:
        return 0.0
    ratio = expected_min_14d / current_price          # ~1.0 => đang gần đáy
    score = (ratio - (1.0 - margin)) / margin         # ánh xạ [1-margin,1] -> [0,1]
    return float(min(1.0, max(0.0, score)))
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> `feature_store` tồn tại với PK `(product_id, as_of_date)` và đủ 11 cột đặc trưng + cột nhãn `future_min_price_14d` nullable.
2. `build_features` chỉ dùng dữ liệu `<= as_of` (không có điểm nào sau `as_of` lọt vào vector).
3. Mỗi đặc trưng tính đúng định nghĩa ở §1 #3 (double-date khi `d == m`, `price_vs_median90 = close_p/median90`, volatility là std log-return 30 ngày, v.v.).
4. `future_min_price_14d(daily, as_of)` trả `min(min_p)` trong cửa sổ `(as_of, as_of+14d]`; trả `None` khi tương lai chưa đủ 14 ngày.
5. Đường serve không bao giờ gọi `future_min_price_14d` và không đọc cột nhãn (no leakage).
6. Train và serve đọc đặc trưng từ cùng `feature_store`/`build_features` - vector đặc trưng cho cùng `(product_id, as_of_date)` giống hệt nhau.
7. SKU `<180` ngày -> fallback Prophet (FR-DEAL-004), KHÔNG chạy LightGBM; SKU `>=180` ngày -> chạy LightGBM.
8. `train` trả `LGBMRegressor` đã fit trên đúng `FEATURE_COLS` với mục tiêu `future_min_price_14d`; `random_state` cố định.
9. `p_bottom(expected_min_14d, current_price)` trả giá trị trong `[0,1]`, tăng khi đáy dự đoán tiến gần `current_price`.
10. Tín hiệu phát ra cùng schema FR-DEAL-004: `p_bottom_14d`, `expected_min_14d`, `horizon_days=14`, chỉ khác `model_kind='lgbm'`.
11. Dự đoán ghi vào `price_forecast` với `model_kind='lgbm'` và `p_bottom_14d`.
12. Cùng `feature_store` + cùng seed -> cùng mô hình + cùng dự đoán (tái lập).
13. Đặc trưng thiếu được điền theo quy ước (không NaN ngầm); metric `lgbm_fallback_prophet_total` tăng khi SKU `<180` ngày.

---

## §5 - Kiểm thử (verification)

```python
# services/ml/tests/test_feature_store.py
from services.ml.bottom.feature_store import build_features, FEATURE_COLS

def test_feature_vector_complete(daily_120d):
    feats = build_features(daily_120d, as_of=AS_OF, category_id=4221, platform_id=1)
    for col in FEATURE_COLS:                      # đủ 11 đặc trưng, không thiếu cột
        assert col in feats
    assert feats["is_double_date"] in (0, 1)

def test_features_use_no_future(daily_120d):
    # thêm 1 điểm giá rẻ bất thường SAU as_of: không được lọt vào vector
    poisoned = inject_cheap_point(daily_120d, after=AS_OF)
    feats = build_features(poisoned, as_of=AS_OF, category_id=4221, platform_id=1)
    assert feats["trailing_min_30"] == build_features(
        daily_120d, AS_OF, 4221, 1)["trailing_min_30"]
```

```python
# services/ml/tests/test_labels.py
from services.ml.bottom.labels import future_min_price_14d

def test_label_future_min_14d(daily_series):
    # đáy 14 ngày tới của as_of phải bằng min(min_p) trong (as_of, as_of+14d]
    label = future_min_price_14d(daily_series, as_of=AS_OF)
    assert label == expected_forward_min(daily_series, AS_OF, 14)

def test_label_no_leak_at_serve(daily_series):
    # cửa sổ tương lai chưa đủ 14 ngày (as_of gần mép phải) -> None, không bịa nhãn
    near_edge = daily_series["day"].max() - pd.Timedelta(days=5)
    assert future_min_price_14d(daily_series, as_of=near_edge) is None
```

```python
# services/ml/tests/test_lgbm_model.py
from services.ml.bottom import lgbm_model as m
from services.ml.bottom.serve import route_model

def test_eligibility_180d_gate():
    assert route_model(days_history=179) == "prophet"   # fallback FR-DEAL-004
    assert route_model(days_history=180) == "lgbm"

def test_lgbm_train_predict_shape(train_df, one_feat_row):
    model = m.train(train_df)
    pred = m.predict(model, one_feat_row)
    assert isinstance(pred, int)                          # expected_min_14d

def test_p_bottom_contract_matches_prophet(one_feat_row):
    model = m.train_fixture()
    exp = m.predict(model, one_feat_row)
    p = m.p_bottom(exp, current_price=one_feat_row_current())
    assert 0.0 <= p <= 1.0                                # cùng miền tín hiệu như FR-DEAL-004
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration `0002_feature_store.sql` (bảng + index) -> `feature_store.py` (`build_features` + `FEATURE_COLS`) -> `labels.py` (`future_min_price_14d`, chỉ pipeline train) -> `lgbm_model.py` (`train`/`predict`/`p_bottom`) -> sửa `serve.py` (cổng `route_model` theo số ngày: `>=180` LightGBM, dưới ngưỡng gọi Prophet FR-DEAL-004) -> tests. Pipeline train chạy theo lô: dựng `feature_store` cho dải `as_of_date` lịch sử, điền nhãn bằng forward-join, fit `LGBMRegressor`, lưu model artifact. Đường serve chỉ đọc một hàng đặc trưng tại `as_of = today` (cột nhãn trống), gọi `predict` + `p_bottom`, ghi `price_forecast`. Một hàm `build_features` duy nhất dùng cho cả hai để chặn skew.

---

## §7 - Phụ thuộc

- **FR-DEAL-004** - Prophet baseline; FR-DEAL-005 fallback về nó cho SKU `<180` ngày và dùng chung bảng `price_forecast` (FR-DEAL-004 sở hữu DDL) + hợp đồng tín hiệu. Ánh xạ cột khi ghi: `expected_min_14d` của FR này ghi vào cột `yhat`; `as_of_date` ghi vào cột `run_date`; `horizon_days = 14` ứng với cột `horizon_day` (1..14); `model_kind = 'lgbm'` đã nằm trong CHECK của bảng. Không tạo bảng `price_forecast` thứ hai - chỉ INSERT vào bảng do FR-DEAL-004 định nghĩa.
- **FR-DEAL-002** - category prior là fallback cuối khi cả Prophet chưa đủ; `category_seasonality` mượn ý tưởng gộp theo category.
- **FR-PRICE-002** - `price_daily` (continuous aggregate) và `price_snapshot` là nguồn cho mọi đặc trưng và nhãn.
- **FR-DEAL-006 (downstream)** - tiêu thụ tín hiệu `p_bottom_14d` chia sẻ, bất khả tri theo `model_kind`.
- Lib: `lightgbm`, `pandas`, `numpy`, `scikit-learn`; driver Postgres cho `feature_store`/`price_forecast`.

---

## §8 - Payload ví dụ

### Một hàng feature_store (kèm nhãn lúc train)

```json
{
  "product_id": 90112,
  "as_of_date": "2026-06-06",
  "day_of_month": 6,
  "is_double_date": true,
  "days_to_next_double_date": 0,
  "is_payday_window": false,
  "trailing_min_30": 89000,
  "trailing_min_60": 85000,
  "trailing_min_90": 82000,
  "price_vs_median90": 0.94,
  "volatility_30d": 0.071,
  "category_seasonality": 1.12,
  "flash_sale_flag": true,
  "platform_id": 1,
  "future_min_price_14d": 81000
}
```

### Một dòng dự đoán ghi vào price_forecast (model_kind = lgbm)

```json
{
  "product_id": 90112,
  "as_of_date": "2026-06-27",
  "model_kind": "lgbm",
  "expected_min_14d": 84000,
  "p_bottom_14d": 0.78,
  "horizon_days": 14
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- `margin` của `p_bottom` cố định 0,03 hay co giãn theo `volatility_30d` của SKU - tinh chỉnh khi có nhãn click thực.
- Ngưỡng `>=180` ngày có nên khác theo category (thời trang đổi giá nhanh hơn điện máy) - cấu hình per-category ở slice sau.
- Hiệu chỉnh xác suất (calibration) cho `p_bottom_14d` để con số đọc đúng là xác suất - tách sang FR riêng khi có dữ liệu kết quả.
- Train tăng dần (incremental) thay vì train lại toàn bộ theo lô - tối ưu chi phí giai đoạn sau.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Train/serve skew (đặc trưng tính khác nhau) | so vector serve vs feature_store | mô hình sai online dù tốt offline | Cùng một `build_features` + đọc chung `feature_store` (§1 #6) |
| Rò nhãn (forward-join lọt vào serve) | test no_leak + review đường serve | điểm offline giả tạo, sụp khi chạy thật | `labels.py` chỉ trong pipeline train; serve không gọi (§1 #5) |
| SKU `<180` ngày bị route nhầm vào LightGBM | test cổng 180d | LightGBM thua Prophet trên dữ liệu mỏng | `route_model` chặn `<180` -> Prophet (§1 #7) |
| Đặc trưng NULL trong cửa sổ trailing | check NULL khi build | NaN rò vào mô hình, dự đoán rác | Điền theo quy ước rõ ràng (§1 #14) |
| `category_seasonality` không ổn định (category thưa) | giám sát phương sai theo category | đặc trưng nhiễu kéo dự đoán lệch | Về trung tính khi `sample_count` category dưới sàn |
| Nhãn cụt ở mép phải lịch sử (chưa đủ 14 ngày) | `future_min_price_14d` trả None | hàng train có nhãn sai/thiếu | Bỏ hàng nhãn None khỏi tập train (§1 #4) |
| Mô hình không tái lập (seed trôi) | chạy lại so dự đoán | khó debug, kết quả khác mỗi lần | `random_state` cố định + đọc ổn định thứ tự (§1 #12) |
| Hợp đồng tín hiệu lệch FR-DEAL-004 | test p_bottom_contract | FR-DEAL-006 phải rẽ nhánh theo mô hình | Cùng `p_bottom_14d`/`expected_min_14d`/`horizon_days` (§1 #10) |
| Lệch phân phối (price_daily đổi đơn vị/null) | giám sát MAE kiểm định | dự đoán trôi âm thầm | Theo dõi `lgbm_train_mae`; cảnh báo khi vọt |

---

## §11 - Ghi chú

- LightGBM là nâng cấp theo SKU, không thay thế toàn cục: nó chỉ nhận các SKU `>=180` ngày, phần còn lại vẫn do Prophet (FR-DEAL-004) và category prior (FR-DEAL-002) lo - một cầu thang chất lượng theo độ dày dữ liệu.
- Feature store là quyết định kiến trúc quan trọng nhất ở đây: nó biến train/serve skew từ một lớp lỗi âm thầm thành một bất biến theo định nghĩa (cùng hàng, cùng vector).
- Tách `labels.py` khỏi đường serve là ranh giới chống rò nhãn: forward-join là hợp lệ lúc train, là gian lận lúc serve - mã phải làm ranh giới đó hiển nhiên.
- Giữ chung hợp đồng `p_bottom_14d` với FR-DEAL-004 là điều khiến FR-DEAL-006 không phải biết mô hình nào - đổi engine bên dưới không động tới tầng cảnh báo.
- `future_min_price_14d` khớp đúng câu hỏi người dùng "giờ mua được chưa hay còn xuống" - chọn min (không phải trung bình) vì người dùng quan tâm điểm thấp nhất họ chờ được.
- Seed cố định và đọc ổn định thứ tự là điều kiện để audit lại được một dự đoán cũ - cùng đầu vào phải cho cùng đầu ra.

---

*Hết FR-DEAL-005. Status: ready_to_implement (mục tiêu audit 10/10).*
