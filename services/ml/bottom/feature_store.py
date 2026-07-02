import pandas as pd
import numpy as np

FEATURE_COLS = [
    "day_of_month", "is_double_date", "days_to_next_double_date",
    "is_payday_window", "trailing_min_30", "trailing_min_60", "trailing_min_90",
    "price_vs_median90", "volatility_30d", "category_seasonality",
    "flash_sale_flag", "platform_id",
]

def _days_to_next_double_date(as_of: pd.Timestamp) -> int:
    y, m, d = as_of.year, as_of.month, as_of.day
    if d <= m:
        target = pd.Timestamp(year=y, month=m, day=m)
    else:
        nm = m + 1
        ny = y
        if nm > 12:
            nm = 1
            ny = y + 1
        target = pd.Timestamp(year=ny, month=nm, day=nm)
    return (target - as_of).days

def _in_payday_window(as_of: pd.Timestamp) -> bool:
    d = as_of.day
    return (25 <= d <= 31) or (1 <= d <= 5)

def _log_return_std(prices: pd.Series) -> float:
    if len(prices) < 2:
        return 0.0
    r = np.log(prices / prices.shift(1)).dropna()
    if len(r) == 0:
        return 0.0
    return float(r.std())

def _category_seasonality(category_id: int, as_of: pd.Timestamp) -> float:
    # Dummy implementation for now, should query from a pre-computed table
    return 1.0

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
    return feats
