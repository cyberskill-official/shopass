import os
from datetime import date, datetime, timezone

import pytest

from bottom import db, run_forecast

DSN = os.environ.get("TEST_DB_URL")
pytestmark = pytest.mark.skipif(not DSN, reason="TEST_DB_URL not set")


def _setup(conn):
    with conn.cursor() as cur:
        cur.execute(
            """
            CREATE TABLE IF NOT EXISTS price_snapshot(
              product_id BIGINT NOT NULL, ts TIMESTAMPTZ NOT NULL, price BIGINT NOT NULL,
              list_price BIGINT, stock INT, sold INT, flash_sale BOOLEAN NOT NULL DEFAULT false,
              PRIMARY KEY(product_id, ts));
            CREATE TABLE IF NOT EXISTS price_forecast(
              product_id BIGINT NOT NULL, run_date DATE NOT NULL,
              horizon_day SMALLINT NOT NULL CHECK(horizon_day BETWEEN 1 AND 14),
              yhat BIGINT NOT NULL, yhat_lower BIGINT NOT NULL, yhat_upper BIGINT NOT NULL,
              p_bottom_14d REAL NOT NULL CHECK(p_bottom_14d BETWEEN 0 AND 1),
              model_kind TEXT NOT NULL CHECK(model_kind IN ('prophet','category_prior','lgbm')),
              scored_at TIMESTAMPTZ NOT NULL DEFAULT now(),
              PRIMARY KEY(product_id, run_date, horizon_day));
            TRUNCATE price_snapshot;
            TRUNCATE price_forecast;
            """
        )
    conn.commit()


def _pmax(conn, pid):
    with conn.cursor() as cur:
        cur.execute("SELECT count(*), max(p_bottom_14d) FROM price_forecast WHERE product_id=%s", (pid,))
        return cur.fetchone()


def test_upsert_writes_alertable_and_is_idempotent():
    conn = db.connect(DSN)
    _setup(conn)
    now = datetime.now(timezone.utc)
    rows = [(100, date(2026, 7, 1), h, 100000, 90000, 110000, 0.85, "lgbm", now) for h in range(1, 15)]
    assert db.upsert_forecasts(conn, rows) == 14
    cnt, pmax = _pmax(conn, 100)
    assert cnt == 14 and abs(pmax - 0.85) < 1e-6  # alertable forecast (deal fires when > 0.7)

    # re-run same PK with a changed probability -> updated in place, still 14 rows
    rows2 = [(100, date(2026, 7, 1), h, 100000, 90000, 110000, 0.20, "lgbm", now) for h in range(1, 15)]
    db.upsert_forecasts(conn, rows2)
    cnt, pmax = _pmax(conn, 100)
    assert cnt == 14 and abs(pmax - 0.20) < 1e-6
    conn.close()


def test_run_reads_history_and_writes_forecast():
    conn = db.connect(DSN)
    _setup(conn)
    with conn.cursor() as cur:
        for i, p in enumerate([100000, 99000, 98000, 97000]):
            cur.execute(
                "INSERT INTO price_snapshot(product_id, ts, price) VALUES (200, %s, %s)",
                (f"2026-06-{10 + i:02d}T00:00:00Z", p),
            )
    conn.commit()

    n = run_forecast.run(conn, date(2026, 7, 1))
    assert n == 14  # 14 horizon rows for product 200
    with conn.cursor() as cur:
        cur.execute("SELECT DISTINCT model_kind FROM price_forecast WHERE product_id=200")
        kinds = sorted(r[0] for r in cur.fetchall())
    assert kinds == ["category_prior"]  # short history -> cold-start prior (CmdStan-free)
    conn.close()
