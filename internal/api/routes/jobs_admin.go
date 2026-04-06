package routes

import (
	"github.com/gofiber/fiber/v3"
)

// JobsAdminDeps contains dependencies for jobs admin routes.
// Auth middleware is inherited from the parent admin route group.
//
// Role Access:
//   - instance_admin: Full access to all job management operations
//   - tenant_admin: View/cancel/retry jobs for their tenant (RLS enforced)
type JobsAdminDeps struct {
	ListJobNamespaces fiber.Handler
	ListJobFunctions  fiber.Handler
	GetJobFunction    fiber.Handler
	DeleteJobFunction fiber.Handler
	GetJobStats       fiber.Handler
	ListWorkers       fiber.Handler
	ListAllJobs       fiber.Handler
	GetJobAdmin       fiber.Handler
	TerminateJob      fiber.Handler
	CancelJobAdmin    fiber.Handler
	RetryJobAdmin     fiber.Handler
	ResubmitJobAdmin  fiber.Handler
	SyncJobs          fiber.Handler
}

// BuildJobsAdminRoutes creates the jobs admin route group.
func BuildJobsAdminRoutes(deps *JobsAdminDeps) *RouteGroup {
	if deps == nil {
		return nil
	}

	return &RouteGroup{
		Name:         "jobs_admin",
		DefaultAuth:  AuthRequired,
		DefaultRoles: []string{"admin", "instance_admin"},
		Routes: []Route{
			// Job management (instance_admin only - uses default roles)
			{Method: "GET", Path: "/jobs/namespaces", Handler: deps.ListJobNamespaces, Summary: "List job namespaces"},
			{Method: "GET", Path: "/jobs/functions", Handler: deps.ListJobFunctions, Summary: "List job functions"},
			{Method: "GET", Path: "/jobs/functions/:namespace/:name", Handler: deps.GetJobFunction, Summary: "Get job function"},
			{Method: "DELETE", Path: "/jobs/functions/:namespace/:name", Handler: deps.DeleteJobFunction, Summary: "Delete job function"},
			{Method: "GET", Path: "/jobs/stats", Handler: deps.GetJobStats, Summary: "Get job stats"},
			{Method: "GET", Path: "/jobs/workers", Handler: deps.ListWorkers, Summary: "List workers"},
			{Method: "POST", Path: "/jobs/sync", Handler: deps.SyncJobs, Summary: "Sync jobs"},

			// Job queue (tenant_admin can access - override roles)
			{Method: "GET", Path: "/jobs", Handler: deps.ListAllJobs, Summary: "List all jobs", Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "GET", Path: "/jobs/:id", Handler: deps.GetJobAdmin, Summary: "Get job (admin)", Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "POST", Path: "/jobs/:id/terminate", Handler: deps.TerminateJob, Summary: "Terminate job", Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "POST", Path: "/jobs/:id/cancel", Handler: deps.CancelJobAdmin, Summary: "Cancel job (admin)", Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "POST", Path: "/jobs/:id/retry", Handler: deps.RetryJobAdmin, Summary: "Retry job (admin)", Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "POST", Path: "/jobs/:id/resubmit", Handler: deps.ResubmitJobAdmin, Summary: "Resubmit job", Roles: []string{"admin", "instance_admin", "tenant_admin"}},

			// Job queue route aliases (frontend compatibility, tenant_admin can access)
			{Method: "GET", Path: "/jobs/queue", Handler: deps.ListAllJobs, Summary: "List all jobs (queue alias)", Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "GET", Path: "/jobs/queue/:id", Handler: deps.GetJobAdmin, Summary: "Get job (queue alias)", Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "POST", Path: "/jobs/queue/:id/cancel", Handler: deps.CancelJobAdmin, Summary: "Cancel job (queue alias)", Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "POST", Path: "/jobs/queue/:id/terminate", Handler: deps.TerminateJob, Summary: "Terminate job (queue alias)", Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "POST", Path: "/jobs/queue/:id/retry", Handler: deps.RetryJobAdmin, Summary: "Retry job (queue alias)", Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "POST", Path: "/jobs/queue/:id/resubmit", Handler: deps.ResubmitJobAdmin, Summary: "Resubmit job (queue alias)", Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		},
	}
}
