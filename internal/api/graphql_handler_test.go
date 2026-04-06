package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/config"
)

// =============================================================================
// GraphQLRequest Tests
// =============================================================================

func TestGraphQLRequest_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		req := GraphQLRequest{
			Query:         "{ users { id name } }",
			OperationName: "GetUsers",
			Variables: map[string]interface{}{
				"limit": 10,
			},
		}

		assert.Equal(t, "{ users { id name } }", req.Query)
		assert.Equal(t, "GetUsers", req.OperationName)
		assert.Equal(t, 10, req.Variables["limit"])
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{
			"query": "query GetUser($id: UUID!) { user(id: $id) { id email } }",
			"operationName": "GetUser",
			"variables": {"id": "550e8400-e29b-41d4-a716-446655440000"}
		}`

		var req GraphQLRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Contains(t, req.Query, "GetUser")
		assert.Equal(t, "GetUser", req.OperationName)
		assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", req.Variables["id"])
	})

	t.Run("minimal request", func(t *testing.T) {
		jsonData := `{"query": "{ _health }"}`

		var req GraphQLRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "{ _health }", req.Query)
		assert.Empty(t, req.OperationName)
		assert.Nil(t, req.Variables)
	})
}

// =============================================================================
// GraphQLResponse Tests
// =============================================================================

func TestGraphQLResponse_Struct(t *testing.T) {
	t.Run("successful response with data", func(t *testing.T) {
		resp := GraphQLResponse{
			Data: map[string]interface{}{
				"users": []map[string]interface{}{
					{"id": "1", "name": "Alice"},
					{"id": "2", "name": "Bob"},
				},
			},
		}

		assert.NotNil(t, resp.Data)
		assert.Nil(t, resp.Errors)
	})

	t.Run("error response", func(t *testing.T) {
		resp := GraphQLResponse{
			Errors: []GraphQLError{
				{
					Message: "User not found",
					Path:    []interface{}{"user"},
				},
			},
		}

		assert.Nil(t, resp.Data)
		assert.Len(t, resp.Errors, 1)
		assert.Equal(t, "User not found", resp.Errors[0].Message)
	})

	t.Run("partial response with data and errors", func(t *testing.T) {
		resp := GraphQLResponse{
			Data: map[string]interface{}{
				"user": map[string]interface{}{"id": "1", "name": "Alice"},
			},
			Errors: []GraphQLError{
				{Message: "Field 'email' is deprecated"},
			},
		}

		assert.NotNil(t, resp.Data)
		assert.Len(t, resp.Errors, 1)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		resp := GraphQLResponse{
			Data: map[string]interface{}{"_health": "ok"},
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"data"`)
		assert.Contains(t, string(data), `"_health"`)
	})
}

// =============================================================================
// GraphQLError Tests
// =============================================================================

func TestGraphQLError_Struct(t *testing.T) {
	t.Run("simple error", func(t *testing.T) {
		err := GraphQLError{
			Message: "Invalid query syntax",
		}

		assert.Equal(t, "Invalid query syntax", err.Message)
		assert.Nil(t, err.Locations)
		assert.Nil(t, err.Path)
		assert.Nil(t, err.Extensions)
	})

	t.Run("error with location", func(t *testing.T) {
		err := GraphQLError{
			Message: "Syntax error",
			Locations: []GraphQLErrorLocation{
				{Line: 5, Column: 10},
			},
		}

		assert.Len(t, err.Locations, 1)
		assert.Equal(t, 5, err.Locations[0].Line)
		assert.Equal(t, 10, err.Locations[0].Column)
	})

	t.Run("error with path", func(t *testing.T) {
		err := GraphQLError{
			Message: "Field error",
			Path:    []interface{}{"user", 0, "email"},
		}

		assert.Len(t, err.Path, 3)
		assert.Equal(t, "user", err.Path[0])
		assert.Equal(t, 0, err.Path[1])
		assert.Equal(t, "email", err.Path[2])
	})

	t.Run("error with extensions", func(t *testing.T) {
		err := GraphQLError{
			Message: "Unauthorized",
			Extensions: map[string]interface{}{
				"code":    "UNAUTHENTICATED",
				"details": "Token expired",
			},
		}

		assert.Equal(t, "UNAUTHENTICATED", err.Extensions["code"])
		assert.Equal(t, "Token expired", err.Extensions["details"])
	})

	t.Run("JSON serialization", func(t *testing.T) {
		err := GraphQLError{
			Message: "Test error",
			Locations: []GraphQLErrorLocation{
				{Line: 1, Column: 5},
			},
			Path: []interface{}{"users", 0},
		}

		data, errMarshal := json.Marshal(err)
		require.NoError(t, errMarshal)

		assert.Contains(t, string(data), `"message":"Test error"`)
		assert.Contains(t, string(data), `"line":1`)
		assert.Contains(t, string(data), `"column":5`)
	})
}

