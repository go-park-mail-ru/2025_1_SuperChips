CREATE SEQUENCE IF NOT EXISTS flow_board_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


CREATE TABLE IF NOT EXISTS board (
    id INTEGER DEFAULT nextval('flow_board_id_seq') PRIMARY KEY,
    author_id INTEGER NOT NULL,
    board_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_private BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY (author_id) REFERENCES flow_user(id) ON DELETE CASCADE,
    UNIQUE (author_id, board_name)
);