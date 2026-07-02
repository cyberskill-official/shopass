-- services/bill/migrations/0003_payment.sql
CREATE TABLE payment (
  id             BIGSERIAL   PRIMARY KEY,
  order_ref      TEXT        NOT NULL UNIQUE,
  subscription_id BIGINT     REFERENCES subscription(id),
  gateway        TEXT        NOT NULL,
  amount         BIGINT      NOT NULL CHECK (amount >= 0),
  fee            BIGINT      NOT NULL DEFAULT 0 CHECK (fee >= 0),
  status         TEXT        NOT NULL DEFAULT 'pending'
                   CHECK (status IN ('pending','paid','failed','mismatch')),
  transaction_id TEXT,
  paid_at        TIMESTAMPTZ,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_payment_pending ON payment (created_at) WHERE status = 'pending';
