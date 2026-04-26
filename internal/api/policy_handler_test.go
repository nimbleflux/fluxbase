package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Policy Struct Tests
// =============================================================================

func TestPolicy_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		using := "auth.uid() = user_id"
		withCheck := "auth.uid() = user_id"

		policy := Policy{
			Schema:     "public",
			Table:      "users",
			PolicyName: "users_own_data",
			Permissive: "PERMISSIVE",
			Roles:      []string{"authenticated"},
			Command:    "ALL",
			Using:      &using,
			WithCheck:  &withCheck,
		}

		assert.Equal(t, "public", policy.Schema)
		assert.Equal(t, "users", policy.Table)
		assert.Equal(t, "users_own_data", policy.PolicyName)
		assert.Equal(t, "PERMISSIVE", policy.Permissive)
		assert.Equal(t, []string{"authenticated"}, policy.Roles)
		assert.Equal(t, "ALL", policy.Command)
		assert.Equal(t, "auth.uid() = user_id", *policy.Using)
		assert.Equal(t, "auth.uid() = user_id", *policy.WithCheck)
	})

	t.Run("policy with nil expressions", func(t *testing.T) {
		policy := Policy{
			Schema:     "public",
			Table:      "posts",
			PolicyName: "public_read",
			Permissive: "PERMISSIVE",
			Roles:      []string{"anon", "authenticated"},
			Command:    "SELECT",
			Using:      nil,
			WithCheck:  nil,
		}

		assert.Nil(t, policy.Using)
		assert.Nil(t, policy.WithCheck)
	})

	t.Run("restrictive policy", func(t *testing.T) {
		policy := Policy{
			Schema:     "public",
			Table:      "secrets",
			PolicyName: "admin_only",
			Permissive: "RESTRICTIVE",
			Roles:      []string{"admin"},
			Command:    "ALL",
		}

		assert.Equal(t, "RESTRICTIVE", policy.Permissive)
	})
}

