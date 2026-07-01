CREATE TABLE app_user (
  id           BIGSERIAL    PRIMARY KEY,
  email        CITEXT       UNIQUE,                 -- case-insensitive unique
  phone        TEXT,
  display_name TEXT,
  locale       TEXT         NOT NULL DEFAULT 'vi-VN',
  status       TEXT         NOT NULL DEFAULT 'active',
  created_at   TIMESTAMPTZ  DEFAULT now()
);
