package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraphQLQuery_Success(t *testing.T) {
	resetGraphQLFlags()
	graphqlPretty = true

	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/graphql")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Contains(t, body["query"], "users")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"data": map[string]interface{}{
				"users": []map[string]interface{}{
					{"id": "1", "email": "test@example.com"},
					{"id": "2", "email": "other@example.com"},
				},
			},
		})
	})
	defer cleanup()

	err := runGraphQLQuery(nil, []string{"{ users { id email } }"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.NotNil(t, result["data"])
}

func TestGraphQLQuery_NoQuery(t *testing.T) {
	resetGraphQLFlags()
	graphqlFile = ""

	err := runGraphQLQuery(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "either provide a query as an argument or use --file")
}

func TestGraphQLQuery_WithVariables(t *testing.T) {
	resetGraphQLFlags()
	graphqlPretty = true
	graphqlVariables = []string{"id=123", "limit=10"}

	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.NotNil(t, body["variables"])

		vars, ok := body["variables"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, float64(123), vars["id"])
		assert.Equal(t, float64(10), vars["limit"])

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{"id": "123", "email": "user@example.com"},
			},
		})
	})
	defer cleanup()

	err := runGraphQLQuery(nil, []string{"query GetUser($id: ID!) { user(id: $id) { email } }"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.NotNil(t, result["data"])
}

func TestGraphQLQuery_GraphQLErrors(t *testing.T) {
	resetGraphQLFlags()
	graphqlPretty = true

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"errors": []map[string]interface{}{
				{"message": "Cannot query field \"nonexistent\" on type \"Query\"."},
			},
		})
	})
	defer cleanup()

	err := runGraphQLQuery(nil, []string{"{ nonexistent }"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GraphQL errors")
	assert.Contains(t, err.Error(), "Cannot query field")
}

func TestGraphQLQuery_APIError(t *testing.T) {
	resetGraphQLFlags()
	graphqlPretty = true

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusServiceUnavailable, "server overloaded")
	})
	defer cleanup()

	err := runGraphQLQuery(nil, []string{"{ users { id } }"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "server overloaded")
}

func TestGraphQLMutation_Success(t *testing.T) {
	resetGraphQLFlags()
	graphqlPretty = true

	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/graphql")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Contains(t, body["query"], "insert_users")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"data": map[string]interface{}{
				"insert_users": map[string]interface{}{
					"returning": []map[string]interface{}{
						{"id": "1", "email": "new@example.com"},
					},
				},
			},
		})
	})
	defer cleanup()

	err := runGraphQLMutation(nil, []string{"mutation { insert_users(objects: [{email: \"new@example.com\"}]) { returning { id } } }"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.NotNil(t, result["data"])
}

func TestGraphQLMutation_APIError(t *testing.T) {
	resetGraphQLFlags()
	graphqlPretty = true

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusUnauthorized, "unauthorized")
	})
	defer cleanup()

	err := runGraphQLMutation(nil, []string{"mutation { insert_users(objects: []) { returning { id } } }"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestGraphQLIntrospect_Success(t *testing.T) {
	resetGraphQLFlags()
	introspectTypesOnly = true
	graphqlPretty = true

	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/graphql")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Contains(t, body["query"], "__schema")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"data": map[string]interface{}{
				"__schema": map[string]interface{}{
					"types": []interface{}{
						map[string]interface{}{"name": "User", "kind": "OBJECT"},
						map[string]interface{}{"name": "Product", "kind": "OBJECT"},
						map[string]interface{}{"name": "__Directive", "kind": "OBJECT"},
					},
				},
			},
		})
	})
	defer cleanup()

	err := runGraphQLIntrospect(nil, []string{})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.NotNil(t, result["data"])
}

func TestGraphQLIntrospect_Full(t *testing.T) {
	resetGraphQLFlags()
	introspectTypesOnly = false
	graphqlPretty = true

	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Contains(t, body["query"], "IntrospectionQuery")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"data": map[string]interface{}{
				"__schema": map[string]interface{}{
					"queryType": map[string]interface{}{"name": "Query"},
				},
			},
		})
	})
	defer cleanup()

	err := runGraphQLIntrospect(nil, []string{})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.NotNil(t, result["data"])
}

func TestGraphQLIntrospect_APIError(t *testing.T) {
	resetGraphQLFlags()
	introspectTypesOnly = false
	graphqlPretty = true

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusForbidden, "introspection disabled")
	})
	defer cleanup()

	err := runGraphQLIntrospect(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "introspection disabled")
}
