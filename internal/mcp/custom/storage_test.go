package custom

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/mcp"

	"github.com/nimbleflux/fluxbase/internal/database"
)

// =============================================================================
// NewStorage Tests
// =============================================================================

func TestNewStorage(t *testing.T) {
	t.Run("creates storage with nil pool", func(t *testing.T) {
		storage := NewStorage(nil)

		require.NotNil(t, storage)
		assert.Nil(t, storage.DB)
	})
}

// =============================================================================
// Storage Struct Tests
// =============================================================================

func TestStorage_Struct(t *testing.T) {
	t.Run("stores database pool", func(t *testing.T) {
		storage := &Storage{
			TenantAware: database.TenantAware{DB: nil},
		}

		assert.Nil(t, storage.DB)
	})
}

// =============================================================================
// Error Constants Tests
// =============================================================================

func TestStorageErrors(t *testing.T) {
	t.Run("ErrToolNotFound has correct message", func(t *testing.T) {
		assert.Equal(t, "custom tool not found", ErrToolNotFound.Error())
	})

	t.Run("ErrResourceNotFound has correct message", func(t *testing.T) {
		assert.Equal(t, "custom resource not found", ErrResourceNotFound.Error())
	})

	t.Run("ErrToolAlreadyExists has correct message", func(t *testing.T) {
		assert.Equal(t, "custom tool with this name already exists in namespace", ErrToolAlreadyExists.Error())
	})

	t.Run("ErrResourceExists has correct message", func(t *testing.T) {
		assert.Equal(t, "custom resource with this URI already exists in namespace", ErrResourceExists.Error())
	})

	t.Run("ErrInvalidInputSchema has correct message", func(t *testing.T) {
		assert.Equal(t, "invalid input schema: must be a valid JSON Schema object", ErrInvalidInputSchema.Error())
	})

	t.Run("errors are distinct", func(t *testing.T) {
		assert.NotEqual(t, ErrToolNotFound, ErrResourceNotFound)
		assert.NotEqual(t, ErrToolAlreadyExists, ErrResourceExists)
	})

	t.Run("errors can be compared with errors.Is", func(t *testing.T) {
		err := ErrToolNotFound
		assert.True(t, errors.Is(err, ErrToolNotFound))
		assert.False(t, errors.Is(err, ErrResourceNotFound))
	})
}

// =============================================================================
// isUniqueViolation Tests
// =============================================================================

func TestIsUniqueViolation(t *testing.T) {
	t.Run("returns false for nil error", func(t *testing.T) {
		result := isUniqueViolation(nil)
		assert.False(t, result)
	})

	t.Run("returns true for duplicate key error", func(t *testing.T) {
		err := errors.New("ERROR: duplicate key value violates unique constraint")
		result := isUniqueViolation(err)
		assert.True(t, result)
	})

	t.Run("returns true for unique constraint error", func(t *testing.T) {
		err := errors.New("unique constraint violation on column")
		result := isUniqueViolation(err)
		assert.True(t, result)
	})

	t.Run("returns true for PostgreSQL error code 23505", func(t *testing.T) {
		err := errors.New("pq: error code 23505")
		result := isUniqueViolation(err)
		assert.True(t, result)
	})

	t.Run("returns false for other errors", func(t *testing.T) {
		err := errors.New("connection refused")
		result := isUniqueViolation(err)
		assert.False(t, result)
	})

	t.Run("returns false for unrelated error", func(t *testing.T) {
		err := errors.New("timeout exceeded")
		result := isUniqueViolation(err)
		assert.False(t, result)
	})
}

// =============================================================================
// contains Helper Tests
// =============================================================================

func TestContains(t *testing.T) {
	t.Run("returns true when substring present", func(t *testing.T) {
		result := contains("hello world", "world")
		assert.True(t, result)
	})

	t.Run("returns false when substring not present", func(t *testing.T) {
		result := contains("hello world", "foo")
		assert.False(t, result)
	})

	t.Run("returns true for exact match", func(t *testing.T) {
		result := contains("test", "test")
		assert.True(t, result)
	})

	t.Run("returns true for substring at start", func(t *testing.T) {
		result := contains("hello world", "hello")
		assert.True(t, result)
	})

	t.Run("returns true for substring at end", func(t *testing.T) {
		result := contains("hello world", "world")
		assert.True(t, result)
	})

	t.Run("returns false for empty string", func(t *testing.T) {
		result := contains("", "test")
		assert.False(t, result)
	})

	t.Run("handles empty substring", func(t *testing.T) {
		result := contains("hello", "")
		assert.True(t, result)
	})

	t.Run("returns false when string shorter than substring", func(t *testing.T) {
		result := contains("hi", "hello")
		assert.False(t, result)
	})
}

