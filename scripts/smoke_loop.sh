#!/usr/bin/env bash
# Smoke test for the core loop on real service processes + a real Postgres:
#   scrape (fake Shopee) -> pricesvc ingest -> price_snapshot -> forecast -> dealsvc -> bottom_alert_log + notif
#
# Prereqs: Go toolchain; a Postgres reachable at DATABASE_URL with the composed
# schema applied (db + track alert_rule + ml price_forecast + deal migrations).
# In CI use the timescale/timescaledb:pg16 service; locally point DATABASE_URL at it.
#
# Usage: DATABASE_URL=postgres://postgres:postgres@localhost:5432/shopass_smoke ./scripts/smoke_loop.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DSN="${DATABASE_URL:?set DATABASE_URL to a Postgres with the composed schema}"
BIN="$(mktemp -d)"
PSQL=(psql "$DSN" -tA)

pids=()
cleanup() { for p in "${pids[@]:-}"; do kill "$p" 2>/dev/null || true; done; }
trap cleanup EXIT

echo "== build service binaries =="
( cd "$ROOT/services/price"  && go build -o "$BIN/pricesvc"  ./cmd/pricesvc )
( cd "$ROOT/services/scrape" && go build -o "$BIN/scrapesvc" ./cmd/scrapesvc )
( cd "$ROOT/services/deal"   && go build -o "$BIN/dealsvc"   ./cmd/dealsvc )

echo "== seed =="
"${PSQL[@]}" >/dev/null <<SQL
TRUNCATE bottom_alert_log, alert_rule, price_forecast, price_snapshot CASCADE;
INSERT INTO platform (id, code, country) VALUES (1,'shopee','VN') ON CONFLICT (id) DO NOTHING;
INSERT INTO app_user (id) VALUES (999) ON CONFLICT (id) DO NOTHING;
INSERT INTO tracked_product (id, platform_id, platform_item_id, first_seen)
  VALUES (100, 1, '555:777', now() - INTERVAL '100 days') ON CONFLICT (id) DO NOTHING;
INSERT INTO alert_rule (user_id, product_id, rule_type, active) VALUES (999,100,'bottom_predicted',true);
SQL

echo "== fake Shopee (:18090) + fake notif (:18091) =="
python3 - <<'PY' & pids+=($!)
import http.server,json
class H(http.server.BaseHTTPRequestHandler):
    def do_GET(s):
        s.send_response(200);s.send_header("Content-Type","application/json");s.end_headers()
        s.wfile.write(json.dumps({"error":0,"data":{"item":{"price":19900000000,"price_before_discount":25000000000,"stock":10,"historical_sold":5}}}).encode())
    def log_message(s,*a):pass
http.server.HTTPServer(("127.0.0.1",18090),H).serve_forever()
PY
python3 - <<'PY' & pids+=($!)
import http.server
class H(http.server.BaseHTTPRequestHandler):
    def do_POST(s):
        s.rfile.read(int(s.headers.get("Content-Length",0)))
        open("/tmp/smoke_notif.log","a").write("1\n")
        s.send_response(200);s.end_headers();s.wfile.write(b"ok")
    def log_message(s,*a):pass
http.server.HTTPServer(("127.0.0.1",18091),H).serve_forever()
PY
: > /tmp/smoke_notif.log

echo "== pricesvc (:18081) =="
DATABASE_URL="$DSN" PRICE_ADDR=":18081" "$BIN/pricesvc" >/tmp/smoke_pricesvc.log 2>&1 & pids+=($!)
sleep 2

echo "== STEP 1: scrape -> price ingest =="
SHOPEE_BASE_URL="http://127.0.0.1:18090" PRICE_BASE_URL="http://127.0.0.1:18081" SCRAPE_SEED="100:555:777" "$BIN/scrapesvc"
snap="$("${PSQL[@]}" -c "SELECT price FROM price_snapshot WHERE product_id=100")"
echo "  price_snapshot(100).price = ${snap:-NONE}"

echo "== STEP 2: forecast (represents ml output; Prophet fit needs CmdStan) =="
"${PSQL[@]}" >/dev/null <<SQL
INSERT INTO price_forecast (product_id, run_date, horizon_day, yhat, yhat_lower, yhat_upper, p_bottom_14d, model_kind, scored_at)
VALUES (100, current_date, 14, 100000, 90000, 110000, 0.85, 'lgbm', now());
SQL

echo "== STEP 3: dealsvc RUN_ONCE -> alert =="
DATABASE_URL="$DSN" NOTIFSVC_URL="http://127.0.0.1:18091/notify" RUN_ONCE=1 "$BIN/dealsvc"
alerts="$("${PSQL[@]}" -c "SELECT count(*) FROM bottom_alert_log WHERE user_id=999 AND product_id=100")"
notifs="$(wc -l < /tmp/smoke_notif.log | tr -d ' ')"

echo "== RESULT =="
if [ "${snap:-}" = "199000" ] && [ "${alerts:-0}" -ge 1 ] && [ "${notifs:-0}" -ge 1 ]; then
  echo "  PASS: scraped 199000 -> price_snapshot -> forecast -> alert fired (bottom_alert_log=$alerts, notif=$notifs)"
else
  echo "  FAIL: snap=$snap alerts=$alerts notifs=$notifs"; exit 1
fi
