package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/test"
)

func setupInvitationHashTest(t *testing.T) *test.TestContext {
	tc := test.NewTestContext(t)
	tc.EnsureAuthSchema()
	tc.ExecuteSQL("DELETE FROM platform.invitation_tokens WHERE email LIKE 'e2e-invite-%'")
	return tc
}

func getTenantID(tc *test.TestContext) string {
	rows := tc.QuerySQL("SELECT id FROM platform.tenants WHERE slug = 'default' LIMIT 1")
	if len(rows) == 0 {
		return ""
	}
	id, _ := rows[0]["id"].(string)
	return id
}

func TestInvitationTokenHashStorage(t *testing.T) {
	tc := setupInvitationHashTest(t)
	defer tc.Close()

	tenantID := getTenantID(tc)
	require.NotEmpty(t, tenantID, "default tenant must exist")

	inviteEmail := fmt.Sprintf("e2e-invite-hash-%d@test.com", time.Now().UnixMilli())
	_, adminToken := tc.CreateDashboardAdminUser(
		fmt.Sprintf("e2e-invite-admin-%d@test.com", time.Now().UnixMilli()),
		"Admin-password-32chars!!",
	)

	resp := tc.NewRequest("POST", "/api/v1/admin/invitations").
		WithAuth(adminToken).
		WithBody(map[string]interface{}{
			"email":     inviteEmail,
			"role":      "tenant_admin",
			"tenant_id": tenantID,
		}).
		Send()
	resp.AssertStatus(201)

	rows := tc.QuerySQL(
		"SELECT token, token_hash FROM platform.invitation_tokens WHERE email = $1 ORDER BY created_at DESC LIMIT 1",
		inviteEmail,
	)
	require.Len(t, rows, 1, "expected exactly one invitation row")

	token, _ := rows[0]["token"].(string)
	tokenHash, _ := rows[0]["token_hash"].(string)
	assert.NotEmpty(t, token, "token (plaintext) should be populated")
	assert.NotEmpty(t, tokenHash, "token_hash should be populated")
	assert.Len(t, tokenHash, 64, "token_hash should be 64 chars (SHA-256 hex)")

	tc.ExecuteSQL("DELETE FROM platform.invitation_tokens WHERE email = $1", inviteEmail)
}

func TestInvitationLazyMigration(t *testing.T) {
	tc := setupInvitationHashTest(t)
	defer tc.Close()

	tenantID := getTenantID(tc)
	require.NotEmpty(t, tenantID, "default tenant must exist")

	legacyEmail := fmt.Sprintf("e2e-invite-legacy-%d@test.com", time.Now().UnixMilli())
	legacyToken := "legacy-invite-token-abc-123"

	tc.ExecuteSQLAsSuperuser(fmt.Sprintf(
		`INSERT INTO platform.invitation_tokens (id, tenant_id, email, token, token_hash, role, invited_by, expires_at, accepted, created_at)
		 VALUES (gen_random_uuid(), '%s', '%s', '%s', NULL, 'viewer', NULL, NOW() + INTERVAL '7 days', false, NOW())`,
		tenantID, legacyEmail, legacyToken,
	))

	rowsBefore := tc.QuerySQL("SELECT token_hash FROM platform.invitation_tokens WHERE token = $1", legacyToken)
	require.Len(t, rowsBefore, 1)
	assert.Nil(t, rowsBefore[0]["token_hash"], "token_hash should be NULL before migration")

	tc.NewRequest("POST", "/api/v1/invitations/"+legacyToken+"/accept").
		WithBody(map[string]interface{}{
			"password": "newpassword123456",
			"name":     "Test User",
		}).
		Send()

	rowsAfter := tc.QuerySQL("SELECT token_hash FROM platform.invitation_tokens WHERE token = $1", legacyToken)
	if len(rowsAfter) == 1 {
		tokenHash, _ := rowsAfter[0]["token_hash"].(string)
		assert.NotEmpty(t, tokenHash, "token_hash should be populated after lazy migration")
	}

	tc.ExecuteSQL("DELETE FROM platform.invitation_tokens WHERE email = $1", legacyEmail)
}
