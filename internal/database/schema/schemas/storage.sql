--
-- pgschema database dump
--

-- Dumped from database version PostgreSQL 18.3
-- Dumped by pgschema version 1.7.4

SET search_path TO storage, public;


--
-- Name: buckets; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS buckets (
    id text,
    name text NOT NULL,
    public boolean DEFAULT false,
    allowed_mime_types text[],
    max_file_size bigint,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    tenant_id uuid,
    CONSTRAINT buckets_pkey PRIMARY KEY (id),
    CONSTRAINT buckets_name_key UNIQUE (name)
);


COMMENT ON TABLE buckets IS 'Storage buckets configuration. Public buckets allow unauthenticated read access.';


COMMENT ON COLUMN storage.buckets.tenant_id IS 'Tenant ID for multi-tenancy. NULL = default tenant (backward compatibility)';

--
-- Name: idx_storage_buckets_tenant_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_storage_buckets_tenant_id ON buckets (tenant_id);

--
-- Name: storage_buckets_name_tenant_not_null; Type: INDEX; Schema: -; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS storage_buckets_name_tenant_not_null ON buckets (name, tenant_id) WHERE (tenant_id IS NOT NULL);


COMMENT ON INDEX storage_buckets_name_tenant_not_null IS 'Ensures bucket names are unique within each tenant';

--
-- Name: storage_buckets_name_tenant_null; Type: INDEX; Schema: -; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS storage_buckets_name_tenant_null ON buckets (name) WHERE (tenant_id IS NULL);


COMMENT ON INDEX storage_buckets_name_tenant_null IS 'Ensures bucket names are unique in the default tenant (NULL tenant_id)';

--
-- Name: buckets; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE buckets ENABLE ROW LEVEL SECURITY;

--
-- Name: buckets; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE buckets FORCE ROW LEVEL SECURITY;

--
-- Name: chunked_upload_sessions; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS chunked_upload_sessions (
    id uuid DEFAULT gen_random_uuid(),
    upload_id text NOT NULL,
    bucket_id text NOT NULL,
    path text NOT NULL,
    total_size bigint NOT NULL,
    chunk_size integer NOT NULL,
    total_chunks integer NOT NULL,
    completed_chunks integer[] DEFAULT '{}'::integer[],
    content_type text,
    metadata jsonb,
    cache_control text,
    owner_id uuid,
    s3_upload_id text,
    s3_part_etags jsonb,
    status text DEFAULT 'active',
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    expires_at timestamptz DEFAULT (now() + '24:00:00'::interval),
    tenant_id uuid,
    CONSTRAINT chunked_upload_sessions_pkey PRIMARY KEY (id),
    CONSTRAINT chunked_upload_sessions_upload_id_key UNIQUE (upload_id),
    CONSTRAINT chunked_upload_sessions_status_check CHECK (status IN ('active'::text, 'completing'::text, 'completed'::text, 'aborted'::text, 'expired'::text))
);


COMMENT ON COLUMN storage.chunked_upload_sessions.tenant_id IS 'Tenant ID for multi-tenancy. NULL = default tenant (backward compatibility)';

--
-- Name: idx_chunked_sessions_bucket; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_chunked_sessions_bucket ON chunked_upload_sessions (bucket_id);

--
-- Name: idx_chunked_sessions_expires; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_chunked_sessions_expires ON chunked_upload_sessions (expires_at) WHERE (status = 'active'::text);

--
-- Name: idx_chunked_sessions_owner; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_chunked_sessions_owner ON chunked_upload_sessions (owner_id) WHERE (owner_id IS NOT NULL);

--
-- Name: idx_chunked_sessions_status; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_chunked_sessions_status ON chunked_upload_sessions (status);

--
-- Name: idx_storage_chunked_sessions_tenant_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_storage_chunked_sessions_tenant_id ON chunked_upload_sessions (tenant_id);

--
-- Name: chunked_upload_sessions; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE chunked_upload_sessions ENABLE ROW LEVEL SECURITY;

