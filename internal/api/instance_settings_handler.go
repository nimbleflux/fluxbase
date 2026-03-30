package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/settings"
)

// InstanceSettingsHandler handles instance-level settings API endpoints
type InstanceSettingsHandler struct {
	settingsSvc *settings.UnifiedService
}

// NewInstanceSettingsHandler creates a new instance settings handler
func NewInstanceSettingsHandler(settingsSvc *settings.UnifiedService) *InstanceSettingsHandler {
	return &InstanceSettingsHandler{
		settingsSvc: settingsSvc,
	}
}

// InstanceSettingsResponse represents the response for instance settings
type InstanceSettingsResponse struct {
	Settings            map[string]any `json:"settings"`
	OverridableSettings []string       `json:"overridable_settings,omitempty"`
}

// UpdateInstanceSettingsRequest represents the request to update instance settings
type UpdateInstanceSettingsRequest struct {
	Settings map[string]any `json:"settings"`
}

// UpdateOverridableSettingsRequest represents the request to update overridable settings
type UpdateOverridableSettingsRequest struct {
	OverridableSettings []string `json:"overridable_settings"`
}

// GetInstanceSettings returns all instance-level settings
// GET /admin/instance/settings
func (h *InstanceSettingsHandler) GetInstanceSettings(c fiber.Ctx) error {
	ctx := c.Context()

	instanceSettings, err := h.settingsSvc.GetInstanceSettings(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get instance settings")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get instance settings")
	}

	return c.JSON(InstanceSettingsResponse{
		Settings:            instanceSettings.Settings,
		OverridableSettings: instanceSettings.OverridableSettings,
	})
}

// UpdateInstanceSettings updates instance-level settings
// PATCH /admin/instance/settings
func (h *InstanceSettingsHandler) UpdateInstanceSettings(c fiber.Ctx) error {
	ctx := c.Context()

	var req UpdateInstanceSettingsRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.Settings == nil {
		return fiber.NewError(fiber.StatusBadRequest, "Settings are required")
	}

	// Update each setting
	for path, value := range req.Settings {
		isSecret := false
		if err := h.settingsSvc.SetInstanceSetting(ctx, path, value, isSecret); err != nil {
			log.Error().Err(err).Str("path", path).Msg("Failed to set instance setting")
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to update instance settings")
		}
	}

	// Return updated settings
	return h.GetInstanceSettings(c)
}

// GetOverridableSettings returns which settings can be overridden by tenants
// GET /admin/instance/settings/overridable
func (h *InstanceSettingsHandler) GetOverridableSettings(c fiber.Ctx) error {
	ctx := c.Context()

	instanceSettings, err := h.settingsSvc.GetInstanceSettings(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get instance settings")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get overridable settings")
	}

	return c.JSON(fiber.Map{
		"overridable_settings": instanceSettings.OverridableSettings,
	})
}

// UpdateOverridableSettings updates which settings can be overridden by tenants
// PUT /admin/instance/settings/overridable
func (h *InstanceSettingsHandler) UpdateOverridableSettings(c fiber.Ctx) error {
	ctx := c.Context()

	var req UpdateOverridableSettingsRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.settingsSvc.SetOverridableSettings(ctx, req.OverridableSettings); err != nil {
		log.Error().Err(err).Msg("Failed to update overridable settings")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update overridable settings")
	}

	return c.JSON(fiber.Map{
		"overridable_settings": req.OverridableSettings,
	})
}