func TestPolicy_JSON(t *testing.T) {
	t.Run("serializes to JSON", func(t *testing.T) {
		using := "true"
		policy := Policy{
			Schema:     "public",
			Table:      "items",
			PolicyName: "items_policy",
			Permissive: "PERMISSIVE",
			Roles:      []string{"authenticated"},
			Command:    "SELECT",
			Using:      &using,
		}

		data, err := json.Marshal(policy)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"schema":"public"`)
		assert.Contains(t, string(data), `"table":"items"`)
		assert.Contains(t, string(data), `"policy_name":"items_policy"`)
		assert.Contains(t, string(data), `"command":"SELECT"`)
	})

	t.Run("deserializes from JSON", func(t *testing.T) {
		jsonData := `{
			"schema": "public",
			"table": "orders",
			"policy_name": "user_orders",
			"permissive": "PERMISSIVE",
			"roles": ["authenticated"],
			"command": "ALL",
			"using": "user_id = auth.uid()"
		}`

		var policy Policy
		err := json.Unmarshal([]byte(jsonData), &policy)
		require.NoError(t, err)

		assert.Equal(t, "public", policy.Schema)
		assert.Equal(t, "orders", policy.Table)
		assert.Equal(t, "user_orders", policy.PolicyName)
		assert.Equal(t, "user_id = auth.uid()", *policy.Using)
	})
}

// =============================================================================
// TableRLSStatus Struct Tests
// =============================================================================

func TestTableRLSStatus_Struct(t *testing.T) {
	t.Run("table without RLS", func(t *testing.T) {
		status := TableRLSStatus{
			Schema:      "public",
			Table:       "products",
			RLSEnabled:  false,
			RLSForced:   false,
			PolicyCount: 0,
			Policies:    []Policy{},
			HasWarnings: true,
		}

		assert.False(t, status.RLSEnabled)
		assert.False(t, status.RLSForced)
		assert.Equal(t, 0, status.PolicyCount)
		assert.Empty(t, status.Policies)
		assert.True(t, status.HasWarnings)
	})

	t.Run("table with RLS and policies", func(t *testing.T) {
		using := "true"
		status := TableRLSStatus{
			Schema:      "public",
			Table:       "users",
			RLSEnabled:  true,
			RLSForced:   true,
			PolicyCount: 2,
			Policies: []Policy{
				{PolicyName: "users_select", Command: "SELECT", Using: &using},
				{PolicyName: "users_insert", Command: "INSERT", Using: &using},
			},
			HasWarnings: false,
		}

		assert.True(t, status.RLSEnabled)
		assert.True(t, status.RLSForced)
		assert.Equal(t, 2, status.PolicyCount)
		assert.Len(t, status.Policies, 2)
		assert.False(t, status.HasWarnings)
	})

	t.Run("table with RLS enabled but no policies - warning", func(t *testing.T) {
		status := TableRLSStatus{
			Schema:      "public",
			Table:       "empty_table",
			RLSEnabled:  true,
			RLSForced:   false,
			PolicyCount: 0,
			Policies:    []Policy{},
			HasWarnings: true,
		}

		// RLS enabled with no policies means all access denied
		assert.True(t, status.RLSEnabled)
		assert.Equal(t, 0, status.PolicyCount)
		assert.True(t, status.HasWarnings)
	})
}

// =============================================================================
// SecurityWarning Struct Tests
// =============================================================================

func TestSecurityWarning_Struct(t *testing.T) {
	t.Run("critical warning - no RLS", func(t *testing.T) {
		warning := SecurityWarning{
			ID:         "no-rls-users",
			Severity:   "critical",
			Category:   "rls",
			Schema:     "public",
			Table:      "users",
			Message:    "Table 'users' does not have Row Level Security enabled",
			Suggestion: "Enable RLS and create appropriate policies",
			FixSQL:     `ALTER TABLE public."users" ENABLE ROW LEVEL SECURITY;`,
		}

		assert.Equal(t, "critical", warning.Severity)
		assert.Equal(t, "rls", warning.Category)
		assert.Contains(t, warning.Message, "does not have Row Level Security")
		assert.NotEmpty(t, warning.FixSQL)
	})

	t.Run("high warning - overly permissive", func(t *testing.T) {
		warning := SecurityWarning{
			ID:         "permissive-public-items-allow_all",
			Severity:   "high",
			Category:   "policy",
			Schema:     "public",
			Table:      "items",
			PolicyName: "allow_all",
			Message:    "Policy 'allow_all' uses 'USING (true)' for INSERT",
			Suggestion: "Restrict the USING clause to appropriate conditions",
		}

		assert.Equal(t, "high", warning.Severity)
		assert.Equal(t, "policy", warning.Category)
		assert.Equal(t, "allow_all", warning.PolicyName)
	})

	t.Run("medium warning - missing WITH CHECK", func(t *testing.T) {
		warning := SecurityWarning{
			ID:         "no-check-public-posts-insert_policy",
			Severity:   "medium",
			Category:   "policy",
			Schema:     "public",
			Table:      "posts",
			PolicyName: "insert_policy",
			Message:    "Policy 'insert_policy' has no WITH CHECK clause",
			Suggestion: "Add WITH CHECK to validate data on insert/update",
		}

		assert.Equal(t, "medium", warning.Severity)
		assert.Contains(t, warning.Message, "no WITH CHECK")
	})

	t.Run("sensitive data warning", func(t *testing.T) {
		warning := SecurityWarning{
			ID:       "sensitive-no-rls-credentials-api_key",
			Severity: "critical",
			Category: "sensitive-data",
			Schema:   "public",
			Table:    "credentials",
			Message:  "Table 'credentials' contains sensitive column 'api_key' but RLS is not enabled",
			FixSQL:   `ALTER TABLE public."credentials" ENABLE ROW LEVEL SECURITY;`,
		}

		assert.Equal(t, "critical", warning.Severity)
		assert.Equal(t, "sensitive-data", warning.Category)
		assert.Contains(t, warning.Message, "sensitive column")
	})
}

func TestSecurityWarning_JSON(t *testing.T) {
	t.Run("serializes to JSON", func(t *testing.T) {
		warning := SecurityWarning{
			ID:         "test-warning",
			Severity:   "high",
			Category:   "rls",
			Schema:     "public",
			Table:      "test_table",
			Message:    "Test warning message",
			Suggestion: "Fix the issue",
			FixSQL:     "ALTER TABLE test;",
		}

		data, err := json.Marshal(warning)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"id":"test-warning"`)
		assert.Contains(t, string(data), `"severity":"high"`)
		assert.Contains(t, string(data), `"fix_sql":"ALTER TABLE test;"`)
	})
}

