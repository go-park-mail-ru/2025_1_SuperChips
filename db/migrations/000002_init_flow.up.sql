CREATE SEQUENCE IF NOT EXISTS flow_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


CREATE TABLE IF NOT EXISTS flow (
    id INTEGER DEFAULT nextval('flow_id_seq') PRIMARY KEY,
    title TEXT,
    description TEXT,
    author_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_private BOOLEAN NOT NULL DEFAULT FALSE,
    media_url TEXT NOT NULL,
    like_count INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (author_id) REFERENCES flow_user(id)
);
