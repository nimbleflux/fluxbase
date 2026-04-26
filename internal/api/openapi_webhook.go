package api

// addWebhookEndpoints adds webhook management endpoints to the spec
func (h *OpenAPIHandler) addWebhookEndpoints(spec *OpenAPISpec) {
	// Webhook schema
	spec.Components.Schemas["Webhook"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":          map[string]string{"type": "string", "format": "uuid"},
			"name":        map[string]string{"type": "string"},
			"url":         map[string]string{"type": "string", "format": "uri"},
			"events":      map[string]interface{}{"type": "array", "items": map[string]string{"type": "string"}},
			"secret":      map[string]string{"type": "string"},
			"is_active":   map[string]string{"type": "boolean"},
			"created_at":  map[string]string{"type": "string", "format": "date-time"},
			"updated_at":  map[string]string{"type": "string", "format": "date-time"},
			"last_status": map[string]string{"type": "string"},
		},
	}

	spec.Components.Schemas["WebhookDelivery"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":            map[string]string{"type": "string", "format": "uuid"},
			"webhook_id":    map[string]string{"type": "string", "format": "uuid"},
			"event":         map[string]string{"type": "string"},
			"payload":       map[string]string{"type": "object"},
			"response_code": map[string]string{"type": "integer"},
			"response_body": map[string]string{"type": "string"},
			"created_at":    map[string]string{"type": "string", "format": "date-time"},
			"delivered_at":  map[string]string{"type": "string", "format": "date-time"},
		},
	}

	// GET/POST /api/v1/webhooks
	spec.Paths["/api/v1/webhooks"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "List webhooks",
			Description: "Get all webhooks",
			OperationID: "webhooks_list",
			Tags:        []string{"Webhooks"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "List of webhooks",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]interface{}{
								"type":  "array",
								"items": map[string]string{"$ref": "#/components/schemas/Webhook"},
							},
						},
					},
				},
			},
		},
		"post": OpenAPIOperation{
			Summary:     "Create webhook",
			Description: "Create a new webhook",
			OperationID: "webhooks_create",
			Tags:        []string{"Webhooks"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]interface{}{
							"type":     "object",
							"required": []string{"name", "url", "events"},
							"properties": map[string]interface{}{
								"name":   map[string]string{"type": "string"},
								"url":    map[string]string{"type": "string", "format": "uri"},
								"events": map[string]interface{}{"type": "array", "items": map[string]string{"type": "string"}},
								"secret": map[string]string{"type": "string"},
							},
						},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"201": {
					Description: "Webhook created",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Webhook"},
						},
					},
				},
			},
		},
	}

	// GET/PATCH/DELETE /api/v1/webhooks/:id
	spec.Paths["/api/v1/webhooks/{id}"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "Get webhook",
			Description: "Get webhook details",
			OperationID: "webhooks_get",
			Tags:        []string{"Webhooks"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Webhook details",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Webhook"},
						},
					},
				},
			},
		},
		"patch": OpenAPIOperation{
			Summary:     "Update webhook",
			Description: "Update a webhook",
			OperationID: "webhooks_update",
			Tags:        []string{"Webhooks"},
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
								"name":      map[string]string{"type": "string"},
								"url":       map[string]string{"type": "string", "format": "uri"},
								"events":    map[string]interface{}{"type": "array", "items": map[string]string{"type": "string"}},
								"is_active": map[string]string{"type": "boolean"},
							},
						},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Webhook updated",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Webhook"},
						},
					},
				},
			},
		},
		"delete": OpenAPIOperation{
			Summary:     "Delete webhook",
			Description: "Delete a webhook",
			OperationID: "webhooks_delete",
			Tags:        []string{"Webhooks"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"204": {
					Description: "Webhook deleted",
				},
			},
		},
	}

	// POST /api/v1/webhooks/:id/test
	spec.Paths["/api/v1/webhooks/{id}/test"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Test webhook",
			Description: "Send a test event to the webhook",
			OperationID: "webhooks_test",
			Tags:        []string{"Webhooks"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Test event sent",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"success":       map[string]string{"type": "boolean"},
									"response_code": map[string]string{"type": "integer"},
									"response_body": map[string]string{"type": "string"},
								},
							},
						},
					},
				},
			},
		},
	}

	// GET /api/v1/webhooks/:id/deliveries
	spec.Paths["/api/v1/webhooks/{id}/deliveries"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "List webhook deliveries",
			Description: "Get delivery history for a webhook",
			OperationID: "webhooks_deliveries",
			Tags:        []string{"Webhooks"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
				{Name: "limit", In: "query", Schema: map[string]string{"type": "integer"}},
				{Name: "offset", In: "query", Schema: map[string]string{"type": "integer"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "List of deliveries",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]interface{}{
								"type":  "array",
								"items": map[string]string{"$ref": "#/components/schemas/WebhookDelivery"},
							},
						},
					},
				},
			},
		},
	}
}
