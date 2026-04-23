package functions

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Storage Construction Tests
// =============================================================================

func TestNewStorage(t *testing.T) {
	t.Run("creates storage with nil database", func(t *testing.T) {
		storage := NewStorage(nil)
		assert.NotNil(t, storage)
		assert.Nil(t, storage.db)
	})
}

// =============================================================================
// EdgeFunction Field Validation Tests
// =============================================================================

func TestEdgeFunction_FieldValidation(t *testing.T) {
	t.Run("name is required", func(t *testing.T) {
		fn := EdgeFunction{}
		assert.Empty(t, fn.Name)
	})

	t.Run("code is required", func(t *testing.T) {
		fn := EdgeFunction{Name: "test"}
		assert.Empty(t, fn.Code)
	})

	t.Run("valid function with all required fields", func(t *testing.T) {
		fn := EdgeFunction{
			Name:           "valid-function",
			Code:           "export default () => 'hello';",
			TimeoutSeconds: 30,
			MemoryLimitMB:  128,
		}

		assert.NotEmpty(t, fn.Name)
		assert.NotEmpty(t, fn.Code)
		assert.Equal(t, 30, fn.TimeoutSeconds)
		assert.Equal(t, 128, fn.MemoryLimitMB)
	})
}

// =============================================================================
// EdgeFunction Name Validation Tests
// =============================================================================

func TestEdgeFunction_NameValidation(t *testing.T) {
	validNames := []string{
		"hello",
		"hello-world",
		"hello_world",
		"hello123",
		"my-function-v2",
		"api_handler_main",
		"a",
		"function-with-many-dashes",
	}

	for _, name := range validNames {
		t.Run("valid: "+name, func(t *testing.T) {
			fn := EdgeFunction{Name: name}
			assert.Equal(t, name, fn.Name)
		})
	}
}

// =============================================================================
// EdgeFunction Namespace Tests
// =============================================================================

func TestEdgeFunction_Namespace(t *testing.T) {
	t.Run("default namespace", func(t *testing.T) {
		fn := EdgeFunction{
			Name:      "test",
			Namespace: "default",
		}
		assert.Equal(t, "default", fn.Namespace)
	})

	t.Run("custom namespace", func(t *testing.T) {
		fn := EdgeFunction{
			Name:      "test",
			Namespace: "production",
		}
		assert.Equal(t, "production", fn.Namespace)
	})

	t.Run("empty namespace", func(t *testing.T) {
		fn := EdgeFunction{Name: "test"}
		assert.Empty(t, fn.Namespace)
	})
}

// =============================================================================
// EdgeFunction Code Bundling Tests
// =============================================================================

func TestEdgeFunction_Bundling(t *testing.T) {
	t.Run("unbundled function", func(t *testing.T) {
		originalCode := `import { serve } from "https://deno.land/std/http/server.ts";
export default () => serve(() => new Response("Hello"));`

		fn := EdgeFunction{
			Name:         "unbundled",
			Code:         originalCode,
			OriginalCode: &originalCode,
			IsBundled:    false,
		}

		assert.False(t, fn.IsBundled)
		assert.Equal(t, originalCode, fn.Code)
		assert.Equal(t, originalCode, *fn.OriginalCode)
	})

	t.Run("bundled function", func(t *testing.T) {
		originalCode := `import { hello } from "./utils.ts";
export default () => hello();`
		bundledCode := `// Bundled output
const hello = () => "Hello";
export default () => hello();`

		fn := EdgeFunction{
			Name:         "bundled",
			Code:         bundledCode,
			OriginalCode: &originalCode,
			IsBundled:    true,
		}

		assert.True(t, fn.IsBundled)
		assert.Equal(t, bundledCode, fn.Code)
		assert.Equal(t, originalCode, *fn.OriginalCode)
	})

	t.Run("function with bundle error", func(t *testing.T) {
		errorMsg := "Module not found: ./missing.ts"
		fn := EdgeFunction{
			Name:        "failed-bundle",
			Code:        "export default () => {};",
			IsBundled:   false,
			BundleError: &errorMsg,
		}

		assert.False(t, fn.IsBundled)
		assert.NotNil(t, fn.BundleError)
		assert.Equal(t, "Module not found: ./missing.ts", *fn.BundleError)
	})
}

