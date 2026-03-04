-- +goose Up
ALTER TABLE oauth_states ADD COLUMN ip_address INET;

-- +goose Down
ALTER TABLE oauth_states DROP COLUMN ip_address;
