package config

import (
	"fmt"
	"time"
)

// MCPOAuthConfig contains OAuth 2.1 settings for MCP authentication
type MCPOAuthConfig struct {
	Enabled             bool          `mapstructure:"enabled"`               // Enable OAuth for MCP (zero-config Claude Desktop integration)
	DCREnabled          bool          `mapstructure:"dcr_enabled"`           // Enable Dynamic Client Registration (RFC 7591)
	AllowedRedirectURIs []string      `mapstructure:"allowed_redirect_uris"` // Allowed OAuth redirect URIs
	TokenExpiry         time.Duration `mapstructure:"token_expiry"`          // Access token expiry (default: 1h)
	RefreshTokenExpiry  time.Duration `mapstructure:"refresh_token_expiry"`  // Refresh token expiry (default: 168h / 7 days)
}

// MCPConfig contains Model Context Protocol server settings
type MCPConfig struct {
	Enabled          bool          `mapstructure:"enabled"`            // Enable MCP server endpoint
	BasePath         string        `mapstructure:"base_path"`          // Base path for MCP endpoints (default: "/mcp")
	SessionTimeout   time.Duration `mapstructure:"session_timeout"`    // Session timeout for stateful connections
	MaxMessageSize   int           `mapstructure:"max_message_size"`   // Maximum message size in bytes
	AllowedTools     []string      `mapstructure:"allowed_tools"`      // Allowed tool names (empty = all enabled)
	AllowedResources []string      `mapstructure:"allowed_resources"`  // Allowed resource URIs (empty = all enabled)
	RateLimitPerMin  int           `mapstructure:"rate_limit_per_min"` // Rate limit per minute per client key

	// Custom MCP tools configuration
	ToolsDir       string `mapstructure:"tools_dir"`         // Directory for custom MCP tool files (default: "/app/mcp-tools")
	AutoLoadOnBoot bool   `mapstructure:"auto_load_on_boot"` // Auto-load custom tools from ToolsDir on startup

	// OAuth 2.1 configuration for MCP
	OAuth MCPOAuthConfig `mapstructure:"oauth"`
}

// DefaultMCPOAuthRedirectURIs returns the default allowed redirect URIs for popular MCP clients
func DefaultMCPOAuthRedirectURIs() []string {
	return []string{
		// Claude Desktop / Claude Code
		"https://claude.ai/api/mcp/auth_callback",
		"https://claude.com/api/mcp/auth_callback",
		// Cursor
		"cursor://anysphere.cursor-mcp/oauth/*/callback",
		"cursor://",
		// VS Code
		"http://127.0.0.1:33418",
		"https://vscode.dev/redirect",
		"vscode://",
		// OpenCode
		"http://127.0.0.1:19876/mcp/oauth/callback",
		// MCP Inspector (development)
		"http://localhost:6274/oauth/callback",
		// ChatGPT
		"https://chatgpt.com/connector_platform_oauth_redirect",
		// Localhost wildcards (development)
		"http://localhost:*",
		"http://127.0.0.1:*",
	}
}

// Validate validates MCP configuration
func (mc *MCPConfig) Validate() error {
	if !mc.Enabled {
		return nil // No validation needed if disabled
	}

	if mc.BasePath == "" {
		return fmt.Errorf("mcp base_path cannot be empty when enabled")
	}

	if mc.SessionTimeout < 0 {
		return fmt.Errorf("mcp session_timeout cannot be negative, got: %v", mc.SessionTimeout)
	}

	if mc.MaxMessageSize < 0 {
		return fmt.Errorf("mcp max_message_size cannot be negative, got: %d", mc.MaxMessageSize)
	}

	if mc.RateLimitPerMin < 0 {
		return fmt.Errorf("mcp rate_limit_per_min cannot be negative, got: %d", mc.RateLimitPerMin)
	}

	// Validate OAuth config if enabled
	if mc.OAuth.Enabled {
		if mc.OAuth.TokenExpiry < 0 {
			return fmt.Errorf("mcp oauth token_expiry cannot be negative, got: %v", mc.OAuth.TokenExpiry)
		}
		if mc.OAuth.RefreshTokenExpiry < 0 {
			return fmt.Errorf("mcp oauth refresh_token_expiry cannot be negative, got: %v", mc.OAuth.RefreshTokenExpiry)
		}
	}

	return nil
}

// SetOAuthDefaults sets default values for OAuth configuration
func (mc *MCPConfig) SetOAuthDefaults() {
	if mc.OAuth.TokenExpiry == 0 {
		mc.OAuth.TokenExpiry = time.Hour // 1 hour default
	}
	if mc.OAuth.RefreshTokenExpiry == 0 {
		mc.OAuth.RefreshTokenExpiry = 168 * time.Hour // 7 days default
	}
	if len(mc.OAuth.AllowedRedirectURIs) == 0 {
		mc.OAuth.AllowedRedirectURIs = DefaultMCPOAuthRedirectURIs()
	}
}
