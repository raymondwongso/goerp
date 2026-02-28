-- +goose Up
CREATE TABLE oauth_accounts (
  id           UUID        PRIMARY KEY DEFAULT uuidv7(),
  user_id      UUID        NOT NULL,
  provider     TEXT        NOT NULL,
  provider_sub TEXT        NOT NULL,
  email        TEXT        NOT NULL,
  last_login   TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(provider, provider_sub)
);

-- +goose Down
DROP TABLE oauth_accounts;
