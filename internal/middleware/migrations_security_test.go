package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/config"
)

func TestMigrationsSecurity_FeatureDisabled(t *testing.T) {
	cfg := &config.MigrationsConfig{
		Enabled: false,
	}

	app := fiber.New()
	app.Use(RequireMigrationsFullSecurity(
		cfg,
		&config.ServerConfig{},
		nil,
		nil,
		100,
		time.Minute,
		nil,
	))
	app.Post("/sync", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(fiber.MethodPost, "/sync", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var body map[string]string
	require.NoError(t, json.Unmarshal(getBody(t, resp), &body))
	assert.Equal(t, "Not Found", body["error"])
}

func TestMigrationsSecurity_NoAuthReturns401(t *testing.T) {
	cfg := &config.MigrationsConfig{
		Enabled: true,
	}

	app := fiber.New()
	app.Use(RequireMigrationsFullSecurity(
		cfg,
		&config.ServerConfig{},
		nil,
		nil,
		100,
		time.Minute,
		nil,
	))
	app.Post("/sync", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(fiber.MethodPost, "/sync", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	var body map[string]string
	require.NoError(t, json.Unmarshal(getBody(t, resp), &body))
	assert.Contains(t, body["error"], "Service key or service_role JWT")
}

func getBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return body
}
