package routes

import (
	"github.com/gofiber/fiber/v3"
)

type InvitationDeps struct {
	ValidateInvitation fiber.Handler
	AcceptInvitation   fiber.Handler
}

func BuildInvitationRoutes(deps *InvitationDeps) *RouteGroup {
	return &RouteGroup{
		Name:   "invitations",
		Prefix: "/api/v1/invitations",
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/:token/validate",
				Handler: deps.ValidateInvitation,
				Summary: "Validate invitation token",
				Auth:    AuthNone,
				Public:  true,
			},
			{
				Method:  "POST",
				Path:    "/:token/accept",
				Handler: deps.AcceptInvitation,
				Summary: "Accept invitation",
				Auth:    AuthNone,
				Public:  true,
			},
		},
	}
}
