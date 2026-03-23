package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/mcp"
)

func TestNewGitHubClient(t *testing.T) {
	t.Run("creates client with token", func(t *testing.T) {
		client := NewGitHubClient("test-token")
		assert.NotNil(t, client)
		assert.Equal(t, "test-token", client.token)
		assert.Equal(t, "https://api.github.com", client.baseURL)
		assert.NotNil(t, client.httpClient)
	})

	t.Run("creates client without token", func(t *testing.T) {
		client := NewGitHubClient("")
		assert.NotNil(t, client)
		assert.Empty(t, client.token)
	})
}

func TestGitHubClient_doRequest(t *testing.T) {
	t.Run("successful GET request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/test/path", r.URL.Path)
			assert.Equal(t, "application/vnd.github.v3+json", r.Header.Get("Accept"))
			assert.Equal(t, "Fluxbase-MCP", r.Header.Get("User-Agent"))
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status": "ok"}`))
		}))
		defer server.Close()

		client := NewGitHubClient("test-token")
		client.baseURL = server.URL

		resp, err := client.doRequest(context.Background(), "GET", "/test/path", nil)
		require.NoError(t, err)
		assert.Contains(t, string(resp), "ok")
	})

	t.Run("handles error status code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message": "Not Found"}`))
		}))
		defer server.Close()

		client := NewGitHubClient("test-token")
		client.baseURL = server.URL

		_, err := client.doRequest(context.Background(), "GET", "/not-found", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status 404")
	})

	t.Run("skips auth header without token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Empty(t, r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}))
		defer server.Close()

		client := NewGitHubClient("")
		client.baseURL = server.URL

		_, err := client.doRequest(context.Background(), "GET", "/public", nil)
		require.NoError(t, err)
	})
}

func TestNewListGitHubIssuesTool(t *testing.T) {
	client := NewGitHubClient("token")
	tool := NewListGitHubIssuesTool(client)
	assert.NotNil(t, tool)
	assert.Equal(t, client, tool.client)
}

func TestListGitHubIssuesTool_Metadata(t *testing.T) {
	tool := NewListGitHubIssuesTool(NewGitHubClient(""))

	t.Run("name", func(t *testing.T) {
		assert.Equal(t, "list_github_issues", tool.Name())
	})

	t.Run("description", func(t *testing.T) {
		desc := tool.Description()
		assert.Contains(t, desc, "List GitHub issues")
		assert.Contains(t, desc, "repository")
	})

	t.Run("input schema", func(t *testing.T) {
		schema := tool.InputSchema()
		assert.Equal(t, "object", schema["type"])

		props := schema["properties"].(map[string]any)
		assert.Contains(t, props, "repository")
		assert.Contains(t, props, "state")
		assert.Contains(t, props, "labels")
		assert.Contains(t, props, "assignee")
		assert.Contains(t, props, "limit")

		required := schema["required"].([]string)
		assert.Contains(t, required, "repository")
	})

	t.Run("required scopes", func(t *testing.T) {
		scopes := tool.RequiredScopes()
		assert.Contains(t, scopes, mcp.ScopeGitHubRead)
	})
}

func TestListGitHubIssuesTool_Execute(t *testing.T) {
	t.Run("missing repository", func(t *testing.T) {
		tool := NewListGitHubIssuesTool(NewGitHubClient(""))
		result, err := tool.Execute(context.Background(), map[string]any{}, nil)
		require.NoError(t, err)
		assert.True(t, result.IsError)
		assert.Contains(t, result.Content[0].Text, "repository is required")
	})

	t.Run("successful list issues", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Contains(t, r.URL.Path, "/repos/owner/repo/issues")
			assert.Contains(t, r.URL.RawQuery, "state=open")
			assert.Contains(t, r.URL.RawQuery, "per_page=30")

			issues := []map[string]any{
				{
					"number":     1,
					"title":      "Test Issue",
					"state":      "open",
					"html_url":   "https://github.com/owner/repo/issues/1",
					"created_at": "2024-01-15T10:00:00Z",
					"updated_at": "2024-01-15T10:00:00Z",
					"labels": []map[string]any{
						{"name": "bug"},
					},
					"assignees": []map[string]any{
						{"login": "user1"},
					},
				},
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(issues)
		}))
		defer server.Close()

		client := NewGitHubClient("token")
		client.baseURL = server.URL
		tool := NewListGitHubIssuesTool(client)

		result, err := tool.Execute(context.Background(), map[string]any{
			"repository": "owner/repo",
		}, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError)

		content := result.Content[0].Text
		assert.Contains(t, content, "Test Issue")
		assert.Contains(t, content, "bug")
	})

	t.Run("filters out pull requests", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			issues := []map[string]any{
				{"number": 1, "title": "Issue", "state": "open"},
				{"number": 2, "title": "PR", "state": "open", "pull_request": map[string]any{}},
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(issues)
		}))
		defer server.Close()

		client := NewGitHubClient("token")
		client.baseURL = server.URL
		tool := NewListGitHubIssuesTool(client)

		result, err := tool.Execute(context.Background(), map[string]any{
			"repository": "owner/repo",
		}, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError)

		content := result.Content[0].Text
		assert.Contains(t, content, "Issue")
		assert.NotContains(t, content, "PR")
	})

	t.Run("with all parameters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Contains(t, r.URL.RawQuery, "state=closed")
			assert.Contains(t, r.URL.RawQuery, "labels=bug,enhancement")
			assert.Contains(t, r.URL.RawQuery, "assignee=testuser")
			assert.Contains(t, r.URL.RawQuery, "per_page=50")

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		}))
		defer server.Close()

		client := NewGitHubClient("token")
		client.baseURL = server.URL
		tool := NewListGitHubIssuesTool(client)

		result, err := tool.Execute(context.Background(), map[string]any{
			"repository": "owner/repo",
			"state":      "closed",
			"labels":     "bug,enhancement",
			"assignee":   "testuser",
			"limit":      float64(50),
		}, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError)
	})

	t.Run("limit capped at 100", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Contains(t, r.URL.RawQuery, "per_page=100")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		}))
		defer server.Close()

		client := NewGitHubClient("token")
		client.baseURL = server.URL
		tool := NewListGitHubIssuesTool(client)

		result, err := tool.Execute(context.Background(), map[string]any{
			"repository": "owner/repo",
			"limit":      float64(200),
		}, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError)
	})
}

func TestNewGetGitHubIssueTool(t *testing.T) {
	client := NewGitHubClient("token")
	tool := NewGetGitHubIssueTool(client)
	assert.NotNil(t, tool)
	assert.Equal(t, client, tool.client)
}

func TestGetGitHubIssueTool_Metadata(t *testing.T) {
	tool := NewGetGitHubIssueTool(NewGitHubClient(""))

	t.Run("name", func(t *testing.T) {
		assert.Equal(t, "get_github_issue", tool.Name())
	})

	t.Run("description", func(t *testing.T) {
		desc := tool.Description()
		assert.Contains(t, desc, "Get details")
		assert.Contains(t, desc, "issue_number")
	})

	t.Run("input schema", func(t *testing.T) {
		schema := tool.InputSchema()
		props := schema["properties"].(map[string]any)
		assert.Contains(t, props, "repository")
		assert.Contains(t, props, "issue_number")

		required := schema["required"].([]string)
		assert.Contains(t, required, "repository")
		assert.Contains(t, required, "issue_number")
	})

	t.Run("required scopes", func(t *testing.T) {
		scopes := tool.RequiredScopes()
		assert.Contains(t, scopes, mcp.ScopeGitHubRead)
	})
}

func TestGetGitHubIssueTool_Execute(t *testing.T) {
	t.Run("missing repository", func(t *testing.T) {
		tool := NewGetGitHubIssueTool(NewGitHubClient(""))
		result, err := tool.Execute(context.Background(), map[string]any{
			"issue_number": float64(1),
		}, nil)
		require.NoError(t, err)
		assert.True(t, result.IsError)
	})

	t.Run("missing issue_number", func(t *testing.T) {
		tool := NewGetGitHubIssueTool(NewGitHubClient(""))
		result, err := tool.Execute(context.Background(), map[string]any{
			"repository": "owner/repo",
		}, nil)
		require.NoError(t, err)
		assert.True(t, result.IsError)
	})

	t.Run("successful get issue", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/repos/owner/repo/issues/42", r.URL.Path)

			issue := map[string]any{
				"number":     42,
				"title":      "Test Issue",
				"state":      "open",
				"body":       "Issue description",
				"html_url":   "https://github.com/owner/repo/issues/42",
				"comments":   5,
				"created_at": "2024-01-15T10:00:00Z",
				"updated_at": "2024-01-15T12:00:00Z",
				"user":       map[string]any{"login": "author"},
				"labels": []map[string]any{
					{"name": "bug"},
				},
				"assignees": []map[string]any{
					{"login": "dev1"},
				},
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(issue)
		}))
		defer server.Close()

		client := NewGitHubClient("token")
		client.baseURL = server.URL
		tool := NewGetGitHubIssueTool(client)

		result, err := tool.Execute(context.Background(), map[string]any{
			"repository":   "owner/repo",
			"issue_number": float64(42),
		}, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError)

		content := result.Content[0].Text
		assert.Contains(t, content, "Test Issue")
		assert.Contains(t, content, "Issue description")
		assert.Contains(t, content, "author")
	})
}

func TestNewCreateGitHubIssueCommentTool(t *testing.T) {
	client := NewGitHubClient("token")
	tool := NewCreateGitHubIssueCommentTool(client)
	assert.NotNil(t, tool)
	assert.Equal(t, client, tool.client)
}

func TestCreateGitHubIssueCommentTool_Metadata(t *testing.T) {
	tool := NewCreateGitHubIssueCommentTool(NewGitHubClient(""))

	t.Run("name", func(t *testing.T) {
		assert.Equal(t, "create_github_issue_comment", tool.Name())
	})

	t.Run("description", func(t *testing.T) {
		desc := tool.Description()
		assert.Contains(t, desc, "Add a comment")
	})

	t.Run("input schema", func(t *testing.T) {
		schema := tool.InputSchema()
		props := schema["properties"].(map[string]any)
		assert.Contains(t, props, "repository")
		assert.Contains(t, props, "issue_number")
		assert.Contains(t, props, "body")

		required := schema["required"].([]string)
		assert.Contains(t, required, "body")
	})

	t.Run("required scopes", func(t *testing.T) {
		scopes := tool.RequiredScopes()
		assert.Contains(t, scopes, mcp.ScopeGitHubWrite)
	})
}

func TestCreateGitHubIssueCommentTool_Execute(t *testing.T) {
	t.Run("missing repository", func(t *testing.T) {
		tool := NewCreateGitHubIssueCommentTool(NewGitHubClient(""))
		result, err := tool.Execute(context.Background(), map[string]any{
			"issue_number": float64(1),
			"body":         "comment",
		}, nil)
		require.NoError(t, err)
		assert.True(t, result.IsError)
	})

	t.Run("missing body", func(t *testing.T) {
		tool := NewCreateGitHubIssueCommentTool(NewGitHubClient(""))
		result, err := tool.Execute(context.Background(), map[string]any{
			"repository":   "owner/repo",
			"issue_number": float64(1),
		}, nil)
		require.NoError(t, err)
		assert.True(t, result.IsError)
	})

	t.Run("successful create comment", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/repos/owner/repo/issues/1/comments", r.URL.Path)

			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			assert.Equal(t, "Test comment", body["body"])

			comment := map[string]any{
				"id":         12345,
				"html_url":   "https://github.com/owner/repo/issues/1#comment-12345",
				"created_at": "2024-01-15T10:00:00Z",
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(comment)
		}))
		defer server.Close()

		client := NewGitHubClient("token")
		client.baseURL = server.URL
		tool := NewCreateGitHubIssueCommentTool(client)

		result, err := tool.Execute(context.Background(), map[string]any{
			"repository":   "owner/repo",
			"issue_number": float64(1),
			"body":         "Test comment",
		}, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError)

		content := result.Content[0].Text
		assert.Contains(t, content, "12345")
	})
}

func TestNewUpdateGitHubIssueLabelsTool(t *testing.T) {
	client := NewGitHubClient("token")
	tool := NewUpdateGitHubIssueLabelsTool(client)
	assert.NotNil(t, tool)
	assert.Equal(t, client, tool.client)
}

func TestUpdateGitHubIssueLabelsTool_Metadata(t *testing.T) {
	tool := NewUpdateGitHubIssueLabelsTool(NewGitHubClient(""))

	t.Run("name", func(t *testing.T) {
		assert.Equal(t, "update_github_issue_labels", tool.Name())
	})

	t.Run("description", func(t *testing.T) {
		desc := tool.Description()
		assert.Contains(t, desc, "Add or remove labels")
	})

	t.Run("input schema", func(t *testing.T) {
		schema := tool.InputSchema()
		props := schema["properties"].(map[string]any)
		assert.Contains(t, props, "repository")
		assert.Contains(t, props, "issue_number")
		assert.Contains(t, props, "add_labels")
		assert.Contains(t, props, "remove_labels")
	})

	t.Run("required scopes", func(t *testing.T) {
		scopes := tool.RequiredScopes()
		assert.Contains(t, scopes, mcp.ScopeGitHubWrite)
	})
}

func TestUpdateGitHubIssueLabelsTool_Execute(t *testing.T) {
	t.Run("missing repository", func(t *testing.T) {
		tool := NewUpdateGitHubIssueLabelsTool(NewGitHubClient(""))
		result, err := tool.Execute(context.Background(), map[string]any{
			"issue_number": float64(1),
		}, nil)
		require.NoError(t, err)
		assert.True(t, result.IsError)
	})

	t.Run("add labels", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			if r.Method == "POST" {
				var body map[string][]string
				_ = json.NewDecoder(r.Body).Decode(&body)
				assert.Contains(t, body["labels"], "bug")
				assert.Contains(t, body["labels"], "enhancement")
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"name": "bug"},
				{"name": "enhancement"},
			})
		}))
		defer server.Close()

		client := NewGitHubClient("token")
		client.baseURL = server.URL
		tool := NewUpdateGitHubIssueLabelsTool(client)

		result, err := tool.Execute(context.Background(), map[string]any{
			"repository":   "owner/repo",
			"issue_number": float64(1),
			"add_labels":   "bug, enhancement",
		}, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError)
	})

	t.Run("remove labels", func(t *testing.T) {
		deleteRequests := []string{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "DELETE" {
				deleteRequests = append(deleteRequests, r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		}))
		defer server.Close()

		client := NewGitHubClient("token")
		client.baseURL = server.URL
		tool := NewUpdateGitHubIssueLabelsTool(client)

		result, err := tool.Execute(context.Background(), map[string]any{
			"repository":    "owner/repo",
			"issue_number":  float64(1),
			"remove_labels": "old-label",
		}, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError)
		assert.Contains(t, deleteRequests[0], "old-label")
	})
}

func TestNewCreateGitHubIssueTool(t *testing.T) {
	client := NewGitHubClient("token")
	tool := NewCreateGitHubIssueTool(client)
	assert.NotNil(t, tool)
	assert.Equal(t, client, tool.client)
}

func TestCreateGitHubIssueTool_Metadata(t *testing.T) {
	tool := NewCreateGitHubIssueTool(NewGitHubClient(""))

	t.Run("name", func(t *testing.T) {
		assert.Equal(t, "create_github_issue", tool.Name())
	})

	t.Run("description", func(t *testing.T) {
		desc := tool.Description()
		assert.Contains(t, desc, "Create a new GitHub issue")
	})

	t.Run("input schema", func(t *testing.T) {
		schema := tool.InputSchema()
		props := schema["properties"].(map[string]any)
		assert.Contains(t, props, "repository")
		assert.Contains(t, props, "title")
		assert.Contains(t, props, "body")
		assert.Contains(t, props, "labels")
		assert.Contains(t, props, "assignees")

		required := schema["required"].([]string)
		assert.Contains(t, required, "repository")
		assert.Contains(t, required, "title")
	})

	t.Run("required scopes", func(t *testing.T) {
		scopes := tool.RequiredScopes()
		assert.Contains(t, scopes, mcp.ScopeGitHubWrite)
	})
}

func TestCreateGitHubIssueTool_Execute(t *testing.T) {
	t.Run("missing repository", func(t *testing.T) {
		tool := NewCreateGitHubIssueTool(NewGitHubClient(""))
		result, err := tool.Execute(context.Background(), map[string]any{
			"title": "Test",
		}, nil)
		require.NoError(t, err)
		assert.True(t, result.IsError)
	})

	t.Run("missing title", func(t *testing.T) {
		tool := NewCreateGitHubIssueTool(NewGitHubClient(""))
		result, err := tool.Execute(context.Background(), map[string]any{
			"repository": "owner/repo",
		}, nil)
		require.NoError(t, err)
		assert.True(t, result.IsError)
	})

	t.Run("successful create issue", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/repos/owner/repo/issues", r.URL.Path)

			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			assert.Equal(t, "New Issue", body["title"])
			assert.Equal(t, "Issue body", body["body"])

			issue := map[string]any{
				"number":     99,
				"title":      "New Issue",
				"html_url":   "https://github.com/owner/repo/issues/99",
				"created_at": "2024-01-15T10:00:00Z",
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(issue)
		}))
		defer server.Close()

		client := NewGitHubClient("token")
		client.baseURL = server.URL
		tool := NewCreateGitHubIssueTool(client)

		result, err := tool.Execute(context.Background(), map[string]any{
			"repository": "owner/repo",
			"title":      "New Issue",
			"body":       "Issue body",
		}, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError)

		content := result.Content[0].Text
		assert.Contains(t, content, "99")
		assert.Contains(t, content, "New Issue")
	})

	t.Run("with labels and assignees", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)

			labels := body["labels"].([]any)
			assert.Len(t, labels, 2)

			assignees := body["assignees"].([]any)
			assert.Len(t, assignees, 1)

			issue := map[string]any{
				"number":     100,
				"title":      "Test",
				"html_url":   "https://github.com/owner/repo/issues/100",
				"created_at": "2024-01-15T10:00:00Z",
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(issue)
		}))
		defer server.Close()

		client := NewGitHubClient("token")
		client.baseURL = server.URL
		tool := NewCreateGitHubIssueTool(client)

		result, err := tool.Execute(context.Background(), map[string]any{
			"repository": "owner/repo",
			"title":      "Test",
			"labels":     "bug, enhancement",
			"assignees":  "user1",
		}, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError)
	})
}

func TestNewTriggerClaudeFixTool(t *testing.T) {
	client := NewGitHubClient("token")
	tool := NewTriggerClaudeFixTool(client)
	assert.NotNil(t, tool)
	assert.Equal(t, client, tool.client)
}

func TestTriggerClaudeFixTool_Metadata(t *testing.T) {
	tool := NewTriggerClaudeFixTool(NewGitHubClient(""))

	t.Run("name", func(t *testing.T) {
		assert.Equal(t, "trigger_claude_fix", tool.Name())
	})

	t.Run("description", func(t *testing.T) {
		desc := tool.Description()
		assert.Contains(t, desc, "Trigger the Claude Fix automation")
		assert.Contains(t, desc, "claude-fix label")
	})

	t.Run("input schema", func(t *testing.T) {
		schema := tool.InputSchema()
		props := schema["properties"].(map[string]any)
		assert.Contains(t, props, "repository")
		assert.Contains(t, props, "issue_number")

		required := schema["required"].([]string)
		assert.Contains(t, required, "repository")
		assert.Contains(t, required, "issue_number")
	})

	t.Run("required scopes", func(t *testing.T) {
		scopes := tool.RequiredScopes()
		assert.Contains(t, scopes, mcp.ScopeGitHubWrite)
	})
}

func TestTriggerClaudeFixTool_Execute(t *testing.T) {
	t.Run("missing repository", func(t *testing.T) {
		tool := NewTriggerClaudeFixTool(NewGitHubClient(""))
		result, err := tool.Execute(context.Background(), map[string]any{
			"issue_number": float64(1),
		}, nil)
		require.NoError(t, err)
		assert.True(t, result.IsError)
	})

	t.Run("missing issue_number", func(t *testing.T) {
		tool := NewTriggerClaudeFixTool(NewGitHubClient(""))
		result, err := tool.Execute(context.Background(), map[string]any{
			"repository": "owner/repo",
		}, nil)
		require.NoError(t, err)
		assert.True(t, result.IsError)
	})

	t.Run("successful trigger", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Contains(t, r.URL.Path, "/labels")

			var body map[string][]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			assert.Contains(t, body["labels"], "claude-fix")

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]map[string]any{{"name": "claude-fix"}})
		}))
		defer server.Close()

		client := NewGitHubClient("token")
		client.baseURL = server.URL
		tool := NewTriggerClaudeFixTool(client)

		result, err := tool.Execute(context.Background(), map[string]any{
			"repository":   "owner/repo",
			"issue_number": float64(5),
		}, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError)

		content := result.Content[0].Text
		assert.Contains(t, content, "triggered")
	})
}
