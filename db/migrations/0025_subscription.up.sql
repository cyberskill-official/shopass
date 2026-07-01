CREATE TABLE subscription (
  id         BIGSERIAL   PRIMARY KEY,
  user_id    BIGINT      NOT NULL REFERENCES app_user(id),
  plan_id    SMALLINT    NOT NULL REFERENCES plan_catalog(id),
  started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  renews_at  TIMESTAMPTZ NOT NULL,
  status     TEXT        NOT NULL DEFAULT 'active'
               CHECK (status IN ('active','past_due','canceled','expired')),
  CHECK (renews_at > started_at)
);

CREATE UNIQUE INDEX uq_sub_active_user
  ON subscription (user_id) WHERE status = 'active';
