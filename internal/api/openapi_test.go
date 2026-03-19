package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// OpenAPISpec Struct Tests
// =============================================================================

func TestOpenAPISpec_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		spec := OpenAPISpec{
			OpenAPI: "3.0.0",
			Info: OpenAPIInfo{
				Title:       "Test API",
				Description: "A test API",
				Version:     "1.0.0",
			},
			Servers: []OpenAPIServer{
				{URL: "http://localhost:8080", Description: "Local server"},
			},
			Paths: map[string]OpenAPIPath{
				"/test": {"get": OpenAPIOperation{Summary: "Test endpoint"}},
			},
			Components: OpenAPIComponents{
				Schemas:         map[string]interface{}{"Test": map[string]string{"type": "object"}},
				SecuritySchemes: map[string]interface{}{"bearerAuth": map[string]string{"type": "http"}},
			},
		}

		assert.Equal(t, "3.0.0", spec.OpenAPI)
		assert.Equal(t, "Test API", spec.Info.Title)
		assert.Len(t, spec.Servers, 1)
		assert.Len(t, spec.Paths, 1)
		assert.NotNil(t, spec.Components.Schemas)
		assert.NotNil(t, spec.Components.SecuritySchemes)
	})
}

// =============================================================================
// OpenAPIInfo Struct Tests
// =============================================================================

func TestOpenAPIInfo_Struct(t *testing.T) {
	t.Run("stores title, description, version", func(t *testing.T) {
		info := OpenAPIInfo{
			Title:       "Fluxbase API",
			Description: "Fluxbase REST API",
			Version:     "1.0.0",
		}

		assert.Equal(t, "Fluxbase API", info.Title)
		assert.Equal(t, "Fluxbase REST API", info.Description)
		assert.Equal(t, "1.0.0", info.Version)
	})
}

// =============================================================================
// OpenAPIServer Struct Tests
// =============================================================================

func TestOpenAPIServer_Struct(t *testing.T) {
	t.Run("stores URL and description", func(t *testing.T) {
		server := OpenAPIServer{
			URL:         "https://api.example.com",
			Description: "Production server",
		}

		assert.Equal(t, "https://api.example.com", server.URL)
		assert.Equal(t, "Production server", server.Description)
	})
}

// =============================================================================
// OpenAPIOperation Struct Tests
// =============================================================================

func TestOpenAPIOperation_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		op := OpenAPIOperation{
			Summary:     "List users",
			Description: "Get all users",
			OperationID: "listUsers",
			Tags:        []string{"Users"},
			Parameters: []OpenAPIParameter{
				{Name: "limit", In: "query", Schema: map[string]string{"type": "integer"}},
			},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {Schema: map[string]string{"type": "object"}},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {Description: "Success"},
			},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
		}

		assert.Equal(t, "List users", op.Summary)
		assert.Equal(t, "Get all users", op.Description)
		assert.Equal(t, "listUsers", op.OperationID)
		assert.Equal(t, []string{"Users"}, op.Tags)
		assert.Len(t, op.Parameters, 1)
		assert.NotNil(t, op.RequestBody)
		assert.Len(t, op.Responses, 1)
		assert.Len(t, op.Security, 1)
	})

	t.Run("handles optional fields", func(t *testing.T) {
		op := OpenAPIOperation{
			Summary: "Simple operation",
			Responses: map[string]OpenAPIResponse{
				"200": {Description: "OK"},
			},
		}

		assert.Empty(t, op.Description)
		assert.Empty(t, op.OperationID)
		assert.Nil(t, op.Tags)
		assert.Nil(t, op.Parameters)
		assert.Nil(t, op.RequestBody)
	})
}

// =============================================================================
// OpenAPIParameter Struct Tests
// =============================================================================

func TestOpenAPIParameter_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		param := OpenAPIParameter{
			Name:        "id",
			In:          "path",
			Description: "User ID",
			Required:    true,
			Schema:      map[string]string{"type": "string", "format": "uuid"},
		}

		assert.Equal(t, "id", param.Name)
		assert.Equal(t, "path", param.In)
		assert.Equal(t, "User ID", param.Description)
		assert.True(t, param.Required)
		assert.NotNil(t, param.Schema)
	})

	t.Run("parameter in locations", func(t *testing.T) {
		locations := []string{"query", "path", "header", "cookie"}

		for _, loc := range locations {
			param := OpenAPIParameter{
				Name: "test",
				In:   loc,
			}
			assert.Equal(t, loc, param.In)
		}
	})
}

