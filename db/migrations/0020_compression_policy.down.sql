SELECT remove_retention_policy('price_snapshot', if_exists => true);
SELECT remove_compression_policy('price_snapshot', if_exists => true);
ALTER TABLE price_snapshot SET (timescaledb.compress = false);
