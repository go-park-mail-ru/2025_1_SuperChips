CREATE TABLE IF NOT EXISTS subscription (
    user_id INT NOT NULL,
    target_id INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, target_id),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES flow_user(id) ON DELETE CASCADE,
    CONSTRAINT fk_target FOREIGN KEY (target_id) REFERENCES flow_user(id) ON DELETE CASCADE
);