// =============================================================================
// containsAt Helper Tests
// =============================================================================

func TestContainsAt(t *testing.T) {
	t.Run("finds substring at specified position", func(t *testing.T) {
		result := containsAt("hello world", "world", 6)
		assert.True(t, result)
	})

	t.Run("finds substring from start", func(t *testing.T) {
		result := containsAt("hello world", "hello", 0)
		assert.True(t, result)
	})

	t.Run("returns false when not at position", func(t *testing.T) {
		result := containsAt("hello world", "hello", 6)
		assert.False(t, result)
	})

	t.Run("handles substring in middle", func(t *testing.T) {
		result := containsAt("abcdefg", "cde", 0)
		assert.True(t, result)
	})
}

// =============================================================================
// CustomTool Struct Tests (from types.go)
// =============================================================================

func TestCustomToolStorage_Struct(t *testing.T) {
	t.Run("stores all fields correctly", func(t *testing.T) {
		id := uuid.New()
		createdBy := uuid.New()
		now := time.Now()

		tool := &CustomTool{
			ID:             id,
			Name:           "my_tool",
			Namespace:      "production",
			Description:    "A test tool",
			Code:           "function handler() {}",
			InputSchema:    map[string]any{"type": "object"},
			RequiredScopes: []string{"read", "write"},
			TimeoutSeconds: 30,
			MemoryLimitMB:  128,
			AllowNet:       true,
			AllowEnv:       false,
			AllowRead:      true,
			AllowWrite:     false,
			Enabled:        true,
			Version:        1,
			CreatedBy:      &createdBy,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		assert.Equal(t, id, tool.ID)
		assert.Equal(t, "my_tool", tool.Name)
		assert.Equal(t, "production", tool.Namespace)
		assert.Equal(t, "A test tool", tool.Description)
		assert.Equal(t, "function handler() {}", tool.Code)
		assert.NotNil(t, tool.InputSchema)
		assert.Equal(t, []string{"read", "write"}, tool.RequiredScopes)
		assert.Equal(t, 30, tool.TimeoutSeconds)
		assert.Equal(t, 128, tool.MemoryLimitMB)
		assert.True(t, tool.AllowNet)
		assert.False(t, tool.AllowEnv)
		assert.True(t, tool.AllowRead)
		assert.False(t, tool.AllowWrite)
		assert.True(t, tool.Enabled)
		assert.Equal(t, 1, tool.Version)
		assert.NotNil(t, tool.CreatedBy)
	})

	t.Run("handles nil created_by", func(t *testing.T) {
		tool := &CustomTool{
			Name:      "test",
			CreatedBy: nil,
		}

		assert.Nil(t, tool.CreatedBy)
	})
}

// =============================================================================
// CustomResource Struct Tests (from types.go)
// =============================================================================

func TestCustomResourceStorage_Struct(t *testing.T) {
	t.Run("stores all fields correctly", func(t *testing.T) {
		id := uuid.New()
		createdBy := uuid.New()
		now := time.Now()

		resource := &CustomResource{
			ID:              id,
			URI:             "fluxbase://tables/{table}",
			Name:            "Table Resource",
			Namespace:       "default",
			Description:     "Access table data",
			MimeType:        "application/json",
			Code:            "function handler(params) {}",
			IsTemplate:      true,
			RequiredScopes:  []string{"read"},
			TimeoutSeconds:  10,
			CacheTTLSeconds: 60,
			Enabled:         true,
			Version:         1,
			CreatedBy:       &createdBy,
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		assert.Equal(t, id, resource.ID)
		assert.Equal(t, "fluxbase://tables/{table}", resource.URI)
		assert.Equal(t, "Table Resource", resource.Name)
		assert.Equal(t, "default", resource.Namespace)
		assert.Equal(t, "Access table data", resource.Description)
		assert.Equal(t, "application/json", resource.MimeType)
		assert.True(t, resource.IsTemplate)
		assert.Equal(t, []string{"read"}, resource.RequiredScopes)
		assert.Equal(t, 10, resource.TimeoutSeconds)
		assert.Equal(t, 60, resource.CacheTTLSeconds)
		assert.True(t, resource.Enabled)
	})
}

// =============================================================================
// CreateToolRequest Struct Tests
// =============================================================================

func TestCreateToolRequest_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		allowNet := true
		allowEnv := false
		enabled := true

		req := &CreateToolRequest{
			Name:           "new_tool",
			Namespace:      "test",
			Description:    "A new tool",
			Code:           "export default handler() {}",
			InputSchema:    map[string]any{"type": "object", "properties": map[string]any{}},
			RequiredScopes: []string{"admin"},
			TimeoutSeconds: 60,
			MemoryLimitMB:  256,
			AllowNet:       &allowNet,
			AllowEnv:       &allowEnv,
			AllowRead:      nil,
			AllowWrite:     nil,
			Enabled:        &enabled,
		}

		assert.Equal(t, "new_tool", req.Name)
		assert.Equal(t, "test", req.Namespace)
		assert.True(t, *req.AllowNet)
		assert.False(t, *req.AllowEnv)
		assert.Nil(t, req.AllowRead)
		assert.Nil(t, req.AllowWrite)
	})

	t.Run("handles nil pointers", func(t *testing.T) {
		req := &CreateToolRequest{
			Name: "minimal",
			Code: "handler()",
		}

		assert.Nil(t, req.AllowNet)
		assert.Nil(t, req.AllowEnv)
		assert.Nil(t, req.AllowRead)
		assert.Nil(t, req.AllowWrite)
		assert.Nil(t, req.Enabled)
	})
}

