CREATE TABLE IF NOT EXISTS flow (
    id INT GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1) PRIMARY KEY,
    title TEXT CHECK (LENGTH(title) <= 128),
    description TEXT CHECK (LENGTH(description) <= 512),
    author_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_private BOOLEAN NOT NULL DEFAULT FALSE,
    media_url TEXT NOT NULL, -- генерится на бэке, так что мне кажется, здесь ограничение не нужно
    like_count INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (author_id) REFERENCES flow_user(id)
);
