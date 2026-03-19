package jobs

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleSatisfiesRequirement(t *testing.T) {
	tests := []struct {
		name         string
		userRole     string
		requiredRole string
		expected     bool
	}{
		// service_role can access everything (highest privilege)
		{"service_role satisfies admin", "service_role", "admin", true},
		{"service_role satisfies authenticated", "service_role", "authenticated", true},
		{"service_role satisfies anon", "service_role", "anon", true},
		{"service_role satisfies custom role", "service_role", "moderator", true},

		// instance_admin can access everything (highest privilege)
		{"instance_admin satisfies admin", "instance_admin", "admin", true},
		{"instance_admin satisfies authenticated", "instance_admin", "authenticated", true},
		{"instance_admin satisfies anon", "instance_admin", "anon", true},
		{"instance_admin satisfies custom role", "instance_admin", "editor", true},

		// Admin can access everything except service_role/instance_admin level
		{"admin satisfies admin", "admin", "admin", true},
		{"admin satisfies authenticated", "admin", "authenticated", true},
		{"admin satisfies anon", "admin", "anon", true},

		// Authenticated can access authenticated and anon
		{"authenticated satisfies authenticated", "authenticated", "authenticated", true},
		{"authenticated satisfies anon", "authenticated", "anon", true},
		{"authenticated does not satisfy admin", "authenticated", "admin", false},

		// Anon can only access anon
		{"anon satisfies anon", "anon", "anon", true},
		{"anon does not satisfy authenticated", "anon", "authenticated", false},
		{"anon does not satisfy admin", "anon", "admin", false},

		// Custom roles are treated as authenticated level
		{"custom role satisfies authenticated", "moderator", "authenticated", true},
		{"custom role satisfies anon", "editor", "anon", true},
		{"custom role does not satisfy admin", "moderator", "admin", false},

		// Custom required roles require exact match
		{"exact match for custom required role", "moderator", "moderator", true},
		{"no match for different custom role", "editor", "moderator", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := roleSatisfiesRequirements(tt.userRole, []string{tt.requiredRole})
			if result != tt.expected {
				t.Errorf("roleSatisfiesRequirements(%q, %q) = %v, want %v",
					tt.userRole, tt.requiredRole, result, tt.expected)
			}
		})
	}
}

// TestEmbeddedSDKEndpoint verifies that the embedded SDK uses the correct API endpoint
// for database operations. This prevents regressions where the endpoint path might
// be accidentally changed back to an incorrect value.
func TestEmbeddedSDKEndpoint(t *testing.T) {
	// Read the embedded SDK file (now in internal/runtime/)
	embeddedSDKCode, err := os.ReadFile("../runtime/embedded_sdk.js")
	if err != nil {
		t.Fatalf("Failed to read embedded_sdk.js: %v", err)
	}

	code := string(embeddedSDKCode)

	// Verify the QueryBuilder uses the correct endpoint path
	correctEndpoint := "/api/v1/tables/"
	incorrectEndpoint := "/api/v1/rest/"

	// The embedded SDK should contain the correct endpoint
	if !strings.Contains(code, correctEndpoint) {
		t.Errorf("Embedded SDK does not contain the correct endpoint %q", correctEndpoint)
	}

	// The embedded SDK should NOT contain the old incorrect endpoint
	if strings.Contains(code, incorrectEndpoint) {
		t.Errorf("Embedded SDK contains the incorrect endpoint %q. "+
			"Database operations in job handlers must use %q for proper routing.",
			incorrectEndpoint, correctEndpoint)
	}

	// Additional validation: ensure QueryBuilder uses buildTablePath which returns the correct path
	// Look for the buildTablePath method that constructs the /api/v1/tables/ path
	buildTablePathIndex := strings.Index(code, "buildTablePath()")
	if buildTablePathIndex == -1 {
		t.Fatal("Could not find buildTablePath() method in embedded SDK")
	}

	// Extract a reasonable section of code after buildTablePath() to check the path construction
	endIndex := buildTablePathIndex + 200
	if endIndex > len(code) {
		endIndex = len(code)
	}
	codeSection := code[buildTablePathIndex:endIndex]

	// Check that buildTablePath returns the correct API endpoint
	if !strings.Contains(codeSection, "/api/v1/tables/") {
		t.Error("buildTablePath() does not construct path with '/api/v1/tables/'")
	}
}

// =============================================================================
// roleSatisfiesRequirements Extended Tests
// =============================================================================

