package api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/branching"
	"github.com/nimbleflux/fluxbase/internal/config"
)

// GitHubWebhookHandler handles GitHub webhook events for database branching
type GitHubWebhookHandler struct {
	manager *branching.Manager
	router  *branching.Router
	config  config.BranchingConfig
}

// NewGitHubWebhookHandler creates a new GitHub webhook handler
func NewGitHubWebhookHandler(manager *branching.Manager, router *branching.Router, cfg config.BranchingConfig) *GitHubWebhookHandler {
	return &GitHubWebhookHandler{
		manager: manager,
		router:  router,
		config:  cfg,
	}
}

// GitHubWebhookPayload represents the common fields in GitHub webhook payloads
type GitHubWebhookPayload struct {
	Action       string              `json:"action"`
	PullRequest  *GitHubPullRequest  `json:"pull_request,omitempty"`
	Issue        *GitHubIssue        `json:"issue,omitempty"`
	Label        *GitHubLabel        `json:"label,omitempty"`
	Repository   *GitHubRepository   `json:"repository,omitempty"`
	Sender       *GitHubUser         `json:"sender,omitempty"`
	Installation *GitHubInstallation `json:"installation,omitempty"`
}

// GitHubIssue represents a GitHub issue
type GitHubIssue struct {
	Number    int           `json:"number"`
	State     string        `json:"state"`
	Title     string        `json:"title"`
	Body      string        `json:"body"`
	HTMLURL   string        `json:"html_url"`
	Labels    []GitHubLabel `json:"labels"`
	Assignees []GitHubUser  `json:"assignees"`
	User      *GitHubUser   `json:"user,omitempty"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
}

// GitHubLabel represents a GitHub label
type GitHubLabel struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

// GitHubPullRequest represents a GitHub pull request
type GitHubPullRequest struct {
	Number  int        `json:"number"`
	State   string     `json:"state"`
	Title   string     `json:"title"`
	HTMLURL string     `json:"html_url"`
	Merged  bool       `json:"merged"`
	Base    *GitHubRef `json:"base,omitempty"`
	Head    *GitHubRef `json:"head,omitempty"`
}

// GitHubRef represents a Git reference (branch)
type GitHubRef struct {
	Ref  string            `json:"ref"`
	SHA  string            `json:"sha"`
	Repo *GitHubRepository `json:"repo,omitempty"`
}

// GitHubRepository represents a GitHub repository
type GitHubRepository struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	HTMLURL  string `json:"html_url"`
}

// GitHubUser represents a GitHub user
type GitHubUser struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
}

// GitHubInstallation represents a GitHub App installation
type GitHubInstallation struct {
	ID int `json:"id"`
}

// HandleWebhook handles incoming GitHub webhook requests.
// The X-Hub-Signature-256 HMAC over the raw body is verified BEFORE the
// payload is parsed, so attacker-controlled bytes are never unmarshalled or
// acted upon without a valid signature. Unsigned deliveries are rejected
// unless branching.allow_unsigned_webhooks is enabled — and even then only for
// repositories configured without a webhook secret.
func (h *GitHubWebhookHandler) HandleWebhook(c fiber.Ctx) error {
	if !h.config.Enabled {
		return SendErrorWithCode(c, fiber.StatusServiceUnavailable, "Database branching is not enabled", "BRANCHING_DISABLED")
	}

	// Get event type from header
	eventType := c.Get("X-GitHub-Event")
	if eventType == "" {
		return SendErrorWithCode(c, fiber.StatusBadRequest, "Missing X-GitHub-Event header", "MISSING_EVENT")
	}

	// Get delivery ID for logging
	deliveryID := c.Get("X-GitHub-Delivery")

	log.Info().
		Str("event", eventType).
		Str("delivery_id", deliveryID).
		Msg("Received GitHub webhook")

	// Retain the raw body: the signature is computed over it and it must not
	// be re-read from the (possibly already consumed) request stream.
	rawBody := c.Body()

	// Verify the signature over the raw body before parsing anything. The
	// per-repository secret means verification also establishes which
	// configured repository the delivery can act on.
	verifiedRepo, err := h.verifySignature(c.RequestCtx(), rawBody, c.Get("X-Hub-Signature-256"))
	if err != nil {
		log.Warn().
			Err(err).
			Str("delivery_id", deliveryID).
			Msg("Webhook signature verification failed")
		return SendErrorWithCode(c, fiber.StatusUnauthorized, "Webhook signature verification failed", "INVALID_SIGNATURE")
	}

	// Parse the payload only after verification succeeded
	var payload GitHubWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return SendErrorWithCode(c, fiber.StatusBadRequest, "Failed to parse webhook payload: "+err.Error(), "INVALID_PAYLOAD")
	}

	// Get repository full name
	if payload.Repository == nil {
		return SendErrorWithCode(c, fiber.StatusBadRequest, "Missing repository in webhook payload", "MISSING_REPOSITORY")
	}

	repoFullName := payload.Repository.FullName

	// A signature-verified delivery may only act on the repository whose
	// secret validated it — the payload cannot redirect it elsewhere.
	if verifiedRepo != "" && repoFullName != verifiedRepo {
		log.Warn().
			Str("delivery_id", deliveryID).
			Str("signed_repository", verifiedRepo).
			Str("payload_repository", repoFullName).
			Msg("Webhook payload repository does not match the signed repository")
		return SendErrorWithCode(c, fiber.StatusUnauthorized, "Webhook signature verification failed", "INVALID_SIGNATURE")
	}

	// Unsigned deliveries (accepted only under allow_unsigned_webhooks) are
	// further restricted to repositories configured without a webhook secret.
	if verifiedRepo == "" && !h.shouldAcceptUnsigned(c.RequestCtx(), repoFullName) {
		log.Warn().
			Str("repository", repoFullName).
			Str("delivery_id", deliveryID).
			Msg("Unsigned webhook rejected - repository not configured for unsigned webhooks")
		return SendErrorWithCode(c, fiber.StatusUnauthorized, "Webhook signature verification failed", "INVALID_SIGNATURE")
	}

	// Handle different event types
	switch eventType {
	case "pull_request":
		return h.handlePullRequestEvent(c, &payload)
	case "issues":
		return h.handleIssueEvent(c, &payload)
	case "ping":
		return h.handlePingEvent(c, &payload)
	default:
		// Ignore other events
		log.Debug().
			Str("event", eventType).
			Str("repository", repoFullName).
			Msg("Ignoring unhandled GitHub event")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "ignored",
			"message": "Event type not handled",
		})
	}
}

// verifySignature verifies the X-Hub-Signature-256 header over the raw body
// BEFORE any payload parsing. It returns the repository whose configured
// webhook secret validated the signature, or "" when the delivery is accepted
// unsigned under the explicit allow_unsigned_webhooks opt-in. An error means
// the delivery must be rejected.
func (h *GitHubWebhookHandler) verifySignature(ctx context.Context, rawBody []byte, signature string) (string, error) {
	configs, err := h.manager.GetStorage().ListGitHubConfigs(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list GitHub configs: %w", err)
	}

	// Separate configs into those with and without a configured secret
	var secretRepos, unsignedRepos []string
	for _, cfg := range configs {
		if cfg == nil {
			continue
		}
		if cfg.WebhookSecret != nil && *cfg.WebhookSecret != "" {
			secretRepos = append(secretRepos, cfg.Repository)
		} else {
			unsignedRepos = append(unsignedRepos, cfg.Repository)
		}
	}

	if signature == "" {
		// No signature provided. Unsigned deliveries are rejected unless the
		// explicit opt-in is set, and even then only accepted for repositories
		// configured without a webhook secret (checked by the caller via
		// shouldAcceptUnsigned once the repository is known from the payload).
		if !h.config.AllowUnsignedWebhooks {
			return "", fmt.Errorf("webhook signature required but not provided")
		}
		if len(unsignedRepos) == 0 {
			return "", fmt.Errorf("no repository is configured for unsigned webhooks")
		}
		log.Warn().
			Strs("repositories", unsignedRepos).
			Msg("GitHub webhook accepted without signature verification - allow_unsigned_webhooks is enabled; configure webhook_secret to require signatures")
		return "", nil
	}

	// Signature provided: try every configured secret against the raw body.
	// The repo identity comes from whichever secret verifies, not from the
	// (untrusted) payload.
	for _, cfg := range configs {
		if cfg == nil || cfg.WebhookSecret == nil || *cfg.WebhookSecret == "" {
			continue
		}
		expected := computeHMACSHA256(rawBody, *cfg.WebhookSecret)
		expectedSignature := "sha256=" + expected
		if hmac.Equal([]byte(signature), []byte(expectedSignature)) {
			return cfg.Repository, nil
		}
	}

	if len(secretRepos) == 0 {
		return "", fmt.Errorf("signature provided but no repository has a webhook secret configured")
	}
	return "", fmt.Errorf("signature mismatch for any configured repository")
}

// shouldAcceptUnsigned reports whether an unsigned delivery for the given
// repository may be processed: only when the repository is configured without
// a webhook secret (explicit opt-in to insecure mode).
func (h *GitHubWebhookHandler) shouldAcceptUnsigned(ctx context.Context, repository string) bool {
	ghConfig, err := h.manager.GetStorage().GetGitHubConfig(ctx, repository)
	if err != nil {
		if !errors.Is(err, branching.ErrGitHubConfigNotFound) {
			log.Error().Err(err).Str("repository", repository).Msg("Failed to get GitHub config for unsigned webhook check")
		}
		return false
	}
	if ghConfig == nil {
		return false
	}
	return ghConfig.WebhookSecret == nil || *ghConfig.WebhookSecret == ""
}

// computeHMACSHA256 computes HMAC-SHA256 of data with the given key
func computeHMACSHA256(data []byte, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

// handlePullRequestEvent handles pull request events
func (h *GitHubWebhookHandler) handlePullRequestEvent(c fiber.Ctx, payload *GitHubWebhookPayload) error {
	if payload.PullRequest == nil {
		return SendErrorWithCode(c, fiber.StatusBadRequest, "Missing pull_request in payload", "MISSING_PULL_REQUEST")
	}

	pr := payload.PullRequest
	repo := payload.Repository.FullName

	log.Info().
		Str("action", payload.Action).
		Int("pr_number", pr.Number).
		Str("repository", repo).
		Msg("Processing pull request event")

	// Get GitHub config for this repository
	ghConfig, err := h.manager.GetStorage().GetGitHubConfig(c.RequestCtx(), repo)
	if err != nil && !errors.Is(err, branching.ErrGitHubConfigNotFound) {
		log.Error().Err(err).Str("repository", repo).Msg("Failed to get GitHub config")
		return SendErrorWithCode(c, fiber.StatusInternalServerError, "Failed to get GitHub configuration", "CONFIG_ERROR")
	}

	// Use default settings if no config
	autoCreate := true
	autoDelete := true
	if ghConfig != nil {
		autoCreate = ghConfig.AutoCreateOnPR
		autoDelete = ghConfig.AutoDeleteOnMerge
	}

	switch payload.Action {
	case "opened", "reopened":
		if !autoCreate {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status":  "skipped",
				"message": "Auto-create is disabled for this repository",
			})
		}
		return h.createBranchForPR(c, repo, pr)

	case "closed":
		if !autoDelete {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status":  "skipped",
				"message": "Auto-delete is disabled for this repository",
			})
		}
		return h.deleteBranchForPR(c, repo, pr)

	case "synchronize":
		// PR was updated (new commits pushed)
		// Could trigger migrations here in the future
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "acknowledged",
			"message": "PR synchronize event acknowledged",
		})

	default:
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "ignored",
			"message": "Pull request action not handled: " + payload.Action,
		})
	}
}

// createBranchForPR creates a database branch for a pull request
func (h *GitHubWebhookHandler) createBranchForPR(c fiber.Ctx, repo string, pr *GitHubPullRequest) error {
	branch, err := h.manager.CreateBranchFromGitHubPR(c.RequestCtx(), repo, pr.Number, pr.HTMLURL)
	if err != nil {
		log.Error().Err(err).
			Str("repository", repo).
			Int("pr_number", pr.Number).
			Msg("Failed to create branch for PR")

		if errors.Is(err, branching.ErrBranchExists) {
			// Branch already exists, return success
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status":  "exists",
				"message": "Branch already exists for this PR",
			})
		}

		if errors.Is(err, branching.ErrMaxBranchesReached) {
			return SendErrorWithCode(c, fiber.StatusForbidden, "Maximum number of branches has been reached", "MAX_BRANCHES_REACHED")
		}

		return SendErrorWithCode(c, fiber.StatusInternalServerError, "Failed to create branch: "+err.Error(), "CREATE_FAILED")
	}

	log.Info().
		Str("branch_slug", branch.Slug).
		Str("repository", repo).
		Int("pr_number", pr.Number).
		Msg("Created branch for PR")

	// Warmup the connection pool
	go func() {
		if err := h.router.WarmupPool(c.RequestCtx(), branch.Slug); err != nil {
			log.Warn().Err(err).Str("slug", branch.Slug).Msg("Failed to warmup branch pool")
		}
	}()

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":      "created",
		"branch_id":   branch.ID,
		"branch_slug": branch.Slug,
		"database":    branch.DatabaseName,
		"pr_number":   pr.Number,
	})
}

// deleteBranchForPR deletes the database branch for a pull request
func (h *GitHubWebhookHandler) deleteBranchForPR(c fiber.Ctx, repo string, pr *GitHubPullRequest) error {
	// Find branch by PR
	branch, err := h.manager.GetStorage().GetBranchByGitHubPR(c.RequestCtx(), repo, pr.Number)
	if err != nil {
		if errors.Is(err, branching.ErrBranchNotFound) {
			// Branch doesn't exist, return success
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status":  "not_found",
				"message": "No branch exists for this PR",
			})
		}
		log.Error().Err(err).
			Str("repository", repo).
			Int("pr_number", pr.Number).
			Msg("Failed to find branch for PR")
		return SendErrorWithCode(c, fiber.StatusInternalServerError, "Failed to find branch: "+err.Error(), "FIND_FAILED")
	}

	// Close the connection pool first
	h.router.ClosePool(branch.Slug)

	// Delete the branch
	if err := h.manager.DeleteBranch(c.RequestCtx(), branch.ID, nil); err != nil {
		log.Error().Err(err).
			Str("repository", repo).
			Int("pr_number", pr.Number).
			Str("branch_id", branch.ID.String()).
			Msg("Failed to delete branch for PR")
		return SendErrorWithCode(c, fiber.StatusInternalServerError, "Failed to delete branch: "+err.Error(), "DELETE_FAILED")
	}

	log.Info().
		Str("branch_slug", branch.Slug).
		Str("repository", repo).
		Int("pr_number", pr.Number).
		Bool("merged", pr.Merged).
		Msg("Deleted branch for PR")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":      "deleted",
		"branch_slug": branch.Slug,
		"pr_number":   pr.Number,
		"merged":      pr.Merged,
	})
}

// handlePingEvent handles GitHub ping events (sent when webhook is first configured)
func (h *GitHubWebhookHandler) handlePingEvent(c fiber.Ctx, payload *GitHubWebhookPayload) error {
	repo := ""
	if payload.Repository != nil {
		repo = payload.Repository.FullName
	}

	log.Info().
		Str("repository", repo).
		Msg("Received GitHub ping event")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "pong",
		"message": "Webhook configured successfully",
	})
}

// handleIssueEvent handles issue events (opened, labeled, closed, etc.)
func (h *GitHubWebhookHandler) handleIssueEvent(c fiber.Ctx, payload *GitHubWebhookPayload) error {
	if payload.Issue == nil {
		return SendErrorWithCode(c, fiber.StatusBadRequest, "Missing issue in payload", "MISSING_ISSUE")
	}

	issue := payload.Issue
	repo := payload.Repository.FullName

	log.Info().
		Str("action", payload.Action).
		Int("issue_number", issue.Number).
		Str("repository", repo).
		Str("title", issue.Title).
		Msg("Processing issue event")

	switch payload.Action {
	case "opened":
		return h.handleIssueOpened(c, repo, issue, payload.Sender)

	case "labeled":
		if payload.Label == nil {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status":  "ignored",
				"message": "No label in payload",
			})
		}
		return h.handleIssueLabeled(c, repo, issue, payload.Label, payload.Sender)

	case "closed":
		return h.handleIssueClosed(c, repo, issue, payload.Sender)

	case "assigned":
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "acknowledged",
			"message": "Issue assignment acknowledged",
			"issue":   issue.Number,
		})

	default:
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "ignored",
			"message": "Issue action not handled: " + payload.Action,
		})
	}
}

// handleIssueOpened handles when a new issue is opened
func (h *GitHubWebhookHandler) handleIssueOpened(c fiber.Ctx, repo string, issue *GitHubIssue, sender *GitHubUser) error {
	senderLogin := ""
	if sender != nil {
		senderLogin = sender.Login
	}

	log.Info().
		Str("repository", repo).
		Int("issue_number", issue.Number).
		Str("title", issue.Title).
		Str("sender", senderLogin).
		Msg("New issue opened")

	// Check if the issue has the claude-fix label from the start
	for _, label := range issue.Labels {
		if label.Name == "claude-fix" {
			log.Info().
				Int("issue_number", issue.Number).
				Msg("Issue opened with claude-fix label - automation will be triggered via GitHub Actions")
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":       "acknowledged",
		"message":      "Issue opened event processed",
		"issue_number": issue.Number,
		"title":        issue.Title,
	})
}

// handleIssueLabeled handles when a label is added to an issue
func (h *GitHubWebhookHandler) handleIssueLabeled(c fiber.Ctx, repo string, issue *GitHubIssue, label *GitHubLabel, sender *GitHubUser) error {
	senderLogin := ""
	if sender != nil {
		senderLogin = sender.Login
	}

	log.Info().
		Str("repository", repo).
		Int("issue_number", issue.Number).
		Str("label", label.Name).
		Str("sender", senderLogin).
		Msg("Label added to issue")

	// Handle specific labels that trigger automation
	switch label.Name {
	case "claude-fix":
		log.Info().
			Int("issue_number", issue.Number).
			Str("repository", repo).
			Msg("claude-fix label detected - GitHub Actions workflow will handle automation")

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":       "acknowledged",
			"message":      "claude-fix label detected, automation triggered via GitHub Actions",
			"issue_number": issue.Number,
			"label":        label.Name,
		})

	case "priority:critical", "priority:high":
		log.Warn().
			Int("issue_number", issue.Number).
			Str("repository", repo).
			Str("priority", label.Name).
			Msg("High priority issue labeled")

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":       "acknowledged",
			"message":      "High priority issue noted",
			"issue_number": issue.Number,
			"label":        label.Name,
		})

	default:
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":       "acknowledged",
			"message":      "Label added",
			"issue_number": issue.Number,
			"label":        label.Name,
		})
	}
}

// handleIssueClosed handles when an issue is closed
func (h *GitHubWebhookHandler) handleIssueClosed(c fiber.Ctx, repo string, issue *GitHubIssue, sender *GitHubUser) error {
	senderLogin := ""
	if sender != nil {
		senderLogin = sender.Login
	}

	log.Info().
		Str("repository", repo).
		Int("issue_number", issue.Number).
		Str("title", issue.Title).
		Str("sender", senderLogin).
		Msg("Issue closed")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":       "acknowledged",
		"message":      "Issue closed",
		"issue_number": issue.Number,
	})
}

// GetWebhookURL returns the webhook URL for configuration
func (h *GitHubWebhookHandler) GetWebhookURL(baseURL string) string {
	return strings.TrimSuffix(baseURL, "/") + "/api/v1/webhooks/github"
}

// fiber:context-methods migrated
