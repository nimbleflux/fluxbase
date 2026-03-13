package tools

import (
	"context"
	"testing"

	"github.com/nimbleflux/fluxbase/internal/mcp"
	"github.com/stretchr/testify/assert"
)

func TestNewSyncJobTool(t *testing.T) {
	t.Run("creates tool with nil storage", func(t *testing.T) {
		tool := NewSyncJobTool(nil)
		assert.NotNil(t, tool)
		assert.Nil(t, tool.storage)
	})
}

func TestSyncJobTool_Metadata(t *testing.T) {
	tool := NewSyncJobTool(nil)

	t.Run("name", func(t *testing.T) {
		assert.Equal(t, "sync_job", tool.Name())
	})

	t.Run("description", func(t *testing.T) {
		desc := tool.Description()
		assert.Contains(t, desc, "Deploy or update a background job")
		assert.Contains(t, desc, "@fluxbase:schedule")
		assert.Contains(t, desc, "@fluxbase:timeout")
		assert.Contains(t, desc, "@fluxbase:max-retries")
	})

	t.Run("input schema", func(t *testing.T) {
		schema := tool.InputSchema()
		assert.Equal(t, "object", schema["type"])

		props := schema["properties"].(map[string]any)
		assert.Contains(t, props, "name")
		assert.Contains(t, props, "code")
		assert.Contains(t, props, "namespace")

		required := schema["required"].([]string)
		assert.Contains(t, required, "name")
		assert.Contains(t, required, "code")
	})

	t.Run("required scopes", func(t *testing.T) {
		scopes := tool.RequiredScopes()
		assert.Contains(t, scopes, mcp.ScopeSyncJobs)
	})
}