func TestRoleSatisfiesRequirements_EmptyRoles(t *testing.T) {
	t.Run("allows all when no roles required", func(t *testing.T) {
		assert.True(t, roleSatisfiesRequirements("anon", []string{}))
		assert.True(t, roleSatisfiesRequirements("authenticated", []string{}))
		assert.True(t, roleSatisfiesRequirements("admin", []string{}))
		assert.True(t, roleSatisfiesRequirements("custom_role", []string{}))
	})

	t.Run("allows all when nil roles", func(t *testing.T) {
		assert.True(t, roleSatisfiesRequirements("anon", nil))
		assert.True(t, roleSatisfiesRequirements("authenticated", nil))
	})
}

func TestRoleSatisfiesRequirements_MultipleRoles(t *testing.T) {
	t.Run("satisfies any of multiple required roles", func(t *testing.T) {
		// User has admin, requires admin OR authenticated - should pass
		assert.True(t, roleSatisfiesRequirements("admin", []string{"admin", "authenticated"}))

		// User is authenticated, requires admin OR authenticated - should pass
		assert.True(t, roleSatisfiesRequirements("authenticated", []string{"admin", "authenticated"}))

		// User is anon, requires admin OR authenticated - should fail
		assert.False(t, roleSatisfiesRequirements("anon", []string{"admin", "authenticated"}))
	})

	t.Run("custom role matches one of multiple", func(t *testing.T) {
		// moderator is one of the required roles
		assert.True(t, roleSatisfiesRequirements("moderator", []string{"moderator", "admin"}))

		// moderator satisfies authenticated which is one of the required
		assert.True(t, roleSatisfiesRequirements("moderator", []string{"authenticated", "admin"}))

		// moderator is treated as authenticated, but anon is also in list
		assert.True(t, roleSatisfiesRequirements("moderator", []string{"anon"}))
	})
}

// =============================================================================
// valueOr Tests
// =============================================================================

func TestValueOr(t *testing.T) {
	t.Run("returns value when pointer is non-nil", func(t *testing.T) {
		val := 42
		result := valueOr(&val, 0)
		assert.Equal(t, 42, result)
	})

	t.Run("returns default when pointer is nil", func(t *testing.T) {
		var ptr *int
		result := valueOr(ptr, 99)
		assert.Equal(t, 99, result)
	})

	t.Run("works with string", func(t *testing.T) {
		val := "hello"
		assert.Equal(t, "hello", valueOr(&val, "default"))

		var nilStr *string
		assert.Equal(t, "default", valueOr(nilStr, "default"))
	})

	t.Run("works with bool", func(t *testing.T) {
		val := true
		assert.True(t, valueOr(&val, false))

		var nilBool *bool
		assert.False(t, valueOr(nilBool, false))
	})

	t.Run("works with time.Time", func(t *testing.T) {
		now := time.Now()
		zero := time.Time{}

		assert.Equal(t, now, valueOr(&now, zero))

		var nilTime *time.Time
		assert.Equal(t, zero, valueOr(nilTime, zero))
	})

	t.Run("works with uuid.UUID", func(t *testing.T) {
		id := uuid.New()
		zero := uuid.UUID{}

		assert.Equal(t, id, valueOr(&id, zero))

		var nilUUID *uuid.UUID
		assert.Equal(t, zero, valueOr(nilUUID, zero))
	})
}

// =============================================================================
// toString Tests
// =============================================================================

func TestToString(t *testing.T) {
	t.Run("returns empty string for nil", func(t *testing.T) {
		assert.Equal(t, "", toString(nil))
	})

	t.Run("returns string as is", func(t *testing.T) {
		assert.Equal(t, "hello", toString("hello"))
		assert.Equal(t, "", toString(""))
	})

	t.Run("converts uuid.UUID pointer", func(t *testing.T) {
		id := uuid.MustParse("12345678-1234-1234-1234-123456789abc")
		result := toString(&id)
		assert.Equal(t, "12345678-1234-1234-1234-123456789abc", result)
	})

	t.Run("returns empty for nil uuid pointer", func(t *testing.T) {
		var nilUUID *uuid.UUID
		assert.Equal(t, "", toString(nilUUID))
	})

	t.Run("converts int to string", func(t *testing.T) {
		assert.Equal(t, "42", toString(42))
		assert.Equal(t, "0", toString(0))
		assert.Equal(t, "-1", toString(-1))
	})

	t.Run("converts float to string", func(t *testing.T) {
		assert.Equal(t, "3.14", toString(3.14))
	})

	t.Run("converts bool to string", func(t *testing.T) {
		assert.Equal(t, "true", toString(true))
		assert.Equal(t, "false", toString(false))
	})
}

