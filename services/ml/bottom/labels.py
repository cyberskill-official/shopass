import pandas as pd

HORIZON_DAYS = 14

def future_min_price_14d(daily: pd.DataFrame, as_of: pd.Timestamp) -> int | None:
    """Forward-join: min giá trong (as_of, as_of + 14d]. CHỈ gọi lúc train.
    Trả None nếu cửa sổ tương lai chưa đủ 14 ngày dữ liệu (tránh nhãn cụt)."""
    end = as_of + pd.Timedelta(days=HORIZON_DAYS)
    fwd = daily[(daily["day"] > as_of) & (daily["day"] <= end)]
    
    # We require the max day in the forward window to be at least `end`,
    # otherwise we don't have the full 14-day picture.
    # Note: this assumes daily has one row per day. If not every day has a row,
    # we might need to check if max() >= end or similar.
    # To be safe against missing days, we check if the max date in the series >= end.
    if fwd.empty or daily["day"].max() < end:
        return None
        
    return int(fwd["min_p"].min())