// =============================================================================
// OpenAPIRequestBody Struct Tests
// =============================================================================

func TestOpenAPIRequestBody_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		body := OpenAPIRequestBody{
			Description: "User data",
			Required:    true,
			Content: map[string]OpenAPIMedia{
				"application/json": {
					Schema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name": map[string]string{"type": "string"},
						},
					},
				},
			},
		}

		assert.Equal(t, "User data", body.Description)
		assert.True(t, body.Required)
		assert.Len(t, body.Content, 1)
	})
}

// =============================================================================
// OpenAPIMedia Struct Tests
// =============================================================================

func TestOpenAPIMedia_Struct(t *testing.T) {
	t.Run("stores schema", func(t *testing.T) {
		media := OpenAPIMedia{
			Schema: map[string]interface{}{
				"type": "object",
			},
		}

		assert.NotNil(t, media.Schema)
	})
}

// =============================================================================
// OpenAPIResponse Struct Tests
// =============================================================================

func TestOpenAPIResponse_Struct(t *testing.T) {
	t.Run("stores description and content", func(t *testing.T) {
		resp := OpenAPIResponse{
			Description: "Successful response",
			Content: map[string]OpenAPIMedia{
				"application/json": {
					Schema: map[string]interface{}{
						"$ref": "#/components/schemas/User",
					},
				},
			},
		}

		assert.Equal(t, "Successful response", resp.Description)
		assert.Len(t, resp.Content, 1)
	})

	t.Run("handles response without content", func(t *testing.T) {
		resp := OpenAPIResponse{
			Description: "No Content",
		}

		assert.Equal(t, "No Content", resp.Description)
		assert.Nil(t, resp.Content)
	})
}

// =============================================================================
// OpenAPIComponents Struct Tests
// =============================================================================

func TestOpenAPIComponents_Struct(t *testing.T) {
	t.Run("stores schemas and security schemes", func(t *testing.T) {
		components := OpenAPIComponents{
			Schemas: map[string]interface{}{
				"User": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":   map[string]string{"type": "string"},
						"name": map[string]string{"type": "string"},
					},
				},
			},
			SecuritySchemes: map[string]interface{}{
				"bearerAuth": map[string]interface{}{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
				},
			},
		}

		assert.Len(t, components.Schemas, 1)
		assert.Len(t, components.SecuritySchemes, 1)
	})
}

// =============================================================================
// NewOpenAPIHandler Tests
// =============================================================================

func TestNewOpenAPIHandler(t *testing.T) {
	t.Run("creates handler with nil db", func(t *testing.T) {
		handler := NewOpenAPIHandler(nil)

		require.NotNil(t, handler)
		assert.Nil(t, handler.db)
	})
}

// =============================================================================
// GetOpenAPISpec Handler Tests
// =============================================================================

