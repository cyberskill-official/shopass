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
            CREATE TABLE IF NOT EXISTS model_run (
              id BIGSERIAL PRIMARY KEY,
              version TEXT NOT NULL,
              model_kind TEXT NOT NULL,
              training_window_start DATE,
              training_window_end DATE,
              feature_set_hash TEXT NOT NULL DEFAULT '',
              backtest_mape REAL,
              backtest_hit_rate REAL,
              gate_passed BOOLEAN NOT NULL DEFAULT false,
              gate_reason TEXT NOT NULL DEFAULT '',
              artifact_path TEXT NOT NULL DEFAULT '',
              created_at TIMESTAMPTZ NOT NULL DEFAULT now()
            );
            CREATE TABLE IF NOT EXISTS price_forecast(
              product_id BIGINT NOT NULL, run_date DATE NOT NULL,
              horizon_day SMALLINT NOT NULL CHECK(horizon_day BETWEEN 1 AND 14),
              yhat BIGINT NOT NULL, yhat_lower BIGINT NOT NULL, yhat_upper BIGINT NOT NULL,
              p_bottom_14d REAL NOT NULL CHECK(p_bottom_14d BETWEEN 0 AND 1),
              model_kind TEXT NOT NULL CHECK(model_kind IN ('prophet','category_prior','lgbm')),
              scored_at TIMESTAMPTZ NOT NULL DEFAULT now(),
              model_run_id BIGINT REFERENCES model_run(id),
              PRIMARY KEY(product_id, run_date, horizon_day));
            TRUNCATE price_snapshot CASCADE;
            TRUNCATE price_forecast CASCADE;
            TRUNCATE model_run CASCADE;
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
    rows = [(100, date(2026, 7, 1), h, 100000, 90000, 110000, 0.85, "lgbm", now, None) for h in range(1, 15)]
    assert db.upsert_forecasts(conn, rows) == 14
    cnt, pmax = _pmax(conn, 100)
    assert cnt == 14 and abs(pmax - 0.85) < 1e-6

    rows2 = [(100, date(2026, 7, 1), h, 100000, 90000, 110000, 0.20, "lgbm", now, None) for h in range(1, 15)]
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
    assert n == 14
    with conn.cursor() as cur:
        cur.execute("SELECT DISTINCT model_kind FROM price_forecast WHERE product_id=200")
        kinds = sorted(r[0] for r in cur.fetchall())
        cur.execute("SELECT count(*) FROM model_run")
        runs = cur.fetchone()[0]
        cur.execute("SELECT gate_passed FROM model_run ORDER BY id DESC LIMIT 1")
        passed = cur.fetchone()[0]
    assert kinds == ["category_prior"]
    assert runs == 1
    assert passed is True
    conn.close()


def test_two_consecutive_runs_record_versions():
    conn = db.connect(DSN)
    _setup(conn)
    with conn.cursor() as cur:
        for i, p in enumerate([100000, 99000, 98000, 97000, 96000]):
            cur.execute(
                "INSERT INTO price_snapshot(product_id, ts, price) VALUES (300, %s, %s)",
                (f"2026-06-{10 + i:02d}T00:00:00Z", p),
            )
    conn.commit()
    assert run_forecast.run(conn, date(2026, 7, 1)) == 14
    assert run_forecast.run(conn, date(2026, 7, 2)) == 14
    with conn.cursor() as cur:
        cur.execute("SELECT count(DISTINCT version) FROM model_run")
        assert cur.fetchone()[0] == 2
    conn.close()