// =============================================================================
// UpdateToolRequest Struct Tests
// =============================================================================

func TestUpdateToolRequest_Struct(t *testing.T) {
	t.Run("stores optional fields", func(t *testing.T) {
		name := "updated_name"
		code := "new code"
		timeout := 45

		req := &UpdateToolRequest{
			Name:           &name,
			Code:           &code,
			TimeoutSeconds: &timeout,
		}

		assert.Equal(t, "updated_name", *req.Name)
		assert.Equal(t, "new code", *req.Code)
		assert.Equal(t, 45, *req.TimeoutSeconds)
		assert.Nil(t, req.Description)
	})

	t.Run("handles all nil fields", func(t *testing.T) {
		req := &UpdateToolRequest{}

		assert.Nil(t, req.Name)
		assert.Nil(t, req.Description)
		assert.Nil(t, req.Code)
		assert.Nil(t, req.TimeoutSeconds)
	})
}

// =============================================================================
// CreateResourceRequest Struct Tests
// =============================================================================

func TestCreateResourceRequest_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		isTemplate := true
		timeout := 15
		cacheTTL := 120

		req := &CreateResourceRequest{
			URI:             "fluxbase://resource",
			Name:            "My Resource",
			Namespace:       "prod",
			Description:     "A resource",
			MimeType:        "text/plain",
			Code:            "handler()",
			IsTemplate:      &isTemplate,
			RequiredScopes:  []string{"read"},
			TimeoutSeconds:  &timeout,
			CacheTTLSeconds: &cacheTTL,
		}

		assert.Equal(t, "fluxbase://resource", req.URI)
		assert.Equal(t, "My Resource", req.Name)
		assert.True(t, *req.IsTemplate)
		assert.Equal(t, 15, *req.TimeoutSeconds)
		assert.Equal(t, 120, *req.CacheTTLSeconds)
	})
}

// =============================================================================
// UpdateResourceRequest Struct Tests
// =============================================================================

func TestUpdateResourceRequest_Struct(t *testing.T) {
	t.Run("stores optional fields", func(t *testing.T) {
		uri := "new://uri"
		name := "New Name"

		req := &UpdateResourceRequest{
			URI:  &uri,
			Name: &name,
		}

		assert.Equal(t, "new://uri", *req.URI)
		assert.Equal(t, "New Name", *req.Name)
		assert.Nil(t, req.Description)
	})
}

// =============================================================================
// ListToolsFilter Struct Tests
// =============================================================================

func TestListToolsFilter_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		filter := ListToolsFilter{
			Namespace:   "production",
			EnabledOnly: true,
			Limit:       10,
			Offset:      5,
		}

		assert.Equal(t, "production", filter.Namespace)
		assert.True(t, filter.EnabledOnly)
		assert.Equal(t, 10, filter.Limit)
		assert.Equal(t, 5, filter.Offset)
	})

	t.Run("defaults to zero values", func(t *testing.T) {
		filter := ListToolsFilter{}

		assert.Equal(t, "", filter.Namespace)
		assert.False(t, filter.EnabledOnly)
		assert.Equal(t, 0, filter.Limit)
		assert.Equal(t, 0, filter.Offset)
	})
}

// =============================================================================
// ListResourcesFilter Struct Tests
// =============================================================================

func TestListResourcesFilter_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		filter := ListResourcesFilter{
			Namespace:   "staging",
			EnabledOnly: false,
			Limit:       25,
			Offset:      10,
		}

		assert.Equal(t, "staging", filter.Namespace)
		assert.False(t, filter.EnabledOnly)
		assert.Equal(t, 25, filter.Limit)
		assert.Equal(t, 10, filter.Offset)
	})
}

// =============================================================================
// ToolExecutionResult Struct Tests
// =============================================================================

