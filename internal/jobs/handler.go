package jobs

import (
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/nimbleflux/fluxbase/internal/auth"
	"github.com/nimbleflux/fluxbase/internal/config"
	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/nimbleflux/fluxbase/internal/logging"
)

/*
JOB PROGRESS REPORTING API DOCUMENTATION

When writing job functions, use the global `Fluxbase` API to report progress:

1. REPORT PROGRESS

   Fluxbase.reportProgress(percent, message, data)

   Parameters:
   - percent (number): Progress percentage from 0 to 100
   - message (string): Human-readable status message
   - data (object, optional): Additional structured data

   Examples:

   // Simple percentage update
   Fluxbase.reportProgress(25, "Processing batch 1 of 4");

   // With absolute progress
   Fluxbase.reportProgress(50, "Processed 500 of 1000 records", {
     processed: 500,
     total: 1000,
     errors: 3
   });

   // Step-based progress
   Fluxbase.reportProgress(33, "Step 1: Validating data", {
     step: "validation",
     itemsValidated: 150
   });

2. GET JOB PAYLOAD

   const payload = Fluxbase.getJobPayload()

   Returns the job's input payload as an object.

3. CHECK CANCELLATION

   if (Fluxbase.checkCancellation()) {
     // Job was cancelled by user, clean up and exit
     return { success: false, error: "Job cancelled by user" };
   }

4. GET JOB CONTEXT

   const context = Fluxbase.getJobContext()

   Returns full job context including:
   - job_id: UUID of the job
   - job_name: Name of the job function
   - namespace: Job namespace
   - retry_count: Current retry attempt
   - user_id: User who submitted the job (if any)
   - payload: Job input data

BEST PRACTICES:

1. Report progress frequently (every 5-10% or after significant steps)
2. Use descriptive messages that help users understand what's happening
3. Include absolute progress in data field when applicable (e.g., "processed: 50, total: 100")
4. Check for cancellation in long-running loops
5. Return structured results:

   // Success
   return {
     success: true,
     result: { recordsProcessed: 1000, errors: 2 }
   };

   // Failure
   return {
     success: false,
     error: "Failed to connect to external API",
     result: { partialData: [...] }
   };

EXAMPLE JOB FUNCTION:

```typescript
// @fluxbase:timeout 600
// @fluxbase:max-retries 3
// @fluxbase:progress-timeout 60

export async function handler(request: Request) {
  const { items } = Fluxbase.getJobPayload();
  const total = items.length;
  let processed = 0;
  const results = [];

  Fluxbase.reportProgress(0, "Starting processing");

  for (const item of items) {
    // Check for cancellation
    if (Fluxbase.checkCancellation()) {
      return {
        success: false,
        error: "Job cancelled",
        result: { processed, results }
      };
    }

    // Process item
    const result = await processItem(item);
    results.push(result);
    processed++;

    // Report progress
    const percent = Math.round((processed / total) * 100);
    Fluxbase.reportProgress(percent, `Processed ${processed} of ${total}`, {
      processed,
      total,
      lastItem: item.id
    });
  }

  return {
    success: true,
    result: {
      totalProcessed: processed,
      results
    }
  };
}
```

ANNOTATIONS:

Configure job behavior using @fluxbase: annotations in code comments:

- @fluxbase:timeout 600               // Max duration in seconds (default: 300)
- @fluxbase:memory 512                // Memory limit in MB (default: 256)
- @fluxbase:max-retries 3             // Max retry attempts (default: 0)
- @fluxbase:progress-timeout 60       // Kill job if no progress for N seconds (default: 60)
- @fluxbase:enabled false             // Disable job function (default: true)
- @fluxbase:allow-read true           // Allow filesystem read (default: false)
- @fluxbase:allow-write true          // Allow filesystem write (default: false)
- @fluxbase:allow-net false           // Disallow network access (default: true)
- @fluxbase:allow-env false           // Disallow env var access (default: true)
- @fluxbase:schedule 0 2 * * *        // Cron schedule (optional)
- @fluxbase:schedule-params {"key": "value"}  // Parameters passed when scheduled (optional)

*/

