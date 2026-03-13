package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nimbleflux/fluxbase/internal/mcp"
	"github.com/rs/zerolog/log"
)

// GitHubClient provides methods for interacting with the GitHub API
type GitHubClient struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

// NewGitHubClient creates a new GitHub API client
func NewGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.github.com",
	}
}

// doRequest performs an authenticated request to the GitHub API
func (c *GitHubClient) doRequest(ctx context.Context, method, path string, body io.Reader) ([]byte, error) {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "Fluxbase-MCP")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// ============================================================================
// LIST ISSUES TOOL
// ============================================================================

// ListGitHubIssuesTool implements the list_github_issues MCP tool
type ListGitHubIssuesTool struct {
	client *GitHubClient
}

// NewListGitHubIssuesTool creates a new list_github_issues tool
func NewListGitHubIssuesTool(client *GitHubClient) *ListGitHubIssuesTool {
	return &ListGitHubIssuesTool{client: client}
}

func (t *ListGitHubIssuesTool) Name() string {
	return "list_github_issues"
}

func (t *ListGitHubIssuesTool) Description() string {
	return `List GitHub issues for a repository with optional filtering.

Parameters:
  - repository: Repository in format "owner/repo" (required)
  - state: Filter by state: open, closed, all (default: open)
  - labels: Comma-separated list of labels to filter by
  - assignee: Filter by assignee username
  - limit: Maximum number of results (default: 30, max: 100)

Returns list of issues with number, title, state, labels, and assignees.`
}

func (t *ListGitHubIssuesTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"repository": map[string]any{
				"type":        "string",
				"description": "Repository in format 'owner/repo'",
			},
			"state": map[string]any{
				"type":        "string",
				"description": "Filter by state: open, closed, all",
				"enum":        []string{"open", "closed", "all"},
				"default":     "open",
			},
			"labels": map[string]any{
				"type":        "string",
				"description": "Comma-separated list of labels to filter by",
			},
			"assignee": map[string]any{
				"type":        "string",
				"description": "Filter by assignee username",
			},
			"limit": map[string]any{
				"type":        "integer",
				"description": "Maximum number of results (default: 30, max: 100)",
				"default":     30,
			},
		},
		"required": []string{"repository"},
	}
}

func (t *ListGitHubIssuesTool) RequiredScopes() []string {
	return []string{mcp.ScopeGitHubRead}
}

func (t *ListGitHubIssuesTool) Execute(ctx context.Context, args map[string]any, authCtx *mcp.AuthContext) (*mcp.ToolResult, error) {
	repo, ok := args["repository"].(string)
	if !ok || repo == "" {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent("repository is required")},
			IsError: true,
		}, nil
	}

	// Build query parameters
	params := []string{}

	state := "open"
	if s, ok := args["state"].(string); ok && s != "" {
		state = s
	}
	params = append(params, "state="+state)

	if labels, ok := args["labels"].(string); ok && labels != "" {
		params = append(params, "labels="+labels)
	}

	if assignee, ok := args["assignee"].(string); ok && assignee != "" {
		params = append(params, "assignee="+assignee)
	}

	limit := 30
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
		if limit > 100 {
			limit = 100
		}
	}
	params = append(params, fmt.Sprintf("per_page=%d", limit))

	path := fmt.Sprintf("/repos/%s/issues?%s", repo, strings.Join(params, "&"))

	respBody, err := t.client.doRequest(ctx, "GET", path, nil)
	if err != nil {
		log.Error().Err(err).Str("repository", repo).Msg("MCP: list_github_issues - failed")
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to list issues: %v", err))},
			IsError: true,
		}, nil
	}

	var issues []map[string]any
	if err := json.Unmarshal(respBody, &issues); err != nil {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to parse response: %v", err))},
			IsError: true,
		}, nil
	}

	// Filter out pull requests (GitHub API returns PRs in issues endpoint)
	filteredIssues := make([]map[string]any, 0)
	for _, issue := range issues {
		if _, hasPR := issue["pull_request"]; !hasPR {
			// Simplify the response
			simplified := map[string]any{
				"number":     issue["number"],
				"title":      issue["title"],
				"state":      issue["state"],
				"html_url":   issue["html_url"],
				"created_at": issue["created_at"],
				"updated_at": issue["updated_at"],
			}

			if labels, ok := issue["labels"].([]any); ok {
				labelNames := make([]string, 0, len(labels))
				for _, l := range labels {
					if label, ok := l.(map[string]any); ok {
						if name, ok := label["name"].(string); ok {
							labelNames = append(labelNames, name)
						}
					}
				}
				simplified["labels"] = labelNames
			}

			if assignees, ok := issue["assignees"].([]any); ok {
				assigneeNames := make([]string, 0, len(assignees))
				for _, a := range assignees {
					if assignee, ok := a.(map[string]any); ok {
						if login, ok := assignee["login"].(string); ok {
							assigneeNames = append(assigneeNames, login)
						}
					}
				}
				simplified["assignees"] = assigneeNames
			}

			filteredIssues = append(filteredIssues, simplified)
		}
	}

	log.Info().
		Str("repository", repo).
		Int("count", len(filteredIssues)).
		Msg("MCP: list_github_issues - success")

	resultJSON, _ := json.MarshalIndent(filteredIssues, "", "  ")
	return &mcp.ToolResult{
		Content: []mcp.Content{mcp.TextContent(string(resultJSON))},
	}, nil
}

