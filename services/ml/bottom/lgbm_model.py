import lightgbm as lgb
import numpy as np
import pandas as pd
from .feature_store import FEATURE_COLS

SEED = 42
DEFAULT_MARGIN = 0.03  # biên P(bottom): đáy dự đoán >= current*(1-margin) -> coi như gần đáy

def train(df: pd.DataFrame) -> lgb.LGBMRegressor:
    """Train hồi quy future_min_price_14d trên đúng FEATURE_COLS, seed cố định."""
    model = lgb.LGBMRegressor(random_state=SEED, n_estimators=400, learning_rate=0.05)
    model.fit(df[FEATURE_COLS], df["future_min_price_14d"])
    return model

def predict(model: lgb.LGBMRegressor, feats: dict) -> int:
    x = np.array([[feats[c] for c in FEATURE_COLS]])
    return int(model.predict(x)[0])  # expected_min_14d

def p_bottom(expected_min_14d: int, current_price: int, margin: float = DEFAULT_MARGIN) -> float:
    """P(bottom within 14d): cao khi đáy dự đoán không thấp hơn current quá margin."""
    if current_price <= 0:
        return 0.0
    ratio = expected_min_14d / current_price          # ~1.0 => đang gần đáy
    score = (ratio - (1.0 - margin)) / margin         # ánh xạ [1-margin,1] -> [0,1]
    return float(min(1.0, max(0.0, score)))
