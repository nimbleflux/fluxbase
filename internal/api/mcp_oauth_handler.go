package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/fluxbase-eu/fluxbase/internal/auth"
	"github.com/fluxbase-eu/fluxbase/internal/config"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// MCPOAuthHandler handles OAuth 2.1 authentication for MCP clients
type MCPOAuthHandler struct {
	db          *pgxpool.Pool
	config      *config.MCPConfig
	authService *auth.Service
	baseURL     string
	publicURL   string
}

// NewMCPOAuthHandler creates a new MCP OAuth handler
func NewMCPOAuthHandler(db *pgxpool.Pool, cfg *config.MCPConfig, authService *auth.Service, baseURL, publicURL string) *MCPOAuthHandler {
	return &MCPOAuthHandler{
		db:          db,
		config:      cfg,
		authService: authService,
		baseURL:     baseURL,
		publicURL:   publicURL,
	}
}

// RegisterRoutes registers OAuth routes
// wellKnownGroup is for /.well-known endpoints (public, no auth)
// mcpGroup is for /mcp/oauth/* endpoints
func (h *MCPOAuthHandler) RegisterRoutes(app fiber.Router, mcpGroup fiber.Router) {
	// Public discovery endpoints (no auth required)
	app.Get("/.well-known/oauth-authorization-server", h.handleAuthorizationServerMetadata)
	app.Get("/.well-known/oauth-protected-resource", h.handleProtectedResourceMetadata)
	app.Get("/.well-known/oauth-protected-resource/mcp", h.handleProtectedResourceMetadata)

	// OAuth endpoints (under /mcp/oauth/*)
	oauth := mcpGroup.Group("/oauth")

	// Dynamic Client Registration (public)
	oauth.Post("/register", h.handleClientRegistration)

	// Authorization endpoints
	oauth.Get("/authorize", h.handleAuthorize)
	oauth.Post("/authorize", h.handleAuthorizeConsent)

	// Token endpoints
	oauth.Post("/token", h.handleToken)
	oauth.Post("/revoke", h.handleRevoke)
}

// handleAuthorizationServerMetadata returns OAuth 2.0 Authorization Server Metadata
// RFC 8414: https://datatracker.ietf.org/doc/html/rfc8414
func (h *MCPOAuthHandler) handleAuthorizationServerMetadata(c *fiber.Ctx) error {
	if !h.config.OAuth.Enabled {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "OAuth is not enabled for MCP",
		})
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
		"token_endpoint_auth_methods_supported": []string{"none"}, // Public clients only (PKCE required)
		"code_challenge_methods_supported":      []string{"S256"},
		"scopes_supported": []string{
			"read:tables", "write:tables",
			"execute:functions", "execute:rpc",
			"read:storage", "write:storage",
			"execute:jobs", "read:vectors",
			"read:schema", "admin:ddl",
		},
	}

	// Add DCR endpoint if enabled
	if h.config.OAuth.DCREnabled {
		metadata["registration_endpoint"] = issuer + h.config.BasePath + "/oauth/register"
	}

	return c.JSON(metadata)
}

// handleProtectedResourceMetadata returns OAuth 2.0 Protected Resource Metadata
// This tells clients where to get authorization
func (h *MCPOAuthHandler) handleProtectedResourceMetadata(c *fiber.Ctx) error {
	if !h.config.OAuth.Enabled {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "OAuth is not enabled for MCP",
		})
	}

	issuer := h.getIssuer()

	return c.JSON(fiber.Map{
		"resource":                 issuer + h.config.BasePath,
		"authorization_servers":    []string{issuer},
		"scopes_supported":         []string{"read:tables", "write:tables", "read:schema"},
		"bearer_methods_supported": []string{"header"},
	})
}

