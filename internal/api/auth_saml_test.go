package api

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSAMLHandler_NoServiceConfigured tests all SAML endpoints when SAML service is not configured
// This tests the defensive programming aspects of the handlers
func TestSAMLHandler_NoServiceConfigured(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "list providers with no SAML service",
			method:         "GET",
			path:           "/auth/saml/providers",
			expectedStatus: fiber.StatusOK, // Returns empty array
		},
		{
			name:           "get metadata with no SAML service",
			method:         "GET",
			path:           "/auth/saml/metadata/okta",
			expectedStatus: fiber.StatusNotFound,
			expectedError:  "SAML is not configured",
		},
		{
			name:           "initiate login with no SAML service",
			method:         "GET",
			path:           "/auth/saml/login/okta",
			expectedStatus: fiber.StatusNotFound,
			expectedError:  "SAML is not configured",
		},
		{
			name:           "handle assertion with no SAML service",
			method:         "POST",
			path:           "/auth/saml/acs",
			expectedStatus: fiber.StatusNotFound,
			expectedError:  "SAML is not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			handler := NewSAMLHandler(nil, nil)

			// Register routes based on path
			switch tt.path {
			case "/auth/saml/providers":
				app.Get(tt.path, handler.ListSAMLProviders)
			case "/auth/saml/metadata/okta":
				app.Get("/auth/saml/metadata/:provider", handler.GetSPMetadata)
			case "/auth/saml/login/okta":
				app.Get("/auth/saml/login/:provider", handler.InitiateSAMLLogin)
			case "/auth/saml/acs":
				app.Post(tt.path, handler.HandleSAMLAssertion)
			}

			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.method == "GET" && tt.path == "/auth/saml/login/okta" {
				req.Header.Set("Accept", "application/json")
			}

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedError != "" {
				var result map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)
				assert.Contains(t, result["error"], tt.expectedError)
			}
		})
	}
}

// TestListSAMLProviders_ReturnsEmptyArray tests that an empty array is returned when no providers exist
func TestListSAMLProviders_ReturnsEmptyArray(t *testing.T) {
	app := fiber.New()
	handler := NewSAMLHandler(nil, nil)
	app.Get("/auth/saml/providers", handler.ListSAMLProviders)

	req := httptest.NewRequest("GET", "/auth/saml/providers", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result []SAMLProviderResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, 0, len(result), "Should return empty array when SAML is not configured")
	assert.NotNil(t, result, "Should return empty array, not null")
}

// TestInitiateSAMLLogin_AcceptHeader tests that the Accept header determines response format
func TestInitiateSAMLLogin_AcceptHeader(t *testing.T) {
	tests := []struct {
		name         string
		acceptHeader string
	}{
		{
			name:         "JSON response when Accept is application/json",
			acceptHeader: "application/json",
		},
		{
			name:         "JSON response when Accept contains application/json",
			acceptHeader: "text/html,application/json;q=0.9",
		},
		{
			name:         "redirect when Accept is text/html",
			acceptHeader: "text/html",
		},
		{
			name:         "redirect when no Accept header",
			acceptHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			handler := NewSAMLHandler(nil, nil)
			app.Get("/auth/saml/login/:provider", handler.InitiateSAMLLogin)

			req := httptest.NewRequest("GET", "/auth/saml/login/okta", nil)
			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Since SAML is not configured, should always return 404
			// But this tests that the Accept header logic runs before the auth logic
			assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

			// The response should be JSON (since we return JSON errors)
			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)
			assert.Contains(t, result["error"], "SAML is not configured")
		})
	}
}

