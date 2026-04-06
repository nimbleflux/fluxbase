package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

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

func TestFunctionConfig_Struct(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		var config FunctionConfig
		assert.Empty(t, config.Description)
		assert.False(t, config.IsPublic)
		assert.False(t, config.AllowUnauthenticated)
		assert.Equal(t, 0, config.Timeout)
		assert.Equal(t, 0, config.Memory)
		assert.False(t, config.AllowNet)
		assert.False(t, config.AllowEnv)
	})

	t.Run("all fields", func(t *testing.T) {
		config := FunctionConfig{
			Description:          "Test function",
			IsPublic:             true,
			AllowUnauthenticated: true,
			Timeout:              60,
			Memory:               256,
			AllowNet:             true,
			AllowEnv:             true,
			DisableLogs:          true,
			CorsOrigins:          "*",
			RateLimitPerMinute:   100,
			RateLimitPerHour:     1000,
			RateLimitPerDay:      10000,
		}

		assert.Equal(t, "Test function", config.Description)
		assert.True(t, config.IsPublic)
		assert.Equal(t, 60, config.Timeout)
		assert.Equal(t, 256, config.Memory)
		assert.Equal(t, 100, config.RateLimitPerMinute)
	})
}

func TestParseFluxbaseAnnotations(t *testing.T) {
	t.Run("empty code returns defaults", func(t *testing.T) {
		config := parseFluxbaseAnnotations("")
		assert.Equal(t, 30, config.Timeout)
		assert.Equal(t, 128, config.Memory)
		assert.True(t, config.AllowNet)
		assert.True(t, config.AllowEnv)
		assert.False(t, config.IsPublic)
		assert.False(t, config.AllowUnauthenticated)
	})

	t.Run("public annotation", func(t *testing.T) {
		code := `// @fluxbase:public
export default function() {}`
		config := parseFluxbaseAnnotations(code)
		assert.True(t, config.IsPublic)
	})

	t.Run("allow-unauthenticated annotation", func(t *testing.T) {
		code := `// @fluxbase:allow-unauthenticated
export default function() {}`
		config := parseFluxbaseAnnotations(code)
		assert.True(t, config.AllowUnauthenticated)
	})

	t.Run("description annotation", func(t *testing.T) {
		code := `// @fluxbase:description Handle user registration and validation
export default function() {}`
		config := parseFluxbaseAnnotations(code)
		assert.Equal(t, "Handle user registration and validation", config.Description)
	})

	t.Run("timeout annotation", func(t *testing.T) {
		code := `// @fluxbase:timeout 60
export default function() {}`
		config := parseFluxbaseAnnotations(code)
		assert.Equal(t, 60, config.Timeout)
	})

	t.Run("timeout annotation - invalid value", func(t *testing.T) {
		code := `// @fluxbase:timeout invalid
export default function() {}`
		config := parseFluxbaseAnnotations(code)
		assert.Equal(t, 30, config.Timeout) // default
	})

	t.Run("memory annotation", func(t *testing.T) {
		code := `// @fluxbase:memory 256
export default function() {}`
		config := parseFluxbaseAnnotations(code)
		assert.Equal(t, 256, config.Memory)
	})

	t.Run("cors-origins annotation", func(t *testing.T) {
		code := `// @fluxbase:cors-origins https://example.com,https://app.example.com
export default function() {}`
		config := parseFluxbaseAnnotations(code)
		assert.Equal(t, "https://example.com,https://app.example.com", config.CorsOrigins)
	})

	t.Run("deny-net annotation", func(t *testing.T) {
		code := `// @fluxbase:deny-net
export default function() {}`
		config := parseFluxbaseAnnotations(code)
		assert.False(t, config.AllowNet)
	})

	t.Run("deny-env annotation", func(t *testing.T) {
		code := `// @fluxbase:deny-env
export default function() {}`
		config := parseFluxbaseAnnotations(code)
		assert.False(t, config.AllowEnv)
	})

	t.Run("disable-logs annotation", func(t *testing.T) {
		code := `// @fluxbase:disable-logs
export default function() {}`
		config := parseFluxbaseAnnotations(code)
		assert.True(t, config.DisableLogs)
	})

	t.Run("rate-limit per minute", func(t *testing.T) {
		code := `// @fluxbase:rate-limit 100/min
export default function() {}`
		config := parseFluxbaseAnnotations(code)
		assert.Equal(t, 100, config.RateLimitPerMinute)
		assert.Equal(t, 0, config.RateLimitPerHour)
		assert.Equal(t, 0, config.RateLimitPerDay)
	})

	t.Run("rate-limit per hour", func(t *testing.T) {
		code := `// @fluxbase:rate-limit 1000/hour
export default function() {}`
		config := parseFluxbaseAnnotations(code)
		assert.Equal(t, 1000, config.RateLimitPerHour)
	})

	t.Run("rate-limit per day", func(t *testing.T) {
		code := `// @fluxbase:rate-limit 10000/day
export default function() {}`
		config := parseFluxbaseAnnotations(code)
		assert.Equal(t, 10000, config.RateLimitPerDay)
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
		config := parseFluxbaseAnnotations(code)
		assert.Equal(t, "API endpoint", config.Description)
		assert.True(t, config.IsPublic)
		assert.True(t, config.AllowUnauthenticated)
		assert.Equal(t, 45, config.Timeout)
		assert.Equal(t, 512, config.Memory)
		assert.Equal(t, 50, config.RateLimitPerMinute)
		assert.Equal(t, "*", config.CorsOrigins)
	})

	t.Run("block comment style", func(t *testing.T) {
		code := `/*
 * @fluxbase:description Block comment function
 * @fluxbase:public
 * @fluxbase:timeout 120
 */
export default function() {}`
		config := parseFluxbaseAnnotations(code)
		assert.Equal(t, "Block comment function", config.Description)
		assert.True(t, config.IsPublic)
		assert.Equal(t, 120, config.Timeout)
	})

	t.Run("case insensitivity", func(t *testing.T) {
		code := `// @fluxbase:PUBLIC
// @fluxbase:TIMEOUT 60
// @fluxbase:Description Test`
		config := parseFluxbaseAnnotations(code)
		assert.True(t, config.IsPublic)
		assert.Equal(t, 60, config.Timeout)
		assert.Equal(t, "Test", config.Description)
	})
}

