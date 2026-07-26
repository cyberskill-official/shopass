"""Backtest evaluation gate for forecast publishes (R26).

Publish when candidate MAPE is within MAPE_RATIO of the trailing best for the
same model_kind (default 1.2x). Otherwise fall back to Prophet/category_prior
baseline or suppress p_bottom.
"""
from __future__ import annotations

from dataclasses import dataclass

import numpy as np
import pandas as pd

# Candidate may be at most this multiple of the trailing 30-day best MAPE.
MAPE_RATIO = 1.2
MIN_BACKTEST_POINTS = 3


@dataclass(frozen=True)
class GateResult:
    passed: bool
    reason: str
    candidate_mape: float | None
    best_mape: float | None
    suppress_p_bottom: bool
    use_baseline_fallback: bool


def mape(y_true: np.ndarray, y_pred: np.ndarray) -> float:
    """Mean absolute percentage error; ignores zero actuals."""
    y_true = np.asarray(y_true, dtype=float)
    y_pred = np.asarray(y_pred, dtype=float)
    mask = np.abs(y_true) > 1e-9
    if not mask.any():
        return float("inf")
    return float(np.mean(np.abs((y_true[mask] - y_pred[mask]) / y_true[mask])))


def holdout_mape(history: pd.DataFrame, holdout_days: int = 7) -> float | None:
    """Naive last-value forecast MAPE on the trailing holdout window.

    Used as a cheap quality proxy when a full model backtest is unavailable
    (cold-start / short series). Returns None when history is too short.
    """
    if history is None or history.empty:
        return None
    hist = history.sort_values("ds")
    days = hist["ds"].dt.normalize().nunique() if hasattr(hist["ds"], "dt") else hist["ds"].nunique()
    if days < holdout_days + MIN_BACKTEST_POINTS:
        return None
    # Daily close: last price per calendar day.
    daily = hist.copy()
    daily["day"] = pd.to_datetime(daily["ds"]).dt.normalize()
    closes = daily.groupby("day", as_index=False)["y"].last().sort_values("day")
    if len(closes) < holdout_days + 1:
        return None
    train = closes.iloc[:-holdout_days]
    test = closes.iloc[-holdout_days:]
    last = float(train["y"].iloc[-1])
    preds = np.full(len(test), last)
    return mape(test["y"].to_numpy(), preds)


def evaluate_gate(
    candidate_mape: float | None,
    best_trailing_mape: float | None,
    *,
    ratio: float = MAPE_RATIO,
) -> GateResult:
    """Decide whether a candidate model run may publish forecasts."""
    if candidate_mape is None:
        # Insufficient history: allow cold-start prior (p_bottom already 0).
        return GateResult(
            passed=True,
            reason="insufficient_history_allow_cold_start",
            candidate_mape=None,
            best_mape=best_trailing_mape,
            suppress_p_bottom=False,
            use_baseline_fallback=False,
        )
    if best_trailing_mape is None or best_trailing_mape <= 0:
        return GateResult(
            passed=True,
            reason="no_baseline_accept",
            candidate_mape=candidate_mape,
            best_mape=best_trailing_mape,
            suppress_p_bottom=False,
            use_baseline_fallback=False,
        )
    limit = best_trailing_mape * ratio
    if candidate_mape <= limit:
        return GateResult(
            passed=True,
            reason="within_ratio",
            candidate_mape=candidate_mape,
            best_mape=best_trailing_mape,
            suppress_p_bottom=False,
            use_baseline_fallback=False,
        )
    return GateResult(
        passed=False,
        reason=f"mape={candidate_mape:.4f}>limit={limit:.4f}",
        candidate_mape=candidate_mape,
        best_mape=best_trailing_mape,
        suppress_p_bottom=True,
        use_baseline_fallback=True,
    )


def apply_gate_to_forecast(df: pd.DataFrame, gate: GateResult) -> pd.DataFrame:
    """Zero p_bottom when the gate suppresses publish of bottom signals."""
    out = df.copy()
    if gate.suppress_p_bottom and "p_bottom_14d" in out.columns:
        out["p_bottom_14d"] = 0.0
    return out
