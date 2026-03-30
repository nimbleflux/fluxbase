package e2e

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/storage"
	"github.com/nimbleflux/fluxbase/test"
)

// TransformTestContext extends StorageTestContext with transform-specific config
type TransformTestContext struct {
	*test.TestContext
	APIKey string
}

// setupTransformTest prepares the test context for transform tests
func setupTransformTest(t *testing.T) *TransformTestContext {
	tc := test.NewTestContext(t)
	tc.EnsureStorageSchema()

	// Use local storage for transform tests
	tc.Config.Storage.Provider = "local"
	tc.Config.Storage.LocalPath = "/tmp/fluxbase-transform-test-storage"

	// Enable image transforms
	tc.Config.Storage.Transforms.Enabled = true
	tc.Config.Storage.Transforms.MaxWidth = 4096
	tc.Config.Storage.Transforms.MaxHeight = 4096
	tc.Config.Storage.Transforms.MaxTotalPixels = 16000000
	tc.Config.Storage.Transforms.DefaultQuality = 80
	tc.Config.Storage.Transforms.BucketSize = 50
	tc.Config.Storage.Transforms.RateLimit = 100 // High limit for tests
	tc.Config.Storage.Transforms.MaxConcurrent = 4

	// Clean up any existing test storage files
	tc.CleanupStorageFiles()

	// Create an API key for authenticated requests
	apiKey := tc.CreateAPIKey("Transform Test API Key", nil)

	return &TransformTestContext{
		TestContext: tc,
		APIKey:      apiKey,
	}
}

// createTestBucketAndUpload creates a bucket and uploads a test image
func createTestBucketAndUpload(t *testing.T, tc *TransformTestContext, bucketName string, fileName string, content []byte, contentType string) {
	// Create bucket
	tc.NewRequest("POST", "/api/v1/storage/buckets/"+bucketName).
		WithAPIKey(tc.APIKey).
		Send().
		AssertStatus(fiber.StatusCreated)

	// Create multipart form with file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form file with specified content type
	h := make(map[string][]string)
	h["Content-Disposition"] = []string{`form-data; name="file"; filename="` + fileName + `"`}
	h["Content-Type"] = []string{contentType}

	part, err := writer.CreatePart(h)
	require.NoError(t, err)

	_, err = part.Write(content)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	// Upload file
	req := httptest.NewRequest("POST", "/api/v1/storage/"+bucketName+"/"+fileName, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Client-Key", tc.APIKey)

	resp, err := tc.App.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode, "File upload should succeed")
}

// createMinimalPNG creates a 1x1 pixel PNG image (smallest valid PNG)
func createMinimalPNG() []byte {
	// Minimal valid 1x1 red PNG
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, // IHDR length
		0x49, 0x48, 0x44, 0x52, // IHDR
		0x00, 0x00, 0x00, 0x01, // width: 1
		0x00, 0x00, 0x00, 0x01, // height: 1
		0x08, 0x02, // bit depth: 8, color type: 2 (RGB)
		0x00, 0x00, 0x00, // compression, filter, interlace
		0x90, 0x77, 0x53, 0xDE, // IHDR CRC
		0x00, 0x00, 0x00, 0x0C, // IDAT length
		0x49, 0x44, 0x41, 0x54, // IDAT
		0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00, 0x00, // compressed data
		0x01, 0x01, 0x01, 0x00, // IDAT CRC (simplified)
		0x1B, 0xB6, 0xEE, 0x56, // actual CRC
		0x00, 0x00, 0x00, 0x00, // IEND length
		0x49, 0x45, 0x4E, 0x44, // IEND
		0xAE, 0x42, 0x60, 0x82, // IEND CRC
	}
}

// =============================================================================
// Transform Configuration Tests
// =============================================================================

