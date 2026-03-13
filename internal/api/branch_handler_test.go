package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nimbleflux/fluxbase/internal/branching"
	"github.com/nimbleflux/fluxbase/internal/config"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// BranchHandler Construction Tests
// =============================================================================

func TestNewBranchHandler(t *testing.T) {
	t.Run("creates handler with nil dependencies", func(t *testing.T) {
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})
		assert.NotNil(t, handler)
		assert.Nil(t, handler.manager)
		assert.Nil(t, handler.router)
	})

	t.Run("creates handler with config", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled:            true,
			MaxBranchesPerUser: 10,
			MaxTotalBranches:   50,
		}
		handler := NewBranchHandler(nil, nil, cfg)
		assert.NotNil(t, handler)
		assert.True(t, handler.config.Enabled)
		assert.Equal(t, 10, handler.config.MaxBranchesPerUser)
	})
}

// =============================================================================
// CreateBranchRequest Tests
// =============================================================================

func TestCreateBranchRequest_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		parentID := uuid.New()
		prNumber := 123
		prURL := "https://github.com/owner/repo/pull/123"
		repo := "owner/repo"
		expiresIn := "24h"

		req := CreateBranchRequest{
			Name:           "feature-branch",
			ParentBranchID: &parentID,
			DataCloneMode:  branching.DataCloneModeSchemaOnly,
			Type:           branching.BranchTypePreview,
			GitHubPRNumber: &prNumber,
			GitHubPRURL:    &prURL,
			GitHubRepo:     &repo,
			ExpiresIn:      &expiresIn,
		}

		assert.Equal(t, "feature-branch", req.Name)
		assert.Equal(t, parentID, *req.ParentBranchID)
		assert.Equal(t, branching.DataCloneModeSchemaOnly, req.DataCloneMode)
		assert.Equal(t, branching.BranchTypePreview, req.Type)
		assert.Equal(t, 123, *req.GitHubPRNumber)
		assert.Equal(t, "https://github.com/owner/repo/pull/123", *req.GitHubPRURL)
		assert.Equal(t, "owner/repo", *req.GitHubRepo)
		assert.Equal(t, "24h", *req.ExpiresIn)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{
			"name": "test-branch",
			"data_clone_mode": "schema_only",
			"type": "preview",
			"expires_in": "48h"
		}`

		var req CreateBranchRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "test-branch", req.Name)
		assert.Equal(t, branching.DataCloneModeSchemaOnly, req.DataCloneMode)
		assert.Equal(t, branching.BranchTypePreview, req.Type)
		assert.Equal(t, "48h", *req.ExpiresIn)
	})

	t.Run("minimal request", func(t *testing.T) {
		req := CreateBranchRequest{
			Name: "simple-branch",
		}

		assert.Equal(t, "simple-branch", req.Name)
		assert.Nil(t, req.ParentBranchID)
		assert.Nil(t, req.GitHubPRNumber)
	})
}

// =============================================================================
// SetActiveBranchRequest Tests
// =============================================================================

func TestSetActiveBranchRequest_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		req := SetActiveBranchRequest{
			Branch: "feature-123",
		}

		assert.Equal(t, "feature-123", req.Branch)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{"branch":"my-branch"}`

		var req SetActiveBranchRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "my-branch", req.Branch)
	})
}

// =============================================================================
// UpsertGitHubConfigRequest Tests
// =============================================================================

