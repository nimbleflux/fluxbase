package routes

import (
	"github.com/gofiber/fiber/v3"
)

type JobsDeps struct {
	RequireJobsEnabled fiber.Handler
	RequireAuth        fiber.Handler
	RequireScope       func(...string) fiber.Handler
	SubmitJob          fiber.Handler
	GetJob             fiber.Handler
	ListJobs           fiber.Handler
	CancelJob          fiber.Handler
	RetryJob           fiber.Handler
	GetJobLogsUser     fiber.Handler

	// Middleware for tenant context
	TenantMiddleware   fiber.Handler
	TenantDBMiddleware fiber.Handler
}

func BuildJobsRoutes(deps *JobsDeps) *RouteGroup {
	if deps == nil {
		return nil
	}
	// Build middlewares for tenant context
	var middlewares []Middleware
	middlewares = append(middlewares, Middleware{
		Name: "RequireJobsEnabled", Handler: deps.RequireJobsEnabled,
	})
	if deps.TenantMiddleware != nil {
		middlewares = append(middlewares, Middleware{
			Name: "TenantContext", Handler: deps.TenantMiddleware,
		})
	}
	if deps.TenantDBMiddleware != nil {
		middlewares = append(middlewares, Middleware{
			Name: "TenantDBContext", Handler: deps.TenantDBMiddleware,
		})
	}

	return &RouteGroup{
		Name:        "jobs",
		Middlewares: middlewares,
		Routes: []Route{
			{
				Method:  "POST",
				Path:    "/api/v1/jobs/submit",
				Handler: deps.SubmitJob,
				Summary: "Submit a new job",
				Auth:    AuthRequired,
			},
			{
				Method:  "GET",
				Path:    "/api/v1/jobs/:id/logs",
				Handler: deps.GetJobLogsUser,
				Summary: "Get job logs",
				Auth:    AuthRequired,
			},
			{
				Method:  "GET",
				Path:    "/api/v1/jobs/:id",
				Handler: deps.GetJob,
				Summary: "Get a job by ID",
				Auth:    AuthRequired,
			},
			{
				Method:  "GET",
				Path:    "/api/v1/jobs",
				Handler: deps.ListJobs,
				Summary: "List jobs",
				Auth:    AuthRequired,
			},
			{
				Method:  "POST",
				Path:    "/api/v1/jobs/:id/cancel",
				Handler: deps.CancelJob,
				Summary: "Cancel a job",
				Auth:    AuthRequired,
			},
			{
				Method:  "POST",
				Path:    "/api/v1/jobs/:id/retry",
				Handler: deps.RetryJob,
				Summary: "Retry a job",
				Auth:    AuthRequired,
			},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Required: deps.RequireAuth,
		},
	}
}
