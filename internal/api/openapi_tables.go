package api

import (
	"fmt"
	"strings"

	"github.com/nimbleflux/fluxbase/internal/database"
)

// addTableToSpec adds paths and schema for a table
func (h *OpenAPIHandler) addTableToSpec(spec *OpenAPISpec, table database.TableInfo) {
	tableName := table.Name
	schemaName := table.Schema

	// Build the path (same logic as REST handler)
	path := h.buildTablePath(table)
	pathWithID := path + "/{id}"

	// Generate schema
	schemaRef := h.generateTableSchema(spec, table)

	// Add paths
	spec.Paths[path] = OpenAPIPath{
		"get":    h.generateListOperation(tableName, schemaName, schemaRef),
		"post":   h.generateCreateOperation(tableName, schemaName, schemaRef),
		"patch":  h.generateBatchUpdateOperation(tableName, schemaName, schemaRef),
		"delete": h.generateBatchDeleteOperation(tableName, schemaName, schemaRef),
	}

	spec.Paths[pathWithID] = OpenAPIPath{
		"get":    h.generateGetOperation(tableName, schemaName, schemaRef),
		"put":    h.generateReplaceOperation(tableName, schemaName, schemaRef),
		"patch":  h.generateUpdateOperation(tableName, schemaName, schemaRef),
		"delete": h.generateDeleteOperation(tableName, schemaName, schemaRef),
	}
}

// buildTablePath builds the REST API path for a table
func (h *OpenAPIHandler) buildTablePath(table database.TableInfo) string {
	tableName := table.Name
	if !strings.HasSuffix(tableName, "s") {
		switch {
		case strings.HasSuffix(tableName, "y"):
			tableName = strings.TrimSuffix(tableName, "y") + "ies"
		case strings.HasSuffix(tableName, "x") ||
			strings.HasSuffix(tableName, "ch") ||
			strings.HasSuffix(tableName, "sh"):
			tableName += "es"
		default:
			tableName += "s"
		}
	}

	// All database tables/views are under /api/tables/ prefix
	if table.Schema != "public" {
		return "/api/v1/tables/" + table.Schema + "/" + tableName
	}
	return "/api/v1/tables/" + tableName
}

// generateTableSchema generates JSON schema for a table
func (h *OpenAPIHandler) generateTableSchema(spec *OpenAPISpec, table database.TableInfo) string {
	schemaName := table.Schema + "." + table.Name

	properties := make(map[string]interface{})
	required := []string{}

	for _, col := range table.Columns {
		properties[col.Name] = h.columnToSchema(col)

		if !col.IsNullable && col.DefaultValue == nil {
			required = append(required, col.Name)
		}
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	spec.Components.Schemas[schemaName] = schema
	return "#/components/schemas/" + schemaName
}

// columnToSchema converts a column to JSON schema
func (h *OpenAPIHandler) columnToSchema(col database.ColumnInfo) map[string]interface{} {
	schema := make(map[string]interface{})

	// Map PostgreSQL types to JSON Schema types
	// NOTE: Array check must come BEFORE int check to handle _int4, _int8, etc.
	switch {
	case strings.Contains(col.DataType, "array") || strings.HasSuffix(col.DataType, "[]") || strings.HasPrefix(col.DataType, "_"):
		schema["type"] = "array"
		schema["items"] = map[string]string{"type": "string"}
	case strings.Contains(col.DataType, "int"):
		schema["type"] = "integer"
	case strings.Contains(col.DataType, "numeric") || strings.Contains(col.DataType, "decimal") || strings.Contains(col.DataType, "float") || strings.Contains(col.DataType, "double"):
		schema["type"] = "number"
	case strings.Contains(col.DataType, "bool"):
		schema["type"] = "boolean"
	case strings.Contains(col.DataType, "json"):
		schema["type"] = "object"
	case strings.Contains(col.DataType, "timestamp") || strings.Contains(col.DataType, "date"):
		schema["type"] = "string"
		schema["format"] = "date-time"
	case strings.Contains(col.DataType, "uuid"):
		schema["type"] = "string"
		schema["format"] = "uuid"
	default:
		schema["type"] = "string"
	}

	if col.IsNullable {
		schema["nullable"] = true
	}

	return schema
}

// generateListOperation generates GET operation for listing records
func (h *OpenAPIHandler) generateListOperation(tableName, schemaName, schemaRef string) OpenAPIOperation {
	return OpenAPIOperation{
		Summary:     fmt.Sprintf("List %s.%s records", schemaName, tableName),
		Description: "Query and filter records with PostgREST-compatible syntax",
		OperationID: fmt.Sprintf("list_%s_%s", schemaName, tableName),
		Tags:        []string{"Tables"},
		Parameters:  h.getQueryParameters(),
		Responses: map[string]OpenAPIResponse{
			"200": {
				Description: "Successful response",
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]interface{}{
							"type": "array",
							"items": map[string]string{
								"$ref": schemaRef,
							},
						},
					},
				},
			},
		},
	}
}