// handleClientRegistration handles Dynamic Client Registration (DCR)
// RFC 7591: https://datatracker.ietf.org/doc/html/rfc7591
func (h *MCPOAuthHandler) handleClientRegistration(c *fiber.Ctx) error {
	if !h.config.OAuth.Enabled || !h.config.OAuth.DCREnabled {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":             "invalid_request",
			"error_description": "Dynamic Client Registration is not enabled",
		})
	}

	var req struct {
		ClientName   string   `json:"client_name"`
		RedirectURIs []string `json:"redirect_uris"`
		Scopes       string   `json:"scope"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_request",
			"error_description": "Invalid request body",
		})
	}

	// Validate client name
	if req.ClientName == "" {
		req.ClientName = "MCP Client"
	}

	// Validate redirect URIs
	if len(req.RedirectURIs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_redirect_uri",
			"error_description": "At least one redirect_uri is required",
		})
	}

	for _, uri := range req.RedirectURIs {
		if !h.isRedirectURIAllowed(uri) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":             "invalid_redirect_uri",
				"error_description": fmt.Sprintf("Redirect URI not allowed: %s", uri),
			})
		}
	}

	// Parse and validate scopes (default to safe scopes)
	scopes := []string{"read:tables", "read:schema"}
	if req.Scopes != "" {
		scopes = strings.Split(req.Scopes, " ")
	}

	// Generate client ID
	clientID, err := generateSecureToken("mcp_", 32)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate client ID")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             "server_error",
			"error_description": "Failed to generate client credentials",
		})
	}

	// Insert client into database
	_, err = h.db.Exec(c.Context(), `
		INSERT INTO auth.mcp_oauth_clients (client_id, client_name, client_type, redirect_uris, scopes, metadata)
		VALUES ($1, $2, 'public', $3, $4, $5)
	`, clientID, req.ClientName, req.RedirectURIs, scopes, map[string]any{
		"user_agent":    c.Get("User-Agent"),
		"registered_at": time.Now().UTC(),
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to register OAuth client")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             "server_error",
			"error_description": "Failed to register client",
		})
	}

	log.Info().
		Str("client_id", clientID).
		Str("client_name", req.ClientName).
		Strs("redirect_uris", req.RedirectURIs).
		Msg("MCP OAuth client registered")

	// Return client credentials
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

// handleAuthorize handles the authorization endpoint
func (h *MCPOAuthHandler) handleAuthorize(c *fiber.Ctx) error {
	if !h.config.OAuth.Enabled {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "OAuth is not enabled",
		})
	}

	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	responseType := c.Query("response_type")
	scope := c.Query("scope")
	state := c.Query("state")
	codeChallenge := c.Query("code_challenge")
	codeChallengeMethod := c.Query("code_challenge_method")

	// Validate required parameters
	if clientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_request",
			"error_description": "client_id is required",
		})
	}

	if responseType != "code" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "unsupported_response_type",
			"error_description": "Only response_type=code is supported",
		})
	}

	// PKCE is required for public clients
	if codeChallenge == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_request",
			"error_description": "code_challenge is required (PKCE)",
		})
	}

	if codeChallengeMethod != "S256" && codeChallengeMethod != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_request",
			"error_description": "code_challenge_method must be S256",
		})
	}
	if codeChallengeMethod == "" {
		codeChallengeMethod = "S256"
	}

	// Look up client
	var client struct {
		ClientID     string   `db:"client_id"`
		ClientName   string   `db:"client_name"`
		RedirectURIs []string `db:"redirect_uris"`
		Scopes       []string `db:"scopes"`
		IsActive     bool     `db:"is_active"`
	}

	err := h.db.QueryRow(c.Context(), `
		SELECT client_id, client_name, redirect_uris, scopes, is_active
		FROM auth.mcp_oauth_clients
		WHERE client_id = $1
	`, clientID).Scan(&client.ClientID, &client.ClientName, &client.RedirectURIs, &client.Scopes, &client.IsActive)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_client",
			"error_description": "Client not found",
		})
	}

	if !client.IsActive {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_client",
			"error_description": "Client is inactive",
		})
	}

	// Validate redirect URI
	if redirectURI == "" && len(client.RedirectURIs) == 1 {
		redirectURI = client.RedirectURIs[0]
	}

	if !h.isURIInList(redirectURI, client.RedirectURIs) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_redirect_uri",
			"error_description": "Redirect URI not registered for this client",
		})
	}

	// Parse requested scopes (default to client's scopes)
	requestedScopes := client.Scopes
	if scope != "" {
		requestedScopes = strings.Split(scope, " ")
		// Validate scopes are subset of client's allowed scopes
		for _, s := range requestedScopes {
			if !h.contains(client.Scopes, s) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":             "invalid_scope",
					"error_description": fmt.Sprintf("Scope '%s' not allowed for this client", s),
				})
			}
		}
	}

	// Check if user is authenticated
	// Try to get user from session cookie or Bearer token
	userID := h.extractUserFromRequest(c)

	// If no user is authenticated, redirect to login
	// The login page should redirect back here after authentication
	if userID == nil {
		// Build the authorization URL to return to after login
		authURL := h.getIssuer() + h.config.BasePath + "/oauth/authorize?" + string(c.Request().URI().QueryString())

		// For now, return an error asking user to authenticate first
		// In a full implementation, redirect to login page with return_to parameter
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":             "login_required",
			"error_description": "User authentication required. Please log in to Fluxbase first.",
			"login_url":         h.getIssuer() + "/auth/login?return_to=" + url.QueryEscape(authURL),
		})
	}

	// Generate authorization code
	code, err := generateSecureToken("", 32)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate authorization code")
		return h.redirectWithError(c, redirectURI, "server_error", "Failed to generate authorization code", state)
	}

	// Store authorization code (now includes user_id)
	_, err = h.db.Exec(c.Context(), `
		INSERT INTO auth.mcp_oauth_codes (code, client_id, user_id, redirect_uri, scopes, code_challenge, code_challenge_method, state)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, code, clientID, userID, redirectURI, requestedScopes, codeChallenge, codeChallengeMethod, state)

	if err != nil {
		log.Error().Err(err).Msg("Failed to store authorization code")
		return h.redirectWithError(c, redirectURI, "server_error", "Failed to process authorization", state)
	}

	log.Debug().
		Str("client_id", clientID).
		Str("code", code[:8]+"...").
		Msg("MCP OAuth authorization code issued")

	// Redirect with code
	redirectURL, _ := url.Parse(redirectURI)
	q := redirectURL.Query()
	q.Set("code", code)
	if state != "" {
		q.Set("state", state)
	}
	redirectURL.RawQuery = q.Encode()

	return c.Redirect(redirectURL.String(), fiber.StatusFound)
}

