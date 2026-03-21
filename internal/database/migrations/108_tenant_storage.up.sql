--
-- TENANT-SCOPED STORAGE
-- Adds tenant_id columns to storage tables with RLS policies for tenant isolation
-- Key design: tenant_id = NULL means "default tenant" (backward compatibility)
-- Bucket names are unique per tenant (not globally)
--

-- ============================================================================
-- PHASE 1: ADD TENANT_ID COLUMNS
-- ============================================================================

-- Add tenant_id to storage.buckets
ALTER TABLE storage.buckets ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;

-- Add tenant_id to storage.objects
ALTER TABLE storage.objects ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;

-- Add tenant_id to storage.object_permissions
ALTER TABLE storage.object_permissions ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;

-- Add tenant_id to storage.chunked_upload_sessions
ALTER TABLE storage.chunked_upload_sessions ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;

-- ============================================================================
-- PHASE 2: CREATE INDEXES
-- ============================================================================

-- Indexes for tenant-based queries
CREATE INDEX IF NOT EXISTS idx_storage_buckets_tenant_id ON storage.buckets(tenant_id);
CREATE INDEX IF NOT EXISTS idx_storage_objects_tenant_id ON storage.objects(tenant_id);
CREATE INDEX IF NOT EXISTS idx_storage_object_permissions_tenant_id ON storage.object_permissions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_storage_chunked_sessions_tenant_id ON storage.chunked_upload_sessions(tenant_id);

-- ============================================================================
-- PHASE 3: UPDATE BUCKET NAME UNIQUENESS
-- ============================================================================
-- Bucket names should be unique per tenant, not globally
-- We use partial indexes to handle NULL tenant_id correctly

-- Drop the existing global unique constraint on bucket name
ALTER TABLE storage.buckets DROP CONSTRAINT IF EXISTS storage_buckets_name_key;

-- Create partial unique indexes for tenant-scoped bucket names
-- NULL tenant_id = default tenant (backward compatibility)
CREATE UNIQUE INDEX IF NOT EXISTS storage_buckets_name_tenant_not_null
    ON storage.buckets(name, tenant_id)
    WHERE tenant_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS storage_buckets_name_tenant_null
    ON storage.buckets(name)
    WHERE tenant_id IS NULL;

COMMENT ON INDEX storage_buckets_name_tenant_not_null IS 'Ensures bucket names are unique within each tenant';
COMMENT ON INDEX storage_buckets_name_tenant_null IS 'Ensures bucket names are unique in the default tenant (NULL tenant_id)';

-- ============================================================================
-- PHASE 4: AUTO-POPULATE TRIGGER
-- ============================================================================
-- Automatically set tenant_id from session context on INSERT

CREATE OR REPLACE FUNCTION storage.set_tenant_id()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    -- Only set tenant_id if it's NULL and we have tenant context
    IF NEW.tenant_id IS NULL THEN
        NEW.tenant_id := NULLIF(current_setting('app.current_tenant_id', true), '')::uuid;
    END IF;
    RETURN NEW;
END;
$$;

COMMENT ON FUNCTION storage.set_tenant_id() IS 'Auto-populates tenant_id from session context on INSERT operations';

-- Create triggers for auto-population
DROP TRIGGER IF EXISTS storage_buckets_set_tenant_id ON storage.buckets;
CREATE TRIGGER storage_buckets_set_tenant_id
    BEFORE INSERT ON storage.buckets
    FOR EACH ROW
    EXECUTE FUNCTION storage.set_tenant_id();

DROP TRIGGER IF EXISTS storage_objects_set_tenant_id ON storage.objects;
CREATE TRIGGER storage_objects_set_tenant_id
    BEFORE INSERT ON storage.objects
    FOR EACH ROW
    EXECUTE FUNCTION storage.set_tenant_id();

