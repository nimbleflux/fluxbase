package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/database"
)

// auditTimeout bounds the audit insert so a slow database never stalls a
// tool call or resource read.
const auditTimeout = 3 * time.Second

// maxAuditErrorLength caps the stored error text.
const maxAuditErrorLength = 1024

// writeAudit records one MCP operation in the durable audit trail
// (mcp.audit_log). Best effort: failures are logged and never affect the
// RPC response. Retention/pruning of this table is handled by the platform's
// audit log retention job, like the other audit tables.
func (s *Server) writeAudit(ctx context.Context, authCtx *AuthContext, operation string, arguments map[string]any, durationMs int64, success bool, errMsg string) {
	if s.db == nil {
		return
	}

	if authCtx == nil {
		authCtx = &AuthContext{}
	}

	if len(errMsg) > maxAuditErrorLength {
		errMsg = errMsg[:maxAuditErrorLength]
	}

	var argsJSON []byte
	if arguments != nil {
		argsJSON, _ = json.Marshal(arguments)
	} else {
		argsJSON = []byte("null")
	}

	tenantID := database.TenantFromContext(ctx)

	auditCtx, cancel := context.WithTimeout(ctx, auditTimeout)
	defer cancel()

	err := database.WrapWithServiceRoleAndTenant(auditCtx, s.db, tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(auditCtx, `
			INSERT INTO mcp.audit_log (tenant_id, auth_type, user_id, client_key_id, tool, arguments, duration_ms, success, error)
			VALUES (NULLIF($1, '')::uuid, $2, $3, NULLIF($4, ''), $5, $6::jsonb, $7, $8, NULLIF($9, ''))
		`, tenantID, authCtx.AuthType, authCtx.UserID, authCtx.ClientKeyID, operation, string(argsJSON), durationMs, success, errMsg)
		return err
	})
	if err != nil {
		log.Debug().Err(err).Str("operation", operation).Msg("MCP: Failed to write audit log entry")
	}
}

// sanitizeMCPError converts an error into a client-safe message. Database
// (pgconn) errors are routed through the shared sanitizer so raw server
// messages — which can embed SQL fragments or internal details — never reach
// MCP clients. Application-level errors keep their original message.
func sanitizeMCPError(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return database.SanitizeErrorMessage(err)
	}
	return err.Error()
}
