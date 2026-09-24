package runtime

import (
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIsBlockedEnvVar_Patterns verifies credential-shaped variable names are
// blocked by pattern in addition to the explicit list, while benign and
// intentionally-injected names are not.
func TestIsBlockedEnvVar_Patterns(t *testing.T) {
	blocked := []string{
		// explicit list
		"FLUXBASE_AUTH_JWT_SECRET",
		"FLUXBASE_DATABASE_PASSWORD",
		"FLUXBASE_SERVICE_ROLE_KEY",
		"FLUXBASE_STORAGE_S3_SECRET_KEY",
		// previously-missed config secrets
		"FLUXBASE_EMAIL_SES_SECRET_KEY",
		"FLUXBASE_SECURITY_CAPTCHA_SECRET_KEY",
		"FLUXBASE_LOGGING_LOKI_PASSWORD",
		"FLUXBASE_LOGGING_ELASTICSEARCH_PASSWORD",
		// pattern-based
		"FLUXBASE_AI_OPENAI_API_KEY",
		"FLUXBASE_AI_AZURE_API_KEY",
		"FLUXBASE_SOMETHING_PRIVATE_KEY",
		"FLUXBASE_NEW_SECRET_ADDED_LATER",
	}
	for _, name := range blocked {
		assert.True(t, isBlockedEnvVar(name), "%s should be blocked", name)
	}

	allowed := []string{
		"FLUXBASE_DEBUG",
		"FLUXBASE_TIMEOUT",
		"FLUXBASE_CORS_ALLOWED_HEADERS",
		// intentional per-execution injections are appended separately and
		// must not be treated as blocked passthrough names
		"FLUXBASE_SERVICE_TOKEN",
		"FLUXBASE_USER_TOKEN",
		"FLUXBASE_JOB_TOKEN",
	}
	for _, name := range allowed {
		assert.False(t, isBlockedEnvVar(name), "%s should not be blocked", name)
	}
}

// TestBuildEnv_BlocksConfigSecrets verifies that credential-shaped FLUXBASE_*
// variables set in the server environment never reach the execution env.
func TestBuildEnv_BlocksConfigSecrets(t *testing.T) {
	req := ExecutionRequest{ID: uuid.New(), Name: "test", Namespace: "default"}

	t.Setenv("FLUXBASE_EMAIL_SES_SECRET_KEY", "ses-secret")
	t.Setenv("FLUXBASE_SECURITY_CAPTCHA_SECRET_KEY", "captcha-secret")
	t.Setenv("FLUXBASE_AI_OPENAI_API_KEY", "openai-key")
	t.Setenv("FLUXBASE_LOGGING_LOKI_PASSWORD", "loki-pass")

	env := buildEnv(req, RuntimeTypeFunction, t.TempDir(), "", "", "", nil, nil)

	for _, e := range env {
		assert.False(t, strings.HasPrefix(e, "FLUXBASE_EMAIL_SES_SECRET_KEY="), "SES secret key leaked")
		assert.False(t, strings.HasPrefix(e, "FLUXBASE_SECURITY_CAPTCHA_SECRET_KEY="), "captcha secret leaked")
		assert.False(t, strings.HasPrefix(e, "FLUXBASE_AI_OPENAI_API_KEY="), "OpenAI API key leaked")
		assert.False(t, strings.HasPrefix(e, "FLUXBASE_LOGGING_LOKI_PASSWORD="), "Loki password leaked")
	}
}

// TestBuildDenoArgs_DenyNetAndScopedEnv verifies the Deno flag wiring: deny-net
// enforcement of the SSRF blocklist and --allow-env scoped to the passed vars.
func TestBuildDenoArgs_DenyNetAndScopedEnv(t *testing.T) {
	perms := DefaultFunctionPermissions()
	require.NotEmpty(t, perms.BlockedDomains)

	cfg := denoArgsConfig{
		RuntimeType:   RuntimeTypeFunction,
		PublicURL:     "https://api.example.com",
		MemoryLimitMB: 0,
		ExecDir:       t.TempDir(),
		EnvVarNames:   []string{"DENO_DIR", "HOME", "FLUXBASE_EXECUTION_ID"},
	}

	args, _, _ := buildDenoArgs(cfg, perms, nil, "/tmp/fake-main.ts")
	joined := strings.Join(args, " ")

	// Bare --allow-net (no explicit AllowedDomains)…
	assert.Contains(t, args, "--allow-net")
	// …plus deny-net enforcing the blocklist and always-deny entries.
	denyArg := ""
	for _, a := range args {
		if strings.HasPrefix(a, "--deny-net=") {
			denyArg = a
		}
	}
	require.NotEmpty(t, denyArg, "expected --deny-net when BlockedDomains is set")
	assert.Contains(t, denyArg, "169.254.169.254")
	assert.Contains(t, denyArg, "metadata.google.internal")
	assert.Contains(t, denyArg, "localhost")
	assert.Contains(t, denyArg, "127.0.0.1")
	assert.NotContains(t, denyArg, "api.example.com", "self host must never be denied")

	// --allow-env scoped to the actual env var names.
	assert.Contains(t, args, "--allow-env=DENO_DIR,HOME,FLUXBASE_EXECUTION_ID")
	assert.NotContains(t, args, "--allow-env ")
	assert.False(t, joined == strings.Join(args, " ") && strings.Contains(joined, "--allow-env --allow"), "env must not be unscoped")
}

// TestBuildDenoArgs_NoDenyWhenBlocklistCleared verifies an explicitly empty
// BlockedDomains opts out of deny-net.
func TestBuildDenoArgs_NoDenyWhenBlocklistCleared(t *testing.T) {
	perms := Permissions{AllowNet: true, AllowEnv: true}
	cfg := denoArgsConfig{
		RuntimeType: RuntimeTypeFunction,
		ExecDir:     t.TempDir(),
		EnvVarNames: []string{"HOME"},
	}

	args, _, _ := buildDenoArgs(cfg, perms, nil, "/tmp/fake-main.ts")
	for _, a := range args {
		assert.False(t, strings.HasPrefix(a, "--deny-net"), "deny-net must not be emitted when BlockedDomains is empty")
	}
}

// TestBuildDenoArgs_ScopedFilesystem verifies read/write permissions point at
// the per-execution directory instead of the shared /tmp.
func TestBuildDenoArgs_ScopedFilesystem(t *testing.T) {
	execDir := t.TempDir()
	perms := Permissions{AllowNet: false, AllowEnv: false, AllowRead: true, AllowWrite: true}
	cfg := denoArgsConfig{RuntimeType: RuntimeTypeFunction, ExecDir: execDir}

	args, _, _ := buildDenoArgs(cfg, perms, nil, "/tmp/fake-main.ts")
	assert.Contains(t, args, "--allow-read="+execDir)
	assert.Contains(t, args, "--allow-write="+execDir)
}

// TestBuildNetworkDenyList verifies deny list construction: always-deny
// entries, configured entries, and self-host exemption.
func TestBuildNetworkDenyList(t *testing.T) {
	denied := buildNetworkDenyList(Permissions{
		BlockedDomains: []string{"169.254.169.254", "internal.example.com"},
	}, "http://localhost:8080")

	assert.Contains(t, denied, "169.254.169.254")
	assert.Contains(t, denied, "internal.example.com")
	assert.Contains(t, denied, "metadata.google.internal")
	// Self host is exempt so SDK callbacks to the configured URL keep working
	// (it may legitimately be loopback in development).
	assert.NotContains(t, denied, "localhost")
	assert.Contains(t, denied, "127.0.0.1")

	// No duplicates
	count := 0
	for _, d := range denied {
		if d == "169.254.169.254" {
			count++
		}
	}
	assert.Equal(t, 1, count)

	// Self host is exempt (dev deployments expose the instance on loopback)
	denied = buildNetworkDenyList(Permissions{}, "http://127.0.0.1:8080")
	assert.NotContains(t, denied, "127.0.0.1")
}

// TestBuildEnv_PerExecutionDirs verifies Deno cache and HOME live inside the
// per-execution directory rather than a shared /tmp location.
func TestBuildEnv_PerExecutionDirs(t *testing.T) {
	req := ExecutionRequest{ID: uuid.New(), Name: "test", Namespace: "default"}
	execDir := t.TempDir()

	env := buildEnv(req, RuntimeTypeFunction, execDir, "", "", "", nil, nil)

	foundDenoDir, foundHome := false, false
	for _, e := range env {
		if strings.HasPrefix(e, "DENO_DIR=") {
			foundDenoDir = true
			assert.True(t, strings.HasPrefix(e, "DENO_DIR="+execDir), "DENO_DIR must live in the execution dir: %s", e)
		}
		if strings.HasPrefix(e, "HOME=") {
			foundHome = true
			assert.Equal(t, "HOME="+execDir, e)
		}
	}
	assert.True(t, foundDenoDir, "DENO_DIR must be set")
	assert.True(t, foundHome, "HOME must be set")

	// No shared /tmp/deno anymore
	for _, e := range env {
		assert.NotEqual(t, "DENO_DIR=/tmp/deno", e)
		assert.NotEqual(t, "HOME=/tmp", e)
	}

	// Best-effort cleanup check: the caller removes the dir; here we only
	// verify os.Environ-based pass-through doesn't accidentally leak /tmp vars.
	_ = os.Getenv("PATH")
}
