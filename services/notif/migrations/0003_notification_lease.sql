-- TASK-NOTIF-003: lease columns for fan-out claim/retry
ALTER TABLE notification
  ADD COLUMN attempts    INTEGER     NOT NULL DEFAULT 0,
  ADD COLUMN lease_until TIMESTAMPTZ,
  ADD COLUMN last_error  TEXT;