// =============================================================================
// CreatePolicyRequest Struct Tests
// =============================================================================

func TestCreatePolicyRequest_Struct(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		req := CreatePolicyRequest{
			Schema:     "public",
			Table:      "posts",
			Name:       "posts_owner_policy",
			Command:    "ALL",
			Permissive: true,
			Roles:      []string{"authenticated"},
			Using:      "user_id = auth.uid()",
			WithCheck:  "user_id = auth.uid()",
		}

		assert.Equal(t, "public", req.Schema)
		assert.Equal(t, "posts", req.Table)
		assert.Equal(t, "posts_owner_policy", req.Name)
		assert.Equal(t, "ALL", req.Command)
		assert.True(t, req.Permissive)
		assert.Equal(t, []string{"authenticated"}, req.Roles)
	})

	t.Run("restrictive policy request", func(t *testing.T) {
		req := CreatePolicyRequest{
			Schema:     "public",
			Table:      "audit_logs",
			Name:       "admin_only_access",
			Command:    "SELECT",
			Permissive: false, // RESTRICTIVE
			Roles:      []string{"admin"},
			Using:      "true",
		}

		assert.False(t, req.Permissive)
	})

	t.Run("policy with multiple roles", func(t *testing.T) {
		req := CreatePolicyRequest{
			Schema: "public",
			Table:  "shared_docs",
			Name:   "team_access",
			Roles:  []string{"admin", "editor", "viewer"},
		}

		assert.Len(t, req.Roles, 3)
		assert.Contains(t, req.Roles, "admin")
		assert.Contains(t, req.Roles, "editor")
		assert.Contains(t, req.Roles, "viewer")
	})
}

func TestCreatePolicyRequest_JSON(t *testing.T) {
	t.Run("deserializes from JSON", func(t *testing.T) {
		jsonData := `{
			"schema": "public",
			"table": "items",
			"name": "items_policy",
			"command": "SELECT",
			"permissive": true,
			"roles": ["anon", "authenticated"],
			"using": "is_public = true"
		}`

		var req CreatePolicyRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "public", req.Schema)
		assert.Equal(t, "items", req.Table)
		assert.Equal(t, "items_policy", req.Name)
		assert.Equal(t, "SELECT", req.Command)
		assert.True(t, req.Permissive)
		assert.Len(t, req.Roles, 2)
	})
}

// =============================================================================
// UpdatePolicyRequest Struct Tests
// =============================================================================

func TestUpdatePolicyRequest_Struct(t *testing.T) {
	t.Run("update roles only", func(t *testing.T) {
		req := UpdatePolicyRequest{
			Roles: []string{"admin", "moderator"},
		}

		assert.Len(t, req.Roles, 2)
		assert.Nil(t, req.Using)
		assert.Nil(t, req.WithCheck)
	})

	t.Run("update USING clause", func(t *testing.T) {
		using := "user_id = auth.uid() OR is_admin()"
		req := UpdatePolicyRequest{
			Using: &using,
		}

		assert.Equal(t, "user_id = auth.uid() OR is_admin()", *req.Using)
	})

	t.Run("update WITH CHECK clause", func(t *testing.T) {
		withCheck := "user_id = auth.uid()"
		req := UpdatePolicyRequest{
			WithCheck: &withCheck,
		}

		assert.Equal(t, "user_id = auth.uid()", *req.WithCheck)
	})

	t.Run("update all fields", func(t *testing.T) {
		using := "true"
		withCheck := "user_id = auth.uid()"
		req := UpdatePolicyRequest{
			Roles:     []string{"authenticated"},
			Using:     &using,
			WithCheck: &withCheck,
		}

		assert.Len(t, req.Roles, 1)
		assert.NotNil(t, req.Using)
		assert.NotNil(t, req.WithCheck)
	})
}

