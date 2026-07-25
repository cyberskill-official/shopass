-- Harden refresh_token lookup by rotation family and enforce dsar_request → app_user.
-- Safe for both runners: deploy/migrate.sh (file ledger) and db/internal/migrate (golang-migrate).
CREATE INDEX IF NOT EXISTS idx_rt_family ON refresh_token (family_id);

-- Drop orphan DSAR rows before adding the FK (dev/demo DBs may have seeded user_ids).
DELETE FROM dsar_request d
 WHERE NOT EXISTS (SELECT 1 FROM app_user u WHERE u.id = d.user_id);

ALTER TABLE dsar_request
  DROP CONSTRAINT IF EXISTS dsar_request_user_id_fkey;

ALTER TABLE dsar_request
  ADD CONSTRAINT dsar_request_user_id_fkey
  FOREIGN KEY (user_id) REFERENCES app_user(id);
