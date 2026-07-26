"""Persist model_run rows and resolve trailing best MAPE (R26)."""
from __future__ import annotations

import hashlib
import json
import os
from datetime import date, datetime, timezone
from pathlib import Path

FEATURE_SET = ("is_double_date", "is_payday_window", "flash_sale", "y", "ds")


def feature_set_hash() -> str:
    payload = json.dumps(sorted(FEATURE_SET), separators=(",", ":"))
    return hashlib.sha256(payload.encode()).hexdigest()[:16]


def artifact_dir(version: str) -> Path:
    root = Path(os.environ.get("MODEL_ARTIFACT_DIR", "/tmp/shopass-model-artifacts"))
    path = root / version
    path.mkdir(parents=True, exist_ok=True)
    return path


def next_version(run_date: date, model_kind: str) -> str:
    stamp = datetime.now(timezone.utc).strftime("%H%M%S")
    return f"{run_date.isoformat()}-{model_kind}-{stamp}"


def best_trailing_mape(conn, model_kind: str, *, days: int = 30) -> float | None:
    with conn.cursor() as cur:
        cur.execute(
            """
            SELECT MIN(backtest_mape)
            FROM model_run
            WHERE model_kind = %s
              AND gate_passed = true
              AND backtest_mape IS NOT NULL
              AND created_at >= now() - (%s * INTERVAL '1 day')
            """,
            (model_kind, days),
        )
        row = cur.fetchone()
    if not row or row[0] is None:
        return None
    return float(row[0])


def insert_model_run(
    conn,
    *,
    version: str,
    model_kind: str,
    training_window_start: date | None,
    training_window_end: date | None,
    feature_hash: str,
    backtest_mape: float | None,
    backtest_hit_rate: float | None,
    gate_passed: bool,
    gate_reason: str,
    artifact_path: str,
) -> int:
    with conn.cursor() as cur:
        cur.execute(
            """
            INSERT INTO model_run (
              version, model_kind, training_window_start, training_window_end,
              feature_set_hash, backtest_mape, backtest_hit_rate,
              gate_passed, gate_reason, artifact_path
            ) VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)
            RETURNING id
            """,
            (
                version,
                model_kind,
                training_window_start,
                training_window_end,
                feature_hash,
                backtest_mape,
                backtest_hit_rate,
                gate_passed,
                gate_reason,
                artifact_path,
            ),
        )
        run_id = int(cur.fetchone()[0])
    conn.commit()
    return run_id


def write_artifact_stub(path: Path, meta: dict) -> None:
    path.write_text(json.dumps(meta, indent=2, default=str) + "\n", encoding="utf-8")
