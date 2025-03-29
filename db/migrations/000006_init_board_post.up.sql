CREATE TABLE IF NOT EXISTS board_post (
    board_id INTEGER NOT NULL,
    flow_id INTEGER NOT NULL,
    saved_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (board_id, flow_id),
    FOREIGN KEY (board_id) REFERENCES board(id) ON DELETE CASCADE,
    FOREIGN KEY (flow_id) REFERENCES flow(id) ON DELETE CASCADE
);