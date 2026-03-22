DROP INDEX IF EXISTS idx_short_urls_user_id;

ALTER TABLE short_urls DROP COLUMN IF EXISTS user_id;