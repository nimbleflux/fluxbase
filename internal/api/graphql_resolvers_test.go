package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// GraphQLResolverFactory Construction Tests
// =============================================================================

func TestNewGraphQLResolverFactory(t *testing.T) {
	t.Run("creates factory with nil dependencies", func(t *testing.T) {
		factory := NewGraphQLResolverFactory(nil, nil)
		assert.NotNil(t, factory)
		assert.Nil(t, factory.db)
		assert.Nil(t, factory.schemaCache)
	})
}

// =============================================================================
// RLSContext Tests
// =============================================================================

func TestRLSContext_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		ctx := &RLSContext{
			UserID: "550e8400-e29b-41d4-a716-446655440000",
			Role:   "authenticated",
			Claims: map[string]interface{}{
				"email": "user@example.com",
				"sub":   "550e8400-e29b-41d4-a716-446655440000",
			},
		}

		assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", ctx.UserID)
		assert.Equal(t, "authenticated", ctx.Role)
		assert.Equal(t, "user@example.com", ctx.Claims["email"])
	})

	t.Run("empty context", func(t *testing.T) {
		ctx := &RLSContext{
			Claims: make(map[string]interface{}),
		}

		assert.Empty(t, ctx.UserID)
		assert.Empty(t, ctx.Role)
		assert.Empty(t, ctx.Claims)
	})

	t.Run("nil claims map", func(t *testing.T) {
		ctx := &RLSContext{
			UserID: "test-user",
			Role:   "admin",
			Claims: nil,
		}

		assert.Equal(t, "test-user", ctx.UserID)
		assert.Nil(t, ctx.Claims)
	})
}

// =============================================================================
// mapAppRoleToDatabaseRole Tests
// =============================================================================

