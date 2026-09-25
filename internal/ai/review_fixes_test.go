package ai

import (
	"strings"
	"testing"

	"github.com/nimbleflux/fluxbase/internal/ai/integrations"
	"github.com/nimbleflux/fluxbase/internal/auth"
)

// ───────────────────────── A1: WS database-role resolution ─────────────────────────

func TestResolveDatabaseRole(t *testing.T) {
	str := func(s string) interface{} { return s }

	tests := []struct {
		name    string
		rlsRole interface{}
		userRol interface{}
		want    string
	}{
		// Service-key / client-key paths set rls_role directly (already a DB role).
		{"service key role passthrough", str("service_role"), nil, "service_role"},
		{"client key authenticated passthrough", str("authenticated"), nil, "authenticated"},
		{"anon passthrough", str("anon"), nil, "anon"},

		// JWT (OptionalAuth) paths: only user_role is set; rls_role missing.
		{"jwt regular user", nil, str("authenticated"), "authenticated"},
		{"jwt instance admin maps to service_role", nil, str("instance_admin"), "service_role"},
		{"jwt app admin maps to authenticated", nil, str("admin"), "authenticated"},
		{"jwt tenant admin maps to authenticated", nil, str("tenant_admin"), "authenticated"},
		{"jwt tenant service", nil, str("tenant_service"), "tenant_service"},
		{"jwt anon role", nil, str("anon"), "anon"},

		// No identity at all.
		{"anonymous connection", nil, nil, "anon"},
		{"empty strings ignored", str(""), str(""), "anon"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveDatabaseRole(tt.rlsRole, tt.userRol); got != tt.want {
				t.Errorf("ResolveDatabaseRole(%v, %v) = %q, want %q", tt.rlsRole, tt.userRol, got, tt.want)
			}
		})
	}
}

// ───────────────────────── A4: impersonation gate + claims swap ─────────────────────────

func TestIsInstanceAdminPrincipal(t *testing.T) {
	if isInstanceAdminPrincipal(nil) {
		t.Error("nil context must not be an instance admin")
	}

	// Dashboard instance_admin JWT (role resolves to service_role at DB level
	// but claims carry the original role) — the A4 gate fix.
	claimsAdmin := &ChatContext{
		Role:   "service_role",
		Claims: &auth.TokenClaims{Role: "instance_admin"},
	}
	if !isInstanceAdminPrincipal(claimsAdmin) {
		t.Error("claims.Role=instance_admin should pass the impersonation gate")
	}

	// Service-key path where rls_role carried the app role.
	if !isInstanceAdminPrincipal(&ChatContext{Role: "instance_admin"}) {
		t.Error("Role=instance_admin should pass the impersonation gate")
	}

	// Regular users must not pass.
	if isInstanceAdminPrincipal(&ChatContext{Role: "authenticated"}) {
		t.Error("regular authenticated user must not pass the impersonation gate")
	}
	if isInstanceAdminPrincipal(&ChatContext{Role: "service_role", Claims: &auth.TokenClaims{Role: "authenticated"}}) {
		t.Error("service_role connection without instance_admin claims must not pass")
	}
}

// ───────────────────────── A2: conversation ownership ─────────────────────────

func TestConversationOwnedBy(t *testing.T) {
	userA := "11111111-1111-1111-1111-111111111111"
	userB := "22222222-2222-2222-2222-222222222222"

	// Same owner.
	if !conversationOwnedBy(&ConversationState{UserID: &userA}, &userA) {
		t.Error("owner should be able to resume their conversation")
	}
	// Anonymous chatbot + anonymous caller (both empty).
	if !conversationOwnedBy(&ConversationState{}, nil) {
		t.Error("both sides anonymous should be allowed")
	}
	// Hijack attempts must be denied.
	if conversationOwnedBy(&ConversationState{UserID: &userA}, &userB) {
		t.Error("another user's conversation must be rejected")
	}
	if conversationOwnedBy(&ConversationState{UserID: &userA}, nil) {
		t.Error("anonymous caller must not resume an owned conversation")
	}
	if conversationOwnedBy(&ConversationState{}, &userA) {
		t.Error("authenticated caller must not adopt an anonymous conversation")
	}
	if conversationOwnedBy(nil, &userA) {
		t.Error("nil state must be rejected")
	}
}

// ───────────────────────── A10: anonymous rate-limit buckets ─────────────────────────