// generateCreateOperation generates POST operation for creating records
func (h *OpenAPIHandler) generateCreateOperation(tableName, schemaName, schemaRef string) OpenAPIOperation {
	return OpenAPIOperation{
		Summary:     fmt.Sprintf("Create %s.%s record(s)", schemaName, tableName),
		Description: "Create a single record or batch insert multiple records",
		OperationID: fmt.Sprintf("create_%s_%s", schemaName, tableName),
		Tags:        []string{"Tables"},
		RequestBody: &OpenAPIRequestBody{
			Required: true,
			Content: map[string]OpenAPIMedia{
				"application/json": {
					Schema: map[string]interface{}{
						"oneOf": []interface{}{
							map[string]string{"$ref": schemaRef},
							map[string]interface{}{
								"type": "array",
								"items": map[string]string{
									"$ref": schemaRef,
								},
							},
						},
					},
				},
			},
		},
		Responses: map[string]OpenAPIResponse{
			"201": {
				Description: "Created successfully",
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]interface{}{
							"oneOf": []interface{}{
								map[string]string{"$ref": schemaRef},
								map[string]interface{}{
									"type": "array",
									"items": map[string]string{
										"$ref": schemaRef,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// generateGetOperation generates GET by ID operation
func (h *OpenAPIHandler) generateGetOperation(tableName, schemaName, schemaRef string) OpenAPIOperation {
	return OpenAPIOperation{
		Summary:     fmt.Sprintf("Get %s.%s by ID", schemaName, tableName),
		OperationID: fmt.Sprintf("get_%s_%s", schemaName, tableName),
		Tags:        []string{"Tables"},
		Parameters: []OpenAPIParameter{
			{
				Name:        "id",
				In:          "path",
				Description: "Record ID",
				Required:    true,
				Schema:      map[string]string{"type": "string"},
			},
		},
		Responses: map[string]OpenAPIResponse{
			"200": {
				Description: "Successful response",
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]string{"$ref": schemaRef},
					},
				},
			},
			"404": {
				Description: "Record not found",
			},
		},
	}
}

// generateUpdateOperation generates PATCH operation
func (h *OpenAPIHandler) generateUpdateOperation(tableName, schemaName, schemaRef string) OpenAPIOperation {
	return OpenAPIOperation{
		Summary:     fmt.Sprintf("Update %s.%s", schemaName, tableName),
		OperationID: fmt.Sprintf("update_%s_%s", schemaName, tableName),
		Tags:        []string{"Tables"},
		Parameters: []OpenAPIParameter{
			{
				Name:        "id",
				In:          "path",
				Description: "Record ID",
				Required:    true,
				Schema:      map[string]string{"type": "string"},
			},
		},
		RequestBody: &OpenAPIRequestBody{
			Required: true,
			Content: map[string]OpenAPIMedia{
				"application/json": {
					Schema: map[string]string{"$ref": schemaRef},
				},
			},
		},
		Responses: map[string]OpenAPIResponse{
			"200": {
				Description: "Updated successfully",
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]string{"$ref": schemaRef},
					},
				},
			},
		},
	}
}

// generateReplaceOperation generates PUT operation
func (h *OpenAPIHandler) generateReplaceOperation(tableName, schemaName, schemaRef string) OpenAPIOperation {
	return OpenAPIOperation{
		Summary:     fmt.Sprintf("Replace %s.%s", schemaName, tableName),
		OperationID: fmt.Sprintf("replace_%s_%s", schemaName, tableName),
		Tags:        []string{"Tables"},
		Parameters: []OpenAPIParameter{
			{
				Name:        "id",
				In:          "path",
				Description: "Record ID",
				Required:    true,
				Schema:      map[string]string{"type": "string"},
			},
		},
		RequestBody: &OpenAPIRequestBody{
			Required: true,
			Content: map[string]OpenAPIMedia{
				"application/json": {
					Schema: map[string]string{"$ref": schemaRef},
				},
			},
		},
		Responses: map[string]OpenAPIResponse{
			"200": {
				Description: "Replaced successfully",
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]string{"$ref": schemaRef},
					},
				},
			},
		},
	}
}

