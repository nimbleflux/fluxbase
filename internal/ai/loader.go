package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Loader handles loading chatbot definitions from the filesystem
type Loader struct {
	chatbotsDir string
}

// NewLoader creates a new chatbot loader
func NewLoader(chatbotsDir string) *Loader {
	return &Loader{
		chatbotsDir: chatbotsDir,
	}
}

// LoadAll loads all chatbots from the chatbots directory
func (l *Loader) LoadAll() ([]*Chatbot, error) {
	// Check if directory exists
	info, err := os.Stat(l.chatbotsDir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warn().Str("dir", l.chatbotsDir).Msg("Chatbots directory does not exist, skipping load")
			return nil, nil
		}
		return nil, fmt.Errorf("failed to stat chatbots directory: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("chatbots path is not a directory: %s", l.chatbotsDir)
	}

	var chatbots []*Chatbot

	// Walk the directory looking for chatbot definitions
	err = filepath.Walk(l.chatbotsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, but continue walking into them
		if info.IsDir() {
			// Skip hidden directories and node_modules
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "node_modules" || info.Name() == "_shared" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process index.ts files
		if info.Name() != "index.ts" {
			return nil
		}

		// Load the chatbot
		chatbot, err := l.loadChatbot(path)
		if err != nil {
			log.Warn().Err(err).Str("path", path).Msg("Failed to load chatbot")
			return nil // Continue with other chatbots
		}

		chatbots = append(chatbots, chatbot)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk chatbots directory: %w", err)
	}

	log.Info().Int("count", len(chatbots)).Str("dir", l.chatbotsDir).Msg("Loaded chatbots from filesystem")

	return chatbots, nil
}

// loadChatbot loads a single chatbot from a file
func (l *Loader) loadChatbot(path string) (*Chatbot, error) {
	// Read the file
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	code := string(content)

	// Extract chatbot name from directory name
	// e.g., /path/to/chatbots/location-assistant/index.ts -> location-assistant
	dir := filepath.Dir(path)
	name := filepath.Base(dir)

	// Determine namespace from parent directory (if nested)
	// e.g., /path/to/chatbots/analytics/reports/index.ts -> analytics
	relPath, err := filepath.Rel(l.chatbotsDir, dir)
	if err != nil {
		relPath = name
	}

	namespace := "default"
	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) > 1 {
		namespace = parts[0]
		name = parts[len(parts)-1]
	}

	// Parse configuration from annotations
	config := ParseChatbotConfig(code)

	// Parse description from JSDoc
	description := ParseDescription(code)

	// Create chatbot
	chatbot := &Chatbot{
		ID:           uuid.New().String(),
		Name:         name,
		Namespace:    namespace,
		Description:  description,
		Code:         code,
		OriginalCode: code,
		IsBundled:    false, // Not bundled yet
		Enabled:      true,
		Source:       "filesystem",
		Version:      1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Apply parsed configuration
	chatbot.ApplyConfig(config)

	log.Debug().
		Str("name", name).
		Str("namespace", namespace).
		Strs("allowed_tables", chatbot.AllowedTables).
		Strs("allowed_operations", chatbot.AllowedOperations).
		Int("max_tokens", chatbot.MaxTokens).
		Float64("temperature", chatbot.Temperature).
		Msg("Loaded chatbot from filesystem")

	return chatbot, nil
}

// LoadOne loads a single chatbot by name and namespace
func (l *Loader) LoadOne(namespace, name string) (*Chatbot, error) {
	var path string

	if namespace == "default" {
		path = filepath.Join(l.chatbotsDir, name, "index.ts")
	} else {
		path = filepath.Join(l.chatbotsDir, namespace, name, "index.ts")
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("chatbot not found: %s/%s", namespace, name)
	}

	return l.loadChatbot(path)
}

// GetChatbotsDir returns the chatbots directory
func (l *Loader) GetChatbotsDir() string {
	return l.chatbotsDir
}

// ChatbotExists checks if a chatbot exists in the filesystem
func (l *Loader) ChatbotExists(namespace, name string) bool {
	var path string

	if namespace == "default" {
		path = filepath.Join(l.chatbotsDir, name, "index.ts")
	} else {
		path = filepath.Join(l.chatbotsDir, namespace, name, "index.ts")
	}

	_, err := os.Stat(path)
	return err == nil
}

// WatchForChanges sets up a file watcher for the chatbots directory
// Returns a channel that receives updates when files change
// The caller is responsible for closing the returned channel by cancelling the context
func (l *Loader) WatchForChanges(ctx context.Context) (<-chan ChatbotChange, error) {
	// Create fsnotify watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	// Create channel for changes
	changes := make(chan ChatbotChange, 10)

	// Start watching goroutine
	go l.watchLoop(ctx, watcher, changes)

	return changes, nil
}

// watchLoop runs the file watching loop
func (l *Loader) watchLoop(ctx context.Context, watcher *fsnotify.Watcher, changes chan<- ChatbotChange) {
	defer close(changes)
	defer func() { _ = watcher.Close() }()

	// Track pending changes with debounce timer
	var (
		pendingMutex  sync.Mutex
		pendingChange *ChatbotChange
		debounceTimer *time.Timer
	)

	// Track watched directories to avoid duplicates
	watchedDirs := make(map[string]bool)
	var watchedDirsMutex sync.Mutex

	// Helper to add directory to watcher (idempotent)
	addWatchDir := func(dir string) error {
		watchedDirsMutex.Lock()
		defer watchedDirsMutex.Unlock()

		if watchedDirs[dir] {
			return nil
		}

		if err := watcher.Add(dir); err != nil {
			return err
		}

		watchedDirs[dir] = true
		log.Debug().Str("dir", dir).Msg("Added directory to watcher")
		return nil
	}

	// Helper to send change with debouncing
	sendChange := func(change ChatbotChange) {
		pendingMutex.Lock()
		defer pendingMutex.Unlock()

		// Stop existing timer if any
		if debounceTimer != nil {
			debounceTimer.Stop()
		}

		// Store the pending change
		pendingChange = &change

		// Set debounce timer (150ms)
		debounceTimer = time.AfterFunc(150*time.Millisecond, func() {
			pendingMutex.Lock()
			if pendingChange != nil {
				changes <- *pendingChange
				pendingChange = nil
			}
			pendingMutex.Unlock()
		})
	}

	// Helper to check if file should be watched
	shouldWatch := func(path string) bool {
		// Filter for .ts and .js files only
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".ts" && ext != ".js" {
			return false
		}

		// Filter out temporary editor files
		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") || // Hidden files
			strings.HasSuffix(base, "~") || // Backup files
			strings.Contains(base, ".swp") || // Vim swap files
			strings.Contains(base, "~") || // Other temp files
			strings.HasPrefix(base, "#") && strings.HasSuffix(base, "#") { // Emacs lock files
			return false
		}

		return true
	}

	// Helper to extract namespace and name from path
	parsePath := func(path string) (namespace, name string, err error) {
		// Get relative path from chatbotsDir
		relPath, err := filepath.Rel(l.chatbotsDir, path)
		if err != nil {
			return "", "", err
		}

		// Expected format: namespace/name/index.ts or namespace/name.ts
		parts := strings.Split(relPath, string(filepath.Separator))
		if len(parts) < 1 {
			return "", "", fmt.Errorf("invalid path: %s", path)
		}

		// Check for namespace/name/index.ts format
		if len(parts) >= 3 && parts[len(parts)-1] == "index.ts" {
			namespace = parts[len(parts)-3]
			name = parts[len(parts)-2]
			return namespace, name, nil
		}

		// Check for namespace/name.ts format
		if len(parts) == 2 {
			namespace = parts[len(parts)-2]
			// Remove extension from name
			name = strings.TrimSuffix(parts[len(parts)-1], filepath.Ext(parts[len(parts)-1]))
			return namespace, name, nil
		}

		return "", "", fmt.Errorf("unexpected path format: %s", path)
	}

	// Recursively add directory and all subdirectories to watcher
	addWatchRecursively := func(rootDir string) error {
		err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if err := addWatchDir(path); err != nil {
					log.Debug().Err(err).Str("dir", path).Msg("Failed to watch directory")
				}
			}
			return nil
		})
		return err
	}

	// Add initial watch on chatbots directory and subdirectories
	if err := addWatchRecursively(l.chatbotsDir); err != nil {
		log.Error().Err(err).Str("dir", l.chatbotsDir).Msg("Failed to watch chatbots directory")
		return
	}

	log.Info().Str("dir", l.chatbotsDir).Msg("Started watching chatbots directory for changes")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Stopping chatbots directory watcher")
			return

		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// If a directory was created, watch it
			if event.Has(fsnotify.Create) {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					_ = addWatchDir(event.Name)
				}
			}

			// Skip if not a file we care about
			if !shouldWatch(event.Name) {
				continue
			}

			// Parse namespace and name from path
			namespace, name, err := parsePath(event.Name)
			if err != nil {
				log.Debug().Err(err).Str("path", event.Name).Msg("Failed to parse chatbot path")
				continue
			}

			// Determine change type
			changeType := "modified"
			if event.Has(fsnotify.Create) {
				changeType = "created"
			} else if event.Has(fsnotify.Remove) {
				changeType = "deleted"
			}

			// Send change with debouncing
			change := ChatbotChange{
				Type:      changeType,
				Namespace: namespace,
				Name:      name,
				Path:      event.Name,
			}
			sendChange(change)

			log.Debug().
				Str("type", changeType).
				Str("namespace", namespace).
				Str("name", name).
				Str("path", event.Name).
				Msg("Chatbot file change detected")

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Error().Err(err).Msg("File watcher error")
		}
	}
}

// ChatbotChange represents a change to a chatbot file
type ChatbotChange struct {
	Type      string // "created", "modified", "deleted"
	Namespace string
	Name      string
	Path      string
}

// ParseChatbotFromCode parses a chatbot from code string (for SDK-based syncing)
func (l *Loader) ParseChatbotFromCode(code string, namespace string) (*Chatbot, error) {
	// Parse configuration from annotations
	config := ParseChatbotConfig(code)

	// Parse description from JSDoc
	description := ParseDescription(code)

	// Create chatbot
	chatbot := &Chatbot{
		ID:           uuid.New().String(),
		Namespace:    namespace,
		Description:  description,
		Code:         code,
		OriginalCode: code,
		IsBundled:    false,
		Enabled:      true,
		Source:       "sdk",
		Version:      1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Apply parsed configuration
	chatbot.ApplyConfig(config)

	return chatbot, nil
}
