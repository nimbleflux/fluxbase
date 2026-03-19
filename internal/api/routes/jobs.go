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
}

func BuildJobsRoutes(deps *JobsDeps) *RouteGroup {
	if deps == nil {
		return nil
	}
	return &RouteGroup{
		Name: "jobs",
		Routes: []Route{
			{
				Method:  "POST",
				Path:    "/api/v1/jobs/submit",
				Handler: deps.SubmitJob,
				Middlewares: []Middleware{
					{Name: "RequireJobsEnabled", Handler: deps.RequireJobsEnabled},
					{Name: "RequireAuth", Handler: deps.RequireAuth},
				},
				Summary: "Submit a new job",
				Auth:    AuthRequired,
			},
			{
				Method:  "GET",
				Path:    "/api/v1/jobs/:id/logs",
				Handler: deps.GetJobLogsUser,
				Middlewares: []Middleware{
					{Name: "RequireJobsEnabled", Handler: deps.RequireJobsEnabled},
					{Name: "RequireAuth", Handler: deps.RequireAuth},
				},
				Summary: "Get job logs",
				Auth:    AuthRequired,
			},
			{
				Method:  "GET",
				Path:    "/api/v1/jobs/:id",
				Handler: deps.GetJob,
				Middlewares: []Middleware{
					{Name: "RequireJobsEnabled", Handler: deps.RequireJobsEnabled},
					{Name: "RequireAuth", Handler: deps.RequireAuth},
				},
				Summary: "Get a job by ID",
				Auth:    AuthRequired,
			},
			{
				Method:  "GET",
				Path:    "/api/v1/jobs",
				Handler: deps.ListJobs,
				Middlewares: []Middleware{
					{Name: "RequireJobsEnabled", Handler: deps.RequireJobsEnabled},
					{Name: "RequireAuth", Handler: deps.RequireAuth},
				},
				Summary: "List jobs",
				Auth:    AuthRequired,
			},
			{
				Method:  "POST",
				Path:    "/api/v1/jobs/:id/cancel",
				Handler: deps.CancelJob,
				Middlewares: []Middleware{
					{Name: "RequireJobsEnabled", Handler: deps.RequireJobsEnabled},
					{Name: "RequireAuth", Handler: deps.RequireAuth},
				},
				Summary: "Cancel a job",
				Auth:    AuthRequired,
			},
			{
				Method:  "POST",
				Path:    "/api/v1/jobs/:id/retry",
				Handler: deps.RetryJob,
				Middlewares: []Middleware{
					{Name: "RequireJobsEnabled", Handler: deps.RequireJobsEnabled},
					{Name: "RequireAuth", Handler: deps.RequireAuth},
				},
				Summary: "Retry a job",
				Auth:    AuthRequired,
			},
		},
	}
}