--
-- Name: chunked_upload_sessions; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE chunked_upload_sessions FORCE ROW LEVEL SECURITY;

--
-- Name: objects; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS objects (
    id uuid DEFAULT gen_random_uuid(),
    bucket_id text,
    path text NOT NULL,
    mime_type text,
    size bigint,
    metadata jsonb,
    owner_id uuid,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    name text GENERATED ALWAYS AS (path) STORED,
    tenant_id uuid,
    CONSTRAINT objects_pkey PRIMARY KEY (id),
    CONSTRAINT objects_bucket_id_path_key UNIQUE (bucket_id, path),
    CONSTRAINT objects_bucket_id_fkey FOREIGN KEY (bucket_id) REFERENCES buckets (id) ON DELETE CASCADE
);


COMMENT ON TABLE objects IS 'Storage objects metadata. All file operations are tracked here for RLS enforcement.';


COMMENT ON COLUMN storage.objects.path IS 'Full path to the object within the bucket. Also accessible via the "name" column for Supabase compatibility.';


COMMENT ON COLUMN storage.objects.name IS 'Supabase-compatible alias for path column. Automatically synchronized with path.';


COMMENT ON COLUMN storage.objects.tenant_id IS 'Tenant ID for multi-tenancy. NULL = default tenant (backward compatibility)';

--
-- Name: idx_storage_objects_bucket_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_storage_objects_bucket_id ON objects (bucket_id);

--
-- Name: idx_storage_objects_owner_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_storage_objects_owner_id ON objects (owner_id);

--
-- Name: idx_storage_objects_tenant_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_storage_objects_tenant_id ON objects (tenant_id);

--
-- Name: objects; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE objects ENABLE ROW LEVEL SECURITY;

--
-- Name: objects; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE objects FORCE ROW LEVEL SECURITY;

--
-- Name: object_permissions; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS object_permissions (
    id uuid DEFAULT gen_random_uuid(),
    object_id uuid,
    user_id uuid,
    permission text NOT NULL,
    created_at timestamptz DEFAULT now(),
    tenant_id uuid,
    CONSTRAINT object_permissions_pkey PRIMARY KEY (id),
    CONSTRAINT object_permissions_object_id_user_id_key UNIQUE (object_id, user_id),
    CONSTRAINT object_permissions_object_id_fkey FOREIGN KEY (object_id) REFERENCES objects (id) ON DELETE CASCADE,
    CONSTRAINT object_permissions_permission_check CHECK (permission IN ('read'::text, 'write'::text))
);


COMMENT ON TABLE object_permissions IS 'Tracks file sharing permissions between users';


COMMENT ON COLUMN storage.object_permissions.permission IS 'Permission level: read (download only) or write (download, update, delete)';


COMMENT ON COLUMN storage.object_permissions.tenant_id IS 'Tenant ID for multi-tenancy. NULL = default tenant (backward compatibility)';

--
-- Name: idx_storage_object_permissions_object_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_storage_object_permissions_object_id ON object_permissions (object_id);

--
-- Name: idx_storage_object_permissions_tenant_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_storage_object_permissions_tenant_id ON object_permissions (tenant_id);

--
-- Name: idx_storage_object_permissions_user_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_storage_object_permissions_user_id ON object_permissions (user_id);

--
-- Name: object_permissions; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE object_permissions ENABLE ROW LEVEL SECURITY;

--
-- Name: object_permissions; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE object_permissions FORCE ROW LEVEL SECURITY;

--
-- Name: bucket_exists(text, uuid); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION bucket_exists(
    p_bucket_name text,
    p_tenant_id uuid DEFAULT NULL
)
RETURNS boolean
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = public, storage
AS $$
BEGIN
    -- If tenant_id is provided, check for bucket in that tenant
    IF p_tenant_id IS NOT NULL THEN
        RETURN EXISTS (
            SELECT 1 FROM buckets
            WHERE name = p_bucket_name AND tenant_id = p_tenant_id
        );
    END IF;

    -- If no tenant_id, check for bucket in default tenant (NULL)
    -- This maintains backward compatibility
    RETURN EXISTS (
        SELECT 1 FROM buckets
        WHERE name = p_bucket_name AND tenant_id IS NULL
    );
