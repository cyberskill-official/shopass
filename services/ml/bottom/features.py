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
