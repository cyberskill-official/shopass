import pandas as pd
import pytest

from bottom.gate import (
    MAPE_RATIO,
    apply_gate_to_forecast,
    evaluate_gate,
    holdout_mape,
    mape,
)


def test_mape_basic():
    assert mape([100, 100], [110, 90]) == pytest.approx(0.1)


def test_gate_accepts_within_ratio():
    best = 0.10
    g = evaluate_gate(0.11, best)
    assert g.passed is True
    assert g.use_baseline_fallback is False


def test_gate_rejects_above_ratio():
    best = 0.10
    bad = best * MAPE_RATIO + 0.01
    g = evaluate_gate(bad, best)
    assert g.passed is False
    assert g.suppress_p_bottom is True
    assert g.use_baseline_fallback is True


def test_gate_allows_cold_start_without_mape():
    g = evaluate_gate(None, 0.05)
    assert g.passed is True
    assert g.reason == "insufficient_history_allow_cold_start"


def test_apply_gate_suppresses_p_bottom():
    df = pd.DataFrame({"p_bottom_14d": [0.9, 0.8], "model_kind": ["lgbm", "lgbm"]})
    g = evaluate_gate(0.5, 0.1)
    assert g.passed is False
    out = apply_gate_to_forecast(df, g)
    assert list(out["p_bottom_14d"]) == [0.0, 0.0]


def test_holdout_mape_none_on_short_history():
    hist = pd.DataFrame({
        "ds": pd.to_datetime(["2026-01-01", "2026-01-02"]),
        "y": [100, 110],
    })
    assert holdout_mape(hist) is None