// ============================================================================
// GET ISSUE TOOL
// ============================================================================

// GetGitHubIssueTool implements the get_github_issue MCP tool
type GetGitHubIssueTool struct {
	client *GitHubClient
}

// NewGetGitHubIssueTool creates a new get_github_issue tool
func NewGetGitHubIssueTool(client *GitHubClient) *GetGitHubIssueTool {
	return &GetGitHubIssueTool{client: client}
}

func (t *GetGitHubIssueTool) Name() string {
	return "get_github_issue"
}

func (t *GetGitHubIssueTool) Description() string {
	return `Get details of a specific GitHub issue.

Parameters:
  - repository: Repository in format "owner/repo" (required)
  - issue_number: Issue number (required)

Returns full issue details including body, comments count, and metadata.`
}

func (t *GetGitHubIssueTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"repository": map[string]any{
				"type":        "string",
				"description": "Repository in format 'owner/repo'",
			},
			"issue_number": map[string]any{
				"type":        "integer",
				"description": "Issue number",
			},
		},
		"required": []string{"repository", "issue_number"},
	}
}

func (t *GetGitHubIssueTool) RequiredScopes() []string {
	return []string{mcp.ScopeGitHubRead}
}

func (t *GetGitHubIssueTool) Execute(ctx context.Context, args map[string]any, authCtx *mcp.AuthContext) (*mcp.ToolResult, error) {
	repo, ok := args["repository"].(string)
	if !ok || repo == "" {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent("repository is required")},
			IsError: true,
		}, nil
	}

	issueNum, ok := args["issue_number"].(float64)
	if !ok {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent("issue_number is required")},
			IsError: true,
		}, nil
	}

	path := fmt.Sprintf("/repos/%s/issues/%d", repo, int(issueNum))

	respBody, err := t.client.doRequest(ctx, "GET", path, nil)
	if err != nil {
		log.Error().Err(err).Str("repository", repo).Int("issue", int(issueNum)).Msg("MCP: get_github_issue - failed")
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to get issue: %v", err))},
			IsError: true,
		}, nil
	}

	var issue map[string]any
	if err := json.Unmarshal(respBody, &issue); err != nil {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to parse response: %v", err))},
			IsError: true,
		}, nil
	}

	// Simplify the response
	result := map[string]any{
		"number":         issue["number"],
		"title":          issue["title"],
		"state":          issue["state"],
		"body":           issue["body"],
		"html_url":       issue["html_url"],
		"comments_count": issue["comments"],
		"created_at":     issue["created_at"],
		"updated_at":     issue["updated_at"],
		"closed_at":      issue["closed_at"],
	}

	if user, ok := issue["user"].(map[string]any); ok {
		result["author"] = user["login"]
	}

	if labels, ok := issue["labels"].([]any); ok {
		labelNames := make([]string, 0, len(labels))
		for _, l := range labels {
			if label, ok := l.(map[string]any); ok {
				if name, ok := label["name"].(string); ok {
					labelNames = append(labelNames, name)
				}
			}
		}
		result["labels"] = labelNames
	}

	if assignees, ok := issue["assignees"].([]any); ok {
		assigneeNames := make([]string, 0, len(assignees))
		for _, a := range assignees {
			if assignee, ok := a.(map[string]any); ok {
				if login, ok := assignee["login"].(string); ok {
					assigneeNames = append(assigneeNames, login)
				}
			}
		}
		result["assignees"] = assigneeNames
	}

	log.Info().
		Str("repository", repo).
		Int("issue", int(issueNum)).
		Msg("MCP: get_github_issue - success")

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{
		Content: []mcp.Content{mcp.TextContent(string(resultJSON))},
	}, nil
}

