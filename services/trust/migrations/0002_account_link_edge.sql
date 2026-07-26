-- services/trust/migrations/0002_account_link_edge.sql
CREATE TABLE IF NOT EXISTS account_link_edge (
  a_user    BIGINT NOT NULL REFERENCES app_user(id),
  b_user    BIGINT NOT NULL REFERENCES app_user(id),
  link_type TEXT   NOT NULL,
  weight    REAL   NOT NULL DEFAULT 1.0,
  PRIMARY KEY (a_user, b_user, link_type),
  CHECK (a_user < b_user)
);
CREATE INDEX IF NOT EXISTS idx_edge_b ON account_link_edge (b_user, link_type);
