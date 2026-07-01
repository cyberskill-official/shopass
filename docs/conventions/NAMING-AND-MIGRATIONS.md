# Naming and Migration Conventions — SănDeal

## Table Naming
- snake_case, plural: `app_user`, `price_snapshot`, `tracked_product`
- Primary key always named `id`
- Timestamp columns suffixed `_at` with type `TIMESTAMPTZ DEFAULT now()`
- Foreign key columns: `<referenced_table>_id` (e.g. `user_id`, `product_id`)

## Data Types
- Primary keys: `BIGSERIAL` / `BIGINT` for tables that can grow; `SMALLINT` for small static tables (e.g. `platform`)
- Money: `BIGINT` (VND), never float/numeric
- Text with case-insensitive unique: `CITEXT`

## Migration Rules
- Each migration is a pair of `NNNN_name.up.sql` + `NNNN_name.down.sql`
- Numbering is sequential and immutable after merge
- Never edit a merged migration — add a new one
- Extensions (citext, timescaledb) in first migration before dependent tables
- Each table created by its owner FR; other FRs extend via ALTER TABLE