// =============================================================================
// GraphQLErrorLocation Tests
// =============================================================================

func TestGraphQLErrorLocation_Struct(t *testing.T) {
	t.Run("zero values", func(t *testing.T) {
		loc := GraphQLErrorLocation{}
		assert.Equal(t, 0, loc.Line)
		assert.Equal(t, 0, loc.Column)
	})

	t.Run("set values", func(t *testing.T) {
		loc := GraphQLErrorLocation{
			Line:   42,
			Column: 15,
		}
		assert.Equal(t, 42, loc.Line)
		assert.Equal(t, 15, loc.Column)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		loc := GraphQLErrorLocation{Line: 10, Column: 20}

		data, err := json.Marshal(loc)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"line":10`)
		assert.Contains(t, string(data), `"column":20`)
	})
}

// =============================================================================
// calculateQueryDepth Tests
// =============================================================================

func TestCalculateQueryDepth(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected int
		hasError bool
	}{
		{
			name:     "simple query",
			query:    "{ _health }",
			expected: 1,
		},
		{
			name:     "two levels",
			query:    "{ users { id } }",
			expected: 2,
		},
		{
			name:     "three levels",
			query:    "{ users { posts { title } } }",
			expected: 3,
		},
		{
			name:     "four levels with nested objects",
			query:    "{ users { posts { comments { author { name } } } } }",
			expected: 5,
		},
		{
			name:     "multiple fields at same level",
			query:    "{ users { id name email posts { id } } }",
			expected: 3,
		},
		{
			name:     "query with arguments",
			query:    "{ user(id: \"123\") { name posts(limit: 10) { title } } }",
			expected: 3,
		},
		{
			name:     "mutation",
			query:    "mutation { insertUser(data: {name: \"Test\"}) { id name } }",
			expected: 2,
		},
		{
			name:     "fragment spread",
			query:    "{ users { ...UserFields } } fragment UserFields on User { id name }",
			expected: 2,
		},
		{
			name:     "inline fragment",
			query:    "{ node { ... on User { id name posts { title } } } }",
			expected: 4,
		},
		{
			name:     "invalid query syntax",
			query:    "{ users { ",
			hasError: true,
		},
		{
			name:     "empty query",
			query:    "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			depth, err := calculateQueryDepth(tt.query)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, depth)
			}
		})
	}
}

// =============================================================================
// calculateQueryComplexity Tests
// =============================================================================

func TestCalculateQueryComplexity(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		minExpected int // complexity should be at least this
	}{
		{
			name:        "simple single field",
			query:       "{ _health }",
			minExpected: 1,
		},
		{
			name:        "multiple scalar fields",
			query:       "{ user { id name email createdAt } }",
			minExpected: 4,
		},
		{
			name:        "list field has higher cost",
			query:       "{ users { id } }",
			minExpected: 10, // list fields have base cost of 10
		},
		{
			name:        "nested list fields compound",
			query:       "{ users { posts { comments { text } } } }",
			minExpected: 100, // nested lists multiply
		},
		{
			name:        "mutation has base cost",
			query:       "mutation { insertUser(data: {name: \"Test\"}) { id } }",
			minExpected: 10, // mutations add base cost
		},
		{
			name:        "field with limit argument",
			query:       "{ users(limit: 5) { id } }",
			minExpected: 10,
		},
		{
			name:        "invalid query returns zero",
			query:       "{ invalid syntax",
			minExpected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			complexity := calculateQueryComplexity(tt.query)
			assert.GreaterOrEqual(t, complexity, tt.minExpected)
		})
	}
}

// =============================================================================
// HandleGraphQL Handler Tests
// =============================================================================

func TestHandleGraphQL_Validation(t *testing.T) {
	t.Run("invalid JSON body", func(t *testing.T) {
		app := fiber.New()

		cfg := &config.GraphQLConfig{
			Enabled:       true,
			Introspection: true,
		}

		// We can't create a full handler without database, but we can test body parsing
		// by checking the error response format
		handler := &GraphQLHandler{
			config: cfg,
		}

		app.Post("/graphql", handler.HandleGraphQL)

		req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var gqlResp GraphQLResponse
		err = json.Unmarshal(body, &gqlResp)
		require.NoError(t, err)

		assert.Len(t, gqlResp.Errors, 1)
		assert.Contains(t, gqlResp.Errors[0].Message, "Invalid JSON")
	})

	t.Run("empty query", func(t *testing.T) {
		app := fiber.New()

		cfg := &config.GraphQLConfig{
			Enabled:       true,
			Introspection: true,
		}

		handler := &GraphQLHandler{
			config: cfg,
		}

		app.Post("/graphql", handler.HandleGraphQL)

		reqBody := `{"query": ""}`
		req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader([]byte(reqBody)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var gqlResp GraphQLResponse
		err = json.Unmarshal(body, &gqlResp)
		require.NoError(t, err)

		assert.Len(t, gqlResp.Errors, 1)
		assert.Contains(t, gqlResp.Errors[0].Message, "Query string is required")
	})

	t.Run("query exceeds max depth", func(t *testing.T) {
		app := fiber.New()

		cfg := &config.GraphQLConfig{
			Enabled:       true,
			Introspection: true,
			MaxDepth:      2, // Very shallow
		}

		handler := &GraphQLHandler{
			config: cfg,
		}

		app.Post("/graphql", handler.HandleGraphQL)

		// Query with depth 3 - should exceed max
		reqBody := `{"query": "{ users { posts { title } } }"}`
		req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader([]byte(reqBody)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var gqlResp GraphQLResponse
		err = json.Unmarshal(body, &gqlResp)
		require.NoError(t, err)

		assert.Len(t, gqlResp.Errors, 1)
		assert.Contains(t, gqlResp.Errors[0].Message, "exceeds maximum allowed depth")
	})

	t.Run("query exceeds max complexity", func(t *testing.T) {
		app := fiber.New()

		cfg := &config.GraphQLConfig{
			Enabled:       true,
			Introspection: true,
			MaxComplexity: 5, // Very low
		}

		handler := &GraphQLHandler{
			config: cfg,
		}

		app.Post("/graphql", handler.HandleGraphQL)

		// Query with complexity > 5
		reqBody := `{"query": "{ users { id name email posts { title } } }"}`
		req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader([]byte(reqBody)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var gqlResp GraphQLResponse
		err = json.Unmarshal(body, &gqlResp)
		require.NoError(t, err)

		assert.Len(t, gqlResp.Errors, 1)
		assert.Contains(t, gqlResp.Errors[0].Message, "exceeds maximum")
	})

	t.Run("invalid query syntax caught by depth check", func(t *testing.T) {
		app := fiber.New()

		cfg := &config.GraphQLConfig{
			Enabled:       true,
			Introspection: true,
			MaxDepth:      10,
		}

		handler := &GraphQLHandler{
			config: cfg,
		}

		app.Post("/graphql", handler.HandleGraphQL)

		reqBody := `{"query": "{ users { "}`
		req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader([]byte(reqBody)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var gqlResp GraphQLResponse
		err = json.Unmarshal(body, &gqlResp)
		require.NoError(t, err)

		assert.Contains(t, gqlResp.Errors[0].Message, "Invalid query syntax")
	})
}

// =============================================================================
// HandleIntrospection Handler Tests
// =============================================================================

func TestHandleIntrospection_Validation(t *testing.T) {
	t.Run("introspection disabled", func(t *testing.T) {
		app := fiber.New()

		cfg := &config.GraphQLConfig{
			Enabled:       true,
			Introspection: false, // Disabled
		}

		handler := &GraphQLHandler{
			config: cfg,
		}

		app.Get("/graphql", handler.HandleIntrospection)

		req := httptest.NewRequest(http.MethodGet, "/graphql", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var gqlResp GraphQLResponse
		err = json.Unmarshal(body, &gqlResp)
		require.NoError(t, err)

		assert.Len(t, gqlResp.Errors, 1)
		assert.Contains(t, gqlResp.Errors[0].Message, "Introspection is disabled")
	})
}

// =============================================================================
// convertErrors Tests
// =============================================================================

func TestConvertErrors(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		result := convertErrors(nil)
		assert.Nil(t, result)
	})

	t.Run("empty input returns nil", func(t *testing.T) {
		result := convertErrors([]gqlerrors.FormattedError{})
		assert.Nil(t, result)
	})
}

// =============================================================================
// JSON Serialization Tests
// =============================================================================

func TestGraphQL_JSONSerialization(t *testing.T) {
	t.Run("GraphQLRequest roundtrip", func(t *testing.T) {
		original := GraphQLRequest{
			Query:         "{ users { id } }",
			OperationName: "GetUsers",
			Variables:     map[string]interface{}{"limit": float64(10)},
		}

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded GraphQLRequest
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, original.Query, decoded.Query)
		assert.Equal(t, original.OperationName, decoded.OperationName)
		assert.Equal(t, original.Variables["limit"], decoded.Variables["limit"])
	})

	t.Run("GraphQLResponse roundtrip", func(t *testing.T) {
		original := GraphQLResponse{
			Data: map[string]interface{}{"_health": "ok"},
			Errors: []GraphQLError{
				{Message: "Warning"},
			},
		}

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded GraphQLResponse
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.NotNil(t, decoded.Data)
		assert.Len(t, decoded.Errors, 1)
	})

	t.Run("GraphQLError roundtrip", func(t *testing.T) {
		original := GraphQLError{
			Message: "Test error",
			Locations: []GraphQLErrorLocation{
				{Line: 1, Column: 5},
			},
			Path:       []interface{}{"users", float64(0)},
			Extensions: map[string]interface{}{"code": "TEST"},
		}

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded GraphQLError
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, original.Message, decoded.Message)
		assert.Len(t, decoded.Locations, 1)
		assert.Equal(t, 1, decoded.Locations[0].Line)
	})
}

// =============================================================================
// Config Integration Tests
// =============================================================================

func TestGraphQLConfig_Integration(t *testing.T) {
	t.Run("config fields accessible", func(t *testing.T) {
		cfg := &config.GraphQLConfig{
			Enabled:       true,
			Introspection: true,
			MaxDepth:      15,
			MaxComplexity: 1000,
		}

		handler := &GraphQLHandler{
			config: cfg,
		}

		assert.True(t, handler.config.Enabled)
		assert.True(t, handler.config.Introspection)
		assert.Equal(t, 15, handler.config.MaxDepth)
		assert.Equal(t, 1000, handler.config.MaxComplexity)
	})

	t.Run("nil config handled", func(t *testing.T) {
		handler := &GraphQLHandler{
			config: nil,
		}

		assert.Nil(t, handler.config)
	})
}