// handleAuthorizeConsent handles POST to authorization endpoint (consent form submission)
func (h *MCPOAuthHandler) handleAuthorizeConsent(c *fiber.Ctx) error {
	// For now, redirect to GET handler
	// In a full implementation, this would process user consent
	return h.handleAuthorize(c)
}

// handleToken handles the token endpoint
func (h *MCPOAuthHandler) handleToken(c *fiber.Ctx) error {
	if !h.config.OAuth.Enabled {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "OAuth is not enabled",
		})
	}

	grantType := c.FormValue("grant_type")

	switch grantType {
	case "authorization_code":
		return h.handleAuthorizationCodeGrant(c)
	case "refresh_token":
		return h.handleRefreshTokenGrant(c)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "unsupported_grant_type",
			"error_description": "Only authorization_code and refresh_token grants are supported",
		})
	}
}

// handleAuthorizationCodeGrant exchanges an authorization code for tokens
func (h *MCPOAuthHandler) handleAuthorizationCodeGrant(c *fiber.Ctx) error {
	code := c.FormValue("code")
	clientID := c.FormValue("client_id")
	redirectURI := c.FormValue("redirect_uri")
	codeVerifier := c.FormValue("code_verifier")

	if code == "" || clientID == "" || codeVerifier == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_request",
			"error_description": "code, client_id, and code_verifier are required",
		})
	}

	// Look up and validate authorization code
	var authCode struct {
		ClientID            string
		UserID              *string
		RedirectURI         string
		Scopes              []string
		CodeChallenge       string
		CodeChallengeMethod string
		ExpiresAt           time.Time
	}

	err := h.db.QueryRow(c.Context(), `
		SELECT client_id, user_id, redirect_uri, scopes, code_challenge, code_challenge_method, expires_at
		FROM auth.mcp_oauth_codes
		WHERE code = $1
	`, code).Scan(
		&authCode.ClientID, &authCode.UserID, &authCode.RedirectURI,
		&authCode.Scopes, &authCode.CodeChallenge, &authCode.CodeChallengeMethod, &authCode.ExpiresAt,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_grant",
			"error_description": "Invalid authorization code",
		})
	}

	// Delete the code (one-time use)
	_, _ = h.db.Exec(c.Context(), `DELETE FROM auth.mcp_oauth_codes WHERE code = $1`, code)

	// Validate code hasn't expired
	if time.Now().After(authCode.ExpiresAt) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_grant",
			"error_description": "Authorization code has expired",
		})
	}

	// Validate client ID matches
	if authCode.ClientID != clientID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_grant",
			"error_description": "Client ID mismatch",
		})
	}

	// Validate redirect URI matches (if provided)
	if redirectURI != "" && authCode.RedirectURI != redirectURI {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_grant",
			"error_description": "Redirect URI mismatch",
		})
	}

	// Verify PKCE code verifier
	if !h.verifyPKCE(codeVerifier, authCode.CodeChallenge, authCode.CodeChallengeMethod) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_grant",
			"error_description": "Invalid code_verifier",
		})
	}

	// Generate tokens
	accessToken, err := generateSecureToken("mcp_at_", 32)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             "server_error",
			"error_description": "Failed to generate access token",
		})
	}

	refreshToken, err := generateSecureToken("mcp_rt_", 32)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             "server_error",
			"error_description": "Failed to generate refresh token",
		})
	}

	accessTokenExpiry := time.Now().Add(h.config.OAuth.TokenExpiry)
	refreshTokenExpiry := time.Now().Add(h.config.OAuth.RefreshTokenExpiry)

	// Store tokens
	accessTokenHash := hashToken(accessToken)
	refreshTokenHash := hashToken(refreshToken)

	_, err = h.db.Exec(c.Context(), `
		INSERT INTO auth.mcp_oauth_tokens (token_type, token_hash, client_id, user_id, scopes, expires_at)
		VALUES ('access', $1, $2, $3, $4, $5)
	`, accessTokenHash, clientID, authCode.UserID, authCode.Scopes, accessTokenExpiry)

	if err != nil {
		log.Error().Err(err).Msg("Failed to store access token")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             "server_error",
			"error_description": "Failed to store access token",
		})
	}

	_, err = h.db.Exec(c.Context(), `
		INSERT INTO auth.mcp_oauth_tokens (token_type, token_hash, client_id, user_id, scopes, expires_at)
		VALUES ('refresh', $1, $2, $3, $4, $5)
	`, refreshTokenHash, clientID, authCode.UserID, authCode.Scopes, refreshTokenExpiry)

	if err != nil {
		log.Error().Err(err).Msg("Failed to store refresh token")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             "server_error",
			"error_description": "Failed to store refresh token",
		})
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

