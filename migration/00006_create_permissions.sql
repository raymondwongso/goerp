-- +goose Up
CREATE TABLE permissions (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  resource TEXT NOT NULL,
  action   TEXT NOT NULL,
  UNIQUE(resource, action)
);

-- +goose Down
DROP TABLE permissions;
