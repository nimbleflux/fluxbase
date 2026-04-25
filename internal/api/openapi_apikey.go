package api

// addAPIKeyEndpoints adds API key management endpoints to the spec
func (h *OpenAPIHandler) addAPIKeyEndpoints(spec *OpenAPISpec) {
	// API Key schema
	spec.Components.Schemas["APIKey"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":                    map[string]string{"type": "string", "format": "uuid"},
			"name":                  map[string]string{"type": "string"},
			"description":           map[string]string{"type": "string"},
			"key_prefix":            map[string]string{"type": "string"},
			"scopes":                map[string]interface{}{"type": "array", "items": map[string]string{"type": "string"}},
			"rate_limit_per_minute": map[string]string{"type": "integer"},
			"expires_at":            map[string]string{"type": "string", "format": "date-time"},
			"created_at":            map[string]string{"type": "string", "format": "date-time"},
			"last_used_at":          map[string]string{"type": "string", "format": "date-time"},
			"is_active":             map[string]string{"type": "boolean"},
		},
	}

	spec.Components.Schemas["CreateAPIKeyRequest"] = map[string]interface{}{
		"type":     "object",
		"required": []string{"name", "scopes"},
		"properties": map[string]interface{}{
			"name":                  map[string]string{"type": "string"},
			"description":           map[string]string{"type": "string"},
			"scopes":                map[string]interface{}{"type": "array", "items": map[string]string{"type": "string"}},
			"rate_limit_per_minute": map[string]string{"type": "integer"},
			"expires_at":            map[string]string{"type": "string", "format": "date-time"},
		},
	}

	spec.Components.Schemas["CreateAPIKeyResponse"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":                    map[string]string{"type": "string", "format": "uuid"},
			"name":                  map[string]string{"type": "string"},
			"key":                   map[string]string{"type": "string", "description": "The full API key (only shown once)"},
			"key_prefix":            map[string]string{"type": "string"},
			"scopes":                map[string]interface{}{"type": "array", "items": map[string]string{"type": "string"}},
			"rate_limit_per_minute": map[string]string{"type": "integer"},
			"expires_at":            map[string]string{"type": "string", "format": "date-time"},
			"created_at":            map[string]string{"type": "string", "format": "date-time"},
		},
	}

	// GET /api/v1/client-keys - List client keys
	spec.Paths["/api/v1/client-keys"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "List client keys",
			Description: "Get client keys. Regular users see only their own keys; admins can see all keys. When the system setting 'app.auth.allow_user_client_keys' is disabled, only admins can access this endpoint.",
			OperationID: "apikeys_list",
			Tags:        []string{"Client Keys"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "List of client keys",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]interface{}{
								"type":  "array",
								"items": map[string]string{"$ref": "#/components/schemas/APIKey"},
							},
						},
					},
				},
			},
		},
		"post": OpenAPIOperation{
			Summary:     "Create API key",
			Description: "Create a new API key for the authenticated user. When the system setting 'app.auth.allow_user_client_keys' is disabled, only admins can access this endpoint.",
			OperationID: "apikeys_create",
			Tags:        []string{"Client Keys"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]string{"$ref": "#/components/schemas/CreateAPIKeyRequest"},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"201": {
					Description: "API key created",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/CreateAPIKeyResponse"},
						},
					},
				},
			},
		},
	}

	// GET/PATCH/DELETE /api/v1/client-keys/:id
	spec.Paths["/api/v1/client-keys/{id}"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "Get API key",
			Description: "Get details of a specific API key. Users can only view their own keys; admins can view any key. When the system setting 'app.auth.allow_user_client_keys' is disabled, only admins can access this endpoint.",
			OperationID: "apikeys_get",
			Tags:        []string{"Client Keys"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "API key details",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/APIKey"},
						},
					},
				},
				"404": {
					Description: "API key not found",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
		"patch": OpenAPIOperation{
			Summary:     "Update API key",
			Description: "Update an existing API key. Users can only update their own keys. When the system setting 'app.auth.allow_user_client_keys' is disabled, only admins can access this endpoint.",
			OperationID: "apikeys_update",
			Tags:        []string{"Client Keys"},
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
								"name":                  map[string]string{"type": "string"},
								"description":           map[string]string{"type": "string"},
								"scopes":                map[string]interface{}{"type": "array", "items": map[string]string{"type": "string"}},
								"rate_limit_per_minute": map[string]string{"type": "integer"},
							},
						},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "API key updated",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/APIKey"},
						},
					},
				},
			},
		},
		"delete": OpenAPIOperation{
			Summary:     "Delete API key",
			Description: "Delete an API key. Users can only delete their own keys. When the system setting 'app.auth.allow_user_client_keys' is disabled, only admins can access this endpoint.",
			OperationID: "apikeys_delete",
			Tags:        []string{"Client Keys"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"204": {
					Description: "API key deleted",
				},
			},
		},
	}

	// POST /api/v1/client-keys/:id/revoke
	spec.Paths["/api/v1/client-keys/{id}/revoke"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Revoke API key",
			Description: "Revoke an API key (deactivate without deleting). Users can only revoke their own keys. When the system setting 'app.auth.allow_user_client_keys' is disabled, only admins can access this endpoint.",
			OperationID: "apikeys_revoke",
			Tags:        []string{"Client Keys"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "API key revoked",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/APIKey"},
						},
					},
				},
			},
		},
	}
}
