CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- GIN trigram index để truy vấn similarity không quét toàn bảng
CREATE INDEX IF NOT EXISTS idx_tp_title_trgm
  ON tracked_product USING gin (title gin_trgm_ops);

-- Hàng đợi duyệt tay cho merge confidence thấp (DEC-PRICE-23)
CREATE TABLE canonical_review_queue (
  id            BIGSERIAL PRIMARY KEY,
  product_id    BIGINT      NOT NULL REFERENCES tracked_product(id),
  candidate_key TEXT        NOT NULL,              -- canonical_key đề xuất gộp vào
  confidence    REAL        NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
  status        TEXT        NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending','approved','rejected')),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  decided_at    TIMESTAMPTZ,
  UNIQUE (product_id, candidate_key)
);

CREATE INDEX idx_crq_pending ON canonical_review_queue (status, created_at)
  WHERE status = 'pending';
