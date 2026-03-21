--
-- ROLLBACK: TENANT-SCOPED STORAGE
-- Reverts tenant isolation changes for storage tables
--

-- ============================================================================
-- PHASE 1: DROP RLS POLICIES
-- ============================================================================

-- Storage buckets
DROP POLICY IF EXISTS storage_buckets_admin ON storage.buckets;
DROP POLICY IF EXISTS storage_buckets_public_view ON storage.buckets;

-- Storage objects
DROP POLICY IF EXISTS storage_objects_admin ON storage.objects;
DROP POLICY IF EXISTS storage_objects_owner ON storage.objects;
DROP POLICY IF EXISTS storage_objects_public_read ON storage.objects;
DROP POLICY IF EXISTS storage_objects_shared_read ON storage.objects;
DROP POLICY IF EXISTS storage_objects_shared_write ON storage.objects;
DROP POLICY IF EXISTS storage_objects_shared_delete ON storage.objects;
DROP POLICY IF EXISTS storage_objects_insert ON storage.objects;

-- Storage object permissions
DROP POLICY IF EXISTS storage_object_permissions_admin ON storage.object_permissions;
DROP POLICY IF EXISTS storage_object_permissions_owner_manage ON storage.object_permissions;
DROP POLICY IF EXISTS storage_object_permissions_view_shared ON storage.object_permissions;

-- Chunked upload sessions
DROP POLICY IF EXISTS storage_chunked_sessions_admin ON storage.chunked_upload_sessions;
DROP POLICY IF EXISTS storage_chunked_sessions_owner ON storage.chunked_upload_sessions;
DROP POLICY IF EXISTS storage_chunked_sessions_insert ON storage.chunked_upload_sessions;

-- ============================================================================
-- PHASE 2: DROP HELPER FUNCTION
-- ============================================================================

DROP FUNCTION IF EXISTS storage.has_tenant_access(UUID);

-- ============================================================================
-- PHASE 3: DROP TRIGGERS AND FUNCTION
-- ============================================================================

DROP TRIGGER IF EXISTS storage_buckets_set_tenant_id ON storage.buckets;
DROP TRIGGER IF EXISTS storage_objects_set_tenant_id ON storage.objects;
DROP TRIGGER IF EXISTS storage_object_permissions_set_tenant_id ON storage.object_permissions;
DROP TRIGGER IF EXISTS storage_chunked_sessions_set_tenant_id ON storage.chunked_upload_sessions;

DROP FUNCTION IF EXISTS storage.set_tenant_id();

-- ============================================================================
-- PHASE 4: RESTORE ORIGINAL HELPER FUNCTIONS
-- ============================================================================

-- Restore bucket_exists to non-tenant-aware version
DROP FUNCTION IF EXISTS storage.bucket_exists(TEXT, UUID);
DROP FUNCTION IF EXISTS storage.bucket_exists(TEXT);

CREATE OR REPLACE FUNCTION storage.bucket_exists(bucket_name TEXT)
RETURNS BOOLEAN
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = public, storage
AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM storage.buckets
        WHERE name = bucket_name
    );
END;
$$;

COMMENT ON FUNCTION storage.bucket_exists(TEXT) IS 'SECURITY DEFINER function to check if a bucket exists, bypassing RLS. Used by storage handler to validate bucket existence before upload.';

-- Restore get_bucket_settings to non-tenant-aware version
DROP FUNCTION IF EXISTS storage.get_bucket_settings(TEXT, UUID);
DROP FUNCTION IF EXISTS storage.get_bucket_settings(TEXT);

CREATE OR REPLACE FUNCTION storage.get_bucket_settings(bucket_name TEXT)
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
    RETURN QUERY
    SELECT b.max_file_size, b.allowed_mime_types
    FROM storage.buckets b
    WHERE b.name = bucket_name;
END;
$$;

COMMENT ON FUNCTION storage.get_bucket_settings(TEXT) IS 'SECURITY DEFINER function to get bucket settings, bypassing RLS. Used by storage handler to validate upload constraints.';

-- Restore is_bucket_public to non-tenant-aware version
DROP FUNCTION IF EXISTS storage.is_bucket_public(TEXT, UUID);
DROP FUNCTION IF EXISTS storage.is_bucket_public(TEXT);

CREATE OR REPLACE FUNCTION storage.is_bucket_public(bucket_name TEXT)
RETURNS BOOLEAN
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = public, storage
AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM storage.buckets
        WHERE id = bucket_name AND public = true
    );
END;
$$;

COMMENT ON FUNCTION storage.is_bucket_public(TEXT) IS 'Check if a bucket is public, bypassing RLS to prevent infinite recursion';

-- ============================================================================
-- PHASE 5: DROP UNIQUE INDEXES AND RESTORE CONSTRAINT
-- ============================================================================

DROP INDEX IF EXISTS storage_buckets_name_tenant_not_null;
DROP INDEX IF EXISTS storage_buckets_name_tenant_null;

-- Restore global unique constraint on bucket name
ALTER TABLE storage.buckets ADD CONSTRAINT storage_buckets_name_key UNIQUE (name);

