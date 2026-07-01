INSERT INTO platform (id, code, country, base_url) VALUES
  (1, 'shopee', 'VN', 'https://shopee.vn'),
  (2, 'tiktok', 'VN', 'https://www.tiktok.com'),
  (3, 'lazada', 'VN', 'https://www.lazada.vn')
ON CONFLICT (code) DO NOTHING;
