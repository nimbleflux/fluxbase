package mcp

// SupportedScopes is the full set of MCP scopes the authorization server can
// grant. It covers every scope referenced by tool RequiredScopes
// implementations plus the custom tool/resource scopes. This is the list
// advertised in the OAuth authorization server metadata and the whitelist
// applied during dynamic client registration and authorization.
var SupportedScopes = []string{
	// Tables
	ScopeReadTables,
	ScopeWriteTables,

	// Functions / RPC
	ScopeExecuteFunctions,
	ScopeExecuteRPC,

	// Storage
	ScopeReadStorage,
	ScopeWriteStorage,

	// Jobs
	ScopeExecuteJobs,

	// Vector / AI
	ScopeReadVectors,

	// SQL and HTTP execution
	ScopeExecuteSQL,
	ScopeExecuteHTTP,

	// Schema / projects
	ScopeReadSchema,
	ScopeReadProjects,
	ScopeWriteProjects,

	// Admin
	ScopeAdminSchemas,
	ScopeAdminDDL,

	// Sync (admin-level code deployment)
	ScopeSyncFunctions,
	ScopeSyncJobs,
	ScopeSyncRPC,
	ScopeSyncMigrations,
	ScopeSyncChatbots,

	// Branching
	ScopeBranchRead,
	ScopeBranchWrite,
	ScopeBranchAccess,

	// GitHub
	ScopeGitHubRead,
	ScopeGitHubWrite,

	// Custom tools and resources
	"execute:custom",
	"read:custom",
}

var supportedScopeSet = func() map[string]bool {
	set := make(map[string]bool, len(SupportedScopes))
	for _, s := range SupportedScopes {
		set[s] = true
	}
	return set
}()

// privilegedScopePrefixes list the scope prefixes that grant dangerous
// (admin-level) capabilities. They may only be granted when the authorizing
// user holds an admin role.
var privilegedScopePrefixes = []string{"admin:", "sync:", "branch:", "github:"}

// IsSupportedScope reports whether the scope is in the supported set.
func IsSupportedScope(scope string) bool {
	return supportedScopeSet[scope]
}

// FilterSupportedScopes splits requested scopes into the supported subset and
// the unknown ones (which must be rejected with invalid_scope).
func FilterSupportedScopes(requested []string) (supported, unknown []string) {
	for _, s := range requested {
		if s == "" {
			continue
		}
		if supportedScopeSet[s] {
			supported = append(supported, s)
		} else {
			unknown = append(unknown, s)
		}
	}
	return supported, unknown
}

// IsPrivilegedScope reports whether a scope grants admin-level capabilities
// (DDL, code/migration sync, branch admin, GitHub integration).
func IsPrivilegedScope(scope string) bool {
	for _, prefix := range privilegedScopePrefixes {
		if len(scope) >= len(prefix) && scope[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// HasPrivilegedScope reports whether any of the scopes is privileged.
func HasPrivilegedScope(scopes []string) bool {
	for _, s := range scopes {
		if IsPrivilegedScope(s) {
			return true
		}
	}
	return false
}