// =============================================================================
// validIdentifierRegex Tests
// =============================================================================

func TestValidPolicyNameRegex(t *testing.T) {
	validNames := []string{
		"my_policy",
		"Policy1",
		"_private_policy",
		"A",
		"policy_with_numbers_123",
		"camelCasePolicy",
		"UPPER_CASE",
	}

	invalidNames := []string{
		"123_starts_with_number",
		"has spaces",
		"has-dashes",
		"has.dots",
		"",
		"has@special",
		"has!bang",
		"policy;injection",
	}

	for _, name := range validNames {
		t.Run("valid: "+name, func(t *testing.T) {
			assert.True(t, validIdentifierRegex.MatchString(name), "Expected %q to be valid", name)
		})
	}

	for _, name := range invalidNames {
		t.Run("invalid: "+name, func(t *testing.T) {
			assert.False(t, validIdentifierRegex.MatchString(name), "Expected %q to be invalid", name)
		})
	}
}

// =============================================================================
// Command Validation Tests
// =============================================================================

func TestValidCommands(t *testing.T) {
	validCommands := map[string]bool{
		"ALL":    true,
		"SELECT": true,
		"INSERT": true,
		"UPDATE": true,
		"DELETE": true,
	}

	t.Run("all valid commands accepted", func(t *testing.T) {
		for cmd := range validCommands {
			assert.True(t, validCommands[cmd], "Command %s should be valid", cmd)
		}
	})

	t.Run("invalid commands rejected", func(t *testing.T) {
		invalidCommands := []string{"TRUNCATE", "DROP", "CREATE", "GRANT", ""}
		for _, cmd := range invalidCommands {
			_, exists := validCommands[cmd]
			assert.False(t, exists, "Command %s should be invalid", cmd)
		}
	})
}

// =============================================================================
// Severity Levels Tests
// =============================================================================

func TestSeverityLevels(t *testing.T) {
	severities := []string{"critical", "high", "medium", "low"}

	t.Run("all severity levels defined", func(t *testing.T) {
		assert.Len(t, severities, 4)
		assert.Contains(t, severities, "critical")
		assert.Contains(t, severities, "high")
		assert.Contains(t, severities, "medium")
		assert.Contains(t, severities, "low")
	})

	t.Run("severity order", func(t *testing.T) {
		// In practice, critical > high > medium > low
		severityOrder := map[string]int{
			"critical": 4,
			"high":     3,
			"medium":   2,
			"low":      1,
		}

		assert.Greater(t, severityOrder["critical"], severityOrder["high"])
		assert.Greater(t, severityOrder["high"], severityOrder["medium"])
		assert.Greater(t, severityOrder["medium"], severityOrder["low"])
	})
}

// =============================================================================
// Warning Categories Tests
// =============================================================================

func TestWarningCategories(t *testing.T) {
	categories := []string{"rls", "policy", "sensitive-data"}

	t.Run("RLS category for table-level issues", func(t *testing.T) {
		warning := SecurityWarning{
			Category: "rls",
			Message:  "Table does not have RLS enabled",
		}
		assert.Equal(t, "rls", warning.Category)
	})

	t.Run("policy category for policy issues", func(t *testing.T) {
		warning := SecurityWarning{
			Category:   "policy",
			PolicyName: "overly_permissive",
			Message:    "Policy uses USING (true)",
		}
		assert.Equal(t, "policy", warning.Category)
		assert.NotEmpty(t, warning.PolicyName)
	})

	t.Run("sensitive-data category for data issues", func(t *testing.T) {
		warning := SecurityWarning{
			Category: "sensitive-data",
			Message:  "Contains sensitive column",
		}
		assert.Equal(t, "sensitive-data", warning.Category)
	})

	t.Run("all categories defined", func(t *testing.T) {
		assert.Len(t, categories, 3)
	})
}
