-- +goose Up
CREATE TABLE IF NOT EXISTS keys (
    id BIGSERIAL PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    users_id BIGINT REFERENCES users(id),
    creation_date TIMESTAMPTZ NOT NULL DEFAULT now(),
    usages_left INT NOT NULL DEFAULT 1
);

-- +goose Down
DROP TABLE keys;