// =============================================================================
// JobStatus Tests
// =============================================================================

func TestJobsHandler_JobStatusConstants(t *testing.T) {
	t.Run("status values are correct", func(t *testing.T) {
		assert.Equal(t, JobStatus("pending"), JobStatusPending)
		assert.Equal(t, JobStatus("running"), JobStatusRunning)
		assert.Equal(t, JobStatus("completed"), JobStatusCompleted)
		assert.Equal(t, JobStatus("failed"), JobStatusFailed)
		assert.Equal(t, JobStatus("cancelled"), JobStatusCancelled)
		assert.Equal(t, JobStatus("interrupted"), JobStatusInterrupted)
	})
}

func TestJobsHandler_WorkerStatusConstants(t *testing.T) {
	t.Run("worker status values are correct", func(t *testing.T) {
		assert.Equal(t, WorkerStatus("active"), WorkerStatusActive)
		assert.Equal(t, WorkerStatus("draining"), WorkerStatusDraining)
		assert.Equal(t, WorkerStatus("stopped"), WorkerStatusStopped)
	})
}

// =============================================================================
// Job Struct Tests
// =============================================================================

func TestJobsHandler_JobStruct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		id := uuid.New()
		funcID := uuid.New()
		userID := uuid.New()
		workerID := uuid.New()
		payload := `{"key":"value"}`
		result := `{"success":true}`
		progress := `{"percent":50}`
		errorMsg := "test error"
		role := "authenticated"
		email := "user@example.com"
		now := time.Now()
		later := now.Add(time.Hour)

		job := Job{
			ID:                     id,
			Namespace:              "test",
			JobFunctionID:          &funcID,
			JobName:                "test_job",
			Status:                 JobStatusRunning,
			Payload:                &payload,
			Result:                 &result,
			Progress:               &progress,
			Priority:               5,
			MaxRetries:             3,
			RetryCount:             1,
			MaxDurationSeconds:     intPtr(300),
			ProgressTimeoutSeconds: intPtr(60),
			ErrorMessage:           &errorMsg,
			WorkerID:               &workerID,
			CreatedBy:              &userID,
			UserRole:               &role,
			UserEmail:              &email,
			ScheduledAt:            &now,
			StartedAt:              &now,
			CompletedAt:            &later,
			CreatedAt:              now,
		}

		assert.Equal(t, id, job.ID)
		assert.Equal(t, "test", job.Namespace)
		assert.Equal(t, &funcID, job.JobFunctionID)
		assert.Equal(t, "test_job", job.JobName)
		assert.Equal(t, JobStatusRunning, job.Status)
		assert.Equal(t, &payload, job.Payload)
		assert.Equal(t, 5, job.Priority)
		assert.Equal(t, 3, job.MaxRetries)
		assert.Equal(t, 1, job.RetryCount)
	})
}

func TestJobsHandler_JobFunctionStruct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		id := uuid.New()
		desc := "Test job function"
		code := "export function handler() {}"
		schedule := "0 * * * *"
		now := time.Now()

		fn := JobFunction{
			ID:                     id,
			Name:                   "test_function",
			Namespace:              "default",
			Description:            &desc,
			Code:                   &code,
			IsBundled:              true,
			Enabled:                true,
			Schedule:               &schedule,
			TimeoutSeconds:         300,
			MemoryLimitMB:          256,
			MaxRetries:             3,
			ProgressTimeoutSeconds: 60,
			AllowNet:               true,
			AllowEnv:               true,
			AllowRead:              false,
			AllowWrite:             false,
			RequireRoles:           []string{"authenticated"},
			DisableExecutionLogs:   false,
			Version:                1,
			Source:                 "filesystem",
			CreatedAt:              now,
			UpdatedAt:              now,
		}

		assert.Equal(t, id, fn.ID)
		assert.Equal(t, "test_function", fn.Name)
		assert.Equal(t, "default", fn.Namespace)
		assert.Equal(t, &desc, fn.Description)
		assert.True(t, fn.Enabled)
		assert.True(t, fn.AllowNet)
		assert.Equal(t, []string{"authenticated"}, fn.RequireRoles)
	})
}