// ============================================================================
// CREATE ISSUE COMMENT TOOL
// ============================================================================

// CreateGitHubIssueCommentTool implements the create_github_issue_comment MCP tool
type CreateGitHubIssueCommentTool struct {
	client *GitHubClient
}

// NewCreateGitHubIssueCommentTool creates a new create_github_issue_comment tool
func NewCreateGitHubIssueCommentTool(client *GitHubClient) *CreateGitHubIssueCommentTool {
	return &CreateGitHubIssueCommentTool{client: client}
}

func (t *CreateGitHubIssueCommentTool) Name() string {
	return "create_github_issue_comment"
}

func (t *CreateGitHubIssueCommentTool) Description() string {
	return `Add a comment to a GitHub issue.

Parameters:
  - repository: Repository in format "owner/repo" (required)
  - issue_number: Issue number (required)
  - body: Comment body in markdown (required)

Returns the created comment details.`
}

func (t *CreateGitHubIssueCommentTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"repository": map[string]any{
				"type":        "string",
				"description": "Repository in format 'owner/repo'",
			},
			"issue_number": map[string]any{
				"type":        "integer",
				"description": "Issue number",
			},
			"body": map[string]any{
				"type":        "string",
				"description": "Comment body in markdown",
			},
		},
		"required": []string{"repository", "issue_number", "body"},
	}
}

func (t *CreateGitHubIssueCommentTool) RequiredScopes() []string {
	return []string{mcp.ScopeGitHubWrite}
}

func (t *CreateGitHubIssueCommentTool) Execute(ctx context.Context, args map[string]any, authCtx *mcp.AuthContext) (*mcp.ToolResult, error) {
	repo, ok := args["repository"].(string)
	if !ok || repo == "" {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent("repository is required")},
			IsError: true,
		}, nil
	}

	issueNum, ok := args["issue_number"].(float64)
	if !ok {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent("issue_number is required")},
			IsError: true,
		}, nil
	}

	body, ok := args["body"].(string)
	if !ok || body == "" {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent("body is required")},
			IsError: true,
		}, nil
	}

	path := fmt.Sprintf("/repos/%s/issues/%d/comments", repo, int(issueNum))

	payload := map[string]string{"body": body}
	payloadJSON, _ := json.Marshal(payload)

	respBody, err := t.client.doRequest(ctx, "POST", path, strings.NewReader(string(payloadJSON)))
	if err != nil {
		log.Error().Err(err).Str("repository", repo).Int("issue", int(issueNum)).Msg("MCP: create_github_issue_comment - failed")
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to create comment: %v", err))},
			IsError: true,
		}, nil
	}

	var comment map[string]any
	if err := json.Unmarshal(respBody, &comment); err != nil {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to parse response: %v", err))},
			IsError: true,
		}, nil
	}

	result := map[string]any{
		"id":         comment["id"],
		"html_url":   comment["html_url"],
		"created_at": comment["created_at"],
	}

	log.Info().
		Str("repository", repo).
		Int("issue", int(issueNum)).
		Msg("MCP: create_github_issue_comment - success")

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{
		Content: []mcp.Content{mcp.TextContent(string(resultJSON))},
	}, nil
}

