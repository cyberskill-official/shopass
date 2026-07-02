"""Runnable forecast job: read price history -> forecast -> write price_forecast.

Usage: DATABASE_URL=postgres://... python -m bottom.run_forecast

Model policy (DEC-DEAL-40 / DEC-DEAL-32): a mature SKU (>= MATURE_DAYS distinct
days) is fit with Prophet; an immature SKU uses the cold-start prior baseline
(model_kind='category_prior'), which needs no Stan backend.
"""
import os
from datetime import date

import pandas as pd

from . import db
from .prophet_baseline import forecast_bottom

MATURE_DAYS = 90


def forecast_for(history: pd.DataFrame) -> pd.DataFrame:
    days = history["ds"].dt.date.nunique() if not history.empty else 0
    if days < MATURE_DAYS:
        prior = int(history["y"].median()) if not history.empty else 0
        return forecast_bottom(history, prior_median=prior)
    return forecast_bottom(history)  # Prophet (requires CmdStan)


def run(conn, run_date: date | None = None) -> int:
    run_date = run_date or date.today()
    written = 0
    for pid in db.distinct_products(conn):
        hist = db.read_history(conn, pid)
        if hist.empty:
            continue
        hist["ds"] = pd.to_datetime(hist["ds"])
        fdf = forecast_for(hist)
        written += db.upsert_forecasts(conn, db.df_to_rows(pid, run_date, fdf))
    return written


def main():
    conn = db.connect(os.environ["DATABASE_URL"])
    n = run(conn)
    print(f"wrote {n} price_forecast rows")


if __name__ == "__main__":
    main()
