CREATE TABLE IF NOT EXISTS invitation_user (
    id INT GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1) PRIMARY KEY,
    invitation_id INT NOT NULL,
    user_id INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES flow_user(id) ON DELETE CASCADE,
    FOREIGN KEY (invitation_id) REFERENCES board_invitation(id) ON DELETE CASCADE
);