-- This script runs automatically when PostgreSQL container starts for the first time
DO
$do$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_database WHERE datname = 'scrappd'
   ) THEN
      CREATE DATABASE scrappd;
   END IF;
END
$do$;

\c scrappd

-- Create roles
CREATE ROLE scrappd_app WITH LOGIN PASSWORD 'scrappd-go';
CREATE ROLE scrappd_readonly WITH LOGIN PASSWORD 'scrappd-viewer';
CREATE ROLE scrappd_admin WITH LOGIN PASSWORD 'scrappd-admin' SUPERUSER;

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create schemas
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS content;
CREATE SCHEMA IF NOT EXISTS social;
CREATE SCHEMA IF NOT EXISTS marketplace;
CREATE SCHEMA IF NOT EXISTS analytics;

-- Grant permissions to scrappd_app
GRANT USAGE ON SCHEMA auth TO scrappd_app;
GRANT USAGE ON SCHEMA content TO scrappd_app;
GRANT USAGE ON SCHEMA social TO scrappd_app;
GRANT USAGE ON SCHEMA marketplace TO scrappd_app;
GRANT USAGE ON SCHEMA analytics TO scrappd_app;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA auth TO scrappd_app;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA content TO scrappd_app;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA social TO scrappd_app;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA marketplace TO scrappd_app;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA analytics TO scrappd_app;

GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA auth TO scrappd_app;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA content TO scrappd_app;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA social TO scrappd_app;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA marketplace TO scrappd_app;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA analytics TO scrappd_app;

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

-- Grant read-only permissions
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

-- Set search path for database
ALTER DATABASE scrappd SET search_path TO auth, content, social, marketplace, analytics, public;