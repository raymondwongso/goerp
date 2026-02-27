-- +goose Up
CREATE TABLE oauth_states (
  state         TEXT        PRIMARY KEY,
  code_verifier TEXT        NOT NULL,
  redirect_to   TEXT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at    TIMESTAMPTZ NOT NULL DEFAULT now() + INTERVAL '5 minutes'
);

-- +goose Down
DROP TABLE oauth_states;
