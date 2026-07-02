# SănDeal deploy stack

A Docker Compose stack that runs the core value loop on real services and a real TimescaleDB: scrape a price, store it, forecast the bottom, fire the alert, notify. Requires Docker with the Compose plugin.

## Bring it up

```
docker compose -f deploy/docker-compose.yml up -d --build
```

This starts:

- `db` - Postgres 16 + TimescaleDB
- `migrate` - one-shot; applies the shared foundation and every service's forward migrations (see `migrate.sh`), then exits
- `notifsvc` - notification receiver on `:8082/notify`
- `pricesvc` - price read + ingest API on `:8081`
- `dealsvc` - nightly bottom-price scorer (cron, 02:00 Asia/Ho_Chi_Minh)
- `web` - Next.js UI on `:3000`

Check: `docker compose -f deploy/docker-compose.yml ps` and open `http://localhost:3000`.

## Run the loop (one-shot jobs)

The scraper and the forecast job are under the `jobs` profile, run on demand:

```
# 1. seed a product to track (the scraper writes price_snapshot for it)
docker compose -f deploy/docker-compose.yml exec db psql -U postgres -d shopass -c \
  "INSERT INTO app_user(id) VALUES (999) ON CONFLICT DO NOTHING; \
   INSERT INTO tracked_product(id, platform_id, platform_item_id, first_seen) \
     VALUES (100, 1, '555:777', now() - INTERVAL '100 days') ON CONFLICT DO NOTHING; \
   INSERT INTO alert_rule(user_id, product_id, rule_type, active) VALUES (999,100,'bottom_predicted',true);"

# 2. scrape -> price ingest (SHOPEE_BASE_URL can point at a fixture for a dry run)
SCRAPE_SEED=100:555:777 docker compose -f deploy/docker-compose.yml run --rm scrapesvc

# 3. forecast -> price_forecast (Prophet for mature SKUs; needs enough history)
docker compose -f deploy/docker-compose.yml run --rm mlforecast

# 4. fire the nightly score once (instead of waiting for cron)
docker compose -f deploy/docker-compose.yml run --rm -e RUN_ONCE=1 dealsvc
```

Then inspect `price_snapshot`, `price_forecast`, and `bottom_alert_log` in `db`.

## Acceptance check

`scripts/smoke_loop.sh` walks one product from a scraped price to a fired alert and asserts the result. Point it at the compose database:

```
DATABASE_URL=postgres://postgres:postgres@localhost:5432/shopass?sslmode=disable ./scripts/smoke_loop.sh
```

## Notes

- All services share one database (`shopass`); price owns `price_snapshot`, ml owns `price_forecast`, deal reads both. `platform` is seeded by its migration (shopee=1, tiktok=2, lazada=3).
- The Go build image is `golang:1.25` (matches `go.mod`); bump it if you align to 1.22 per the hygiene note in `docs/AUDIT-REPORT.md`.
- `mlforecast`'s image installs CmdStan so the mature-SKU Prophet fit works.
- Not yet in the stack: the durable scrape queue/scheduler (FR-SCRAPE-001), a cron for the scraper and forecast jobs, and the API gateway. Add them once the loop is exercised.
