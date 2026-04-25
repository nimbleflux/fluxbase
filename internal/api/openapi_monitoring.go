package api

// addMonitoringEndpoints adds monitoring endpoints to the spec
func (h *OpenAPIHandler) addMonitoringEndpoints(spec *OpenAPISpec) {
	// Metrics schema
	spec.Components.Schemas["SystemMetrics"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"cpu_usage":    map[string]string{"type": "number"},
			"memory_usage": map[string]string{"type": "number"},
			"disk_usage":   map[string]string{"type": "number"},
			"connections":  map[string]string{"type": "integer"},
			"requests":     map[string]string{"type": "integer"},
			"timestamp":    map[string]string{"type": "string", "format": "date-time"},
		},
	}

	spec.Components.Schemas["HealthStatus"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"status":   map[string]string{"type": "string"},
			"database": map[string]string{"type": "string"},
			"storage":  map[string]string{"type": "string"},
			"uptime":   map[string]string{"type": "integer"},
		},
	}

	// GET /api/v1/monitoring/metrics
	spec.Paths["/api/v1/monitoring/metrics"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "Get system metrics",
			Description: "Get current system metrics",
			OperationID: "monitoring_metrics",
			Tags:        []string{"Monitoring"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "System metrics",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/SystemMetrics"},
						},
					},
				},
			},
		},
	}

	// GET /api/v1/monitoring/health
	spec.Paths["/api/v1/monitoring/health"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "Get health status",
			Description: "Get system health status",
			OperationID: "monitoring_health",
			Tags:        []string{"Monitoring"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Health status",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/HealthStatus"},
						},
					},
				},
			},
		},
	}

	// GET /api/v1/monitoring/logs
	spec.Paths["/api/v1/monitoring/logs"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "Get system logs",
			Description: "Get recent system logs",
			OperationID: "monitoring_logs",
			Tags:        []string{"Monitoring"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "level", In: "query", Description: "Log level filter", Schema: map[string]string{"type": "string"}},
				{Name: "limit", In: "query", Schema: map[string]string{"type": "integer"}},
				{Name: "since", In: "query", Description: "Start time", Schema: map[string]string{"type": "string", "format": "date-time"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "System logs",
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
										"metadata":  map[string]string{"type": "object"},
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
