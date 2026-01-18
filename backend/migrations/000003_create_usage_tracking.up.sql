-- 000003_add_usage_tracking.up.sql
-- Add usage_tracking table to content schema for freemium limits

CREATE TABLE IF NOT EXISTS content.usage_tracking (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    
    -- Usage period (monthly)
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    
    -- Counters
    items_processed INTEGER NOT NULL DEFAULT 0,
    items_limit INTEGER, -- NULL means unlimited (Pro users)
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure one record per user per period
    UNIQUE(user_id, period_start)
);

-- Create indexes for usage_tracking
CREATE INDEX IF NOT EXISTS idx_usage_tracking_user_id ON content.usage_tracking(user_id);
CREATE INDEX IF NOT EXISTS idx_usage_tracking_period ON content.usage_tracking(period_start, period_end);
CREATE INDEX IF NOT EXISTS idx_usage_tracking_user_current ON content.usage_tracking(user_id, period_start, period_end);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION content.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_items_updated_at') THEN
        CREATE TRIGGER update_items_updated_at BEFORE UPDATE ON content.items
            FOR EACH ROW EXECUTE FUNCTION content.update_updated_at_column();
    END IF;
END;
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_usage_tracking_updated_at') THEN
        CREATE TRIGGER update_usage_tracking_updated_at BEFORE UPDATE ON content.usage_tracking
            FOR EACH ROW EXECUTE FUNCTION content.update_updated_at_column();
    END IF;
END;
$$;
