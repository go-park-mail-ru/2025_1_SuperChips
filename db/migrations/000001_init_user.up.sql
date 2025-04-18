CREATE TABLE IF NOT EXISTS flow_user (
    id INT GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1) PRIMARY KEY,
    username TEXT NOT NULL UNIQUE CHECK(LENGTH(username) <= 128),
    avatar TEXT DEFAULT '',
    public_name TEXT NOT NULL CHECK(LENGTH(public_name) <= 128),
    email TEXT NOT NULL UNIQUE CHECK(LENGTH(email) <= 128),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    password TEXT NOT NULL,
    birthday DATE,
    about TEXT CHECK(LENGTH(about) <= 2048),
    jwt_version INTEGER NOT NULL DEFAULT 1
);
