ALTER TABLE flow_user
ADD COLUMN IF NOT EXISTS is_external_avatar BOOL;
