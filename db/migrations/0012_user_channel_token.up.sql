CREATE TABLE user_channel_token (
  user_id    BIGINT      NOT NULL REFERENCES app_user(id),
  channel    TEXT        NOT NULL CHECK (channel IN ('push','email','sms')),
  platform   TEXT        NOT NULL CHECK (platform IN ('ios','android','web','email','sms')),
  address    TEXT        NOT NULL,
  verified   BOOLEAN     NOT NULL DEFAULT false,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, channel, platform)
);
