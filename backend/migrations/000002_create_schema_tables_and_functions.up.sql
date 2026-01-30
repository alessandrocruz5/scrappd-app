-- 000002_create_schema_tables_and_functions.up.sql
-- ============================================
-- AUTH SCHEMA - Users & Authentication
-- ============================================

CREATE TABLE IF NOT EXISTS auth.users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100),
    password_hash VARCHAR(255) NOT NULL,
    profile_image_url TEXT,
    bio TEXT,
    
    -- Subscription & tier
    subscription_tier VARCHAR(20) DEFAULT 'free' CHECK (subscription_tier IN ('free', 'pro', 'creator')),
    stripe_customer_id VARCHAR(255) UNIQUE,
    subscription_status VARCHAR(20) DEFAULT 'active' CHECK (subscription_status IN ('active', 'cancelled', 'expired', 'trialing')),
    subscription_expires_at TIMESTAMP WITH TIME ZONE,
    
    -- Usage limits (resets monthly)
    monthly_bg_removals_used INTEGER DEFAULT 0,
    monthly_bg_removals_limit INTEGER DEFAULT 50,
    monthly_storage_used_mb DECIMAL(10, 2) DEFAULT 0,
    monthly_storage_limit_mb INTEGER DEFAULT 500,
    
    -- Social stats
    follower_count INTEGER DEFAULT 0,
    following_count INTEGER DEFAULT 0,
    is_verified BOOLEAN DEFAULT FALSE,
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS auth.refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS auth.password_resets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- CONTENT SCHEMA - Projects, Pages, Items
-- ============================================

CREATE TABLE IF NOT EXISTS content.projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    cover_image_url TEXT,
    
    -- Privacy & sharing
    visibility VARCHAR(20) DEFAULT 'private' CHECK (visibility IN ('private', 'unlisted', 'public')),
    is_template BOOLEAN DEFAULT FALSE,
    template_price DECIMAL(10, 2),
    
    -- Stats
    view_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    fork_count INTEGER DEFAULT 0,
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS content.pages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES content.projects(id) ON DELETE CASCADE,
    page_number INTEGER NOT NULL,
    title VARCHAR(200),
    
    -- Canvas configuration
    canvas_width INTEGER DEFAULT 1080,
    canvas_height INTEGER DEFAULT 1920,
    background_color VARCHAR(7) DEFAULT '#FFFFFF',
    background_image_url TEXT,
    background_pattern VARCHAR(50),
    
    -- Template data
    layout_template JSONB,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(project_id, page_number)
);

CREATE TABLE IF NOT EXISTS content.items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    
    -- Original image
    original_image_key VARCHAR(500) NOT NULL,
    original_image_url TEXT NOT NULL,
    original_file_size_bytes BIGINT,
    original_width INTEGER,
    original_height INTEGER,
    
    -- Processed image
    processed_image_key VARCHAR(500),
    processed_image_url TEXT,
    processed_file_size_bytes BIGINT,
    
    -- ML Processing
    processing_status VARCHAR(20) DEFAULT 'pending' CHECK (
        processing_status IN ('pending', 'processing', 'completed', 'failed')
    ),
    ml_model_version VARCHAR(50),
    processing_started_at TIMESTAMP WITH TIME ZONE,
    processing_completed_at TIMESTAMP WITH TIME ZONE,
    processing_error TEXT,
    
    -- Metadata
    mime_type VARCHAR(50),
    item_name VARCHAR(200),
    item_category VARCHAR(50),
    tags TEXT[],
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS content.page_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    page_id UUID NOT NULL REFERENCES content.pages(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES content.items(id) ON DELETE CASCADE,
    
    -- Position & transformation
    position_x DECIMAL(10, 2) NOT NULL,
    position_y DECIMAL(10, 2) NOT NULL,
    width DECIMAL(10, 2) NOT NULL,
    height DECIMAL(10, 2) NOT NULL,
    rotation DECIMAL(6, 2) DEFAULT 0,
    z_index INTEGER DEFAULT 0,
    opacity DECIMAL(3, 2) DEFAULT 1.0,
    
    -- Filters/effects
    filters JSONB,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- SOCIAL SCHEMA - Follows, Likes, Comments
-- ============================================

CREATE TABLE IF NOT EXISTS social.follows (
    follower_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    following_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_id, following_id),
    CHECK (follower_id != following_id)
);

