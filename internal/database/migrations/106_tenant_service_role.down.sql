-- Migration: 106_tenant_service_role (rollback)
-- Description: Remove tenant_service role

-- Revoke role membership
REVOKE tenant_service FROM fluxbase_app;

-- Revoke sequence privileges
REVOKE USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public FROM tenant_service;
REVOKE USAGE, SELECT ON ALL SEQUENCES IN SCHEMA ai FROM tenant_service;
REVOKE USAGE, SELECT ON ALL SEQUENCES IN SCHEMA jobs FROM tenant_service;
REVOKE USAGE, SELECT ON ALL SEQUENCES IN SCHEMA functions FROM tenant_service;
REVOKE USAGE, SELECT ON ALL SEQUENCES IN SCHEMA storage FROM tenant_service;
REVOKE USAGE, SELECT ON ALL SEQUENCES IN SCHEMA auth FROM tenant_service;

-- Revoke table privileges
REVOKE SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public FROM tenant_service;
REVOKE SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA ai FROM tenant_service;
REVOKE SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA jobs FROM tenant_service;
REVOKE SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA functions FROM tenant_service;
REVOKE SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA storage FROM tenant_service;
REVOKE SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA auth FROM tenant_service;

-- Revoke schema usage
REVOKE USAGE ON SCHEMA public FROM tenant_service;
REVOKE USAGE ON SCHEMA ai FROM tenant_service;
REVOKE USAGE ON SCHEMA jobs FROM tenant_service;
REVOKE USAGE ON SCHEMA functions FROM tenant_service;
REVOKE USAGE ON SCHEMA storage FROM tenant_service;
REVOKE USAGE ON SCHEMA auth FROM tenant_service;

-- Drop the role
DROP ROLE IF EXISTS tenant_service;
