package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	cfg := New()
	assert.Equal(t, Version, cfg.Version)
	assert.NotNil(t, cfg.Profiles)
	assert.Empty(t, cfg.Profiles)
	assert.Equal(t, "table", cfg.Defaults.Output)
}

func TestLoadOrCreate_Nonexistent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg, err := LoadOrCreate(path)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, Version, cfg.Version)
	assert.Empty(t, cfg.Profiles)
}

func TestLoadOrCreate_Existing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	// Write a config file
	content := `version: "1"
current_profile: dev
profiles:
  dev:
    name: dev
    server: http://localhost:8080
    credentials:
      api_key: test-key-123
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	cfg, err := LoadOrCreate(path)
	require.NoError(t, err)
	assert.Equal(t, "dev", cfg.CurrentProfile)
	require.NotNil(t, cfg.Profiles["dev"])
	assert.Equal(t, "http://localhost:8080", cfg.Profiles["dev"].Server)
	assert.Equal(t, "test-key-123", cfg.Profiles["dev"].Credentials.APIKey)
}

func TestLoad_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("{{invalid yaml"), 0o600))

	_, err := Load(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config")
}

func TestLoad_NotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config file not found")
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "config.yaml")

	cfg := New()
	cfg.CurrentProfile = "prod"
	cfg.SetProfile(&Profile{
		Name:   "prod",
		Server: "https://api.example.com",
		Credentials: &Credentials{
			AccessToken:  "access-token-xyz",
			RefreshToken: "refresh-token-xyz",
			ExpiresAt:    1700000000,
			APIKey:       "",
		},
		User: &UserInfo{
			ID:            "user-123",
			Email:         "test@example.com",
			Role:          "admin",
			EmailVerified: true,
		},
		DefaultNamespace: "ns1",
		OutputFormat:     "json",
		DefaultBranch:    "feature-branch",
	})
	cfg.SetProfile(&Profile{
		Name:   "dev",
		Server: "http://localhost:8080",
	})
	cfg.Defaults = Defaults{
		Output:    "json",
		NoHeaders: true,
	}

	// Save
	require.NoError(t, cfg.Save(path))

	// Verify file was created with correct permissions
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	// Load
	loaded, err := Load(path)
	require.NoError(t, err)

	// Verify all fields survived round-trip
	assert.Equal(t, "prod", loaded.CurrentProfile)
	require.NotNil(t, loaded.Profiles["prod"])
	assert.Equal(t, "https://api.example.com", loaded.Profiles["prod"].Server)
	assert.Equal(t, "access-token-xyz", loaded.Profiles["prod"].Credentials.AccessToken)
	assert.Equal(t, "refresh-token-xyz", loaded.Profiles["prod"].Credentials.RefreshToken)
	assert.Equal(t, int64(1700000000), loaded.Profiles["prod"].Credentials.ExpiresAt)
	assert.Equal(t, "user-123", loaded.Profiles["prod"].User.ID)
	assert.Equal(t, "test@example.com", loaded.Profiles["prod"].User.Email)
	assert.Equal(t, "admin", loaded.Profiles["prod"].User.Role)
	assert.True(t, loaded.Profiles["prod"].User.EmailVerified)
	assert.Equal(t, "ns1", loaded.Profiles["prod"].DefaultNamespace)
	assert.Equal(t, "json", loaded.Profiles["prod"].OutputFormat)
	assert.Equal(t, "feature-branch", loaded.Profiles["prod"].DefaultBranch)
	assert.Equal(t, "json", loaded.Defaults.Output)
	assert.True(t, loaded.Defaults.NoHeaders)
}

func TestSave_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "deep", "nested", "config.yaml")

	cfg := New()
	require.NoError(t, cfg.Save(path))

	_, err := os.Stat(path)
	require.NoError(t, err)
}

func TestGetProfile(t *testing.T) {
	cfg := New()
	cfg.SetProfile(&Profile{Name: "dev", Server: "http://localhost:8080"})
	cfg.SetProfile(&Profile{Name: "prod", Server: "https://api.example.com"})

	t.Run("by name", func(t *testing.T) {
		p, err := cfg.GetProfile("dev")
		require.NoError(t, err)
		assert.Equal(t, "http://localhost:8080", p.Server)
	})

	t.Run("current profile", func(t *testing.T) {
		cfg.CurrentProfile = "prod"
		p, err := cfg.GetProfile("")
		require.NoError(t, err)
		assert.Equal(t, "https://api.example.com", p.Server)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := cfg.GetProfile("staging")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("no current profile", func(t *testing.T) {
		cfg.CurrentProfile = ""
		_, err := cfg.GetProfile("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no profile specified")
	})
}

func TestSetProfile(t *testing.T) {
	cfg := New()

	// Add new profile
	cfg.SetProfile(&Profile{Name: "dev", Server: "http://localhost:8080"})
	assert.NotNil(t, cfg.Profiles["dev"])

	// Overwrite existing
	cfg.SetProfile(&Profile{Name: "dev", Server: "http://newhost:9090"})
	assert.Equal(t, "http://newhost:9090", cfg.Profiles["dev"].Server)
}

func TestDeleteProfile(t *testing.T) {
	cfg := New()
	cfg.SetProfile(&Profile{Name: "dev", Server: "http://localhost:8080"})
	cfg.SetProfile(&Profile{Name: "prod", Server: "https://api.example.com"})
	cfg.CurrentProfile = "dev"

	t.Run("delete non-current", func(t *testing.T) {
		require.NoError(t, cfg.DeleteProfile("prod"))
		assert.Nil(t, cfg.Profiles["prod"])
		assert.Equal(t, "dev", cfg.CurrentProfile)
	})

	t.Run("delete current switches to another", func(t *testing.T) {
		cfg.SetProfile(&Profile{Name: "staging", Server: "http://staging:8080"})
		require.NoError(t, cfg.DeleteProfile("dev"))
		assert.NotEqual(t, "dev", cfg.CurrentProfile)
		assert.True(t, cfg.CurrentProfile == "staging") // picks first available
	})

	t.Run("delete nonexistent", func(t *testing.T) {
		err := cfg.DeleteProfile("nonexistent")
		assert.Error(t, err)
	})
}

func TestDeleteProfile_LastProfile(t *testing.T) {
	cfg := New()
	cfg.SetProfile(&Profile{Name: "only", Server: "http://localhost:8080"})
	cfg.CurrentProfile = "only"

	require.NoError(t, cfg.DeleteProfile("only"))
	assert.Empty(t, cfg.CurrentProfile)
}

func TestListProfiles(t *testing.T) {
	cfg := New()
	cfg.SetProfile(&Profile{Name: "dev", Server: "http://localhost:8080"})
	cfg.SetProfile(&Profile{Name: "prod", Server: "https://api.example.com"})

	list := cfg.ListProfiles()
	assert.Len(t, list, 2)
	assert.Contains(t, list, "dev")
	assert.Contains(t, list, "prod")
}

func TestProfile_HasCredentials(t *testing.T) {
	t.Run("nil credentials", func(t *testing.T) {
		p := &Profile{}
		assert.False(t, p.HasCredentials())
	})

	t.Run("empty credentials", func(t *testing.T) {
		p := &Profile{Credentials: &Credentials{}}
		assert.False(t, p.HasCredentials())
	})

	t.Run("with access token", func(t *testing.T) {
		p := &Profile{Credentials: &Credentials{AccessToken: "token"}}
		assert.True(t, p.HasCredentials())
	})

	t.Run("with API key", func(t *testing.T) {
		p := &Profile{Credentials: &Credentials{APIKey: "key"}}
		assert.True(t, p.HasCredentials())
	})
}

func TestProfile_IsTokenExpired(t *testing.T) {
	t.Run("nil credentials", func(t *testing.T) {
		p := &Profile{}
		assert.False(t, p.IsTokenExpired())
	})

	t.Run("zero expires_at", func(t *testing.T) {
		p := &Profile{Credentials: &Credentials{ExpiresAt: 0}}
		assert.False(t, p.IsTokenExpired())
	})

	t.Run("currentUnixTime returns 0, so any positive ExpiresAt is not expired", func(t *testing.T) {
		// Note: currentUnixTime() currently returns 0
		p := &Profile{Credentials: &Credentials{ExpiresAt: 9999999999}}
		assert.False(t, p.IsTokenExpired())
	})
}

func TestDefaultConfigDir(t *testing.T) {
	dir := DefaultConfigDir()
	assert.Contains(t, dir, ".fluxbase")
}

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()
	assert.Contains(t, path, ".fluxbase")
	assert.Contains(t, path, "config.yaml")
}
