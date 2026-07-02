import pytest
import pandas as pd
from bottom import lgbm_model as m
from bottom.serve import route_model

def test_eligibility_180d_gate():
    assert route_model(days_history=179) == "prophet"
    assert route_model(days_history=180) == "lgbm"

@pytest.fixture
def train_df():
    df = pd.DataFrame({
        "day_of_month": [1, 2, 3, 4, 5],
        "is_double_date": [0, 0, 0, 1, 0],
        "days_to_next_double_date": [3, 2, 1, 0, 30],
        "is_payday_window": [1, 1, 1, 1, 1],
        "trailing_min_30": [90, 90, 90, 90, 90],
        "trailing_min_60": [80, 80, 80, 80, 80],
        "trailing_min_90": [80, 80, 80, 80, 80],
        "price_vs_median90": [1.0, 1.0, 1.0, 0.9, 0.95],
        "volatility_30d": [0.01, 0.02, 0.01, 0.03, 0.02],
        "category_seasonality": [1.0, 1.0, 1.0, 1.0, 1.0],
        "flash_sale_flag": [0, 0, 0, 1, 0],
        "platform_id": [1, 1, 1, 1, 1],
        "future_min_price_14d": [85, 85, 85, 80, 85]
    })
    return df

@pytest.fixture
def one_feat_row():
    return {
        "day_of_month": 4,
        "is_double_date": 1,
        "days_to_next_double_date": 0,
        "is_payday_window": 1,
        "trailing_min_30": 90,
        "trailing_min_60": 80,
        "trailing_min_90": 80,
        "price_vs_median90": 0.9,
        "volatility_30d": 0.03,
        "category_seasonality": 1.0,
        "flash_sale_flag": 1,
        "platform_id": 1
    }

def test_lgbm_train_predict_shape(train_df, one_feat_row):
    model = m.train(train_df)
    pred = m.predict(model, one_feat_row)
    assert isinstance(pred, int)

def test_p_bottom_contract_matches_prophet(train_df, one_feat_row):
    model = m.train(train_df)
    exp = m.predict(model, one_feat_row)
    p = m.p_bottom(exp, current_price=90)
    assert 0.0 <= p <= 1.0
