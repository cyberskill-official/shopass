ALTER TABLE price_snapshot SET (
  timescaledb.compress,
  timescaledb.compress_segmentby = 'product_id'
);

SELECT add_compression_policy('price_snapshot', INTERVAL '30 days');

SELECT add_retention_policy('price_snapshot', INTERVAL '18 months');
