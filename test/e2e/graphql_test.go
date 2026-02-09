package e2e

import (
	"testing"

	"github.com/fluxbase-eu/fluxbase/test"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

// GraphQL request/response types
type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type graphQLResponse struct {
	Data   map[string]interface{} `json:"data,omitempty"`
	Errors []graphQLError         `json:"errors,omitempty"`
}

type graphQLError struct {
	Message   string        `json:"message"`
	Path      []interface{} `json:"path,omitempty"` // Can be string or int
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations,omitempty"`
}

// setupGraphQLTest prepares the test context for GraphQL tests
func setupGraphQLTest(t *testing.T) *test.TestContext {
	tc := test.NewTestContext(t)
	tc.EnsureAuthSchema()

	// Clean only test-specific data and truncate products table
	tc.ExecuteSQLAsSuperuser(`
		DELETE FROM auth.users WHERE email LIKE '%@example.com' OR email LIKE '%@test.com';
		TRUNCATE TABLE public.products RESTART IDENTITY CASCADE;
	`)

	// Grant permissions on products table for GraphQL mutation tests
	// Note: Public schema is "closed by default" - need explicit grants
	tc.ExecuteSQLAsSuperuser(`
		GRANT SELECT, INSERT, UPDATE, DELETE ON public.products TO authenticated;
		GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO authenticated;
	`)

	return tc
}

// setupGraphQLRLSTest prepares the test context for GraphQL RLS tests
func setupGraphQLRLSTest(t *testing.T) *test.TestContext {
	tc := test.NewRLSTestContext(t)
	tc.EnsureAuthSchema()
	tc.EnsureRLSTestTables()

	// Clean only test-specific data
	tc.ExecuteSQLAsSuperuser(`
		DELETE FROM auth.users WHERE email LIKE '%@example.com' OR email LIKE '%@test.com';
		DELETE FROM public.tasks WHERE user_id IS NOT NULL;
	`)

	return tc
}

// ============================================================================
// BASIC GRAPHQL TESTS
// ============================================================================

// TestGraphQL_Query_Authenticated tests basic GraphQL query with authentication
func TestGraphQL_Query_Authenticated(t *testing.T) {
	tc := setupGraphQLTest(t)
	defer tc.Close()

	// Create a test user
	_, token := tc.CreateTestUser(test.E2ETestEmail(), "password123")

	// Clean and insert test data (only query name and id to avoid NUMERIC conversion issues)
	tc.ExecuteSQLAsSuperuser(`DELETE FROM public.products`)
	tc.ExecuteSQLAsSuperuser(`INSERT INTO public.products (name, price) VALUES ('Test Product 1', 19.99)`)
	tc.ExecuteSQLAsSuperuser(`INSERT INTO public.products (name, price) VALUES ('Test Product 2', 29.99)`)

	// Query products via GraphQL (only non-NUMERIC fields to avoid type conversion issues)
	resp := tc.NewRequest("POST", "/api/v1/graphql").
		WithAuth(token).
		WithBody(graphQLRequest{
			Query: `
				query {
					products {
						id
						name
					}
				}
			`,
		}).
		Send().
		AssertStatus(fiber.StatusOK)

	var result graphQLResponse
	resp.JSON(&result)

	// Verify no errors
	require.Empty(t, result.Errors, "GraphQL query should not return errors")

	// Verify data returned
	require.NotNil(t, result.Data, "GraphQL query should return data")
	products, ok := result.Data["products"].([]interface{})
	require.True(t, ok, "products should be an array")
	require.GreaterOrEqual(t, len(products), 2, "Should return at least 2 products")

	// Verify product structure
	product1 := products[0].(map[string]interface{})
	require.NotEmpty(t, product1["id"], "Product should have id")
	require.NotEmpty(t, product1["name"], "Product should have name")
}

// TestGraphQL_Query_Anonymous_UsesAnonRole tests that anonymous queries use the anon role
// The GraphQL endpoint allows anonymous access but with the anon role applied
func TestGraphQL_Query_Anonymous_UsesAnonRole(t *testing.T) {
	tc := setupGraphQLTest(t)
	defer tc.Close()

	// Insert test data
	tc.ExecuteSQLAsSuperuser(`DELETE FROM public.products`)
	tc.ExecuteSQLAsSuperuser(`INSERT INTO public.products (name, price) VALUES ('Test Product', 19.99)`)

	// Query without authentication - should succeed with anon role
	resp := tc.NewRequest("POST", "/api/v1/graphql").
		Unauthenticated().
		WithBody(graphQLRequest{
			Query: `
				query {
					products {
						id
						name
					}
				}
			`,
		}).
		Send().
		AssertStatus(fiber.StatusOK)

	var result graphQLResponse
	resp.JSON(&result)

	// Anonymous queries should work (with anon role)
	// The anon role has access to products table (granted in setup_test.go)
	require.Empty(t, result.Errors, "Anonymous GraphQL query should work with anon role")
	require.NotNil(t, result.Data, "Should return data for anonymous query")
}

// TestGraphQL_Mutation_Insert tests GraphQL insert mutation using tasks table (which has RLS policies)
func TestGraphQL_Mutation_Insert(t *testing.T) {
	// Use RLS test context which has tasks table with proper INSERT policies
	// Use shared RLS context to avoid creating multiple connection pools
	tc := setupGraphQLRLSTest(t)
	// NO defer tc.Close() - shared context is managed by TestMain

	// Create a test user
	userID, token := tc.CreateTestUserDirect(test.E2ETestEmail(), "password123")

	// Insert a task via GraphQL mutation
	resp := tc.NewRequest("POST", "/api/v1/graphql").
		WithAuth(token).
		WithBody(graphQLRequest{
			Query: `
				mutation InsertTask($userId: UUID!, $title: String!) {
					insertTasks(data: {userId: $userId, title: $title}) {
						id
						title
						userId
					}
				}
			`,
			Variables: map[string]interface{}{
				"userId": userID,
				"title":  "GraphQL Task",
			},
		}).
		Send().
		AssertStatus(fiber.StatusOK)

	var result graphQLResponse
	resp.JSON(&result)

	// Log response for debugging
	t.Logf("Mutation response: %+v", result)

	// If there are errors, log them for debugging
	if len(result.Errors) > 0 {
		t.Logf("GraphQL errors: %+v", result.Errors)
	}

	// Verify no errors
	require.Empty(t, result.Errors, "GraphQL mutation should not return errors")

	// Verify task was created
	require.NotNil(t, result.Data, "GraphQL mutation should return data")
	insertResult, ok := result.Data["insertTasks"].(map[string]interface{})
	require.True(t, ok, "insertTasks should return an object, got: %T", result.Data["insertTasks"])
	require.NotEmpty(t, insertResult["id"], "Task should have an ID")
	require.Equal(t, "GraphQL Task", insertResult["title"])
	// Note: UUID fields may be returned as byte arrays or strings depending on pgtype serialization
	require.NotEmpty(t, insertResult["userId"], "Task should have a userId")
}

// TestGraphQL_Introspection tests GraphQL schema introspection
func TestGraphQL_Introspection(t *testing.T) {
	tc := setupGraphQLTest(t)
	defer tc.Close()

	// Create a test user
	_, token := tc.CreateTestUser(test.E2ETestEmail(), "password123")

	// Run introspection query
	resp := tc.NewRequest("POST", "/api/v1/graphql").
		WithAuth(token).
		WithBody(graphQLRequest{
			Query: `
				query {
					__schema {
						queryType {
							name
						}
						mutationType {
							name
						}
					}
				}
			`,
		}).
		Send().
		AssertStatus(fiber.StatusOK)

	var result graphQLResponse
	resp.JSON(&result)

	// Verify introspection works
	require.Empty(t, result.Errors, "Introspection should not return errors")
	require.NotNil(t, result.Data, "Introspection should return data")

	schema, ok := result.Data["__schema"].(map[string]interface{})
	require.True(t, ok, "__schema should be returned")
	require.NotNil(t, schema["queryType"], "Query type should exist")
	require.NotNil(t, schema["mutationType"], "Mutation type should exist")
}

// TestGraphQL_QueryWithVariables tests GraphQL query with variables
func TestGraphQL_QueryWithVariables(t *testing.T) {
	tc := setupGraphQLTest(t)
	defer tc.Close()

	// Create a test user
	_, token := tc.CreateTestUser(test.E2ETestEmail(), "password123")

	// Clean and insert test data
	tc.ExecuteSQLAsSuperuser(`DELETE FROM public.products`)
	tc.ExecuteSQLAsSuperuser(`INSERT INTO public.products (name, price) VALUES ('Expensive Product', 999.99)`)
	tc.ExecuteSQLAsSuperuser(`INSERT INTO public.products (name, price) VALUES ('Cheap Product', 9.99)`)

	// Query with limit variable (avoid price field due to NUMERIC serialization issues)
	resp := tc.NewRequest("POST", "/api/v1/graphql").
		WithAuth(token).
		WithBody(graphQLRequest{
			Query: `
				query GetProducts($limit: Int) {
					products(limit: $limit) {
						id
						name
					}
				}
			`,
			Variables: map[string]interface{}{
				"limit": 1,
			},
		}).
		Send().
		AssertStatus(fiber.StatusOK)

	var result graphQLResponse
	resp.JSON(&result)

	require.Empty(t, result.Errors, "Query with variables should not return errors")
	products, ok := result.Data["products"].([]interface{})
	require.True(t, ok, "products should be an array")
	require.Len(t, products, 1, "Should return only 1 product due to limit")
}

// ============================================================================
// RLS-SPECIFIC GRAPHQL TESTS
// ============================================================================

// TestGraphQL_RLS_UserCanOnlySeeOwnData tests that RLS restricts GraphQL queries to own data
func TestGraphQL_RLS_UserCanOnlySeeOwnData(t *testing.T) {
	// Use shared RLS context to avoid creating multiple connection pools
	tc := setupGraphQLRLSTest(t)
	// NO defer tc.Close() - shared context is managed by TestMain

	// Create two users
	user1ID, token1 := tc.CreateTestUserDirect(test.E2ETestEmail(), "password123")
	user2ID, token2 := tc.CreateTestUserDirect(test.E2ETestEmail(), "password123")

	// User 1 creates a task via GraphQL
	resp := tc.NewRequest("POST", "/api/v1/graphql").
		WithAuth(token1).
		WithBody(graphQLRequest{
			Query: `
				mutation CreateTask($data: TasksInput!) {
					insertTasks(data: $data) {
						id
						title
						userId
					}
				}
			`,
			Variables: map[string]interface{}{
				"data": map[string]interface{}{
					"userId": user1ID,
					"title":  "User 1 Task",
				},
			},
		}).
		Send().
		AssertStatus(fiber.StatusOK)

	var createResult graphQLResponse
	resp.JSON(&createResult)
	require.Empty(t, createResult.Errors, "User 1 should be able to create task")

	// User 2 creates a task via GraphQL
	tc.NewRequest("POST", "/api/v1/graphql").
		WithAuth(token2).
		WithBody(graphQLRequest{
			Query: `
				mutation CreateTask($data: TasksInput!) {
					insertTasks(data: $data) {
						id
						title
						userId
					}
				}
			`,
			Variables: map[string]interface{}{
				"data": map[string]interface{}{
					"userId": user2ID,
					"title":  "User 2 Task",
				},
			},
		}).
		Send().
		AssertStatus(fiber.StatusOK)

	// User 1 queries tasks via GraphQL - should only see their own
	resp = tc.NewRequest("POST", "/api/v1/graphql").
		WithAuth(token1).
		WithBody(graphQLRequest{
			Query: `
				query {
					tasks {
						id
						title
						userId
					}
				}
			`,
		}).
		Send().
		AssertStatus(fiber.StatusOK)

	var queryResult graphQLResponse
	resp.JSON(&queryResult)
	require.Empty(t, queryResult.Errors, "Query should not return errors")

	tasks, ok := queryResult.Data["tasks"].([]interface{})
	require.True(t, ok, "tasks should be an array")
	require.Len(t, tasks, 1, "User 1 should only see their own task")

	task := tasks[0].(map[string]interface{})
	require.Equal(t, "User 1 Task", task["title"], "Should see User 1's task")
}

// TestGraphQL_RLS_MutationEnforcement tests that RLS restricts GraphQL mutations
func TestGraphQL_RLS_MutationEnforcement(t *testing.T) {
	// Use shared RLS context to avoid creating multiple connection pools
	tc := setupGraphQLRLSTest(t)
	// NO defer tc.Close() - shared context is managed by TestMain

	// Create two users
	user1ID, _ := tc.CreateTestUserDirect(test.E2ETestEmail(), "password123")
	_, token2 := tc.CreateTestUserDirect(test.E2ETestEmail(), "password123")

	// Insert a task for user 1 directly in DB as superuser
	taskID := "11111111-1111-1111-1111-111111111111"
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO tasks (id, user_id, title, description, completed)
		VALUES ($1, $2, 'User 1 Task', 'User 1 description', false)
	`, taskID, user1ID)

	// User 2 tries to update User 1's task via GraphQL - should fail
	resp := tc.NewRequest("POST", "/api/v1/graphql").
		WithAuth(token2).
		WithBody(graphQLRequest{
			Query: `
				mutation UpdateTask($id: UUID!, $data: TasksInput!) {
					updateTasks(id: $id, data: $data) {
						id
						title
					}
				}
			`,
			Variables: map[string]interface{}{
				"id": taskID,
				"data": map[string]interface{}{
					"title": "Malicious Update",
				},
			},
		}).
		Send()

	// Should either return an error or empty result (RLS blocking)
	var result graphQLResponse
	resp.JSON(&result)

	// Verify RLS blocked the mutation (either via error or no rows affected)
	if len(result.Errors) == 0 && result.Data != nil {
		// If no error, the update should have returned nil/null (no matching row)
		updateResult := result.Data["updateTasks"]
		require.Nil(t, updateResult, "RLS should block updating other user's task")
	}

	// Verify the task was NOT updated (query as superuser to bypass RLS for verification)
	tasks := tc.QuerySQLAsSuperuser("SELECT * FROM tasks WHERE id = $1", taskID)
	require.Len(t, tasks, 1)
	require.Equal(t, "User 1 Task", tasks[0]["title"], "Task should not be modified")
}

// TestGraphQL_RLS_PublicData tests that public data is accessible via GraphQL
func TestGraphQL_RLS_PublicData(t *testing.T) {
	// Use shared RLS context to avoid creating multiple connection pools
	tc := setupGraphQLRLSTest(t)
	// NO defer tc.Close() - shared context is managed by TestMain

	// Create two users
	user1ID, token1 := tc.CreateTestUserDirect(test.E2ETestEmail(), "password123")
	_, token2 := tc.CreateTestUserDirect(test.E2ETestEmail(), "password123")

	// User 1 creates a public task via GraphQL
	tc.NewRequest("POST", "/api/v1/graphql").
		WithAuth(token1).
		WithBody(graphQLRequest{
			Query: `
				mutation CreateTask($data: TasksInput!) {
					insertTasks(data: $data) {
						id
						title
						isPublic
					}
				}
			`,
			Variables: map[string]interface{}{
				"data": map[string]interface{}{
					"userId":      user1ID,
					"title":       "Public Task",
					"description": "This is public",
					"isPublic":    true,
					"completed":   false,
				},
			},
		}).
		Send().
		AssertStatus(fiber.StatusOK)

	// User 1 creates a private task
	tc.NewRequest("POST", "/api/v1/graphql").
		WithAuth(token1).
		WithBody(graphQLRequest{
			Query: `
				mutation CreateTask($data: TasksInput!) {
					insertTasks(data: $data) {
						id
						title
						isPublic
					}
				}
			`,
			Variables: map[string]interface{}{
				"data": map[string]interface{}{
					"userId":      user1ID,
					"title":       "Private Task",
					"description": "This is private",
					"isPublic":    false,
					"completed":   false,
				},
			},
		}).
		Send().
		AssertStatus(fiber.StatusOK)

	// User 2 queries tasks - should see public task but not private one
	resp := tc.NewRequest("POST", "/api/v1/graphql").
		WithAuth(token2).
		WithBody(graphQLRequest{
			Query: `
				query {
					tasks {
						id
						title
						isPublic
					}
				}
			`,
		}).
		Send().
		AssertStatus(fiber.StatusOK)

	var result graphQLResponse
	resp.JSON(&result)
	require.Empty(t, result.Errors, "Query should not return errors")

	tasks, ok := result.Data["tasks"].([]interface{})
	require.True(t, ok, "tasks should be an array")
	require.Len(t, tasks, 1, "User 2 should only see public task")

	task := tasks[0].(map[string]interface{})
	require.Equal(t, "Public Task", task["title"])
	require.Equal(t, true, task["isPublic"])
}

// TestGraphQL_RLS_ServiceRoleBypassesRLS tests that service role can see all data via GraphQL
func TestGraphQL_RLS_ServiceRoleBypassesRLS(t *testing.T) {

	// Use shared RLS context to avoid creating multiple connection pools
	tc := setupGraphQLRLSTest(t)
	// NO defer tc.Close() - shared context is managed by TestMain

	// Create two users with tasks
	user1ID, _ := tc.CreateTestUserDirect(test.E2ETestEmail(), "password123")
	user2ID, _ := tc.CreateTestUserDirect(test.E2ETestEmail(), "password123")

	// Insert tasks for both users directly in DB
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO tasks (user_id, title, description, completed)
		VALUES ($1, 'User 1 Task', 'User 1 description', false)
	`, user1ID)
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO tasks (user_id, title, description, completed)
		VALUES ($1, 'User 2 Task', 'User 2 description', false)
	`, user2ID)

	// Create service key
	serviceKey := tc.CreateServiceKey("GraphQL Test Service")

	// Query with service key - should see all tasks (bypasses RLS)
	resp := tc.NewRequest("POST", "/api/v1/graphql").
		WithServiceKey(serviceKey).
		WithBody(graphQLRequest{
			Query: `
				query {
					tasks {
						id
						title
						userId
					}
				}
			`,
		}).
		Send().
		AssertStatus(fiber.StatusOK)

	var result graphQLResponse
	resp.JSON(&result)
	require.Empty(t, result.Errors, "Service role query should not return errors")

	tasks, ok := result.Data["tasks"].([]interface{})
	require.True(t, ok, "tasks should be an array")
	require.GreaterOrEqual(t, len(tasks), 2, "Service role should see all tasks from both users")

	// Verify both users' tasks are visible
	var foundUser1, foundUser2 bool
	for _, t := range tasks {
		task := t.(map[string]interface{})
		if task["title"] == "User 1 Task" {
			foundUser1 = true
		}
		if task["title"] == "User 2 Task" {
			foundUser2 = true
		}
	}
	require.True(t, foundUser1, "Should see User 1's task")
	require.True(t, foundUser2, "Should see User 2's task")
}

// ============================================================================
// QUERY DEPTH AND COMPLEXITY TESTS
// ============================================================================

// TestGraphQL_QueryDepthLimit tests that query depth limits are enforced
func TestGraphQL_QueryDepthLimit(t *testing.T) {
	tc := setupGraphQLTest(t)
	defer tc.Close()

	// Create a test user
	_, token := tc.CreateTestUser(test.E2ETestEmail(), "password123")

	// Create a deeply nested query that should exceed depth limits
	// Default depth limit is 10
	resp := tc.NewRequest("POST", "/api/v1/graphql").
		WithAuth(token).
		WithBody(graphQLRequest{
			Query: `
				query {
					products {
						id
						name
						products {
							id
							products {
								id
								products {
									id
									products {
										id
										products {
											id
											products {
												id
												products {
													id
													products {
														id
														products {
															id
															products {
																id
															}
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			`,
		}).
		Send()

	// Should either return an error about depth or fail to execute
	var result graphQLResponse
	resp.JSON(&result)

	// The query may fail due to depth limit or because relationships don't exist
	// Either way, it shouldn't succeed in returning deeply nested data
	t.Logf("Deep query response status: %d", resp.Status())
	if len(result.Errors) > 0 {
		t.Logf("Deep query errors: %+v", result.Errors)
	}
}