func TestRateLimitIdentifier(t *testing.T) {
	uid := "user-1"
	if got := rateLimitIdentifier(&ChatContext{UserID: &uid, IPAddress: "1.2.3.4:5"}); got != "user-1" {
		t.Errorf("authenticated identifier = %q, want user id", got)
	}
	ipCtx := &ChatContext{IPAddress: "10.0.0.7:51234"}
	if got := rateLimitIdentifier(ipCtx); got != "ip:10.0.0.7:51234" {
		t.Errorf("anonymous identifier = %q, want per-IP bucket", got)
	}
	if got := rateLimitIdentifier(&ChatContext{}); got != "anonymous" {
		t.Errorf("no identity = %q, want anonymous", got)
	}
	if got := rateLimitIdentifier(nil); got != "anonymous" {
		t.Errorf("nil ctx = %q, want anonymous", got)
	}
}

// ───────────────────────── B1: user-isolation predicate ─────────────────────────

func TestUserIsolationCondition(t *testing.T) {
	// Include-global (legacy default): caller's docs OR no user_id.
	incl := userIsolationCondition("d.metadata", 5, true)
	for _, want := range []string{
		`d.metadata->>'user_id' = $5`,
		`d.metadata->>'user_id' IS NULL`,
		`NOT (d.metadata ? 'user_id')`,
	} {
		if !strings.Contains(incl, want) {
			t.Errorf("include-global predicate missing %q; got %s", want, incl)
		}
	}

	// Strict (filtered links / deletes): only the caller's docs — global
	// docs must NOT match.
	strict := userIsolationCondition("metadata", 3, false)
	if !strings.Contains(strict, `metadata->>'user_id' = $3`) {
		t.Errorf("strict predicate missing equality clause; got %s", strict)
	}
	if strings.Contains(strict, "IS NULL") || strings.Contains(strict, "? 'user_id'") {
		t.Errorf("strict predicate must not leak global docs; got %s", strict)
	}
}

// ───────────────────────── B2: link access-level default ─────────────────────────

func TestNormalizeAccessLevel(t *testing.T) {
	if got := normalizeAccessLevel(""); got != string(AccessLevelFiltered) {
		t.Errorf("empty access level = %q, want filtered default", got)
	}
	if got := normalizeAccessLevel("bogus"); got != string(AccessLevelFiltered) {
		t.Errorf("invalid access level = %q, want filtered default", got)
	}
	if got := normalizeAccessLevel("full"); got != string(AccessLevelFull) {
		t.Errorf("explicit full = %q, want full opt-in preserved", got)
	}
	if got := normalizeAccessLevel("tiered"); got != string(AccessLevelTiered) {
		t.Errorf("explicit tiered = %q, want tiered", got)
	}
}

// ───────────────────────── B3: export user_id metadata ─────────────────────────

func TestExportMetadataUserID(t *testing.T) {
	owner := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	columns := []string{"id", "user_id", "name"}

	// Zero-config convention: source column literally named user_id.
	if got, ok := exportMetadataUserID(columns, "", &owner); !ok || got != owner {
		t.Errorf("default user_id column: got (%q, %v), want (%q, true)", got, ok, owner)
	}
	// Explicit override column.
	if _, ok := exportMetadataUserID(columns, "owner_id", &owner); ok {
		t.Error("override column not present in table must not stamp metadata")
	}
	if got, ok := exportMetadataUserID([]string{"id", "owner_id"}, "owner_id", &owner); !ok || got != owner {
		t.Errorf("override column mapping failed: got (%q, %v)", got, ok)
	}
	// No identity → no stamp (document stays global).
	if _, ok := exportMetadataUserID(columns, "", nil); ok {
		t.Error("export without owner identity must not stamp user_id metadata")
	}
	empty := ""
	if _, ok := exportMetadataUserID(columns, "", &empty); ok {
		t.Error("empty owner identity must not stamp user_id metadata")
	}
	// Table without the column → no stamp.
	if _, ok := exportMetadataUserID([]string{"id", "name"}, "", &owner); ok {
		t.Error("table without per-user column must not stamp user_id metadata")
	}
}

// ───────────────────────── A6: provider config secret encryption ─────────────────────────

func TestIsSecretConfigKey(t *testing.T) {
	for _, k := range []string{"api_key", "ApiKey", "client_secret", "refresh_token", "TOKEN"} {
		if !isSecretConfigKey(k) {
			t.Errorf("isSecretConfigKey(%q) = false, want true", k)
		}
	}
	for _, k := range []string{"model", "base_url", "endpoint", "temperature"} {
		if isSecretConfigKey(k) {
			t.Errorf("isSecretConfigKey(%q) = true, want false", k)
		}
	}
}

