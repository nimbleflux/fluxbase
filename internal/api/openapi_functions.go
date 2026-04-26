package api

// addFunctionsEndpoints adds edge functions endpoints to the spec
func (h *OpenAPIHandler) addFunctionsEndpoints(spec *OpenAPISpec) {
	// Function schemas
	spec.Components.Schemas["EdgeFunction"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"namespace":   map[string]string{"type": "string"},
			"name":        map[string]string{"type": "string"},
			"description": map[string]string{"type": "string"},
			"version":     map[string]string{"type": "string"},
			"created_at":  map[string]string{"type": "string", "format": "date-time"},
			"updated_at":  map[string]string{"type": "string", "format": "date-time"},
		},
	}

	// GET /api/v1/functions
	spec.Paths["/api/v1/functions"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "List functions",
			Description: "Get all edge functions",
			OperationID: "functions_list",
			Tags:        []string{"Functions"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "List of functions",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]interface{}{
								"type":  "array",
								"items": map[string]string{"$ref": "#/components/schemas/EdgeFunction"},
							},
						},
					},
				},
			},
		},
	}

	// POST /api/v1/functions/:namespace/:name - Invoke function
	spec.Paths["/api/v1/functions/{namespace}/{name}"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Invoke function",
			Description: "Invoke an edge function",
			OperationID: "functions_invoke",
			Tags:        []string{"Functions"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "namespace", In: "path", Required: true, Schema: map[string]string{"type": "string"}},
				{Name: "name", In: "path", Required: true, Schema: map[string]string{"type": "string"}},
			},
			RequestBody: &OpenAPIRequestBody{
				Description: "Function input (optional)",
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]string{"type": "object"},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Function response",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"type": "object"},
						},
					},
				},
			},
		},
		"get": OpenAPIOperation{
			Summary:     "Invoke function (GET)",
			Description: "Invoke an edge function with GET request",
			OperationID: "functions_invoke_get",
			Tags:        []string{"Functions"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "namespace", In: "path", Required: true, Schema: map[string]string{"type": "string"}},
				{Name: "name", In: "path", Required: true, Schema: map[string]string{"type": "string"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Function response",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"type": "object"},
						},
					},
				},
			},
		},
	}
}