// handleRefreshTokenGrant exchanges a refresh token for new tokens
func (h *MCPOAuthHandler) handleRefreshTokenGrant(c *fiber.Ctx) error {
	refreshToken := c.FormValue("refresh_token")
	clientID := c.FormValue("client_id")

	if refreshToken == "" || clientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_request",
			"error_description": "refresh_token and client_id are required",
		})
	}

	refreshTokenHash := hashToken(refreshToken)

	// Look up refresh token
	var token struct {
		ID       string
		ClientID string
		UserID   *string
		Scopes   []string
	}

	err := h.db.QueryRow(c.Context(), `
		SELECT id, client_id, user_id, scopes
		FROM auth.mcp_oauth_tokens
		WHERE token_hash = $1 AND token_type = 'refresh' AND NOT is_revoked AND expires_at > NOW()
	`, refreshTokenHash).Scan(&token.ID, &token.ClientID, &token.UserID, &token.Scopes)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_grant",
			"error_description": "Invalid refresh token",
		})
	}

	// Validate client ID matches
	if token.ClientID != clientID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_grant",
			"error_description": "Client ID mismatch",
		})
	}

	// Revoke old refresh token (rotation)
	_, _ = h.db.Exec(c.Context(), `
		UPDATE auth.mcp_oauth_tokens
		SET is_revoked = true, revoked_at = NOW(), revoked_reason = 'rotated'
		WHERE id = $1
	`, token.ID)

	// Generate new tokens
	newAccessToken, _ := generateSecureToken("mcp_at_", 32)
	newRefreshToken, _ := generateSecureToken("mcp_rt_", 32)

	accessTokenExpiry := time.Now().Add(h.config.OAuth.TokenExpiry)
	refreshTokenExpiry := time.Now().Add(h.config.OAuth.RefreshTokenExpiry)

	// Store new tokens
	accessTokenHash := hashToken(newAccessToken)
	newRefreshTokenHash := hashToken(newRefreshToken)

	_, _ = h.db.Exec(c.Context(), `
		INSERT INTO auth.mcp_oauth_tokens (token_type, token_hash, client_id, user_id, scopes, expires_at)
		VALUES ('access', $1, $2, $3, $4, $5)
	`, accessTokenHash, clientID, token.UserID, token.Scopes, accessTokenExpiry)

	_, _ = h.db.Exec(c.Context(), `
		INSERT INTO auth.mcp_oauth_tokens (token_type, token_hash, client_id, user_id, scopes, expires_at)
		VALUES ('refresh', $1, $2, $3, $4, $5)
	`, newRefreshTokenHash, clientID, token.UserID, token.Scopes, refreshTokenExpiry)

	return c.JSON(fiber.Map{
		"access_token":  newAccessToken,
		"token_type":    "Bearer",
		"expires_in":    int(h.config.OAuth.TokenExpiry.Seconds()),
		"refresh_token": newRefreshToken,
		"scope":         strings.Join(token.Scopes, " "),
	})
}