// =============================================================================
// EdgeFunction Version Tests
// =============================================================================

func TestEdgeFunction_Version(t *testing.T) {
	t.Run("initial version is 0", func(t *testing.T) {
		fn := EdgeFunction{Name: "test"}
		assert.Equal(t, 0, fn.Version)
	})

	t.Run("version increments", func(t *testing.T) {
		fn := EdgeFunction{
			Name:    "versioned",
			Version: 5,
		}
		assert.Equal(t, 5, fn.Version)
	})
}

// =============================================================================
// EdgeFunction Execution Logging Tests
// =============================================================================

func TestEdgeFunction_ExecutionLogging(t *testing.T) {
	t.Run("execution logging enabled by default", func(t *testing.T) {
		fn := EdgeFunction{Name: "test"}
		assert.False(t, fn.DisableExecutionLogs)
	})

	t.Run("execution logging disabled", func(t *testing.T) {
		fn := EdgeFunction{
			Name:                 "no-logs",
			DisableExecutionLogs: true,
		}
		assert.True(t, fn.DisableExecutionLogs)
	})
}

// =============================================================================
// EdgeFunctionExecution JSON Serialization Tests
// =============================================================================

func TestEdgeFunctionExecution_JSONSerialization(t *testing.T) {
	t.Run("pending execution", func(t *testing.T) {
		exec := EdgeFunctionExecution{
			ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			FunctionID:  uuid.MustParse("660e8400-e29b-41d4-a716-446655440000"),
			TriggerType: "http",
			Status:      "pending",
			ExecutedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		}

		data, err := json.Marshal(exec)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"trigger_type":"http"`)
		assert.Contains(t, string(data), `"status":"pending"`)
	})

	t.Run("completed execution with all fields", func(t *testing.T) {
		completedAt := time.Date(2024, 1, 15, 10, 30, 5, 0, time.UTC)
		duration := 5000
		statusCode := 200
		result := `{"message":"success"}`
		logs := "Function started\nProcessing...\nDone"

		exec := EdgeFunctionExecution{
			ID:          uuid.New(),
			FunctionID:  uuid.New(),
			TriggerType: "http",
			Status:      "success",
			StatusCode:  &statusCode,
			DurationMs:  &duration,
			Result:      &result,
			Logs:        &logs,
			ExecutedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			CompletedAt: &completedAt,
		}

		data, err := json.Marshal(exec)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"status":"success"`)
		assert.Contains(t, string(data), `"status_code":200`)
		assert.Contains(t, string(data), `"duration_ms":5000`)
	})

	t.Run("failed execution with error", func(t *testing.T) {
		errorMsg := "TypeError: Cannot read property 'x' of undefined"
		errorStack := "TypeError: Cannot read property 'x' of undefined\n    at handler (function.ts:10:5)"

		exec := EdgeFunctionExecution{
			ID:           uuid.New(),
			FunctionID:   uuid.New(),
			TriggerType:  "cron",
			Status:       "error",
			ErrorMessage: &errorMsg,
			ErrorStack:   &errorStack,
			ExecutedAt:   time.Now(),
		}

		data, err := json.Marshal(exec)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"status":"error"`)
		assert.Contains(t, string(data), `"error_message"`)
		assert.Contains(t, string(data), `"error_stack"`)
	})
}

// =============================================================================
// EdgeFunctionExecution Trigger Payload Tests
// =============================================================================

func TestEdgeFunctionExecution_TriggerPayload(t *testing.T) {
	t.Run("HTTP trigger with request body", func(t *testing.T) {
		payload := `{"method":"POST","path":"/api/hello","body":{"name":"World"}}`

		exec := EdgeFunctionExecution{
			TriggerType:    "http",
			TriggerPayload: &payload,
		}

		assert.Equal(t, "http", exec.TriggerType)
		assert.NotNil(t, exec.TriggerPayload)

		var parsed map[string]interface{}
		err := json.Unmarshal([]byte(*exec.TriggerPayload), &parsed)
		require.NoError(t, err)
		assert.Equal(t, "POST", parsed["method"])
	})

	t.Run("cron trigger without payload", func(t *testing.T) {
		exec := EdgeFunctionExecution{
			TriggerType: "cron",
		}

		assert.Equal(t, "cron", exec.TriggerType)
		assert.Nil(t, exec.TriggerPayload)
	})

	t.Run("webhook trigger with event data", func(t *testing.T) {
		payload := `{"event":"user.created","data":{"id":"123","email":"test@example.com"}}`

		exec := EdgeFunctionExecution{
			TriggerType:    "webhook",
			TriggerPayload: &payload,
		}

		assert.Equal(t, "webhook", exec.TriggerType)
		assert.NotNil(t, exec.TriggerPayload)
	})
}

// =============================================================================
// FunctionFile Path Validation Tests
// =============================================================================

func TestFunctionFile_PathValidation(t *testing.T) {
	validPaths := []string{
		"utils.ts",
		"helpers/db.ts",
		"lib/auth/jwt.ts",
		"_shared/cors.ts",
		"types.d.ts",
		"index.js",
	}

	for _, path := range validPaths {
		t.Run("valid: "+path, func(t *testing.T) {
			file := FunctionFile{
				ID:       uuid.New(),
				FilePath: path,
				Content:  "export const x = 1;",
			}
			assert.Equal(t, path, file.FilePath)
		})
	}
}

// =============================================================================
// SharedModule Path Validation Tests
// =============================================================================

func TestSharedModule_PathValidation(t *testing.T) {
	t.Run("shared module path format", func(t *testing.T) {
		paths := []string{
			"_shared/cors.ts",
			"_shared/auth/jwt.ts",
			"_shared/utils/http.ts",
			"_shared/db/postgres.ts",
		}

		for _, path := range paths {
			module := SharedModule{
				ID:         uuid.New(),
				ModulePath: path,
				Content:    "export const x = 1;",
			}
			assert.Equal(t, path, module.ModulePath)
			assert.Contains(t, module.ModulePath, "_shared/")
		}
	})
}

// =============================================================================
// SharedModule JSON Serialization Tests
// =============================================================================

func TestSharedModule_JSONSerialization(t *testing.T) {
	t.Run("basic module", func(t *testing.T) {
		module := SharedModule{
			ID:         uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			ModulePath: "_shared/utils.ts",
			Content:    "export const add = (a, b) => a + b;",
			Version:    1,
			CreatedAt:  time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		}

		data, err := json.Marshal(module)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"module_path":"_shared/utils.ts"`)
		assert.Contains(t, string(data), `"version":1`)
	})

	t.Run("module with description", func(t *testing.T) {
		desc := "Utility functions for edge functions"
		module := SharedModule{
			ID:          uuid.New(),
			ModulePath:  "_shared/utils.ts",
			Content:     "export const x = 1;",
			Description: &desc,
			Version:     2,
		}

		data, err := json.Marshal(module)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"description":"Utility functions for edge functions"`)
	})
}

// =============================================================================
// EdgeFunction CORS Field Tests
// =============================================================================

func TestEdgeFunction_CORSFields(t *testing.T) {
	t.Run("default CORS (nil values)", func(t *testing.T) {
		fn := EdgeFunction{Name: "no-cors"}

		assert.Nil(t, fn.CorsOrigins)
		assert.Nil(t, fn.CorsMethods)
		assert.Nil(t, fn.CorsHeaders)
		assert.Nil(t, fn.CorsCredentials)
		assert.Nil(t, fn.CorsMaxAge)
	})

	t.Run("custom CORS settings", func(t *testing.T) {
		origins := "https://app.example.com"
		methods := "GET,POST,OPTIONS"
		headers := "Content-Type,Authorization"
		credentials := true
		maxAge := 3600

		fn := EdgeFunction{
			Name:            "custom-cors",
			CorsOrigins:     &origins,
			CorsMethods:     &methods,
			CorsHeaders:     &headers,
			CorsCredentials: &credentials,
			CorsMaxAge:      &maxAge,
		}

		assert.Equal(t, "https://app.example.com", *fn.CorsOrigins)
		assert.Equal(t, "GET,POST,OPTIONS", *fn.CorsMethods)
		assert.Equal(t, "Content-Type,Authorization", *fn.CorsHeaders)
		assert.True(t, *fn.CorsCredentials)
		assert.Equal(t, 3600, *fn.CorsMaxAge)
	})

	t.Run("wildcard CORS", func(t *testing.T) {
		origins := "*"
		methods := "*"
		headers := "*"

		fn := EdgeFunction{
			Name:        "wildcard-cors",
			CorsOrigins: &origins,
			CorsMethods: &methods,
			CorsHeaders: &headers,
		}

		assert.Equal(t, "*", *fn.CorsOrigins)
		assert.Equal(t, "*", *fn.CorsMethods)
		assert.Equal(t, "*", *fn.CorsHeaders)
	})
}

// =============================================================================
// EdgeFunction Rate Limit Field Tests
// =============================================================================

func TestEdgeFunction_RateLimitFields(t *testing.T) {
	t.Run("no rate limits (nil values)", func(t *testing.T) {
		fn := EdgeFunction{Name: "unlimited"}

		assert.Nil(t, fn.RateLimitPerMinute)
		assert.Nil(t, fn.RateLimitPerHour)
		assert.Nil(t, fn.RateLimitPerDay)
	})

	t.Run("all rate limits set", func(t *testing.T) {
		perMin := 10
		perHour := 100
		perDay := 1000

		fn := EdgeFunction{
			Name:               "rate-limited",
			RateLimitPerMinute: &perMin,
			RateLimitPerHour:   &perHour,
			RateLimitPerDay:    &perDay,
		}

		assert.Equal(t, 10, *fn.RateLimitPerMinute)
		assert.Equal(t, 100, *fn.RateLimitPerHour)
		assert.Equal(t, 1000, *fn.RateLimitPerDay)
	})

	t.Run("only minute rate limit", func(t *testing.T) {
		perMin := 60

		fn := EdgeFunction{
			Name:               "minute-limited",
			RateLimitPerMinute: &perMin,
		}

		assert.Equal(t, 60, *fn.RateLimitPerMinute)
		assert.Nil(t, fn.RateLimitPerHour)
		assert.Nil(t, fn.RateLimitPerDay)
	})
}

// =============================================================================
// EdgeFunction CreatedBy Field Tests
// =============================================================================

func TestEdgeFunction_CreatedBy(t *testing.T) {
	t.Run("filesystem source (no creator)", func(t *testing.T) {
		fn := EdgeFunction{
			Name:   "fs-function",
			Source: "filesystem",
		}

		assert.Nil(t, fn.CreatedBy)
		assert.Equal(t, "filesystem", fn.Source)
	})

	t.Run("API source with creator", func(t *testing.T) {
		creatorID := uuid.New()
		fn := EdgeFunction{
			Name:      "api-function",
			Source:    "api",
			CreatedBy: &creatorID,
		}

		assert.NotNil(t, fn.CreatedBy)
		assert.Equal(t, creatorID, *fn.CreatedBy)
		assert.Equal(t, "api", fn.Source)
	})
}

// =============================================================================
// EdgeFunctionSummary Field Tests
// =============================================================================

func TestEdgeFunctionSummary_Fields(t *testing.T) {
	t.Run("summary contains all metadata", func(t *testing.T) {
		desc := "A test function"
		schedule := "*/5 * * * *"
		rateLimit := 100

		summary := EdgeFunctionSummary{
			ID:                 uuid.New(),
			Name:               "test-function",
			Namespace:          "default",
			Description:        &desc,
			IsBundled:          true,
			Version:            3,
			CronSchedule:       &schedule,
			Enabled:            true,
			TimeoutSeconds:     60,
			MemoryLimitMB:      256,
			AllowNet:           true,
			AllowEnv:           true,
			AllowRead:          false,
			AllowWrite:         false,
			RateLimitPerMinute: &rateLimit,
			Source:             "api",
		}

		assert.Equal(t, "test-function", summary.Name)
		assert.Equal(t, "default", summary.Namespace)
		assert.True(t, summary.IsBundled)
		assert.Equal(t, 3, summary.Version)
		assert.Equal(t, "*/5 * * * *", *summary.CronSchedule)
		assert.True(t, summary.Enabled)
		assert.Equal(t, 60, summary.TimeoutSeconds)
		assert.Equal(t, 256, summary.MemoryLimitMB)
		assert.True(t, summary.AllowNet)
		assert.True(t, summary.AllowEnv)
		assert.False(t, summary.AllowRead)
		assert.False(t, summary.AllowWrite)
		assert.Equal(t, 100, *summary.RateLimitPerMinute)
		assert.Equal(t, "api", summary.Source)
	})
}

// =============================================================================
// EdgeFunction Timestamp Tests
// =============================================================================

func TestEdgeFunction_Timestamps(t *testing.T) {
	t.Run("timestamps are set on creation", func(t *testing.T) {
		now := time.Now()
		fn := EdgeFunction{
			Name:      "timestamped",
			CreatedAt: now,
			UpdatedAt: now,
		}

		assert.Equal(t, now, fn.CreatedAt)
		assert.Equal(t, now, fn.UpdatedAt)
	})

	t.Run("updated_at changes on update", func(t *testing.T) {
		created := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		updated := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

		fn := EdgeFunction{
			Name:      "updated",
			CreatedAt: created,
			UpdatedAt: updated,
		}

		assert.True(t, fn.UpdatedAt.After(fn.CreatedAt))
	})
}

// =============================================================================
// EdgeFunctionExecution Duration Tests
// =============================================================================

func TestAllowedFunctionColumns(t *testing.T) {
	t.Run("essential columns are whitelisted", func(t *testing.T) {
		essential := []string{
			"name", "namespace", "description", "code",
			"original_code", "enabled", "cron_schedule", "version",
			"timeout_seconds", "memory_limit_mb",
			"allow_net", "allow_env", "allow_read", "allow_write",
			"allow_unauthenticated", "is_public",
			"is_bundled", "bundle_error", "source",
			"cors_origins", "cors_methods", "cors_headers",
			"cors_credentials", "cors_max_age",
			"rate_limit_per_minute", "rate_limit_per_hour", "rate_limit_per_day",
			"disable_execution_logs", "needs_rebundle",
			"created_by",
		}
		for _, col := range essential {
			assert.True(t, allowedFunctionColumns[col], "column %q should be in whitelist", col)
		}
	})

	t.Run("protected columns are NOT whitelisted", func(t *testing.T) {
		protected := []string{"id", "created_at", "updated_at", "tenant_id"}
		for _, col := range protected {
			assert.False(t, allowedFunctionColumns[col], "protected column %q must NOT be in whitelist", col)
		}
	})

	t.Run("nonsense columns are NOT whitelisted", func(t *testing.T) {
		nonsense := []string{"admin_override", "password", "secret"}
		for _, col := range nonsense {
			assert.False(t, allowedFunctionColumns[col], "nonsense column %q must NOT be in whitelist", col)
		}
	})
}

func TestEdgeFunctionExecution_Duration(t *testing.T) {
	t.Run("fast execution", func(t *testing.T) {
		duration := 50 // 50ms

		exec := EdgeFunctionExecution{
			Status:     "success",
			DurationMs: &duration,
		}

		assert.Equal(t, 50, *exec.DurationMs)
	})

	t.Run("slow execution", func(t *testing.T) {
		duration := 29500 // 29.5 seconds

		exec := EdgeFunctionExecution{
			Status:     "success",
			DurationMs: &duration,
		}

		assert.Equal(t, 29500, *exec.DurationMs)
	})

	t.Run("timeout execution", func(t *testing.T) {
		duration := 30000 // 30 seconds (timeout)
		errorMsg := "Function execution timed out"

		exec := EdgeFunctionExecution{
			Status:       "timeout",
			DurationMs:   &duration,
			ErrorMessage: &errorMsg,
		}

		assert.Equal(t, "timeout", exec.Status)
		assert.Equal(t, 30000, *exec.DurationMs)
	})
}