// generateDeleteOperation generates DELETE operation
func (h *OpenAPIHandler) generateDeleteOperation(tableName, schemaName, schemaRef string) OpenAPIOperation {
	return OpenAPIOperation{
		Summary:     fmt.Sprintf("Delete %s.%s", schemaName, tableName),
		OperationID: fmt.Sprintf("delete_%s_%s", schemaName, tableName),
		Tags:        []string{"Tables"},
		Parameters: []OpenAPIParameter{
			{
				Name:        "id",
				In:          "path",
				Description: "Record ID",
				Required:    true,
				Schema:      map[string]string{"type": "string"},
			},
		},
		Responses: map[string]OpenAPIResponse{
			"204": {
				Description: "Deleted successfully",
			},
			"404": {
				Description: "Record not found",
			},
		},
	}
}

// generateBatchUpdateOperation generates batch PATCH operation
func (h *OpenAPIHandler) generateBatchUpdateOperation(tableName, schemaName, schemaRef string) OpenAPIOperation {
	return OpenAPIOperation{
		Summary:     fmt.Sprintf("Batch update %s.%s records", schemaName, tableName),
		Description: "Update multiple records matching the filter criteria",
		OperationID: fmt.Sprintf("batch_update_%s_%s", schemaName, tableName),
		Tags:        []string{"Tables"},
		Parameters:  h.getQueryParameters(),
		RequestBody: &OpenAPIRequestBody{
			Required: true,
			Content: map[string]OpenAPIMedia{
				"application/json": {
					Schema: map[string]string{"$ref": schemaRef},
				},
			},
		},
		Responses: map[string]OpenAPIResponse{
			"200": {
				Description: "Updated successfully",
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]interface{}{
							"type": "array",
							"items": map[string]string{
								"$ref": schemaRef,
							},
						},
					},
				},
			},
		},
	}
}

// generateBatchDeleteOperation generates batch DELETE operation
func (h *OpenAPIHandler) generateBatchDeleteOperation(tableName, schemaName, schemaRef string) OpenAPIOperation {
	return OpenAPIOperation{
		Summary:     fmt.Sprintf("Batch delete %s.%s records", schemaName, tableName),
		Description: "Delete multiple records matching the filter criteria (requires at least one filter)",
		OperationID: fmt.Sprintf("batch_delete_%s_%s", schemaName, tableName),
		Tags:        []string{"Tables"},
		Parameters:  h.getQueryParameters(),
		Responses: map[string]OpenAPIResponse{
			"200": {
				Description: "Deleted successfully",
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"deleted": map[string]string{"type": "integer"},
								"records": map[string]interface{}{
									"type": "array",
									"items": map[string]string{
										"$ref": schemaRef,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// getQueryParameters returns common query parameters
func (h *OpenAPIHandler) getQueryParameters() []OpenAPIParameter {
	return []OpenAPIParameter{
		{
			Name:        "select",
			In:          "query",
			Description: "Columns to select (comma-separated)",
			Schema:      map[string]string{"type": "string"},
		},
		{
			Name:        "order",
			In:          "query",
			Description: "Order by column (e.g., name.asc, created_at.desc)",
			Schema:      map[string]string{"type": "string"},
		},
		{
			Name:        "limit",
			In:          "query",
			Description: "Limit number of results",
			Schema:      map[string]string{"type": "integer"},
		},
		{
			Name:        "offset",
			In:          "query",
			Description: "Offset for pagination",
			Schema:      map[string]string{"type": "integer"},
		},
		{
			Name:        "filter",
			In:          "query",
			Description: "Filter using column.operator=value (e.g., name.eq=John, age.gt=18)",
			Schema:      map[string]string{"type": "string"},
		},
	}
}
