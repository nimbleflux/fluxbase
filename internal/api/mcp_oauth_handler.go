package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/auth"
	"github.com/nimbleflux/fluxbase/internal/config"
	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/nimbleflux/fluxbase/internal/mcp"
	"github.com/nimbleflux/fluxbase/internal/middleware"
)

type MCPOAuthHandler struct {
	db          *database.Connection
	config      *config.MCPConfig
	authService *auth.Service
	baseURL     string
	publicURL   string
}

func NewMCPOAuthHandler(db *database.Connection, cfg *config.MCPConfig, authService *auth.Service, baseURL, publicURL string) *MCPOAuthHandler {
	return &MCPOAuthHandler{
		db:          db,
		config:      cfg,
		authService: authService,
		baseURL:     baseURL,
		publicURL:   publicURL,
	}
}

func (h *MCPOAuthHandler) HandleAuthorizationServerMetadata(c fiber.Ctx) error {
	if !h.config.OAuth.Enabled {
		return SendNotFound(c, "OAuth is not enabled for MCP")
	}

	issuer := h.getIssuer()

	metadata := fiber.Map{
		"issuer":                                issuer,
		"authorization_endpoint":                issuer + h.config.BasePath + "/oauth/authorize",
		"token_endpoint":                        issuer + h.config.BasePath + "/oauth/token",
		"revocation_endpoint":                   issuer + h.config.BasePath + "/oauth/revoke",
		"response_types_supported":              []string{"code"},
		"response_modes_supported":              []string{"query"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
		"token_endpoint_auth_methods_supported": []string{"none"},
		"code_challenge_methods_supported":      []string{"S256"},
		"scopes_supported":                      mcp.SupportedScopes,
	}

	if h.config.OAuth.DCREnabled {
		metadata["registration_endpoint"] = issuer + h.config.BasePath + "/oauth/register"
	}

	return c.JSON(metadata)
}

func (h *MCPOAuthHandler) HandleProtectedResourceMetadata(c fiber.Ctx) error {
	if !h.config.OAuth.Enabled {
		return SendNotFound(c, "OAuth is not enabled for MCP")
	}

	issuer := h.getIssuer()

	return c.JSON(fiber.Map{
		"resource":                 issuer + h.config.BasePath,
		"authorization_servers":    []string{issuer},
		"scopes_supported":         []string{"tables:read", "tables:write", "read:schema"},
		"bearer_methods_supported": []string{"header"},
	})
}

func (h *MCPOAuthHandler) HandleClientRegistration(c fiber.Ctx) error {
	if !h.config.OAuth.Enabled || !h.config.OAuth.DCREnabled {
		return SendNotFound(c, "Dynamic Client Registration is not enabled")
	}

	var req struct {
		ClientName   string   `json:"client_name"`
		RedirectURIs []string `json:"redirect_uris"`
		Scopes       string   `json:"scope"`
	}

	if err := ParseBody(c, &req); err != nil {
		return err
	}

	if req.ClientName == "" {
		req.ClientName = "MCP Client"
	}

	if len(req.RedirectURIs) == 0 {
		return SendErrorWithCode(c, fiber.StatusBadRequest, "At least one redirect_uri is required", "invalid_redirect_uri")
	}

	for _, uri := range req.RedirectURIs {
		if !h.isRedirectURIAllowed(uri) {
			return SendErrorWithCode(c, fiber.StatusBadRequest, fmt.Sprintf("Redirect URI not allowed: %s", uri), "invalid_redirect_uri")
		}
	}

	scopes := []string{"tables:read", "read:schema"}
	if req.Scopes != "" {
		requested := strings.Fields(req.Scopes)
		supported, unknown := mcp.FilterSupportedScopes(requested)
		if len(unknown) > 0 {
			return SendErrorWithCode(c, fiber.StatusBadRequest,
				fmt.Sprintf("Unsupported scopes: %s", strings.Join(unknown, ", ")), "invalid_scope")
		}
		if len(supported) > 0 {
			scopes = supported
		}
	}

	clientID, err := generateSecureToken("mcp_", 32)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate client ID")
		return SendInternalError(c, "Failed to generate client credentials")
	}

	tenantCtx := middleware.CtxWithTenant(c)
	tenantID := database.TenantFromContext(tenantCtx)
	err = database.WrapWithServiceRoleAndTenant(tenantCtx, h.db, tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(tenantCtx, `
			INSERT INTO auth.mcp_oauth_clients (client_id, client_name, client_type, redirect_uris, scopes, metadata)
			VALUES ($1, $2, 'public', $3, $4, $5)
		`, clientID, req.ClientName, req.RedirectURIs, scopes, map[string]any{
			"user_agent":    c.Get("User-Agent"),
			"registered_at": time.Now().UTC(),
		})
		return err
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to register OAuth client")
		return SendInternalError(c, "Failed to register client")
	}

	log.Info().
		Str("client_id", clientID).
		Str("client_name", req.ClientName).
		Strs("redirect_uris", req.RedirectURIs).
		Msg("MCP OAuth client registered")

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"client_id":                  clientID,
		"client_name":                req.ClientName,
		"redirect_uris":              req.RedirectURIs,
		"scope":                      strings.Join(scopes, " "),
		"token_endpoint_auth_method": "none",
		"grant_types":                []string{"authorization_code", "refresh_token"},
		"response_types":             []string{"code"},
	})
}

