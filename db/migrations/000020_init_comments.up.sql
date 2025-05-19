CREATE TABLE IF NOT EXISTS comment (
    id INT GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1) PRIMARY KEY,
    author_id INT,
    flow_id INT,
    contents TEXT,
    like_count INT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY author_id REFERENCES flow_user(id),
    FOREIGN KEY flow_id REFERENCES flow(id) ON DELETE CASCADE
);