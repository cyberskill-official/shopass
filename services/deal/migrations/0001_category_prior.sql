-- services/deal/migrations/0001_category_prior.sql
-- Gộp thống kê theo category_id, CHỈ từ SKU đã MATURE (>= 90 ngày lịch sử).
-- mature_sku: 1 dòng/SKU MATURE kèm median giá riêng và độ sâu giảm gần nhất.
CREATE MATERIALIZED VIEW category_prior AS
  WITH mature_sku AS (
    SELECT tp.category_id,
           percentile_cont(0.5) WITHIN GROUP (ORDER BY pd.close_p) AS sku_median,
           max(pd.max_p) AS sku_list,    -- xấp xỉ giá niêm yết = đỉnh quan sát
           min(pd.min_p) AS sku_floor
    FROM tracked_product tp
    JOIN price_daily pd ON pd.product_id = tp.id
    WHERE tp.category_id IS NOT NULL
      AND now() - tp.first_seen >= INTERVAL '90 days'   -- chỉ SKU MATURE
    GROUP BY tp.id, tp.category_id
  )
  SELECT category_id,
         percentile_cont(0.5) WITHIN GROUP (ORDER BY sku_median)              AS median_price,
         percentile_cont(0.5) WITHIN GROUP (
           ORDER BY (sku_list - sku_floor)::float / NULLIF(sku_list, 0))      AS typical_discount_depth,
         count(*)                                                             AS sample_count
  FROM mature_sku
  GROUP BY category_id;

CREATE UNIQUE INDEX idx_category_prior_cat ON category_prior (category_id);
-- Refresh hằng ngày (REFRESH ... CONCURRENTLY, đăng ký ở scheduler deal-svc, §1 #11).
