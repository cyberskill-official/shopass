#!/bin/sh
# Applies the shared foundation (db/migrations *.up.sql) then each service's
# forward migrations to DATABASE_URL, in dependency order. Runs on real
# TimescaleDB (no shim). Skips golang-migrate reverse (.down.sql) files.
set -eu
DB="${DATABASE_URL:?set DATABASE_URL}"
apply() { echo ">> $1"; psql -v ON_ERROR_STOP=1 "$DB" -f "$1"; }

for f in /migrations/db/*.up.sql; do apply "$f"; done
for d in auth track price ml deal scrape notif comply cart affil bill; do
  for f in /migrations/services/"$d"/migrations/*.sql; do
    [ -e "$f" ] || continue
    case "$f" in *.down.sql) continue ;; esac
    apply "$f"
  done
done
echo "migrations complete"