DROP TRIGGER IF EXISTS storage_object_permissions_set_tenant_id ON storage.object_permissions;
CREATE TRIGGER storage_object_permissions_set_tenant_id
    BEFORE INSERT ON storage.object_permissions
    FOR EACH ROW
    EXECUTE FUNCTION storage.set_tenant_id();

DROP TRIGGER IF EXISTS storage_chunked_sessions_set_tenant_id ON storage.chunked_upload_sessions;
CREATE TRIGGER storage_chunked_sessions_set_tenant_id
    BEFORE INSERT ON storage.chunked_upload_sessions
    FOR EACH ROW
    EXECUTE FUNCTION storage.set_tenant_id();

-- ============================================================================
-- PHASE 5: UPDATE HELPER FUNCTIONS TO BE TENANT-AWARE
-- ============================================================================

-- Update bucket_exists to be tenant-aware
DROP FUNCTION IF EXISTS storage.bucket_exists(TEXT);
CREATE OR REPLACE FUNCTION storage.bucket_exists(
    p_bucket_name TEXT,
    p_tenant_id UUID DEFAULT NULL
)
RETURNS BOOLEAN
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = public, storage
AS $$
BEGIN
    -- If tenant_id is provided, check for bucket in that tenant
    IF p_tenant_id IS NOT NULL THEN
        RETURN EXISTS (
            SELECT 1 FROM storage.buckets
            WHERE name = p_bucket_name AND tenant_id = p_tenant_id
        );
    END IF;

    -- If no tenant_id, check for bucket in default tenant (NULL)
    -- This maintains backward compatibility
    RETURN EXISTS (
        SELECT 1 FROM storage.buckets
        WHERE name = p_bucket_name AND tenant_id IS NULL
    );
END;
$$;

COMMENT ON FUNCTION storage.bucket_exists(TEXT, UUID) IS
    'Check if a bucket exists in a specific tenant context. p_tenant_id = NULL checks default tenant. SECURITY DEFINER bypasses RLS.';

-- Overload for backward compatibility (uses session context)
CREATE OR REPLACE FUNCTION storage.bucket_exists(p_bucket_name TEXT)
RETURNS BOOLEAN
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = public, storage
AS $$
DECLARE
    v_tenant_id UUID;
BEGIN
    -- Get tenant from session context
    v_tenant_id := NULLIF(current_setting('app.current_tenant_id', true), '')::uuid;

    -- Use the tenant-aware version
    RETURN storage.bucket_exists(p_bucket_name, v_tenant_id);
END;
$$;

COMMENT ON FUNCTION storage.bucket_exists(TEXT) IS
    'Check if a bucket exists using session tenant context. SECURITY DEFINER bypasses RLS.';

-- Update get_bucket_settings to be tenant-aware
DROP FUNCTION IF EXISTS storage.get_bucket_settings(TEXT);
CREATE OR REPLACE FUNCTION storage.get_bucket_settings(
    p_bucket_name TEXT,
    p_tenant_id UUID DEFAULT NULL
)
RETURNS TABLE (
    max_file_size BIGINT,
    allowed_mime_types TEXT[]
)
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = public, storage
AS $$
BEGIN
    -- If tenant_id is provided, get settings for bucket in that tenant
    IF p_tenant_id IS NOT NULL THEN
        RETURN QUERY
        SELECT b.max_file_size, b.allowed_mime_types
        FROM storage.buckets b
        WHERE b.name = p_bucket_name AND b.tenant_id = p_tenant_id;
        RETURN;
    END IF;

    -- If no tenant_id, get settings for bucket in default tenant (NULL)
    RETURN QUERY
    SELECT b.max_file_size, b.allowed_mime_types
    FROM storage.buckets b
    WHERE b.name = p_bucket_name AND b.tenant_id IS NULL;
END;
$$;

COMMENT ON FUNCTION storage.get_bucket_settings(TEXT, UUID) IS
    'Get bucket settings for a specific tenant context. p_tenant_id = NULL checks default tenant. SECURITY DEFINER bypasses RLS.';