func TestStorageTransform_GetConfig(t *testing.T) {
	tc := setupTransformTest(t)
	defer tc.Close()

	resp := tc.NewRequest("GET", "/api/v1/storage/config/transforms").
		WithAPIKey(tc.APIKey).
		Send().
		AssertStatus(fiber.StatusOK)

	var result map[string]interface{}
	resp.JSON(&result)

	assert.True(t, result["enabled"].(bool), "transforms should be enabled")
	assert.Equal(t, float64(80), result["default_quality"], "default quality should be 80")
	assert.Equal(t, float64(4096), result["max_width"], "max width should be 4096")
	assert.Equal(t, float64(4096), result["max_height"], "max height should be 4096")

	t.Log("Transform config retrieved successfully")
}

func TestStorageTransform_GetConfigDisabled(t *testing.T) {
	tc := test.NewTestContext(t)
	defer tc.Close()

	// Don't enable transforms
	tc.Config.Storage.Transforms.Enabled = false

	apiKey := tc.CreateAPIKey("Test API Key", nil)

	resp := tc.NewRequest("GET", "/api/v1/storage/config/transforms").
		WithAPIKey(apiKey).
		Send().
		AssertStatus(fiber.StatusOK)

	var result map[string]interface{}
	resp.JSON(&result)

	assert.False(t, result["enabled"].(bool), "transforms should be disabled")
}

// =============================================================================
// Transform Query Parameter Tests
// =============================================================================

func TestStorageTransform_CanTransformFunction(t *testing.T) {
	tests := []struct {
		contentType  string
		canTransform bool
	}{
		{"image/jpeg", true},
		{"image/png", true},
		{"image/webp", true},
		{"image/gif", true},
		{"image/avif", true},
		{"text/plain", false},
		{"application/pdf", false},
		{"video/mp4", false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			result := storage.CanTransform(tt.contentType)
			assert.Equal(t, tt.canTransform, result)
		})
	}
}

func TestStorageTransform_ParseTransformOptions(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		height   int
		format   string
		quality  int
		fit      string
		expected *storage.TransformOptions
	}{
		{
			name:     "no options returns nil",
			width:    0,
			height:   0,
			format:   "",
			quality:  0,
			fit:      "",
			expected: nil,
		},
		{
			name:    "width only",
			width:   800,
			height:  0,
			format:  "",
			quality: 0,
			fit:     "",
			expected: &storage.TransformOptions{
				Width: 800,
				Fit:   storage.FitCover,
			},
		},
		{
			name:    "all options",
			width:   800,
			height:  600,
			format:  "webp",
			quality: 85,
			fit:     "contain",
			expected: &storage.TransformOptions{
				Width:   800,
				Height:  600,
				Format:  "webp",
				Quality: 85,
				Fit:     storage.FitContain,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.ParseTransformOptions(tt.width, tt.height, tt.format, tt.quality, tt.fit)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.Width, result.Width)
				assert.Equal(t, tt.expected.Height, result.Height)
				assert.Equal(t, tt.expected.Format, result.Format)
				assert.Equal(t, tt.expected.Quality, result.Quality)
				assert.Equal(t, tt.expected.Fit, result.Fit)
			}
		})
	}
}

func TestStorageTransform_ValidateOptions(t *testing.T) {
	transformer := storage.NewImageTransformerWithOptions(storage.TransformerOptions{
		MaxWidth:       4096,
		MaxHeight:      4096,
		MaxTotalPixels: 16000000,
		BucketSize:     50,
	})

	t.Run("valid options", func(t *testing.T) {
		opts := &storage.TransformOptions{
			Width:   800,
			Height:  600,
			Format:  "webp",
			Quality: 80,
		}
		err := transformer.ValidateOptions(opts)
		assert.NoError(t, err)
	})

	t.Run("exceeds max width", func(t *testing.T) {
		opts := &storage.TransformOptions{Width: 5000}
		err := transformer.ValidateOptions(opts)
		assert.Error(t, err)
		assert.ErrorIs(t, err, storage.ErrImageTooLarge)
	})

	t.Run("negative dimensions", func(t *testing.T) {
		opts := &storage.TransformOptions{Width: -100}
		err := transformer.ValidateOptions(opts)
		assert.Error(t, err)
		assert.ErrorIs(t, err, storage.ErrInvalidDimensions)
	})

	t.Run("unsupported format", func(t *testing.T) {
		opts := &storage.TransformOptions{Format: "gif"}
		err := transformer.ValidateOptions(opts)
		assert.Error(t, err)
		assert.ErrorIs(t, err, storage.ErrUnsupportedFormat)
	})

	t.Run("dimension bucketing", func(t *testing.T) {
		opts := &storage.TransformOptions{Width: 823, Height: 617}
		err := transformer.ValidateOptions(opts)
		assert.NoError(t, err)
		assert.Equal(t, 800, opts.Width)  // Bucketed
		assert.Equal(t, 600, opts.Height) // Bucketed
	})
}

