package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/nimbleflux/fluxbase/internal/auth"
)

type RESTDeps struct {
	RequireAuth  fiber.Handler
	RequireScope func(...string) fiber.Handler
	HandleTables fiber.Handler
	HandleQuery  fiber.Handler
	HandleById   fiber.Handler
}

func BuildRESTRoutes(deps *RESTDeps) *RouteGroup {
	readScope := []Middleware{{Name: "RequireScope(tables:read)", Handler: deps.RequireScope(auth.ScopeTablesRead)}}
	writeScope := []Middleware{{Name: "RequireScope(tables:write)", Handler: deps.RequireScope(auth.ScopeTablesWrite)}}

	return &RouteGroup{
		Name:   "rest",
		Prefix: "/api/v1/tables",
		Middlewares: []Middleware{
			{Name: "RequireAuth", Handler: deps.RequireAuth},
		},
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/",
				Handler: deps.HandleTables,
				Summary: "List all tables (admin only)",
				Auth:    AuthRequired,
			},
			{
				Method:      "POST",
				Path:        "/:schema/:table/query",
				Handler:     deps.HandleQuery,
				Summary:     "Query table with complex filters",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeTablesRead},
				Middlewares: readScope,
			},
			{
				Method:      "POST",
				Path:        "/:schema/query",
				Handler:     deps.HandleQuery,
				Summary:     "Query table in public schema",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeTablesRead},
				Middlewares: readScope,
			},
			{
				Method:      "GET",
				Path:        "/:schema/:table/:id",
				Handler:     deps.HandleById,
				Summary:     "Get single row by ID",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeTablesRead},
				Middlewares: readScope,
			},
			{
				Method:      "PUT",
				Path:        "/:schema/:table/:id",
				Handler:     deps.HandleById,
				Summary:     "Replace row by ID",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeTablesWrite},
				Middlewares: writeScope,
			},
			{
				Method:      "PATCH",
				Path:        "/:schema/:table/:id",
				Handler:     deps.HandleById,
				Summary:     "Update row by ID",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeTablesWrite},
				Middlewares: writeScope,
			},
			{
				Method:      "DELETE",
				Path:        "/:schema/:table/:id",
				Handler:     deps.HandleById,
				Summary:     "Delete row by ID",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeTablesWrite},
				Middlewares: writeScope,
			},
			{
				Method:      "GET",
				Path:        "/:schema/:table",
				Handler:     deps.HandleTables,
				Summary:     "List rows from table",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeTablesRead},
				Middlewares: readScope,
			},
			{
				Method:      "POST",
				Path:        "/:schema/:table",
				Handler:     deps.HandleTables,
				Summary:     "Create row in table",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeTablesWrite},
				Middlewares: writeScope,
			},
			{
				Method:      "PATCH",
				Path:        "/:schema/:table",
				Handler:     deps.HandleTables,
				Summary:     "Batch update rows",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeTablesWrite},
				Middlewares: writeScope,
			},
			{
				Method:      "DELETE",
				Path:        "/:schema/:table",
				Handler:     deps.HandleTables,
				Summary:     "Batch delete rows",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeTablesWrite},
				Middlewares: writeScope,
			},
			{
				Method:      "GET",
				Path:        "/:schema",
				Handler:     deps.HandleTables,
				Summary:     "List rows from public schema table",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeTablesRead},
				Middlewares: readScope,
			},
			{
				Method:      "POST",
				Path:        "/:schema",
				Handler:     deps.HandleTables,
				Summary:     "Create row in public schema table",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeTablesWrite},
				Middlewares: writeScope,
			},
			{
				Method:      "PATCH",
				Path:        "/:schema",
				Handler:     deps.HandleTables,
				Summary:     "Batch update public schema rows",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeTablesWrite},
				Middlewares: writeScope,
			},
			{
				Method:      "DELETE",
				Path:        "/:schema",
				Handler:     deps.HandleTables,
				Summary:     "Batch delete public schema rows",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeTablesWrite},
				Middlewares: writeScope,
			},
		},
	}
}
