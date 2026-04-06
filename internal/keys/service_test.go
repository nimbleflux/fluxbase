package keys

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestGenerateKey(t *testing.T) {
	t.Run("returns correct prefix for each key type", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			keyType        string
			expectedPrefix string
		}{
			{"anon key", KeyTypeAnon, KeyPrefixAnon},
			{"publishable key", KeyTypePublishable, KeyPrefixPublishable},
			{"tenant service key", KeyTypeTenantService, KeyPrefixTenantService},
			{"global service key", KeyTypeGlobalService, KeyPrefixGlobalService},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				fullKey, _, keyPrefix, err := GenerateKey(tt.keyType)

				require.NoError(t, err)
				assert.True(t, strings.HasPrefix(fullKey, tt.expectedPrefix),
					"full key should start with %s, got %s", tt.expectedPrefix, fullKey)
				assert.Equal(t, tt.expectedPrefix[:len(tt.expectedPrefix)-1], keyPrefix[:len(tt.expectedPrefix)-1],
					"key prefix should contain the prefix base")
			})
		}
	})

	t.Run("returns non-empty hash", func(t *testing.T) {
		t.Parallel()

		for _, keyType := range []string{KeyTypeAnon, KeyTypePublishable, KeyTypeTenantService, KeyTypeGlobalService} {
			_, keyHash, _, err := GenerateKey(keyType)

			require.NoError(t, err)
			assert.NotEmpty(t, keyHash, "hash should not be empty for key type %s", keyType)
			assert.True(t, strings.HasPrefix(keyHash, "$2a$"),
				"hash should be a bcrypt hash for key type %s", keyType)
		}
	})

	t.Run("returns non-empty prefix", func(t *testing.T) {
		t.Parallel()

		for _, keyType := range []string{KeyTypeAnon, KeyTypePublishable, KeyTypeTenantService, KeyTypeGlobalService} {
			_, _, keyPrefix, err := GenerateKey(keyType)

			require.NoError(t, err)
			assert.NotEmpty(t, keyPrefix, "prefix should not be empty for key type %s", keyType)
		}
	})

	t.Run("full key format is prefix plus body", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name    string
			keyType string
			prefix  string
		}{
			{"anon", KeyTypeAnon, KeyPrefixAnon},
			{"publishable", KeyTypePublishable, KeyPrefixPublishable},
			{"tenant_service", KeyTypeTenantService, KeyPrefixTenantService},
			{"global_service", KeyTypeGlobalService, KeyPrefixGlobalService},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				fullKey, _, _, err := GenerateKey(tt.keyType)

				require.NoError(t, err)
				assert.True(t, strings.HasPrefix(fullKey, tt.prefix),
					"key must start with %s", tt.prefix)

				body := strings.TrimPrefix(fullKey, tt.prefix)
				assert.Equal(t, keyBodyLength, len(body),
					"key body should be %d characters, got %d", keyBodyLength, len(body))
				for i, c := range body {
					assert.True(t, strings.ContainsRune(alphanumeric, c),
						"body character at index %d (%c) is not alphanumeric", i, c)
				}
			})
		}
	})

	t.Run("generated keys are unique", func(t *testing.T) {
		t.Parallel()

		for _, keyType := range []string{KeyTypeAnon, KeyTypePublishable, KeyTypeTenantService, KeyTypeGlobalService} {
			key1, hash1, _, err := GenerateKey(keyType)
			require.NoError(t, err)

			key2, hash2, _, err := GenerateKey(keyType)
			require.NoError(t, err)

			assert.NotEqual(t, key1, key2, "two generated keys should differ for type %s", keyType)
			assert.NotEqual(t, hash1, hash2, "two generated hashes should differ for type %s", keyType)
		}
	})

	t.Run("invalid key type returns ErrInvalidKeyType", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name    string
			keyType string
		}{
			{"empty string", ""},
			{"unknown type", "unknown"},
			{"singular service", "service"},
			{"uppercase", "ANON"},
			{"with spaces", " anon "},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				fullKey, keyHash, keyPrefix, err := GenerateKey(tt.keyType)

				assert.ErrorIs(t, err, ErrInvalidKeyType)
				assert.Empty(t, fullKey)
				assert.Empty(t, keyHash)
				assert.Empty(t, keyPrefix)
			})
		}
	})

	t.Run("generated hash verifies against full key", func(t *testing.T) {
		t.Parallel()

		for _, keyType := range []string{KeyTypeAnon, KeyTypePublishable, KeyTypeTenantService, KeyTypeGlobalService} {
			fullKey, keyHash, _, err := GenerateKey(keyType)

			require.NoError(t, err)
			assert.True(t, VerifyKey(fullKey, keyHash),
				"generated key should verify against its hash for type %s", keyType)
		}
	})
}