END;
$$;

--
-- Name: bucket_exists(text, uuid); Type: FUNCTION; Schema: -; Owner: -
--

COMMENT ON FUNCTION bucket_exists(text, uuid) IS 'Check if a bucket exists in a specific tenant context. p_tenant_id = NULL checks default tenant. SECURITY DEFINER bypasses RLS.';

--
-- Name: bucket_exists(text); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION bucket_exists(
    bucket_name text
)
RETURNS boolean
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
    RETURN bucket_exists(bucket_name, v_tenant_id);
END;
$$;

--
-- Name: bucket_exists(text); Type: FUNCTION; Schema: -; Owner: -
--

COMMENT ON FUNCTION bucket_exists(text) IS 'Check if a bucket exists using session tenant context. SECURITY DEFINER bypasses RLS.';

--
-- Name: foldername(text); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION foldername(
    name text
)
RETURNS text[]
LANGUAGE plpgsql
IMMUTABLE
AS $$
DECLARE
    path_parts TEXT[];
    folder_parts TEXT[];
BEGIN
    IF name IS NULL OR name = '' THEN
        RETURN ARRAY[]::TEXT[];
    END IF;

    -- Split the path by '/' to get folder structure
    path_parts := string_to_array(name, '/');

    -- Remove the last element (filename) to get just folders
    IF array_length(path_parts, 1) > 1 THEN
        folder_parts := path_parts[1:array_length(path_parts, 1) - 1];
    ELSE
        -- No folders, just a filename at root
        folder_parts := ARRAY[]::TEXT[];
    END IF;

    RETURN folder_parts;
END;
$$;

--
-- Name: foldername(text); Type: FUNCTION; Schema: -; Owner: -
--

COMMENT ON FUNCTION foldername(text) IS 'Supabase-compatible function that extracts folder path components from an object name/path. Returns array of folder names. Use [1] to get first folder, [2] for second, etc.';

--
-- Name: get_bucket_settings(text, uuid); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION get_bucket_settings(
    p_bucket_name text,
    p_tenant_id uuid DEFAULT NULL
)
RETURNS TABLE(max_file_size bigint, allowed_mime_types text[])
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
        FROM buckets b
        WHERE b.name = p_bucket_name AND b.tenant_id = p_tenant_id;
        RETURN;
    END IF;

    -- If no tenant_id, get settings for bucket in default tenant (NULL)
    RETURN QUERY
    SELECT b.max_file_size, b.allowed_mime_types
    FROM buckets b
    WHERE b.name = p_bucket_name AND b.tenant_id IS NULL;
END;
$$;

--
-- Name: get_bucket_settings(text, uuid); Type: FUNCTION; Schema: -; Owner: -
--

COMMENT ON FUNCTION get_bucket_settings(text, uuid) IS 'Get bucket settings for a specific tenant context. p_tenant_id = NULL checks default tenant. SECURITY DEFINER bypasses RLS.';

--
-- Name: get_bucket_settings(text); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION get_bucket_settings(
    bucket_name text
)
RETURNS TABLE(max_file_size bigint, allowed_mime_types text[])
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
    RETURN QUERY SELECT * FROM get_bucket_settings(bucket_name, v_tenant_id);
END;
$$;

--
-- Name: get_bucket_settings(text); Type: FUNCTION; Schema: -; Owner: -
--

COMMENT ON FUNCTION get_bucket_settings(text) IS 'Get bucket settings using session tenant context. SECURITY DEFINER bypasses RLS.';

--
-- Name: has_object_permission(uuid, uuid, text); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION has_object_permission(
    p_object_id uuid,
    p_user_id uuid,
    p_permission text
)
RETURNS boolean
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = public, storage
AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM object_permissions
        WHERE object_id = p_object_id
        AND user_id = p_user_id
        AND (permission = p_permission OR (p_permission = 'read' AND permission = 'write'))
    );
