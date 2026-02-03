-- 000001_create_db_schemas.up.sql
-- ============================================
-- DATABASE SETUP (Cloud SQL Compatible)
-- ============================================

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================
-- CREATE SCHEMAS
-- ============================================

CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS content;
CREATE SCHEMA IF NOT EXISTS social;
CREATE SCHEMA IF NOT EXISTS marketplace;
CREATE SCHEMA IF NOT EXISTS analytics;

-- Set search path (wrapped in exception handler for Cloud SQL)
DO $$
BEGIN
    EXECUTE format('ALTER DATABASE %I SET search_path TO auth, content, social, marketplace, analytics, public', current_database());
EXCEPTION
    WHEN insufficient_privilege THEN
        RAISE NOTICE 'Cannot alter database search_path - skipping';
END
$$;

-- ============================================
-- GRANT PERMISSIONS TO scrappd_user
-- (This is the user we created in Cloud SQL console)
-- ============================================

-- Grant schema usage
GRANT USAGE ON SCHEMA auth TO scrappd_user;
GRANT USAGE ON SCHEMA content TO scrappd_user;
GRANT USAGE ON SCHEMA social TO scrappd_user;
GRANT USAGE ON SCHEMA marketplace TO scrappd_user;
GRANT USAGE ON SCHEMA analytics TO scrappd_user;

-- Grant ALL privileges on tables
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA auth TO scrappd_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA content TO scrappd_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA social TO scrappd_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA marketplace TO scrappd_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA analytics TO scrappd_user;

-- Grant sequence privileges (for ID generation)
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA auth TO scrappd_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA content TO scrappd_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA social TO scrappd_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA marketplace TO scrappd_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA analytics TO scrappd_user;

-- Grant function execution
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA auth TO scrappd_user;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA content TO scrappd_user;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA social TO scrappd_user;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA marketplace TO scrappd_user;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA analytics TO scrappd_user;

-- Set default privileges for future tables
ALTER DEFAULT PRIVILEGES IN SCHEMA auth GRANT ALL ON TABLES TO scrappd_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA content GRANT ALL ON TABLES TO scrappd_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA social GRANT ALL ON TABLES TO scrappd_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA marketplace GRANT ALL ON TABLES TO scrappd_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA analytics GRANT ALL ON TABLES TO scrappd_user;

ALTER DEFAULT PRIVILEGES IN SCHEMA auth GRANT ALL ON SEQUENCES TO scrappd_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA content GRANT ALL ON SEQUENCES TO scrappd_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA social GRANT ALL ON SEQUENCES TO scrappd_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA marketplace GRANT ALL ON SEQUENCES TO scrappd_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA analytics GRANT ALL ON SEQUENCES TO scrappd_user;