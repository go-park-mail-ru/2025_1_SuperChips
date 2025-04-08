CREATE SEQUENCE IF NOT EXISTS flow_user_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE IF NOT EXISTS flow_user (
    id INTEGER DEFAULT nextval('flow_user_id_seq') PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    avatar TEXT DEFAULT '',
    public_name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    password TEXT NOT NULL,
    birthday DATE,
    about TEXT,
    jwt_version INTEGER NOT NULL DEFAULT 1
);

