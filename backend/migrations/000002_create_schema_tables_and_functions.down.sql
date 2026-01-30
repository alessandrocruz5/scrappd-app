-- 000002_create_schema_tables_and_functions.down.sql
-- ============================================
-- ROLLBACK SCHEMA TABLES AND FUNCTIONS
-- ============================================

-- ============================================
-- DROP TRIGGERS (in reverse order)
-- ============================================

-- Social triggers
DROP TRIGGER IF EXISTS like_delete_trigger ON social.project_likes;
DROP TRIGGER IF EXISTS like_insert_trigger ON social.project_likes;
DROP TRIGGER IF EXISTS follow_delete_trigger ON social.follows;
DROP TRIGGER IF EXISTS follow_insert_trigger ON social.follows;
DROP TRIGGER IF EXISTS update_comments_updated_at ON social.project_comments;

-- Marketplace triggers
DROP TRIGGER IF EXISTS update_earnings_updated_at ON marketplace.creator_earnings;

-- Content triggers
DROP TRIGGER IF EXISTS update_page_items_updated_at ON content.page_items;
DROP TRIGGER IF EXISTS update_items_updated_at ON content.items;
DROP TRIGGER IF EXISTS update_pages_updated_at ON content.pages;
DROP TRIGGER IF EXISTS update_projects_updated_at ON content.projects;

-- Auth triggers
DROP TRIGGER IF EXISTS update_users_updated_at ON auth.users;

-- ============================================
-- DROP HELPER FUNCTIONS
-- ============================================

DROP FUNCTION IF EXISTS decrement_like_count();
DROP FUNCTION IF EXISTS increment_like_count();
DROP FUNCTION IF EXISTS decrement_follower_counts();
DROP FUNCTION IF EXISTS increment_follower_counts();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- ============================================
-- DROP INDEXES
-- ============================================

-- Analytics schema indexes
DROP INDEX IF EXISTS analytics.idx_daily_stats_stat_date;
DROP INDEX IF EXISTS analytics.idx_usage_logs_created_at;
DROP INDEX IF EXISTS analytics.idx_usage_logs_action_type;
DROP INDEX IF EXISTS analytics.idx_usage_logs_user_id;
DROP INDEX IF EXISTS analytics.idx_project_views_viewed_at;
DROP INDEX IF EXISTS analytics.idx_project_views_project_id;

-- Marketplace schema indexes
DROP INDEX IF EXISTS marketplace.idx_payouts_status;
DROP INDEX IF EXISTS marketplace.idx_payouts_user_id;
DROP INDEX IF EXISTS marketplace.idx_template_purchases_project_id;
DROP INDEX IF EXISTS marketplace.idx_template_purchases_seller_id;
DROP INDEX IF EXISTS marketplace.idx_template_purchases_buyer_id;

-- Social schema indexes
DROP INDEX IF EXISTS social.idx_project_comments_parent_id;
DROP INDEX IF EXISTS social.idx_project_comments_project_id;
DROP INDEX IF EXISTS social.idx_project_likes_project_id;
DROP INDEX IF EXISTS social.idx_follows_following_id;
DROP INDEX IF EXISTS social.idx_follows_follower_id;

-- Content schema indexes
DROP INDEX IF EXISTS content.idx_page_items_z_index;
DROP INDEX IF EXISTS content.idx_page_items_item_id;
DROP INDEX IF EXISTS content.idx_page_items_page_id;
DROP INDEX IF EXISTS content.idx_items_deleted_at;
DROP INDEX IF EXISTS content.idx_items_created_at;
DROP INDEX IF EXISTS content.idx_items_processing_status;
DROP INDEX IF EXISTS content.idx_items_user_id;
DROP INDEX IF EXISTS content.idx_pages_project_id;
DROP INDEX IF EXISTS content.idx_projects_deleted_at;
DROP INDEX IF EXISTS content.idx_projects_created_at;
DROP INDEX IF EXISTS content.idx_projects_is_template;
DROP INDEX IF EXISTS content.idx_projects_visibility;
DROP INDEX IF EXISTS content.idx_projects_user_id;

-- Auth schema indexes
DROP INDEX IF EXISTS auth.idx_refresh_tokens_expires_at;
DROP INDEX IF EXISTS auth.idx_refresh_tokens_user_id;
DROP INDEX IF EXISTS auth.idx_users_deleted_at;
DROP INDEX IF EXISTS auth.idx_users_stripe_customer_id;
DROP INDEX IF EXISTS auth.idx_users_subscription_tier;
DROP INDEX IF EXISTS auth.idx_users_username;
DROP INDEX IF EXISTS auth.idx_users_email;

-- ============================================
-- DROP TABLES (in dependency order - children first)
-- ============================================

-- Analytics schema tables
DROP TABLE IF EXISTS analytics.daily_stats;
DROP TABLE IF EXISTS analytics.usage_logs;
DROP TABLE IF EXISTS analytics.project_views;

-- Marketplace schema tables
DROP TABLE IF EXISTS marketplace.payouts;
DROP TABLE IF EXISTS marketplace.creator_earnings;
DROP TABLE IF EXISTS marketplace.template_purchases;

-- Social schema tables
DROP TABLE IF EXISTS social.project_comments;
DROP TABLE IF EXISTS social.project_likes;
DROP TABLE IF EXISTS social.follows;

-- Content schema tables (drop in dependency order)
DROP TABLE IF EXISTS content.page_items;
DROP TABLE IF EXISTS content.pages;
DROP TABLE IF EXISTS content.items;
DROP TABLE IF EXISTS content.projects;

-- Auth schema tables (drop in dependency order)
DROP TABLE IF EXISTS auth.password_resets;
DROP TABLE IF EXISTS auth.refresh_tokens;
DROP TABLE IF EXISTS auth.users;