func TestHashKey(t *testing.T) {
	t.Run("returns a bcrypt hash", func(t *testing.T) {
		t.Parallel()

		key := "fb_pk_sometestkeyvalue1234567890abcd"
		hash, err := HashKey(key)

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, key, hash)
		assert.True(t, strings.HasPrefix(hash, "$2a$"), "hash should start with bcrypt identifier $2a$")
	})

	t.Run("different inputs produce different hashes", func(t *testing.T) {
		t.Parallel()

		hash1, err1 := HashKey("first_key_value")
		hash2, err2 := HashKey("second_key_value")

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2, "different inputs should produce different hashes")
	})

	t.Run("same input produces different hashes due to salt", func(t *testing.T) {
		t.Parallel()

		input := "same_key_value"

		hash1, err1 := HashKey(input)
		hash2, err2 := HashKey(input)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2, "same input should produce different hashes due to bcrypt salting")
	})

	t.Run("empty string can be hashed", func(t *testing.T) {
		t.Parallel()

		hash, err := HashKey("")

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.True(t, strings.HasPrefix(hash, "$2a$"))
	})

	t.Run("hash is valid bcrypt format", func(t *testing.T) {
		t.Parallel()

		key := "fb_gsk_testkey1234567890"
		hash, err := HashKey(key)

		require.NoError(t, err)
		// bcrypt.CompareHashAndPassword will return nil on match
		err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(key))
		assert.NoError(t, err, "hash should be a valid bcrypt hash that matches the key")
	})
}

func TestVerifyKey(t *testing.T) {
	t.Run("correct key matches hash", func(t *testing.T) {
		t.Parallel()

		key := "fb_tsk_correctkey1234567890abcdef"
		hash, err := HashKey(key)
		require.NoError(t, err)

		assert.True(t, VerifyKey(key, hash), "correct key should match its hash")
	})

	t.Run("wrong key does not match", func(t *testing.T) {
		t.Parallel()

		key := "fb_tsk_correctkey1234567890abcdef"
		wrongKey := "fb_tsk_wrongkey1234567890abcdef"
		hash, err := HashKey(key)
		require.NoError(t, err)

		assert.False(t, VerifyKey(wrongKey, hash), "wrong key should not match the hash")
	})

	t.Run("empty key against empty hash", func(t *testing.T) {
		t.Parallel()

		hash, err := HashKey("")
		require.NoError(t, err)

		assert.True(t, VerifyKey("", hash), "empty key should match hash of empty key")
	})

	t.Run("non-empty key against hash of empty", func(t *testing.T) {
		t.Parallel()

		hash, err := HashKey("")
		require.NoError(t, err)

		assert.False(t, VerifyKey("fb_pk_somekey", hash), "non-empty key should not match hash of empty string")
	})

	t.Run("empty key against hash of non-empty", func(t *testing.T) {
		t.Parallel()

		hash, err := HashKey("fb_pk_somekey")
		require.NoError(t, err)

		assert.False(t, VerifyKey("", hash), "empty key should not match hash of non-empty key")
	})

	t.Run("invalid hash returns false", func(t *testing.T) {
		t.Parallel()

		assert.False(t, VerifyKey("some_key", "not-a-hash"), "invalid hash format should return false")
		assert.False(t, VerifyKey("some_key", ""), "empty hash should return false")
	})

	t.Run("round-trip with GenerateKey", func(t *testing.T) {
		t.Parallel()

		fullKey, keyHash, _, err := GenerateKey(KeyTypePublishable)
		require.NoError(t, err)

		assert.True(t, VerifyKey(fullKey, keyHash), "key from GenerateKey should verify against its hash")
	})
}

func TestExtractPrefix(t *testing.T) {
	t.Run("each key type prefix is correctly extracted", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			key            string
			expectedPrefix string
		}{
			{"anon key", KeyPrefixAnon + "abcdefghijklmnopqrstuvwxyz123456", KeyPrefixAnon},
			{"publishable key", KeyPrefixPublishable + "abcdefghijklmnopqrstuvwxyz123456", KeyPrefixPublishable},
			{"tenant service key", KeyPrefixTenantService + "abcdefghijklmnopqrstuvwxyz123456", KeyPrefixTenantService},
			{"global service key", KeyPrefixGlobalService + "abcdefghijklmnopqrstuvwxyz123456", KeyPrefixGlobalService},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				prefix := ExtractPrefix(tt.key)
				assert.Equal(t, tt.expectedPrefix, prefix)
			})
		}
	})

	t.Run("unknown prefix returns empty string", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name string
			key  string
		}{
			{"unknown prefix", "unknown_something12345678"},
			{"partial prefix match", "fb_"},
			{"random string", "some_random_key_value"},
			{"similar but wrong", "fb_pk2_something12345678"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				prefix := ExtractPrefix(tt.key)
				assert.Empty(t, prefix, "unknown prefix should return empty string for key: %s", tt.key)
			})
		}
	})

	t.Run("empty string returns empty string", func(t *testing.T) {
		t.Parallel()

		prefix := ExtractPrefix("")
		assert.Empty(t, prefix)
	})

	t.Run("extracted prefix from generated key", func(t *testing.T) {
		t.Parallel()

		for _, keyType := range []string{KeyTypeAnon, KeyTypePublishable, KeyTypeTenantService, KeyTypeGlobalService} {
			fullKey, _, _, err := GenerateKey(keyType)
			require.NoError(t, err)

			prefix := ExtractPrefix(fullKey)
			assert.NotEmpty(t, prefix, "ExtractPrefix should return non-empty for generated key of type %s", keyType)
			assert.True(t, strings.HasPrefix(fullKey, prefix),
				"full key should start with the extracted prefix for type %s", keyType)
		}
	})
}

