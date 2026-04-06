package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Users List ---

func TestUsersList_Success(t *testing.T) {
	resetUsersFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/users")
		assert.Equal(t, "app", r.URL.Query().Get("type"))
		assert.Equal(t, "100", r.URL.Query().Get("limit"))

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"users": []map[string]interface{}{
				{"id": "user-1", "email": "user1@example.com", "role": "user", "is_active": true},
				{"id": "user-2", "email": "user2@example.com", "role": "user", "is_active": true},
			},
			"total": 2,
		})
	})
	defer cleanup()

	err := runUsersList(nil, []string{})
	require.NoError(t, err)

	// formatter.Print(result.Users) outputs a JSON array
	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Len(t, result, 2)
	assert.Equal(t, "user1@example.com", result[0]["email"])
}

func TestUsersList_Empty(t *testing.T) {
	resetUsersFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"users": []map[string]interface{}{},
			"total": 0,
		})
	})
	defer cleanup()

	err := runUsersList(nil, []string{})
	require.NoError(t, err)
}

func TestUsersList_WithSearch(t *testing.T) {
	resetUsersFlags()
	usersSearchQuery = "john"

	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "john", r.URL.Query().Get("search"))

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"users": []map[string]interface{}{
				{"id": "user-1", "email": "john@example.com", "role": "user"},
			},
			"total": 1,
		})
	})
	defer cleanup()

	err := runUsersList(nil, []string{})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Len(t, result, 1)
}

func TestUsersList_APIError(t *testing.T) {
	resetUsersFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "database error")
	})
	defer cleanup()

	err := runUsersList(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

// --- Users Get ---

func TestUsersGet_Success(t *testing.T) {
	resetUsersFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/users/user-123")
		assert.Equal(t, "app", r.URL.Query().Get("type"))

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"id":             "user-123",
			"email":          "user@example.com",
			"role":           "user",
			"email_verified": true,
			"is_active":      true,
		})
	})
	defer cleanup()

	err := runUsersGet(nil, []string{"user-123"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "user@example.com", result["email"])
}

func TestUsersGet_NotFound(t *testing.T) {
	resetUsersFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "user not found")
	})
	defer cleanup()

	err := runUsersGet(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

// --- Users Invite ---

func TestUsersInvite_Success(t *testing.T) {
	resetUsersFlags()
	appUserEmail = "newuser@example.com"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		// DoPost appends ?type=app to the path directly
		assert.Contains(t, r.URL.String(), "/api/v1/admin/users/invite")
		assert.Contains(t, r.URL.String(), "type=app")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Equal(t, "newuser@example.com", body["email"])

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"message": "Invitation sent",
		})
	})
	defer cleanup()

	err := runUsersInvite(nil, []string{})
	require.NoError(t, err)
}

func TestUsersInvite_APIError(t *testing.T) {
	resetUsersFlags()
	appUserEmail = "newuser@example.com"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusConflict, "user already exists")
	})
	defer cleanup()

	err := runUsersInvite(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user already exists")
}

// --- Users Delete ---

func TestUsersDelete_Success(t *testing.T) {
	resetUsersFlags()
	appUserForce = true

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		// DoDelete appends ?type=app to the path directly
		assert.Contains(t, r.URL.String(), "/api/v1/admin/users/user-123")
		assert.Contains(t, r.URL.String(), "type=app")

		w.WriteHeader(http.StatusNoContent)
	})
	defer cleanup()

	err := runUsersDelete(nil, []string{"user-123"})
	require.NoError(t, err)
}

func TestUsersDelete_APIError(t *testing.T) {
	resetUsersFlags()
	appUserForce = true

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "user not found")
	})
	defer cleanup()

	err := runUsersDelete(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}
