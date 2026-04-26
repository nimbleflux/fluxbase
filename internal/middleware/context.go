package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func GetUserID(c fiber.Ctx) string {
	userID, _ := c.Locals("user_id").(string)
	return userID
}

func GetUserIDUUID(c fiber.Ctx) (*uuid.UUID, error) {
	s := GetUserID(c)
	if s == "" {
		return nil, nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func GetUserRole(c fiber.Ctx) string {
	role, _ := c.Locals("user_role").(string)
	return role
}

func GetAuthType(c fiber.Ctx) string {
	authType, _ := c.Locals("auth_type").(string)
	return authType
}

func GetClientKeyID(c fiber.Ctx) string {
	id, _ := c.Locals("client_key_id").(string)
	return id
}

func GetServiceKeyID(c fiber.Ctx) string {
	id, _ := c.Locals("service_key_id").(string)
	return id
}

func GetTenantSlug(c fiber.Ctx) string {
	slug, _ := c.Locals("tenant_slug").(string)
	return slug
}

func GetTenantID(c fiber.Ctx) string {
	id, _ := c.Locals("tenant_id").(string)
	return id
}

func GetRLSUserID(c fiber.Ctx) string {
	userID, _ := c.Locals("rls_user_id").(string)
	return userID
}

func GetRLSRole(c fiber.Ctx) string {
	role, _ := c.Locals("rls_role").(string)
	return role
}

func GetNamespace(c fiber.Ctx) string {
	namespace := c.Query("namespace")
	if namespace == "" {
		return "default"
	}
	return namespace
}