// handleRevoke handles token revocation
func (h *MCPOAuthHandler) handleRevoke(c *fiber.Ctx) error {
	if !h.config.OAuth.Enabled {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "OAuth is not enabled",
		})
	}

	token := c.FormValue("token")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_request",
			"error_description": "token is required",
		})
	}

	tokenHash := hashToken(token)

	// Revoke the token (and any related refresh tokens)
	result, err := h.db.Exec(c.Context(), `
		UPDATE auth.mcp_oauth_tokens
		SET is_revoked = true, revoked_at = NOW(), revoked_reason = 'user_revoked'
		WHERE token_hash = $1
	`, tokenHash)

	if err != nil {
		log.Error().Err(err).Msg("Failed to revoke token")
	}

	// Per RFC 7009, always return 200 OK even if token wasn't found
	rowsAffected := result.RowsAffected()
	if rowsAffected > 0 {
		log.Debug().Msg("MCP OAuth token revoked")
	}

	return c.SendStatus(fiber.StatusOK)
}

// ValidateAccessToken validates an MCP OAuth access token and returns the auth context
func (h *MCPOAuthHandler) ValidateAccessToken(c *fiber.Ctx, token string) (clientID string, userID *string, scopes []string, err error) {
	tokenHash := hashToken(token)

	err = h.db.QueryRow(c.Context(), `
		SELECT client_id, user_id, scopes
		FROM auth.mcp_oauth_tokens
		WHERE token_hash = $1 AND token_type = 'access' AND NOT is_revoked AND expires_at > NOW()
	`, tokenHash).Scan(&clientID, &userID, &scopes)

	return
}

// Helper functions

func (h *MCPOAuthHandler) getIssuer() string {
	if h.publicURL != "" {
		return h.publicURL
	}
	return h.baseURL
}

