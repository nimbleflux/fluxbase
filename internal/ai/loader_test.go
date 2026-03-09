package ai

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLoader(t *testing.T) {
	t.Run("creates loader with specified directory", func(t *testing.T) {
		loader := NewLoader("/path/to/chatbots")
		assert.Equal(t, "/path/to/chatbots", loader.chatbotsDir)
		assert.Equal(t, "/path/to/chatbots", loader.GetChatbotsDir())
	})
}

func TestLoader_LoadAll(t *testing.T) {
	t.Run("returns nil for non-existent directory", func(t *testing.T) {
		loader := NewLoader("/nonexistent/path/to/chatbots")
		chatbots, err := loader.LoadAll()
		require.NoError(t, err)
		assert.Nil(t, chatbots)
	})

	t.Run("returns error if path is a file not directory", func(t *testing.T) {
		// Create a temporary file
		tmpFile, err := os.CreateTemp("", "test-chatbot-*.txt")
		require.NoError(t, err)
		defer func() { _ = os.Remove(tmpFile.Name()) }()
		_ = tmpFile.Close()

		loader := NewLoader(tmpFile.Name())
		chatbots, err := loader.LoadAll()
		require.Error(t, err)
		assert.Nil(t, chatbots)
		assert.Contains(t, err.Error(), "not a directory")
	})

	t.Run("loads chatbots from valid directory", func(t *testing.T) {
		// Create a temporary directory structure
		tmpDir, err := os.MkdirTemp("", "chatbots-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		// Create chatbot directory and file
		chatbotDir := filepath.Join(tmpDir, "test-bot")
		err = os.MkdirAll(chatbotDir, 0o755)
		require.NoError(t, err)

		chatbotCode := `
/**
 * Test chatbot
 * @description A test chatbot for unit tests
 */
export default async function handler(ctx) {
	return { message: "Hello" };
}
`
		err = os.WriteFile(filepath.Join(chatbotDir, "index.ts"), []byte(chatbotCode), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)
		chatbots, err := loader.LoadAll()
		require.NoError(t, err)
		assert.Len(t, chatbots, 1)
		assert.Equal(t, "test-bot", chatbots[0].Name)
		assert.Equal(t, "default", chatbots[0].Namespace)
		assert.Equal(t, "filesystem", chatbots[0].Source)
		assert.True(t, chatbots[0].Enabled)
	})

	t.Run("loads nested chatbots with namespace", func(t *testing.T) {
		// Create a temporary directory structure
		tmpDir, err := os.MkdirTemp("", "chatbots-nested-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		// Create nested chatbot directory
		chatbotDir := filepath.Join(tmpDir, "analytics", "reports")
		err = os.MkdirAll(chatbotDir, 0o755)
		require.NoError(t, err)

		chatbotCode := `export default async function handler(ctx) { return {}; }`
		err = os.WriteFile(filepath.Join(chatbotDir, "index.ts"), []byte(chatbotCode), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)
		chatbots, err := loader.LoadAll()
		require.NoError(t, err)
		assert.Len(t, chatbots, 1)
		assert.Equal(t, "reports", chatbots[0].Name)
		assert.Equal(t, "analytics", chatbots[0].Namespace)
	})

	t.Run("skips hidden directories", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "chatbots-hidden-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		// Create hidden directory
		hiddenDir := filepath.Join(tmpDir, ".hidden")
		err = os.MkdirAll(hiddenDir, 0o755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(hiddenDir, "index.ts"), []byte("export default {}"), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)
		chatbots, err := loader.LoadAll()
		require.NoError(t, err)
		assert.Empty(t, chatbots)
	})

	t.Run("skips node_modules directory", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "chatbots-nodemodules-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		// Create node_modules directory
		nodeModulesDir := filepath.Join(tmpDir, "node_modules", "some-package")
		err = os.MkdirAll(nodeModulesDir, 0o755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(nodeModulesDir, "index.ts"), []byte("export default {}"), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)
		chatbots, err := loader.LoadAll()
		require.NoError(t, err)
		assert.Empty(t, chatbots)
	})

	t.Run("skips _shared directory", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "chatbots-shared-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		// Create _shared directory
		sharedDir := filepath.Join(tmpDir, "_shared")
		err = os.MkdirAll(sharedDir, 0o755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(sharedDir, "index.ts"), []byte("export default {}"), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)
		chatbots, err := loader.LoadAll()
		require.NoError(t, err)
		assert.Empty(t, chatbots)
	})

	t.Run("only processes index.ts files", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "chatbots-indexts-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		// Create directory with non-index.ts file
		botDir := filepath.Join(tmpDir, "bot")
		err = os.MkdirAll(botDir, 0o755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(botDir, "other.ts"), []byte("export default {}"), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)
		chatbots, err := loader.LoadAll()
		require.NoError(t, err)
		assert.Empty(t, chatbots)
	})
}

func TestLoader_LoadOne(t *testing.T) {
	t.Run("loads chatbot from default namespace", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "chatbots-loadone-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		// Create chatbot
		botDir := filepath.Join(tmpDir, "my-bot")
		err = os.MkdirAll(botDir, 0o755)
		require.NoError(t, err)

		chatbotCode := `
/**
 * @description My test bot
 */
export default async function handler(ctx) { return {}; }
`
		err = os.WriteFile(filepath.Join(botDir, "index.ts"), []byte(chatbotCode), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)
		chatbot, err := loader.LoadOne("default", "my-bot")
		require.NoError(t, err)
		assert.NotNil(t, chatbot)
		assert.Equal(t, "my-bot", chatbot.Name)
	})

	t.Run("loads chatbot from custom namespace", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "chatbots-loadone-ns-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		// Create namespaced chatbot
		botDir := filepath.Join(tmpDir, "custom-ns", "my-bot")
		err = os.MkdirAll(botDir, 0o755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(botDir, "index.ts"), []byte("export default {}"), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)
		chatbot, err := loader.LoadOne("custom-ns", "my-bot")
		require.NoError(t, err)
		assert.NotNil(t, chatbot)
	})

	t.Run("returns error for non-existent chatbot", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "chatbots-loadone-notfound-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		loader := NewLoader(tmpDir)
		chatbot, err := loader.LoadOne("default", "nonexistent")
		require.Error(t, err)
		assert.Nil(t, chatbot)
		assert.Contains(t, err.Error(), "chatbot not found")
	})
}

