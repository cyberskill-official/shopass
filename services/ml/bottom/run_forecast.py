"""Runnable forecast job: read price history -> forecast -> gate -> write price_forecast.

Usage: DATABASE_URL=postgres://... python -m bottom.run_forecast

Model policy (DEC-DEAL-40 / DEC-DEAL-32): a mature SKU (>= MATURE_DAYS distinct
days) is fit with Prophet; an immature SKU uses the cold-start prior baseline
(model_kind='category_prior'), which needs no Stan backend.

R26: each job records a model_run, evaluates a holdout MAPE gate, and suppresses
p_bottom (or falls back to category_prior) when the gate fails.
"""
from __future__ import annotations

import logging
import os
import urllib.request
from datetime import date

import pandas as pd

from . import db
from .gate import apply_gate_to_forecast, evaluate_gate, holdout_mape
from .model_run import (
    artifact_dir,
    best_trailing_mape,
    feature_set_hash,
    insert_model_run,
    next_version,
    write_artifact_stub,
)
from .prophet_baseline import forecast_bottom

log = logging.getLogger("mlforecast")

MATURE_DAYS = 90


def forecast_for(history: pd.DataFrame) -> pd.DataFrame:
    days = history["ds"].dt.date.nunique() if not history.empty else 0
    if days < MATURE_DAYS:
        prior = int(history["y"].median()) if not history.empty else 0
        return forecast_bottom(history, prior_median=prior)
    return forecast_bottom(history)  # Prophet (requires CmdStan)


def _push_gate_trip(reason: str) -> None:
    """Best-effort Pushgateway counter for R13 alert ShopassModelGateTripped."""
    url = os.environ.get("PUSHGATEWAY_URL", "").rstrip("/")
    if not url:
        return
    body = (
        "# TYPE shopass_ml_gate_trips_total counter\n"
        f'shopass_ml_gate_trips_total{{reason="{reason[:64]}"}} 1\n'
    )
    try:
        req = urllib.request.Request(
            f"{url}/metrics/job/shopass_ml/instance/gate",
            data=body.encode(),
            method="POST",
        )
        urllib.request.urlopen(req, timeout=3)
    except Exception as exc:  # noqa: BLE001 — observability must not fail the job
        log.warning("pushgateway gate metric failed: %s", exc)


def run(conn, run_date: date | None = None) -> int:
    run_date = run_date or date.today()
    written = 0
    fhash = feature_set_hash()

    # One model_run per job invocation (aggregate quality across products).
    # Candidate MAPE = mean of per-product holdout MAPEs that are available.
    product_mapes: list[float] = []
    forecasts: list[tuple[int, pd.DataFrame, date | None, date | None]] = []

    for pid in db.distinct_products(conn):
        hist = db.read_history(conn, pid)
        if hist.empty:
            continue
        hist["ds"] = pd.to_datetime(hist["ds"])
        fdf = forecast_for(hist)
        m = holdout_mape(hist)
        if m is not None:
            product_mapes.append(m)
        tw_start = hist["ds"].min().date() if not hist.empty else None
        tw_end = hist["ds"].max().date() if not hist.empty else None
        forecasts.append((pid, fdf, tw_start, tw_end))

    if not forecasts:
        return 0

    # Dominant kind for the run (mature products prefer prophet).
    kinds = [fdf["model_kind"].iloc[0] for _, fdf, _, _ in forecasts]
    model_kind = "prophet" if "prophet" in kinds else kinds[0]
    candidate_mape = sum(product_mapes) / len(product_mapes) if product_mapes else None
    best = best_trailing_mape(conn, model_kind)
    gate = evaluate_gate(candidate_mape, best)

    version = next_version(run_date, model_kind)
    art = artifact_dir(version) / "meta.json"
    write_artifact_stub(
        art,
        {
            "version": version,
            "model_kind": model_kind,
            "feature_set_hash": fhash,
            "candidate_mape": candidate_mape,
            "best_mape": best,
            "gate": gate.__dict__,
            "products": len(forecasts),
        },
    )

    run_id = insert_model_run(
        conn,
        version=version,
        model_kind=model_kind,
        training_window_start=min((s for _, _, s, _ in forecasts if s), default=None),
        training_window_end=max((e for _, _, _, e in forecasts if e), default=None),
        feature_hash=fhash,
        backtest_mape=candidate_mape,
        backtest_hit_rate=None,
        gate_passed=gate.passed,
        gate_reason=gate.reason,
        artifact_path=str(art),
    )

    if not gate.passed:
        log.warning("R26 gate failed: %s — suppressing p_bottom / using baseline", gate.reason)
        _push_gate_trip(gate.reason)

    for pid, fdf, _, _ in forecasts:
        out = fdf
        if not gate.passed and gate.use_baseline_fallback:
            # Re-emit cold-start prior (safe p_bottom=0) instead of a bad mature fit.
            hist = db.read_history(conn, pid)
            if not hist.empty:
                hist["ds"] = pd.to_datetime(hist["ds"])
                prior = int(hist["y"].median())
                out = forecast_bottom(hist, prior_median=prior)
        out = apply_gate_to_forecast(out, gate)
        # Failed gate: do not publish bad mature forecasts — only baseline rows.
        if not gate.passed and str(out["model_kind"].iloc[0]) == "prophet":
            continue
        written += db.upsert_forecasts(
            conn, db.df_to_rows(pid, run_date, out, model_run_id=run_id)
        )
    return written


def main():
    logging.basicConfig(level=logging.INFO)
    conn = db.connect(os.environ["DATABASE_URL"])
    n = run(conn)
    print(f"wrote {n} price_forecast rows")


if __name__ == "__main__":
    main()