// =============================================================================
// Job JSON Serialization Tests
// =============================================================================

func TestJobsHandler_JobJSONSerialization(t *testing.T) {
	t.Run("serializes job to JSON", func(t *testing.T) {
		id := uuid.MustParse("12345678-1234-1234-1234-123456789abc")
		job := Job{
			ID:        id,
			Namespace: "default",
			JobName:   "test_job",
			Status:    JobStatusPending,
			Priority:  1,
		}

		data, err := json.Marshal(job)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"id":"12345678-1234-1234-1234-123456789abc"`)
		assert.Contains(t, string(data), `"namespace":"default"`)
		assert.Contains(t, string(data), `"job_name":"test_job"`)
		assert.Contains(t, string(data), `"status":"pending"`)
	})

	t.Run("deserializes job from JSON", func(t *testing.T) {
		jsonData := `{
			"id": "12345678-1234-1234-1234-123456789abc",
			"namespace": "default",
			"job_name": "test_job",
			"status": "running",
			"priority": 5
		}`

		var job Job
		err := json.Unmarshal([]byte(jsonData), &job)
		require.NoError(t, err)

		assert.Equal(t, uuid.MustParse("12345678-1234-1234-1234-123456789abc"), job.ID)
		assert.Equal(t, "default", job.Namespace)
		assert.Equal(t, "test_job", job.JobName)
		assert.Equal(t, JobStatusRunning, job.Status)
		assert.Equal(t, 5, job.Priority)
	})
}

func TestJobFunction_JSONSerialization(t *testing.T) {
	t.Run("serializes job function to JSON", func(t *testing.T) {
		id := uuid.MustParse("12345678-1234-1234-1234-123456789abc")
		fn := JobFunction{
			ID:             id,
			Name:           "my_function",
			Namespace:      "jobs",
			Enabled:        true,
			TimeoutSeconds: 300,
			RequireRoles:   []string{"admin", "authenticated"},
		}

		data, err := json.Marshal(fn)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"name":"my_function"`)
		assert.Contains(t, string(data), `"enabled":true`)
		assert.Contains(t, string(data), `"require_roles":["admin","authenticated"]`)
	})
}

// =============================================================================
// Handler Struct Tests
// =============================================================================

func TestJobsHandler_Struct(t *testing.T) {
	t.Run("fields are accessible", func(t *testing.T) {
		h := &Handler{}

		assert.Nil(t, h.storage)
		assert.Nil(t, h.loader)
		assert.Nil(t, h.manager)
		assert.Nil(t, h.scheduler)
		assert.Nil(t, h.config)
		assert.Nil(t, h.authService)
		assert.Nil(t, h.loggingService)
	})
}

func TestJobsHandler_SetScheduler(t *testing.T) {
	t.Run("sets scheduler", func(t *testing.T) {
		h := &Handler{}
		scheduler := &Scheduler{}

		h.SetScheduler(scheduler)

		assert.Equal(t, scheduler, h.scheduler)
	})

	t.Run("allows nil scheduler", func(t *testing.T) {
		h := &Handler{}
		h.SetScheduler(nil)

		assert.Nil(t, h.scheduler)
	})
}

// =============================================================================
// ErrDuplicateJob Tests
// =============================================================================

func TestErrDuplicateJob(t *testing.T) {
	t.Run("error message is correct", func(t *testing.T) {
		assert.Equal(t, "duplicate job already pending or running", ErrDuplicateJob.Error())
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkRoleSatisfiesRequirements_ServiceRole(b *testing.B) {
	roles := []string{"admin", "authenticated"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		roleSatisfiesRequirements("service_role", roles)
	}
}

func BenchmarkRoleSatisfiesRequirements_CustomRole(b *testing.B) {
	roles := []string{"admin", "moderator", "editor"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		roleSatisfiesRequirements("moderator", roles)
	}
}

func BenchmarkValueOr_NonNil(b *testing.B) {
	val := 42
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		valueOr(&val, 0)
	}
}

func BenchmarkValueOr_Nil(b *testing.B) {
	var ptr *int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		valueOr(ptr, 99)
	}
}

func BenchmarkToString_String(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		toString("test")
	}
}

func BenchmarkToString_UUID(b *testing.B) {
	id := uuid.New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		toString(&id)
	}
}

// =============================================================================
// Helper functions for tests
// =============================================================================

func intPtr(v int) *int {
	return &v
}