-- ============================================================================
-- PHASE 6: DROP INDEXES
-- ============================================================================

DROP INDEX IF EXISTS idx_storage_buckets_tenant_id;
DROP INDEX IF EXISTS idx_storage_objects_tenant_id;
DROP INDEX IF EXISTS idx_storage_object_permissions_tenant_id;
DROP INDEX IF EXISTS idx_storage_chunked_sessions_tenant_id;

-- ============================================================================
-- PHASE 7: DROP TENANT_ID COLUMNS
-- ============================================================================

ALTER TABLE storage.chunked_upload_sessions DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE storage.object_permissions DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE storage.objects DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE storage.buckets DROP COLUMN IF EXISTS tenant_id;

-- ============================================================================
-- PHASE 8: RESTORE ORIGINAL RLS POLICIES
-- ============================================================================

-- Storage buckets - original policies
CREATE POLICY storage_buckets_admin ON storage.buckets
    FOR ALL
    USING (auth.current_user_role() IN ('dashboard_admin', 'service_role'))
    WITH CHECK (auth.current_user_role() IN ('dashboard_admin', 'service_role'));

CREATE POLICY storage_buckets_public_view ON storage.buckets
    FOR SELECT
    USING (public = true);

-- Storage objects - original policies
CREATE POLICY storage_objects_admin ON storage.objects
    FOR ALL
    USING (auth.current_user_role() IN ('dashboard_admin', 'service_role'))
    WITH CHECK (auth.current_user_role() IN ('dashboard_admin', 'service_role'));

CREATE POLICY storage_objects_owner ON storage.objects
    FOR ALL
    USING (auth.current_user_id() = owner_id)
    WITH CHECK (auth.current_user_id() = owner_id);

CREATE POLICY storage_objects_public_read ON storage.objects
    FOR SELECT
    USING (storage.is_bucket_public(bucket_id));

CREATE POLICY storage_objects_shared_read ON storage.objects
    FOR SELECT
    USING (
        auth.current_user_id() IS NOT NULL
        AND storage.has_object_permission(id, auth.current_user_id(), 'read')
    );

CREATE POLICY storage_objects_shared_write ON storage.objects
    FOR UPDATE
    USING (
        auth.current_user_id() IS NOT NULL
        AND storage.has_object_permission(id, auth.current_user_id(), 'write')
    );

CREATE POLICY storage_objects_shared_delete ON storage.objects
    FOR DELETE
    USING (
        auth.current_user_id() IS NOT NULL
        AND storage.has_object_permission(id, auth.current_user_id(), 'write')
    );

CREATE POLICY storage_objects_insert ON storage.objects
    FOR INSERT
    WITH CHECK (
        auth.current_user_role() IN ('dashboard_admin', 'service_role')
        OR (auth.current_user_id() IS NOT NULL AND auth.current_user_id() = owner_id)
    );

-- Storage object permissions - original policies
CREATE POLICY storage_object_permissions_admin ON storage.object_permissions
    FOR ALL
    USING (auth.current_user_role() IN ('dashboard_admin', 'service_role'))
    WITH CHECK (auth.current_user_role() IN ('dashboard_admin', 'service_role'));

CREATE POLICY storage_object_permissions_owner_manage ON storage.object_permissions
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM storage.objects
            WHERE objects.id = object_permissions.object_id
            AND objects.owner_id = auth.current_user_id()
        )
    )
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM storage.objects
            WHERE objects.id = object_permissions.object_id
            AND objects.owner_id = auth.current_user_id()
        )
    );

CREATE POLICY storage_object_permissions_view_shared ON storage.object_permissions
    FOR SELECT
    USING (user_id = auth.current_user_id());

-- Chunked upload sessions - original policies
CREATE POLICY storage_chunked_sessions_admin ON storage.chunked_upload_sessions
    FOR ALL
    USING (auth.current_user_role() IN ('dashboard_admin', 'service_role'))
    WITH CHECK (auth.current_user_role() IN ('dashboard_admin', 'service_role'));

CREATE POLICY storage_chunked_sessions_owner ON storage.chunked_upload_sessions
    FOR ALL
    USING (auth.current_user_id() = owner_id)
    WITH CHECK (auth.current_user_id() = owner_id);

CREATE POLICY storage_chunked_sessions_insert ON storage.chunked_upload_sessions
    FOR INSERT
    WITH CHECK (
        auth.current_user_role() IN ('dashboard_admin', 'service_role')
        OR (auth.current_user_id() IS NOT NULL AND auth.current_user_id() = owner_id)
    );

-- ============================================================================
-- PHASE 9: GRANT PERMISSIONS
-- ============================================================================

GRANT EXECUTE ON FUNCTION storage.bucket_exists(TEXT) TO anon, authenticated, service_role;
GRANT EXECUTE ON FUNCTION storage.get_bucket_settings(TEXT) TO anon, authenticated, service_role;
GRANT EXECUTE ON FUNCTION storage.is_bucket_public(TEXT) TO anon, authenticated, service_role;
