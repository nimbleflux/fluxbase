package api

import (
	"context"
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// EnsureDefaultBucketRecords creates storage.buckets DB records for the
// configured default buckets. This complements the physical bucket creation
// done by storage.Service.EnsureDefaultBuckets(), which only creates
// directories/S3 buckets without inserting DB rows.
//
// Buckets listed in publicBuckets are created with public-read visibility
// (storage.default_public_buckets) and self-healed to it: an operator has no
// way to discover the flag in a seeded row, and an app bucket that renders
// public object URLs is broken while it stays private.
//
// The default tenant's UUID is looked up from platform.tenants. If no default
// tenant exists, rows are inserted with NULL tenant_id (legacy behavior).
func EnsureDefaultBucketRecords(ctx context.Context, db *pgxpool.Pool, bucketNames, publicBuckets []string) error {
	// Resolve the default tenant ID
	var defaultTenantID *string
	var tenantIDStr string
	err := db.QueryRow(ctx, `
		SELECT id::text FROM platform.tenants
		WHERE is_default = true AND deleted_at IS NULL
		LIMIT 1
	`).Scan(&tenantIDStr)
	if err == nil {
		defaultTenantID = &tenantIDStr
	}

	// Self-heal: earlier versions seeded default buckets with id = gen_random_uuid()
	// while every INSERT into storage.objects uses the bucket NAME as bucket_id.
	// That mismatch violated objects_bucket_id_fkey on fresh instances. Align
	// id = name so existing rows resolve. Safe: the FK bug blocked all object
	// inserts on these buckets, so nothing references the old UUID ids.
	if _, err := db.Exec(ctx, `
		UPDATE storage.buckets
		SET id = name
		WHERE name = ANY($1::text[])
		  AND id <> name
	`, bucketNames); err != nil {
		return fmt.Errorf("failed to align default bucket ids with names: %w", err)
	}

	for _, name := range bucketNames {
		var exists bool
		err := db.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM storage.buckets
				WHERE name = $1
				AND (($2::uuid IS NULL AND tenant_id IS NULL) OR tenant_id = $2::uuid)
			)
		`, name, defaultTenantID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check bucket record for %q: %w", name, err)
		}
		if exists {
			continue
		}

		// Use the bucket name as its id. storage.objects.bucket_id is always
		// populated with the name (from the URL path), so id must equal name
		// for objects_bucket_id_fkey to resolve. This matches CreateBucket.
		public := slices.Contains(publicBuckets, name)
		_, err = db.Exec(ctx, `
			INSERT INTO storage.buckets (id, name, public, tenant_id)
			VALUES ($1, $1, $3, $2)
			ON CONFLICT (name) DO NOTHING
		`, name, defaultTenantID, public)
		if err != nil {
			return fmt.Errorf("failed to create bucket record for %q: %w", name, err)
		}
		log.Info().Str("bucket", name).Bool("public", public).Msg("Created default bucket DB record")
	}

	// Heal rows seeded before the bucket was declared public — seeding never
	// revisits existing rows, so deployments that upgrade keep the private
	// visibility their app-level public URLs depend on.
	for _, name := range publicBuckets {
		tag, err := db.Exec(ctx, `
			UPDATE storage.buckets
			SET public = true
			WHERE name = $1
			  AND public = false
		`, name)
		if err != nil {
			return fmt.Errorf("failed to heal bucket visibility for %q: %w", name, err)
		}
		if tag.RowsAffected() > 0 {
			log.Info().Str("bucket", name).Msg("Healed default bucket to public-read")
		}
	}

	return nil
}
