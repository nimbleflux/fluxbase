--
-- MULTI-TENANCY: FIX INSTANCE ADMIN TENANT MODE
-- When an instance admin selects a tenant (acting as tenant admin),
-- they should NOT have instance admin privileges for that session.
-- This ensures proper tenant isolation.
--

-- Update is_instance_admin to check for tenant context
CREATE OR REPLACE FUNCTION is_instance_admin(p_user_id UUID) RETURNS BOOLEAN AS $$
DECLARE
    tenant_context UUID;
BEGIN
    IF p_user_id IS NULL THEN
        RETURN false;
    END IF;

    -- Check if a tenant context is set (acting as tenant admin mode)
    -- If so, the user is NOT acting as an instance admin for this session
    BEGIN
        tenant_context := current_setting('app.current_tenant_id', true)::UUID;
        IF tenant_context IS NOT NULL THEN
            -- User is in tenant admin mode, not instance admin mode
            RETURN false;
        END IF;
    EXCEPTION WHEN OTHERS THEN
        -- No tenant context set, continue with instance admin check
        NULL;
    END;

    -- Also check JWT claims for tenant context
    BEGIN
        DECLARE
            claims JSONB;
        BEGIN
            claims := auth.jwt();
            IF claims ? 'tenant_id' AND (claims->>'tenant_id') IS NOT NULL THEN
                -- User is in tenant admin mode via JWT
                RETURN false;
            END IF;
        END;
    EXCEPTION WHEN OTHERS THEN
        NULL;
    END;

    -- No tenant context, check if user is actually an instance admin
    RETURN EXISTS (
        SELECT 1 FROM platform.users pu
        WHERE pu.id = p_user_id
        AND pu.role = 'instance_admin'
        AND pu.deleted_at IS NULL
        AND pu.is_active = true
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER STABLE;

COMMENT ON FUNCTION is_instance_admin(UUID) IS 'Returns true if user is an instance admin AND not in tenant admin mode. When acting as tenant admin, returns false to ensure tenant isolation.';

-- Add a helper function to check if user is in tenant admin mode
CREATE OR REPLACE FUNCTION is_tenant_admin_mode() RETURNS BOOLEAN AS $$
BEGIN
    -- Check session variable
    BEGIN
        IF current_setting('app.current_tenant_id', true) IS NOT NULL THEN
            RETURN true;
        END IF;
    EXCEPTION WHEN OTHERS THEN
        NULL;
    END;

    -- Check JWT claims
    BEGIN
        DECLARE
            claims JSONB;
        BEGIN
            claims := auth.jwt();
            IF claims ? 'tenant_id' AND (claims->>'tenant_id') IS NOT NULL THEN
                RETURN true;
            END IF;
        END;
    EXCEPTION WHEN OTHERS THEN
        NULL;
    END;

    RETURN false;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER STABLE;

COMMENT ON FUNCTION is_tenant_admin_mode() IS 'Returns true if the current session is in tenant admin mode (acting as a tenant admin).';

DO $$
BEGIN
    RAISE NOTICE 'Migration 117 complete: Updated is_instance_admin to respect tenant admin mode';
END $$;
