-- +goose Up
CREATE TABLE oauth_accounts (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      UUID        NOT NULL REFERENCES users(id),
  provider     TEXT        NOT NULL,
  provider_sub TEXT        NOT NULL,
  email        TEXT        NOT NULL,
  last_login   TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(provider, provider_sub)
);

CREATE INDEX idx_oauth_accounts_user ON oauth_accounts(user_id);

-- +goose Down
DROP TABLE oauth_accounts;