// ============================================================================
// UPDATE ISSUE LABELS TOOL
// ============================================================================

// UpdateGitHubIssueLabelsTool implements the update_github_issue_labels MCP tool
type UpdateGitHubIssueLabelsTool struct {
	client *GitHubClient
}

// NewUpdateGitHubIssueLabelsTool creates a new update_github_issue_labels tool
func NewUpdateGitHubIssueLabelsTool(client *GitHubClient) *UpdateGitHubIssueLabelsTool {
	return &UpdateGitHubIssueLabelsTool{client: client}
}

func (t *UpdateGitHubIssueLabelsTool) Name() string {
	return "update_github_issue_labels"
}

func (t *UpdateGitHubIssueLabelsTool) Description() string {
	return `Add or remove labels from a GitHub issue.

Parameters:
  - repository: Repository in format "owner/repo" (required)
  - issue_number: Issue number (required)
  - add_labels: Labels to add (comma-separated)
  - remove_labels: Labels to remove (comma-separated)

Returns the updated list of labels.`
}

func (t *UpdateGitHubIssueLabelsTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"repository": map[string]any{
				"type":        "string",
				"description": "Repository in format 'owner/repo'",
			},
			"issue_number": map[string]any{
				"type":        "integer",
				"description": "Issue number",
			},
			"add_labels": map[string]any{
				"type":        "string",
				"description": "Labels to add (comma-separated)",
			},
			"remove_labels": map[string]any{
				"type":        "string",
				"description": "Labels to remove (comma-separated)",
			},
		},
		"required": []string{"repository", "issue_number"},
	}
}

func (t *UpdateGitHubIssueLabelsTool) RequiredScopes() []string {
	return []string{mcp.ScopeGitHubWrite}
}

func (t *UpdateGitHubIssueLabelsTool) Execute(ctx context.Context, args map[string]any, authCtx *mcp.AuthContext) (*mcp.ToolResult, error) {
	repo, ok := args["repository"].(string)
	if !ok || repo == "" {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent("repository is required")},
			IsError: true,
		}, nil
	}

	issueNum, ok := args["issue_number"].(float64)
	if !ok {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent("issue_number is required")},
			IsError: true,
		}, nil
	}

	// Add labels
	if addLabels, ok := args["add_labels"].(string); ok && addLabels != "" {
		labels := strings.Split(addLabels, ",")
		for i := range labels {
			labels[i] = strings.TrimSpace(labels[i])
		}

		path := fmt.Sprintf("/repos/%s/issues/%d/labels", repo, int(issueNum))
		payload := map[string][]string{"labels": labels}
		payloadJSON, _ := json.Marshal(payload)

		_, err := t.client.doRequest(ctx, "POST", path, strings.NewReader(string(payloadJSON)))
		if err != nil {
			log.Error().Err(err).Str("repository", repo).Int("issue", int(issueNum)).Msg("MCP: update_github_issue_labels - failed to add labels")
			return &mcp.ToolResult{
				Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to add labels: %v", err))},
				IsError: true,
			}, nil
		}
	}

	// Remove labels
	if removeLabels, ok := args["remove_labels"].(string); ok && removeLabels != "" {
		labels := strings.Split(removeLabels, ",")
		for _, label := range labels {
			label = strings.TrimSpace(label)
			path := fmt.Sprintf("/repos/%s/issues/%d/labels/%s", repo, int(issueNum), label)

			_, err := t.client.doRequest(ctx, "DELETE", path, nil)
			if err != nil {
				// Log but don't fail - label might not exist
				log.Warn().Err(err).Str("repository", repo).Int("issue", int(issueNum)).Str("label", label).Msg("MCP: update_github_issue_labels - failed to remove label")
			}
		}
	}

	// Get current labels
	path := fmt.Sprintf("/repos/%s/issues/%d/labels", repo, int(issueNum))
	respBody, err := t.client.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to get labels: %v", err))},
			IsError: true,
		}, nil
	}

	var labels []map[string]any
	if err := json.Unmarshal(respBody, &labels); err != nil {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to parse response: %v", err))},
			IsError: true,
		}, nil
	}

	labelNames := make([]string, 0, len(labels))
	for _, l := range labels {
		if name, ok := l["name"].(string); ok {
			labelNames = append(labelNames, name)
		}
	}

	result := map[string]any{
		"issue_number": int(issueNum),
		"labels":       labelNames,
	}

	log.Info().
		Str("repository", repo).
		Int("issue", int(issueNum)).
		Strs("labels", labelNames).
		Msg("MCP: update_github_issue_labels - success")

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{
		Content: []mcp.Content{mcp.TextContent(string(resultJSON))},
	}, nil
}

