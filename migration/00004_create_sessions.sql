-- +goose Up
CREATE TABLE sessions (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id         UUID        NOT NULL REFERENCES users(id),
  ip_address      INET,
  user_agent      TEXT,
  is_revoked      BOOLEAN     NOT NULL DEFAULT false,
  absolute_expiry TIMESTAMPTZ NOT NULL DEFAULT now() + INTERVAL '30 days',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_seen_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_active ON sessions(id, absolute_expiry) WHERE NOT is_revoked;

-- +goose Down
DROP TABLE sessions;
