package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/settings"
)

type InstanceSettingsHandler struct {
	settingsSvc *settings.UnifiedService
}

func NewInstanceSettingsHandler(settingsSvc *settings.UnifiedService) *InstanceSettingsHandler {
	return &InstanceSettingsHandler{
		settingsSvc: settingsSvc,
	}
}

type InstanceSettingsResponse struct {
	Settings            map[string]any `json:"settings"`
	OverridableSettings []string       `json:"overridable_settings,omitempty"`
}

type UpdateInstanceSettingsRequest struct {
	Settings map[string]any `json:"settings"`
}

type UpdateOverridableSettingsRequest struct {
	OverridableSettings []string `json:"overridable_settings"`
}

func (h *InstanceSettingsHandler) GetInstanceSettings(c fiber.Ctx) error {
	ctx := c.Context()

	instanceSettings, err := h.settingsSvc.GetInstanceSettings(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get instance settings")
		return SendInternalError(c, "Failed to get instance settings")
	}

	return c.JSON(InstanceSettingsResponse{
		Settings:            instanceSettings.Settings,
		OverridableSettings: instanceSettings.OverridableSettings,
	})
}

func (h *InstanceSettingsHandler) UpdateInstanceSettings(c fiber.Ctx) error {
	ctx := c.Context()

	var req UpdateInstanceSettingsRequest
	if err := ParseBody(c, &req); err != nil {
		return err
	}

	if req.Settings == nil {
		return SendBadRequest(c, "Settings are required", ErrCodeInvalidInput)
	}

	for path, value := range req.Settings {
		isSecret := false
		if err := h.settingsSvc.SetInstanceSetting(ctx, path, value, isSecret); err != nil {
			log.Error().Err(err).Str("path", path).Msg("Failed to set instance setting")
			return SendInternalError(c, "Failed to update instance settings")
		}
	}

	return h.GetInstanceSettings(c)
}

func (h *InstanceSettingsHandler) GetOverridableSettings(c fiber.Ctx) error {
	ctx := c.Context()

	instanceSettings, err := h.settingsSvc.GetInstanceSettings(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get instance settings")
		return SendInternalError(c, "Failed to get overridable settings")
	}

	return c.JSON(fiber.Map{
		"overridable_settings": instanceSettings.OverridableSettings,
	})
}

func (h *InstanceSettingsHandler) UpdateOverridableSettings(c fiber.Ctx) error {
	ctx := c.Context()

	var req UpdateOverridableSettingsRequest
	if err := ParseBody(c, &req); err != nil {
		return err
	}

	if err := h.settingsSvc.SetOverridableSettings(ctx, req.OverridableSettings); err != nil {
		log.Error().Err(err).Msg("Failed to update overridable settings")
		return SendInternalError(c, "Failed to update overridable settings")
	}

	return c.JSON(fiber.Map{
		"overridable_settings": req.OverridableSettings,
	})
}
