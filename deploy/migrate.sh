#!/bin/sh
# Applies the shared foundation (db/migrations *.up.sql) then each service's
# forward migrations to DATABASE_URL, in dependency order. Runs on real
# TimescaleDB (no shim). Skips golang-migrate reverse (.down.sql) files.
set -eu
DB="${DATABASE_URL:?set DATABASE_URL}"

# The migration key includes its source path, not just the basename. Different
# services may legitimately have migrations named 0001_*.sql.
psql -v ON_ERROR_STOP=1 -c "CREATE TABLE IF NOT EXISTS applied_migration_files (filename VARCHAR(255) PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP);" "$DB"

has_platform=$(psql -t -A -c "SELECT CASE WHEN to_regclass('public.platform') IS NULL THEN '' ELSE '1' END;" "$DB")
applied_count=$(psql -t -A -c "SELECT count(*) FROM applied_migration_files;" "$DB")

migration_key() {
  printf '%s' "$1" | sed 's#^/migrations/##'
}

sql_literal() {
  # Migration keys are repository-controlled paths, but quote them anyway so
  # psql receives valid SQL without relying on psql variable substitution.
  # `psql -c` sends SQL directly to the server and does not expand :variables.
  printf '%s' "$1" | sed "s/'/''/g"
}

mark_applied() {
  key=$(migration_key "$1")
  key_sql=$(sql_literal "$key")
  psql -v ON_ERROR_STOP=1 \
    -c "INSERT INTO applied_migration_files (filename) VALUES ('$key_sql') ON CONFLICT DO NOTHING;" "$DB"
}

# A pre-existing database has no reliable way to prove that every migration in
# this repository was applied. Never auto-baseline it: that can silently skip a
# schema change. An operator who has independently verified the database may
# deliberately set this exact acknowledgement once.
if [ "$has_platform" = "1" ] && [ "$applied_count" = "0" ]; then
  if [ "${MIGRATION_BASELINE:-}" != "acknowledge-existing-schema" ]; then
    echo "refusing to auto-baseline an existing database; verify its schema and set MIGRATION_BASELINE=acknowledge-existing-schema deliberately" >&2
    exit 1
  fi
  echo "baselining an independently verified existing database"
  for f in /migrations/db/*.up.sql; do mark_applied "$f"; done
  for d in auth track price ml deal scrape notif comply cart affil bill; do
    for f in /migrations/services/"$d"/migrations/*.sql; do
      [ -e "$f" ] || continue
      case "$f" in *.down.sql) continue ;; esac
      mark_applied "$f"
    done
  done
fi

apply() {
  key=$(migration_key "$1")
  key_sql=$(sql_literal "$key")
  already_applied=$(psql -t -A -c "SELECT 1 FROM applied_migration_files WHERE filename = '$key_sql';" "$DB")
  if [ "$already_applied" = "1" ]; then
    return
  fi
  echo ">> $1"
  psql -v ON_ERROR_STOP=1 "$DB" -f "$1"
  mark_applied "$1"
}

for f in /migrations/db/*.up.sql; do apply "$f"; done
for d in auth track price ml deal scrape notif comply cart affil bill; do
  for f in /migrations/services/"$d"/migrations/*.sql; do
    [ -e "$f" ] || continue
    case "$f" in *.down.sql) continue ;; esac
    apply "$f"
  done
done
echo "migrations complete"
