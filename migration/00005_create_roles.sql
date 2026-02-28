-- +goose Up
CREATE TABLE roles (
  id          UUID        PRIMARY KEY DEFAULT uuidv7(),
  name        TEXT        NOT NULL UNIQUE,
  description TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE roles;
