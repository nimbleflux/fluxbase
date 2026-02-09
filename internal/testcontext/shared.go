// Package testcontext provides thread-safe storage for shared test context.
// This package is used to share server/app instances across tests without
// creating import cycles.
package testcontext

import (
	"sync"
)

// Package-level storage for shared server/app instances
// Using string keys to identify different packages
var (
	mu      sync.RWMutex
	servers = make(map[string]interface{})
	apps    = make(map[string]interface{})
)

// SetServer stores a server instance for the given package key.
// This allows tests to access the server instance without creating import cycles.
func SetServer(key string, server interface{}) {
	mu.Lock()
	defer mu.Unlock()
	servers[key] = server
}

// GetServer retrieves a server instance for the given package key.
// Returns nil if no server is stored for the key.
func GetServer(key string) interface{} {
	mu.RLock()
	defer mu.RUnlock()
	return servers[key]
}

// SetApp stores an app instance for the given package key.
// This allows tests to access the app instance without creating import cycles.
func SetApp(key string, app interface{}) {
	mu.Lock()
	defer mu.Unlock()
	apps[key] = app
}

// GetApp retrieves an app instance for the given package key.
// Returns nil if no app is stored for the key.
func GetApp(key string) interface{} {
	mu.RLock()
	defer mu.RUnlock()
	return apps[key]
}

// ClearAll removes all stored servers and apps.
// This should be called during cleanup to prevent memory leaks.
func ClearAll() {
	mu.Lock()
	defer mu.Unlock()
	servers = make(map[string]interface{})
	apps = make(map[string]interface{})
}