func (h *MCPOAuthHandler) isRedirectURIAllowed(uri string) bool {
	allowedPatterns := h.config.OAuth.AllowedRedirectURIs
	if len(allowedPatterns) == 0 {
		// Apply defaults if not configured
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
	// Exact match
	if uri == pattern {
		return true
	}

	// Scheme-only match (e.g., "vscode://")
	if strings.HasSuffix(pattern, "://") {
		return strings.HasPrefix(uri, pattern)
	}

	// Wildcard port match (e.g., "http://localhost:*")
	if strings.Contains(pattern, ":*") {
		prefix := strings.Split(pattern, ":*")[0]
		if strings.HasPrefix(uri, prefix+":") {
			// Check that everything after the port is a valid path
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

	// Wildcard path match (e.g., "cursor://anysphere.cursor-mcp/oauth/*/callback")
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

func (h *MCPOAuthHandler) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// extractUserFromRequest attempts to extract a user ID from the request
// Checks: Authorization Bearer token, access_token cookie
func (h *MCPOAuthHandler) extractUserFromRequest(c *fiber.Ctx) *string {
	// Try Authorization header (Bearer token)
	authHeader := c.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Skip MCP OAuth tokens (they start with "mcp_at_")
		if !strings.HasPrefix(token, "mcp_at_") && h.authService != nil {
			claims, err := h.authService.ValidateToken(token)
			if err == nil {
				// Check if token has been revoked
				isRevoked, err := h.authService.IsTokenRevoked(c.Context(), claims.ID)
				if err == nil && !isRevoked {
					return &claims.UserID
				}
			}
		}
	}

	// Try access_token cookie
	accessToken := c.Cookies("access_token")
	if accessToken != "" && h.authService != nil {
		claims, err := h.authService.ValidateToken(accessToken)
		if err == nil {
			isRevoked, err := h.authService.IsTokenRevoked(c.Context(), claims.ID)
			if err == nil && !isRevoked {
				return &claims.UserID
			}
		}
	}

	return nil
}

func (h *MCPOAuthHandler) verifyPKCE(verifier, challenge, method string) bool {
	if method != "S256" {
		return false
	}

	// S256: BASE64URL(SHA256(code_verifier)) == code_challenge
	hash := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(hash[:])

	return computed == challenge
}

func (h *MCPOAuthHandler) redirectWithError(c *fiber.Ctx, redirectURI, errorCode, errorDesc, state string) error {
	u, _ := url.Parse(redirectURI)
	q := u.Query()
	q.Set("error", errorCode)
	q.Set("error_description", errorDesc)
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()
	return c.Redirect(u.String(), fiber.StatusFound)
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(prefix string, length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(bytes), nil
}

// hashToken creates a SHA256 hash of a token for storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// matchRedirectURI checks if a URI matches a pattern (standalone version for testing)
// Supports exact matches, wildcard ports (e.g., "http://localhost:*"),
// and wildcard path segments (e.g., "cursor://host/oauth/*/callback")
func matchRedirectURI(pattern, uri string) bool {
	// Exact match
	if uri == pattern {
		return true
	}

	// Scheme-only match (e.g., "vscode://")
	if strings.HasSuffix(pattern, "://") {
		return strings.HasPrefix(uri, pattern)
	}

	// Wildcard port match (e.g., "http://localhost:*")
	if strings.Contains(pattern, ":*") {
		prefix := strings.Split(pattern, ":*")[0]
		if strings.HasPrefix(uri, prefix+":") {
			// Check that everything after the port is a valid path
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

	// Wildcard path match (e.g., "cursor://anysphere.cursor-mcp/oauth/*/callback")
	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			return strings.HasPrefix(uri, parts[0]) && strings.HasSuffix(uri, parts[1])
		}
	}

	return false
}

// verifyPKCE verifies a PKCE code verifier against a challenge (standalone version for testing)
func verifyPKCE(verifier, challenge, method string) bool {
	if method != "S256" {
		return false
	}

	// S256: BASE64URL(SHA256(code_verifier)) == code_challenge
	hash := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(hash[:])

	return computed == challenge
}

// generateRandomString generates a random alphanumeric string of the specified length
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

// nullIfEmpty returns nil if the string is empty, otherwise a pointer to the string
func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
