# Database migrations

## Two runners (reconcile, do not fork)

This repository has two forward-migration paths that apply the **same** SQL
under `db/migrations/*.up.sql` (plus service-local files under
`services/*/migrations/` for the compose ledger):

| Runner | Entry | Schema tracking | Down |
| --- | --- | --- | --- |
| Compose / ops | `deploy/migrate.sh` | `applied_migration_files` (path key) | Skips `*.down.sql` |
| Go library / tests | `db/internal/migrate` (golang-migrate) | `schema_migrations` (version) | Supports `Down` / `Steps` |

**Rules so dual-runner stays coherent:**

1. Add every shared schema change as a new numbered pair
   `NNNN_name.up.sql` + `NNNN_name.down.sql` in `db/migrations/`.
2. Never edit an already-shipped `NNNN_*.up.sql`; ship a new forward migration.
3. Prefer idempotent DDL (`IF NOT EXISTS` / `IF EXISTS`) when a change may be
   applied by either runner on DBs that already received a partial fix.
4. Service-only migrations stay under `services/<svc>/migrations/` and are
   applied only by `deploy/migrate.sh` (not by golang-migrate on `db/migrations`).
5. Do not auto-baseline an existing DB in `migrate.sh` without the explicit
   `MIGRATION_BASELINE=acknowledge-existing-schema` acknowledgement.

When debugging "migration applied in compose but not in Go tests" (or the
reverse), compare both ledgers against the files on disk rather than inventing
a third runner.