// ============================================================================
// CREATE ISSUE TOOL
// ============================================================================

// CreateGitHubIssueTool implements the create_github_issue MCP tool
type CreateGitHubIssueTool struct {
	client *GitHubClient
}

// NewCreateGitHubIssueTool creates a new create_github_issue tool
func NewCreateGitHubIssueTool(client *GitHubClient) *CreateGitHubIssueTool {
	return &CreateGitHubIssueTool{client: client}
}

func (t *CreateGitHubIssueTool) Name() string {
	return "create_github_issue"
}

func (t *CreateGitHubIssueTool) Description() string {
	return `Create a new GitHub issue.

Parameters:
  - repository: Repository in format "owner/repo" (required)
  - title: Issue title (required)
  - body: Issue body in markdown
  - labels: Comma-separated list of labels to add
  - assignees: Comma-separated list of usernames to assign

Returns the created issue details.`
}

func (t *CreateGitHubIssueTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"repository": map[string]any{
				"type":        "string",
				"description": "Repository in format 'owner/repo'",
			},
			"title": map[string]any{
				"type":        "string",
				"description": "Issue title",
			},
			"body": map[string]any{
				"type":        "string",
				"description": "Issue body in markdown",
			},
			"labels": map[string]any{
				"type":        "string",
				"description": "Comma-separated list of labels to add",
			},
			"assignees": map[string]any{
				"type":        "string",
				"description": "Comma-separated list of usernames to assign",
			},
		},
		"required": []string{"repository", "title"},
	}
}

func (t *CreateGitHubIssueTool) RequiredScopes() []string {
	return []string{mcp.ScopeGitHubWrite}
}

func (t *CreateGitHubIssueTool) Execute(ctx context.Context, args map[string]any, authCtx *mcp.AuthContext) (*mcp.ToolResult, error) {
	repo, ok := args["repository"].(string)
	if !ok || repo == "" {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent("repository is required")},
			IsError: true,
		}, nil
	}

	title, ok := args["title"].(string)
	if !ok || title == "" {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent("title is required")},
			IsError: true,
		}, nil
	}

	payload := map[string]any{
		"title": title,
	}

	if body, ok := args["body"].(string); ok && body != "" {
		payload["body"] = body
	}

	if labelsStr, ok := args["labels"].(string); ok && labelsStr != "" {
		labels := strings.Split(labelsStr, ",")
		for i := range labels {
			labels[i] = strings.TrimSpace(labels[i])
		}
		payload["labels"] = labels
	}

	if assigneesStr, ok := args["assignees"].(string); ok && assigneesStr != "" {
		assignees := strings.Split(assigneesStr, ",")
		for i := range assignees {
			assignees[i] = strings.TrimSpace(assignees[i])
		}
		payload["assignees"] = assignees
	}

	path := fmt.Sprintf("/repos/%s/issues", repo)
	payloadJSON, _ := json.Marshal(payload)

	respBody, err := t.client.doRequest(ctx, "POST", path, strings.NewReader(string(payloadJSON)))
	if err != nil {
		log.Error().Err(err).Str("repository", repo).Str("title", title).Msg("MCP: create_github_issue - failed")
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to create issue: %v", err))},
			IsError: true,
		}, nil
	}

	var issue map[string]any
	if err := json.Unmarshal(respBody, &issue); err != nil {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to parse response: %v", err))},
			IsError: true,
		}, nil
	}

	result := map[string]any{
		"number":     issue["number"],
		"title":      issue["title"],
		"html_url":   issue["html_url"],
		"created_at": issue["created_at"],
	}

	log.Info().
		Str("repository", repo).
		Float64("issue", issue["number"].(float64)).
		Str("title", title).
		Msg("MCP: create_github_issue - success")

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{
		Content: []mcp.Content{mcp.TextContent(string(resultJSON))},
	}, nil
}

