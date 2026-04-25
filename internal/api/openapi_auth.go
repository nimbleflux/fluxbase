package api

// addAuthEndpoints adds authentication endpoints to the spec
func (h *OpenAPIHandler) addAuthEndpoints(spec *OpenAPISpec) {
	// User schema
	spec.Components.Schemas["User"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":         map[string]string{"type": "string", "format": "uuid"},
			"email":      map[string]string{"type": "string", "format": "email"},
			"created_at": map[string]string{"type": "string", "format": "date-time"},
			"updated_at": map[string]string{"type": "string", "format": "date-time"},
		},
	}

	// Token response schema
	spec.Components.Schemas["TokenResponse"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"access_token":  map[string]string{"type": "string"},
			"refresh_token": map[string]string{"type": "string"},
			"token_type":    map[string]string{"type": "string"},
			"expires_in":    map[string]string{"type": "integer"},
			"user":          map[string]string{"$ref": "#/components/schemas/User"},
		},
	}

	// Signup request schema
	spec.Components.Schemas["SignupRequest"] = map[string]interface{}{
		"type":     "object",
		"required": []string{"email", "password"},
		"properties": map[string]interface{}{
			"email":    map[string]string{"type": "string", "format": "email"},
			"password": map[string]string{"type": "string", "minLength": "8"},
		},
	}

	// Signin request schema
	spec.Components.Schemas["SigninRequest"] = map[string]interface{}{
		"type":     "object",
		"required": []string{"email", "password"},
		"properties": map[string]interface{}{
			"email":    map[string]string{"type": "string", "format": "email"},
			"password": map[string]string{"type": "string"},
		},
	}

	// Refresh request schema
	spec.Components.Schemas["RefreshRequest"] = map[string]interface{}{
		"type":     "object",
		"required": []string{"refresh_token"},
		"properties": map[string]interface{}{
			"refresh_token": map[string]string{"type": "string"},
		},
	}

	// Magic link request schema
	spec.Components.Schemas["MagicLinkRequest"] = map[string]interface{}{
		"type":     "object",
		"required": []string{"email"},
		"properties": map[string]interface{}{
			"email": map[string]string{"type": "string", "format": "email"},
		},
	}

	// Error response schema
	spec.Components.Schemas["Error"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"error": map[string]string{"type": "string"},
		},
	}

	// POST /api/v1/auth/signup
	spec.Paths["/api/v1/auth/signup"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Sign up a new user",
			Description: "Create a new user account with email and password",
			OperationID: "auth_signup",
			Tags:        []string{"Authentication"},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]string{"$ref": "#/components/schemas/SignupRequest"},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "User created successfully",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/TokenResponse"},
						},
					},
				},
				"400": {
					Description: "Invalid request",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
	}

	// POST /api/v1/auth/signin
	spec.Paths["/api/v1/auth/signin"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Sign in a user",
			Description: "Authenticate with email and password to get access tokens",
			OperationID: "auth_signin",
			Tags:        []string{"Authentication"},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]string{"$ref": "#/components/schemas/SigninRequest"},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Successfully authenticated",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/TokenResponse"},
						},
					},
				},
				"401": {
					Description: "Invalid credentials",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
	}

	// POST /api/v1/auth/signout
	spec.Paths["/api/v1/auth/signout"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Sign out a user",
			Description: "Invalidate the current session",
			OperationID: "auth_signout",
			Tags:        []string{"Authentication"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Successfully signed out",
				},
				"401": {
					Description: "Unauthorized",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
	}

	// POST /api/v1/auth/refresh
	spec.Paths["/api/v1/auth/refresh"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Refresh access token",
			Description: "Get a new access token using a refresh token",
			OperationID: "auth_refresh",
			Tags:        []string{"Authentication"},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]string{"$ref": "#/components/schemas/RefreshRequest"},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "New access token issued",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/TokenResponse"},
						},
					},
				},
				"401": {
					Description: "Invalid refresh token",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
	}

	// GET /api/v1/auth/user
	spec.Paths["/api/v1/auth/user"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "Get current user",
			Description: "Get the authenticated user's information",
			OperationID: "auth_get_user",
			Tags:        []string{"Authentication"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "User information",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/User"},
						},
					},
				},
				"401": {
					Description: "Unauthorized",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
		"patch": OpenAPIOperation{
			Summary:     "Update current user",
			Description: "Update the authenticated user's information",
			OperationID: "auth_update_user",
			Tags:        []string{"Authentication"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]string{"$ref": "#/components/schemas/User"},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "User updated successfully",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/User"},
						},
					},
				},
				"401": {
					Description: "Unauthorized",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
	}

	// POST /api/v1/auth/magiclink
	spec.Paths["/api/v1/auth/magiclink"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Request magic link",
			Description: "Send a magic link to the user's email for passwordless authentication",
			OperationID: "auth_magiclink",
			Tags:        []string{"Authentication"},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]string{"$ref": "#/components/schemas/MagicLinkRequest"},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Magic link sent successfully",
				},
				"400": {
					Description: "Invalid request",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
	}

	// GET /api/v1/auth/magiclink/verify
	spec.Paths["/api/v1/auth/magiclink/verify"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "Verify magic link",
			Description: "Verify a magic link token and authenticate the user",
			OperationID: "auth_magiclink_verify",
			Tags:        []string{"Authentication"},
			Parameters: []OpenAPIParameter{
				{
					Name:        "token",
					In:          "query",
					Description: "Magic link token",
					Required:    true,
					Schema:      map[string]string{"type": "string"},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Successfully authenticated",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/TokenResponse"},
						},
					},
				},
				"401": {
					Description: "Invalid or expired token",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
	}
}
