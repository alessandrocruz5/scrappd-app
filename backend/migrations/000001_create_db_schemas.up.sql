-- 000001_create_db_schemas.up.sql
-- ============================================
-- DATABASE SETUP
-- ============================================

-- Connect to scrappd database
-- \c scrappd

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================
-- CREATE ROLES (Principle of Least Privilege)
-- ============================================

-- Application role (your Go backend connects as this)
CREATE ROLE scrappd_app WITH LOGIN PASSWORD 'scrappd-go';

-- Read-only role (for analytics, reporting, etc.)
CREATE ROLE scrappd_readonly WITH LOGIN PASSWORD 'scrappd-viewer';

-- Admin role (for migrations, maintenance)
CREATE ROLE scrappd_admin WITH LOGIN PASSWORD 'scrappd-admin' SUPERUSER;

-- ============================================
-- CREATE SCHEMAS
-- ============================================

-- SELECT has_database_privilege('apa', 'CONNECT');

-- SELECT datname 
-- FROM pg_database 
-- WHERE has_database_privilege(datname, 'CONNECT')
-- ORDER BY datname;

CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS content;
CREATE SCHEMA IF NOT EXISTS social;
CREATE SCHEMA IF NOT EXISTS marketplace;
CREATE SCHEMA IF NOT EXISTS analytics;

-- Set search path for the database
ALTER DATABASE scrappd SET search_path TO auth, content, social, marketplace, analytics, public;

-- ============================================
-- GRANT APPROPRIATE PERMISSIONS
-- ============================================

-- Grant schema usage to app role (NOT PUBLIC!)
GRANT USAGE ON SCHEMA auth TO scrappd_app;
GRANT USAGE ON SCHEMA content TO scrappd_app;
GRANT USAGE ON SCHEMA social TO scrappd_app;
GRANT USAGE ON SCHEMA marketplace TO scrappd_app;
GRANT USAGE ON SCHEMA analytics TO scrappd_app;

-- Grant ALL privileges on tables to app role
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA auth TO scrappd_app;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA content TO scrappd_app;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA social TO scrappd_app;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA marketplace TO scrappd_app;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA analytics TO scrappd_app;

-- Grant sequence privileges (for ID generation)
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA auth TO scrappd_app;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA content TO scrappd_app;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA social TO scrappd_app;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA marketplace TO scrappd_app;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA analytics TO scrappd_app;

-- Grant function execution
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA auth TO scrappd_app;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA content TO scrappd_app;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA social TO scrappd_app;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA marketplace TO scrappd_app;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA analytics TO scrappd_app;

-- Set default privileges for future tables
ALTER DEFAULT PRIVILEGES IN SCHEMA auth GRANT ALL ON TABLES TO scrappd_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA content GRANT ALL ON TABLES TO scrappd_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA social GRANT ALL ON TABLES TO scrappd_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA marketplace GRANT ALL ON TABLES TO scrappd_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA analytics GRANT ALL ON TABLES TO scrappd_app;

ALTER DEFAULT PRIVILEGES IN SCHEMA auth GRANT ALL ON SEQUENCES TO scrappd_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA content GRANT ALL ON SEQUENCES TO scrappd_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA social GRANT ALL ON SEQUENCES TO scrappd_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA marketplace GRANT ALL ON SEQUENCES TO scrappd_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA analytics GRANT ALL ON SEQUENCES TO scrappd_app;

-- ============================================
-- READ-ONLY ROLE (for analytics/reporting)
-- ============================================

GRANT USAGE ON SCHEMA auth TO scrappd_readonly;
GRANT USAGE ON SCHEMA content TO scrappd_readonly;
GRANT USAGE ON SCHEMA social TO scrappd_readonly;
GRANT USAGE ON SCHEMA marketplace TO scrappd_readonly;
GRANT USAGE ON SCHEMA analytics TO scrappd_readonly;

GRANT SELECT ON ALL TABLES IN SCHEMA auth TO scrappd_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA content TO scrappd_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA social TO scrappd_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA marketplace TO scrappd_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA analytics TO scrappd_readonly;

ALTER DEFAULT PRIVILEGES IN SCHEMA auth GRANT SELECT ON TABLES TO scrappd_readonly;
ALTER DEFAULT PRIVILEGES IN SCHEMA content GRANT SELECT ON TABLES TO scrappd_readonly;
ALTER DEFAULT PRIVILEGES IN SCHEMA social GRANT SELECT ON TABLES TO scrappd_readonly;
ALTER DEFAULT PRIVILEGES IN SCHEMA marketplace GRANT SELECT ON TABLES TO scrappd_readonly;
ALTER DEFAULT PRIVILEGES IN SCHEMA analytics GRANT SELECT ON TABLES TO scrappd_readonly;