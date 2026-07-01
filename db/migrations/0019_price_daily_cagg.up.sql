CREATE MATERIALIZED VIEW price_daily
  WITH (timescaledb.continuous) AS
  SELECT product_id,
         time_bucket('1 day', ts) AS day,
         min(price)      AS min_p,
         max(price)      AS max_p,
         last(price, ts) AS close_p
  FROM price_snapshot
  GROUP BY product_id, day
  WITH NO DATA;

SELECT add_continuous_aggregate_policy('price_daily',
  start_offset => INTERVAL '3 days',
  end_offset   => INTERVAL '1 hour',
  schedule_interval => INTERVAL '1 hour');