// Handler manages HTTP endpoints for jobs
type Handler struct {
	storage        *Storage
	loader         *Loader
	manager        *Manager
	scheduler      *Scheduler
	config         *config.JobsConfig
	authService    *auth.Service
	loggingService *logging.Service
}

// SetScheduler sets the scheduler for the handler
func (h *Handler) SetScheduler(scheduler *Scheduler) {
	h.scheduler = scheduler
}

// GetStorage returns the jobs storage
func (h *Handler) GetStorage() *Storage {
	return h.storage
}

// roleSatisfiesRequirements checks if the user's role satisfies ANY of the required roles (OR semantics)
// using a hierarchy where: service_role/instance_admin > admin > authenticated > anon
func roleSatisfiesRequirements(userRole string, requiredRoles []string) bool {
	// If no roles required, allow all
	if len(requiredRoles) == 0 {
		return true
	}

	// Service roles bypass all checks
	if userRole == "service_role" || userRole == "instance_admin" || userRole == "tenant_service" {
		return true
	}

	// Define role hierarchy levels (higher number = more privileged)
	roleLevel := map[string]int{
		"anon":           0,
		"authenticated":  1,
		"admin":          2,
		"instance_admin": 3,
		"service_role":   3,
		"tenant_service": 3,
	}

	userLevel, userOk := roleLevel[userRole]
	// If user role is not in hierarchy, it's treated as authenticated level
	// (e.g., custom roles like "moderator", "editor" are at authenticated level)
	if !userOk {
		userLevel = roleLevel["authenticated"]
	}

	// Check if user's role satisfies ANY of the required roles
	for _, requiredRole := range requiredRoles {
		requiredLevel, requiredOk := roleLevel[requiredRole]

		// If the required role is not in the hierarchy, require exact match
		if !requiredOk {
			if userRole == requiredRole {
				return true
			}
			continue
		}

		// Check hierarchy
		if userLevel >= requiredLevel {
			return true
		}
	}

	return false
}

// NewHandler creates a new jobs handler
// npmRegistry and jsrRegistry are optional - if provided, they configure custom registries for Deno bundling
func NewHandler(db *database.Connection, cfg *config.JobsConfig, manager *Manager, authService *auth.Service, loggingService *logging.Service, npmRegistry, jsrRegistry string) (*Handler, error) {
	storage := NewStorage(db)
	loader, err := NewLoader(storage, cfg, npmRegistry, jsrRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to create loader: %w", err)
	}

	return &Handler{
		storage:        storage,
		loader:         loader,
		manager:        manager,
		config:         cfg,
		authService:    authService,
		loggingService: loggingService,
	}, nil
}

// RegisterAdminRoutes registers admin-only routes
func (h *Handler) RegisterAdminRoutes(app *fiber.App) {
	admin := app.Group("/api/v1/admin/jobs")

	// Admin endpoints
	admin.Post("/sync", h.SyncJobs)
	admin.Get("/functions", h.ListJobFunctions)
	admin.Get("/functions/:namespace/:name", h.GetJobFunction)
	admin.Put("/functions/:namespace/:name", h.UpdateJobFunction)
	admin.Delete("/functions/:namespace/:name", h.DeleteJobFunction)
	admin.Get("/stats", h.GetJobStats)
	admin.Get("/workers", h.ListWorkers)

	// Queue operations - admin can see and manage all jobs across users
	admin.Get("/queue", h.ListAllJobs)
	admin.Get("/queue/:id/logs", h.GetJobLogs) // More specific routes must come first
	admin.Post("/queue/:id/terminate", h.TerminateJob)
	admin.Post("/queue/:id/cancel", h.CancelJobAdmin)
	admin.Post("/queue/:id/retry", h.RetryJobAdmin)
	admin.Post("/queue/:id/resubmit", h.ResubmitJobAdmin)
	admin.Get("/queue/:id", h.GetJobAdmin) // Less specific route comes last
}

// fiber:context-methods migrated
