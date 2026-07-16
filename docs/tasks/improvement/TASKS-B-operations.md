# Part B - production operations (R11-R22)

---

## R11 - TLS + reverse proxy (Caddy)

Wave 1 | Effort M | Depends: R1 | Stephen input: creds (domain DNS access)

Why: nothing terminates TLS; services speak plain HTTP on published ports. The cyberos repo already runs the Caddy-in-front pattern in production - reuse it.

Steps:
1. Copy the Caddy service pattern from `cyberos/deploy` (compose service + Caddyfile with automatic certs).
2. Route: `app.<domain>` -> web, `api.<domain>` -> gateway (R1). Redirect HTTP->HTTPS. Add security headers (HSTS, X-Content-Type-Options, frame-ancestors).
3. Parameterize domain via `deploy/.env.example`.
4. Local story: self-signed/`localhost` config documented in `deploy/README.md`.
5. Ask Stephen: point DNS A records for the chosen domain (R8) at the VPS.

Acceptance: HTTPS with valid cert on both hosts once DNS exists; HTTP redirects; headers present.

Verify: `curl -sI https://api.<domain>/healthz` output; SSL Labs or `testssl.sh` summary in ledger.

Human review: check HSTS max-age is sane (start low, e.g., 1 day, raise later); confirm no service still reachable over plain HTTP from outside.

---

## R12 - automated backups + restore drill

Wave 1 | Effort M | Depends: - | Stephen input: account + creds (S3-compatible storage)

Why: `deploy/README.md:74` documents only a hand-run pg_dump. Price history is the moat; losing the DB resets the company clock to zero.

Steps:
1. Add a backup sidecar (cron container or host cron): nightly `pg_dump | gzip` -> S3-compatible bucket (Backblaze/Wasabi/Vultr object storage), 30-day retention, filename with date + git sha of schema.
2. Add wal-g (or pgBackRest) for WAL archiving + weekly base backup -> point-in-time recovery. If too heavy for wave 1, ship nightly dumps now and log PITR as follow-up.
3. Alert on backup failure or staleness > 26h (wire into R13's Prometheus or a dead-man's-snitch style ping).
4. Write `deploy/RESTORE-RUNBOOK.md`: exact restore commands, expected RTO/RPO (propose RPO 24h now, 5min with WAL; RTO 1h).
5. Run one full restore drill into a scratch container; record timings.

Acceptance: backup object appears in bucket on schedule; drill restored a working DB; runbook exists.

Verify: bucket listing, drill transcript, and row-count comparison in ledger.

Human review: Stephen provisions the bucket + keys; reviewer re-runs the restore drill once personally.

---

## R13 - Prometheus + alert rules for declared NFRs

Wave 1 | Effort M | Depends: - | Stephen input: -

Why: OTel instrumentation and `deploy/grafana/dashboards/overview.json` exist, but no Prometheus service scrapes anything, and no alert fires on anything. NFR-INFRA-001/002 targets (p95 <300ms cache, <500ms chart, 99.5% uptime) are unverifiable.

Steps:
1. Add prometheus service to compose; expose/confirm each Go service's metrics endpoint (check how the OTel meter exports - Prometheus exporter or OTLP; add the exporter if missing).
2. Wire Grafana (add service if not present) with the existing dashboard JSON provisioned.
3. Alert rules: p95 latency per NFR, error-rate burn (fast+slow window per NFR-INFRA-002), scrape job success rate, alert-dispatch failures, FCM 429 count, backup staleness (R12), disk usage.
4. Alertmanager route -> Telegram (cheapest) for now.
5. Document the SLO table in `obs/README.md` or `deploy/README.md`.

Acceptance: Grafana shows live metrics from all core services; a forced error triggers a Telegram alert.

Verify: screenshot/exported panel JSON + alert test transcript in ledger.

Human review: confirm the alert actually reached your Telegram; sanity-check thresholds against the NFR docs.

---

## R14 - centralized logs (Loki)

Wave 1 | Effort M | Depends: - | Stephen input: -

Why: logs live per-container; incident work means ssh + `docker logs` per service.

Steps:
1. Add Loki + promtail (or Vector) to compose, scraping container logs with service labels.
2. Grafana datasource + one "errors across services" saved query.
3. Confirm Go services log structured JSON (they use slog-style in dealsvc; verify others) so fields parse; fix stragglers.
4. Retention: 14-30 days local volume; document.

