import pytest
import pandas as pd
from bottom.labels import future_min_price_14d

def test_label_future_min_14d():
    dates = pd.date_range("2026-01-01", "2026-06-20", freq="D")
    df = pd.DataFrame({
        "day": dates,
        "min_p": [10000] * len(dates)
    })
    
    as_of = pd.Timestamp("2026-06-06")
    
    # Make a dip in the next 14 days
    df.loc[df["day"] == "2026-06-15", "min_p"] = 8000
    
    label = future_min_price_14d(df, as_of=as_of)
    assert label == 8000

def test_label_no_leak_at_serve():
    dates = pd.date_range("2026-01-01", "2026-06-20", freq="D")
    df = pd.DataFrame({
        "day": dates,
        "min_p": [10000] * len(dates)
    })
    
    # Not enough future data (less than 14 days)
    as_of = pd.Timestamp("2026-06-10")
    label = future_min_price_14d(df, as_of=as_of)
    assert label is None
