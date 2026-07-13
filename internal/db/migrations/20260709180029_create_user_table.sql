-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name   TEXT         NOT NULL CHECK (full_name <> ''),
    email       TEXT         NOT NULL UNIQUE CHECK (email <> ''),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email ON users(email);



-- +goose Down
DROP TABLE IF EXISTS users;
