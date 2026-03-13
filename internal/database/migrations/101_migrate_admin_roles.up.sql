--
-- MULTI-TENANCY: MIGRATE ADMIN ROLES
-- Creates default tenant and migrates existing admin users to new role model
--
-- Migration Strategy:
--   - First setup admin → instance_admin (global admin)
--   - Other dashboard_admin users → tenant_admin of default tenant
--   - Creates tenant memberships for migrated admins
--

DO $$
DECLARE
    default_tenant_id UUID;
    first_admin_id UUID;
    admin_count INTEGER;
BEGIN
    -- Check if default tenant already exists
    SELECT id INTO default_tenant_id FROM tenants WHERE is_default = true LIMIT 1;
    
    IF default_tenant_id IS NULL THEN
        -- Step 1: Create default tenant
        INSERT INTO tenants (id, slug, name, is_default, metadata, created_at)
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
        
        -- Create membership for instance admin in default tenant (as tenant_admin for data access)
        INSERT INTO tenant_memberships (tenant_id, user_id, role, created_at)
        VALUES (default_tenant_id, first_admin_id, 'tenant_admin', NOW())
        ON CONFLICT (tenant_id, user_id) DO UPDATE SET role = 'tenant_admin', updated_at = NOW();
    END IF;
    
    -- Step 4: Migrate other dashboard_admin to tenant_admin
    -- Count how many will be migrated
    SELECT COUNT(*) INTO admin_count
    FROM dashboard.users
    WHERE role = 'dashboard_admin'
    AND (id != first_admin_id OR first_admin_id IS NULL)
    AND deleted_at IS NULL;
    
    IF admin_count > 0 THEN
        -- Create tenant memberships for other dashboard_admins
        INSERT INTO tenant_memberships (tenant_id, user_id, role, created_at)
        SELECT 
            default_tenant_id,
            du.id,
            'tenant_admin',
            NOW()
        FROM dashboard.users du
        WHERE du.role = 'dashboard_admin'
        AND (du.id != first_admin_id OR first_admin_id IS NULL)
        AND du.deleted_at IS NULL
        ON CONFLICT (tenant_id, user_id) DO UPDATE SET role = 'tenant_admin', updated_at = NOW();
        
        -- Update their dashboard role to tenant_admin
        UPDATE dashboard.users 
        SET role = 'tenant_admin',
            updated_at = NOW()
        WHERE role = 'dashboard_admin'
        AND (id != first_admin_id OR first_admin_id IS NULL)
        AND deleted_at IS NULL;
        
        RAISE NOTICE 'Migrated % dashboard_admin users to tenant_admin', admin_count;
    ELSE
        RAISE NOTICE 'No additional dashboard_admin users to migrate';
    END IF;
    
    -- Step 5: Ensure dashboard_user role users also get membership (read-only access)
    INSERT INTO tenant_memberships (tenant_id, user_id, role, created_at)
    SELECT 
        default_tenant_id,
        du.id,
        'tenant_member',
        NOW()
    FROM dashboard.users du
    WHERE du.role = 'dashboard_user'
    AND du.deleted_at IS NULL
    AND du.is_active = true
    ON CONFLICT (tenant_id, user_id) DO NOTHING;
    
    RAISE NOTICE 'Admin role migration complete';
END $$;
