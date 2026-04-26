package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nimbleflux/fluxbase/internal/functions"
	"github.com/nimbleflux/fluxbase/internal/mcp"
)

func TestNewSyncFunctionTool(t *testing.T) {
	t.Run("creates tool with nil storage", func(t *testing.T) {
		tool := NewSyncFunctionTool(nil)
		assert.NotNil(t, tool)
		assert.Nil(t, tool.storage)
	})
}

func TestSyncFunctionTool_Metadata(t *testing.T) {
	tool := NewSyncFunctionTool(nil)

	t.Run("name", func(t *testing.T) {
		assert.Equal(t, "sync_function", tool.Name())
	})

	t.Run("description", func(t *testing.T) {
		desc := tool.Description()
		assert.Contains(t, desc, "Deploy or update an edge function")
		assert.Contains(t, desc, "@fluxbase:public")
		assert.Contains(t, desc, "@fluxbase:timeout")
		assert.Contains(t, desc, "@fluxbase:rate-limit")
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
		assert.Contains(t, scopes, mcp.ScopeSyncFunctions)
	})
}

func TestSyncFunctionTool_Execute_Validation(t *testing.T) {
	tool := NewSyncFunctionTool(nil)

	t.Run("missing name", func(t *testing.T) {
		_, err := tool.Execute(context.Background(), map[string]any{
			"code": "export default function() {}",
		}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("missing code", func(t *testing.T) {
		_, err := tool.Execute(context.Background(), map[string]any{
			"name": "test-fn",
		}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "code is required")
	})

	t.Run("invalid name format", func(t *testing.T) {
		_, err := tool.Execute(context.Background(), map[string]any{
			"name": "invalid name!",
			"code": "export default function() {}",
		}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid function name")
	})
}

func TestParseFunctionConfig_Annotations(t *testing.T) {
	t.Run("empty code returns defaults", func(t *testing.T) {
		config := functions.ParseFunctionConfig("")
		assert.Equal(t, 30, config.Timeout)
		assert.Equal(t, 128, config.Memory)
		assert.True(t, config.AllowNet)
		assert.True(t, config.AllowEnv)
		assert.True(t, config.IsPublic)
		assert.False(t, config.AllowUnauthenticated)
	})

	t.Run("public annotation", func(t *testing.T) {
		code := `// @fluxbase:public
export default function() {}`
		config := functions.ParseFunctionConfig(code)
		assert.True(t, config.IsPublic)
	})

	t.Run("allow-unauthenticated annotation", func(t *testing.T) {
		code := `// @fluxbase:allow-unauthenticated
export default function() {}`
		config := functions.ParseFunctionConfig(code)
		assert.True(t, config.AllowUnauthenticated)
	})

	t.Run("description annotation", func(t *testing.T) {
		code := `// @fluxbase:description Handle user registration and validation
export default function() {}`
		config := functions.ParseFunctionConfig(code)
		assert.Equal(t, "Handle user registration and validation", config.Description)
	})

	t.Run("timeout annotation", func(t *testing.T) {
		code := `// @fluxbase:timeout 60
export default function() {}`
		config := functions.ParseFunctionConfig(code)
		assert.Equal(t, 60, config.Timeout)
	})

	t.Run("timeout annotation - invalid value", func(t *testing.T) {
		code := `// @fluxbase:timeout invalid
export default function() {}`
		config := functions.ParseFunctionConfig(code)
		assert.Equal(t, 30, config.Timeout)
	})

	t.Run("memory annotation", func(t *testing.T) {
		code := `// @fluxbase:memory 256
export default function() {}`
		config := functions.ParseFunctionConfig(code)
		assert.Equal(t, 256, config.Memory)
	})

	t.Run("cors-origins annotation", func(t *testing.T) {
		code := `// @fluxbase:cors-origins https://example.com,https://app.example.com
export default function() {}`
		config := functions.ParseFunctionConfig(code)
		assert.NotNil(t, config.CorsOrigins)
		assert.Equal(t, "https://example.com,https://app.example.com", *config.CorsOrigins)
	})

	t.Run("deny-net annotation", func(t *testing.T) {
		code := `// @fluxbase:deny-net
export default function() {}`
		config := functions.ParseFunctionConfig(code)
		assert.False(t, config.AllowNet)
	})

	t.Run("deny-env annotation", func(t *testing.T) {
		code := `// @fluxbase:deny-env
export default function() {}`
		config := functions.ParseFunctionConfig(code)
		assert.False(t, config.AllowEnv)
	})

	t.Run("disable-logs annotation", func(t *testing.T) {
		code := `// @fluxbase:disable-logs
export default function() {}`
		config := functions.ParseFunctionConfig(code)
		assert.True(t, config.DisableExecutionLogs)
	})

	t.Run("rate-limit per minute", func(t *testing.T) {
		code := `// @fluxbase:rate-limit 100/min
export default function() {}`
		config := functions.ParseFunctionConfig(code)
		assert.NotNil(t, config.RateLimitPerMinute)
		assert.Equal(t, 100, *config.RateLimitPerMinute)
		assert.Nil(t, config.RateLimitPerHour)
		assert.Nil(t, config.RateLimitPerDay)
	})

	t.Run("rate-limit per hour", func(t *testing.T) {
		code := `// @fluxbase:rate-limit 1000/hour
export default function() {}`
		config := functions.ParseFunctionConfig(code)
		assert.NotNil(t, config.RateLimitPerHour)
		assert.Equal(t, 1000, *config.RateLimitPerHour)
	})

	t.Run("rate-limit per day", func(t *testing.T) {
		code := `// @fluxbase:rate-limit 10000/day
export default function() {}`
		config := functions.ParseFunctionConfig(code)
		assert.NotNil(t, config.RateLimitPerDay)
		assert.Equal(t, 10000, *config.RateLimitPerDay)
	})

	t.Run("multiple annotations", func(t *testing.T) {
		code := `// @fluxbase:description API endpoint
// @fluxbase:public
// @fluxbase:allow-unauthenticated
// @fluxbase:timeout 45
// @fluxbase:memory 512
// @fluxbase:rate-limit 50/min
// @fluxbase:cors-origins *
export default async function handler(req: Request) {
  return new Response("Hello!");
}`
		config := functions.ParseFunctionConfig(code)
		assert.Equal(t, "API endpoint", config.Description)
		assert.True(t, config.IsPublic)
		assert.True(t, config.AllowUnauthenticated)
		assert.Equal(t, 45, config.Timeout)
		assert.Equal(t, 512, config.Memory)
		assert.NotNil(t, config.RateLimitPerMinute)
		assert.Equal(t, 50, *config.RateLimitPerMinute)
		assert.NotNil(t, config.CorsOrigins)
		assert.Equal(t, "*", *config.CorsOrigins)
	})

	t.Run("block comment style", func(t *testing.T) {
		code := `/*
 * @fluxbase:description Block comment function
 * @fluxbase:public
 * @fluxbase:timeout 120
 */
export default function() {}`
		config := functions.ParseFunctionConfig(code)
		assert.Equal(t, "Block comment function", config.Description)
		assert.True(t, config.IsPublic)
		assert.Equal(t, 120, config.Timeout)
	})

	t.Run("case insensitivity", func(t *testing.T) {
		code := `// @fluxbase:PUBLIC
// @fluxbase:TIMEOUT 60
// @fluxbase:Description Test`
		config := functions.ParseFunctionConfig(code)
		assert.True(t, config.IsPublic)
		assert.Equal(t, 60, config.Timeout)
		assert.Equal(t, "Test", config.Description)
	})

	t.Run("public false annotation", func(t *testing.T) {
		code := `// @fluxbase:public false
export default function() {}`
		config := functions.ParseFunctionConfig(code)
		assert.False(t, config.IsPublic)
	})
}

func TestIsValidFunctionName(t *testing.T) {
	t.Run("valid names", func(t *testing.T) {
		validNames := []string{
			"myfunction",
			"my_function",
			"my-function",
			"MyFunction",
			"_private",
			"fn1",
			"a",
		}

		for _, name := range validNames {
			assert.True(t, isValidFunctionName(name), "Expected %q to be valid", name)
		}
	})

	t.Run("invalid names", func(t *testing.T) {
		invalidNames := []string{
			"",
			"1function",
			"-function",
			"my function",
			"my.function",
			"my@function",
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		}

		for _, name := range invalidNames {
			assert.False(t, isValidFunctionName(name), "Expected %q to be invalid", name)
		}
	})

	t.Run("boundary - exactly 63 characters", func(t *testing.T) {
		name := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		assert.Equal(t, 63, len(name))
		assert.True(t, isValidFunctionName(name))
	})

	t.Run("boundary - 64 characters", func(t *testing.T) {
		name := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		assert.Equal(t, 64, len(name))
		assert.False(t, isValidFunctionName(name))
	})
}
