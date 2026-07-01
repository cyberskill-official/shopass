CREATE TABLE user_activity_coin_task (
  id          BIGSERIAL   PRIMARY KEY,
  user_id     BIGINT      NOT NULL REFERENCES app_user(id),
  platform_id SMALLINT    NOT NULL REFERENCES platform(id),
  task_type   TEXT        NOT NULL,
  due_date    DATE        NOT NULL,
  done        BOOLEAN     NOT NULL DEFAULT false,
  UNIQUE (user_id, platform_id, task_type, due_date)
);

CREATE INDEX idx_coin_due ON user_activity_coin_task (user_id, due_date) WHERE done = false;
