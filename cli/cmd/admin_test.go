package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Admin Users List ---

func TestAdminUsersList_Success(t *testing.T) {
	resetAdminUsersFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/users")
		assert.Equal(t, "dashboard", r.URL.Query().Get("type"))

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"users": []map[string]interface{}{
				{"id": "user-1", "email": "admin@example.com", "role": "instance_admin"},
				{"id": "user-2", "email": "admin2@example.com", "role": "tenant_admin"},
			},
			"total": 2,
		})
	})
	defer cleanup()

	err := runAdminUsersList(nil, []string{})
	require.NoError(t, err)

	// formatter.Print(result.Users) outputs a JSON array, not the wrapper object
	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Len(t, result, 2)
	assert.Equal(t, "admin@example.com", result[0]["email"])
}

func TestAdminUsersList_Empty(t *testing.T) {
	resetAdminUsersFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"users": []map[string]interface{}{},
			"total": 0,
		})
	})
	defer cleanup()

	err := runAdminUsersList(nil, []string{})
	require.NoError(t, err)
}

func TestAdminUsersList_APIError(t *testing.T) {
	resetAdminUsersFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "database error")
	})
	defer cleanup()

	err := runAdminUsersList(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

// --- Admin Users Get ---

func TestAdminUsersGet_Success(t *testing.T) {
	resetAdminUsersFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/users/user-123")
		assert.Equal(t, "dashboard", r.URL.Query().Get("type"))

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"id":     "user-123",
			"email":  "admin@example.com",
			"role":   "instance_admin",
			"active": true,
		})
	})
	defer cleanup()

	err := runAdminUsersGet(nil, []string{"user-123"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "admin@example.com", result["email"])
}

func TestAdminUsersGet_NotFound(t *testing.T) {
	resetAdminUsersFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "user not found")
	})
	defer cleanup()

	err := runAdminUsersGet(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

// --- Admin Users Invite ---

func TestAdminUsersInvite_Success(t *testing.T) {
	resetAdminUsersFlags()
	adminUserEmail = "newadmin@example.com"
	adminUserRole = "tenant_admin"

	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/invitations")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Equal(t, "newadmin@example.com", body["email"])
		assert.Equal(t, "tenant_admin", body["role"])

		respondJSON(w, http.StatusCreated, map[string]interface{}{
			"token": "invite-token-abc",
			"email": "newadmin@example.com",
		})
	})
	defer cleanup()

	err := runAdminUsersInvite(nil, []string{})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "newadmin@example.com", result["email"])
}

func TestAdminUsersInvite_APIError(t *testing.T) {
	resetAdminUsersFlags()
	adminUserEmail = "newadmin@example.com"
	adminUserRole = "tenant_admin"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusConflict, "invitation already exists")
	})
	defer cleanup()

	err := runAdminUsersInvite(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invitation already exists")
}

// --- Admin Users Delete ---

func TestAdminUsersDelete_Success(t *testing.T) {
	resetAdminUsersFlags()
	adminUserForce = true

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		// DoDelete appends ?type=dashboard to the path directly
		assert.Contains(t, r.URL.String(), "/api/v1/admin/users/user-123")
		assert.Contains(t, r.URL.String(), "type=dashboard")

		w.WriteHeader(http.StatusNoContent)
	})
	defer cleanup()

	err := runAdminUsersDelete(nil, []string{"user-123"})
	require.NoError(t, err)
}

func TestAdminUsersDelete_APIError(t *testing.T) {
	resetAdminUsersFlags()
	adminUserForce = true

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "admin user not found")
	})
	defer cleanup()

	err := runAdminUsersDelete(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin user not found")
}

// --- Admin Invitations List ---

func TestAdminInvitationsList_Success(t *testing.T) {
	resetAdminInvitationsFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/invitations")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"invitations": []map[string]interface{}{
				{"token": "abc123def456", "email": "admin@example.com", "role": "tenant_admin"},
			},
		})
	})
	defer cleanup()

	err := runAdminInvitationsList(nil, []string{})
	require.NoError(t, err)

	// formatter.Print(result.Invitations) outputs a JSON array
	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Len(t, result, 1)
	assert.Equal(t, "admin@example.com", result[0]["email"])
}