-- Overload for backward compatibility (uses session context)
CREATE OR REPLACE FUNCTION storage.get_bucket_settings(p_bucket_name TEXT)
RETURNS TABLE (
    max_file_size BIGINT,
    allowed_mime_types TEXT[]
)
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = public, storage
AS $$
DECLARE
    v_tenant_id UUID;
BEGIN
    -- Get tenant from session context
    v_tenant_id := NULLIF(current_setting('app.current_tenant_id', true), '')::uuid;

    -- Use the tenant-aware version
    RETURN QUERY SELECT * FROM storage.get_bucket_settings(p_bucket_name, v_tenant_id);
END;
$$;

COMMENT ON FUNCTION storage.get_bucket_settings(TEXT) IS
    'Get bucket settings using session tenant context. SECURITY DEFINER bypasses RLS.';

-- Update is_bucket_public to be tenant-aware
DROP FUNCTION IF EXISTS storage.is_bucket_public(TEXT);
CREATE OR REPLACE FUNCTION storage.is_bucket_public(
    p_bucket_name TEXT,
    p_tenant_id UUID DEFAULT NULL
)
RETURNS BOOLEAN
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = public, storage
AS $$
BEGIN
    -- If tenant_id is provided, check bucket in that tenant
    IF p_tenant_id IS NOT NULL THEN
        RETURN EXISTS (
            SELECT 1 FROM storage.buckets
            WHERE name = p_bucket_name
            AND tenant_id = p_tenant_id
            AND public = true
        );
    END IF;

    -- If no tenant_id, check bucket in default tenant (NULL)
    RETURN EXISTS (
        SELECT 1 FROM storage.buckets
        WHERE name = p_bucket_name
        AND tenant_id IS NULL
        AND public = true
    );
END;
$$;

COMMENT ON FUNCTION storage.is_bucket_public(TEXT, UUID) IS
    'Check if a bucket is public in a specific tenant context. SECURITY DEFINER bypasses RLS.';

-- Overload for backward compatibility (uses session context)
CREATE OR REPLACE FUNCTION storage.is_bucket_public(p_bucket_name TEXT)
RETURNS BOOLEAN
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = public, storage
AS $$
DECLARE
    v_tenant_id UUID;
BEGIN
    -- Get tenant from session context
    v_tenant_id := NULLIF(current_setting('app.current_tenant_id', true), '')::uuid;

    -- Use the tenant-aware version
    RETURN storage.is_bucket_public(p_bucket_name, v_tenant_id);
END;
$$;

COMMENT ON FUNCTION storage.is_bucket_public(TEXT) IS
    'Check if a bucket is public using session tenant context. SECURITY DEFINER bypasses RLS.';

-- ============================================================================
-- PHASE 6: UPDATE RLS POLICIES FOR TENANT ISOLATION
-- ============================================================================
-- All policies now check tenant_id to ensure isolation between tenants
-- NULL tenant_id = default tenant (backward compatibility)

-- ============================================================================
-- STORAGE BUCKETS RLS
-- ============================================================================

-- Drop existing policies
DROP POLICY IF EXISTS storage_buckets_admin ON storage.buckets;
DROP POLICY IF EXISTS storage_buckets_public_view ON storage.buckets;

-- Helper function to check tenant access
CREATE OR REPLACE FUNCTION storage.has_tenant_access(p_tenant_id UUID)
RETURNS BOOLEAN
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
AS $$
DECLARE
    v_current_tenant TEXT;
BEGIN
    -- Get current tenant from session
    v_current_tenant := current_setting('app.current_tenant_id', true);

    -- If no tenant context, only allow access to NULL tenant (default)
    IF v_current_tenant IS NULL OR v_current_tenant = '' THEN
        RETURN p_tenant_id IS NULL;
    END IF;

    -- Allow access if tenant matches
    RETURN p_tenant_id::text = v_current_tenant;
