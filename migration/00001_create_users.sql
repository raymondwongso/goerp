-- +goose Up
CREATE TABLE users (
  id            UUID        PRIMARY KEY DEFAULT uuidv7(),
  email         TEXT        NOT NULL UNIQUE,
  display_name  TEXT,
  avatar_url    TEXT,
  is_active     BOOLEAN     NOT NULL DEFAULT true,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE users;
