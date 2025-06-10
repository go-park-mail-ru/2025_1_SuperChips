CREATE TABLE IF NOT EXISTS board_coauthor (
    id INT GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1) PRIMARY KEY,
    board_id INT NOT NULL,
    coauthor_id INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (coauthor_id) REFERENCES flow_user(id) ON DELETE CASCADE,
    FOREIGN KEY (board_id) REFERENCES board(id) ON DELETE CASCADE,
    UNIQUE (board_id, coauthor_id)
);