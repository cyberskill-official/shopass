# SănDeal - audit and rework report

Date: 2026-07-03. Scope: deep audit and rework of the implementation that a prior model produced against the 90-FR backlog, driving unit, integration, and e2e tests to green.

## Verdict

The implementation compiled and its unit tests "passed" only on paper - several packages did not compile, tests were drifted from the code they tested, and the JS and Python test harnesses were misconfigured. After the rework below, every actual code and test defect is fixed. What does not run in this sandbox is gated purely on infrastructure the sandbox cannot provide (Docker, TimescaleDB, CmdStan), not on code.

Core invariants were respected by the original code: money is BIGINT/int64 everywhere (no float in money fields or columns), and platform session tokens are actively guarded and redacted so they never leave the client. The bugs were mechanical, not architectural.

## How this was verified

A real toolchain was stood up in the sandbox rather than reasoning about the code:

- Go 1.26 (the modules declare `go 1.25.0`; see hygiene note).
- PostgreSQL 18 (rootless, via conda) for the `deal` integration suite.
- A Babel-based jest harness installed with pnpm, off a working copy, for the extension and web suites.
- A Python venv with Prophet + LightGBM for the ml suite.

## Green status (measured, not asserted)

| Area | Result |
|---|---|
| Go - db, obs, region, secrets | build + unit green (db migration tests are Docker-gated, see below) |
| Go - 11 services | build + vet + unit green |
| Go - deal integration (`deal/batch`) | 6 tests green against live Postgres |
| Extension (MV3) | tsc clean + 87 jest tests across 34 suites green |
| Web (Next.js) | tsc clean + 23 jest tests green |
| ml | 12 of 13 pytest green (1 needs CmdStan, see below) |

## Defects found and fixed

Compile breaks in production code (the code was never compiled):

- `deal/internal/deal/service.go` referenced `fakesale.VerdictUndefined`, which does not exist. Fixed to the real `Unknown` verdict.
- `bill/internal/api/ipn.go` used `time.Duration` without importing `time`. Added the import.

Test-versus-implementation drift (the tests were never run):

- `bill` - `NewReconcileJob` (2 args vs 3) and `NewIPNHandler` (2 vs 3); added the missing `SubscriptionActivator` mocks.
- `comply/ecom` - the tests wanted a seedable in-memory repo but the impl was Postgres-only. Introduced a `store` interface: production keeps the pgx `*Repo`, the unit test uses an in-memory fake seeded to match migration `0008`.
- `cart/api` - the handler correctly reads the gateway-injected `X-User-Id` header (TASK-CART-002), but the tests passed the user via a request context value that cannot cross an HTTP boundary. Fixed the tests to use the header.

Integration-test defects (`deal/batch`):

- The test assumed a schema that did not exist: a phantom `url` column on `tracked_product`, wrong `price_forecast` columns (`as_of_date`, `expected_min_14d` vs the real `run_date`, `horizon_day`, `yhat*`), missing FK seed rows for `platform` and `app_user`, an invalid `rule_type` enum value (`price_drop`), and a 1e-9 float tolerance on a `real` (float32) column. All corrected to the real schema and realistic tolerances.

TikTok adapter (`scrape`):

- `Fetch` launched a real Playwright browser and then returned a hardcoded price, so the unit test could not run without a browser. Added a `Renderer` seam: production uses Playwright, the unit test injects a fake - matching the shopee adapter's dependency-injection pattern.

ml portability bug:

- `bottom/prophet_baseline.py` called `cmdstanpy.set_cmdstan_path("~/.cmdstan/cmdstan-2.39.0")` at import time, which crashed every consumer of the `bottom` package on any machine lacking that exact path. Guarded it to use a system CmdStan only when present, otherwise fall back to Prophet's bundled backend.

Frontend test harness:

- Extension - jest 30 + ts-jest 29 + TypeScript 6 is unworkable, and `package.json` was corrupted (its `dependencies` was the entire flattened `node_modules` tree). Replaced ts-jest with Babel (tsc still type-checks separately) and cleaned `package.json` to real dependencies.

## Environment-gated tests (pass elsewhere, not in this sandbox)

These are correct tests that need infrastructure the sandbox cannot provide. They run on a developer machine or CI:

- `db/internal/migrate` (3 tests) - testcontainers spinning up `timescale/timescaledb:latest-pg16`. Needs Docker (none in the sandbox). This is the proper migration test and also exercises the TimescaleDB hypertable path.
- `ml test_prophet_forecast_shape` (1 test) - needs a CmdStan backend to fit a Prophet model. Stephen's Mac has `~/.cmdstan/cmdstan-2.39.0` (which the old hardcoded path pointed at).

## Repo hygiene (recommended, not yet applied to git)

- A full virtualenv (`services/ml/.venv`) and `audit/node_modules` are committed. They inflate the repo and are platform-specific (the committed `node_modules` are a macOS build). Add a `.gitignore` and untrack them:

```
git rm -r --cached services/ml/.venv audit/node_modules extension/node_modules web/node_modules
```

- ~~`go.mod` files declare `go 1.25.0`; SHIP-GUIDE mandates 1.22.~~ Resolved by R10 (2026-07-26): SHIP-GUIDE + CI + Dockerfile + `toolchain go1.25.12` aligned; govulncheck in CI.

## How to run each suite

```
# Go (unit)
cd services/<svc> && go test ./...

# Go deal integration (needs a Postgres at TEST_DB_URL with the composed schema)
export TEST_DB_URL="postgres://postgres:postgres@127.0.0.1:5432/shopass_deal_test?sslmode=disable"
cd services/deal && go test ./internal/batch/

# db migration tests (needs Docker)
cd db && go test ./...

# Extension
cd extension && npm test        # after: npm install (Babel-based jest)

# Web
cd web && npm test

# ml (needs prophet + cmdstan for the full suite)
cd services/ml && pytest tests/
```

## Feature build (2026-07-03) - the three P1 gaps + Lazada

After the rework, four features were built and verified with the same real toolchain. Details and per-feature test counts are in `docs/TASK-COVERAGE.md`; this is how to run them.

```
# TASK-PRICE-004 cross-platform compare (5 integration tests) - part of pricesvc
cd services/price && TEST_DB_URL="postgres://.../shopass_price_test?sslmode=disable" go test -p 1 ./...

# Lazada: Go orchestrator adapter now dispatches to the farm (4 unit tests)
cd services/scrape && go test ./internal/adapters/lazada/...
# Lazada/TikTok farm extraction (TypeScript): tsc --noEmit clean; extraction logic verified.
# The browser-backed farm adapter.test.ts needs Playwright Chromium (not in this sandbox).

# TASK-WEB-005 GraphQL BFF (10 tests on Node's built-in runner)
cd services/bff && npm install && npm test        # runs tsc then node --test

# TASK-AUTH-004 social login core (auth unit + Postgres integration)
cd services/auth && TEST_DB_URL="postgres://.../shopass_auth_test?sslmode=disable" go test -p 1 ./...
```

Two honesty notes carried over from the sandbox limits above:

- Docker itself is still not run here (there is no Docker in the sandbox). The `deploy/migrate.sh` chain, including the new `0007_social_identity` migration, was replayed against a real Postgres and applies cleanly; a `docker compose up` has not been exercised.
- The compare endpoint (in `pricesvc`) and the Lazada change (in `scrapesvc`) ride compose services that already exist. The BFF and the auth service are built and tested but are not yet in `docker-compose.yml` - wiring them needs the API gateway and the track service, which are also not in the core stack. `make migrate` does create `social_identity`.
