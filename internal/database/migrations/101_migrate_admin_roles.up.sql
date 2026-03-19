--
-- MULTI-TENANCY: MIGRATE ADMIN ROLES
-- Creates default tenant and migrates existing admin users to new role model
--
-- Migration Strategy:
--   - First setup admin → instance_admin (global admin)
--   - Other dashboard_admin users → tenant_admin of default tenant
--   - Note: tenant_memberships are for auth.users, not dashboard.users
--    Dashboard roles (instance_admin, tenant_admin) are separate from tenant memberships
--

DO $$
DECLARE
    default_tenant_id UUID;
    first_admin_id UUID;
    admin_count INTEGER;
BEGIN
    -- Check if default tenant already exists in platform schema
    SELECT id INTO default_tenant_id FROM platform.tenants WHERE is_default = true LIMIT 1;

    IF default_tenant_id IS NULL THEN
        -- Step 1: Create default tenant
        INSERT INTO platform.tenants (id, slug, name, is_default, metadata, created_at)
        VALUES (
            gen_random_uuid(),
            'default',
            'Default Tenant',
            true,
            '{"description": "Default tenant for backward compatibility - all existing data belongs to this tenant", "migrated": true}'::jsonb,
            NOW()
        )
        RETURNING id INTO default_tenant_id;

        RAISE NOTICE 'Created default tenant with ID: %', default_tenant_id;
    ELSE
        RAISE NOTICE 'Default tenant already exists with ID: %', default_tenant_id;
    END IF;

    -- Step 2: Get first admin (setup admin - oldest dashboard user)
    SELECT id INTO first_admin_id
    FROM dashboard.users
    WHERE deleted_at IS NULL AND is_active = true
    ORDER BY created_at ASC
    LIMIT 1;

    IF first_admin_id IS NOT NULL THEN
        -- Step 3: Migrate first admin to instance_admin
        UPDATE dashboard.users
        SET role = 'instance_admin',
            updated_at = NOW()
        WHERE id = first_admin_id
        AND role != 'instance_admin';

        RAISE NOTICE 'Migrated first admin % to instance_admin', first_admin_id;
    END IF;

    -- Step 4: Migrate other dashboard_admin to tenant_admin role
    -- Note: This updates dashboard.users role. Not tenant_memberships
    -- tenant_memberships is for auth.users (application users), not dashboard.users
    SELECT COUNT(*) INTO admin_count
    FROM dashboard.users
    WHERE role = 'dashboard_admin'
    AND (id != first_admin_id OR first_admin_id IS NULL)
    AND deleted_at IS NULL;

    IF admin_count > 0 THEN
        -- Update their dashboard role to tenant_admin
        UPDATE dashboard.users
        SET role = 'tenant_admin',
            updated_at = NOW()
        WHERE role = 'dashboard_admin'
        AND (id != first_admin_id OR first_admin_id IS NULL)
        AND deleted_at IS NULL;

    END IF;

    -- Step 5: Ensure default tenant membership for instance_admin
    IF first_admin_id IS NOT NULL THEN
        INSERT INTO platform.tenant_memberships (tenant_id, user_id, role, created_at)
        VALUES (default_tenant_id, first_admin_id, 'tenant_admin', NOW())
        ON CONFLICT (tenant_id, user_id) DO NOTHING;
    END IF;

    RAISE NOTICE 'Admin role migration complete';
END $$;