func TestParseKeyType(t *testing.T) {
	t.Run("each prefix maps to correct key type", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name         string
			prefix       string
			expectedType string
		}{
			{"anon prefix", KeyPrefixAnon, KeyTypeAnon},
			{"publishable prefix", KeyPrefixPublishable, KeyTypePublishable},
			{"tenant service prefix", KeyPrefixTenantService, KeyTypeTenantService},
			{"global service prefix", KeyPrefixGlobalService, KeyTypeGlobalService},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				keyType := ParseKeyType(tt.prefix)
				assert.Equal(t, tt.expectedType, keyType)
			})
		}
	})

	t.Run("unknown prefix returns empty string", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name   string
			prefix string
		}{
			{"empty string", ""},
			{"unknown prefix", "fb_unknown_"},
			{"partial prefix", "fb_pk"},
			{"random string", "not_a_prefix"},
			{"case wrong", "FB_PK_"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				keyType := ParseKeyType(tt.prefix)
				assert.Empty(t, keyType, "unknown prefix %q should return empty string", tt.prefix)
			})
		}
	})
}

func TestExtractKeyPrefix(t *testing.T) {
	t.Run("extracts prefix with 8 extra characters", func(t *testing.T) {
		t.Parallel()

		// extractKeyPrefix takes the first (prefixLen + 8) characters
		tests := []struct {
			name      string
			fullKey   string
			prefixLen int
			expected  string
		}{
			{
				name:      "anon key",
				fullKey:   KeyPrefixAnon + "abcdefghijklmnopqrstuvwxyz123456",
				prefixLen: len(KeyPrefixAnon),
				expected:  (KeyPrefixAnon + "abcdefghijklmnopqrstuvwxyz123456")[:len(KeyPrefixAnon)+8],
			},
			{
				name:      "publishable key",
				fullKey:   KeyPrefixPublishable + "abcdefghijklmnopqrstuvwxyz123456",
				prefixLen: len(KeyPrefixPublishable),
				expected:  (KeyPrefixPublishable + "abcdefghijklmnopqrstuvwxyz123456")[:len(KeyPrefixPublishable)+8],
			},
			{
				name:      "tenant service key",
				fullKey:   KeyPrefixTenantService + "abcdefghijklmnopqrstuvwxyz123456",
				prefixLen: len(KeyPrefixTenantService),
				expected:  (KeyPrefixTenantService + "abcdefghijklmnopqrstuvwxyz123456")[:len(KeyPrefixTenantService)+8],
			},
			{
				name:      "global service key",
				fullKey:   KeyPrefixGlobalService + "abcdefghijklmnopqrstuvwxyz123456",
				prefixLen: len(KeyPrefixGlobalService),
				expected:  (KeyPrefixGlobalService + "abcdefghijklmnopqrstuvwxyz123456")[:len(KeyPrefixGlobalService)+8],
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				result := extractKeyPrefix(tt.fullKey, tt.prefixLen)
				assert.Equal(t, tt.expected, result)
				assert.Equal(t, tt.prefixLen+8, len(result),
					"extracted prefix should be prefixLen + 8 characters")
			})
		}
	})

	t.Run("short key returns full key", func(t *testing.T) {
		t.Parallel()

		shortKey := "fb_"
		result := extractKeyPrefix(shortKey, 3)
		assert.Equal(t, shortKey, result, "key shorter than prefixLen+8 should be returned as-is")
	})

	t.Run("exact length key returns full key", func(t *testing.T) {
		t.Parallel()

		exactLen := len(KeyPrefixPublishable) + 8
		key := KeyPrefixPublishable + "abcdefgh"
		assert.Equal(t, exactLen, len(key))

		result := extractKeyPrefix(key, len(KeyPrefixPublishable))
		assert.Equal(t, key, result)
	})

	t.Run("via GenerateKey returns correct format", func(t *testing.T) {
		t.Parallel()

		for _, keyType := range []string{KeyTypeAnon, KeyTypePublishable, KeyTypeTenantService, KeyTypeGlobalService} {
			prefix, err := getPrefixForKeyType(keyType)
			require.NoError(t, err)

			_, _, keyPrefix, err := GenerateKey(keyType)
			require.NoError(t, err)

			assert.Equal(t, len(prefix)+8, len(keyPrefix),
				"returned prefix from GenerateKey should be prefixLen+8 for type %s", keyType)
		}
	})
}
