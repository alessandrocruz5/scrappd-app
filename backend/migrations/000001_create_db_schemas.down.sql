-- 000001_create_db_schemas.down.sql
-- ============================================
-- ROLLBACK DATABASE SETUP
-- ============================================

-- Revoke permissions and drop default privileges
-- (in reverse order of creation)

-- Revoke from read-only role
REVOKE SELECT ON ALL TABLES IN SCHEMA analytics FROM scrappd_readonly;
REVOKE SELECT ON ALL TABLES IN SCHEMA marketplace FROM scrappd_readonly;
REVOKE SELECT ON ALL TABLES IN SCHEMA social FROM scrappd_readonly;
REVOKE SELECT ON ALL TABLES IN SCHEMA content FROM scrappd_readonly;
REVOKE SELECT ON ALL TABLES IN SCHEMA auth FROM scrappd_readonly;

REVOKE USAGE ON SCHEMA analytics FROM scrappd_readonly;
REVOKE USAGE ON SCHEMA marketplace FROM scrappd_readonly;
REVOKE USAGE ON SCHEMA social FROM scrappd_readonly;
REVOKE USAGE ON SCHEMA content FROM scrappd_readonly;
REVOKE USAGE ON SCHEMA auth FROM scrappd_readonly;

-- Revoke from app role
REVOKE ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA analytics FROM scrappd_app;
REVOKE ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA marketplace FROM scrappd_app;
REVOKE ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA social FROM scrappd_app;
REVOKE ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA content FROM scrappd_app;
REVOKE ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA auth FROM scrappd_app;

REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA analytics FROM scrappd_app;
REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA marketplace FROM scrappd_app;
REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA social FROM scrappd_app;
REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA content FROM scrappd_app;
REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA auth FROM scrappd_app;

REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA analytics FROM scrappd_app;
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA marketplace FROM scrappd_app;
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA social FROM scrappd_app;
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA content FROM scrappd_app;
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA auth FROM scrappd_app;

REVOKE USAGE ON SCHEMA analytics FROM scrappd_app;
REVOKE USAGE ON SCHEMA marketplace FROM scrappd_app;
REVOKE USAGE ON SCHEMA social FROM scrappd_app;
REVOKE USAGE ON SCHEMA content FROM scrappd_app;
REVOKE USAGE ON SCHEMA auth FROM scrappd_app;

-- Drop schemas (CASCADE will drop all objects within)
DROP SCHEMA IF EXISTS analytics CASCADE;
DROP SCHEMA IF EXISTS marketplace CASCADE;
DROP SCHEMA IF EXISTS social CASCADE;
DROP SCHEMA IF EXISTS content CASCADE;
DROP SCHEMA IF EXISTS auth CASCADE;

-- Drop roles
DROP ROLE IF EXISTS scrappd_admin;
DROP ROLE IF EXISTS scrappd_readonly;
DROP ROLE IF EXISTS scrappd_app;

-- Drop extensions (optional - usually kept)
-- DROP EXTENSION IF EXISTS "pgcrypto";
-- DROP EXTENSION IF EXISTS "uuid-ossp";