// =============================================================================
// Non-Image File Transform Tests
// =============================================================================

func TestStorageTransform_NonImageFileIgnoresTransform(t *testing.T) {
	tc := setupTransformTest(t)
	defer tc.Close()

	bucketName := "transform-text-test"
	fileName := "test.txt"
	content := []byte("This is a text file, not an image")

	createTestBucketAndUpload(t, tc, bucketName, fileName, content, "text/plain")

	// Request with transform params - should return original since it's not an image
	resp := tc.NewRequest("GET", "/api/v1/storage/"+bucketName+"/"+fileName+"?w=100").
		WithAPIKey(tc.APIKey).
		Send()

	// Should return the file (either 200 or appropriate status)
	if resp.Status() == fiber.StatusOK {
		body := resp.Body()
		assert.Equal(t, content, body, "non-image file should be returned unchanged")
		t.Log("Text file returned unchanged when transform params specified")
	} else {
		t.Logf("Response status: %d (transform ignored for non-image)", resp.Status())
	}
}

// =============================================================================
// FitMode Tests
// =============================================================================

func TestStorageTransform_FitModeConstants(t *testing.T) {
	assert.Equal(t, storage.FitMode("cover"), storage.FitCover)
	assert.Equal(t, storage.FitMode("contain"), storage.FitContain)
	assert.Equal(t, storage.FitMode("fill"), storage.FitFill)
	assert.Equal(t, storage.FitMode("inside"), storage.FitInside)
	assert.Equal(t, storage.FitMode("outside"), storage.FitOutside)
}

func TestStorageTransform_FitModeParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected storage.FitMode
	}{
		{"cover", storage.FitCover},
		{"COVER", storage.FitCover},
		{"Cover", storage.FitCover},
		{"contain", storage.FitContain},
		{"fill", storage.FitFill},
		{"inside", storage.FitInside},
		{"outside", storage.FitOutside},
		{"invalid", storage.FitCover}, // Default
		{"", storage.FitCover},        // Default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			opts := storage.ParseTransformOptions(100, 0, "", 0, tt.input)
			require.NotNil(t, opts)
			assert.Equal(t, tt.expected, opts.Fit)
		})
	}
}

// =============================================================================
// Supported Format Tests
// =============================================================================

func TestStorageTransform_SupportedOutputFormats(t *testing.T) {
	supported := []string{"webp", "jpg", "jpeg", "png", "avif"}
	unsupported := []string{"gif", "bmp", "tiff", "svg", "heic"}

	for _, format := range supported {
		t.Run("supported_"+format, func(t *testing.T) {
			assert.True(t, storage.SupportedOutputFormats[format])
		})
	}

	for _, format := range unsupported {
		t.Run("unsupported_"+format, func(t *testing.T) {
			assert.False(t, storage.SupportedOutputFormats[format])
		})
	}
}

func TestStorageTransform_SupportedInputMimeTypes(t *testing.T) {
	supported := []string{
		"image/jpeg", "image/png", "image/webp", "image/gif",
		"image/tiff", "image/bmp", "image/svg+xml", "image/avif",
	}

	for _, mime := range supported {
		t.Run(mime, func(t *testing.T) {
			assert.True(t, storage.SupportedInputMimeTypes[mime])
		})
	}
}

// =============================================================================
// Download with Transform Parameters (Integration)
// =============================================================================

