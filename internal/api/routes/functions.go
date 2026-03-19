package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/nimbleflux/fluxbase/internal/auth"
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
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/api/v1/functions",
				Handler: deps.ListFunctions,
				Middlewares: []Middleware{
					{Name: "RequireFunctionsEnabled", Handler: deps.RequireFunctionsEnabled},
					{Name: "RequireAuth", Handler: deps.RequireAuth},
					{Name: "RequireScope(functions:read)", Handler: deps.RequireScope(auth.ScopeFunctionsRead)},
				},
				Summary: "List all functions",
				Auth:    AuthRequired,
				Scopes:  []string{auth.ScopeFunctionsRead},
			},
			{
				Method:  "GET",
				Path:    "/api/v1/functions/:name",
				Handler: deps.GetFunction,
				Middlewares: []Middleware{
					{Name: "RequireFunctionsEnabled", Handler: deps.RequireFunctionsEnabled},
					{Name: "RequireAuth", Handler: deps.RequireAuth},
					{Name: "RequireScope(functions:read)", Handler: deps.RequireScope(auth.ScopeFunctionsRead)},
				},
				Summary: "Get a function by name",
				Auth:    AuthRequired,
				Scopes:  []string{auth.ScopeFunctionsRead},
			},
			{
				Method:  "POST",
				Path:    "/api/v1/functions",
				Handler: deps.CreateFunction,
				Middlewares: []Middleware{
					{Name: "RequireFunctionsEnabled", Handler: deps.RequireFunctionsEnabled},
					{Name: "RequireAuth", Handler: deps.RequireAuth},
					{Name: "RequireScope(functions:execute)", Handler: deps.RequireScope(auth.ScopeFunctionsExecute)},
				},
				Summary: "Create a new function",
				Auth:    AuthRequired,
				Scopes:  []string{auth.ScopeFunctionsExecute},
			},
			{
				Method:  "PUT",
				Path:    "/api/v1/functions/:name",
				Handler: deps.UpdateFunction,
				Middlewares: []Middleware{
					{Name: "RequireFunctionsEnabled", Handler: deps.RequireFunctionsEnabled},
					{Name: "RequireAuth", Handler: deps.RequireAuth},
					{Name: "RequireScope(functions:execute)", Handler: deps.RequireScope(auth.ScopeFunctionsExecute)},
				},
				Summary: "Update a function",
				Auth:    AuthRequired,
				Scopes:  []string{auth.ScopeFunctionsExecute},
			},
			{
				Method:  "DELETE",
				Path:    "/api/v1/functions/:name",
				Handler: deps.DeleteFunction,
				Middlewares: []Middleware{
					{Name: "RequireFunctionsEnabled", Handler: deps.RequireFunctionsEnabled},
					{Name: "RequireAuth", Handler: deps.RequireAuth},
					{Name: "RequireScope(functions:execute)", Handler: deps.RequireScope(auth.ScopeFunctionsExecute)},
				},
				Summary: "Delete a function",
				Auth:    AuthRequired,
				Scopes:  []string{auth.ScopeFunctionsExecute},
			},
			{
				Method:  "POST",
				Path:    "/api/v1/functions/:name/invoke",
				Handler: deps.InvokeFunction,
				Middlewares: []Middleware{
					{Name: "RequireFunctionsEnabled", Handler: deps.RequireFunctionsEnabled},
					{Name: "OptionalAuth", Handler: deps.OptionalAuth},
					{Name: "RequireScope(functions:execute)", Handler: deps.RequireScope(auth.ScopeFunctionsExecute)},
				},
				Summary: "Invoke a function",
				Auth:    AuthOptional,
				Scopes:  []string{auth.ScopeFunctionsExecute},
			},
			{
				Method:  "GET",
				Path:    "/api/v1/functions/:name/invoke",
				Handler: deps.InvokeFunction,
				Middlewares: []Middleware{
					{Name: "RequireFunctionsEnabled", Handler: deps.RequireFunctionsEnabled},
					{Name: "OptionalAuth", Handler: deps.OptionalAuth},
					{Name: "RequireScope(functions:execute)", Handler: deps.RequireScope(auth.ScopeFunctionsExecute)},
				},
				Summary: "Invoke a function (GET for health checks)",
				Auth:    AuthOptional,
				Scopes:  []string{auth.ScopeFunctionsExecute},
			},
			{
				Method:  "GET",
				Path:    "/api/v1/functions/:name/executions",
				Handler: deps.GetExecutions,
				Middlewares: []Middleware{
					{Name: "RequireFunctionsEnabled", Handler: deps.RequireFunctionsEnabled},
					{Name: "RequireAuth", Handler: deps.RequireAuth},
					{Name: "RequireScope(functions:read)", Handler: deps.RequireScope(auth.ScopeFunctionsRead)},
				},
				Summary: "Get function execution history",
				Auth:    AuthRequired,
				Scopes:  []string{auth.ScopeFunctionsRead},
			},
			{
				Method:  "GET",
				Path:    "/api/v1/functions/shared",
				Handler: deps.ListSharedModules,
				Middlewares: []Middleware{
					{Name: "RequireFunctionsEnabled", Handler: deps.RequireFunctionsEnabled},
					{Name: "RequireAuth", Handler: deps.RequireAuth},
					{Name: "RequireScope(functions:read)", Handler: deps.RequireScope(auth.ScopeFunctionsRead)},
				},
				Summary: "List shared modules",
				Auth:    AuthRequired,
				Scopes:  []string{auth.ScopeFunctionsRead},
			},
			{
				Method:  "GET",
				Path:    "/api/v1/functions/shared/*",
				Handler: deps.GetSharedModule,
				Middlewares: []Middleware{
					{Name: "RequireFunctionsEnabled", Handler: deps.RequireFunctionsEnabled},
					{Name: "RequireAuth", Handler: deps.RequireAuth},
					{Name: "RequireScope(functions:read)", Handler: deps.RequireScope(auth.ScopeFunctionsRead)},
				},
				Summary: "Get a shared module",
				Auth:    AuthRequired,
				Scopes:  []string{auth.ScopeFunctionsRead},
			},
			{
				Method:  "POST",
				Path:    "/api/v1/functions/shared",
				Handler: deps.CreateSharedModule,
				Middlewares: []Middleware{
					{Name: "RequireFunctionsEnabled", Handler: deps.RequireFunctionsEnabled},
					{Name: "RequireAuth", Handler: deps.RequireAuth},
					{Name: "RequireScope(functions:execute)", Handler: deps.RequireScope(auth.ScopeFunctionsExecute)},
				},
				Summary: "Create a shared module",
				Auth:    AuthRequired,
				Scopes:  []string{auth.ScopeFunctionsExecute},
			},
			{
				Method:  "PUT",
				Path:    "/api/v1/functions/shared/*",
				Handler: deps.UpdateSharedModule,
				Middlewares: []Middleware{
					{Name: "RequireFunctionsEnabled", Handler: deps.RequireFunctionsEnabled},
					{Name: "RequireAuth", Handler: deps.RequireAuth},
					{Name: "RequireScope(functions:execute)", Handler: deps.RequireScope(auth.ScopeFunctionsExecute)},
				},
				Summary: "Update a shared module",
				Auth:    AuthRequired,
				Scopes:  []string{auth.ScopeFunctionsExecute},
			},
			{
				Method:  "DELETE",
				Path:    "/api/v1/functions/shared/*",
				Handler: deps.DeleteSharedModule,
				Middlewares: []Middleware{
					{Name: "RequireFunctionsEnabled", Handler: deps.RequireFunctionsEnabled},
					{Name: "RequireAuth", Handler: deps.RequireAuth},
					{Name: "RequireScope(functions:execute)", Handler: deps.RequireScope(auth.ScopeFunctionsExecute)},
				},
				Summary: "Delete a shared module",
				Auth:    AuthRequired,
				Scopes:  []string{auth.ScopeFunctionsExecute},
			},
		},
	}
}
