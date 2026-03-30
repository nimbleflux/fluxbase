package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  string
	}{
		{"empty string", "", "****"},
		{"single char", "a", "****"},
		{"exactly 8 chars", "abcdefgh", "****"},
		{"9 chars", "abcdefghi", "abcd*fghi"},
		{"16 chars", "abcdefghijklmnop", "abcd********mnop"},
		{"long token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0", "eyJh********************************************************wIn0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskToken(tt.token)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{"zero", 0, "0 B"},
		{"1 byte", 1, "1 B"},
		{"1023 bytes", 1023, "1023 B"},
		{"1 KB", 1024, "1.0 KB"},
		{"1.5 KB", 1536, "1.5 KB"},
		{"1 MB", 1048576, "1.0 MB"},
		{"1 GB", 1073741824, "1.0 GB"},
		{"1 TB", 1099511627776, "1.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatBytes(tt.bytes)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name    string
		seconds int64
		want    string
	}{
		{"zero", 0, "0s"},
		{"1 second", 1, "1s"},
		{"59 seconds", 59, "59s"},
		{"1 minute", 60, "1m 0s"},
		{"90 seconds", 90, "1m 30s"},
		{"1 hour", 3600, "1h 0m 0s"},
		{"1 hour 30 min", 5400, "1h 30m 0s"},
		{"complex", 3661, "1h 1m 1s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDuration(tt.seconds)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"shorter than max", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"longer with ellipsis", "hello world", 8, "hello..."},
		{"max 3", "hello", 3, "hel"},
		{"max 2", "hello", 2, "he"},
		{"empty string", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateString(tt.input, tt.maxLen)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStringPtr(t *testing.T) {
	val := "test"
	ptr := StringPtr(val)
	assert.NotNil(t, ptr)
	assert.Equal(t, "test", *ptr)
}

func TestInt64Ptr(t *testing.T) {
	val := int64(42)
	ptr := Int64Ptr(val)
	assert.NotNil(t, ptr)
	assert.Equal(t, int64(42), *ptr)
}

func TestBoolPtr(t *testing.T) {
	val := true
	ptr := BoolPtr(val)
	assert.NotNil(t, ptr)
	assert.True(t, *ptr)
}