func TestParseRateLimit(t *testing.T) {
	t.Run("per minute", func(t *testing.T) {
		config := FunctionConfig{}
		parseRateLimit("100/min", &config)
		assert.Equal(t, 100, config.RateLimitPerMinute)
	})

	t.Run("per minute - alternate format", func(t *testing.T) {
		config := FunctionConfig{}
		parseRateLimit("50/minute", &config)
		assert.Equal(t, 50, config.RateLimitPerMinute)
	})

	t.Run("per hour", func(t *testing.T) {
		config := FunctionConfig{}
		parseRateLimit("1000/hour", &config)
		assert.Equal(t, 1000, config.RateLimitPerHour)
	})

	t.Run("per hour - alternate format", func(t *testing.T) {
		config := FunctionConfig{}
		parseRateLimit("500/hr", &config)
		assert.Equal(t, 500, config.RateLimitPerHour)
	})

	t.Run("per day", func(t *testing.T) {
		config := FunctionConfig{}
		parseRateLimit("10000/day", &config)
		assert.Equal(t, 10000, config.RateLimitPerDay)
	})

	t.Run("invalid format - no slash", func(t *testing.T) {
		config := FunctionConfig{}
		parseRateLimit("100", &config)
		assert.Equal(t, 0, config.RateLimitPerMinute)
		assert.Equal(t, 0, config.RateLimitPerHour)
		assert.Equal(t, 0, config.RateLimitPerDay)
	})

	t.Run("invalid format - non-numeric count", func(t *testing.T) {
		config := FunctionConfig{}
		parseRateLimit("abc/min", &config)
		assert.Equal(t, 0, config.RateLimitPerMinute)
	})

	t.Run("invalid format - zero count", func(t *testing.T) {
		config := FunctionConfig{}
		parseRateLimit("0/min", &config)
		assert.Equal(t, 0, config.RateLimitPerMinute)
	})

	t.Run("invalid format - unknown period", func(t *testing.T) {
		config := FunctionConfig{}
		parseRateLimit("100/week", &config)
		assert.Equal(t, 0, config.RateLimitPerMinute)
		assert.Equal(t, 0, config.RateLimitPerHour)
		assert.Equal(t, 0, config.RateLimitPerDay)
	})

	t.Run("handles whitespace", func(t *testing.T) {
		config := FunctionConfig{}
		parseRateLimit(" 100 / min ", &config)
		assert.Equal(t, 100, config.RateLimitPerMinute)
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
			"",            // empty
			"1function",   // starts with number
			"-function",   // starts with hyphen
			"my function", // contains space
			"my.function", // contains dot
			"my@function", // contains special char
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // > 63 chars
		}

		for _, name := range invalidNames {
			assert.False(t, isValidFunctionName(name), "Expected %q to be invalid", name)
		}
	})

	t.Run("boundary - exactly 63 characters", func(t *testing.T) {
		name := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // 63 chars
		assert.Equal(t, 63, len(name))
		assert.True(t, isValidFunctionName(name))
	})

	t.Run("boundary - 64 characters", func(t *testing.T) {
		name := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // 64 chars
		assert.Equal(t, 64, len(name))
		assert.False(t, isValidFunctionName(name))
	})
}