END;
$$;

COMMENT ON FUNCTION storage.has_tenant_access(UUID) IS
    'Check if current session has access to the specified tenant. SECURITY DEFINER for RLS policy use.';

-- Admins and service roles can do everything with buckets (within their tenant context)
CREATE POLICY storage_buckets_admin ON storage.buckets
    FOR ALL
    USING (
        storage.has_tenant_access(tenant_id)
        AND auth.current_user_role() IN ('dashboard_admin', 'service_role', 'admin', 'instance_admin')
    )
    WITH CHECK (
        storage.has_tenant_access(tenant_id)
        AND auth.current_user_role() IN ('dashboard_admin', 'service_role', 'admin', 'instance_admin')
    );

COMMENT ON POLICY storage_buckets_admin ON storage.buckets IS 'Dashboard admins and service role have full access to buckets within their tenant context';

-- Anyone can view public buckets within their tenant
CREATE POLICY storage_buckets_public_view ON storage.buckets
    FOR SELECT
    USING (
        storage.has_tenant_access(tenant_id)
        AND public = true
    );

COMMENT ON POLICY storage_buckets_public_view ON storage.buckets IS 'Public buckets are visible to users within the same tenant';

-- ============================================================================
-- STORAGE OBJECTS RLS
-- ============================================================================

-- Drop existing policies
DROP POLICY IF EXISTS storage_objects_admin ON storage.objects;
DROP POLICY IF EXISTS storage_objects_owner ON storage.objects;
DROP POLICY IF EXISTS storage_objects_public_read ON storage.objects;
DROP POLICY IF EXISTS storage_objects_shared_read ON storage.objects;
DROP POLICY IF EXISTS storage_objects_shared_write ON storage.objects;
DROP POLICY IF EXISTS storage_objects_shared_delete ON storage.objects;
DROP POLICY IF EXISTS storage_objects_insert ON storage.objects;

-- Admins and service roles can do everything with objects (within tenant)
CREATE POLICY storage_objects_admin ON storage.objects
    FOR ALL
    USING (
        storage.has_tenant_access(tenant_id)
        AND auth.current_user_role() IN ('dashboard_admin', 'service_role', 'admin', 'instance_admin')
    )
    WITH CHECK (
        storage.has_tenant_access(tenant_id)
        AND auth.current_user_role() IN ('dashboard_admin', 'service_role', 'admin', 'instance_admin')
    );

COMMENT ON POLICY storage_objects_admin ON storage.objects IS 'Dashboard admins and service role have full access to objects within their tenant context';

-- Owners can do everything with their files (within tenant)
CREATE POLICY storage_objects_owner ON storage.objects
    FOR ALL
    USING (
        storage.has_tenant_access(tenant_id)
        AND auth.current_user_id() = owner_id
    )
    WITH CHECK (
        storage.has_tenant_access(tenant_id)
        AND auth.current_user_id() = owner_id
    );

COMMENT ON POLICY storage_objects_owner ON storage.objects IS 'Users can fully manage their own files within their tenant context';

-- Anyone can read files in public buckets (within tenant)
CREATE POLICY storage_objects_public_read ON storage.objects
    FOR SELECT
    USING (
        storage.has_tenant_access(tenant_id)
        AND storage.is_bucket_public(bucket_id, tenant_id)
    );

COMMENT ON POLICY storage_objects_public_read ON storage.objects IS 'Files in public buckets are readable by users within the same tenant';

-- Users can read files shared with them (within tenant)
CREATE POLICY storage_objects_shared_read ON storage.objects
    FOR SELECT
    USING (
        storage.has_tenant_access(tenant_id)
        AND auth.current_user_id() IS NOT NULL
        AND storage.has_object_permission(id, auth.current_user_id(), 'read')
    );

COMMENT ON POLICY storage_objects_shared_read ON storage.objects IS 'Users can read files shared with them within their tenant context';

