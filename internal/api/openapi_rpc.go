package api

// addRPCEndpoints adds RPC endpoints to the spec
func (h *OpenAPIHandler) addRPCEndpoints(spec *OpenAPISpec) {
	// RPC schemas
	spec.Components.Schemas["RPCProcedure"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"namespace":   map[string]string{"type": "string"},
			"name":        map[string]string{"type": "string"},
			"description": map[string]string{"type": "string"},
			"is_public":   map[string]string{"type": "boolean"},
			"parameters":  map[string]interface{}{"type": "array", "items": map[string]string{"type": "object"}},
		},
	}

	spec.Components.Schemas["RPCExecution"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":           map[string]string{"type": "string", "format": "uuid"},
			"procedure":    map[string]string{"type": "string"},
			"status":       map[string]string{"type": "string"},
			"result":       map[string]string{"type": "object"},
			"error":        map[string]string{"type": "string"},
			"started_at":   map[string]string{"type": "string", "format": "date-time"},
			"completed_at": map[string]string{"type": "string", "format": "date-time"},
		},
	}

	// GET /api/v1/rpc/procedures
	spec.Paths["/api/v1/rpc/procedures"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "List RPC procedures",
			Description: "Get all available RPC procedures",
			OperationID: "rpc_procedures_list",
			Tags:        []string{"RPC"},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "List of procedures",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]interface{}{
								"type":  "array",
								"items": map[string]string{"$ref": "#/components/schemas/RPCProcedure"},
							},
						},
					},
				},
			},
		},
	}

	// POST /api/v1/rpc/:namespace/:name
	spec.Paths["/api/v1/rpc/{namespace}/{name}"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Invoke RPC procedure",
			Description: "Call a remote procedure",
			OperationID: "rpc_invoke",
			Tags:        []string{"RPC"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "namespace", In: "path", Required: true, Schema: map[string]string{"type": "string"}},
				{Name: "name", In: "path", Required: true, Schema: map[string]string{"type": "string"}},
			},
			RequestBody: &OpenAPIRequestBody{
				Description: "Procedure arguments",
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]string{"type": "object"},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Procedure result",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/RPCExecution"},
						},
					},
				},
			},
		},
	}

	// GET /api/v1/rpc/executions/:id
	spec.Paths["/api/v1/rpc/executions/{id}"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "Get execution status",
			Description: "Get the status of an RPC execution",
			OperationID: "rpc_execution_get",
			Tags:        []string{"RPC"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Execution details",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/RPCExecution"},
						},
					},
				},
			},
		},
	}

	// GET /api/v1/rpc/executions/:id/logs
	spec.Paths["/api/v1/rpc/executions/{id}/logs"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "Get execution logs",
			Description: "Get logs for an RPC execution",
			OperationID: "rpc_execution_logs",
			Tags:        []string{"RPC"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Execution logs",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"timestamp": map[string]string{"type": "string", "format": "date-time"},
										"level":     map[string]string{"type": "string"},
										"message":   map[string]string{"type": "string"},
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
