ALTER TABLE app_user
  ADD COLUMN pwd_hash         TEXT,
  ADD COLUMN referral_code_id BIGINT;
