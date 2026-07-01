CREATE TABLE consent_policy (
  id             BIGSERIAL   PRIMARY KEY,
  purpose_key    TEXT        NOT NULL,
  version        INTEGER     NOT NULL,
  title_vi       TEXT        NOT NULL,
  body_vi        TEXT        NOT NULL,
  effective_from TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (purpose_key, version)
);

-- Seed purpose loi (moi purpose mot ban chinh sach v1).
INSERT INTO consent_policy (purpose_key, version, title_vi, body_vi) VALUES
  ('cart_read', 1, 'Doc gio hang va voucher cua ban',
   'SanDeal doc gio hang/voucher tren trinh duyet cua ban de toi uu. KHONG gui cookie/token.'),
  ('price_tracking', 1, 'Theo doi gia theo tai khoan',
   'Luu san pham ban theo doi de canh bao khi gia thay doi.'),
  ('marketing_notification', 1, 'Nhan thong bao khuyen mai',
   'Gui alert sale va goi y san pham. Co the thu hoi bat ky luc nao.'),
  ('analytics_b2b', 1, 'Dong gop du lieu xu huong an danh',
   'Du lieu gia da an danh hoa (k-anonymity) phuc vu bao cao thi truong.');
