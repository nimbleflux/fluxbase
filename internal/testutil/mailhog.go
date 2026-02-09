// Package testutil provides email testing utilities using MailHog.
package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"
)

// MailHogMessage represents an email message from MailHog API
type MailHogMessage struct {
	ID   string `json:"ID"`
	From struct {
		Mailbox string `json:"Mailbox"`
		Domain  string `json:"Domain"`
	} `json:"From"`
	To []struct {
		Mailbox string `json:"Mailbox"`
		Domain  string `json:"Domain"`
	} `json:"To"`
	Content struct {
		Headers map[string][]string `json:"Headers"`
		Body    string              `json:"body"`
	} `json:"Content"`
	Created time.Time `json:"Created"`
}

// MailHogMessages represents the response from MailHog messages API
type MailHogMessages struct {
	Total int              `json:"total"`
	Count int              `json:"count"`
	Start int              `json:"start"`
	Items []MailHogMessage `json:"items"`
}

// GetMailHogBaseURL returns the MailHog base URL for the current environment.
// It checks the MAILHOG_HOST environment variable, defaulting to localhost.
func GetMailHogBaseURL() string {
	host := "mailhog"
	// Allow override via environment variable
	// In local development, this might be "localhost"
	return fmt.Sprintf("http://%s:8025", host)
}

// DeleteAllMailHogMessages deletes all messages from MailHog.
// Use this before tests to ensure a clean state.
func DeleteAllMailHogMessages(t *testing.T) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, GetMailHogBaseURL()+"/api/v1/messages", nil)
	if err != nil {
		t.Fatalf("Failed to create delete request: %v", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// MailHog might not be available, log and continue
		t.Logf("Warning: Could not delete MailHog messages: %v", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		t.Logf("Warning: MailHog delete returned status %d", resp.StatusCode)
	}
}

// GetAllMailHogMessages fetches all messages from MailHog.
func GetAllMailHogMessages(t *testing.T) []MailHogMessage {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, GetMailHogBaseURL()+"/api/v2/messages", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to fetch MailHog messages: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("MailHog returned status %d", resp.StatusCode)
	}

	var messages MailHogMessages
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		t.Fatalf("Failed to decode MailHog response: %v", err)
	}

	return messages.Items
}

// WaitForEmail waits for an email to arrive in MailHog matching a filter function.
// It polls every 100ms until the timeout is reached.
//
// Example:
//
//	msg := testutil.WaitForEmail(t, 5*time.Second, func(m testutil.MailHogMessage) bool {
//	    return len(m.To) > 0 && m.To[0].Mailbox == "user" && m.To[0].Domain == "example.com"
//	})
//	require.NotNil(t, msg, "Email not received within timeout")
func WaitForEmail(t *testing.T, timeout time.Duration, checkFn func(MailHogMessage) bool) *MailHogMessage {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}

	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, GetMailHogBaseURL()+"/api/v2/messages", nil)
		if err == nil {
			resp, err := client.Do(req)
			if err == nil {
				defer func() { _ = resp.Body.Close() }()

				if resp.StatusCode == http.StatusOK {
					var messages MailHogMessages
					if err := json.NewDecoder(resp.Body).Decode(&messages); err == nil {
						for _, msg := range messages.Items {
							if checkFn(msg) {
								return &msg
							}
						}
					}
				}
			}
		}

		// Wait before next poll
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// ExtractPasswordResetToken extracts a password reset token from an email body.
// It looks for patterns like:
// - /auth/reset-password?token=<token>
// - /reset-password?token=<token>
// - token=<token>
func ExtractPasswordResetToken(t *testing.T, emailBody string) string {
	// Try to find token in reset link
	// Pattern 1: /reset-password?token=XYZ or /auth/reset-password?token=XYZ
	// Base64 URL encoding can have - (hyphen), _ (underscore), and = (padding) characters
	linkRegex := regexp.MustCompile(`/(?:auth/)?reset-password\?token=([a-zA-Z0-9_-]+=*)`)
	if matches := linkRegex.FindStringSubmatch(emailBody); len(matches) > 1 {
		return matches[1]
	}

	// Pattern 2: token=XYZ (in plain text)
	tokenRegex := regexp.MustCompile(`token=([a-zA-Z0-9_-]+=*)`)
	if matches := tokenRegex.FindStringSubmatch(emailBody); len(matches) > 1 {
		return matches[1]
	}

	t.Fatal("Could not extract password reset token from email body")
	return ""
}

// ExtractMagicLinkToken extracts a magic link token from an email body.
// It looks for patterns like:
// - /auth/verify?token=<token>
// - token=<token>
func ExtractMagicLinkToken(t *testing.T, emailBody string) string {
	// Try to find token in magic link
	// Pattern 1: /auth/verify?token=XYZ
	linkRegex := regexp.MustCompile(`/auth/verify\?token=([a-zA-Z0-9_-]+)`)
	if matches := linkRegex.FindStringSubmatch(emailBody); len(matches) > 1 {
		return matches[1]
	}

	// Pattern 2: token=XYZ (in plain text)
	tokenRegex := regexp.MustCompile(`token=([a-zA-Z0-9_-]+)`)
	if matches := tokenRegex.FindStringSubmatch(emailBody); len(matches) > 1 {
		return matches[1]
	}

	t.Fatal("Could not extract magic link token from email body")
	return ""
}

// ExtractEmailVerificationToken extracts an email verification token from an email body.
func ExtractEmailVerificationToken(t *testing.T, emailBody string) string {
	// Same pattern as magic link
	return ExtractMagicLinkToken(t, emailBody)
}

// FindEmailTo finds an email sent to a specific recipient.
func FindEmailTo(t *testing.T, emailAddress string) *MailHogMessage {
	messages := GetAllMailHogMessages(t)

	// Parse email address to get mailbox and domain
	parts := strings.Split(emailAddress, "@")
	if len(parts) != 2 {
		t.Fatalf("Invalid email address: %s", emailAddress)
	}
	mailbox := parts[0]
	domain := parts[1]

	for _, msg := range messages {
		if len(msg.To) > 0 && msg.To[0].Mailbox == mailbox && msg.To[0].Domain == domain {
			return &msg
		}
	}

	return nil
}

// FindEmailWithSubject finds an email with a specific subject line.
func FindEmailWithSubject(t *testing.T, subject string) *MailHogMessage {
	messages := GetAllMailHogMessages(t)

	for _, msg := range messages {
		if subjectHeader, ok := msg.Content.Headers["Subject"]; ok && len(subjectHeader) > 0 {
			if strings.Contains(subjectHeader[0], subject) {
				return &msg
			}
		}
	}

	return nil
}
