CREATE TABLE processing_activity (
  id               BIGSERIAL   PRIMARY KEY,
  name             TEXT        NOT NULL,
  purpose_key      TEXT        NOT NULL,
  data_categories  TEXT[]      NOT NULL,
  started_at       TIMESTAMPTZ NOT NULL,
  cross_border     BOOLEAN     NOT NULL DEFAULT false,
  recipient_country TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (NOT cross_border OR recipient_country IS NOT NULL)
);