func TestAdminInvitationsList_Empty(t *testing.T) {
	resetAdminInvitationsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"invitations": []map[string]interface{}{},
		})
	})
	defer cleanup()

	err := runAdminInvitationsList(nil, []string{})
	require.NoError(t, err)
}

func TestAdminInvitationsList_APIError(t *testing.T) {
	resetAdminInvitationsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "database error")
	})
	defer cleanup()

	err := runAdminInvitationsList(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestAdminInvitationsList_WithAccepted(t *testing.T) {
	resetAdminInvitationsFlags()
	invIncludeAccepted = true

	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "true", r.URL.Query().Get("include_accepted"))
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"invitations": []map[string]interface{}{
				{"token": "abc123", "email": "admin@example.com", "role": "tenant_admin"},
			},
		})
	})
	defer cleanup()

	err := runAdminInvitationsList(nil, []string{})
	require.NoError(t, err)

	// formatter.Print(result.Invitations) outputs a JSON array
	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Len(t, result, 1)
}

// --- Admin Invitations Revoke ---

func TestAdminInvitationsRevoke_Success(t *testing.T) {
	resetAdminInvitationsFlags()
	invForce = true

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/invitations/invite-token")

		w.WriteHeader(http.StatusNoContent)
	})
	defer cleanup()

	err := runAdminInvitationsRevoke(nil, []string{"invite-token"})
	require.NoError(t, err)
}

func TestAdminInvitationsRevoke_APIError(t *testing.T) {
	resetAdminInvitationsFlags()
	invForce = true

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "invitation not found")
	})
	defer cleanup()

	err := runAdminInvitationsRevoke(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invitation not found")
}

// --- Admin Sessions List ---

func TestAdminSessionsList_Success(t *testing.T) {
	resetAdminSessionsFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/auth/sessions")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"sessions": []map[string]interface{}{
				{"id": "sess-1", "user_email": "admin@example.com", "ip_address": "127.0.0.1"},
			},
			"total": 1,
		})
	})
	defer cleanup()

	err := runAdminSessionsList(nil, []string{})
	require.NoError(t, err)

	// formatter.Print(result.Sessions) outputs a JSON array
	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Len(t, result, 1)
	assert.Equal(t, "admin@example.com", result[0]["user_email"])
}

func TestAdminSessionsList_Empty(t *testing.T) {
	resetAdminSessionsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"sessions": []map[string]interface{}{},
			"total":    0,
		})
	})
	defer cleanup()

	err := runAdminSessionsList(nil, []string{})
	require.NoError(t, err)
}

func TestAdminSessionsList_APIError(t *testing.T) {
	resetAdminSessionsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "session store error")
	})
	defer cleanup()

	err := runAdminSessionsList(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session store error")
}

// --- Admin Sessions Revoke ---

func TestAdminSessionsRevoke_Success(t *testing.T) {
	resetAdminSessionsFlags()
	sessionForce = true

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/auth/sessions/sess-123")

		w.WriteHeader(http.StatusNoContent)
	})
	defer cleanup()

	err := runAdminSessionsRevoke(nil, []string{"sess-123"})
	require.NoError(t, err)
}

func TestAdminSessionsRevoke_APIError(t *testing.T) {
	resetAdminSessionsFlags()
	sessionForce = true

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "session not found")
	})
	defer cleanup()

	err := runAdminSessionsRevoke(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

// --- Admin Sessions Revoke All ---

func TestAdminSessionsRevokeAll_Success(t *testing.T) {
	resetAdminSessionsFlags()
	sessionForce = true

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/auth/sessions/user/user-123")

		w.WriteHeader(http.StatusNoContent)
	})
	defer cleanup()

	err := runAdminSessionsRevokeAll(nil, []string{"user-123"})
	require.NoError(t, err)
}

func TestAdminSessionsRevokeAll_APIError(t *testing.T) {
	resetAdminSessionsFlags()
	sessionForce = true

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "user not found")
	})
	defer cleanup()

	err := runAdminSessionsRevokeAll(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}
