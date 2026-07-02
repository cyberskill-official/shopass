import pytest
import pandas as pd
from bottom.feature_store import build_features, FEATURE_COLS

def test_feature_vector_complete():
    # Setup dummy history data
    dates = pd.date_range("2026-01-01", "2026-06-06", freq="D")
    df = pd.DataFrame({
        "day": dates,
        "close_p": [10000] * len(dates),
        "min_p": [9000] * len(dates),
        "flash_sale": [False] * (len(dates) - 1) + [True]
    })
    
    as_of = pd.Timestamp("2026-06-06")
    feats = build_features(df, as_of=as_of, category_id=4221, platform_id=1)
    
    for col in FEATURE_COLS:
        assert col in feats, f"Missing feature: {col}"
    
    assert feats["is_double_date"] == 1
    assert feats["flash_sale_flag"] == 1

def test_features_use_no_future():
    dates = pd.date_range("2026-01-01", "2026-06-08", freq="D")
    df = pd.DataFrame({
        "day": dates,
        "close_p": [10000] * len(dates),
        "min_p": [9000] * len(dates),
        "flash_sale": [False] * len(dates)
    })
    
    as_of = pd.Timestamp("2026-06-06")
    
    # Inject a cheap price in the future
    df.loc[df["day"] == "2026-06-07", "min_p"] = 1000
    
    feats = build_features(df, as_of=as_of, category_id=4221, platform_id=1)
    assert feats["trailing_min_30"] == 9000, "Should not include future data"