// authorizeParams carries the original OAuth authorize request parameters.
type authorizeParams struct {
	ClientID            string
	RedirectURI         string
	ResponseType        string
	Scope               string
	State               string
	CodeChallenge       string
	CodeChallengeMethod string
}

// mcpOAuthClient mirrors a row of auth.mcp_oauth_clients.
type mcpOAuthClient struct {
	ClientID     string
	ClientName   string
	RedirectURIs []string
	Scopes       []string
	IsActive     bool
}

func (h *MCPOAuthHandler) loadClient(ctx context.Context, clientID string) (*mcpOAuthClient, error) {
	var client mcpOAuthClient
	err := database.WrapWithServiceRole(ctx, h.db, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `
			SELECT client_id, client_name, redirect_uris, scopes, is_active
			FROM auth.mcp_oauth_clients
			WHERE client_id = $1
		`, clientID).Scan(&client.ClientID, &client.ClientName, &client.RedirectURIs, &client.Scopes, &client.IsActive)
	})
	if err != nil {
		return nil, err
	}
	return &client, nil
}

// resolveRequestedScopes validates the requested scopes for an authorize call
// against both the client's registered scopes and the platform-wide supported
// scope list. It returns the final granted scope set.
func resolveRequestedScopes(client *mcpOAuthClient, scopeParam string) ([]string, error) {
	requested := client.Scopes
	if scopeParam != "" {
		requested = strings.Split(scopeParam, " ")
		for _, s := range requested {
			if !containsString(client.Scopes, s) {
				return nil, fmt.Errorf("scope '%s' not allowed for this client", s)
			}
		}
	}

	// Defense in depth: even scopes registered by an old DCR record must be
	// part of the currently supported set.
	supported, unknown := mcp.FilterSupportedScopes(requested)
	if len(unknown) > 0 {
		return nil, fmt.Errorf("unsupported scope: %s", strings.Join(unknown, ", "))
	}
	if len(supported) == 0 {
		return nil, fmt.Errorf("no valid scopes requested")
	}
	return supported, nil
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// HandleAuthorize validates the OAuth authorize request. Authenticated users
// are shown an interactive consent page; the code is only issued when they
// approve via the consent POST endpoint.
func (h *MCPOAuthHandler) HandleAuthorize(c fiber.Ctx) error {
	if !h.config.OAuth.Enabled {
		return SendNotFound(c, "OAuth is not enabled")
	}

	params := authorizeParams{
		ClientID:            c.Query("client_id"),
		RedirectURI:         c.Query("redirect_uri"),
		ResponseType:        c.Query("response_type"),
		Scope:               c.Query("scope"),
		State:               c.Query("state"),
		CodeChallenge:       c.Query("code_challenge"),
		CodeChallengeMethod: c.Query("code_challenge_method"),
	}

	client, _, err := h.validateAuthorizeRequest(c.RequestCtx(), &params)
	if err != nil {
		return SendErrorWithCode(c, fiber.StatusBadRequest, err.Error(), h.authorizeErrorCode(err))
	}

	userID, role := h.sessionIdentity(c)
	if userID == nil {
		authURL := h.getIssuer() + h.config.BasePath + "/oauth/authorize?" + string(c.Request().URI().QueryString())
		loginURL := h.getIssuer() + "/admin/login?return_to=" + url.QueryEscape(authURL)
		return c.Redirect().Status(fiber.StatusFound).To(loginURL)
	}

	// Scope validation happens here (after we know the client and before the
	// user approves) so the consent page always shows the effective scope set.
	scopes, scopeErr := resolveRequestedScopes(client, params.Scope)
	if scopeErr != nil {
		return SendErrorWithCode(c, fiber.StatusBadRequest, scopeErr.Error(), "invalid_scope")
	}
	if mcp.HasPrivilegedScope(scopes) && !isPrivilegedOAuthRole(role) {
		log.Warn().
			Str("client_id", client.ClientID).
			Str("role", role).
			Msg("MCP OAuth: non-admin user attempted to authorize privileged scopes")
		return SendErrorWithCode(c, fiber.StatusForbidden,
			"Privileged scopes require an admin account", "invalid_scope")
	}

	// Render the interactive consent page; the code is only issued on approval.
	return h.renderConsentPage(c, client, scopes, &params)
}

// HandleAuthorizeConsent processes the consent form POST. The code is only
// issued when the form carries a valid CSRF nonce and the user approved.
func (h *MCPOAuthHandler) HandleAuthorizeConsent(c fiber.Ctx) error {
	if !h.config.OAuth.Enabled {
		return SendNotFound(c, "OAuth is not enabled")
	}

	params := authorizeParams{
		ClientID:            c.FormValue("client_id"),
		RedirectURI:         c.FormValue("redirect_uri"),
		ResponseType:        c.FormValue("response_type"),
		Scope:               c.FormValue("scope"),
		State:               c.FormValue("state"),
		CodeChallenge:       c.FormValue("code_challenge"),
		CodeChallengeMethod: c.FormValue("code_challenge_method"),
	}

	client, redirectURI, err := h.validateAuthorizeRequest(c.RequestCtx(), &params)
	if err != nil {
		return SendErrorWithCode(c, fiber.StatusBadRequest, err.Error(), h.authorizeErrorCode(err))
	}

	// CSRF: double-submit nonce — the hidden form field must match the
	// HttpOnly cookie set when the consent page was rendered.
	nonce := c.FormValue("csrf_nonce")
	if nonce == "" || nonce != c.Cookies(consentCSRFCookieName) {
		return SendErrorWithCode(c, fiber.StatusForbidden, "Invalid or missing consent token. Please restart the authorization.", "invalid_request")
	}

	userID, role := h.sessionIdentity(c)
	if userID == nil {
		return SendErrorWithCode(c, fiber.StatusUnauthorized, "Session expired. Please restart the authorization.", "invalid_request")
	}

	// Deny (or any non-approve action) redirects back with access_denied.
	if c.FormValue("action") != "approve" {
		return h.redirectWithError(c, redirectURI, "access_denied", "The user denied the authorization request", params.State)
	}

	scopes, scopeErr := resolveRequestedScopes(client, params.Scope)
	if scopeErr != nil {
		return SendErrorWithCode(c, fiber.StatusBadRequest, scopeErr.Error(), "invalid_scope")
	}
	if mcp.HasPrivilegedScope(scopes) && !isPrivilegedOAuthRole(role) {
		return SendErrorWithCode(c, fiber.StatusForbidden, "Privileged scopes require an admin account", "invalid_scope")
	}

	code, err := generateSecureToken("", 32)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate authorization code")
		return h.redirectWithError(c, redirectURI, "server_error", "Failed to generate authorization code", params.State)
	}

	tenantCtx := middleware.CtxWithTenant(c)
	tenantID := database.TenantFromContext(tenantCtx)
	err = database.WrapWithServiceRoleAndTenant(tenantCtx, h.db, tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(tenantCtx, `
			INSERT INTO auth.mcp_oauth_codes (code, client_id, user_id, redirect_uri, scopes, code_challenge, code_challenge_method, state)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, hashToken(code), client.ClientID, userID, redirectURI, scopes, params.CodeChallenge, params.CodeChallengeMethod, params.State)
		return err
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to store authorization code")
		return h.redirectWithError(c, redirectURI, "server_error", "Failed to process authorization", params.State)
	}

	log.Debug().
		Str("client_id", client.ClientID).
		Str("code", code[:8]+"...").
		Msg("MCP OAuth authorization code issued after user consent")

	redirectURL, _ := url.Parse(redirectURI)
	q := redirectURL.Query()
	q.Set("code", code)
	if params.State != "" {
		q.Set("state", params.State)
	}
	redirectURL.RawQuery = q.Encode()

	return c.Redirect().Status(fiber.StatusFound).To(redirectURL.String())
}

// validateAuthorizeRequest performs the protocol-level checks shared by the
// authorize GET and the consent POST. It returns the client and the resolved
// redirect URI.
func (h *MCPOAuthHandler) validateAuthorizeRequest(ctx context.Context, params *authorizeParams) (*mcpOAuthClient, string, error) {
	if params.ClientID == "" {
		return nil, "", fmt.Errorf("client_id is required")
	}
	if params.ResponseType != "code" {
		return nil, "", fmt.Errorf("only response_type=code is supported")
	}
	if params.CodeChallenge == "" {
		return nil, "", fmt.Errorf("code_challenge is required (PKCE)")
	}
	if params.CodeChallengeMethod != "S256" && params.CodeChallengeMethod != "" {
		return nil, "", fmt.Errorf("code_challenge_method must be S256")
	}
	if params.CodeChallengeMethod == "" {
		params.CodeChallengeMethod = "S256"
	}

	client, err := h.loadClient(ctx, params.ClientID)
	if err != nil {
		return nil, "", fmt.Errorf("client not found")
	}
	if !client.IsActive {
		return nil, "", fmt.Errorf("client is inactive")
	}

	redirectURI := params.RedirectURI
	if redirectURI == "" && len(client.RedirectURIs) == 1 {
		redirectURI = client.RedirectURIs[0]
	}
	if !h.isURIInList(redirectURI, client.RedirectURIs) {
		return nil, "", fmt.Errorf("redirect URI not registered for this client")
	}
	return client, redirectURI, nil
}

// authorizeErrorCode maps a validation error to an OAuth error code.
func (h *MCPOAuthHandler) authorizeErrorCode(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "required"):
		return "invalid_request"
	case strings.Contains(msg, "response_type"):
		return "unsupported_response_type"
	case strings.Contains(msg, "redirect URI"):
		return "invalid_redirect_uri"
	case strings.Contains(msg, "client"):
		return "invalid_client"
	default:
		return "invalid_request"
	}
}

// isPrivilegedOAuthRole reports whether the authorizing user's role may grant
// privileged (admin-level) MCP scopes.
func isPrivilegedOAuthRole(role string) bool {
	return role == "admin" || role == "instance_admin"
}

func (h *MCPOAuthHandler) HandleToken(c fiber.Ctx) error {
	if !h.config.OAuth.Enabled {
		return SendNotFound(c, "OAuth is not enabled")
	}

	grantType := c.FormValue("grant_type")

	switch grantType {
	case "authorization_code":
		return h.handleAuthorizationCodeGrant(c)
	case "refresh_token":
		return h.handleRefreshTokenGrant(c)
	default:
		return SendErrorWithCode(c, fiber.StatusBadRequest, "Only authorization_code and refresh_token grants are supported", "unsupported_grant_type")
	}
}

func (h *MCPOAuthHandler) handleAuthorizationCodeGrant(c fiber.Ctx) error {
	code := c.FormValue("code")
	clientID := c.FormValue("client_id")
	redirectURI := c.FormValue("redirect_uri")
	codeVerifier := c.FormValue("code_verifier")

	if code == "" || clientID == "" || codeVerifier == "" {
		return SendErrorWithCode(c, fiber.StatusBadRequest, "code, client_id, and code_verifier are required", "invalid_request")
	}

	var authCode struct {
		ClientID            string
		UserID              *string
		RedirectURI         string
		Scopes              []string
		CodeChallenge       string
		CodeChallengeMethod string
		ExpiresAt           time.Time
	}

	// Codes are stored hashed; claiming is a single atomic DELETE ... RETURNING
	// so a code can never be redeemed twice (no validate-then-delete race).
	codeHash := hashToken(code)
	err := database.WrapWithServiceRole(c.RequestCtx(), h.db, func(tx pgx.Tx) error {
		return tx.QueryRow(c.RequestCtx(), `
			DELETE FROM auth.mcp_oauth_codes
			WHERE code = $1
			RETURNING client_id, user_id, redirect_uri, scopes, code_challenge, code_challenge_method, expires_at
		`, codeHash).Scan(
			&authCode.ClientID, &authCode.UserID, &authCode.RedirectURI,
			&authCode.Scopes, &authCode.CodeChallenge, &authCode.CodeChallengeMethod, &authCode.ExpiresAt,
		)
	})
	if err != nil {
		return SendErrorWithCode(c, fiber.StatusBadRequest, "Invalid authorization code", "invalid_grant")
	}

	if time.Now().After(authCode.ExpiresAt) {
		return SendErrorWithCode(c, fiber.StatusBadRequest, "Authorization code has expired", "invalid_grant")
	}

	if authCode.ClientID != clientID {
		return SendErrorWithCode(c, fiber.StatusBadRequest, "Client ID mismatch", "invalid_grant")
	}

	if redirectURI != "" && authCode.RedirectURI != redirectURI {
		return SendErrorWithCode(c, fiber.StatusBadRequest, "Redirect URI mismatch", "invalid_grant")
	}

	if !h.verifyPKCE(codeVerifier, authCode.CodeChallenge, authCode.CodeChallengeMethod) {
		return SendErrorWithCode(c, fiber.StatusBadRequest, "Invalid code_verifier", "invalid_grant")
	}

	accessToken, err := generateSecureToken("mcp_at_", 32)
	if err != nil {
		return SendInternalError(c, "Failed to generate access token")
	}

	refreshToken, err := generateSecureToken("mcp_rt_", 32)
	if err != nil {
		return SendInternalError(c, "Failed to generate refresh token")
	}

	accessTokenExpiry := time.Now().Add(h.config.OAuth.TokenExpiry)
	refreshTokenExpiry := time.Now().Add(h.config.OAuth.RefreshTokenExpiry)

	accessTokenHash := hashToken(accessToken)
	refreshTokenHash := hashToken(refreshToken)

	tenantCtx := middleware.CtxWithTenant(c)
	tenantID := database.TenantFromContext(tenantCtx)

	err = database.WrapWithServiceRoleAndTenant(tenantCtx, h.db, tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(tenantCtx, `
			INSERT INTO auth.mcp_oauth_tokens (token_type, token_hash, client_id, user_id, scopes, expires_at)
			VALUES ('access', $1, $2, $3, $4, $5)
		`, accessTokenHash, clientID, authCode.UserID, authCode.Scopes, accessTokenExpiry)
		return err
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to store access token")
		return SendInternalError(c, "Failed to store access token")
	}

	err = database.WrapWithServiceRoleAndTenant(tenantCtx, h.db, tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(tenantCtx, `
			INSERT INTO auth.mcp_oauth_tokens (token_type, token_hash, client_id, user_id, scopes, expires_at)
			VALUES ('refresh', $1, $2, $3, $4, $5)
		`, refreshTokenHash, clientID, authCode.UserID, authCode.Scopes, refreshTokenExpiry)
		return err
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to store refresh token")
		return SendInternalError(c, "Failed to store refresh token")
	}

	log.Info().
		Str("client_id", clientID).
		Strs("scopes", authCode.Scopes).
		Msg("MCP OAuth tokens issued")

	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    int(h.config.OAuth.TokenExpiry.Seconds()),
		"refresh_token": refreshToken,
		"scope":         strings.Join(authCode.Scopes, " "),
	})
}

func (h *MCPOAuthHandler) handleRefreshTokenGrant(c fiber.Ctx) error {
	refreshToken := c.FormValue("refresh_token")
	clientID := c.FormValue("client_id")

	if refreshToken == "" || clientID == "" {
		return SendErrorWithCode(c, fiber.StatusBadRequest, "refresh_token and client_id are required", "invalid_request")
	}

	refreshTokenHash := hashToken(refreshToken)

	var token struct {
		ID       string
		ClientID string
		UserID   *string
		Scopes   []string
	}

	err := database.WrapWithServiceRole(c.RequestCtx(), h.db, func(tx pgx.Tx) error {
		return tx.QueryRow(c.RequestCtx(), `
			SELECT id, client_id, user_id, scopes
			FROM auth.mcp_oauth_tokens
			WHERE token_hash = $1 AND token_type = 'refresh' AND NOT is_revoked AND expires_at > NOW()
		`, refreshTokenHash).Scan(&token.ID, &token.ClientID, &token.UserID, &token.Scopes)
	})
	if err != nil {
		return SendErrorWithCode(c, fiber.StatusBadRequest, "Invalid refresh token", "invalid_grant")
	}

	if token.ClientID != clientID {
		return SendErrorWithCode(c, fiber.StatusBadRequest, "Client ID mismatch", "invalid_grant")
	}

	// Rotate: revoke the refresh token and every token derived from it
	// (notably the associated access token), so a rotated refresh token can
	// no longer be used to keep old access tokens alive.
	_ = database.WrapWithServiceRole(c.RequestCtx(), h.db, func(tx pgx.Tx) error {
		_, err := tx.Exec(c.RequestCtx(), `
			UPDATE auth.mcp_oauth_tokens
			SET is_revoked = true, revoked_at = NOW(), revoked_reason = 'rotated'
			WHERE (id = $1 OR parent_token_id = $1) AND NOT is_revoked
		`, token.ID)
		return err
	})

	newAccessToken, _ := generateSecureToken("mcp_at_", 32)
	newRefreshToken, _ := generateSecureToken("mcp_rt_", 32)

	accessTokenExpiry := time.Now().Add(h.config.OAuth.TokenExpiry)
	refreshTokenExpiry := time.Now().Add(h.config.OAuth.RefreshTokenExpiry)

	accessTokenHash := hashToken(newAccessToken)
	newRefreshTokenHash := hashToken(newRefreshToken)

	tenantCtx := middleware.CtxWithTenant(c)
	tenantID := database.TenantFromContext(tenantCtx)

	_ = database.WrapWithServiceRoleAndTenant(tenantCtx, h.db, tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(tenantCtx, `
			INSERT INTO auth.mcp_oauth_tokens (token_type, token_hash, client_id, user_id, scopes, parent_token_id, expires_at)
			VALUES ('access', $1, $2, $3, $4, $5, $6)
		`, accessTokenHash, clientID, token.UserID, token.Scopes, token.ID, accessTokenExpiry)
		return err
	})

	_ = database.WrapWithServiceRoleAndTenant(tenantCtx, h.db, tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(tenantCtx, `
			INSERT INTO auth.mcp_oauth_tokens (token_type, token_hash, client_id, user_id, scopes, parent_token_id, expires_at)
			VALUES ('refresh', $1, $2, $3, $4, $5, $6)
		`, newRefreshTokenHash, clientID, token.UserID, token.Scopes, token.ID, refreshTokenExpiry)
		return err
	})

	return c.JSON(fiber.Map{
		"access_token":  newAccessToken,
		"token_type":    "Bearer",
		"expires_in":    int(h.config.OAuth.TokenExpiry.Seconds()),
		"refresh_token": newRefreshToken,
		"scope":         strings.Join(token.Scopes, " "),
	})
}

func (h *MCPOAuthHandler) HandleRevoke(c fiber.Ctx) error {
	if !h.config.OAuth.Enabled {
		return SendNotFound(c, "OAuth is not enabled")
	}

	token := c.FormValue("token")
	if token == "" {
		return SendErrorWithCode(c, fiber.StatusBadRequest, "token is required", "invalid_request")
	}

	tokenHash := hashToken(token)

	var rowsAffected int64
	err := database.WrapWithServiceRole(c.RequestCtx(), h.db, func(tx pgx.Tx) error {
		tag, err := tx.Exec(c.RequestCtx(), `
			UPDATE auth.mcp_oauth_tokens
			SET is_revoked = true, revoked_at = NOW(), revoked_reason = 'user_revoked'
			WHERE token_hash = $1
		`, tokenHash)
		if err != nil {
			return err
		}
		rowsAffected = tag.RowsAffected()
		return nil
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to revoke token")
	}

	if rowsAffected > 0 {
		log.Debug().Msg("MCP OAuth token revoked")
	}

	return c.SendStatus(fiber.StatusOK)
}

// ValidateAccessToken validates an MCP OAuth access token and returns the
// client ID, the authorizing user's ID (when the token represents a user),
// the granted scopes, and the user's role (empty for anonymous tokens).
func (h *MCPOAuthHandler) ValidateAccessToken(c fiber.Ctx, token string) (clientID string, userID *string, scopes []string, role string, err error) {
	tokenHash := hashToken(token)

	err = database.WrapWithServiceRole(c.RequestCtx(), h.db, func(tx pgx.Tx) error {
		return tx.QueryRow(c.RequestCtx(), `
			SELECT t.client_id, t.user_id, t.scopes, COALESCE(u.role, '')
			FROM auth.mcp_oauth_tokens t
			LEFT JOIN platform.users u ON u.id = t.user_id
			WHERE t.token_hash = $1 AND t.token_type = 'access' AND NOT t.is_revoked AND t.expires_at > NOW()
		`, tokenHash).Scan(&clientID, &userID, &scopes, &role)
	})

	return
}

func (h *MCPOAuthHandler) getIssuer() string {
	if h.publicURL != "" {
		return h.publicURL
	}
	return h.baseURL
}

func (h *MCPOAuthHandler) isRedirectURIAllowed(uri string) bool {
	allowedPatterns := h.config.OAuth.AllowedRedirectURIs
	if len(allowedPatterns) == 0 {
		allowedPatterns = config.DefaultMCPOAuthRedirectURIs()
	}

	for _, pattern := range allowedPatterns {
		if h.matchURIPattern(uri, pattern) {
			return true
		}
	}
	return false
}

func (h *MCPOAuthHandler) matchURIPattern(uri, pattern string) bool {
	if uri == pattern {
		return true
	}

	if strings.HasSuffix(pattern, "://") {
		return strings.HasPrefix(uri, pattern)
	}

	if strings.Contains(pattern, ":*") {
		prefix := strings.Split(pattern, ":*")[0]
		if strings.HasPrefix(uri, prefix+":") {
			parsedURI, err := url.Parse(uri)
			if err != nil {
				return false
			}
			parsedPattern, err := url.Parse(strings.Replace(pattern, ":*", ":1234", 1))
			if err != nil {
				return false
			}
			return parsedURI.Hostname() == parsedPattern.Hostname()
		}
	}

	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			return strings.HasPrefix(uri, parts[0]) && strings.HasSuffix(uri, parts[1])
		}
	}

	return false
}

func (h *MCPOAuthHandler) isURIInList(uri string, list []string) bool {
	for _, u := range list {
		if u == uri || h.matchURIPattern(uri, u) {
			return true
		}
	}
	return false
}

// sessionIdentity resolves the signed-in user (via bearer token or session
// cookies) and returns their user ID together with their role, which gates
// privileged scope grants.
func (h *MCPOAuthHandler) sessionIdentity(c fiber.Ctx) (userID *string, role string) {
	validate := func(token string) (*string, string) {
		if token == "" || h.authService == nil {
			return nil, ""
		}
		claims, err := h.authService.JWTManager().ValidateToken(token)
		if err != nil {
			return nil, ""
		}
		isRevoked, err := h.authService.TokenBlacklistService().IsTokenRevoked(c.RequestCtx(), claims.ID, "", time.Time{})
		if err != nil || isRevoked {
			return nil, ""
		}
		id := claims.UserID
		return &id, claims.Role
	}

	authHeader := c.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		// MCP access tokens are not session tokens; skip JWT validation for them.
		if !strings.HasPrefix(token, "mcp_at_") {
			if id, r := validate(token); id != nil {
				return id, r
			}
		}
	}

	if id, r := validate(c.Cookies(AccessTokenCookieName)); id != nil {
		return id, r
	}

	adminToken := c.Cookies("fluxbase_admin_token")
	if adminToken != "" {
		if len(adminToken) >= 2 && adminToken[0] == '"' && adminToken[len(adminToken)-1] == '"' {
			adminToken = adminToken[1 : len(adminToken)-1]
		}
		if id, r := validate(adminToken); id != nil {
			return id, r
		}
	}

	return nil, ""
}

// consentCSRFCookieName is the cookie used for the consent form's
// double-submit CSRF nonce.
const consentCSRFCookieName = "mcp_consent_csrf"

// renderConsentPage serves the standalone consent HTML page (server-rendered,
// no external assets) listing the client and requested scopes with
// approve/deny buttons. The original authorize parameters travel as hidden
// fields and the CSRF nonce is stored in both the form and an HttpOnly cookie.
func (h *MCPOAuthHandler) renderConsentPage(c fiber.Ctx, client *mcpOAuthClient, scopes []string, params *authorizeParams) error {
	nonce, err := generateSecureToken("", 24)
	if err != nil {
		return SendInternalError(c, "Failed to generate consent token")
	}

	c.Cookie(&fiber.Cookie{
		Name:     consentCSRFCookieName,
		Value:    nonce,
		Path:     h.config.BasePath + "/oauth/",
		MaxAge:   900, // 15 minutes
		HTTPOnly: true,
		SameSite: "Lax",
	})

	hidden := func(name, value string) string {
		return fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`, name, html.EscapeString(value))
	}

	var scopeItems strings.Builder
	for _, s := range scopes {
		scopeItems.WriteString(fmt.Sprintf(`<li><code>%s</code></li>`, html.EscapeString(s)))
	}

	page := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Authorize %s — Fluxbase MCP</title>
<style>
body{font-family:system-ui,-apple-system,sans-serif;background:#f5f6f8;margin:0;display:flex;align-items:center;justify-content:center;min-height:100vh;color:#1a1a2e}
.card{background:#fff;border-radius:12px;box-shadow:0 2px 12px rgba(0,0,0,.08);max-width:420px;width:90%%;padding:32px}
h1{font-size:1.15rem;margin:0 0 8px}
p{font-size:.9rem;line-height:1.5;color:#555;margin:.4rem 0}
ul{padding-left:1.2rem;font-size:.85rem}
.buttons{display:flex;gap:12px;margin-top:24px}
button{flex:1;padding:10px 0;border-radius:8px;border:0;font-size:.95rem;cursor:pointer}
.approve{background:#2563eb;color:#fff}
.approve:hover{background:#1d4ed8}
.deny{background:#e5e7eb;color:#374151}
.deny:hover{background:#d1d5db}
code{background:#f0f1f5;padding:1px 5px;border-radius:4px;font-size:.8rem}
</style>
</head>
<body>
<div class="card">
<h1>Authorize "%s"?</h1>
<p>The application <strong>%s</strong> is requesting access to your Fluxbase account with the following scopes:</p>
<ul>%s</ul>
<p>If you approve, %s will be able to perform the actions listed above.</p>
<form method="post" action="%s/oauth/authorize/consent">
%s%s%s%s%s%s%s%s
<button class="approve" type="submit" name="action" value="approve">Approve</button>
<button class="deny" type="submit" name="action" value="deny">Deny</button>
</form>
</div>
</body>
</html>`,
		html.EscapeString(client.ClientName),
		html.EscapeString(client.ClientName),
		html.EscapeString(client.ClientName),
		scopeItems.String(),
		html.EscapeString(client.ClientName),
		html.EscapeString(h.config.BasePath),
		hidden("client_id", params.ClientID),
		hidden("redirect_uri", params.RedirectURI),
		hidden("response_type", params.ResponseType),
		hidden("scope", params.Scope),
		hidden("state", params.State),
		hidden("code_challenge", params.CodeChallenge),
		hidden("code_challenge_method", params.CodeChallengeMethod),
		hidden("csrf_nonce", nonce),
	)

	c.Set("Content-Type", "text/html; charset=utf-8")
	c.Set("Cache-Control", "no-store")
	return c.SendString(page)
}

func (h *MCPOAuthHandler) verifyPKCE(verifier, challenge, method string) bool {
	if method != "S256" {
		return false
	}

	hash := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(hash[:])

	return computed == challenge
}

func (h *MCPOAuthHandler) redirectWithError(c fiber.Ctx, redirectURI, errorCode, errorDesc, state string) error {
	u, _ := url.Parse(redirectURI)
	q := u.Query()
	q.Set("error", errorCode)
	q.Set("error_description", errorDesc)
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()
	return c.Redirect().Status(fiber.StatusFound).To(u.String())
}

func generateSecureToken(prefix string, length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(bytes), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func matchRedirectURI(pattern, uri string) bool {
	if uri == pattern {
		return true
	}

	if strings.HasSuffix(pattern, "://") {
		return strings.HasPrefix(uri, pattern)
	}

	if strings.Contains(pattern, ":*") {
		prefix := strings.Split(pattern, ":*")[0]
		if strings.HasPrefix(uri, prefix+":") {
			parsedURI, err := url.Parse(uri)
			if err != nil {
				return false
			}
			parsedPattern, err := url.Parse(strings.Replace(pattern, ":*", ":1234", 1))
			if err != nil {
				return false
			}
			return parsedURI.Hostname() == parsedPattern.Hostname()
		}
	}

	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			return strings.HasPrefix(uri, parts[0]) && strings.HasSuffix(uri, parts[1])
		}
	}

	return false
}

func verifyPKCE(verifier, challenge, method string) bool {
	if method != "S256" {
		return false
	}

	hash := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(hash[:])

	return computed == challenge
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	for i := range bytes {
		bytes[i] = charset[bytes[i]%byte(len(charset))]
	}
	return string(bytes)
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// fiber:context-methods migrated
