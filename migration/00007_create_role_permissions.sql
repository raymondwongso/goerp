-- +goose Up
CREATE TABLE role_permissions (
  role_id       UUID NOT NULL,
  permission_id UUID NOT NULL,
  PRIMARY KEY (role_id, permission_id)
);

-- +goose Down
DROP TABLE role_permissions;
