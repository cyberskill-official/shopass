# Naming Conventions and Migrations (FR-INFRA-002)

## Naming Conventions
- Table names must be `snake_case`, singular (e.g., `app_user`, not `app_users`).
- Primary key column must be named `id`.
- For large tables or business entities, `id` must be `BIGSERIAL` or `BIGINT`.
- Small, static tables (like `platform`) can use `SMALLINT`.
- Timestamp columns must have the `_at` suffix and use type `TIMESTAMPTZ DEFAULT now()`.
- Email columns should use `CITEXT` to enforce case-insensitive uniqueness.

## Migration Rules
- Use `golang-migrate`.
- Each migration consists of a sequential pair of files: `NNNN_name.up.sql` and `NNNN_name.down.sql`.
- A migration file is immutable once merged. DO NOT modify existing files; add a new migration instead.
- `0001_extensions.up.sql` handles extensions (e.g., `CITEXT`) and must run before any tables depend on it.
