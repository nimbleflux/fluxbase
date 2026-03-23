package resources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nimbleflux/fluxbase/internal/mcp"
)

func TestNewFunctionsResource(t *testing.T) {
	t.Run("creates resource with nil storage", func(t *testing.T) {
		resource := NewFunctionsResource(nil)
		assert.NotNil(t, resource)
		assert.Nil(t, resource.storage)
	})
}

func TestFunctionsResource_Metadata(t *testing.T) {
	resource := NewFunctionsResource(nil)

	t.Run("URI", func(t *testing.T) {
		assert.Equal(t, "fluxbase://functions", resource.URI())
	})

	t.Run("Name", func(t *testing.T) {
		assert.Equal(t, "Edge Functions", resource.Name())
	})

	t.Run("Description", func(t *testing.T) {
		desc := resource.Description()
		assert.Contains(t, desc, "edge functions")
		assert.Contains(t, desc, "metadata")
	})

	t.Run("MimeType", func(t *testing.T) {
		assert.Equal(t, "application/json", resource.MimeType())
	})

	t.Run("RequiredScopes", func(t *testing.T) {
		scopes := resource.RequiredScopes()
		assert.Contains(t, scopes, mcp.ScopeExecuteFunctions)
	})
}

func TestFunctionsResource_Read(t *testing.T) {
	t.Run("returns error when storage is nil", func(t *testing.T) {
		resource := NewFunctionsResource(nil)
		_, err := resource.Read(context.Background(), nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "storage not available")
	})
}
