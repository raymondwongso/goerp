-- +goose Up
CREATE TABLE permissions (
  id       UUID PRIMARY KEY DEFAULT uuidv7(),
  resource TEXT NOT NULL,
  action   TEXT NOT NULL,
  UNIQUE(resource, action)
);

-- +goose Down
DROP TABLE permissions;
