package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/mcp"
)

// =============================================================================
// isInternalSchema Tests
// =============================================================================

func TestIsInternalSchema(t *testing.T) {
	internalSchemas := []string{
		"auth",
		"storage",
		"realtime",
		"functions",
		"jobs",
		"rpc",
		"logging",
		"ai",
		"branching",
		"pg_catalog",
		"information_schema",
	}

	for _, schema := range internalSchemas {
		t.Run(schema+" is internal", func(t *testing.T) {
			assert.True(t, isInternalSchema(schema))
		})
	}

	publicSchemas := []string{
		"public",
		"custom_schema",
		"myapp",
		"users",
		"",
	}

	for _, schema := range publicSchemas {
		t.Run(schema+" is not internal", func(t *testing.T) {
			assert.False(t, isInternalSchema(schema))
		})
	}
}

// =============================================================================
// canAccessInternalSchemas Tests
// =============================================================================

func TestCanAccessInternalSchemas(t *testing.T) {
	t.Run("service role has access", func(t *testing.T) {
		authCtx := &mcp.AuthContext{
			IsServiceRole: true,
		}
		assert.True(t, canAccessInternalSchemas(authCtx))
	})

	t.Run("user with admin:schemas scope has access", func(t *testing.T) {
		authCtx := &mcp.AuthContext{
			IsServiceRole: false,
			Scopes:        []string{mcp.ScopeAdminSchemas},
		}
		assert.True(t, canAccessInternalSchemas(authCtx))
	})

	t.Run("user with wildcard scope has access", func(t *testing.T) {
		authCtx := &mcp.AuthContext{
			IsServiceRole: false,
			Scopes:        []string{"*"},
		}
		assert.True(t, canAccessInternalSchemas(authCtx))
	})

	t.Run("user without admin scope denied", func(t *testing.T) {
		authCtx := &mcp.AuthContext{
			IsServiceRole: false,
			Scopes:        []string{mcp.ScopeReadTables, mcp.ScopeWriteTables},
		}
		assert.False(t, canAccessInternalSchemas(authCtx))
	})

	t.Run("user with empty scopes denied", func(t *testing.T) {
		authCtx := &mcp.AuthContext{
			IsServiceRole: false,
			Scopes:        []string{},
		}
		assert.False(t, canAccessInternalSchemas(authCtx))
	})
}

// =============================================================================
// SchemaResource Metadata Tests
// =============================================================================

func TestSchemaResource_URI(t *testing.T) {
	r := &SchemaResource{}
	assert.Equal(t, "fluxbase://schema/tables", r.URI())
}

func TestSchemaResource_Name(t *testing.T) {
	r := &SchemaResource{}
	assert.Equal(t, "Database Schema", r.Name())
}

func TestSchemaResource_Description(t *testing.T) {
	r := &SchemaResource{}
	desc := r.Description()
	assert.Contains(t, desc, "database schema")
	assert.Contains(t, desc, "tables")
}

func TestSchemaResource_MimeType(t *testing.T) {
	r := &SchemaResource{}
	assert.Equal(t, "application/json", r.MimeType())
}

func TestSchemaResource_RequiredScopes(t *testing.T) {
	r := &SchemaResource{}
	scopes := r.RequiredScopes()
	require.Len(t, scopes, 1)
	assert.Equal(t, mcp.ScopeReadTables, scopes[0])
}

func TestNewSchemaResource(t *testing.T) {
	t.Run("creates with nil schema cache", func(t *testing.T) {
		r := NewSchemaResource(nil)
		require.NotNil(t, r)
		assert.Nil(t, r.schemaCache)
	})
}

// =============================================================================
// TableResource Metadata Tests
// =============================================================================

func TestTableResource_URI(t *testing.T) {
	r := &TableResource{}
	assert.Equal(t, "fluxbase://schema/tables/{schema}/{table}", r.URI())
}

func TestTableResource_Name(t *testing.T) {
	r := &TableResource{}
	assert.Equal(t, "Table Details", r.Name())
}

func TestTableResource_Description(t *testing.T) {
	r := &TableResource{}
	desc := r.Description()
	assert.Contains(t, desc, "table")
}

func TestTableResource_MimeType(t *testing.T) {
	r := &TableResource{}
	assert.Equal(t, "application/json", r.MimeType())
}

func TestTableResource_RequiredScopes(t *testing.T) {
	r := &TableResource{}
	scopes := r.RequiredScopes()
	require.Len(t, scopes, 1)
	assert.Equal(t, mcp.ScopeReadTables, scopes[0])
}

func TestTableResource_IsTemplate(t *testing.T) {
	r := &TableResource{}
	assert.True(t, r.IsTemplate())
}

func TestNewTableResource(t *testing.T) {
	t.Run("creates with nil schema cache", func(t *testing.T) {
		r := NewTableResource(nil)
		require.NotNil(t, r)
		assert.Nil(t, r.schemaCache)
	})
}

// =============================================================================
// TableResource.MatchURI Tests
// =============================================================================

func TestTableResource_MatchURI(t *testing.T) {
	r := &TableResource{}

	tests := []struct {
		name           string
		uri            string
		expectedMatch  bool
		expectedParams map[string]string
	}{
		{
			name:          "valid public.users URI",
			uri:           "fluxbase://schema/tables/public/users",
			expectedMatch: true,
			expectedParams: map[string]string{
				"schema": "public",
				"table":  "users",
			},
		},
		{
			name:          "valid auth.sessions URI",
			uri:           "fluxbase://schema/tables/auth/sessions",
			expectedMatch: true,
			expectedParams: map[string]string{
				"schema": "auth",
				"table":  "sessions",
			},
		},
		{
			name:          "valid custom_schema.my_table URI",
			uri:           "fluxbase://schema/tables/custom_schema/my_table",
			expectedMatch: true,
			expectedParams: map[string]string{
				"schema": "custom_schema",
				"table":  "my_table",
			},
		},
		{
			name:           "invalid prefix",
			uri:            "other://schema/tables/public/users",
			expectedMatch:  false,
			expectedParams: nil,
		},
		{
			name:           "missing table",
			uri:            "fluxbase://schema/tables/public",
			expectedMatch:  false,
			expectedParams: nil,
		},
		{
			name:           "missing schema and table",
			uri:            "fluxbase://schema/tables/",
			expectedMatch:  false,
			expectedParams: nil,
		},
		{
			name:           "wrong path",
			uri:            "fluxbase://storage/buckets",
			expectedMatch:  false,
			expectedParams: nil,
		},
		{
			name:           "empty URI",
			uri:            "",
			expectedMatch:  false,
			expectedParams: nil,
		},
		{
			name:          "extra path segments handled",
			uri:           "fluxbase://schema/tables/public/users/extra",
			expectedMatch: true,
			expectedParams: map[string]string{
				"schema": "public",
				"table":  "users/extra", // SplitN with 2 keeps rest together
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, matched := r.MatchURI(tt.uri)
			assert.Equal(t, tt.expectedMatch, matched)
			if tt.expectedMatch {
				assert.Equal(t, tt.expectedParams, params)
			} else {
				assert.Nil(t, params)
			}
		})
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkIsInternalSchema_Internal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isInternalSchema("auth")
	}
}

func BenchmarkIsInternalSchema_Public(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isInternalSchema("public")
	}
}

func BenchmarkTableResource_MatchURI(b *testing.B) {
	r := &TableResource{}
	uri := "fluxbase://schema/tables/public/users"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.MatchURI(uri)
	}
}
