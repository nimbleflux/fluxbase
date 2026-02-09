package api

import (
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestDetectContentType_Utils(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		// Image types
		{"jpg extension", "photo.jpg", "image/jpeg"},
		{"jpeg extension", "photo.jpeg", "image/jpeg"},
		{"png extension", "image.png", "image/png"},
		{"gif extension", "animation.gif", "image/gif"},

		// Document types
		{"pdf extension", "document.pdf", "application/pdf"},
		{"txt extension", "readme.txt", "text/plain"},
		{"html extension", "page.html", "text/html"},

		// Data formats
		{"json extension", "data.json", "application/json"},
		{"xml extension", "config.xml", "application/xml"},

		// Archive types
		{"zip extension", "archive.zip", "application/zip"},

		// Media types
		{"mp4 extension", "video.mp4", "video/mp4"},
		{"mp3 extension", "audio.mp3", "audio/mpeg"},

		// Case insensitivity
		{"uppercase JPG", "photo.JPG", "image/jpeg"},
		{"mixed case PnG", "image.PnG", "image/png"},

		// Unknown extensions
		{"unknown extension", "file.unknown", "application/octet-stream"},
		{"no extension", "filename", "application/octet-stream"},
		{"empty filename", "", "application/octet-stream"},
		{"dot only", ".", "application/octet-stream"},

		// Multiple dots
		{"multiple dots", "archive.tar.gz", "application/octet-stream"},
		{"multiple dots with known ext", "photo.backup.jpg", "image/jpeg"},

		// Hidden files
		{"hidden file with extension", ".gitignore.txt", "text/plain"},
		{"hidden file without extension", ".gitignore", "application/octet-stream"},

		// Path-like filenames
		{"path with extension", "path/to/file.pdf", "application/pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectContentType(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectContentType_AllKnownTypes(t *testing.T) {
	// Verify all known extensions are mapped correctly
	knownTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".pdf":  "application/pdf",
		".txt":  "text/plain",
		".html": "text/html",
		".json": "application/json",
		".xml":  "application/xml",
		".zip":  "application/zip",
		".mp4":  "video/mp4",
		".mp3":  "audio/mpeg",
	}

	for ext, expectedType := range knownTypes {
		t.Run(ext, func(t *testing.T) {
			result := detectContentType("file" + ext)
			assert.Equal(t, expectedType, result, "Expected %s for extension %s", expectedType, ext)
		})
	}
}

func TestGetUserID_Utils(t *testing.T) {
	app := fiber.New()

	tests := []struct {
		name     string
		setupCtx func(fiber.Ctx)
		expected string
	}{
		{
			name: "returns user_id when present as string",
			setupCtx: func(c fiber.Ctx) {
				c.Locals("user_id", "user-123")
			},
			expected: "user-123",
		},
		{
			name: "returns anonymous when user_id is nil",
			setupCtx: func(c fiber.Ctx) {
				// Don't set user_id
			},
			expected: "anonymous",
		},
		{
			name: "returns anonymous when user_id is not a string",
			setupCtx: func(c fiber.Ctx) {
				c.Locals("user_id", 12345) // int instead of string
			},
			expected: "anonymous",
		},
		{
			name: "returns user_id for empty string",
			setupCtx: func(c fiber.Ctx) {
				c.Locals("user_id", "")
			},
			expected: "",
		},
		{
			name: "returns uuid as user_id",
			setupCtx: func(c fiber.Ctx) {
				c.Locals("user_id", "550e8400-e29b-41d4-a716-446655440000")
			},
			expected: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name: "returns email as user_id",
			setupCtx: func(c fiber.Ctx) {
				c.Locals("user_id", "test@example.com")
			},
			expected: "test@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new context for each test
			c := app.AcquireCtx(&fasthttp.RequestCtx{})
			defer app.ReleaseCtx(c)

			// Setup the context
			tt.setupCtx(c)

			// Call the function
			result := getUserID(c)

			assert.Equal(t, tt.expected, result)
		})
	}
}
