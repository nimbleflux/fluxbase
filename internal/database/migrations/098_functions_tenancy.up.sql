--
-- MULTI-TENANCY: RLS HELPER FUNCTIONS
-- Creates PostgreSQL functions for tenant context and membership checking
-- Note: References public.tenants and dashboard.users which get moved to platform schema in later migrations
--

-- Get current tenant ID from JWT claims or session context
-- Falls back to default tenant for backward compatibility
CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS UUID AS $$
DECLARE
    claims JSONB;
    tid UUID;
    default_id UUID;
BEGIN
    -- Try JWT claims first
    BEGIN
        claims := auth.jwt();
        IF claims ? 'tenant_id' AND (claims->>'tenant_id') IS NOT NULL THEN
            -- Validate it's a valid UUID
            tid := (claims->>'tenant_id')::UUID;
            -- Verify tenant exists and is not deleted
            IF EXISTS (SELECT 1 FROM tenants WHERE id = tid AND deleted_at IS NULL) THEN
                RETURN tid;
            END IF;
        END IF;
    EXCEPTION WHEN OTHERS THEN
        NULL;
    END;
    
    -- Fall back to session variable (set by middleware)
    BEGIN
        tid := NULL;
        tid := current_setting('app.current_tenant_id', true)::UUID;
        IF tid IS NOT NULL THEN
            -- Verify tenant exists and is not deleted
            IF EXISTS (SELECT 1 FROM tenants WHERE id = tid AND deleted_at IS NULL) THEN
                RETURN tid;
            END IF;
        END IF;
    EXCEPTION WHEN OTHERS THEN
        NULL;
    END;
    
    -- Return default tenant for backward compatibility
    SELECT id INTO default_id FROM tenants WHERE is_default = true AND deleted_at IS NULL LIMIT 1;
    RETURN default_id;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION current_tenant_id() IS 
'Returns the current tenant ID from JWT claims, session variable, or default tenant';

-- Check if user has specific role in tenant
CREATE OR REPLACE FUNCTION user_has_tenant_role(
    p_user_id UUID,
    p_tenant_id UUID,
    p_role TEXT
) RETURNS BOOLEAN AS $$
BEGIN
    IF p_user_id IS NULL OR p_tenant_id IS NULL THEN
        RETURN false;
    END IF;
    
    RETURN EXISTS (
        SELECT 1 FROM tenant_memberships tm
        INNER JOIN tenants t ON t.id = tm.tenant_id
        WHERE tm.user_id = p_user_id
        AND tm.tenant_id = p_tenant_id
        AND tm.role = p_role
        AND t.deleted_at IS NULL
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION user_has_tenant_role(UUID, UUID, TEXT) IS
'Checks if a user has a specific role in a tenant. SECURITY DEFINER to bypass RLS.';

-- Check if user is instance admin (dashboard.users at this point, renamed to platform in migration 103)
CREATE OR REPLACE FUNCTION is_instance_admin(p_user_id UUID) RETURNS BOOLEAN AS $$
BEGIN
    IF p_user_id IS NULL THEN
        RETURN false;
    END IF;
    
    RETURN EXISTS (
        SELECT 1 FROM dashboard.users du
        WHERE du.id = p_user_id
        AND du.role = 'instance_admin'
        AND (du.deleted_at IS NULL OR du.is_active = true)
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION is_instance_admin(UUID) IS
'Checks if a user is an instance-level admin. SECURITY DEFINER to bypass RLS.';

-- Get user's effective tenant role for current tenant
CREATE OR REPLACE FUNCTION current_tenant_role() RETURNS TEXT AS $$
DECLARE
    uid UUID;
    tid UUID;
BEGIN
    uid := auth.uid();
    tid := current_tenant_id();
    
    IF uid IS NULL OR tid IS NULL THEN
        RETURN NULL;
    END IF;
    
    RETURN (
        SELECT tm.role FROM tenant_memberships tm
        INNER JOIN tenants t ON t.id = tm.tenant_id
        WHERE tm.user_id = uid 
        AND tm.tenant_id = tid
        AND t.deleted_at IS NULL
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION current_tenant_role() IS
'Return the current user role in the current tenant';

-- Get all tenant IDs for a user
CREATE OR REPLACE FUNCTION user_tenant_ids(p_user_id UUID) RETURNS UUID[] AS $$
BEGIN
    IF p_user_id IS NULL THEN
        RETURN ARRAY[]::UUID[];
    END IF;
    
    RETURN ARRAY(
        SELECT tm.tenant_id FROM tenant_memberships tm
        INNER JOIN tenants t ON t.id = tm.tenant_id
        WHERE tm.user_id = p_user_id
        AND t.deleted_at IS NULL
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION user_tenant_ids(UUID) IS
'Returns all tenant IDs that a user is a member of';

-- Check if user is member of tenant (any role)
CREATE OR REPLACE FUNCTION user_is_tenant_member(
    p_user_id UUID,
    p_tenant_id UUID
) RETURNS BOOLEAN AS $$
BEGIN
    IF p_user_id IS NULL OR p_tenant_id IS NULL THEN
        RETURN false;
    END IF;
    
    RETURN EXISTS (
        SELECT 1 FROM tenant_memberships tm
        INNER JOIN tenants t ON t.id = tm.tenant_id
        WHERE tm.user_id = p_user_id
        AND tm.tenant_id = p_tenant_id
        AND t.deleted_at IS NULL
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION user_is_tenant_member(UUID, UUID) IS
'Checks if a user is a member of a tenant (any role). SECURITY DEFINER to bypass RLS.';

-- Get tenant info by slug
CREATE OR REPLACE FUNCTION get_tenant_by_slug(p_slug TEXT) RETURNS TABLE (
    id UUID,
    name TEXT,
    is_default BOOLEAN,
    metadata JSONB,
    created_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT t.id, t.name, t.is_default, t.metadata, t.created_at
    FROM tenants t
    WHERE t.slug = p_slug
    AND t.deleted_at IS NULL;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION get_tenant_by_slug(TEXT) IS
'Gets tenant information by slug. SECURITY DEFINER to bypass RLS.';

-- Get tenant info by ID
CREATE OR REPLACE FUNCTION get_tenant_by_id(p_tenant_id UUID) RETURNS TABLE (
    id UUID,
    slug TEXT,
    name TEXT,
    is_default BOOLEAN,
    metadata JSONB,
    created_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT t.id, t.slug, t.name, t.is_default, t.metadata, t.created_at
    FROM tenants t
    WHERE t.id = p_tenant_id
    AND t.deleted_at IS NULL;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION get_tenant_by_id(UUID) IS
'Gets tenant information by ID. SECURITY DEFINER to bypass RLS.';
