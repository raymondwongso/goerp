-- +goose Up
CREATE TABLE user_roles (
  user_id    UUID        NOT NULL,
  role_id    UUID        NOT NULL,
  granted_by UUID,
  granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, role_id)
);

-- +goose Down
DROP TABLE user_roles;
