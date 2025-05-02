CREATE INDEX idx_board_flow_count ON board (flow_count DESC);
CREATE INDEX idx_board_is_private_created_at ON board (is_private, created_at DESC);
CREATE INDEX idx_board_created_at ON board (created_at DESC);
CREATE INDEX idx_board_author_id ON board (author_id);
CREATE INDEX idx_board_is_private ON board (is_private);
CREATE INDEX idx_board_board_name_search ON board USING GIN(to_tsvector('english', board_name));
