CREATE TABLE ecommerce_obligation (
  id             BIGSERIAL   PRIMARY KEY,
  obligation_key TEXT        NOT NULL,
  description_vi TEXT        NOT NULL,
  status         TEXT        NOT NULL DEFAULT 'not_started'
                   CHECK (status IN ('not_started','submitted','approved','done','n_a')),
  due_at         TIMESTAMPTZ,
  completed_at   TIMESTAMPTZ,
  source_law     TEXT        NOT NULL,   -- 'ND_52_2013' | 'ND_85_2021' | 'DRAFT_2025'
  version        INTEGER     NOT NULL DEFAULT 1,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (obligation_key, version)
);

CREATE TABLE yearly_transaction_count (
  year   INTEGER PRIMARY KEY,
  count  BIGINT  NOT NULL DEFAULT 0
);

-- Nguong cau hinh versioned (mot cho), khong hardcode rai rac.
CREATE TABLE compliance_threshold (
  key     TEXT    NOT NULL,
  value   BIGINT  NOT NULL,
  version INTEGER NOT NULL DEFAULT 1,
  UNIQUE (key, version)
);
INSERT INTO compliance_threshold (key, value) VALUES
  ('foreign_platform_yearly_tx', 100000); -- nguong NĐ 85/2021

INSERT INTO ecommerce_obligation (obligation_key, description_vi, source_law) VALUES
  ('moit_registration', 'Dang ky/thong bao website TMĐT voi Bo Cong Thuong', 'ND_52_2013'),
  ('affiliate_disclosure', 'Cong bo quan he affiliate (du thao Luat TMĐT 2025 - cho luat chot)', 'DRAFT_2025'),
  ('livestream_disclosure', 'Cong bo noi dung livestream thuong mai (du thao 2025 - cho luat chot)', 'DRAFT_2025');