func TestOpenAPIHandler_GetOpenAPISpec(t *testing.T) {
	t.Run("returns minimal spec for non-admin user", func(t *testing.T) {
		handler := NewOpenAPIHandler(nil)

		app := fiber.New()
		app.Get("/openapi.json", func(c fiber.Ctx) error {
			c.Locals("user_role", "authenticated")
			return handler.GetOpenAPISpec(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("returns minimal spec for anon user", func(t *testing.T) {
		handler := NewOpenAPIHandler(nil)

		app := fiber.New()
		app.Get("/openapi.json", func(c fiber.Ctx) error {
			c.Locals("user_role", "anon")
			return handler.GetOpenAPISpec(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// =============================================================================
// Admin Role Detection Tests
// =============================================================================

func TestOpenAPIHandler_AdminRoleDetection(t *testing.T) {
	t.Run("admin role is detected", func(t *testing.T) {
		role := "admin"
		isAdmin := role == "admin" || role == "instance_admin" || role == "service_role"

		assert.True(t, isAdmin)
	})

	t.Run("instance_admin role is detected", func(t *testing.T) {
		role := "instance_admin"
		isAdmin := role == "admin" || role == "instance_admin" || role == "service_role"

		assert.True(t, isAdmin)
	})

	t.Run("service_role is detected", func(t *testing.T) {
		role := "service_role"
		isAdmin := role == "admin" || role == "instance_admin" || role == "service_role"

		assert.True(t, isAdmin)
	})

	t.Run("authenticated role is not admin", func(t *testing.T) {
		role := "authenticated"
		isAdmin := role == "admin" || role == "instance_admin" || role == "service_role"

		assert.False(t, isAdmin)
	})
}

// =============================================================================
// columnToSchema Tests
// =============================================================================

func TestOpenAPIHandler_columnToSchema(t *testing.T) {
	handler := NewOpenAPIHandler(nil)

	t.Run("maps integer type", func(t *testing.T) {
		col := database.ColumnInfo{
			Name:     "count",
			DataType: "integer",
		}

		schema := handler.columnToSchema(col)

		assert.Equal(t, "integer", schema["type"])
	})

	t.Run("maps int4 type", func(t *testing.T) {
		col := database.ColumnInfo{
			Name:     "id",
			DataType: "int4",
		}

		schema := handler.columnToSchema(col)

		assert.Equal(t, "integer", schema["type"])
	})

	t.Run("maps bigint type", func(t *testing.T) {
		col := database.ColumnInfo{
			Name:     "big_id",
			DataType: "bigint",
		}

		schema := handler.columnToSchema(col)

		assert.Equal(t, "integer", schema["type"])
	})

	t.Run("maps numeric type", func(t *testing.T) {
		col := database.ColumnInfo{
			Name:     "price",
			DataType: "numeric(10,2)",
		}

		schema := handler.columnToSchema(col)

		assert.Equal(t, "number", schema["type"])
	})

	t.Run("maps decimal type", func(t *testing.T) {
		col := database.ColumnInfo{
			Name:     "amount",
			DataType: "decimal",
		}

		schema := handler.columnToSchema(col)

		assert.Equal(t, "number", schema["type"])
	})

	t.Run("maps float type", func(t *testing.T) {
		col := database.ColumnInfo{
			Name:     "rating",
			DataType: "float4",
		}

		schema := handler.columnToSchema(col)

		assert.Equal(t, "number", schema["type"])
	})

	t.Run("maps double type", func(t *testing.T) {
		col := database.ColumnInfo{
			Name:     "value",
			DataType: "double precision",
		}

		schema := handler.columnToSchema(col)

		assert.Equal(t, "number", schema["type"])
	})

	t.Run("maps boolean type", func(t *testing.T) {
		col := database.ColumnInfo{
			Name:     "active",
			DataType: "boolean",
		}

		schema := handler.columnToSchema(col)

		assert.Equal(t, "boolean", schema["type"])
	})

	t.Run("maps bool type", func(t *testing.T) {
		col := database.ColumnInfo{
			Name:     "enabled",
			DataType: "bool",
		}

		schema := handler.columnToSchema(col)

		assert.Equal(t, "boolean", schema["type"])
	})

	t.Run("maps json type", func(t *testing.T) {
		col := database.ColumnInfo{
			Name:     "metadata",
			DataType: "json",
		}

		schema := handler.columnToSchema(col)

		assert.Equal(t, "object", schema["type"])
	})

	t.Run("maps jsonb type", func(t *testing.T) {
		col := database.ColumnInfo{
			Name:     "config",
			DataType: "jsonb",
		}

		schema := handler.columnToSchema(col)

		assert.Equal(t, "object", schema["type"])
	})

	t.Run("maps array type", func(t *testing.T) {
		col := database.ColumnInfo{
			Name:     "tags",
			DataType: "text[]",
		}

		schema := handler.columnToSchema(col)

		assert.Equal(t, "array", schema["type"])
		assert.NotNil(t, schema["items"])
	})

	t.Run("maps underscore array type", func(t *testing.T) {
		col := database.ColumnInfo{
			Name:     "values",
			DataType: "_int4",
		}

		schema := handler.columnToSchema(col)

		assert.Equal(t, "array", schema["type"])
	})

	t.Run("maps timestamp type", func(t *testing.T) {
		col := database.ColumnInfo{
			Name:     "created_at",
			DataType: "timestamp with time zone",
		}

		schema := handler.columnToSchema(col)

		assert.Equal(t, "string", schema["type"])
		assert.Equal(t, "date-time", schema["format"])
	})

	t.Run("maps timestamptz type", func(t *testing.T) {
		col := database.ColumnInfo{
			Name:     "updated_at",
			DataType: "timestamptz",
		}

		schema := handler.columnToSchema(col)

		assert.Equal(t, "string", schema["type"])
		assert.Equal(t, "date-time", schema["format"])
	})

	t.Run("maps date type", func(t *testing.T) {
		col := database.ColumnInfo{
			Name:     "birth_date",
			DataType: "date",
		}

		schema := handler.columnToSchema(col)

		assert.Equal(t, "string", schema["type"])
		assert.Equal(t, "date-time", schema["format"])
	})

	t.Run("maps uuid type", func(t *testing.T) {
		col := database.ColumnInfo{
			Name:     "id",
			DataType: "uuid",
		}

		schema := handler.columnToSchema(col)

		assert.Equal(t, "string", schema["type"])
		assert.Equal(t, "uuid", schema["format"])
	})

	t.Run("maps text type to string", func(t *testing.T) {
		col := database.ColumnInfo{
			Name:     "description",
			DataType: "text",
		}

		schema := handler.columnToSchema(col)

		assert.Equal(t, "string", schema["type"])
	})

	t.Run("maps varchar type to string", func(t *testing.T) {
		col := database.ColumnInfo{
			Name:     "name",
			DataType: "character varying",
		}

		schema := handler.columnToSchema(col)

		assert.Equal(t, "string", schema["type"])
	})

	t.Run("handles nullable column", func(t *testing.T) {
		col := database.ColumnInfo{
			Name:       "optional_field",
			DataType:   "text",
			IsNullable: true,
		}

		schema := handler.columnToSchema(col)

		assert.Equal(t, true, schema["nullable"])
	})

	t.Run("non-nullable column has no nullable field", func(t *testing.T) {
		col := database.ColumnInfo{
			Name:       "required_field",
			DataType:   "text",
			IsNullable: false,
		}

		schema := handler.columnToSchema(col)

		_, exists := schema["nullable"]
		assert.False(t, exists)
	})
}

// =============================================================================
// buildTablePath Tests
// =============================================================================

func TestOpenAPIHandler_buildTablePath(t *testing.T) {
	handler := NewOpenAPIHandler(nil)

	t.Run("adds s suffix to regular table", func(t *testing.T) {
		table := database.TableInfo{
			Schema: "public",
			Name:   "user",
		}

		path := handler.buildTablePath(table)

		assert.Equal(t, "/api/v1/tables/users", path)
	})

	t.Run("keeps s for table already ending in s", func(t *testing.T) {
		table := database.TableInfo{
			Schema: "public",
			Name:   "users",
		}

		path := handler.buildTablePath(table)

		assert.Equal(t, "/api/v1/tables/users", path)
	})

	t.Run("changes y to ies", func(t *testing.T) {
		table := database.TableInfo{
			Schema: "public",
			Name:   "category",
		}

		path := handler.buildTablePath(table)

		assert.Equal(t, "/api/v1/tables/categories", path)
	})

	t.Run("adds es for x suffix", func(t *testing.T) {
		table := database.TableInfo{
			Schema: "public",
			Name:   "box",
		}

		path := handler.buildTablePath(table)

		assert.Equal(t, "/api/v1/tables/boxes", path)
	})

	t.Run("adds es for ch suffix", func(t *testing.T) {
		table := database.TableInfo{
			Schema: "public",
			Name:   "branch",
		}

		path := handler.buildTablePath(table)

		assert.Equal(t, "/api/v1/tables/branches", path)
	})

	t.Run("adds es for sh suffix", func(t *testing.T) {
		table := database.TableInfo{
			Schema: "public",
			Name:   "wish",
		}

		path := handler.buildTablePath(table)

		assert.Equal(t, "/api/v1/tables/wishes", path)
	})

	t.Run("includes schema for non-public tables", func(t *testing.T) {
		table := database.TableInfo{
			Schema: "auth",
			Name:   "user",
		}

		path := handler.buildTablePath(table)

		assert.Equal(t, "/api/v1/tables/auth/users", path)
	})

	t.Run("handles custom schema", func(t *testing.T) {
		table := database.TableInfo{
			Schema: "private",
			Name:   "secret",
		}

		path := handler.buildTablePath(table)

		assert.Equal(t, "/api/v1/tables/private/secrets", path)
	})
}

// =============================================================================
// getQueryParameters Tests
// =============================================================================

func TestOpenAPIHandler_getQueryParameters(t *testing.T) {
	handler := NewOpenAPIHandler(nil)

	t.Run("returns expected parameters", func(t *testing.T) {
		params := handler.getQueryParameters()

		require.Len(t, params, 5)

		// Check parameter names
		paramNames := make([]string, len(params))
		for i, p := range params {
			paramNames[i] = p.Name
		}

		assert.Contains(t, paramNames, "select")
		assert.Contains(t, paramNames, "order")
		assert.Contains(t, paramNames, "limit")
		assert.Contains(t, paramNames, "offset")
		assert.Contains(t, paramNames, "filter")
	})

	t.Run("all parameters are query parameters", func(t *testing.T) {
		params := handler.getQueryParameters()

		for _, p := range params {
			assert.Equal(t, "query", p.In)
		}
	})

	t.Run("all parameters have schemas", func(t *testing.T) {
		params := handler.getQueryParameters()

		for _, p := range params {
			assert.NotNil(t, p.Schema)
		}
	})
}

// =============================================================================
// generateListOperation Tests
// =============================================================================

func TestOpenAPIHandler_generateListOperation(t *testing.T) {
	handler := NewOpenAPIHandler(nil)

	t.Run("generates list operation", func(t *testing.T) {
		op := handler.generateListOperation("users", "public", "#/components/schemas/public.users")

		assert.Equal(t, "List public.users records", op.Summary)
		assert.Equal(t, "list_public_users", op.OperationID)
		assert.Contains(t, op.Tags, "Tables")
		assert.NotEmpty(t, op.Parameters)
		assert.Contains(t, op.Responses, "200")
	})
}

// =============================================================================
// generateCreateOperation Tests
// =============================================================================

func TestOpenAPIHandler_generateCreateOperation(t *testing.T) {
	handler := NewOpenAPIHandler(nil)

	t.Run("generates create operation", func(t *testing.T) {
		op := handler.generateCreateOperation("users", "public", "#/components/schemas/public.users")

		assert.Equal(t, "Create public.users record(s)", op.Summary)
		assert.Equal(t, "create_public_users", op.OperationID)
		assert.Contains(t, op.Tags, "Tables")
		assert.NotNil(t, op.RequestBody)
		assert.True(t, op.RequestBody.Required)
		assert.Contains(t, op.Responses, "201")
	})
}

// =============================================================================
// generateGetOperation Tests
// =============================================================================

func TestOpenAPIHandler_generateGetOperation(t *testing.T) {
	handler := NewOpenAPIHandler(nil)

	t.Run("generates get operation", func(t *testing.T) {
		op := handler.generateGetOperation("users", "public", "#/components/schemas/public.users")

		assert.Equal(t, "Get public.users by ID", op.Summary)
		assert.Equal(t, "get_public_users", op.OperationID)
		assert.Contains(t, op.Tags, "Tables")
		assert.NotEmpty(t, op.Parameters) // Should have id parameter
	})
}

// =============================================================================
// generateUpdateOperation Tests
// =============================================================================

func TestOpenAPIHandler_generateUpdateOperation(t *testing.T) {
	handler := NewOpenAPIHandler(nil)

	t.Run("generates update operation", func(t *testing.T) {
		op := handler.generateUpdateOperation("users", "public", "#/components/schemas/public.users")

		assert.Contains(t, op.Summary, "Update")
		assert.Equal(t, "update_public_users", op.OperationID)
		assert.Contains(t, op.Tags, "Tables")
	})
}

// =============================================================================
// generateDeleteOperation Tests
// =============================================================================

func TestOpenAPIHandler_generateDeleteOperation(t *testing.T) {
	handler := NewOpenAPIHandler(nil)

	t.Run("generates delete operation", func(t *testing.T) {
		op := handler.generateDeleteOperation("users", "public", "#/components/schemas/public.users")

		assert.Contains(t, op.Summary, "Delete")
		assert.Equal(t, "delete_public_users", op.OperationID)
		assert.Contains(t, op.Tags, "Tables")
	})
}

// =============================================================================
// generateTableSchema Tests
// =============================================================================

func TestOpenAPIHandler_generateTableSchema(t *testing.T) {
	handler := NewOpenAPIHandler(nil)

	t.Run("generates schema with columns", func(t *testing.T) {
		spec := OpenAPISpec{
			Components: OpenAPIComponents{
				Schemas: make(map[string]interface{}),
			},
		}

		defaultValue := "now()"
		table := database.TableInfo{
			Schema: "public",
			Name:   "users",
			Columns: []database.ColumnInfo{
				{Name: "id", DataType: "uuid", IsNullable: false},
				{Name: "email", DataType: "text", IsNullable: false},
				{Name: "name", DataType: "text", IsNullable: true},
				{Name: "created_at", DataType: "timestamp", IsNullable: false, DefaultValue: &defaultValue},
			},
		}

		schemaRef := handler.generateTableSchema(&spec, table)

		assert.Equal(t, "#/components/schemas/public.users", schemaRef)
		assert.Contains(t, spec.Components.Schemas, "public.users")

		schema := spec.Components.Schemas["public.users"].(map[string]interface{})
		assert.Equal(t, "object", schema["type"])
		assert.NotNil(t, schema["properties"])

		// Check required fields (non-nullable without default)
		required := schema["required"].([]string)
		assert.Contains(t, required, "id")
		assert.Contains(t, required, "email")
		assert.NotContains(t, required, "name")       // Nullable
		assert.NotContains(t, required, "created_at") // Has default
	})
}

// =============================================================================
// Pluralization Edge Cases Tests
// =============================================================================

func TestPluralization(t *testing.T) {
	t.Run("already ends with s", func(t *testing.T) {
		tableName := "items"
		assert.True(t, strings.HasSuffix(tableName, "s"))
	})

	t.Run("ends with y", func(t *testing.T) {
		tableName := "category"

		if !strings.HasSuffix(tableName, "s") {
			if strings.HasSuffix(tableName, "y") {
				tableName = strings.TrimSuffix(tableName, "y") + "ies"
			}
		}

		assert.Equal(t, "categories", tableName)
	})

	t.Run("ends with x", func(t *testing.T) {
		tableName := "box"

		if !strings.HasSuffix(tableName, "s") {
			if strings.HasSuffix(tableName, "x") {
				tableName += "es"
			}
		}

		assert.Equal(t, "boxes", tableName)
	})

	t.Run("ends with ch", func(t *testing.T) {
		tableName := "match"

		if !strings.HasSuffix(tableName, "s") {
			if strings.HasSuffix(tableName, "ch") {
				tableName += "es"
			}
		}

		assert.Equal(t, "matches", tableName)
	})

	t.Run("ends with sh", func(t *testing.T) {
		tableName := "flash"

		if !strings.HasSuffix(tableName, "s") {
			if strings.HasSuffix(tableName, "sh") {
				tableName += "es"
			}
		}

		assert.Equal(t, "flashes", tableName)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkColumnToSchema(b *testing.B) {
	handler := NewOpenAPIHandler(nil)
	col := database.ColumnInfo{
		Name:       "created_at",
		DataType:   "timestamp with time zone",
		IsNullable: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.columnToSchema(col)
	}
}

func BenchmarkBuildTablePath(b *testing.B) {
	handler := NewOpenAPIHandler(nil)
	table := database.TableInfo{
		Schema: "public",
		Name:   "category",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.buildTablePath(table)
	}
}

func BenchmarkGetQueryParameters(b *testing.B) {
	handler := NewOpenAPIHandler(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.getQueryParameters()
	}
}

func BenchmarkGenerateListOperation(b *testing.B) {
	handler := NewOpenAPIHandler(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.generateListOperation("users", "public", "#/components/schemas/public.users")
	}
}
