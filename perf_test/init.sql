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
CREATE TABLE IF NOT EXISTS flow (
    id INT GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1) PRIMARY KEY,
    title TEXT CHECK (LENGTH(title) <= 128),
    description TEXT CHECK (LENGTH(description) <= 1024),
    author_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_private BOOLEAN NOT NULL DEFAULT FALSE,
    media_url TEXT NOT NULL,
    like_count INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (author_id) REFERENCES flow_user(id)
);
ALTER TABLE flow_user
ADD CONSTRAINT unique_email_username UNIQUE (email, username);CREATE TABLE IF NOT EXISTS flow_like (
    user_id INTEGER NOT NULL,
    flow_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, flow_id),
    FOREIGN KEY (user_id) REFERENCES flow_user(id) ON DELETE CASCADE,
    FOREIGN KEY (flow_id) REFERENCES flow(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS board (
    id INT GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1) PRIMARY KEY,
    author_id INTEGER NOT NULL,
    board_name TEXT NOT NULL CHECK (LENGTH(board_name) <= 128),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_private BOOLEAN NOT NULL DEFAULT FALSE,
    flow_count INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (author_id) REFERENCES flow_user(id) ON DELETE CASCADE,
    UNIQUE (author_id, board_name)
);

CREATE TABLE IF NOT EXISTS board_post (
    board_id INTEGER NOT NULL,
    flow_id INTEGER NOT NULL,
    saved_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (board_id, flow_id),
    FOREIGN KEY (board_id) REFERENCES board(id) ON DELETE CASCADE,
    FOREIGN KEY (flow_id) REFERENCES flow(id) ON DELETE CASCADE
);
ALTER TABLE flow
ADD COLUMN IF NOT EXISTS width INTEGER,
ADD COLUMN IF NOT EXISTS height INTEGER;ALTER TABLE flow_user
ADD COLUMN IF NOT EXISTS external_id TEXT;CREATE INDEX IF NOT EXISTS idx_flow_user_username_search ON flow_user USING GIN(to_tsvector('english', username));

CREATE INDEX IF NOT EXISTS idx_flow_user_id ON flow_user (id);
CREATE INDEX IF NOT EXISTS idx_flow_title_description_search ON flow USING GIN(to_tsvector('english', title || ' ' || description));
CREATE INDEX IF NOT EXISTS idx_flow_is_private ON flow (is_private);
CREATE INDEX IF NOT EXISTS idx_flow_author_id ON flow (author_id);

ALTER TABLE flow_user
ADD COLUMN IF NOT EXISTS subscriber_count INTEGER NOT NULL DEFAULT 0;ALTER TABLE flow_user
ADD COLUMN IF NOT EXISTS is_external_avatar BOOL;

CREATE TABLE IF NOT EXISTS subscription (
    user_id INT NOT NULL,
    target_id INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, target_id),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES flow_user(id) ON DELETE CASCADE,
    CONSTRAINT fk_target FOREIGN KEY (target_id) REFERENCES flow_user(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_board_flow_count ON board (flow_count DESC);
CREATE INDEX IF NOT EXISTS idx_board_is_private_created_at ON board (is_private, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_board_created_at ON board (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_board_author_id ON board (author_id);
CREATE INDEX IF NOT EXISTS idx_board_is_private ON board (is_private);
CREATE INDEX IF NOT EXISTS idx_board_board_name_search ON board USING GIN(to_tsvector('english', board_name));

CREATE TABLE IF NOT EXISTS chat (
    id INT GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1) PRIMARY KEY,
    user1 TEXT NOT NULL,
    user2 TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user1, user2),
    CONSTRAINT fk_user1 FOREIGN KEY (user1) REFERENCES flow_user(username),
    CONSTRAINT fk_user2 FOREIGN KEY (user2) REFERENCES flow_user(username)
);
CREATE TABLE message (
    id INT GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1) PRIMARY KEY,
    content TEXT NOT NULL,
    sender TEXT NOT NULL,
    recipient TEXT NOT NULL,
    chat_id INT NOT NULL,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    is_read BOOLEAN DEFAULT FALSE,
    CONSTRAINT fk_chat FOREIGN KEY (chat_id) REFERENCES chat(id) ON DELETE CASCADE
);
CREATE TABLE contact (
    id INT GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1) PRIMARY KEY,
    user_username TEXT NOT NULL,
    contact_username TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_username, contact_username),
    CONSTRAINT fk_user FOREIGN KEY (user_username) REFERENCES flow_user(username) ON DELETE CASCADE,
    CONSTRAINT fk_contact FOREIGN KEY (contact_username) REFERENCES flow_user(username) ON DELETE CASCADE
);

ALTER TABLE message
ADD COLUMN IF NOT EXISTS sent BOOL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS color (
    id INT GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1) PRIMARY KEY,
    flow_id INT,
    color_hex TEXT,
    FOREIGN KEY (flow_id) REFERENCES flow(id)
);

CREATE TABLE IF NOT EXISTS comment (
    id INT GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1) PRIMARY KEY,
    author_id INT,
    flow_id INT,
    contents TEXT,
    like_count INT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (author_id) REFERENCES flow_user(id),
    FOREIGN KEY (flow_id) REFERENCES flow(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS notification (
    id INT GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1) PRIMARY KEY,
    author_id INTEGER NOT NULL,
    receiver_id INTEGER NOT NULL,
    notification_type TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    additional JSONB NOT NULL DEFAULT '{}'::jsonb,
    FOREIGN KEY (author_id) REFERENCES flow_user(id) ON DELETE CASCADE,
    FOREIGN KEY (receiver_id) REFERENCES flow_user(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS comment_like (
    user_id INTEGER NOT NULL,
    comment_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, comment_id),
    FOREIGN KEY (user_id) REFERENCES flow_user(id) ON DELETE CASCADE,
    FOREIGN KEY (comment_id) REFERENCES comment(id) ON DELETE CASCADE
);

ALTER TABLE comment
ALTER COLUMN like_count SET DEFAULT 0;
