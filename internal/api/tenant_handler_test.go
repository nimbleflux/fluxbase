package api

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/tenantdb"
)

// =============================================================================
// TenantHandler Construction Tests
// =============================================================================

func TestNewTenantHandler(t *testing.T) {
	t.Run("creates handler with nil dependencies", func(t *testing.T) {
		handler := NewTenantHandler(nil, nil, nil, nil, nil, nil)
		assert.NotNil(t, handler)
		assert.Nil(t, handler.DB)
		assert.Nil(t, handler.Manager)
		assert.Nil(t, handler.Storage)
	})
}

// =============================================================================
// CreateTenantRequest Tests
// =============================================================================

func TestCreateTenantRequest_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		email := "admin@example.com"
		userID := uuid.New().String()

		req := CreateTenantRequest{
			Slug:             "acme-corp",
			Name:             "Acme Corporation",
			AutoGenerateKeys: true,
			AdminEmail:       &email,
			AdminUserID:      &userID,
			SendKeysToEmail:  true,
		}

		assert.Equal(t, "acme-corp", req.Slug)
		assert.Equal(t, "Acme Corporation", req.Name)
		assert.True(t, req.AutoGenerateKeys)
		assert.Equal(t, email, *req.AdminEmail)
		assert.Equal(t, userID, *req.AdminUserID)
		assert.True(t, req.SendKeysToEmail)
	})

	t.Run("minimal request", func(t *testing.T) {
		req := CreateTenantRequest{
			Slug: "minimal-tenant",
			Name: "Minimal Tenant",
		}

		assert.Equal(t, "minimal-tenant", req.Slug)
		assert.Equal(t, "Minimal Tenant", req.Name)
		assert.False(t, req.AutoGenerateKeys)
		assert.Nil(t, req.AdminEmail)
		assert.Nil(t, req.AdminUserID)
		assert.False(t, req.SendKeysToEmail)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{
			"slug": "test-tenant",
			"name": "Test Tenant",
			"auto_generate_keys": true,
			"admin_email": "admin@test.com",
			"send_keys_to_email": true
		}`

		var req CreateTenantRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "test-tenant", req.Slug)
		assert.Equal(t, "Test Tenant", req.Name)
		assert.True(t, req.AutoGenerateKeys)
		require.NotNil(t, req.AdminEmail)
		assert.Equal(t, "admin@test.com", *req.AdminEmail)
		assert.True(t, req.SendKeysToEmail)
	})

	t.Run("JSON deserialization with metadata", func(t *testing.T) {
		jsonData := `{
			"slug": "meta-tenant",
			"name": "Meta Tenant",
			"metadata": {"plan": "enterprise", "region": "us-east"}
		}`

		var req CreateTenantRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "meta-tenant", req.Slug)
		require.NotNil(t, req.Metadata)
		assert.Equal(t, "enterprise", req.Metadata["plan"])
		assert.Equal(t, "us-east", req.Metadata["region"])
	})
}

// =============================================================================
// UpdateTenantRequest Tests
// =============================================================================

func TestUpdateTenantRequest_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		name := "Updated Name"
		req := UpdateTenantRequest{
			Name:     &name,
			Metadata: map[string]interface{}{"updated": true},
		}

		assert.Equal(t, "Updated Name", *req.Name)
		assert.Equal(t, true, req.Metadata["updated"])
	})

	t.Run("empty request", func(t *testing.T) {
		req := UpdateTenantRequest{}
		assert.Nil(t, req.Name)
		assert.Nil(t, req.Metadata)
	})
}

// =============================================================================
// AssignAdminRequest Tests
// =============================================================================

func TestAssignAdminRequest_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		req := AssignAdminRequest{
			UserID: uuid.New().String(),
		}
		assert.NotEmpty(t, req.UserID)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		userID := uuid.New()
		jsonData := `{"user_id":"` + userID.String() + `"}`

		var req AssignAdminRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)
		assert.Equal(t, userID.String(), req.UserID)
	})
}

// =============================================================================
// TenantResponse Tests
// =============================================================================

func TestTenantResponse(t *testing.T) {
	t.Run("tenantToResponse maps all fields", func(t *testing.T) {
		now := time.Now()
		dbName := "tenant_acme"
		tenant := &tenantdb.Tenant{
			ID:        uuid.New().String(),
			Slug:      "acme-corp",
			Name:      "Acme Corporation",
			DBName:    &dbName,
			Status:    tenantdb.TenantStatusActive,
			IsDefault: false,
			Metadata:  map[string]interface{}{"plan": "pro"},
			CreatedAt: now,
			UpdatedAt: now,
		}

		resp := tenantToResponse(tenant)

		assert.Equal(t, tenant.ID, resp.ID)
		assert.Equal(t, "acme-corp", resp.Slug)
		assert.Equal(t, "Acme Corporation", resp.Name)
		require.NotNil(t, resp.DbName)
		assert.Equal(t, "tenant_acme", *resp.DbName)
		assert.Equal(t, "active", resp.Status)
		assert.False(t, resp.IsDefault)
		assert.Equal(t, "pro", resp.Metadata["plan"])
	})

	t.Run("tenantToResponse with nil DbName", func(t *testing.T) {
		tenant := &tenantdb.Tenant{
			ID:        uuid.New().String(),
			Slug:      "default-tenant",
			Name:      "Default",
			DBName:    nil,
			Status:    tenantdb.TenantStatusActive,
			IsDefault: true,
		}

		resp := tenantToResponse(tenant)
		assert.Nil(t, resp.DbName)
		assert.True(t, resp.IsDefault)
	})
}

// =============================================================================
// isValidSlug Tests
// =============================================================================

func TestIsValidSlug(t *testing.T) {
	tests := []struct {
		name     string
		slug     string
		expected bool
	}{
		{"valid simple slug", "acme", true},
		{"valid hyphenated slug", "acme-corp", true},
		{"valid with numbers", "tenant-123", true},
		{"valid with trailing number", "tenant1", true},
		{"empty string", "", false},
		{"too long", "a-very-long-slug-that-exceeds-the-maximum-allowed-length-of-63-characters-xxxxxxxxx", false},
		{"starts with number", "1tenant", false},
		{"starts with hyphen", "-tenant", false},
		{"ends with hyphen", "tenant-", false},
		{"contains uppercase", "Acme", false},
		{"contains space", "acme corp", false},
		{"contains underscore", "acme_corp", false},
		{"contains special char", "acme@corp", false},
		{"single char", "a", true},
		{"max length 63", "a-very-long-slug-that-is-exactly-sixty-three-characters-xxxxxxx", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isValidSlug(tt.slug))
		})
	}
}
