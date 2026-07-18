-- +goose Up
CREATE TABLE IF NOT EXISTS users(
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    image_url TEXT
);

-- +goose Down
DROP TABLE users;