END;
$$;

--
-- Name: has_object_permission(uuid, uuid, text); Type: FUNCTION; Schema: -; Owner: -
--

COMMENT ON FUNCTION has_object_permission(uuid, uuid, text) IS 'Check if user has permission on object, bypassing RLS to prevent infinite recursion';

--
-- Name: has_tenant_access(uuid); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION has_tenant_access(
    p_tenant_id uuid
)
RETURNS boolean
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = public, storage
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

--
-- Name: has_tenant_access(uuid); Type: FUNCTION; Schema: -; Owner: -
--

COMMENT ON FUNCTION has_tenant_access(uuid) IS 'Check if current session has access to the specified tenant. SECURITY DEFINER for RLS policy use.';

--
-- Name: is_bucket_public(text, uuid); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION is_bucket_public(
    p_bucket_name text,
    p_tenant_id uuid DEFAULT NULL
)
RETURNS boolean
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = public, storage
AS $$
BEGIN
    -- If tenant_id is provided, check bucket in that tenant
    IF p_tenant_id IS NOT NULL THEN
        RETURN EXISTS (
            SELECT 1 FROM buckets
            WHERE name = p_bucket_name
            AND tenant_id = p_tenant_id
            AND public = true
        );
    END IF;

    -- If no tenant_id, check bucket in default tenant (NULL)
    RETURN EXISTS (
        SELECT 1 FROM buckets
        WHERE name = p_bucket_name
        AND tenant_id IS NULL
        AND public = true
    );
END;
$$;

--
-- Name: is_bucket_public(text, uuid); Type: FUNCTION; Schema: -; Owner: -
--

COMMENT ON FUNCTION is_bucket_public(text, uuid) IS 'Check if a bucket is public in a specific tenant context. SECURITY DEFINER bypasses RLS.';

--
-- Name: is_bucket_public(text); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION is_bucket_public(
    bucket_name text
)
RETURNS boolean
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
    RETURN is_bucket_public(bucket_name, v_tenant_id);
END;
$$;

--
-- Name: is_bucket_public(text); Type: FUNCTION; Schema: -; Owner: -
--

COMMENT ON FUNCTION is_bucket_public(text) IS 'Check if a bucket is public using session tenant context. SECURITY DEFINER bypasses RLS.';

--
-- Name: set_tenant_id(); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION set_tenant_id()
RETURNS trigger
LANGUAGE plpgsql
VOLATILE
AS $$
BEGIN
    -- Only set tenant_id if it's NULL and we have tenant context
    IF NEW.tenant_id IS NULL THEN
        NEW.tenant_id := NULLIF(current_setting('app.current_tenant_id', true), '')::uuid;
    END IF;
    RETURN NEW;
END;
$$;

--
-- Name: set_tenant_id(); Type: FUNCTION; Schema: -; Owner: -
--

COMMENT ON FUNCTION set_tenant_id() IS 'Auto-populates tenant_id from session context on INSERT operations';

--
-- Name: user_can_access_object(uuid, text); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION user_can_access_object(
    p_object_id uuid,
    p_required_permission text DEFAULT 'read'
)
RETURNS boolean
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = public, storage
AS $$
DECLARE
    v_owner_id UUID;
    v_bucket_public BOOLEAN;
    v_has_permission BOOLEAN;
    v_user_role TEXT;
