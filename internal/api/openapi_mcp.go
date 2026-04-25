package api

// addMCPEndpoints adds custom MCP tools and resources endpoints to the spec
func (h *OpenAPIHandler) addMCPEndpoints(spec *OpenAPISpec) {
	// MCP Tool schema
	spec.Components.Schemas["MCPTool"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":              map[string]string{"type": "string", "format": "uuid"},
			"name":            map[string]string{"type": "string"},
			"namespace":       map[string]string{"type": "string"},
			"description":     map[string]string{"type": "string"},
			"code":            map[string]string{"type": "string"},
			"input_schema":    map[string]string{"type": "object"},
			"required_scopes": map[string]interface{}{"type": "array", "items": map[string]string{"type": "string"}},
			"timeout_seconds": map[string]string{"type": "integer"},
			"memory_limit_mb": map[string]string{"type": "integer"},
			"allow_net":       map[string]string{"type": "boolean"},
			"allow_env":       map[string]string{"type": "boolean"},
			"allow_read":      map[string]string{"type": "boolean"},
			"allow_write":     map[string]string{"type": "boolean"},
			"enabled":         map[string]string{"type": "boolean"},
			"created_at":      map[string]string{"type": "string", "format": "date-time"},
			"updated_at":      map[string]string{"type": "string", "format": "date-time"},
		},
	}

	// MCP Resource schema
	spec.Components.Schemas["MCPResource"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":              map[string]string{"type": "string", "format": "uuid"},
			"uri":             map[string]string{"type": "string"},
			"name":            map[string]string{"type": "string"},
			"namespace":       map[string]string{"type": "string"},
			"description":     map[string]string{"type": "string"},
			"mime_type":       map[string]string{"type": "string"},
			"code":            map[string]string{"type": "string"},
			"required_scopes": map[string]interface{}{"type": "array", "items": map[string]string{"type": "string"}},
			"timeout_seconds": map[string]string{"type": "integer"},
			"memory_limit_mb": map[string]string{"type": "integer"},
			"allow_net":       map[string]string{"type": "boolean"},
			"allow_env":       map[string]string{"type": "boolean"},
			"is_template":     map[string]string{"type": "boolean"},
			"enabled":         map[string]string{"type": "boolean"},
			"created_at":      map[string]string{"type": "string", "format": "date-time"},
			"updated_at":      map[string]string{"type": "string", "format": "date-time"},
		},
	}

	// GET/POST /api/v1/mcp/tools
	spec.Paths["/api/v1/mcp/tools"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "List custom MCP tools",
			Description: "Get all custom MCP tools. Requires admin access.",
			OperationID: "mcp_tools_list",
			Tags:        []string{"MCP Tools"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "namespace", In: "query", Description: "Filter by namespace", Schema: map[string]string{"type": "string"}},
				{Name: "enabled_only", In: "query", Description: "Only return enabled tools", Schema: map[string]string{"type": "boolean"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "List of MCP tools",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"tools": map[string]interface{}{"type": "array", "items": map[string]string{"$ref": "#/components/schemas/MCPTool"}},
									"count": map[string]string{"type": "integer"},
								},
							},
						},
					},
				},
			},
		},
		"post": OpenAPIOperation{
			Summary:     "Create MCP tool",
			Description: "Create a new custom MCP tool. Requires admin access.",
			OperationID: "mcp_tools_create",
			Tags:        []string{"MCP Tools"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]interface{}{
							"type":     "object",
							"required": []string{"name", "code"},
							"properties": map[string]interface{}{
								"name":            map[string]string{"type": "string"},
								"namespace":       map[string]string{"type": "string"},
								"description":     map[string]string{"type": "string"},
								"code":            map[string]string{"type": "string"},
								"input_schema":    map[string]string{"type": "object"},
								"timeout_seconds": map[string]string{"type": "integer"},
								"memory_limit_mb": map[string]string{"type": "integer"},
								"allow_net":       map[string]string{"type": "boolean"},
								"allow_env":       map[string]string{"type": "boolean"},
								"enabled":         map[string]string{"type": "boolean"},
							},
						},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"201": {
					Description: "Tool created",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/MCPTool"},
						},
					},
				},
			},
		},
	}

	// GET/PUT/DELETE /api/v1/mcp/tools/{id}
	spec.Paths["/api/v1/mcp/tools/{id}"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "Get MCP tool",
			Description: "Get a custom MCP tool by ID. Requires admin access.",
			OperationID: "mcp_tools_get",
			Tags:        []string{"MCP Tools"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Tool details",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/MCPTool"},
						},
					},
				},
			},
		},
		"put": OpenAPIOperation{
			Summary:     "Update MCP tool",
			Description: "Update a custom MCP tool. Requires admin access.",
			OperationID: "mcp_tools_update",
			Tags:        []string{"MCP Tools"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"name":            map[string]string{"type": "string"},
								"namespace":       map[string]string{"type": "string"},
								"description":     map[string]string{"type": "string"},
								"code":            map[string]string{"type": "string"},
								"input_schema":    map[string]string{"type": "object"},
								"timeout_seconds": map[string]string{"type": "integer"},
								"memory_limit_mb": map[string]string{"type": "integer"},
								"allow_net":       map[string]string{"type": "boolean"},
								"allow_env":       map[string]string{"type": "boolean"},
								"enabled":         map[string]string{"type": "boolean"},
							},
						},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Tool updated",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/MCPTool"},
						},
					},
				},
			},
		},
		"delete": OpenAPIOperation{
			Summary:     "Delete MCP tool",
			Description: "Delete a custom MCP tool. Requires admin access.",
			OperationID: "mcp_tools_delete",
			Tags:        []string{"MCP Tools"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"204": {
					Description: "Tool deleted",
				},
			},
		},
	}

	// POST /api/v1/mcp/tools/sync
	spec.Paths["/api/v1/mcp/tools/sync"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Sync MCP tool",
			Description: "Create or update a tool by name (upsert). Requires admin access.",
			OperationID: "mcp_tools_sync",
			Tags:        []string{"MCP Tools"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]interface{}{
							"type":     "object",
							"required": []string{"name", "code"},
							"properties": map[string]interface{}{
								"name":            map[string]string{"type": "string"},
								"namespace":       map[string]string{"type": "string"},
								"description":     map[string]string{"type": "string"},
								"code":            map[string]string{"type": "string"},
								"input_schema":    map[string]string{"type": "object"},
								"timeout_seconds": map[string]string{"type": "integer"},
								"memory_limit_mb": map[string]string{"type": "integer"},
								"allow_net":       map[string]string{"type": "boolean"},
								"allow_env":       map[string]string{"type": "boolean"},
								"enabled":         map[string]string{"type": "boolean"},
							},
						},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Tool synced",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/MCPTool"},
						},
					},
				},
			},
		},
	}

	// POST /api/v1/mcp/tools/{id}/test
	spec.Paths["/api/v1/mcp/tools/{id}/test"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Test MCP tool",
			Description: "Execute a tool with test arguments. Requires admin access.",
			OperationID: "mcp_tools_test",
			Tags:        []string{"MCP Tools"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"args": map[string]string{"type": "object"},
							},
						},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Test result",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"success": map[string]string{"type": "boolean"},
									"result":  map[string]string{"type": "object"},
								},
							},
						},
					},
				},
			},
		},
	}

	// GET/POST /api/v1/mcp/resources
	spec.Paths["/api/v1/mcp/resources"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "List custom MCP resources",
			Description: "Get all custom MCP resources. Requires admin access.",
			OperationID: "mcp_resources_list",
			Tags:        []string{"MCP Tools"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "namespace", In: "query", Description: "Filter by namespace", Schema: map[string]string{"type": "string"}},
				{Name: "enabled_only", In: "query", Description: "Only return enabled resources", Schema: map[string]string{"type": "boolean"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "List of MCP resources",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"resources": map[string]interface{}{"type": "array", "items": map[string]string{"$ref": "#/components/schemas/MCPResource"}},
									"count":     map[string]string{"type": "integer"},
								},
							},
						},
					},
				},
			},
		},
		"post": OpenAPIOperation{
			Summary:     "Create MCP resource",
			Description: "Create a new custom MCP resource. Requires admin access.",
			OperationID: "mcp_resources_create",
			Tags:        []string{"MCP Tools"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]interface{}{
							"type":     "object",
							"required": []string{"uri", "name", "code"},
							"properties": map[string]interface{}{
								"uri":             map[string]string{"type": "string"},
								"name":            map[string]string{"type": "string"},
								"namespace":       map[string]string{"type": "string"},
								"description":     map[string]string{"type": "string"},
								"mime_type":       map[string]string{"type": "string"},
								"code":            map[string]string{"type": "string"},
								"timeout_seconds": map[string]string{"type": "integer"},
								"memory_limit_mb": map[string]string{"type": "integer"},
								"allow_net":       map[string]string{"type": "boolean"},
								"allow_env":       map[string]string{"type": "boolean"},
								"enabled":         map[string]string{"type": "boolean"},
							},
						},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"201": {
					Description: "Resource created",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/MCPResource"},
						},
					},
				},
			},
		},
	}

	// GET/PUT/DELETE /api/v1/mcp/resources/{id}
	spec.Paths["/api/v1/mcp/resources/{id}"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "Get MCP resource",
			Description: "Get a custom MCP resource by ID. Requires admin access.",
			OperationID: "mcp_resources_get",
			Tags:        []string{"MCP Tools"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Resource details",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/MCPResource"},
						},
					},
				},
			},
		},
		"put": OpenAPIOperation{
			Summary:     "Update MCP resource",
			Description: "Update a custom MCP resource. Requires admin access.",
			OperationID: "mcp_resources_update",
			Tags:        []string{"MCP Tools"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"uri":             map[string]string{"type": "string"},
								"name":            map[string]string{"type": "string"},
								"namespace":       map[string]string{"type": "string"},
								"description":     map[string]string{"type": "string"},
								"mime_type":       map[string]string{"type": "string"},
								"code":            map[string]string{"type": "string"},
								"timeout_seconds": map[string]string{"type": "integer"},
								"memory_limit_mb": map[string]string{"type": "integer"},
								"allow_net":       map[string]string{"type": "boolean"},
								"allow_env":       map[string]string{"type": "boolean"},
								"enabled":         map[string]string{"type": "boolean"},
							},
						},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Resource updated",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/MCPResource"},
						},
					},
				},
			},
		},
		"delete": OpenAPIOperation{
			Summary:     "Delete MCP resource",
			Description: "Delete a custom MCP resource. Requires admin access.",
			OperationID: "mcp_resources_delete",
			Tags:        []string{"MCP Tools"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"204": {
					Description: "Resource deleted",
				},
			},
		},
	}

	// POST /api/v1/mcp/resources/sync
	spec.Paths["/api/v1/mcp/resources/sync"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Sync MCP resource",
			Description: "Create or update a resource by URI (upsert). Requires admin access.",
			OperationID: "mcp_resources_sync",
			Tags:        []string{"MCP Tools"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]interface{}{
							"type":     "object",
							"required": []string{"uri", "name", "code"},
							"properties": map[string]interface{}{
								"uri":             map[string]string{"type": "string"},
								"name":            map[string]string{"type": "string"},
								"namespace":       map[string]string{"type": "string"},
								"description":     map[string]string{"type": "string"},
								"mime_type":       map[string]string{"type": "string"},
								"code":            map[string]string{"type": "string"},
								"timeout_seconds": map[string]string{"type": "integer"},
								"memory_limit_mb": map[string]string{"type": "integer"},
								"allow_net":       map[string]string{"type": "boolean"},
								"allow_env":       map[string]string{"type": "boolean"},
								"enabled":         map[string]string{"type": "boolean"},
							},
						},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Resource synced",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/MCPResource"},
						},
					},
				},
			},
		},
	}

	// POST /api/v1/mcp/resources/{id}/test
	spec.Paths["/api/v1/mcp/resources/{id}/test"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Test MCP resource",
			Description: "Read a resource with test parameters. Requires admin access.",
			OperationID: "mcp_resources_test",
			Tags:        []string{"MCP Tools"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"params": map[string]string{"type": "object"},
							},
						},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Test result",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"success":  map[string]string{"type": "boolean"},
									"contents": map[string]interface{}{"type": "array", "items": map[string]string{"type": "object"}},
								},
							},
						},
					},
				},
			},
		},
	}
}