func TestSyncJobTool_Execute_Validation(t *testing.T) {
	tool := NewSyncJobTool(nil)

	t.Run("missing name", func(t *testing.T) {
		_, err := tool.Execute(context.Background(), map[string]any{
			"code": "export default function() {}",
		}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("missing code", func(t *testing.T) {
		_, err := tool.Execute(context.Background(), map[string]any{
			"name": "test-job",
		}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "code is required")
	})

	t.Run("invalid name format", func(t *testing.T) {
		_, err := tool.Execute(context.Background(), map[string]any{
			"name": "invalid job name!",
			"code": "export default function() {}",
		}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid job name")
	})
}

func TestJobConfig_Struct(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		var config JobConfig
		assert.Empty(t, config.Description)
		assert.Empty(t, config.Schedule)
		assert.Equal(t, 0, config.Timeout)
		assert.Equal(t, 0, config.Memory)
		assert.Equal(t, 0, config.MaxRetries)
		assert.Nil(t, config.RequireRoles)
	})

	t.Run("all fields", func(t *testing.T) {
		config := JobConfig{
			Description:  "Daily cleanup",
			Schedule:     "0 0 * * *",
			Timeout:      600,
			Memory:       512,
			MaxRetries:   5,
			RequireRoles: []string{"admin"},
			AllowNet:     true,
			AllowEnv:     true,
			DisableLogs:  true,
		}

		assert.Equal(t, "Daily cleanup", config.Description)
		assert.Equal(t, "0 0 * * *", config.Schedule)
		assert.Equal(t, 600, config.Timeout)
		assert.Equal(t, 5, config.MaxRetries)
	})
}

func TestParseJobAnnotations(t *testing.T) {
	t.Run("empty code returns defaults", func(t *testing.T) {
		config := parseJobAnnotations("")
		assert.Equal(t, 300, config.Timeout)
		assert.Equal(t, 256, config.Memory)
		assert.Equal(t, 3, config.MaxRetries)
		assert.True(t, config.AllowNet)
		assert.True(t, config.AllowEnv)
		assert.Empty(t, config.Schedule)
	})

	t.Run("schedule annotation with quotes", func(t *testing.T) {
		code := `// @fluxbase:schedule "0 */5 * * *"
export default function() {}`
		config := parseJobAnnotations(code)
		assert.Equal(t, "0 */5 * * *", config.Schedule)
	})

	t.Run("schedule annotation with single quotes", func(t *testing.T) {
		code := `// @fluxbase:schedule '0 0 * * *'
export default function() {}`
		config := parseJobAnnotations(code)
		assert.Equal(t, "0 0 * * *", config.Schedule)
	})

	t.Run("description annotation", func(t *testing.T) {
		code := `// @fluxbase:description Daily cleanup job
export default function() {}`
		config := parseJobAnnotations(code)
		assert.Equal(t, "Daily cleanup job", config.Description)
	})

	t.Run("timeout annotation", func(t *testing.T) {
		code := `// @fluxbase:timeout 600
export default function() {}`
		config := parseJobAnnotations(code)
		assert.Equal(t, 600, config.Timeout)
	})

	t.Run("memory annotation", func(t *testing.T) {
		code := `// @fluxbase:memory 512
export default function() {}`
		config := parseJobAnnotations(code)
		assert.Equal(t, 512, config.Memory)
	})

	t.Run("max-retries annotation", func(t *testing.T) {
		code := `// @fluxbase:max-retries 5
export default function() {}`
		config := parseJobAnnotations(code)
		assert.Equal(t, 5, config.MaxRetries)
	})

	t.Run("max-retries zero", func(t *testing.T) {
		code := `// @fluxbase:max-retries 0
export default function() {}`
		config := parseJobAnnotations(code)
		assert.Equal(t, 0, config.MaxRetries)
	})

	t.Run("require-role annotation", func(t *testing.T) {
		code := `// @fluxbase:require-role admin, service
export default function() {}`
		config := parseJobAnnotations(code)
		assert.Equal(t, []string{"admin", "service"}, config.RequireRoles)
	})

	t.Run("deny-net annotation", func(t *testing.T) {
		code := `// @fluxbase:deny-net
export default function() {}`
		config := parseJobAnnotations(code)
		assert.False(t, config.AllowNet)
	})

	t.Run("deny-env annotation", func(t *testing.T) {
		code := `// @fluxbase:deny-env
export default function() {}`
		config := parseJobAnnotations(code)
		assert.False(t, config.AllowEnv)
	})

	t.Run("disable-logs annotation", func(t *testing.T) {
		code := `// @fluxbase:disable-logs
export default function() {}`
		config := parseJobAnnotations(code)
		assert.True(t, config.DisableLogs)
	})

	t.Run("multiple annotations", func(t *testing.T) {
		code := `// @fluxbase:schedule "0 0 * * *"
// @fluxbase:description Nightly data cleanup
// @fluxbase:timeout 1800
// @fluxbase:memory 1024
// @fluxbase:max-retries 5
// @fluxbase:disable-logs
export default async function cleanup() {
  // Cleanup logic
}`
		config := parseJobAnnotations(code)
		assert.Equal(t, "0 0 * * *", config.Schedule)
		assert.Equal(t, "Nightly data cleanup", config.Description)
		assert.Equal(t, 1800, config.Timeout)
		assert.Equal(t, 1024, config.Memory)
		assert.Equal(t, 5, config.MaxRetries)
		assert.True(t, config.DisableLogs)
	})

	t.Run("block comment style", func(t *testing.T) {
		code := `/*
 * @fluxbase:schedule "0 */15 * * *"
 * @fluxbase:description Every 15 minutes
 */
export default function() {}`
		config := parseJobAnnotations(code)
		assert.Equal(t, "0 */15 * * *", config.Schedule)
		assert.Equal(t, "Every 15 minutes", config.Description)
	})
}

func TestIsValidCronExpression(t *testing.T) {
	t.Run("valid expressions", func(t *testing.T) {
		validExpressions := []string{
			"* * * * *",      // Every minute
			"0 * * * *",      // Every hour
			"0 0 * * *",      // Daily at midnight
			"0 0 * * 0",      // Weekly on Sunday
			"0 0 1 * *",      // Monthly on 1st
			"*/5 * * * *",    // Every 5 minutes
			"0 */2 * * *",    // Every 2 hours
			"0 0 */3 * *",    // Every 3 days
			"0 9-17 * * 1-5", // Weekdays 9-5
			"0 0,12 * * *",   // Noon and midnight
			"0 0 1,15 * *",   // 1st and 15th
			"0 0 * * * *",    // 6-field (with seconds)
		}

		for _, expr := range validExpressions {
			assert.True(t, isValidCronExpression(expr), "Expected %q to be valid", expr)
		}
	})

	t.Run("invalid expressions", func(t *testing.T) {
		invalidExpressions := []string{
			"",                // Empty
			"* * * *",         // Too few fields
			"* * * * * * *",   // Too many fields
			"abc * * * *",     // Invalid characters
			"* * * * * hello", // Invalid characters in field
		}

		for _, expr := range invalidExpressions {
			assert.False(t, isValidCronExpression(expr), "Expected %q to be invalid", expr)
		}
	})

	t.Run("special cron characters", func(t *testing.T) {
		specialExpressions := []string{
			"0 0 L * *",   // Last day of month
			"0 0 W * *",   // Nearest weekday
			"0 0 * * 1#2", // Second Monday
			"0 0 15W * *", // Nearest weekday to 15th
			"? * * * *",   // Question mark (no specific value)
		}

		for _, expr := range specialExpressions {
			assert.True(t, isValidCronExpression(expr), "Expected %q to be valid", expr)
		}
	})
}