BEGIN
    v_user_role := auth.current_user_role();

    -- Admins and service roles can access everything
    IF v_user_role IN ('dashboard_admin', 'service_role') THEN
        RETURN TRUE;
    END IF;

    -- Get object owner and bucket public status
    SELECT o.owner_id, b.public INTO v_owner_id, v_bucket_public
    FROM objects o
    JOIN buckets b ON b.id = o.bucket_id
    WHERE o.id = p_object_id;

    -- If object not found, deny access
    IF NOT FOUND THEN
        RETURN FALSE;
    END IF;

    -- Check if user is the owner
    IF v_owner_id = auth.current_user_id() THEN
        RETURN TRUE;
    END IF;

    -- Check if bucket is public (read-only for non-owners)
    IF v_bucket_public AND p_required_permission = 'read' THEN
        RETURN TRUE;
    END IF;

    -- Check object_permissions table for explicit shares
    IF auth.current_user_id() IS NOT NULL THEN
        SELECT EXISTS(
            SELECT 1 FROM object_permissions
            WHERE object_id = p_object_id
            AND user_id = auth.current_user_id()
            AND (permission = 'write' OR (permission = 'read' AND p_required_permission = 'read'))
        ) INTO v_has_permission;

        IF v_has_permission THEN
            RETURN TRUE;
        END IF;
    END IF;

    RETURN FALSE;
END;
$$;

--
-- Name: user_can_access_object(uuid, text); Type: FUNCTION; Schema: -; Owner: -
--

COMMENT ON FUNCTION user_can_access_object(uuid, text) IS 'Checks if the current user can access a storage object with the required permission (read or write). Returns TRUE if: user is admin/service role, user owns the object, object is in public bucket (read only), or user has been granted permission via object_permissions table.';

--
-- Cross-schema FKs moved to post-schema-fks.sql
-- buckets_tenant_id_fkey, chunked_upload_sessions_tenant_id_fkey, objects_owner_id_fkey,
-- objects_tenant_id_fkey, object_permissions_tenant_id_fkey, object_permissions_user_id_fkey
--

--
-- Name: storage_buckets_admin; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY storage_buckets_admin ON buckets TO PUBLIC USING (has_tenant_access(tenant_id) AND (auth.current_user_role() = ANY (ARRAY['dashboard_admin', 'service_role', 'admin', 'instance_admin']))) WITH CHECK (has_tenant_access(tenant_id) AND (auth.current_user_role() = ANY (ARRAY['dashboard_admin', 'service_role', 'admin', 'instance_admin'])));

--
-- Name: storage_buckets_public_view; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY storage_buckets_public_view ON buckets FOR SELECT TO PUBLIC USING (has_tenant_access(tenant_id) AND (public = true));

--
-- Name: storage_chunked_sessions_admin; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY storage_chunked_sessions_admin ON chunked_upload_sessions TO PUBLIC USING (has_tenant_access(tenant_id) AND (auth.current_user_role() = ANY (ARRAY['dashboard_admin', 'service_role', 'admin', 'instance_admin']))) WITH CHECK (has_tenant_access(tenant_id) AND (auth.current_user_role() = ANY (ARRAY['dashboard_admin', 'service_role', 'admin', 'instance_admin'])));

--
-- Name: storage_chunked_sessions_insert; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY storage_chunked_sessions_insert ON chunked_upload_sessions FOR INSERT TO PUBLIC WITH CHECK (has_tenant_access(tenant_id) AND ((auth.current_user_role() = ANY (ARRAY['dashboard_admin', 'service_role', 'admin', 'instance_admin'])) OR ((auth.current_user_id() IS NOT NULL) AND (auth.current_user_id() = owner_id))));

--
-- Name: storage_chunked_sessions_owner; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY storage_chunked_sessions_owner ON chunked_upload_sessions TO PUBLIC USING (has_tenant_access(tenant_id) AND (auth.current_user_id() = owner_id)) WITH CHECK (has_tenant_access(tenant_id) AND (auth.current_user_id() = owner_id));

--
-- Name: storage_object_permissions_admin; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY storage_object_permissions_admin ON object_permissions TO PUBLIC USING (has_tenant_access(tenant_id) AND (auth.current_user_role() = ANY (ARRAY['dashboard_admin', 'service_role', 'admin', 'instance_admin']))) WITH CHECK (has_tenant_access(tenant_id) AND (auth.current_user_role() = ANY (ARRAY['dashboard_admin', 'service_role', 'admin', 'instance_admin'])));

