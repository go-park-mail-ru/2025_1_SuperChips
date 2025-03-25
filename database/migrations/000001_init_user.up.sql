CREATE TABLE IF NOT EXISTS flow_user (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    avatar TEXT,
    public_name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    password TEXT NOT NULL,
    birthday DATE,
    about TEXT,
    jwt_version INTEGER NOT NULL DEFAULT 1
);
