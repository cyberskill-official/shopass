CREATE TABLE affiliate_network (
  id                  SMALLSERIAL PRIMARY KEY,
  code                TEXT     NOT NULL UNIQUE
                        CHECK (code IN ('involve_asia','accesstrade')),
  platform_id         SMALLINT NOT NULL REFERENCES platform(id),
  base_url            TEXT     NOT NULL,   -- gốc deep link network
  target_param        TEXT     NOT NULL,   -- tên tham số bọc URL đích
  sub_id_param        TEXT     NOT NULL,   -- tên tham số mang sub_id (last-click)
  postback_secret_ref TEXT     NOT NULL,   -- Vault key path; KHÔNG cleartext (§1 #2)
  active              BOOLEAN  NOT NULL DEFAULT true
);

INSERT INTO affiliate_network (code, platform_id, base_url, target_param, sub_id_param, postback_secret_ref)
VALUES
  ('involve_asia', 1, 'https://go.involve.asia/aff', 'url', 'sub_id', 'affil/involve_asia/postback_secret'),
  ('accesstrade',  1, 'https://go.isclix.com/deep_link', 'url_enc', 'utm_content', 'affil/accesstrade/postback_secret');
