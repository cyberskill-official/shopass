CREATE TABLE proxy_usage (
  day          DATE        NOT NULL,
  provider     TEXT        NOT NULL,           -- 'brightdata','oxylabs','decodo','iproyal',...
  tier         TEXT        NOT NULL,           -- 'enterprise'|'mid'|'budget'
  country      TEXT        NOT NULL,
  bytes_used   BIGINT      NOT NULL DEFAULT 0,
  cost_micro_usd BIGINT    NOT NULL DEFAULT 0, -- chi phí ước tính, micro-USD (số nguyên)
  PRIMARY KEY (day, provider, tier, country)
);