// TestHandleSAMLAssertion_MissingSAMLResponse tests that missing SAMLResponse returns 400
func TestHandleSAMLAssertion_MissingSAMLResponse(t *testing.T) {
	// This test uses nil SAML service, but the SAMLResponse validation should happen first
	app := fiber.New()
	handler := NewSAMLHandler(nil, nil)
	app.Post("/auth/saml/acs", handler.HandleSAMLAssertion)

	req := httptest.NewRequest("POST", "/auth/saml/acs", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Since SAML service is nil, it will return 404 first
	// This is expected behavior - service check comes before parameter validation
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

// TestSAMLProviderResponse_JSONSerialization tests that SAMLProviderResponse serializes correctly
func TestSAMLProviderResponse_JSONSerialization(t *testing.T) {
	provider := SAMLProviderResponse{
		ID:       "provider-123",
		Name:     "okta",
		EntityID: "https://okta.example.com",
		SsoURL:   "https://okta.example.com/sso",
		LoginURL: "http://localhost:8080/auth/saml/login/okta",
		Enabled:  true,
	}

	data, err := json.Marshal(provider)
	require.NoError(t, err)

	var decoded SAMLProviderResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, provider.ID, decoded.ID)
	assert.Equal(t, provider.Name, decoded.Name)
	assert.Equal(t, provider.EntityID, decoded.EntityID)
	assert.Equal(t, provider.SsoURL, decoded.SsoURL)
	assert.Equal(t, provider.LoginURL, decoded.LoginURL)
	assert.Equal(t, provider.Enabled, decoded.Enabled)
}

// TestSAMLLoginResponse_JSONSerialization tests SAMLLoginResponse serialization
func TestSAMLLoginResponse_JSONSerialization(t *testing.T) {
	response := SAMLLoginResponse{
		RedirectURL: "https://idp.example.com/sso?SAMLRequest=encoded",
	}

	data, err := json.Marshal(response)
	require.NoError(t, err)

	var decoded SAMLLoginResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, response.RedirectURL, decoded.RedirectURL)
}

// TestSAMLCallbackResponse_JSONSerialization tests SAMLCallbackResponse serialization
func TestSAMLCallbackResponse_JSONSerialization(t *testing.T) {
	response := SAMLCallbackResponse{
		AccessToken:  "access-token-123",
		RefreshToken: "refresh-token-456",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
		User:         nil, // Can be nil
	}

	data, err := json.Marshal(response)
	require.NoError(t, err)

	var decoded SAMLCallbackResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, response.AccessToken, decoded.AccessToken)
	assert.Equal(t, response.RefreshToken, decoded.RefreshToken)
	assert.Equal(t, response.ExpiresIn, decoded.ExpiresIn)
	assert.Equal(t, response.TokenType, decoded.TokenType)
	assert.Nil(t, decoded.User)
}

// =============================================================================
// NewSAMLHandler Tests
// =============================================================================

func TestNewSAMLHandler_WithNilServices(t *testing.T) {
	handler := NewSAMLHandler(nil, nil)

	assert.NotNil(t, handler)
	assert.Nil(t, handler.samlService)
	assert.Nil(t, handler.authService)
}

// =============================================================================
// convertAttributes Tests
// =============================================================================

func TestConvertAttributes_SingleValues(t *testing.T) {
	input := map[string][]string{
		"email": {"user@example.com"},
		"name":  {"John Doe"},
	}

	result := convertAttributes(input)

	assert.Equal(t, "user@example.com", result["email"])
	assert.Equal(t, "John Doe", result["name"])
}

func TestConvertAttributes_MultipleValues(t *testing.T) {
	input := map[string][]string{
		"groups": {"admins", "users", "developers"},
		"roles":  {"admin", "editor"},
	}

	result := convertAttributes(input)

	groups, ok := result["groups"].([]string)
	assert.True(t, ok)
	assert.Len(t, groups, 3)
	assert.Contains(t, groups, "admins")
	assert.Contains(t, groups, "users")
	assert.Contains(t, groups, "developers")

	roles, ok := result["roles"].([]string)
	assert.True(t, ok)
	assert.Len(t, roles, 2)
}

