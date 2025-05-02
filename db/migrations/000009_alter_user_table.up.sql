CREATE INDEX IF NOT EXISTS idx_flow_user_username_search ON flow_user USING GIN(to_tsvector('english', username));
CREATE INDEX IF NOT EXISTS idx_flow_user_id ON flow_user (id);
