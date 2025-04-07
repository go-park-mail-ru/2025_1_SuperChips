CREATE TABLE IF NOT EXISTS flow_like (
    user_id INTEGER NOT NULL,
    flow_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, flow_id),
    FOREIGN KEY (user_id) REFERENCES flow_user(id) ON DELETE CASCADE,
    FOREIGN KEY (flow_id) REFERENCES flow(id) ON DELETE CASCADE
);