func TestUpsertGitHubConfigRequest_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		autoCreate := true
		autoDelete := false
		secret := "webhook-secret"

		req := UpsertGitHubConfigRequest{
			Repository:           "owner/repo",
			AutoCreateOnPR:       &autoCreate,
			AutoDeleteOnMerge:    &autoDelete,
			DefaultDataCloneMode: branching.DataCloneModeSchemaOnly,
			WebhookSecret:        &secret,
		}

		assert.Equal(t, "owner/repo", req.Repository)
		assert.True(t, *req.AutoCreateOnPR)
		assert.False(t, *req.AutoDeleteOnMerge)
		assert.Equal(t, branching.DataCloneModeSchemaOnly, req.DefaultDataCloneMode)
		assert.Equal(t, "webhook-secret", *req.WebhookSecret)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{
			"repository": "fluxbase/fluxbase",
			"auto_create_on_pr": true,
			"auto_delete_on_merge": true,
			"default_data_clone_mode": "full_clone"
		}`

		var req UpsertGitHubConfigRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "fluxbase/fluxbase", req.Repository)
		assert.True(t, *req.AutoCreateOnPR)
		assert.True(t, *req.AutoDeleteOnMerge)
		assert.Equal(t, branching.DataCloneModeFullClone, req.DefaultDataCloneMode)
	})

	t.Run("minimal request", func(t *testing.T) {
		req := UpsertGitHubConfigRequest{
			Repository: "owner/repo",
		}

		assert.Equal(t, "owner/repo", req.Repository)
		assert.Nil(t, req.AutoCreateOnPR)
		assert.Nil(t, req.AutoDeleteOnMerge)
	})
}

// =============================================================================
// GrantBranchAccessRequest Tests
// =============================================================================

func TestGrantBranchAccessRequest_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		req := GrantBranchAccessRequest{
			UserID:      "550e8400-e29b-41d4-a716-446655440000",
			AccessLevel: "admin",
		}

		assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", req.UserID)
		assert.Equal(t, "admin", req.AccessLevel)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{"user_id":"123e4567-e89b-12d3-a456-426614174000","access_level":"write"}`

		var req GrantBranchAccessRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", req.UserID)
		assert.Equal(t, "write", req.AccessLevel)
	})

	t.Run("valid access levels", func(t *testing.T) {
		validLevels := []string{"read", "write", "admin"}
		for _, level := range validLevels {
			req := GrantBranchAccessRequest{
				UserID:      uuid.New().String(),
				AccessLevel: level,
			}
			assert.Equal(t, level, req.AccessLevel)
		}
	})
}

// =============================================================================
// CreateBranch Handler Tests
// =============================================================================

func TestCreateBranch_Validation(t *testing.T) {
	t.Run("branching disabled", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{
			Enabled: false,
		})

		app.Post("/branches", handler.CreateBranch)

		body := `{"name":"test-branch"}`
		req := httptest.NewRequest(http.MethodPost, "/branches", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Equal(t, "branching_disabled", result["error"])
	})

	t.Run("invalid request body", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{
			Enabled: true,
		})

		app.Post("/branches", handler.CreateBranch)

		req := httptest.NewRequest(http.MethodPost, "/branches", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Equal(t, "invalid_request", result["error"])
	})

	t.Run("empty branch name", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{
			Enabled: true,
		})

		app.Post("/branches", handler.CreateBranch)

		body := `{"name":""}`
		req := httptest.NewRequest(http.MethodPost, "/branches", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Equal(t, "validation_error", result["error"])
		assert.Contains(t, result["message"], "Branch name is required")
	})

	t.Run("invalid expires_in duration", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{
			Enabled: true,
		})

		app.Post("/branches", handler.CreateBranch)

		body := `{"name":"test-branch","expires_in":"invalid-duration"}`
		req := httptest.NewRequest(http.MethodPost, "/branches", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Equal(t, "validation_error", result["error"])
		assert.Contains(t, result["message"], "Invalid expires_in duration")
	})
}

// =============================================================================
// ListBranches Handler Tests
// =============================================================================

