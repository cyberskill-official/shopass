# obs — Shopass observability contracts (R13)

Shared Go helpers: Prometheus metrics (`HTTPObserve`, `MetricsHandler`), OTel tracing, structured logging helpers.

## SLO table (NFR → alerts)

| NFR | Signal | Alert rule | Threshold |
|-----|--------|------------|-----------|
| NFR-INFRA-001 | `http_request_duration_ms` at gateway | `ShopassCacheP95High` | p95 > 300ms for 5m |
| NFR-INFRA-001 | chart/price routes | `ShopassChartP95High` | p95 > 500ms for 5m |
| NFR-INFRA-002 | `http_requests_total` 5xx ratio | `ShopassErrorRateHigh` | > 1% for 10m |
| Availability | blackbox `probe_success` | `ShopassServiceDown` | fail for 2m |
| R17 scrape | `shopass_job_last_success_unixtime{job_name="scrape"}` | `ShopassScrapeJobStale` | age > 450s |
| R17 forecast | `shopass_job_last_success_unixtime{job_name="forecast"}` | `ShopassForecastJobStale` | age > 36h |
| R26 model gate | `shopass_ml_gate_trips_total` (Pushgateway) | `ShopassModelGateTripped` | any increase in 1h |

Rules live in `deploy/prometheus/rules/shopass.yml`. Bring up the stack with `deploy/docker-compose.observability.yml` (profile `observability`).

## Metrics endpoints today

| Service | Path |
|---------|------|
| gateway | `:9090/metrics` (separate listener; host map `GATEWAY_METRICS_PORT`, default 9094) |
| authsvc | `:8084/metrics` |

Other core services are covered by blackbox `/healthz` probes until they grow process metrics exporters.

## Telegram alerts

Alertmanager ships with a noop webhook receiver. To deliver to Telegram, replace the `telegram` receiver in `deploy/alertmanager/alertmanager.yml` on the host (or overlay) with a Telegram bot integration and set bot token + chat id via host secrets. Record the credential ask under improvement R13 in `docs/tasks/improvement/LEDGER.md`.
