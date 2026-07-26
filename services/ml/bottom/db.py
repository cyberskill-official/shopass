"""Postgres access for the forecast job: read price history, write price_forecast."""
import os
from datetime import date, datetime, timezone

import pandas as pd
import psycopg2
import psycopg2.extras


def connect(dsn: str | None = None):
    return psycopg2.connect(dsn or os.environ["DATABASE_URL"])


def distinct_products(conn) -> list[int]:
    with conn.cursor() as cur:
        cur.execute("SELECT DISTINCT product_id FROM price_snapshot ORDER BY product_id")
        return [r[0] for r in cur.fetchall()]


def read_history(conn, product_id: int) -> pd.DataFrame:
    with conn.cursor() as cur:
        cur.execute(
            "SELECT ts, price, flash_sale FROM price_snapshot WHERE product_id=%s ORDER BY ts",
            (product_id,),
        )
        rows = cur.fetchall()
    return pd.DataFrame(rows, columns=["ds", "y", "flash_sale"])


_UPSERT = """
INSERT INTO price_forecast
  (product_id, run_date, horizon_day, yhat, yhat_lower, yhat_upper, p_bottom_14d, model_kind, scored_at, model_run_id)
VALUES %s
ON CONFLICT (product_id, run_date, horizon_day) DO UPDATE SET
  yhat = EXCLUDED.yhat, yhat_lower = EXCLUDED.yhat_lower, yhat_upper = EXCLUDED.yhat_upper,
  p_bottom_14d = EXCLUDED.p_bottom_14d, model_kind = EXCLUDED.model_kind, scored_at = EXCLUDED.scored_at,
  model_run_id = EXCLUDED.model_run_id
"""


def upsert_forecasts(conn, rows: list[tuple]) -> int:
    if not rows:
        return 0
    with conn.cursor() as cur:
        psycopg2.extras.execute_values(cur, _UPSERT, rows)
    conn.commit()
    return len(rows)


def df_to_rows(
    product_id: int,
    run_date: date,
    df: pd.DataFrame,
    scored_at: datetime | None = None,
    model_run_id: int | None = None,
) -> list[tuple]:
    scored_at = scored_at or datetime.now(timezone.utc)
    rows = []
    for _, r in df.iterrows():
        rows.append((
            product_id, run_date, int(r["horizon_day"]),
            int(r["yhat"]), int(r["yhat_lower"]), int(r["yhat_upper"]),
            float(r["p_bottom_14d"]), str(r["model_kind"]), scored_at,
            model_run_id,
        ))
    return rows
