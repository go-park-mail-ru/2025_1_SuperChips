ALTER TABLE flow_user
ADD CONSTRAINT unique_email_username UNIQUE (email, username);