--
-- Name: storage_object_permissions_owner_manage; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY storage_object_permissions_owner_manage ON object_permissions TO PUBLIC USING (has_tenant_access(tenant_id) AND (EXISTS ( SELECT 1 FROM objects WHERE ((objects.id = object_permissions.object_id) AND (objects.owner_id = auth.current_user_id()) AND has_tenant_access(objects.tenant_id))))) WITH CHECK (has_tenant_access(tenant_id) AND (EXISTS ( SELECT 1 FROM objects WHERE ((objects.id = object_permissions.object_id) AND (objects.owner_id = auth.current_user_id()) AND has_tenant_access(objects.tenant_id)))));

--
-- Name: storage_object_permissions_view_shared; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY storage_object_permissions_view_shared ON object_permissions FOR SELECT TO PUBLIC USING (has_tenant_access(tenant_id) AND (user_id = auth.current_user_id()));

--
-- Name: storage_objects_admin; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY storage_objects_admin ON objects TO PUBLIC USING (has_tenant_access(tenant_id) AND (auth.current_user_role() = ANY (ARRAY['dashboard_admin', 'service_role', 'admin', 'instance_admin']))) WITH CHECK (has_tenant_access(tenant_id) AND (auth.current_user_role() = ANY (ARRAY['dashboard_admin', 'service_role', 'admin', 'instance_admin'])));

--
-- Name: storage_objects_insert; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY storage_objects_insert ON objects FOR INSERT TO PUBLIC WITH CHECK (has_tenant_access(tenant_id) AND ((auth.current_user_role() = ANY (ARRAY['dashboard_admin', 'service_role', 'admin', 'instance_admin'])) OR ((auth.current_user_id() IS NOT NULL) AND (auth.current_user_id() = owner_id))));

--
-- Name: storage_objects_owner; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY storage_objects_owner ON objects TO PUBLIC USING (has_tenant_access(tenant_id) AND (auth.current_user_id() = owner_id)) WITH CHECK (has_tenant_access(tenant_id) AND (auth.current_user_id() = owner_id));

--
-- Name: storage_objects_public_read; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY storage_objects_public_read ON objects FOR SELECT TO PUBLIC USING (has_tenant_access(tenant_id) AND is_bucket_public(bucket_id, tenant_id));

--
-- Name: storage_objects_shared_delete; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY storage_objects_shared_delete ON objects FOR DELETE TO PUBLIC USING (has_tenant_access(tenant_id) AND (auth.current_user_id() IS NOT NULL) AND has_object_permission(id, auth.current_user_id(), 'write'));

--
-- Name: storage_objects_shared_read; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY storage_objects_shared_read ON objects FOR SELECT TO PUBLIC USING (has_tenant_access(tenant_id) AND (auth.current_user_id() IS NOT NULL) AND has_object_permission(id, auth.current_user_id(), 'read'));

--
-- Name: storage_objects_shared_write; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY storage_objects_shared_write ON objects FOR UPDATE TO PUBLIC USING (has_tenant_access(tenant_id) AND (auth.current_user_id() IS NOT NULL) AND has_object_permission(id, auth.current_user_id(), 'write'));

--
-- Name: storage_buckets_set_tenant_id; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER storage_buckets_set_tenant_id
    BEFORE INSERT ON buckets
    FOR EACH ROW
    EXECUTE FUNCTION set_tenant_id();

--
-- Name: storage_chunked_sessions_set_tenant_id; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER storage_chunked_sessions_set_tenant_id
    BEFORE INSERT ON chunked_upload_sessions
    FOR EACH ROW
    EXECUTE FUNCTION set_tenant_id();

--
-- Name: storage_object_permissions_set_tenant_id; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER storage_object_permissions_set_tenant_id
    BEFORE INSERT ON object_permissions
    FOR EACH ROW
    EXECUTE FUNCTION set_tenant_id();

--
-- Name: storage_objects_set_tenant_id; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER storage_objects_set_tenant_id
    BEFORE INSERT ON objects
    FOR EACH ROW
    EXECUTE FUNCTION set_tenant_id();