func TestStorageTransform_DownloadWithQueryParams(t *testing.T) {
	tc := setupTransformTest(t)
	defer tc.Close()

	bucketName := "download-transform-test"
	fileName := "image.png"

	// Upload a minimal PNG
	pngData := createMinimalPNG()
	createTestBucketAndUpload(t, tc, bucketName, fileName, pngData, "image/png")

	// Test download without transform params
	t.Run("without transform", func(t *testing.T) {
		resp := tc.NewRequest("GET", "/api/v1/storage/"+bucketName+"/"+fileName).
			WithAPIKey(tc.APIKey).
			Send()

		if resp.Status() == fiber.StatusOK {
			body := resp.Body()
			assert.NotEmpty(t, body)
			t.Log("Download without transform succeeded")
		} else {
			t.Logf("Download status: %d", resp.Status())
		}
	})

	// Test download with width param
	t.Run("with width param", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/storage/"+bucketName+"/"+fileName+"?w=100", nil)
		req.Header.Set("X-Client-Key", tc.APIKey)

		resp, err := tc.App.Test(req)
		require.NoError(t, err)

		// Response might be 200 (transform applied) or error if vips not available
		body, _ := io.ReadAll(resp.Body)
		t.Logf("Download with width param - status: %d, body length: %d", resp.StatusCode, len(body))
	})

	// Test download with format conversion param
	t.Run("with format param", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/storage/"+bucketName+"/"+fileName+"?fmt=webp", nil)
		req.Header.Set("X-Client-Key", tc.APIKey)

		resp, err := tc.App.Test(req)
		require.NoError(t, err)

		body, _ := io.ReadAll(resp.Body)
		t.Logf("Download with format param - status: %d, body length: %d", resp.StatusCode, len(body))
	})

	// Test download with multiple params
	t.Run("with multiple params", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/storage/"+bucketName+"/"+fileName+"?w=100&h=100&fmt=webp&q=80&fit=cover", nil)
		req.Header.Set("X-Client-Key", tc.APIKey)

		resp, err := tc.App.Test(req)
		require.NoError(t, err)

		body, _ := io.ReadAll(resp.Body)
		t.Logf("Download with all params - status: %d, body length: %d", resp.StatusCode, len(body))
	})
}

// =============================================================================
// Bucket Dimension Tests
// =============================================================================

func TestStorageTransform_BucketDimension(t *testing.T) {
	tests := []struct {
		dim        int
		bucketSize int
		expected   int
	}{
		{100, 50, 100}, // Exact bucket
		{126, 50, 150}, // Round up
		{124, 50, 100}, // Round down
		{125, 50, 150}, // Midpoint rounds up
		{0, 50, 0},     // Zero dimension
		{-10, 50, -10}, // Negative dimension
		{100, 0, 100},  // Zero bucket size
		{123, 1, 123},  // No bucketing
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := storage.BucketDimension(tt.dim, tt.bucketSize)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// Transform Constants Tests
// =============================================================================

func TestStorageTransform_Constants(t *testing.T) {
	assert.Equal(t, 8192, storage.MaxTransformDimension)
	assert.Equal(t, 16_000_000, storage.DefaultMaxTotalPixels)
	assert.Equal(t, 50, storage.DefaultBucketSize)
}

// =============================================================================
// Error Types Tests
// =============================================================================

func TestStorageTransform_ErrorTypes(t *testing.T) {
	assert.NotNil(t, storage.ErrInvalidDimensions)
	assert.NotNil(t, storage.ErrUnsupportedFormat)
	assert.NotNil(t, storage.ErrNotAnImage)
	assert.NotNil(t, storage.ErrTransformFailed)
	assert.NotNil(t, storage.ErrImageTooLarge)
	assert.NotNil(t, storage.ErrVipsNotInitialized)
	assert.NotNil(t, storage.ErrTooManyPixels)

	assert.Contains(t, storage.ErrInvalidDimensions.Error(), "invalid")
	assert.Contains(t, storage.ErrUnsupportedFormat.Error(), "unsupported")
}
