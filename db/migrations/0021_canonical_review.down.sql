DROP INDEX IF EXISTS idx_crq_pending;
DROP TABLE IF EXISTS canonical_review_queue;
DROP INDEX IF EXISTS idx_tp_title_trgm;
-- We do not drop the extension to avoid breaking other things
