from . import prophet_baseline

def route_model(days_history: int) -> str:
    """
    Route mô hình dựa trên số ngày lịch sử của SKU.
    DEC-DEAL-40: >=180 ngày -> LightGBM; <180 ngày -> Prophet
    """
    if days_history >= 180:
        return "lgbm"
    return "prophet"
