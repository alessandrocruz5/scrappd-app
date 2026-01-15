-- 000003_add_usage_tracking.down.sql
-- Remove usage_tracking table

-- Drop trigger
DROP TRIGGER IF EXISTS update_items_updated_at ON content.items;
DROP TRIGGER IF EXISTS update_usage_tracking_updated_at ON content.usage_tracking;

-- Drop indexes
DROP INDEX IF EXISTS content.idx_usage_tracking_user_current;
DROP INDEX IF EXISTS content.idx_usage_tracking_period;
DROP INDEX IF EXISTS content.idx_usage_tracking_user_id;

-- Drop function
DROP FUNCTION IF EXISTS content.update_updated_at_column();

-- Drop table
DROP TABLE IF EXISTS content.usage_tracking;