Acceptance: one Grafana Explore query shows interleaved logs from gateway, pricesvc, dealsvc, notifsvc with labels.

Verify: query screenshot in ledger.

Human review: search for a known smoke-test log line end to end.

---

## R15 - CI/CD: GHCR image publish + SSH deploy

Wave 1 | Effort M | Depends: R2 | Stephen input: creds (VPS host, SSH key, GHCR perms)

Why: CI tests but ships nothing; deploys are manual. The cyberos `deploy.yml` (gate -> GHCR -> SSH roll) is a proven in-house pattern - port it.

Steps:
1. Extend `.github/workflows/`: on main merge after tests, build + push images for gateway, pricesvc, dealsvc, notifsvc, authsvc, tracksvc, bff, web, scrapesvc, mlforecast to GHCR with sha + latest tags.
2. Deploy job (environment-gated): ssh to VPS, `docker compose pull && docker compose up -d --remove-orphans`, then run the smoke check (R22's script, or curl healthz set for now).
3. Compose on the server references GHCR images (a `deploy/docker-compose.prod.yml` overlay) instead of local builds.
4. Ask Stephen: VPS choice/creation, SSH deploy key as repo secret, GHCR token if org-scoped.

Acceptance: merge to main produces images and a deployed stack without manual steps; failed smoke blocks the roll (or alerts loudly).

Verify: workflow run URL, `docker compose images` output from server in ledger.

Human review: Stephen adds secrets; reviewer approves the first environment-gated deploy manually.

---

## R16 - zero-downtime deploys + migration guards

Wave 1 | Effort M | Depends: R15 | Stephen input: -

Why: only db has a healthcheck; nothing gates a roll on service health; migrations lack `IF NOT EXISTS` discipline, so a bad schema change can strand running pods mid-deploy.

Steps:
1. Add healthchecks to every service in compose (`/healthz` exists on Go services - verify each; add where missing) and `depends_on: condition: service_healthy` chains.
2. Verify SIGTERM graceful shutdown in each Go service (http server Shutdown with timeout; drain cron in dealsvc); add where missing, with tests where practical.
3. Migration lint script in CI: every new `*.up.sql` must be additive or guarded (`IF NOT EXISTS`, `ADD COLUMN ... DEFAULT` safe patterns); forbid `DROP`/`ALTER TYPE` without a `-- reviewed:` tag.
4. Document the forward-only rule + expand/contract pattern in `docs/conventions/`.

Acceptance: `docker compose up -d` rolls services one by one without failed requests during smoke; migration lint blocks a test-violation PR.

Verify: rolling restart transcript while a curl loop runs (error count 0) in ledger.

Human review: read the migration lint rules; try the curl-loop restart once.

---

## R17 - production scheduling for scrape + forecast jobs

Wave 1 | Effort S | Depends: - | Stephen input: -

Why: dealsvc self-schedules (robfig cron, `services/deal/cmd/dealsvc/main.go:122`), but scrapesvc and mlforecast are compose `profiles: ["jobs"]` one-shots - nothing runs them in prod, so "self-running" is false without a human.

Steps:
1. Pick the pattern and document it: host systemd timers (recommended: visible, survives compose restarts) calling `docker compose run --rm scrapesvc` on the tier cadence and `mlforecast` nightly after scrape; OR long-running mode flags inside the services mirroring dealsvc.
2. Implement (systemd unit files in `deploy/systemd/` + install notes in `deploy/README.md`, wired by R15's deploy).
3. Emit a heartbeat metric per job run; alert (R13) when a scheduled run is missing > 1.5x its period.
4. Ensure overlapping-run safety: the pg queue lease already guards scrape; verify mlforecast is idempotent per (product, run_date).

Acceptance: on the server, both jobs run on schedule for 3 consecutive days with heartbeats visible.

Verify: timer list output, heartbeat graph, job logs in ledger.

Human review: check the missing-run alert by disabling a timer for one cycle.

---

## R18 - wire BFF behind gateway or remove the dead path

Wave 2 | Effort M | Depends: R1 | Stephen input: -

Why: bff is defined in compose but exposes no port and nothing routes to it; web calls services directly. Dead topology confuses agents and hides GraphQL from production reality.

Steps:
1. Decide with evidence: if web uses REST paths today and GraphQL adds no current value, either (a) route gateway -> bff for the read paths web actually needs and migrate web fetches to it, or (b) mark bff experimental, exclude from default compose, and log the decision.
2. If (a): add gateway route, auth header propagation (verified identity from R1), DataLoader caching sanity test, and switch `web/lib` fetchers.
3. Update `docs/TASK-COVERAGE.md` note about the BFF follow-up.

Acceptance: no service defined in compose that nothing can reach; web's data path documented in one diagram (`docs/` or deploy README).

Verify: compose config + a traced request (web -> gateway -> bff -> svc) in ledger.

Human review: approve the (a)/(b) decision; confirm chart page still loads under the new path.

---

## R19 - data retention + chunk policy decision

Wave 2 | Effort S | Depends: - | Stephen input: decision

Why: compression after 30 days exists; retention is undecided. Long history is the product, so deletion must be deliberate, and storage cost must be forecast.

Steps:
1. Measure current growth: rows/day and bytes/day per hypertable from a week of real scraping (post-R24, or fixture-based estimate now).
2. Propose policy: raw `price_snapshot` kept 24 months, `price_daily` continuous aggregate kept forever; chunk interval review (7d now) against write volume.
3. Implement `add_retention_policy` for the raw table only after Stephen signs off; add the policy statements as a migration with the `-- reviewed:` tag (R16).
4. Document in `docs/tasks/DATA-MODEL.md`.

Acceptance: written policy with cost projection; migration applied in staging.

Verify: `SELECT * FROM timescaledb_information.jobs` output in ledger.

Human review: Stephen approves the retention horizon (this is a product decision, defaulting to keep-everything until storage cost says otherwise).

---

## R20 - k6 load test gate vs NFR p95 targets

Wave 3 | Effort M | Depends: R13 | Stephen input: -

Why: NFR-INFRA-001 declares p95 <300ms (cached) and <500ms (chart) with a promised regression gate; no load test exists.

Steps:
1. k6 script: chart endpoint, compare endpoint, wishlist list; ramp to expected launch RPS (derive from PRD's 100k-user math; start 50 RPS mixed).
2. Run against staging weekly via CI cron; export p95s; fail the run if targets breached.
3. Store baseline results in `docs/non-functional-requirements/` next to the NFR doc.

Acceptance: weekly CI job green with recorded p95s under targets (or a triaged issue when not).

Verify: k6 summary + CI link in ledger.

Human review: sanity-check the traffic model against real analytics once R40 has data.

---

## R21 - dependency, image scanning, SBOM

Wave 3 | Effort S | Depends: R15 | Stephen input: -

Why: no automated dependency updates or vulnerability scanning beyond R10's govulncheck; images unscanned.

Steps:
1. Renovate (or Dependabot) config for Go modules, npm (web, extension), pip (ml), grouped weekly.
2. CI: `npm audit --audit-level=high` (web+extension), `pip-audit` (ml), Trivy scan on built images (fail on HIGH+ fixable).
3. SBOM (syft) attached to each release/image push.

Acceptance: first Renovate PRs open; CI blocks a known-vulnerable test dependency; SBOM artifact downloadable.

Verify: CI run links in ledger.

Human review: merge cadence decision for Renovate groups (weekly batch recommended).

---

## R22 - nightly end-to-end smoke in CI

Wave 3 | Effort M | Depends: R15 | Stephen input: -

Why: `make smoke` proves the whole loop (seed -> scrape fixture -> deal -> alert row) but only runs on a laptop. The loop can rot silently.

Steps:
1. GitHub Actions nightly job: compose up the full stack (GH runners support Docker), run `make smoke` with `ALLOW_SEED=1` (R9), assert expected rows, tear down.
2. On failure: open/append a GitHub issue and fire the R13 Telegram alert via webhook.
3. Keep runtime < 10 min (cache images from GHCR - R15).

Acceptance: three consecutive green nightly runs; a deliberately broken fixture turns it red with an alert.

Verify: workflow links + the forced-failure test in ledger.

Human review: confirm alert reached Telegram; spot-check runtime cost.
