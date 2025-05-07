CREATE INDEX IF NOT EXISTS idx_flow_title_description_search ON flow USING GIN(to_tsvector('english', title || ' ' || description));
CREATE INDEX IF NOT EXISTS idx_flow_is_private ON flow (is_private);
CREATE INDEX IF NOT EXISTS idx_flow_author_id ON flow (author_id);
