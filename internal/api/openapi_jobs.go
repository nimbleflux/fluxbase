package api

// addJobsEndpoints adds jobs endpoints to the spec
func (h *OpenAPIHandler) addJobsEndpoints(spec *OpenAPISpec) {
	// Job schemas
	spec.Components.Schemas["Job"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":           map[string]string{"type": "string", "format": "uuid"},
			"function":     map[string]string{"type": "string"},
			"status":       map[string]string{"type": "string"},
			"input":        map[string]string{"type": "object"},
			"output":       map[string]string{"type": "object"},
			"error":        map[string]string{"type": "string"},
			"scheduled_at": map[string]string{"type": "string", "format": "date-time"},
			"started_at":   map[string]string{"type": "string", "format": "date-time"},
			"completed_at": map[string]string{"type": "string", "format": "date-time"},
			"created_at":   map[string]string{"type": "string", "format": "date-time"},
		},
	}

	spec.Components.Schemas["SubmitJobRequest"] = map[string]interface{}{
		"type":     "object",
		"required": []string{"function"},
		"properties": map[string]interface{}{
			"function":     map[string]string{"type": "string"},
			"input":        map[string]string{"type": "object"},
			"scheduled_at": map[string]string{"type": "string", "format": "date-time"},
			"priority":     map[string]string{"type": "integer"},
		},
	}

	// POST /api/v1/jobs/submit
	spec.Paths["/api/v1/jobs/submit"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Submit job",
			Description: "Submit a new job to the queue",
			OperationID: "jobs_submit",
			Tags:        []string{"Jobs"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]string{"$ref": "#/components/schemas/SubmitJobRequest"},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"201": {
					Description: "Job submitted",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Job"},
						},
					},
				},
			},
		},
	}

	// GET /api/v1/jobs
	spec.Paths["/api/v1/jobs"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "List jobs",
			Description: "Get jobs for the authenticated user",
			OperationID: "jobs_list",
			Tags:        []string{"Jobs"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "status", In: "query", Description: "Filter by status", Schema: map[string]string{"type": "string"}},
				{Name: "limit", In: "query", Schema: map[string]string{"type": "integer"}},
				{Name: "offset", In: "query", Schema: map[string]string{"type": "integer"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "List of jobs",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]interface{}{
								"type":  "array",
								"items": map[string]string{"$ref": "#/components/schemas/Job"},
							},
						},
					},
				},
			},
		},
	}

	// GET /api/v1/jobs/:id
	spec.Paths["/api/v1/jobs/{id}"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "Get job",
			Description: "Get job details",
			OperationID: "jobs_get",
			Tags:        []string{"Jobs"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Job details",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Job"},
						},
					},
				},
			},
		},
	}

	// POST /api/v1/jobs/:id/cancel
	spec.Paths["/api/v1/jobs/{id}/cancel"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Cancel job",
			Description: "Cancel a pending or running job",
			OperationID: "jobs_cancel",
			Tags:        []string{"Jobs"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Job cancelled",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Job"},
						},
					},
				},
			},
		},
	}

	// POST /api/v1/jobs/:id/retry
	spec.Paths["/api/v1/jobs/{id}/retry"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Retry job",
			Description: "Retry a failed job",
			OperationID: "jobs_retry",
			Tags:        []string{"Jobs"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Job resubmitted",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Job"},
						},
					},
				},
			},
		},
	}
}
