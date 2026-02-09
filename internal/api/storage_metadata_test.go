package api

import (
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

// =============================================================================
// parseMetadata Tests
// =============================================================================

func TestParseMetadata(t *testing.T) {
	tests := []struct {
		name             string
		postArgs         map[string]string
		expectedMetadata map[string]string
	}{
		{
			name:             "no metadata fields",
			postArgs:         map[string]string{},
			expectedMetadata: map[string]string{},
		},
		{
			name: "single metadata field",
			postArgs: map[string]string{
				"metadata_key1": "value1",
			},
			expectedMetadata: map[string]string{
				"key1": "value1",
			},
		},
		{
			name: "multiple metadata fields",
			postArgs: map[string]string{
				"metadata_title":       "My Document",
				"metadata_description": "A test document",
				"metadata_category":    "test",
			},
			expectedMetadata: map[string]string{
				"title":       "My Document",
				"description": "A test document",
				"category":    "test",
			},
		},
		{
			name: "mixed metadata and non-metadata fields",
			postArgs: map[string]string{
				"metadata_author":    "John Doe",
				"file_name":          "test.txt",
				"metadata_tags":      "tag1,tag2",
				"upload_timestamp":   "2025-01-01",
				"metadata_is_public": "true",
			},
			expectedMetadata: map[string]string{
				"author":    "John Doe",
				"tags":      "tag1,tag2",
				"is_public": "true",
			},
		},
		{
			name: "metadata with special characters in values",
			postArgs: map[string]string{
				"metadata_key with spaces":     "value with spaces",
				"metadata_key-with-dashes":     "value-with-dashes",
				"metadata_key_with_underscore": "value_with_underscore",
			},
			expectedMetadata: map[string]string{
				"key with spaces":     "value with spaces",
				"key-with-dashes":     "value-with-dashes",
				"key_with_underscore": "value_with_underscore",
			},
		},
		{
			name: "empty metadata values",
			postArgs: map[string]string{
				"metadata_key1": "",
				"metadata_key2": "non-empty",
			},
			expectedMetadata: map[string]string{
				"key1": "",
				"key2": "non-empty",
			},
		},
		{
			name: "metadata keys with numbers",
			postArgs: map[string]string{
				"metadata_field1": "value1",
				"metadata_field2": "value2",
				"metadata_123":    "numeric",
			},
			expectedMetadata: map[string]string{
				"field1": "value1",
				"field2": "value2",
				"123":    "numeric",
			},
		},
		{
			name: "non-metadata fields are ignored",
			postArgs: map[string]string{
				"file_name":    "document.pdf",
				"content_type": "application/pdf",
				"file_size":    "1024",
			},
			expectedMetadata: map[string]string{},
		},
		{
			name: "metadata_ prefix without key name",
			postArgs: map[string]string{
				"metadata_": "empty key",
			},
			expectedMetadata: map[string]string{
				"": "empty key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock Fiber app and context
			app := fiber.New()
			c := app.AcquireCtx(&fasthttp.RequestCtx{})

			// Set up the post arguments
			for key, value := range tt.postArgs {
				c.Request().PostArgs().Set(key, value)
			}

			// Call parseMetadata
			result := parseMetadata(c)

			// Verify the result
			assert.Equal(t, tt.expectedMetadata, result)
		})
	}
}
