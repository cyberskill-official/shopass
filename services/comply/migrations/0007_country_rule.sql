CREATE TABLE IF NOT EXISTS country_rule (
  country_code TEXT NOT NULL,
  gate_key     TEXT NOT NULL,
  allowed      BOOLEAN NOT NULL DEFAULT false,
  value        TEXT,
  version      INT NOT NULL DEFAULT 1,
  PRIMARY KEY (country_code, gate_key)
);

INSERT INTO country_rule (country_code, gate_key, allowed, value) VALUES
  ('VN', 'voucher_stacking', true, 'allowed'),
  ('VN', 'affiliate_channel', true, 'allowed'),
  ('VN', 'data_protection_regime', true, 'VN_PDPL'),
  ('ID', 'voucher_stacking', true, 'allowed'),
  ('ID', 'affiliate_channel', true, 'allowed'),
  ('ID', 'data_protection_regime', true, 'ID_PDP'),
  ('TH', 'voucher_stacking', true, 'allowed'),
  ('TH', 'affiliate_channel', true, 'allowed'),
  ('TH', 'data_protection_regime', true, 'TH_PDPA'),
  ('PH', 'voucher_stacking', false, 'deny'),
  ('PH', 'affiliate_channel', true, 'allowed'),
  ('PH', 'data_protection_regime', false, NULL),
  ('MY', 'voucher_stacking', false, 'deny'),
  ('MY', 'affiliate_channel', true, 'allowed'),
  ('MY', 'data_protection_regime', false, NULL),
  ('SG', 'voucher_stacking', true, 'allowed'),
  ('SG', 'affiliate_channel', true, 'allowed'),
  ('SG', 'data_protection_regime', false, NULL),
  ('TW', 'voucher_stacking', true, 'allowed'),
  ('TW', 'affiliate_channel', true, 'allowed'),
  ('TW', 'data_protection_regime', false, NULL)
ON CONFLICT (country_code, gate_key) DO NOTHING;