func TestMapAppRoleToDatabaseRole(t *testing.T) {
	tests := []struct {
		name     string
		appRole  string
		expected string
	}{
		{
			name:     "service_role maps to service_role",
			appRole:  "service_role",
			expected: "service_role",
		},
		{
			name:     "instance_admin maps to service_role",
			appRole:  "instance_admin",
			expected: "service_role",
		},
		{
			name:     "anon maps to anon",
			appRole:  "anon",
			expected: "anon",
		},
		{
			name:     "empty string maps to anon",
			appRole:  "",
			expected: "anon",
		},
		{
			name:     "authenticated maps to authenticated",
			appRole:  "authenticated",
			expected: "authenticated",
		},
		{
			name:     "user maps to authenticated",
			appRole:  "user",
			expected: "authenticated",
		},
		{
			name:     "admin maps to authenticated",
			appRole:  "admin",
			expected: "authenticated",
		},
		{
			name:     "unknown role maps to authenticated",
			appRole:  "custom_role",
			expected: "authenticated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapAppRoleToDatabaseRole(tt.appRole)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// negateOperator Tests
// =============================================================================

func TestNegateOperator(t *testing.T) {
	tests := []struct {
		name     string
		op       FilterOperator
		expected FilterOperator
	}{
		{
			name:     "equal becomes not equal",
			op:       OpEqual,
			expected: OpNotEqual,
		},
		{
			name:     "not equal becomes equal",
			op:       OpNotEqual,
			expected: OpEqual,
		},
		{
			name:     "greater than becomes less or equal",
			op:       OpGreaterThan,
			expected: OpLessOrEqual,
		},
		{
			name:     "greater or equal becomes less than",
			op:       OpGreaterOrEqual,
			expected: OpLessThan,
		},
		{
			name:     "less than becomes greater or equal",
			op:       OpLessThan,
			expected: OpGreaterOrEqual,
		},
		{
			name:     "less or equal becomes greater than",
			op:       OpLessOrEqual,
			expected: OpGreaterThan,
		},
		{
			name:     "in becomes not in",
			op:       OpIn,
			expected: OpNotIn,
		},
		{
			name:     "not in becomes in",
			op:       OpNotIn,
			expected: OpIn,
		},
		{
			name:     "is becomes is not",
			op:       OpIs,
			expected: OpIsNot,
		},
		{
			name:     "is not becomes is",
			op:       OpIsNot,
			expected: OpIs,
		},
		{
			name:     "contains stays contains (special handling)",
			op:       OpContains,
			expected: OpContains,
		},
		{
			name:     "like stays like (no negate)",
			op:       OpLike,
			expected: OpLike,
		},
		{
			name:     "ilike stays ilike (no negate)",
			op:       OpILike,
			expected: OpILike,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := negateOperator(tt.op)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// GraphQL Context Key Tests
// =============================================================================

func TestGraphQLContextKeys(t *testing.T) {
	t.Run("RLS context key is defined", func(t *testing.T) {
		assert.Equal(t, graphqlContextKey("graphql_rls_context"), GraphQLRLSContextKey)
	})
}

// =============================================================================
// Role Mapping Edge Cases
// =============================================================================

func TestMapAppRoleToDatabaseRole_EdgeCases(t *testing.T) {
	t.Run("whitespace string", func(t *testing.T) {
		result := mapAppRoleToDatabaseRole("   ")
		assert.Equal(t, "authenticated", result)
	})

	t.Run("case sensitivity", func(t *testing.T) {
		// The function is case-sensitive
		result := mapAppRoleToDatabaseRole("SERVICE_ROLE")
		assert.Equal(t, "authenticated", result)

		result = mapAppRoleToDatabaseRole("Anon")
		assert.Equal(t, "authenticated", result)
	})

	t.Run("mixed case instance_admin", func(t *testing.T) {
		result := mapAppRoleToDatabaseRole("Instance_Admin")
		assert.Equal(t, "authenticated", result)
	})
}

// =============================================================================
// Integration Scenarios
// =============================================================================

func TestRLSContext_IntegrationScenarios(t *testing.T) {
	t.Run("anonymous user scenario", func(t *testing.T) {
		ctx := &RLSContext{
			UserID: "",
			Role:   "anon",
			Claims: make(map[string]interface{}),
		}

		dbRole := mapAppRoleToDatabaseRole(ctx.Role)
		assert.Equal(t, "anon", dbRole)
	})

	t.Run("authenticated user scenario", func(t *testing.T) {
		ctx := &RLSContext{
			UserID: "user-123",
			Role:   "authenticated",
			Claims: map[string]interface{}{
				"email": "user@example.com",
				"role":  "authenticated",
			},
		}

		dbRole := mapAppRoleToDatabaseRole(ctx.Role)
		assert.Equal(t, "authenticated", dbRole)
	})

	t.Run("service key scenario", func(t *testing.T) {
		ctx := &RLSContext{
			UserID: "",
			Role:   "service_role",
			Claims: map[string]interface{}{
				"role": "service_role",
			},
		}

		dbRole := mapAppRoleToDatabaseRole(ctx.Role)
		assert.Equal(t, "service_role", dbRole)
	})

	t.Run("dashboard admin scenario", func(t *testing.T) {
		ctx := &RLSContext{
			UserID: "admin-123",
			Role:   "instance_admin",
			Claims: map[string]interface{}{
				"email": "admin@example.com",
				"role":  "instance_admin",
			},
		}

		dbRole := mapAppRoleToDatabaseRole(ctx.Role)
		assert.Equal(t, "service_role", dbRole)
	})
}

// =============================================================================
// Filter Operator Completeness Tests
// =============================================================================

func TestFilterOperatorNegation_Completeness(t *testing.T) {
	t.Run("symmetric negation pairs", func(t *testing.T) {
		// Test that negating twice returns the original
		symmetricOps := []FilterOperator{
			OpEqual,
			OpNotEqual,
			OpGreaterThan,
			OpGreaterOrEqual,
			OpLessThan,
			OpLessOrEqual,
			OpIn,
			OpNotIn,
			OpIs,
			OpIsNot,
		}

		for _, op := range symmetricOps {
			negated := negateOperator(op)
			doubleNegated := negateOperator(negated)
			assert.Equal(t, op, doubleNegated, "Double negation of %v should return original", op)
		}
	})
}