func TestToolExecutionResult_Struct(t *testing.T) {
	t.Run("stores successful result", func(t *testing.T) {
		result := ToolExecutionResult{
			Success: true,
			Content: []mcp.Content{
				{Type: mcp.ContentTypeText, Text: "Hello"},
			},
			DurationMs: 150,
			Logs:       "Execution completed",
			Metadata:   map[string]any{"version": 1},
		}

		assert.True(t, result.Success)
		assert.Len(t, result.Content, 1)
		assert.Equal(t, int64(150), result.DurationMs)
		assert.Equal(t, "Execution completed", result.Logs)
		assert.NotNil(t, result.Metadata)
	})

	t.Run("stores error result", func(t *testing.T) {
		result := ToolExecutionResult{
			Success:    false,
			Error:      "Execution failed",
			DurationMs: 10,
		}

		assert.False(t, result.Success)
		assert.Equal(t, "Execution failed", result.Error)
		assert.Empty(t, result.Content)
	})
}

// =============================================================================
// ResourceReadResult Struct Tests
// =============================================================================

func TestResourceReadResult_Struct(t *testing.T) {
	t.Run("stores successful result", func(t *testing.T) {
		result := ResourceReadResult{
			Success: true,
			Content: []mcp.Content{
				{Type: mcp.ContentTypeText, Text: "Resource data"},
			},
			DurationMs: 25,
		}

		assert.True(t, result.Success)
		assert.Len(t, result.Content, 1)
		assert.Equal(t, int64(25), result.DurationMs)
	})
}

// =============================================================================
// Content Struct Tests (mcp.Content)
// =============================================================================

func TestContent_Struct(t *testing.T) {
	t.Run("stores text content", func(t *testing.T) {
		content := mcp.Content{
			Type: mcp.ContentTypeText,
			Text: "Hello World",
		}

		assert.Equal(t, mcp.ContentTypeText, content.Type)
		assert.Equal(t, "Hello World", content.Text)
	})

	t.Run("stores image content", func(t *testing.T) {
		content := mcp.Content{
			Type:     mcp.ContentTypeImage,
			MimeType: "image/png",
			Data:     "base64encodeddata",
		}

		assert.Equal(t, mcp.ContentTypeImage, content.Type)
		assert.Equal(t, "image/png", content.MimeType)
		assert.Equal(t, "base64encodeddata", content.Data)
	})

	t.Run("stores resource content", func(t *testing.T) {
		content := mcp.Content{
			Type: mcp.ContentTypeResource,
			URI:  "fluxbase://tables/users",
		}

		assert.Equal(t, mcp.ContentTypeResource, content.Type)
		assert.Equal(t, "fluxbase://tables/users", content.URI)
	})
}

// =============================================================================
// SyncToolRequest Struct Tests
// =============================================================================

func TestSyncToolRequest_Struct(t *testing.T) {
	t.Run("stores upsert flag with embedded request", func(t *testing.T) {
		req := SyncToolRequest{
			CreateToolRequest: CreateToolRequest{
				Name:        "sync_tool",
				Code:        "handler()",
				Description: "Synced tool",
			},
			Upsert: true,
		}

		assert.Equal(t, "sync_tool", req.Name)
		assert.Equal(t, "handler()", req.Code)
		assert.True(t, req.Upsert)
	})

	t.Run("defaults upsert to false", func(t *testing.T) {
		req := SyncToolRequest{
			CreateToolRequest: CreateToolRequest{
				Name: "tool",
				Code: "code",
			},
		}

		assert.False(t, req.Upsert)
	})
}

// =============================================================================
// SyncResourceRequest Struct Tests
// =============================================================================

func TestSyncResourceRequest_Struct(t *testing.T) {
	t.Run("stores upsert flag with embedded request", func(t *testing.T) {
		req := SyncResourceRequest{
			CreateResourceRequest: CreateResourceRequest{
				URI:  "fluxbase://resource",
				Name: "Synced Resource",
				Code: "handler()",
			},
			Upsert: true,
		}

		assert.Equal(t, "fluxbase://resource", req.URI)
		assert.Equal(t, "Synced Resource", req.Name)
		assert.True(t, req.Upsert)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkIsUniqueViolation_Match(b *testing.B) {
	err := errors.New("ERROR: duplicate key value violates unique constraint")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isUniqueViolation(err)
	}
}

func BenchmarkIsUniqueViolation_NoMatch(b *testing.B) {
	err := errors.New("connection timeout")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isUniqueViolation(err)
	}
}

func BenchmarkContains_Found(b *testing.B) {
	s := "ERROR: duplicate key value violates unique constraint"
	substr := "duplicate key"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		contains(s, substr)
	}
}

func BenchmarkContains_NotFound(b *testing.B) {
	s := "ERROR: connection refused"
	substr := "duplicate key"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		contains(s, substr)
	}
}