-- Users can update files shared with write permission (within tenant)
CREATE POLICY storage_objects_shared_write ON storage.objects
    FOR UPDATE
    USING (
        storage.has_tenant_access(tenant_id)
        AND auth.current_user_id() IS NOT NULL
        AND storage.has_object_permission(id, auth.current_user_id(), 'write')
    );

COMMENT ON POLICY storage_objects_shared_write ON storage.objects IS 'Users can update files shared with write permission within their tenant context';

-- Users can delete files shared with write permission (within tenant)
CREATE POLICY storage_objects_shared_delete ON storage.objects
    FOR DELETE
    USING (
        storage.has_tenant_access(tenant_id)
        AND auth.current_user_id() IS NOT NULL
        AND storage.has_object_permission(id, auth.current_user_id(), 'write')
    );

COMMENT ON POLICY storage_objects_shared_delete ON storage.objects IS 'Users can delete files shared with write permission within their tenant context';

-- Authenticated users can insert objects (within tenant, owner_id must match)
CREATE POLICY storage_objects_insert ON storage.objects
    FOR INSERT
    WITH CHECK (
        storage.has_tenant_access(tenant_id)
        AND (
            auth.current_user_role() IN ('dashboard_admin', 'service_role', 'admin', 'instance_admin')
            OR (auth.current_user_id() IS NOT NULL AND auth.current_user_id() = owner_id)
        )
    );

COMMENT ON POLICY storage_objects_insert ON storage.objects IS 'Users can upload files within their tenant context (owner_id must match their user ID)';

-- ============================================================================
-- STORAGE OBJECT PERMISSIONS RLS
-- ============================================================================

-- Drop existing policies
DROP POLICY IF EXISTS storage_object_permissions_admin ON storage.object_permissions;
DROP POLICY IF EXISTS storage_object_permissions_owner_manage ON storage.object_permissions;
DROP POLICY IF EXISTS storage_object_permissions_view_shared ON storage.object_permissions;

-- Admins can manage all permissions (within tenant)
CREATE POLICY storage_object_permissions_admin ON storage.object_permissions
    FOR ALL
    USING (
        storage.has_tenant_access(tenant_id)
        AND auth.current_user_role() IN ('dashboard_admin', 'service_role', 'admin', 'instance_admin')
    )
    WITH CHECK (
        storage.has_tenant_access(tenant_id)
        AND auth.current_user_role() IN ('dashboard_admin', 'service_role', 'admin', 'instance_admin')
    );

COMMENT ON POLICY storage_object_permissions_admin ON storage.object_permissions IS 'Dashboard admins and service role can manage all file sharing permissions within their tenant context';

-- Owners can share their own files (within tenant)
-- Note: We need to check the object's tenant_id matches the permission's tenant_id
CREATE POLICY storage_object_permissions_owner_manage ON storage.object_permissions
    FOR ALL
    USING (
        storage.has_tenant_access(tenant_id)
        AND EXISTS (
            SELECT 1 FROM storage.objects
            WHERE objects.id = object_permissions.object_id
            AND objects.owner_id = auth.current_user_id()
            AND storage.has_tenant_access(objects.tenant_id)
        )
    )
    WITH CHECK (
        storage.has_tenant_access(tenant_id)
        AND EXISTS (
            SELECT 1 FROM storage.objects
            WHERE objects.id = object_permissions.object_id
            AND objects.owner_id = auth.current_user_id()
            AND storage.has_tenant_access(objects.tenant_id)
        )
    );

COMMENT ON POLICY storage_object_permissions_owner_manage ON storage.object_permissions IS 'File owners can manage sharing permissions for their files within their tenant context';

-- Users can view permissions for files shared with them (within tenant)
CREATE POLICY storage_object_permissions_view_shared ON storage.object_permissions
    FOR SELECT
    USING (
        storage.has_tenant_access(tenant_id)
        AND user_id = auth.current_user_id()
    );

