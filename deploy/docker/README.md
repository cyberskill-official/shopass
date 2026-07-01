# Local Development — Docker Stack

Placeholder documentation for the local development infrastructure.

## Required Services

| Service | Purpose | Image | Notes |
|---------|---------|-------|-------|
| PostgreSQL 16 | Primary OLTP database | `postgres:16` | Port 5432 |
| TimescaleDB 2.x | Time-series price data | `timescale/timescaledb:2-pg16` | Hypertable for `price_snapshot` |
| Redis | Queue / fan-out | `redis:7` | Kafka-Redis Streams alternative |
| Vault / AWS Secrets Manager | Secrets storage | `hashicorp/vault:latest` | Dev mode for local; production uses AWS SM |
| OpenTelemetry Collector | Trace/metric/log pipeline | `otel/opentelemetry-collector-contrib` | Export to Prometheus/Grafana |
| Prometheus | Metrics store | `prom/prometheus` | Scrape OTel endpoints |
| Grafana | Dashboards | `grafana/grafana` | Provisioned dashboards per FR |

## Docker Compose

A `docker-compose.yml` will be created here when the first service FR (FR-INFRA-001/002) is implemented.

## Usage

```bash
# Start all local dependencies
docker compose up -d

# Run migrations
# (command TBD by FR-INFRA-002)

# Stop
docker compose down
```
