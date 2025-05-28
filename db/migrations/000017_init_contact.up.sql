CREATE TABLE contact (
    id INT GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1) PRIMARY KEY,
    user_username TEXT NOT NULL,
    contact_username TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_username, contact_username),
    CONSTRAINT fk_user FOREIGN KEY (user_username) REFERENCES flow_user(username) ON DELETE CASCADE,
    CONSTRAINT fk_contact FOREIGN KEY (contact_username) REFERENCES flow_user(username) ON DELETE CASCADE
);