func TestLoader_ChatbotExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "chatbots-exists-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create one chatbot
	botDir := filepath.Join(tmpDir, "existing-bot")
	err = os.MkdirAll(botDir, 0o755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(botDir, "index.ts"), []byte("export default {}"), 0o644)
	require.NoError(t, err)

	loader := NewLoader(tmpDir)

	t.Run("returns true for existing chatbot", func(t *testing.T) {
		assert.True(t, loader.ChatbotExists("default", "existing-bot"))
	})

	t.Run("returns false for non-existent chatbot", func(t *testing.T) {
		assert.False(t, loader.ChatbotExists("default", "nonexistent"))
	})

	t.Run("returns false for wrong namespace", func(t *testing.T) {
		assert.False(t, loader.ChatbotExists("wrong-ns", "existing-bot"))
	})
}

func TestLoader_WatchForChanges(t *testing.T) {
	t.Run("watches directory for file changes", func(t *testing.T) {
		// Create temporary directory for test
		tmpDir := t.TempDir()
		loader := NewLoader(tmpDir)

		// Create context with cancellation
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Start watching
		changes, err := loader.WatchForChanges(ctx)
		require.NoError(t, err)
		require.NotNil(t, changes)

		// Give the watcher time to start
		time.Sleep(100 * time.Millisecond)

		// Create a chatbot file
		namespaceDir := filepath.Join(tmpDir, "test")
		err = os.MkdirAll(namespaceDir, 0o755)
		require.NoError(t, err)

		// Give the watcher time to detect the new directory
		time.Sleep(100 * time.Millisecond)

		chatbotFile := filepath.Join(namespaceDir, "mybot.ts")
		err = os.WriteFile(chatbotFile, []byte("export default function() {}"), 0o644)
		require.NoError(t, err)

		// Wait for change event (with timeout)
		// Note: fsnotify behavior varies by platform - we may get "created" or "modified"
		select {
		case change := <-changes:
			assert.Contains(t, []string{"created", "modified"}, change.Type)
			assert.Equal(t, "test", change.Namespace)
			assert.Equal(t, "mybot", change.Name)
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for file change event")
		}

		// Modify the file
		err = os.WriteFile(chatbotFile, []byte("export default function() { return 'updated'; }"), 0o644)
		require.NoError(t, err)

		// Wait for modify event
		select {
		case change := <-changes:
			assert.Equal(t, "modified", change.Type)
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for modify event")
		}

		// Cancel context to stop watcher
		cancel()

		// Verify channel is closed after a short delay
		select {
		case _, ok := <-changes:
			if !ok {
				return // Channel closed as expected
			}
		case <-time.After(500 * time.Millisecond):
			// Expected - watcher should stop
		}
	})

	t.Run("filters non-js/ts files", func(t *testing.T) {
		tmpDir := t.TempDir()
		loader := NewLoader(tmpDir)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		changes, err := loader.WatchForChanges(ctx)
		require.NoError(t, err)

		// Give the watcher time to start
		time.Sleep(100 * time.Millisecond)

		// Create a non-JS file
		testFile := filepath.Join(tmpDir, "README.md")
		err = os.WriteFile(testFile, []byte("# Test"), 0o644)
		require.NoError(t, err)

		// Should not receive any events
		select {
		case <-changes:
			t.Fatal("Should not receive events for non-JS/TS files")
		case <-time.After(300 * time.Millisecond):
			// Expected - no events
		}
	})
}

func TestLoader_ParseChatbotFromCode(t *testing.T) {
	loader := NewLoader("/path/to/chatbots")

	t.Run("parses chatbot from code string", func(t *testing.T) {
		code := `
/**
 * Test chatbot for parsing
 * @fluxbase:allowed-tables users, orders
 * @fluxbase:allowed-operations SELECT
 */
export default async function handler(ctx) { return {}; }
`
		chatbot, err := loader.ParseChatbotFromCode(code, "test-namespace")
		require.NoError(t, err)
		assert.NotNil(t, chatbot)
		assert.Equal(t, "test-namespace", chatbot.Namespace)
		assert.Equal(t, "sdk", chatbot.Source)
		assert.True(t, chatbot.Enabled)
		assert.Contains(t, chatbot.Description, "Test chatbot for parsing")
	})

	t.Run("parses chatbot with minimal code", func(t *testing.T) {
		code := `export default async function handler(ctx) { return {}; }`
		chatbot, err := loader.ParseChatbotFromCode(code, "minimal")
		require.NoError(t, err)
		assert.NotNil(t, chatbot)
		assert.Equal(t, "minimal", chatbot.Namespace)
	})
}

func TestChatbotChange_Struct(t *testing.T) {
	change := ChatbotChange{
		Type:      "modified",
		Namespace: "test",
		Name:      "my-bot",
		Path:      "/path/to/chatbot",
	}

	assert.Equal(t, "modified", change.Type)
	assert.Equal(t, "test", change.Namespace)
	assert.Equal(t, "my-bot", change.Name)
	assert.Equal(t, "/path/to/chatbot", change.Path)
}