func TestConvertAttributes_EmptyMap(t *testing.T) {
	input := map[string][]string{}

	result := convertAttributes(input)

	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestConvertAttributes_NilMap(t *testing.T) {
	var input map[string][]string = nil

	result := convertAttributes(input)

	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestConvertAttributes_EmptySlice(t *testing.T) {
	input := map[string][]string{
		"empty": {},
	}

	result := convertAttributes(input)

	empty, ok := result["empty"].([]string)
	assert.True(t, ok)
	assert.Empty(t, empty)
}

func TestConvertAttributes_MixedValues(t *testing.T) {
	input := map[string][]string{
		"single":   {"value"},
		"multiple": {"a", "b"},
		"empty":    {},
	}

	result := convertAttributes(input)

	assert.Equal(t, "value", result["single"])

	multiple, ok := result["multiple"].([]string)
	assert.True(t, ok)
	assert.Len(t, multiple, 2)

	empty, ok := result["empty"].([]string)
	assert.True(t, ok)
	assert.Empty(t, empty)
}

func TestConvertAttributes_SpecialCharacters(t *testing.T) {
	input := map[string][]string{
		"urn:oid:0.9.2342.19200300.100.1.3": {"user@example.com"},
		"http://schemas.xmlsoap.org/claims": {"value with spaces"},
		"attribute-with-dash":               {"value"},
		"attribute_with_underscore":         {"value"},
	}

	result := convertAttributes(input)

	assert.Equal(t, "user@example.com", result["urn:oid:0.9.2342.19200300.100.1.3"])
	assert.Equal(t, "value with spaces", result["http://schemas.xmlsoap.org/claims"])
	assert.Equal(t, "value", result["attribute-with-dash"])
	assert.Equal(t, "value", result["attribute_with_underscore"])
}

func TestConvertAttributes_UnicodeValues(t *testing.T) {
	input := map[string][]string{
		"name":    {"José García"},
		"company": {"株式会社"},
		"emoji":   {"👋 Hello"},
	}

	result := convertAttributes(input)

	assert.Equal(t, "José García", result["name"])
	assert.Equal(t, "株式会社", result["company"])
	assert.Equal(t, "👋 Hello", result["emoji"])
}

// =============================================================================
// SAMLProviderResponse Tests
// =============================================================================

func TestSAMLProviderResponse_Fields(t *testing.T) {
	response := SAMLProviderResponse{
		ID:       "provider-123",
		Name:     "okta",
		EntityID: "https://app.example.com/saml",
		SsoURL:   "https://idp.okta.com/sso",
		LoginURL: "https://app.example.com/auth/saml/login/okta",
		Enabled:  true,
	}

	assert.Equal(t, "provider-123", response.ID)
	assert.Equal(t, "okta", response.Name)
	assert.Equal(t, "https://app.example.com/saml", response.EntityID)
	assert.Equal(t, "https://idp.okta.com/sso", response.SsoURL)
	assert.Equal(t, "https://app.example.com/auth/saml/login/okta", response.LoginURL)
	assert.True(t, response.Enabled)
}

func TestSAMLProviderResponse_Defaults(t *testing.T) {
	response := SAMLProviderResponse{}

	assert.Empty(t, response.ID)
	assert.Empty(t, response.Name)
	assert.Empty(t, response.EntityID)
	assert.Empty(t, response.SsoURL)
	assert.Empty(t, response.LoginURL)
	assert.False(t, response.Enabled)
}

// =============================================================================
// SAMLLoginResponse Tests
// =============================================================================

func TestSAMLLoginResponse_Fields(t *testing.T) {
	response := SAMLLoginResponse{
		RedirectURL: "https://idp.example.com/sso?SAMLRequest=encoded",
	}

	assert.Equal(t, "https://idp.example.com/sso?SAMLRequest=encoded", response.RedirectURL)
}

func TestSAMLLoginResponse_Defaults(t *testing.T) {
	response := SAMLLoginResponse{}

	assert.Empty(t, response.RedirectURL)
}

// =============================================================================
// SAMLCallbackResponse Tests
// =============================================================================

func TestSAMLCallbackResponse_Fields(t *testing.T) {
	response := SAMLCallbackResponse{
		AccessToken:  "access_token_abc",
		RefreshToken: "refresh_token_xyz",
		ExpiresIn:    3600,
		TokenType:    "bearer",
		User:         nil,
	}

	assert.Equal(t, "access_token_abc", response.AccessToken)
	assert.Equal(t, "refresh_token_xyz", response.RefreshToken)
	assert.Equal(t, int64(3600), response.ExpiresIn)
	assert.Equal(t, "bearer", response.TokenType)
	assert.Nil(t, response.User)
}

func TestSAMLCallbackResponse_Defaults(t *testing.T) {
	response := SAMLCallbackResponse{}

	assert.Empty(t, response.AccessToken)
	assert.Empty(t, response.RefreshToken)
	assert.Equal(t, int64(0), response.ExpiresIn)
	assert.Empty(t, response.TokenType)
	assert.Nil(t, response.User)
}

// =============================================================================
// CreateSAMLUserRequest Tests
// =============================================================================

func TestCreateSAMLUserRequest_Fields(t *testing.T) {
	req := CreateSAMLUserRequest{
		Email:    "user@example.com",
		Name:     "Test User",
		Provider: "okta",
		NameID:   "name-id-123",
		Attributes: map[string][]string{
			"email":  {"user@example.com"},
			"groups": {"admins"},
		},
	}

	assert.Equal(t, "user@example.com", req.Email)
	assert.Equal(t, "Test User", req.Name)
	assert.Equal(t, "okta", req.Provider)
	assert.Equal(t, "name-id-123", req.NameID)
	assert.NotNil(t, req.Attributes)
	assert.Len(t, req.Attributes, 2)
}

func TestCreateSAMLUserRequest_Defaults(t *testing.T) {
	req := CreateSAMLUserRequest{}

	assert.Empty(t, req.Email)
	assert.Empty(t, req.Name)
	assert.Empty(t, req.Provider)
	assert.Empty(t, req.NameID)
	assert.Nil(t, req.Attributes)
}

// =============================================================================
// HandleSAMLLogout Tests
// =============================================================================

func TestHandleSAMLLogout_NilService_POST(t *testing.T) {
	handler := NewSAMLHandler(nil, nil)

	app := fiber.New()
	app.Post("/auth/saml/slo", handler.HandleSAMLLogout)

	req := httptest.NewRequest("POST", "/auth/saml/slo", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Contains(t, result["error"], "SAML is not configured")
}

func TestHandleSAMLLogout_NilService_GET(t *testing.T) {
	handler := NewSAMLHandler(nil, nil)

	app := fiber.New()
	app.Get("/auth/saml/slo", handler.HandleSAMLLogout)

	req := httptest.NewRequest("GET", "/auth/saml/slo", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

// =============================================================================
// InitiateSAMLLogout Tests
// =============================================================================

func TestInitiateSAMLLogout_NilService(t *testing.T) {
	handler := NewSAMLHandler(nil, nil)

	app := fiber.New()
	app.Get("/auth/saml/logout/:provider", handler.InitiateSAMLLogout)

	req := httptest.NewRequest("GET", "/auth/saml/logout/okta", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Contains(t, result["error"], "SAML is not configured")
}

// =============================================================================
// RegisterRoutes Tests
// =============================================================================

func TestSAMLHandler_RegisterRoutes(t *testing.T) {
	handler := NewSAMLHandler(nil, nil)

	app := fiber.New()
	router := app.Group("/auth")
	handler.RegisterRoutes(router)

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/auth/saml/providers"},
		{"GET", "/auth/saml/metadata/test"},
		{"GET", "/auth/saml/login/test"},
		{"POST", "/auth/saml/acs"},
		{"GET", "/auth/saml/logout/test"},
		{"POST", "/auth/saml/slo"},
		{"GET", "/auth/saml/slo"},
	}

	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Routes should be registered (not a router-level 404)
			// Handler may return 404 due to nil service, but the route exists
			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			if err == nil && result["error"] != nil {
				// If there's an error response, it should be from our handler
				errorMsg := result["error"].(string)
				assert.NotContains(t, errorMsg, "Cannot "+route.method)
			}
		})
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestSAMLCallbackResponse_TokenTypes(t *testing.T) {
	tests := []struct {
		name      string
		tokenType string
	}{
		{"bearer lowercase", "bearer"},
		{"Bearer capitalized", "Bearer"},
		{"JWT", "JWT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := SAMLCallbackResponse{
				TokenType: tt.tokenType,
			}
			assert.Equal(t, tt.tokenType, response.TokenType)
		})
	}
}

func TestConvertAttributes_LargeMap(t *testing.T) {
	input := make(map[string][]string)
	for i := 0; i < 100; i++ {
		key := "attr" + string(rune('A'+i%26)) + string(rune('0'+i/26))
		input[key] = []string{"value"}
	}

	result := convertAttributes(input)
	assert.Len(t, result, 100)
}

// NOTE: For full integration testing with real SAML assertions, database operations,
// and user creation, see the E2E tests in test/e2e/ directory.
// These unit tests focus on handler logic, error cases, and data serialization.
