package routes

import (
	"github.com/gofiber/fiber/v3"
)

type FunctionsDeps struct {
	RequireFunctionsEnabled fiber.Handler
	RequireAuth             fiber.Handler
	OptionalAuth            fiber.Handler
	RequireScope            func(...string) fiber.Handler
	ListFunctions           fiber.Handler
	GetFunction             fiber.Handler
	CreateFunction          fiber.Handler
	UpdateFunction          fiber.Handler
	DeleteFunction          fiber.Handler
	InvokeFunction          fiber.Handler
	GetExecutions           fiber.Handler
	ListSharedModules       fiber.Handler
	GetSharedModule         fiber.Handler
	CreateSharedModule      fiber.Handler
	UpdateSharedModule      fiber.Handler
	DeleteSharedModule      fiber.Handler
}

func BuildFunctionsRoutes(deps *FunctionsDeps) *RouteGroup {
	return &RouteGroup{
		Name: "functions",
		Middlewares: []Middleware{
			{Name: "RequireFunctionsEnabled", Handler: deps.RequireFunctionsEnabled},
		},
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/api/v1/functions",
				Handler: deps.ListFunctions,
				Summary: "List all functions",
				Auth:    AuthRequired,
				Scopes:  []string{"functions:read"},
			},
			{
				Method:  "GET",
				Path:    "/api/v1/functions/:name",
				Handler: deps.GetFunction,
				Summary: "Get a function by name",
				Auth:    AuthRequired,
				Scopes:  []string{"functions:read"},
			},
			{
				Method:  "POST",
				Path:    "/api/v1/functions",
				Handler: deps.CreateFunction,
				Summary: "Create a new function",
				Auth:    AuthRequired,
				Scopes:  []string{"functions:execute"},
			},
			{
				Method:  "PUT",
				Path:    "/api/v1/functions/:name",
				Handler: deps.UpdateFunction,
				Summary: "Update a function",
				Auth:    AuthRequired,
				Scopes:  []string{"functions:execute"},
			},
			{
				Method:  "DELETE",
				Path:    "/api/v1/functions/:name",
				Handler: deps.DeleteFunction,
				Summary: "Delete a function",
				Auth:    AuthRequired,
				Scopes:  []string{"functions:execute"},
			},
			{
				Method:  "POST",
				Path:    "/api/v1/functions/:name/invoke",
				Handler: deps.InvokeFunction,
				Summary: "Invoke a function",
				Auth:    AuthOptional,
				Scopes:  []string{"functions:execute"},
			},
			{
				Method:  "GET",
				Path:    "/api/v1/functions/:name/invoke",
				Handler: deps.InvokeFunction,
				Summary: "Invoke a function (GET for health checks)",
				Auth:    AuthOptional,
				Scopes:  []string{"functions:execute"},
			},
			{
				Method:  "GET",
				Path:    "/api/v1/functions/:name/executions",
				Handler: deps.GetExecutions,
				Summary: "Get function execution history",
				Auth:    AuthRequired,
				Scopes:  []string{"functions:read"},
			},
			{
				Method:  "GET",
				Path:    "/api/v1/functions/shared",
				Handler: deps.ListSharedModules,
				Summary: "List shared modules",
				Auth:    AuthRequired,
				Scopes:  []string{"functions:read"},
			},
			{
				Method:  "GET",
				Path:    "/api/v1/functions/shared/*",
				Handler: deps.GetSharedModule,
				Summary: "Get a shared module",
				Auth:    AuthRequired,
				Scopes:  []string{"functions:read"},
			},
			{
				Method:  "POST",
				Path:    "/api/v1/functions/shared",
				Handler: deps.CreateSharedModule,
				Summary: "Create a shared module",
				Auth:    AuthRequired,
				Scopes:  []string{"functions:execute"},
			},
			{
				Method:  "PUT",
				Path:    "/api/v1/functions/shared/*",
				Handler: deps.UpdateSharedModule,
				Summary: "Update a shared module",
				Auth:    AuthRequired,
				Scopes:  []string{"functions:execute"},
			},
			{
				Method:  "DELETE",
				Path:    "/api/v1/functions/shared/*",
				Handler: deps.DeleteSharedModule,
				Summary: "Delete a shared module",
				Auth:    AuthRequired,
				Scopes:  []string{"functions:execute"},
			},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Required: deps.RequireAuth,
			Optional: deps.OptionalAuth,
		},
		RequireScope: deps.RequireScope,
	}
}