--
-- Name: update_storage_buckets_updated_at; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER update_storage_buckets_updated_at
    BEFORE UPDATE ON buckets
    FOR EACH ROW
    EXECUTE FUNCTION platform.update_updated_at();

--
-- Name: update_storage_objects_updated_at; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER update_storage_objects_updated_at
    BEFORE UPDATE ON objects
    FOR EACH ROW
    EXECUTE FUNCTION platform.update_updated_at();

--
-- Name: bucket_exists(bucket_name text); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION bucket_exists(bucket_name text) TO anon;

--
-- Name: bucket_exists(bucket_name text); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION bucket_exists(bucket_name text) TO authenticated;

--
-- Name: bucket_exists(bucket_name text); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION bucket_exists(bucket_name text) TO fluxbase_app;

--
-- Name: bucket_exists(bucket_name text); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION bucket_exists(bucket_name text) TO fluxbase_rls_test;

--
-- Name: bucket_exists(bucket_name text); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION bucket_exists(bucket_name text) TO service_role;

--
-- Name: bucket_exists(p_bucket_name text, p_tenant_id uuid); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION bucket_exists(p_bucket_name text, p_tenant_id uuid) TO anon;

--
-- Name: bucket_exists(p_bucket_name text, p_tenant_id uuid); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION bucket_exists(p_bucket_name text, p_tenant_id uuid) TO authenticated;

--
-- Name: bucket_exists(p_bucket_name text, p_tenant_id uuid); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION bucket_exists(p_bucket_name text, p_tenant_id uuid) TO fluxbase_app;

--
-- Name: bucket_exists(p_bucket_name text, p_tenant_id uuid); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION bucket_exists(p_bucket_name text, p_tenant_id uuid) TO fluxbase_rls_test;

--
-- Name: bucket_exists(p_bucket_name text, p_tenant_id uuid); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION bucket_exists(p_bucket_name text, p_tenant_id uuid) TO service_role;

--
-- Name: foldername(name text); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION foldername(name text) TO fluxbase_app;

--
-- Name: foldername(name text); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION foldername(name text) TO fluxbase_rls_test;

--
-- Name: get_bucket_settings(bucket_name text); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION get_bucket_settings(bucket_name text) TO anon;

--
-- Name: get_bucket_settings(bucket_name text); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION get_bucket_settings(bucket_name text) TO authenticated;

--
-- Name: get_bucket_settings(bucket_name text); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION get_bucket_settings(bucket_name text) TO fluxbase_app;

--
-- Name: get_bucket_settings(bucket_name text); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION get_bucket_settings(bucket_name text) TO fluxbase_rls_test;

--
-- Name: get_bucket_settings(bucket_name text); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION get_bucket_settings(bucket_name text) TO service_role;

--
-- Name: get_bucket_settings(p_bucket_name text, p_tenant_id uuid); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION get_bucket_settings(p_bucket_name text, p_tenant_id uuid) TO anon;

--
-- Name: get_bucket_settings(p_bucket_name text, p_tenant_id uuid); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION get_bucket_settings(p_bucket_name text, p_tenant_id uuid) TO authenticated;

--
-- Name: get_bucket_settings(p_bucket_name text, p_tenant_id uuid); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION get_bucket_settings(p_bucket_name text, p_tenant_id uuid) TO fluxbase_app;

--
-- Name: get_bucket_settings(p_bucket_name text, p_tenant_id uuid); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION get_bucket_settings(p_bucket_name text, p_tenant_id uuid) TO fluxbase_rls_test;

--
-- Name: get_bucket_settings(p_bucket_name text, p_tenant_id uuid); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION get_bucket_settings(p_bucket_name text, p_tenant_id uuid) TO service_role;

--
-- Name: has_object_permission(p_object_id uuid, p_user_id uuid, p_permission text); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION has_object_permission(p_object_id uuid, p_user_id uuid, p_permission text) TO fluxbase_app;

