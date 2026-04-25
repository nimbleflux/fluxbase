package api

// addAIEndpoints adds AI/chatbot endpoints to the spec
func (h *OpenAPIHandler) addAIEndpoints(spec *OpenAPISpec) {
	// AI schemas
	spec.Components.Schemas["Chatbot"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":            map[string]string{"type": "string", "format": "uuid"},
			"name":          map[string]string{"type": "string"},
			"description":   map[string]string{"type": "string"},
			"system_prompt": map[string]string{"type": "string"},
			"model":         map[string]string{"type": "string"},
			"is_public":     map[string]string{"type": "boolean"},
			"created_at":    map[string]string{"type": "string", "format": "date-time"},
		},
	}

	spec.Components.Schemas["Conversation"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":         map[string]string{"type": "string", "format": "uuid"},
			"chatbot_id": map[string]string{"type": "string", "format": "uuid"},
			"user_id":    map[string]string{"type": "string", "format": "uuid"},
			"title":      map[string]string{"type": "string"},
			"created_at": map[string]string{"type": "string", "format": "date-time"},
			"updated_at": map[string]string{"type": "string", "format": "date-time"},
		},
	}

	// GET /ai/ws - WebSocket for AI chat
	spec.Paths["/ai/ws"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "AI Chat WebSocket",
			Description: "Establish a WebSocket connection for AI chat. Upgrade to WebSocket protocol required.",
			OperationID: "ai_chat_ws",
			Tags:        []string{"AI"},
			Responses: map[string]OpenAPIResponse{
				"101": {
					Description: "Switching Protocols - WebSocket connection established",
				},
			},
		},
	}

	// GET /api/v1/ai/chatbots
	spec.Paths["/api/v1/ai/chatbots"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "List chatbots",
			Description: "Get all public chatbots",
			OperationID: "ai_chatbots_list",
			Tags:        []string{"AI"},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "List of chatbots",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]interface{}{
								"type":  "array",
								"items": map[string]string{"$ref": "#/components/schemas/Chatbot"},
							},
						},
					},
				},
			},
		},
	}

	// GET /api/v1/ai/chatbots/:id
	spec.Paths["/api/v1/ai/chatbots/{id}"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "Get chatbot",
			Description: "Get chatbot details",
			OperationID: "ai_chatbot_get",
			Tags:        []string{"AI"},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Chatbot details",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Chatbot"},
						},
					},
				},
			},
		},
	}

	// GET /api/v1/ai/conversations
	spec.Paths["/api/v1/ai/conversations"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "List conversations",
			Description: "Get all conversations for the authenticated user",
			OperationID: "ai_conversations_list",
			Tags:        []string{"AI"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "chatbot_id", In: "query", Description: "Filter by chatbot", Schema: map[string]string{"type": "string", "format": "uuid"}},
				{Name: "limit", In: "query", Schema: map[string]string{"type": "integer"}},
				{Name: "offset", In: "query", Schema: map[string]string{"type": "integer"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "List of conversations",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]interface{}{
								"type":  "array",
								"items": map[string]string{"$ref": "#/components/schemas/Conversation"},
							},
						},
					},
				},
			},
		},
	}

	// GET/PATCH/DELETE /api/v1/ai/conversations/:id
	spec.Paths["/api/v1/ai/conversations/{id}"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "Get conversation",
			Description: "Get conversation details",
			OperationID: "ai_conversation_get",
			Tags:        []string{"AI"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Conversation details",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Conversation"},
						},
					},
				},
			},
		},
		"patch": OpenAPIOperation{
			Summary:     "Update conversation",
			Description: "Update conversation (e.g., title)",
			OperationID: "ai_conversation_update",
			Tags:        []string{"AI"},
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
								"title": map[string]string{"type": "string"},
							},
						},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Conversation updated",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Conversation"},
						},
					},
				},
			},
		},
		"delete": OpenAPIOperation{
			Summary:     "Delete conversation",
			Description: "Delete a conversation",
			OperationID: "ai_conversation_delete",
			Tags:        []string{"AI"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"204": {
					Description: "Conversation deleted",
				},
			},
		},
	}
}
