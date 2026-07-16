import numpy as np
import pandas as pd
import pytest
from bottom.prophet_baseline import forecast_bottom, _p_bottom_14d, build_baseline
from bottom.features import HORIZON_DAYS


def _series(days: int, base: int = 100_000) -> pd.DataFrame:
    ds = pd.date_range("2025-01-01", periods=days, freq="D")
    y = base + (np.sin(np.arange(days) / 7.0) * 5_000)  # nhịp nhẹ quanh base
    return pd.DataFrame({"ds": ds, "y": y, "flash_sale": 0})


def test_prophet_forecast_shape():
    # Prove the real shipped entry point works when the Stan backend is healthy.
    try:
        m = build_baseline(seed=0)
        assert getattr(m, "stan_backend", None) is not None
    except Exception as exc:  # pragma: no cover - environment gate
        pytest.fail(f"Prophet/CmdStan backend failed to initialize: {exc!r}")

    out = forecast_bottom(_series(180))  # MATURE -> fit Prophet
    assert len(out) == HORIZON_DAYS
    assert list(out["horizon_day"]) == list(range(1, 15))
    assert (out["model_kind"] == "prophet").all()
    assert out["yhat"].dtype == np.int64


def test_cold_start_uses_prior():
    out = forecast_bottom(_series(10), prior_median=88_000)  # cold-start
    assert (out["model_kind"] == "category_prior").all()
    assert (out["yhat"] == 88_000).all()
    assert (out["p_bottom_14d"] == 0.0).all()  # không hứa đáy khi thiếu dữ liệu


def test_p_bottom_monotonic_with_interval():
    trailing_min = 80_000
    narrow = _p_bottom_14d(np.full(14, 79_000.0), trailing_min)  # cận dưới sát đáy
    wide = _p_bottom_14d(np.full(14, 70_000.0), trailing_min)    # cận dưới phá đáy sâu hơn
    assert 0.0 <= narrow <= wide <= 1.0  # khoảng nới xuống thấp -> xác suất không giảm
