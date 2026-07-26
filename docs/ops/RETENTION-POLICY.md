# Data retention + chunk policy (R19)

**Status:** proposed for founder sign-off (migration already present in repo)  
**Owner:** Stephen Cheng  
**Related:** `docs/tasks/DATA-MODEL.md`, `db/migrations/0020_compression_policy.up.sql`, NFR-PRICE-001

## Decision proposal

| Store | Policy | Rationale |
|-------|--------|-----------|
| `price_snapshot` (raw) | **Retain 18 months** via Timescale `add_retention_policy` | Sale-ảo / chart need ~90d raw fidelity; older raw is low value vs storage |
| `price_daily` (cagg) | **Retain indefinitely** | Long history is the product; bottom prediction needs ≥180d |
| Compression | Compress raw chunks **after 30 days** (`compress_segmentby=product_id`) | Matches DEC-PRICE / unit economics |
| Chunk interval | **7 days** | Balance chunk count vs compress/retention granularity |

This matches the product DATA-MODEL and TASK-PRICE-002. R19’s job is to make the choice **explicit and costed**, not invent a new horizon.

## Already in migrations

`db/migrations/0020_compression_policy.up.sql`:

```sql
SELECT add_compression_policy('price_snapshot', INTERVAL '30 days');
SELECT add_retention_policy('price_snapshot', INTERVAL '18 months');
```

Tag for ops: treat production enablement of this migration as **`-- reviewed: Stephen`** once signed below. Do not shorten retention without a new decision.

## Fixture-based growth estimate (pre-R24)

Until live scrape volume exists, use PRD-scale math:

| Assumption | Value |
|------------|-------|
| Tracked SKUs (year-1 ambition) | 1e6 |
| Delta-only writes / SKU / day | ~2 (90% no-change skip) |
| Raw row size (compressed later) | ~40–80 B uncompressed columnar estimate |
| Raw rows / day | ~2e6 |
| Raw bytes / day (pre-compress) | ~80–160 MB |
| 18-month raw retention | ~45–90 GB uncompressed-order; compress ≥8x target → **~6–12 GB** class |
| `price_daily` forever | ~1 row/SKU/day → year-1 ~0.4 GB order; grows linearly with SKU×days |

**Unit economics check:** NFR-PRICE-001 targets ~0.1–0.2 USD/user/month storage; retention at 18 months raw + forever daily is consistent if delta-only + compression hold.

After R24 has a week of real scrape: replace this table with measured `rows/day` and `pg_total_relation_size`.

## Chunk review

Current: `chunk_time_interval = 7 days`. Revisit if write volume makes 7d chunks &gt; few GB uncompressed (then consider 1d) or if chunk count explodes (then 14d). No change proposed until metrics exist.

## Stephen sign-off

- [ ] Approve **18 months** raw / **forever** daily (or write alternate horizons)
- [ ] Confirm production has applied `0020_compression_policy` (check `timescaledb_information.jobs`)
- [ ] Optional: shorten/lengthen after first storage bill

**Recommendation:** approve as written; revisit at first 100k tracked SKUs or when object/DB storage exceeds budget.
