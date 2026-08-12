import numpy as np
import pandas as pd
from prophet import Prophet
from .features import build_regressor_frame, HORIZON_DAYS

REGRESSORS = ("is_double_date", "is_payday_window", "flash_sale")


import os
from pathlib import Path

# Prefer an existing CmdStan install (any version under ~/.cmdstan/cmdstan-*).
# CI pins cmdstanpy.install_cmdstan(version="2.39.0") and exports CMDSTAN;
# do not hard-code that path here — local installs may use another version.
def _configure_cmdstan() -> None:
    try:
        import cmdstanpy
    except ImportError:
        return
    env = os.environ.get("CMDSTAN")
    if env and Path(env).is_dir():
        cmdstanpy.set_cmdstan_path(env)
        return
    try:
        # Already configured (e.g. after install_cmdstan in CI).
        if cmdstanpy.cmdstan_path():
            return
    except Exception:
        pass
    home = Path.home() / ".cmdstan"
    if not home.is_dir():
        return
    candidates = sorted(home.glob("cmdstan-*"), reverse=True)
    for cand in candidates:
        if cand.is_dir():
            cmdstanpy.set_cmdstan_path(str(cand))
            return


_configure_cmdstan()

def build_baseline(seed: int = 42) -> Prophet:
    """Prophet baseline: mùa vụ năm + tháng, regressor double-date/payday/flash (DEC-DEAL-30)."""
    np.random.seed(seed)
    _configure_cmdstan()
    # Force CMDSTANPY so Prophet does not silently exhaust backends and crash
    # with "no attribute 'stan_backend'" when default discovery fails.
    try:
        m = Prophet(
            seasonality_mode="multiplicative",
            yearly_seasonality=True,
            weekly_seasonality=False,
            daily_seasonality=False,
            mcmc_samples=0,
            uncertainty_samples=1000,
            stan_backend="CMDSTANPY",
        )
    except Exception:
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