CREATE TABLE IF NOT EXISTS social.project_likes (
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES content.projects(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, project_id)
);

CREATE TABLE IF NOT EXISTS social.project_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES content.projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    parent_comment_id UUID REFERENCES social.project_comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- ============================================
-- MARKETPLACE SCHEMA - Templates & Earnings
-- ============================================

CREATE TABLE IF NOT EXISTS marketplace.template_purchases (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    buyer_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES content.projects(id) ON DELETE CASCADE,
    seller_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    
    amount_paid DECIMAL(10, 2) NOT NULL,
    stripe_payment_intent_id VARCHAR(255),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS marketplace.creator_earnings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    
    total_earnings DECIMAL(10, 2) DEFAULT 0,
    withdrawn_amount DECIMAL(10, 2) DEFAULT 0,
    pending_amount DECIMAL(10, 2) DEFAULT 0,
    
    stripe_connect_account_id VARCHAR(255),
    
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS marketplace.payouts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    
    amount DECIMAL(10, 2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    stripe_payout_id VARCHAR(255),
    
    requested_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    failure_reason TEXT
);

-- ============================================
-- ANALYTICS SCHEMA - Views, Logs, Metrics
-- ============================================

CREATE TABLE IF NOT EXISTS analytics.project_views (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES content.projects(id) ON DELETE CASCADE,
    viewer_id UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    ip_address INET,
    user_agent TEXT,
    viewed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS analytics.usage_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    action_type VARCHAR(50) NOT NULL,
    resource_id UUID,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS analytics.daily_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stat_date DATE NOT NULL,
    
    -- User stats
    new_users INTEGER DEFAULT 0,
    active_users INTEGER DEFAULT 0,
    
    -- Content stats
    new_projects INTEGER DEFAULT 0,
    new_items INTEGER DEFAULT 0,
    bg_removals_processed INTEGER DEFAULT 0,
    
    -- Revenue stats
    template_sales_count INTEGER DEFAULT 0,
    template_sales_revenue DECIMAL(10, 2) DEFAULT 0,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(stat_date)
);

-- ============================================
-- INDEXES FOR PERFORMANCE
-- ============================================

-- Auth schema indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON auth.users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON auth.users(username);
CREATE INDEX IF NOT EXISTS idx_users_subscription_tier ON auth.users(subscription_tier);
CREATE INDEX IF NOT EXISTS idx_users_stripe_customer_id ON auth.users(stripe_customer_id);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON auth.users(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON auth.refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON auth.refresh_tokens(expires_at);

-- Content schema indexes
CREATE INDEX IF NOT EXISTS idx_projects_user_id ON content.projects(user_id);
CREATE INDEX IF NOT EXISTS idx_projects_visibility ON content.projects(visibility);
CREATE INDEX IF NOT EXISTS idx_projects_is_template ON content.projects(is_template) WHERE is_template = TRUE;
CREATE INDEX IF NOT EXISTS idx_projects_created_at ON content.projects(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_projects_deleted_at ON content.projects(deleted_at) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_pages_project_id ON content.pages(project_id);

CREATE INDEX IF NOT EXISTS idx_items_user_id ON content.items(user_id);
CREATE INDEX IF NOT EXISTS idx_items_processing_status ON content.items(processing_status);
CREATE INDEX IF NOT EXISTS idx_items_created_at ON content.items(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_items_deleted_at ON content.items(deleted_at) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_page_items_page_id ON content.page_items(page_id);
CREATE INDEX IF NOT EXISTS idx_page_items_item_id ON content.page_items(item_id);
CREATE INDEX IF NOT EXISTS idx_page_items_z_index ON content.page_items(z_index);

-- Social schema indexes
CREATE INDEX IF NOT EXISTS idx_follows_follower_id ON social.follows(follower_id);
CREATE INDEX IF NOT EXISTS idx_follows_following_id ON social.follows(following_id);
CREATE INDEX IF NOT EXISTS idx_project_likes_project_id ON social.project_likes(project_id);
CREATE INDEX IF NOT EXISTS idx_project_comments_project_id ON social.project_comments(project_id);
CREATE INDEX IF NOT EXISTS idx_project_comments_parent_id ON social.project_comments(parent_comment_id);

-- Marketplace schema indexes
CREATE INDEX IF NOT EXISTS idx_template_purchases_buyer_id ON marketplace.template_purchases(buyer_id);
CREATE INDEX IF NOT EXISTS idx_template_purchases_seller_id ON marketplace.template_purchases(seller_id);
CREATE INDEX IF NOT EXISTS idx_template_purchases_project_id ON marketplace.template_purchases(project_id);
CREATE INDEX IF NOT EXISTS idx_payouts_user_id ON marketplace.payouts(user_id);
CREATE INDEX IF NOT EXISTS idx_payouts_status ON marketplace.payouts(status);

-- Analytics schema indexes
CREATE INDEX IF NOT EXISTS idx_project_views_project_id ON analytics.project_views(project_id);
CREATE INDEX IF NOT EXISTS idx_project_views_viewed_at ON analytics.project_views(viewed_at DESC);
CREATE INDEX IF NOT EXISTS idx_usage_logs_user_id ON analytics.usage_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_usage_logs_action_type ON analytics.usage_logs(action_type);
CREATE INDEX IF NOT EXISTS idx_usage_logs_created_at ON analytics.usage_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_daily_stats_stat_date ON analytics.daily_stats(stat_date DESC);

-- ============================================
-- TRIGGERS FOR AUTO-UPDATE TIMESTAMPS
-- ============================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Auth triggers
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_users_updated_at') THEN
        CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON auth.users
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END;
$$;

-- Content triggers
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_projects_updated_at') THEN
        CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON content.projects
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END;
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_pages_updated_at') THEN
        CREATE TRIGGER update_pages_updated_at BEFORE UPDATE ON content.pages
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END;
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_items_updated_at') THEN
        CREATE TRIGGER update_items_updated_at BEFORE UPDATE ON content.items
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END;
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_page_items_updated_at') THEN
        CREATE TRIGGER update_page_items_updated_at BEFORE UPDATE ON content.page_items
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END;
$$;

-- Social triggers
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_comments_updated_at') THEN
        CREATE TRIGGER update_comments_updated_at BEFORE UPDATE ON social.project_comments
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END;
$$;

-- Marketplace triggers
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_earnings_updated_at') THEN
        CREATE TRIGGER update_earnings_updated_at BEFORE UPDATE ON marketplace.creator_earnings
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END;
$$;

-- ============================================
-- HELPER FUNCTIONS
-- ============================================

-- Function to increment follower counts
CREATE OR REPLACE FUNCTION increment_follower_counts()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE auth.users SET follower_count = follower_count + 1 WHERE id = NEW.following_id;
    UPDATE auth.users SET following_count = following_count + 1 WHERE id = NEW.follower_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION decrement_follower_counts()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE auth.users SET follower_count = follower_count - 1 WHERE id = OLD.following_id;
    UPDATE auth.users SET following_count = following_count - 1 WHERE id = OLD.follower_id;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'follow_insert_trigger') THEN
        CREATE TRIGGER follow_insert_trigger AFTER INSERT ON social.follows
            FOR EACH ROW EXECUTE FUNCTION increment_follower_counts();
    END IF;
END;
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'follow_delete_trigger') THEN
        CREATE TRIGGER follow_delete_trigger AFTER DELETE ON social.follows
            FOR EACH ROW EXECUTE FUNCTION decrement_follower_counts();
    END IF;
END;
$$;

-- Function to increment project like counts
CREATE OR REPLACE FUNCTION increment_like_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE content.projects SET like_count = like_count + 1 WHERE id = NEW.project_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION decrement_like_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE content.projects SET like_count = like_count - 1 WHERE id = OLD.project_id;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'like_insert_trigger') THEN
        CREATE TRIGGER like_insert_trigger AFTER INSERT ON social.project_likes
            FOR EACH ROW EXECUTE FUNCTION increment_like_count();
    END IF;
END;
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'like_delete_trigger') THEN
        CREATE TRIGGER like_delete_trigger AFTER DELETE ON social.project_likes
            FOR EACH ROW EXECUTE FUNCTION decrement_like_count();
    END IF;
END;
$$;
