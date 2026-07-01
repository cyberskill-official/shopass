-- services/scrape/migrations/0001_scrape_job.sql
CREATE TYPE scrape_tier AS ENUM ('hot', 'warm', 'cold');

CREATE TABLE scrape_job (
  product_id    BIGINT      PRIMARY KEY REFERENCES tracked_product(id),
  platform_id   SMALLINT    NOT NULL REFERENCES platform(id),
  tier          scrape_tier NOT NULL DEFAULT 'cold',
  next_run_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  attempts      INTEGER     NOT NULL DEFAULT 0,
  last_status   TEXT        NOT NULL DEFAULT 'pending',  -- pending|ok|retry|failed
  locked_until  TIMESTAMPTZ
);

CREATE INDEX idx_job_due ON scrape_job (next_run_at)
  WHERE last_status <> 'failed';
CREATE INDEX idx_job_platform ON scrape_job (platform_id, tier);