// ============================================================================
// TRIGGER CLAUDE FIX TOOL
// ============================================================================

// TriggerClaudeFixTool implements the trigger_claude_fix MCP tool
type TriggerClaudeFixTool struct {
	client *GitHubClient
}

// NewTriggerClaudeFixTool creates a new trigger_claude_fix tool
func NewTriggerClaudeFixTool(client *GitHubClient) *TriggerClaudeFixTool {
	return &TriggerClaudeFixTool{client: client}
}

func (t *TriggerClaudeFixTool) Name() string {
	return "trigger_claude_fix"
}

func (t *TriggerClaudeFixTool) Description() string {
	return `Trigger the Claude Fix automation for a GitHub issue by adding the claude-fix label.

This will trigger the GitHub Actions workflow that:
1. Creates a new branch
2. Uses Claude to analyze and fix the issue
3. Creates a PR for review

Parameters:
  - repository: Repository in format "owner/repo" (required)
  - issue_number: Issue number to trigger fix for (required)

Returns confirmation of the trigger.`
}

func (t *TriggerClaudeFixTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"repository": map[string]any{
				"type":        "string",
				"description": "Repository in format 'owner/repo'",
			},
			"issue_number": map[string]any{
				"type":        "integer",
				"description": "Issue number to trigger fix for",
			},
		},
		"required": []string{"repository", "issue_number"},
	}
}

func (t *TriggerClaudeFixTool) RequiredScopes() []string {
	return []string{mcp.ScopeGitHubWrite}
}

func (t *TriggerClaudeFixTool) Execute(ctx context.Context, args map[string]any, authCtx *mcp.AuthContext) (*mcp.ToolResult, error) {
	repo, ok := args["repository"].(string)
	if !ok || repo == "" {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent("repository is required")},
			IsError: true,
		}, nil
	}

	issueNum, ok := args["issue_number"].(float64)
	if !ok {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent("issue_number is required")},
			IsError: true,
		}, nil
	}

	// Add the claude-fix label to trigger the workflow
	path := fmt.Sprintf("/repos/%s/issues/%d/labels", repo, int(issueNum))
	payload := map[string][]string{"labels": {"claude-fix"}}
	payloadJSON, _ := json.Marshal(payload)

	_, err := t.client.doRequest(ctx, "POST", path, strings.NewReader(string(payloadJSON)))
	if err != nil {
		log.Error().Err(err).Str("repository", repo).Int("issue", int(issueNum)).Msg("MCP: trigger_claude_fix - failed")
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to trigger Claude fix: %v", err))},
			IsError: true,
		}, nil
	}

	result := map[string]any{
		"status":       "triggered",
		"issue_number": int(issueNum),
		"repository":   repo,
		"message":      "Claude fix workflow triggered. The GitHub Actions workflow will now process this issue.",
	}

	log.Info().
		Str("repository", repo).
		Int("issue", int(issueNum)).
		Msg("MCP: trigger_claude_fix - success")

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{
		Content: []mcp.Content{mcp.TextContent(string(resultJSON))},
	}, nil
}
