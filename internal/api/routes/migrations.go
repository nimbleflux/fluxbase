package routes

import "github.com/gofiber/fiber/v3"

type MigrationsDeps struct {
	SecurityMiddleware fiber.Handler
	CreateMigration    fiber.Handler
	ListMigrations     fiber.Handler
	GetMigration       fiber.Handler
	UpdateMigration    fiber.Handler
	DeleteMigration    fiber.Handler
	ApplyMigration     fiber.Handler
	RollbackMigration  fiber.Handler
	ApplyPending       fiber.Handler
	SyncMigrations     fiber.Handler
	GetExecutions      fiber.Handler
}

func BuildMigrationsRoutes(deps *MigrationsDeps) *RouteGroup {
	if deps == nil {
		return nil
	}
	return &RouteGroup{
		Name:   "migrations",
		Prefix: "/api/v1/admin/migrations",
		Routes: []Route{
			{Method: "POST", Path: "/", Handler: deps.CreateMigration, Summary: "Create migration", Auth: AuthRequired},
			{Method: "GET", Path: "/", Handler: deps.ListMigrations, Summary: "List migrations", Auth: AuthRequired},
			{Method: "GET", Path: "/:name", Handler: deps.GetMigration, Summary: "Get migration", Auth: AuthRequired},
			{Method: "PUT", Path: "/:name", Handler: deps.UpdateMigration, Summary: "Update migration", Auth: AuthRequired},
			{Method: "DELETE", Path: "/:name", Handler: deps.DeleteMigration, Summary: "Delete migration", Auth: AuthRequired},
			{Method: "POST", Path: "/:name/apply", Handler: deps.ApplyMigration, Summary: "Apply migration", Auth: AuthRequired},
			{Method: "POST", Path: "/:name/rollback", Handler: deps.RollbackMigration, Summary: "Rollback migration", Auth: AuthRequired},
			{Method: "POST", Path: "/apply-pending", Handler: deps.ApplyPending, Summary: "Apply pending migrations", Auth: AuthRequired},
			{Method: "POST", Path: "/sync", Handler: deps.SyncMigrations, Summary: "Sync migrations", Auth: AuthRequired},
			{Method: "GET", Path: "/:name/executions", Handler: deps.GetExecutions, Summary: "Get execution history", Auth: AuthRequired},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Required: deps.SecurityMiddleware,
		},
	}
}
