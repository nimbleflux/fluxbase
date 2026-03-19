package keys

import (
	"crypto/rand"
	"errors"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

const keyBodyLength = 32

const alphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var (
	ErrInvalidKeyType = errors.New("invalid key type")
	ErrKeyHashFailed  = errors.New("failed to hash key")
)

func GenerateKey(keyType string) (fullKey, keyHash, keyPrefix string, err error) {
	prefix, err := getPrefixForKeyType(keyType)
	if err != nil {
		return "", "", "", err
	}

	keyBody, err := generateAlphanumericString(keyBodyLength)
	if err != nil {
		return "", "", "", err
	}

	fullKey = prefix + keyBody
	keyPrefix = extractKeyPrefix(fullKey, len(prefix))
	keyHash, err = HashKey(fullKey)
	if err != nil {
		return "", "", "", err
	}

	return fullKey, keyHash, keyPrefix, nil
}

func HashKey(key string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
	if err != nil {
		return "", ErrKeyHashFailed
	}
	return string(hash), nil
}

func VerifyKey(key, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(key))
	return err == nil
}

func ExtractPrefix(key string) string {
	prefixes := []string{
		KeyPrefixAnon,
		KeyPrefixPublishable,
		KeyPrefixTenantService,
		KeyPrefixGlobalService,
	}

	for _, prefix := range prefixes {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			return prefix
		}
	}

	return ""
}

func ParseKeyType(prefix string) string {
	switch prefix {
	case KeyPrefixAnon:
		return KeyTypeAnon
	case KeyPrefixPublishable:
		return KeyTypePublishable
	case KeyPrefixTenantService:
		return KeyTypeTenantService
	case KeyPrefixGlobalService:
		return KeyTypeGlobalService
	default:
		return ""
	}
}

func getPrefixForKeyType(keyType string) (string, error) {
	switch keyType {
	case KeyTypeAnon:
		return KeyPrefixAnon, nil
	case KeyTypePublishable:
		return KeyPrefixPublishable, nil
	case KeyTypeTenantService:
		return KeyPrefixTenantService, nil
	case KeyTypeGlobalService:
		return KeyPrefixGlobalService, nil
	default:
		return "", ErrInvalidKeyType
	}
}

func generateAlphanumericString(length int) (string, error) {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphanumeric))))
		if err != nil {
			return "", err
		}
		result[i] = alphanumeric[num.Int64()]
	}
	return string(result), nil
}

func extractKeyPrefix(fullKey string, prefixLen int) string {
	totalLen := prefixLen + 8
	if len(fullKey) < totalLen {
		return fullKey
	}
	return fullKey[:totalLen]
}
