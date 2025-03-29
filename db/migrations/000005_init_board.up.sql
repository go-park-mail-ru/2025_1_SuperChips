CREATE TABLE IF NOT EXISTS board (
    id INTEGER DEFAULT NEXT VALUE FOR flow_id_seq PRIMARY KEY,
    author_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_private BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY (author_id) REFERENCES flow_user(id) ON DELETE CASCADE
);