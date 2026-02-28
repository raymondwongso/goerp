-- +goose Up
CREATE TABLE sessions (
  id              UUID        PRIMARY KEY DEFAULT uuidv7(),
  user_id         UUID        NOT NULL,
  ip_address      INET,
  user_agent      TEXT,
  is_revoked      BOOLEAN     NOT NULL DEFAULT false,
  absolute_expiry TIMESTAMPTZ NOT NULL DEFAULT now() + INTERVAL '30 days',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_seen_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE sessions;