--
-- Name: has_object_permission(p_object_id uuid, p_user_id uuid, p_permission text); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION has_object_permission(p_object_id uuid, p_user_id uuid, p_permission text) TO fluxbase_rls_test;

--
-- Name: has_tenant_access(p_tenant_id uuid); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION has_tenant_access(p_tenant_id uuid) TO anon;

--
-- Name: has_tenant_access(p_tenant_id uuid); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION has_tenant_access(p_tenant_id uuid) TO authenticated;

--
-- Name: has_tenant_access(p_tenant_id uuid); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION has_tenant_access(p_tenant_id uuid) TO fluxbase_app;

--
-- Name: has_tenant_access(p_tenant_id uuid); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION has_tenant_access(p_tenant_id uuid) TO fluxbase_rls_test;

--
-- Name: has_tenant_access(p_tenant_id uuid); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION has_tenant_access(p_tenant_id uuid) TO service_role;

--
-- Name: is_bucket_public(bucket_name text); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION is_bucket_public(bucket_name text) TO fluxbase_app;

--
-- Name: is_bucket_public(bucket_name text); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION is_bucket_public(bucket_name text) TO fluxbase_rls_test;

--
-- Name: is_bucket_public(p_bucket_name text, p_tenant_id uuid); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION is_bucket_public(p_bucket_name text, p_tenant_id uuid) TO anon;

--
-- Name: is_bucket_public(p_bucket_name text, p_tenant_id uuid); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION is_bucket_public(p_bucket_name text, p_tenant_id uuid) TO authenticated;

--
-- Name: is_bucket_public(p_bucket_name text, p_tenant_id uuid); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION is_bucket_public(p_bucket_name text, p_tenant_id uuid) TO fluxbase_app;

--
-- Name: is_bucket_public(p_bucket_name text, p_tenant_id uuid); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION is_bucket_public(p_bucket_name text, p_tenant_id uuid) TO fluxbase_rls_test;

--
-- Name: is_bucket_public(p_bucket_name text, p_tenant_id uuid); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION is_bucket_public(p_bucket_name text, p_tenant_id uuid) TO service_role;

--
-- Name: set_tenant_id(); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION set_tenant_id() TO fluxbase_app;

--
-- Name: set_tenant_id(); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION set_tenant_id() TO fluxbase_rls_test;

--
-- Name: user_can_access_object(p_object_id uuid, p_required_permission text); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION user_can_access_object(p_object_id uuid, p_required_permission text) TO fluxbase_app;

--
-- Name: user_can_access_object(p_object_id uuid, p_required_permission text); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION user_can_access_object(p_object_id uuid, p_required_permission text) TO fluxbase_rls_test;

--
-- Name: buckets; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT SELECT ON TABLE buckets TO anon;

--
-- Name: buckets; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE buckets TO authenticated;

--
-- Name: buckets; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE buckets TO fluxbase_app;

--
-- Name: buckets; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE buckets TO fluxbase_rls_test;

--
-- Name: buckets; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE buckets TO service_role;

--
-- Name: chunked_upload_sessions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE chunked_upload_sessions TO authenticated;

--
-- Name: chunked_upload_sessions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE chunked_upload_sessions TO fluxbase_app;

--
-- Name: chunked_upload_sessions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE chunked_upload_sessions TO fluxbase_rls_test;

--
-- Name: chunked_upload_sessions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE chunked_upload_sessions TO service_role;

--
-- Name: object_permissions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE object_permissions TO authenticated;

--
-- Name: object_permissions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE object_permissions TO fluxbase_app;

--
-- Name: object_permissions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE object_permissions TO fluxbase_rls_test;

--
-- Name: object_permissions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE object_permissions TO service_role;

--
-- Name: objects; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT SELECT ON TABLE objects TO anon;

--
-- Name: objects; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE objects TO authenticated;

--
-- Name: objects; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE objects TO fluxbase_app;

--
-- Name: objects; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE objects TO fluxbase_rls_test;

--
-- Name: objects; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE objects TO service_role;

