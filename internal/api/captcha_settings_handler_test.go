package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// CaptchaSettingsHandler Construction Tests
// =============================================================================

func TestNewCaptchaSettingsHandler(t *testing.T) {
	t.Run("creates handler with nil dependencies", func(t *testing.T) {
		handler := NewCaptchaSettingsHandler(nil, nil, nil, nil, nil)
		assert.NotNil(t, handler)
		assert.Nil(t, handler.settingsService)
		assert.Nil(t, handler.settingsCache)
		assert.Nil(t, handler.secretsService)
		assert.Nil(t, handler.envConfig)
		assert.Nil(t, handler.captchaService)
	})
}

// =============================================================================
// CaptchaSettingsResponse Struct Tests
// =============================================================================

func TestCaptchaSettingsResponse_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		resp := CaptchaSettingsResponse{
			Enabled:        true,
			Provider:       "hcaptcha",
			SiteKey:        "10000000-ffff-ffff-ffff-000000000001",
			SecretKeySet:   true,
			ScoreThreshold: 0.5,
			Endpoints:      []string{"signup", "login"},
			CapServerURL:   "https://cap.example.com",
			CapAPIKeySet:   false,
			Overrides:      make(map[string]OverrideInfo),
		}

		assert.True(t, resp.Enabled)
		assert.Equal(t, "hcaptcha", resp.Provider)
		assert.Equal(t, "10000000-ffff-ffff-ffff-000000000001", resp.SiteKey)
		assert.True(t, resp.SecretKeySet)
		assert.Equal(t, 0.5, resp.ScoreThreshold)
		assert.Len(t, resp.Endpoints, 2)
		assert.Contains(t, resp.Endpoints, "signup")
		assert.Contains(t, resp.Endpoints, "login")
		assert.Equal(t, "https://cap.example.com", resp.CapServerURL)
		assert.False(t, resp.CapAPIKeySet)
		assert.NotNil(t, resp.Overrides)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		resp := CaptchaSettingsResponse{
			Enabled:        true,
			Provider:       "recaptcha_v3",
			SiteKey:        "6LcXXXX",
			SecretKeySet:   true,
			ScoreThreshold: 0.7,
			Endpoints:      []string{"signup", "login", "password_reset"},
			Overrides:      make(map[string]OverrideInfo),
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"enabled":true`)
		assert.Contains(t, string(data), `"provider":"recaptcha_v3"`)
		assert.Contains(t, string(data), `"site_key":"6LcXXXX"`)
		assert.Contains(t, string(data), `"secret_key_set":true`)
		assert.Contains(t, string(data), `"score_threshold":0.7`)
		assert.Contains(t, string(data), `"endpoints"`)
	})

	t.Run("JSON serialization with overrides", func(t *testing.T) {
		resp := CaptchaSettingsResponse{
			Enabled:  true,
			Provider: "turnstile",
			Overrides: map[string]OverrideInfo{
				"enabled": {
					IsOverridden: true,
					EnvVar:       "FLUXBASE_SECURITY_CAPTCHA_ENABLED",
				},
			},
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"_overrides"`)
		assert.Contains(t, string(data), `"is_overridden":true`)
		assert.Contains(t, string(data), `"FLUXBASE_SECURITY_CAPTCHA_ENABLED"`)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{
			"enabled": true,
			"provider": "hcaptcha",
			"site_key": "test-site-key",
			"secret_key_set": false,
			"score_threshold": 0.5,
			"endpoints": ["signup", "login"],
			"cap_server_url": "",
			"cap_api_key_set": false
		}`

		var resp CaptchaSettingsResponse
		err := json.Unmarshal([]byte(jsonData), &resp)
		require.NoError(t, err)

		assert.True(t, resp.Enabled)
		assert.Equal(t, "hcaptcha", resp.Provider)
		assert.Equal(t, "test-site-key", resp.SiteKey)
		assert.False(t, resp.SecretKeySet)
		assert.Equal(t, 0.5, resp.ScoreThreshold)
		assert.Len(t, resp.Endpoints, 2)
	})

	t.Run("sensitive fields not exposed", func(t *testing.T) {
		resp := CaptchaSettingsResponse{
			SecretKeySet: true,
			CapAPIKeySet: true,
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		// Should only contain boolean flags, not actual secrets
		assert.Contains(t, string(data), `"secret_key_set":true`)
		assert.Contains(t, string(data), `"cap_api_key_set":true`)
		assert.NotContains(t, string(data), `"secret_key":`)
		assert.NotContains(t, string(data), `"cap_api_key":`)
	})
}

// =============================================================================
// UpdateCaptchaSettingsRequest Struct Tests
// =============================================================================

func TestUpdateCaptchaSettingsRequest_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		enabled := true
		provider := "recaptcha_v3"
		siteKey := "6LcXXXXXXXXXX"
		secretKey := "secret-key-value"
		scoreThreshold := 0.7
		endpoints := []string{"signup", "login"}
		capServerURL := "https://cap.example.com"
		capAPIKey := "cap-api-key"

		req := UpdateCaptchaSettingsRequest{
			Enabled:        &enabled,
			Provider:       &provider,
			SiteKey:        &siteKey,
			SecretKey:      &secretKey,
			ScoreThreshold: &scoreThreshold,
			Endpoints:      &endpoints,
			CapServerURL:   &capServerURL,
			CapAPIKey:      &capAPIKey,
		}

		assert.NotNil(t, req.Enabled)
		assert.True(t, *req.Enabled)
		assert.NotNil(t, req.Provider)
		assert.Equal(t, "recaptcha_v3", *req.Provider)
		assert.NotNil(t, req.SiteKey)
		assert.Equal(t, "6LcXXXXXXXXXX", *req.SiteKey)
		assert.NotNil(t, req.SecretKey)
		assert.Equal(t, "secret-key-value", *req.SecretKey)
		assert.NotNil(t, req.ScoreThreshold)
		assert.Equal(t, 0.7, *req.ScoreThreshold)
		assert.NotNil(t, req.Endpoints)
		assert.Len(t, *req.Endpoints, 2)
		assert.NotNil(t, req.CapServerURL)
		assert.Equal(t, "https://cap.example.com", *req.CapServerURL)
		assert.NotNil(t, req.CapAPIKey)
		assert.Equal(t, "cap-api-key", *req.CapAPIKey)
	})

	t.Run("JSON deserialization partial update", func(t *testing.T) {
		jsonData := `{
			"enabled": false,
			"provider": "turnstile"
		}`

		var req UpdateCaptchaSettingsRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.NotNil(t, req.Enabled)
		assert.False(t, *req.Enabled)
		assert.NotNil(t, req.Provider)
		assert.Equal(t, "turnstile", *req.Provider)
		assert.Nil(t, req.SiteKey)
		assert.Nil(t, req.SecretKey)
		assert.Nil(t, req.ScoreThreshold)
		assert.Nil(t, req.Endpoints)
	})

	t.Run("JSON deserialization with all fields", func(t *testing.T) {
		jsonData := `{
			"enabled": true,
			"provider": "hcaptcha",
			"site_key": "site-key-123",
			"secret_key": "secret-key-456",
			"score_threshold": 0.6,
			"endpoints": ["signup", "login", "password_reset", "magic_link"],
			"cap_server_url": "https://cap.example.com",
			"cap_api_key": "cap-key"
		}`

		var req UpdateCaptchaSettingsRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.True(t, *req.Enabled)
		assert.Equal(t, "hcaptcha", *req.Provider)
		assert.Equal(t, "site-key-123", *req.SiteKey)
		assert.Equal(t, "secret-key-456", *req.SecretKey)
		assert.Equal(t, 0.6, *req.ScoreThreshold)
		assert.Len(t, *req.Endpoints, 4)
		assert.Equal(t, "https://cap.example.com", *req.CapServerURL)
		assert.Equal(t, "cap-key", *req.CapAPIKey)
	})

	t.Run("empty update request", func(t *testing.T) {
		jsonData := `{}`

		var req UpdateCaptchaSettingsRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Nil(t, req.Enabled)
		assert.Nil(t, req.Provider)
		assert.Nil(t, req.SiteKey)
		assert.Nil(t, req.SecretKey)
		assert.Nil(t, req.ScoreThreshold)
		assert.Nil(t, req.Endpoints)
		assert.Nil(t, req.CapServerURL)
		assert.Nil(t, req.CapAPIKey)
	})
}

// =============================================================================
// Valid Providers Tests
// =============================================================================

func TestValidProviders(t *testing.T) {
	t.Run("hcaptcha is valid", func(t *testing.T) {
		assert.True(t, validProviders["hcaptcha"])
	})

	t.Run("recaptcha_v3 is valid", func(t *testing.T) {
		assert.True(t, validProviders["recaptcha_v3"])
	})

	t.Run("turnstile is valid", func(t *testing.T) {
		assert.True(t, validProviders["turnstile"])
	})

	t.Run("cap is valid", func(t *testing.T) {
		assert.True(t, validProviders["cap"])
	})

	t.Run("invalid providers not in map", func(t *testing.T) {
		invalidProviders := []string{
			"recaptcha",
			"recaptcha_v2",
			"google",
			"cloudflare",
			"",
			"invalid",
			"HCAPTCHA", // case sensitive
		}

		for _, provider := range invalidProviders {
			assert.False(t, validProviders[provider], "Expected %q to be invalid", provider)
		}
	})
}

// =============================================================================
// Valid Endpoints Tests
// =============================================================================

func TestValidEndpoints(t *testing.T) {
	t.Run("signup is valid", func(t *testing.T) {
		assert.True(t, validEndpoints["signup"])
	})

	t.Run("login is valid", func(t *testing.T) {
		assert.True(t, validEndpoints["login"])
	})

	t.Run("password_reset is valid", func(t *testing.T) {
		assert.True(t, validEndpoints["password_reset"])
	})

	t.Run("magic_link is valid", func(t *testing.T) {
		assert.True(t, validEndpoints["magic_link"])
	})

	t.Run("invalid endpoints not in map", func(t *testing.T) {
		invalidEndpointsList := []string{
			"register",
			"signin",
			"forgot_password",
			"reset_password",
			"verify_email",
			"",
			"all",
			"SIGNUP", // case sensitive
		}

		for _, endpoint := range invalidEndpointsList {
			assert.False(t, validEndpoints[endpoint], "Expected %q to be invalid", endpoint)
		}
	})
}

// =============================================================================
// UpdateSettings Handler Validation Tests
// =============================================================================

func TestUpdateCaptchaSettings_Validation(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCaptchaSettingsHandler(nil, nil, nil, nil, nil)

		app.Put("/settings/captcha", handler.UpdateSettings)

		req := httptest.NewRequest(http.MethodPut, "/settings/captcha", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(body, &result)
		assert.Contains(t, result["error"], "Invalid request body")
	})

	t.Run("invalid provider", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCaptchaSettingsHandler(nil, nil, nil, nil, nil)

		app.Put("/settings/captcha", handler.UpdateSettings)

		body := `{"provider": "invalid_provider"}`
		req := httptest.NewRequest(http.MethodPut, "/settings/captcha", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(respBody, &result)
		assert.Contains(t, result["error"], "Invalid provider")
		assert.Equal(t, "INVALID_INPUT", result["code"])
	})

	t.Run("invalid provider - recaptcha instead of recaptcha_v3", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCaptchaSettingsHandler(nil, nil, nil, nil, nil)

		app.Put("/settings/captcha", handler.UpdateSettings)

		body := `{"provider": "recaptcha"}`
		req := httptest.NewRequest(http.MethodPut, "/settings/captcha", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid endpoint", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCaptchaSettingsHandler(nil, nil, nil, nil, nil)

		app.Put("/settings/captcha", handler.UpdateSettings)

		body := `{"endpoints": ["signup", "invalid_endpoint"]}`
		req := httptest.NewRequest(http.MethodPut, "/settings/captcha", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(respBody, &result)
		assert.Contains(t, result["error"], "Invalid endpoint")
		assert.Equal(t, "INVALID_INPUT", result["code"])
	})

	t.Run("score threshold below 0", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCaptchaSettingsHandler(nil, nil, nil, nil, nil)

		app.Put("/settings/captcha", handler.UpdateSettings)

		body := `{"score_threshold": -0.1}`
		req := httptest.NewRequest(http.MethodPut, "/settings/captcha", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(respBody, &result)
		assert.Contains(t, result["error"], "Score threshold must be between 0.0 and 1.0")
		assert.Equal(t, "INVALID_INPUT", result["code"])
	})

	t.Run("score threshold above 1", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCaptchaSettingsHandler(nil, nil, nil, nil, nil)

		app.Put("/settings/captcha", handler.UpdateSettings)

		body := `{"score_threshold": 1.5}`
		req := httptest.NewRequest(http.MethodPut, "/settings/captcha", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(respBody, &result)
		assert.Contains(t, result["error"], "Score threshold must be between 0.0 and 1.0")
	})

	t.Run("valid score threshold at boundary 0", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCaptchaSettingsHandler(nil, nil, nil, nil, nil)

		app.Put("/settings/captcha", handler.UpdateSettings)

		// Score threshold of 0 should be valid - will fail at settings service level
		body := `{"score_threshold": 0.0}`
		req := httptest.NewRequest(http.MethodPut, "/settings/captcha", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Should pass validation (not StatusBadRequest)
		// Will fail later due to nil settingsService
		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("valid score threshold at boundary 1", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCaptchaSettingsHandler(nil, nil, nil, nil, nil)

		app.Put("/settings/captcha", handler.UpdateSettings)

		// Score threshold of 1 should be valid
		body := `{"score_threshold": 1.0}`
		req := httptest.NewRequest(http.MethodPut, "/settings/captcha", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Should pass validation
		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("all valid endpoints accepted", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCaptchaSettingsHandler(nil, nil, nil, nil, nil)

		app.Put("/settings/captcha", handler.UpdateSettings)

		body := `{"endpoints": ["signup", "login", "password_reset", "magic_link"]}`
		req := httptest.NewRequest(http.MethodPut, "/settings/captcha", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Should pass validation
		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("all valid providers accepted", func(t *testing.T) {
		providers := []string{"hcaptcha", "recaptcha_v3", "turnstile", "cap"}

		for _, provider := range providers {
			t.Run(provider, func(t *testing.T) {
				app := newTestApp(t)
				handler := NewCaptchaSettingsHandler(nil, nil, nil, nil, nil)

				app.Put("/settings/captcha", handler.UpdateSettings)

				body := `{"provider": "` + provider + `"}`
				req := httptest.NewRequest(http.MethodPut, "/settings/captcha", bytes.NewReader([]byte(body)))
				req.Header.Set("Content-Type", "application/json")

				resp, err := app.Test(req)
				require.NoError(t, err)
				defer func() { _ = resp.Body.Close() }()

				// Should pass validation
				assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
			})
		}
	})
}

// =============================================================================
// JSON Serialization Tests
// =============================================================================

func TestCaptchaRequests_JSONSerialization(t *testing.T) {
	t.Run("UpdateCaptchaSettingsRequest serializes correctly", func(t *testing.T) {
		enabled := true
		provider := "hcaptcha"
		siteKey := "test-site-key"
		secretKey := "test-secret-key"
		scoreThreshold := 0.5
		endpoints := []string{"signup", "login"}

		req := UpdateCaptchaSettingsRequest{
			Enabled:        &enabled,
			Provider:       &provider,
			SiteKey:        &siteKey,
			SecretKey:      &secretKey,
			ScoreThreshold: &scoreThreshold,
			Endpoints:      &endpoints,
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"enabled":true`)
		assert.Contains(t, string(data), `"provider":"hcaptcha"`)
		assert.Contains(t, string(data), `"site_key":"test-site-key"`)
		assert.Contains(t, string(data), `"secret_key":"test-secret-key"`)
		assert.Contains(t, string(data), `"score_threshold":0.5`)
		assert.Contains(t, string(data), `"endpoints"`)
	})

	t.Run("CaptchaSettingsResponse serializes correctly", func(t *testing.T) {
		resp := CaptchaSettingsResponse{
			Enabled:        true,
			Provider:       "turnstile",
			SiteKey:        "cf-site-key",
			SecretKeySet:   true,
			ScoreThreshold: 0.0, // Not used for turnstile
			Endpoints:      []string{"signup"},
			Overrides:      make(map[string]OverrideInfo),
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"enabled":true`)
		assert.Contains(t, string(data), `"provider":"turnstile"`)
		assert.Contains(t, string(data), `"site_key":"cf-site-key"`)
		assert.Contains(t, string(data), `"secret_key_set":true`)
	})
}

// =============================================================================
// Score Threshold Boundary Tests
// =============================================================================

func TestScoreThresholdBoundaries(t *testing.T) {
	testCases := []struct {
		name       string
		threshold  float64
		shouldFail bool
	}{
		{"valid 0.0", 0.0, false},
		{"valid 0.1", 0.1, false},
		{"valid 0.5", 0.5, false},
		{"valid 0.9", 0.9, false},
		{"valid 1.0", 1.0, false},
		{"invalid -0.1", -0.1, true},
		{"invalid -1.0", -1.0, true},
		{"invalid 1.1", 1.1, true},
		{"invalid 2.0", 2.0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := UpdateCaptchaSettingsRequest{
				ScoreThreshold: &tc.threshold,
			}

			// Validate using the same logic as the handler
			isInvalid := *req.ScoreThreshold < 0.0 || *req.ScoreThreshold > 1.0
			assert.Equal(t, tc.shouldFail, isInvalid, "Expected validation to be %v for threshold %v", tc.shouldFail, tc.threshold)
		})
	}
}

// =============================================================================
// Provider-Specific Settings Tests
// =============================================================================

func TestProviderSpecificSettings(t *testing.T) {
	t.Run("reCAPTCHA v3 settings", func(t *testing.T) {
		// reCAPTCHA v3 uses score threshold
		resp := CaptchaSettingsResponse{
			Provider:       "recaptcha_v3",
			SiteKey:        "6Lc...",
			SecretKeySet:   true,
			ScoreThreshold: 0.7, // Important for v3
		}

		assert.Equal(t, "recaptcha_v3", resp.Provider)
		assert.Equal(t, 0.7, resp.ScoreThreshold)
	})

	t.Run("hCaptcha settings", func(t *testing.T) {
		// hCaptcha doesn't use score threshold
		resp := CaptchaSettingsResponse{
			Provider:       "hcaptcha",
			SiteKey:        "10000000-ffff-ffff-ffff-000000000001",
			SecretKeySet:   true,
			ScoreThreshold: 0.0, // Not applicable
		}

		assert.Equal(t, "hcaptcha", resp.Provider)
	})

	t.Run("Turnstile settings", func(t *testing.T) {
		// Cloudflare Turnstile
		resp := CaptchaSettingsResponse{
			Provider:     "turnstile",
			SiteKey:      "0x...",
			SecretKeySet: true,
		}

		assert.Equal(t, "turnstile", resp.Provider)
	})

	t.Run("Cap settings", func(t *testing.T) {
		// Cap provider with custom server
		resp := CaptchaSettingsResponse{
			Provider:     "cap",
			CapServerURL: "https://cap.example.com",
			CapAPIKeySet: true,
		}

		assert.Equal(t, "cap", resp.Provider)
		assert.Equal(t, "https://cap.example.com", resp.CapServerURL)
		assert.True(t, resp.CapAPIKeySet)
	})
}

// =============================================================================
// Endpoint Configuration Tests
// =============================================================================

func TestEndpointConfiguration(t *testing.T) {
	t.Run("all endpoints enabled", func(t *testing.T) {
		resp := CaptchaSettingsResponse{
			Endpoints: []string{"signup", "login", "password_reset", "magic_link"},
		}

		assert.Len(t, resp.Endpoints, 4)
	})

	t.Run("signup only", func(t *testing.T) {
		resp := CaptchaSettingsResponse{
			Endpoints: []string{"signup"},
		}

		assert.Len(t, resp.Endpoints, 1)
		assert.Contains(t, resp.Endpoints, "signup")
	})

	t.Run("empty endpoints", func(t *testing.T) {
		resp := CaptchaSettingsResponse{
			Endpoints: []string{},
		}

		assert.Empty(t, resp.Endpoints)
	})

	t.Run("nil endpoints", func(t *testing.T) {
		resp := CaptchaSettingsResponse{
			Endpoints: nil,
		}

		assert.Nil(t, resp.Endpoints)
	})
}

// =============================================================================
// Override Information Tests
// =============================================================================

func TestOverrideInformation(t *testing.T) {
	t.Run("no overrides", func(t *testing.T) {
		resp := CaptchaSettingsResponse{
			Overrides: make(map[string]OverrideInfo),
		}

		assert.Empty(t, resp.Overrides)
	})

	t.Run("single override", func(t *testing.T) {
		resp := CaptchaSettingsResponse{
			Overrides: map[string]OverrideInfo{
				"enabled": {
					IsOverridden: true,
					EnvVar:       "FLUXBASE_SECURITY_CAPTCHA_ENABLED",
				},
			},
		}

		assert.Len(t, resp.Overrides, 1)
		assert.True(t, resp.Overrides["enabled"].IsOverridden)
		assert.Equal(t, "FLUXBASE_SECURITY_CAPTCHA_ENABLED", resp.Overrides["enabled"].EnvVar)
	})

	t.Run("multiple overrides", func(t *testing.T) {
		resp := CaptchaSettingsResponse{
			Overrides: map[string]OverrideInfo{
				"enabled": {
					IsOverridden: true,
					EnvVar:       "FLUXBASE_SECURITY_CAPTCHA_ENABLED",
				},
				"provider": {
					IsOverridden: true,
					EnvVar:       "FLUXBASE_SECURITY_CAPTCHA_PROVIDER",
				},
				"secret_key": {
					IsOverridden: true,
					EnvVar:       "FLUXBASE_SECURITY_CAPTCHA_SECRET_KEY",
				},
			},
		}

		assert.Len(t, resp.Overrides, 3)
		assert.True(t, resp.Overrides["enabled"].IsOverridden)
		assert.True(t, resp.Overrides["provider"].IsOverridden)
		assert.True(t, resp.Overrides["secret_key"].IsOverridden)
	})
}