COMMENT ON POLICY storage_object_permissions_view_shared ON storage.object_permissions IS 'Users can view sharing permissions for files shared with them within their tenant context';

-- ============================================================================
-- CHUNKED UPLOAD SESSIONS RLS
-- ============================================================================

-- Drop existing policies
DROP POLICY IF EXISTS storage_chunked_sessions_admin ON storage.chunked_upload_sessions;
DROP POLICY IF EXISTS storage_chunked_sessions_owner ON storage.chunked_upload_sessions;
DROP POLICY IF EXISTS storage_chunked_sessions_insert ON storage.chunked_upload_sessions;

-- Admins and service roles can manage all upload sessions (within tenant)
CREATE POLICY storage_chunked_sessions_admin ON storage.chunked_upload_sessions
    FOR ALL
    USING (
        storage.has_tenant_access(tenant_id)
        AND auth.current_user_role() IN ('dashboard_admin', 'service_role', 'admin', 'instance_admin')
    )
    WITH CHECK (
        storage.has_tenant_access(tenant_id)
        AND auth.current_user_role() IN ('dashboard_admin', 'service_role', 'admin', 'instance_admin')
    );

COMMENT ON POLICY storage_chunked_sessions_admin ON storage.chunked_upload_sessions IS 'Dashboard admins and service role have full access to upload sessions within their tenant context';

-- Owners can manage their own upload sessions (within tenant)
CREATE POLICY storage_chunked_sessions_owner ON storage.chunked_upload_sessions
    FOR ALL
    USING (
        storage.has_tenant_access(tenant_id)
        AND auth.current_user_id() = owner_id
    )
    WITH CHECK (
        storage.has_tenant_access(tenant_id)
        AND auth.current_user_id() = owner_id
    );

COMMENT ON POLICY storage_chunked_sessions_owner ON storage.chunked_upload_sessions IS 'Users can fully manage their own upload sessions within their tenant context';

-- Users can only insert sessions with their own owner_id (within tenant)
CREATE POLICY storage_chunked_sessions_insert ON storage.chunked_upload_sessions
    FOR INSERT
    WITH CHECK (
        storage.has_tenant_access(tenant_id)
        AND (
            auth.current_user_role() IN ('dashboard_admin', 'service_role', 'admin', 'instance_admin')
            OR (auth.current_user_id() IS NOT NULL AND auth.current_user_id() = owner_id)
        )
    );

COMMENT ON POLICY storage_chunked_sessions_insert ON storage.chunked_upload_sessions IS 'Users can only create upload sessions with their own owner_id within their tenant context';

-- ============================================================================
-- GRANT PERMISSIONS
-- ============================================================================

-- Grant execute permissions on new overloaded functions
GRANT EXECUTE ON FUNCTION storage.bucket_exists(TEXT, UUID) TO anon, authenticated, service_role;
GRANT EXECUTE ON FUNCTION storage.get_bucket_settings(TEXT, UUID) TO anon, authenticated, service_role;
GRANT EXECUTE ON FUNCTION storage.is_bucket_public(TEXT, UUID) TO anon, authenticated, service_role;
GRANT EXECUTE ON FUNCTION storage.has_tenant_access(UUID) TO anon, authenticated, service_role;

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON COLUMN storage.buckets.tenant_id IS 'Tenant ID for multi-tenancy. NULL = default tenant (backward compatibility)';
COMMENT ON COLUMN storage.objects.tenant_id IS 'Tenant ID for multi-tenancy. NULL = default tenant (backward compatibility)';
COMMENT ON COLUMN storage.object_permissions.tenant_id IS 'Tenant ID for multi-tenancy. NULL = default tenant (backward compatibility)';
COMMENT ON COLUMN storage.chunked_upload_sessions.tenant_id IS 'Tenant ID for multi-tenancy. NULL = default tenant (backward compatibility)';
