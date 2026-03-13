--
-- MULTI-TENANCY: ROLLBACK ADMIN ROLE MIGRATION
--

DO $$
DECLARE
    default_tenant_id UUID;
BEGIN
    -- Get default tenant ID
    SELECT id INTO default_tenant_id FROM tenants WHERE is_default = true LIMIT 1;
    
    IF default_tenant_id IS NOT NULL THEN
        -- Remove tenant memberships created during migration
        DELETE FROM tenant_memberships WHERE tenant_id = default_tenant_id;
    END IF;
    
    -- Revert instance_admin back to dashboard_admin
    UPDATE dashboard.users 
    SET role = 'dashboard_admin', 
        updated_at = NOW()
    WHERE role = 'instance_admin';
    
    -- Revert tenant_admin back to dashboard_admin (only those we migrated)
    -- We can't distinguish migrated from new, so revert all
    UPDATE dashboard.users 
    SET role = 'dashboard_admin',
        updated_at = NOW()
    WHERE role = 'tenant_admin';
    
    -- Delete default tenant
    DELETE FROM tenants WHERE is_default = true;
    
    RAISE NOTICE 'Admin role migration rollback complete';
END $$;