func TestEncryptDecryptProviderConfig(t *testing.T) {
	s := &Storage{}
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	s.encryptionKey = key

	cfg := map[string]string{
		"api_key":  "sk-super-secret",
		"model":    "gpt-4o",
		"base_url": "https://api.example.com",
	}

	// Encrypt: caller's map must not be mutated, secrets become "enc:"-tagged.
	stored, err := s.encryptProviderConfig(cfg)
	if err != nil {
		t.Fatalf("encryptProviderConfig failed: %v", err)
	}
	if cfg["api_key"] != "sk-super-secret" {
		t.Error("encryptProviderConfig must not mutate the caller's map")
	}
	if !strings.HasPrefix(stored["api_key"], "enc:") {
		t.Errorf("api_key not encrypted at rest: %q", stored["api_key"])
	}
	if stored["model"] != "gpt-4o" || stored["base_url"] != "https://api.example.com" {
		t.Error("non-secret config fields must be stored verbatim")
	}

	// Decrypt: round-trip restores plaintext.
	loaded := s.decryptProviderConfig(stored)
	if loaded["api_key"] != "sk-super-secret" {
		t.Errorf("round-trip api_key = %q, want original plaintext", loaded["api_key"])
	}
	if loaded["model"] != "gpt-4o" {
		t.Error("round-trip must not touch non-secret fields")
	}

	// Legacy plaintext rows pass through (lazy migration).
	legacy := s.decryptProviderConfig(map[string]string{"api_key": "sk-legacy-plaintext"})
	if legacy["api_key"] != "sk-legacy-plaintext" {
		t.Errorf("legacy plaintext = %q, want passthrough", legacy["api_key"])
	}

	// No key configured → plaintext passthrough (encryption off).
	plain := &Storage{}
	out, err := plain.encryptProviderConfig(map[string]string{"api_key": "sk-open"})
	if err != nil {
		t.Fatalf("encrypt without key failed: %v", err)
	}
	if out["api_key"] != "sk-open" {
		t.Errorf("encryption-off write = %q, want plaintext", out["api_key"])
	}

	// Misconfigured key length fails the write loudly.
	bad := &Storage{encryptionKey: []byte("short")}
	if _, err := bad.encryptProviderConfig(map[string]string{"api_key": "sk-x"}); err == nil {
		t.Error("short master key must fail the write, not store plaintext")
	}

	// Sanity: the integrations helper used by the storage layer behaves.
	enc, err := integrations.EncryptSecret("v", key)
	if err != nil || !integrations.IsEncrypted(enc) {
		t.Fatalf("EncryptSecret/IsEncrypted mismatch: %q, %v", enc, err)
	}
}

// ───────────────────────── A7: processing attempts + transient errors ─────────────────────────

func TestProcessingAttemptsFromMetadata(t *testing.T) {
	if got := processingAttemptsFromMetadata(nil); got != 0 {
		t.Errorf("nil metadata = %d, want 0", got)
	}
	if got := processingAttemptsFromMetadata([]byte(`{"processing_attempts":2}`)); got != 2 {
		t.Errorf("attempts = %d, want 2", got)
	}
	if got := processingAttemptsFromMetadata([]byte(`{"other":"x"}`)); got != 0 {
		t.Errorf("missing key = %d, want 0", got)
	}
	if got := processingAttemptsFromMetadata([]byte(`not-json`)); got != 0 {
		t.Errorf("bad json = %d, want 0", got)
	}
}

func TestIsTransientEmbeddingError(t *testing.T) {
	if isTransientEmbeddingError(nil) {
		t.Error("nil error must not be transient")
	}
	for _, msg := range []string{
		"openai embedding returned status 429: slow down",
		"openai embedding returned status 503: unavailable",
		"embedding rate limit exceeded",
		"failed to send embedding request: connection refused",
		"provider overloaded",
	} {
		if !isTransientEmbeddingError(errString(msg)) {
			t.Errorf("%q should be transient", msg)
		}
	}
	for _, msg := range []string{
		"invalid api key",
		"model not found",
		"failed to marshal embedding request",
	} {
		if isTransientEmbeddingError(errString(msg)) {
			t.Errorf("%q should not be transient", msg)
		}
	}
}

type errString string

func (e errString) Error() string { return string(e) }