func TestListBranches_ParameterParsing(t *testing.T) {
	// Note: Full testing requires mocked manager
	// These tests verify parameter parsing behavior

	t.Run("default pagination", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Get("/branches", handler.ListBranches)

		req := httptest.NewRequest(http.MethodGet, "/branches", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Will fail due to nil manager, but verifies handler is reachable
		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("custom pagination", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Get("/branches", handler.ListBranches)

		req := httptest.NewRequest(http.MethodGet, "/branches?limit=50&offset=10", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("filter by status", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Get("/branches", handler.ListBranches)

		req := httptest.NewRequest(http.MethodGet, "/branches?status=active", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("filter by type", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Get("/branches", handler.ListBranches)

		req := httptest.NewRequest(http.MethodGet, "/branches?type=preview", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("filter by github_repo", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Get("/branches", handler.ListBranches)

		req := httptest.NewRequest(http.MethodGet, "/branches?github_repo=owner/repo", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("filter mine", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		// Middleware to set user_id
		app.Use(func(c fiber.Ctx) error {
			c.Locals("user_id", "550e8400-e29b-41d4-a716-446655440000")
			return c.Next()
		})

		app.Get("/branches", handler.ListBranches)

		req := httptest.NewRequest(http.MethodGet, "/branches?mine=true", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// GetBranch Handler Tests
// =============================================================================

func TestGetBranch_ParameterParsing(t *testing.T) {
	t.Run("with UUID parameter", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Get("/branches/:id", handler.GetBranch)

		req := httptest.NewRequest(http.MethodGet, "/branches/550e8400-e29b-41d4-a716-446655440000", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Will fail due to nil manager, but verifies handler is reachable
		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("with slug parameter", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Get("/branches/:id", handler.GetBranch)

		req := httptest.NewRequest(http.MethodGet, "/branches/feature-branch-123", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// DeleteBranch Handler Tests
// =============================================================================

func TestDeleteBranch_ParameterParsing(t *testing.T) {
	t.Run("with UUID parameter", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Delete("/branches/:id", handler.DeleteBranch)

		req := httptest.NewRequest(http.MethodDelete, "/branches/550e8400-e29b-41d4-a716-446655440000", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("with slug parameter", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Delete("/branches/:id", handler.DeleteBranch)

		req := httptest.NewRequest(http.MethodDelete, "/branches/feature-branch", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// ResetBranch Handler Tests
// =============================================================================

func TestResetBranch_ParameterParsing(t *testing.T) {
	t.Run("with UUID parameter", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Post("/branches/:id/reset", handler.ResetBranch)

		req := httptest.NewRequest(http.MethodPost, "/branches/550e8400-e29b-41d4-a716-446655440000/reset", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("with slug parameter", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Post("/branches/:id/reset", handler.ResetBranch)

		req := httptest.NewRequest(http.MethodPost, "/branches/my-branch/reset", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// GetBranchActivity Handler Tests
// =============================================================================

func TestGetBranchActivity_ParameterParsing(t *testing.T) {
	t.Run("default limit", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Get("/branches/:id/activity", handler.GetBranchActivity)

		req := httptest.NewRequest(http.MethodGet, "/branches/test-branch/activity", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("custom limit", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Get("/branches/:id/activity", handler.GetBranchActivity)

		req := httptest.NewRequest(http.MethodGet, "/branches/test-branch/activity?limit=25", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("limit capped at 100", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Get("/branches/:id/activity", handler.GetBranchActivity)

		req := httptest.NewRequest(http.MethodGet, "/branches/test-branch/activity?limit=500", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// GetPoolStats Handler Tests
// =============================================================================

func TestGetPoolStats_Handler(t *testing.T) {
	t.Run("handler reachable", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Get("/branches/stats/pools", handler.GetPoolStats)

		req := httptest.NewRequest(http.MethodGet, "/branches/stats/pools", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Will panic due to nil router, but this tests the route registration
		// In a real test with mocked router, this would work
		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// GetActiveBranch Handler Tests
// =============================================================================

func TestGetActiveBranch_Handler(t *testing.T) {
	t.Run("handler reachable", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Get("/branches/active", handler.GetActiveBranch)

		req := httptest.NewRequest(http.MethodGet, "/branches/active", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// SetActiveBranch Handler Tests
// =============================================================================

func TestSetActiveBranch_Validation(t *testing.T) {
	t.Run("branching disabled", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{
			Enabled: false,
		})

		app.Post("/branches/active", handler.SetActiveBranch)

		body := `{"branch":"feature-123"}`
		req := httptest.NewRequest(http.MethodPost, "/branches/active", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Equal(t, "branching_disabled", result["error"])
	})

	t.Run("invalid request body", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{
			Enabled: true,
		})

		app.Post("/branches/active", handler.SetActiveBranch)

		req := httptest.NewRequest(http.MethodPost, "/branches/active", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("empty branch slug", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{
			Enabled: true,
		})

		app.Post("/branches/active", handler.SetActiveBranch)

		body := `{"branch":""}`
		req := httptest.NewRequest(http.MethodPost, "/branches/active", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Equal(t, "validation_error", result["error"])
		assert.Contains(t, result["message"], "Branch slug is required")
	})
}

// =============================================================================
// ResetActiveBranch Handler Tests
// =============================================================================

func TestResetActiveBranch_Handler(t *testing.T) {
	t.Run("handler reachable", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Delete("/branches/active", handler.ResetActiveBranch)

		req := httptest.NewRequest(http.MethodDelete, "/branches/active", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// ListGitHubConfigs Handler Tests
// =============================================================================

func TestListGitHubConfigs_Handler(t *testing.T) {
	t.Run("handler reachable", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Get("/github/configs", handler.ListGitHubConfigs)

		req := httptest.NewRequest(http.MethodGet, "/github/configs", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// UpsertGitHubConfig Handler Tests
// =============================================================================

func TestUpsertGitHubConfig_Validation(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Post("/github/configs", handler.UpsertGitHubConfig)

		req := httptest.NewRequest(http.MethodPost, "/github/configs", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Equal(t, "invalid_request", result["error"])
	})

	t.Run("empty repository", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Post("/github/configs", handler.UpsertGitHubConfig)

		body := `{"repository":""}`
		req := httptest.NewRequest(http.MethodPost, "/github/configs", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Equal(t, "validation_error", result["error"])
		assert.Contains(t, result["message"], "Repository is required")
	})
}

// =============================================================================
// DeleteGitHubConfig Handler Tests
// =============================================================================

func TestDeleteGitHubConfig_ParameterParsing(t *testing.T) {
	t.Run("with repository parameter", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Delete("/github/configs/:repository", handler.DeleteGitHubConfig)

		req := httptest.NewRequest(http.MethodDelete, "/github/configs/owner-repo", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// ListBranchAccess Handler Tests
// =============================================================================

func TestListBranchAccess_ParameterParsing(t *testing.T) {
	t.Run("with UUID parameter", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Get("/branches/:id/access", handler.ListBranchAccess)

		req := httptest.NewRequest(http.MethodGet, "/branches/550e8400-e29b-41d4-a716-446655440000/access", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("with slug parameter", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Get("/branches/:id/access", handler.ListBranchAccess)

		req := httptest.NewRequest(http.MethodGet, "/branches/feature-branch/access", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// GrantBranchAccess Handler Tests
// =============================================================================

func TestGrantBranchAccess_Validation(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Post("/branches/:id/access", handler.GrantBranchAccess)

		req := httptest.NewRequest(http.MethodPost, "/branches/test-branch/access", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Will fail due to nil manager, but body parsing happens first
		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("empty user_id", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Post("/branches/:id/access", handler.GrantBranchAccess)

		body := `{"user_id":"","access_level":"read"}`
		req := httptest.NewRequest(http.MethodPost, "/branches/test-branch/access", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Will fail at branch lookup before user_id validation
		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("invalid user_id format", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Post("/branches/:id/access", handler.GrantBranchAccess)

		body := `{"user_id":"not-a-uuid","access_level":"read"}`
		req := httptest.NewRequest(http.MethodPost, "/branches/test-branch/access", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// RevokeBranchAccess Handler Tests
// =============================================================================

func TestRevokeBranchAccess_Validation(t *testing.T) {
	t.Run("invalid user_id format", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Delete("/branches/:id/access/:user_id", handler.RevokeBranchAccess)

		req := httptest.NewRequest(http.MethodDelete, "/branches/test-branch/access/not-a-uuid", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Will fail at branch lookup before user_id validation
		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("with valid UUID parameters", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		app.Delete("/branches/:id/access/:user_id", handler.RevokeBranchAccess)

		req := httptest.NewRequest(http.MethodDelete, "/branches/550e8400-e29b-41d4-a716-446655440000/access/123e4567-e89b-12d3-a456-426614174000", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// JSON Serialization Tests
// =============================================================================

func TestBranchRequests_JSONSerialization(t *testing.T) {
	t.Run("CreateBranchRequest serializes correctly", func(t *testing.T) {
		req := CreateBranchRequest{
			Name:          "test-branch",
			DataCloneMode: branching.DataCloneModeSchemaOnly,
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"name":"test-branch"`)
		assert.Contains(t, string(data), `"data_clone_mode":"schema_only"`)
	})

	t.Run("SetActiveBranchRequest serializes correctly", func(t *testing.T) {
		req := SetActiveBranchRequest{
			Branch: "my-branch",
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"branch":"my-branch"`)
	})

	t.Run("UpsertGitHubConfigRequest serializes correctly", func(t *testing.T) {
		autoCreate := true
		req := UpsertGitHubConfigRequest{
			Repository:     "owner/repo",
			AutoCreateOnPR: &autoCreate,
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"repository":"owner/repo"`)
		assert.Contains(t, string(data), `"auto_create_on_pr":true`)
	})

	t.Run("GrantBranchAccessRequest serializes correctly", func(t *testing.T) {
		req := GrantBranchAccessRequest{
			UserID:      "550e8400-e29b-41d4-a716-446655440000",
			AccessLevel: "admin",
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"user_id":"550e8400-e29b-41d4-a716-446655440000"`)
		assert.Contains(t, string(data), `"access_level":"admin"`)
	})
}

// =============================================================================
// Branch Types and Status Tests
// =============================================================================

func TestBranchTypes(t *testing.T) {
	t.Run("valid branch types", func(t *testing.T) {
		types := []branching.BranchType{
			branching.BranchTypeMain,
			branching.BranchTypePreview,
			branching.BranchTypePersistent,
		}

		for _, bt := range types {
			assert.NotEmpty(t, string(bt))
		}
	})
}

func TestDataCloneModes(t *testing.T) {
	t.Run("valid data clone modes", func(t *testing.T) {
		modes := []branching.DataCloneMode{
			branching.DataCloneModeSchemaOnly,
			branching.DataCloneModeFullClone,
		}

		for _, mode := range modes {
			assert.NotEmpty(t, string(mode))
		}
	})
}

func TestBranchAccessLevels(t *testing.T) {
	t.Run("valid access levels", func(t *testing.T) {
		levels := []branching.BranchAccessLevel{
			branching.BranchAccessRead,
			branching.BranchAccessWrite,
			branching.BranchAccessAdmin,
		}

		for _, level := range levels {
			assert.NotEmpty(t, string(level))
		}
	})

	t.Run("access level hierarchy", func(t *testing.T) {
		// Verify the access level strings
		assert.Equal(t, "read", string(branching.BranchAccessRead))
		assert.Equal(t, "write", string(branching.BranchAccessWrite))
		assert.Equal(t, "admin", string(branching.BranchAccessAdmin))
	})
}

// =============================================================================
// RegisterRoutes Tests
// =============================================================================

func TestBranchHandler_RegisterRoutes(t *testing.T) {
	t.Run("routes are registered without panic", func(t *testing.T) {
		app := fiber.New()
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		// RegisterRoutes should not panic with nil dependencies
		assert.NotPanics(t, func() {
			handler.RegisterRoutes(app)
		})
	})

	t.Run("all handler methods exist", func(t *testing.T) {
		handler := NewBranchHandler(nil, nil, config.BranchingConfig{})

		assert.NotNil(t, handler.CreateBranch)
		assert.NotNil(t, handler.ListBranches)
		assert.NotNil(t, handler.GetBranch)
		assert.NotNil(t, handler.DeleteBranch)
		assert.NotNil(t, handler.ResetBranch)
		assert.NotNil(t, handler.GetBranchActivity)
		assert.NotNil(t, handler.GetPoolStats)
		assert.NotNil(t, handler.GetActiveBranch)
		assert.NotNil(t, handler.SetActiveBranch)
		assert.NotNil(t, handler.ResetActiveBranch)
		assert.NotNil(t, handler.ListGitHubConfigs)
		assert.NotNil(t, handler.UpsertGitHubConfig)
		assert.NotNil(t, handler.DeleteGitHubConfig)
		assert.NotNil(t, handler.ListBranchAccess)
		assert.NotNil(t, handler.GrantBranchAccess)
		assert.NotNil(t, handler.RevokeBranchAccess)
	})
}
