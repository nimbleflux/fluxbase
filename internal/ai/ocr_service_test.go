package ai

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockOCRProvider implements OCRProvider for testing
type mockOCRProvider struct {
	name             string
	available        bool
	extractPDFFunc   func(ctx context.Context, data []byte, languages []string) (*OCRResult, error)
	extractImageFunc func(ctx context.Context, data []byte, languages []string) (*OCRResult, error)
	closeFunc        func() error
}

func newMockOCRProvider(available bool) *mockOCRProvider {
	return &mockOCRProvider{
		name:      "mock",
		available: available,
	}
}

func (m *mockOCRProvider) Name() string {
	return m.name
}

func (m *mockOCRProvider) Type() OCRProviderType {
	return OCRProviderTypeTesseract
}

func (m *mockOCRProvider) IsAvailable() bool {
	return m.available
}

func (m *mockOCRProvider) ExtractTextFromPDF(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
	if m.extractPDFFunc != nil {
		return m.extractPDFFunc(ctx, data, languages)
	}
	return &OCRResult{
		Text:       "Extracted text from PDF",
		Pages:      1,
		Confidence: 0.95,
	}, nil
}

func (m *mockOCRProvider) ExtractTextFromImage(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
	if m.extractImageFunc != nil {
		return m.extractImageFunc(ctx, data, languages)
	}
	return &OCRResult{
		Text:       "Extracted text from image",
		Pages:      1,
		Confidence: 0.90,
	}, nil
}

func (m *mockOCRProvider) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func TestNewOCRService(t *testing.T) {
	t.Run("creates disabled service when disabled in config", func(t *testing.T) {
		cfg := OCRServiceConfig{
			Enabled: false,
		}

		service, err := NewOCRService(cfg)
		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.False(t, service.IsEnabled())
	})

	t.Run("sets default language to eng when not provided", func(t *testing.T) {
		// This test would require mocking NewOCRProvider
		// For now we test the config structure
		cfg := OCRServiceConfig{
			Enabled:          false,
			DefaultLanguages: []string{},
		}

		service, err := NewOCRService(cfg)
		require.NoError(t, err)
		assert.False(t, service.IsEnabled())
	})
}

func TestOCRService_IsEnabled(t *testing.T) {
	t.Run("returns true when enabled", func(t *testing.T) {
		service := &OCRService{
			enabled: true,
		}
		assert.True(t, service.IsEnabled())
	})

	t.Run("returns false when disabled", func(t *testing.T) {
		service := &OCRService{
			enabled: false,
		}
		assert.False(t, service.IsEnabled())
	})
}

func TestOCRService_GetDefaultLanguages(t *testing.T) {
	t.Run("returns configured languages", func(t *testing.T) {
		service := &OCRService{
			defaultLanguages: []string{"eng", "deu", "fra"},
		}

		languages := service.GetDefaultLanguages()
		assert.Equal(t, []string{"eng", "deu", "fra"}, languages)
	})

	t.Run("returns empty slice when no languages configured", func(t *testing.T) {
		service := &OCRService{
			defaultLanguages: nil,
		}

		languages := service.GetDefaultLanguages()
		assert.Nil(t, languages)
	})
}

func TestOCRService_ExtractTextFromPDF(t *testing.T) {
	t.Run("returns error when service is disabled", func(t *testing.T) {
		service := &OCRService{
			enabled: false,
		}

		result, err := service.ExtractTextFromPDF(context.Background(), []byte("pdf data"), nil)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not enabled")
	})

	t.Run("uses default languages when none provided", func(t *testing.T) {
		var capturedLanguages []string
		mock := newMockOCRProvider(true)
		mock.extractPDFFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			capturedLanguages = languages
			return &OCRResult{Text: "text", Pages: 1, Confidence: 0.9}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng", "deu"},
		}

		_, err := service.ExtractTextFromPDF(context.Background(), []byte("pdf"), nil)
		require.NoError(t, err)
		assert.Equal(t, []string{"eng", "deu"}, capturedLanguages)
	})

	t.Run("uses provided languages", func(t *testing.T) {
		var capturedLanguages []string
		mock := newMockOCRProvider(true)
		mock.extractPDFFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			capturedLanguages = languages
			return &OCRResult{Text: "text", Pages: 1, Confidence: 0.9}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		_, err := service.ExtractTextFromPDF(context.Background(), []byte("pdf"), []string{"fra", "spa"})
		require.NoError(t, err)
		assert.Equal(t, []string{"fra", "spa"}, capturedLanguages)
	})

	t.Run("returns result on success", func(t *testing.T) {
		mock := newMockOCRProvider(true)
		mock.extractPDFFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			return &OCRResult{
				Text:       "Extracted PDF text",
				Pages:      3,
				Confidence: 0.95,
			}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		result, err := service.ExtractTextFromPDF(context.Background(), []byte("pdf data"), nil)
		require.NoError(t, err)
		assert.Equal(t, "Extracted PDF text", result.Text)
		assert.Equal(t, 3, result.Pages)
		assert.Equal(t, 0.95, result.Confidence)
	})

	t.Run("wraps provider error", func(t *testing.T) {
		mock := newMockOCRProvider(true)
		mock.extractPDFFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			return nil, assert.AnError
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		result, err := service.ExtractTextFromPDF(context.Background(), []byte("pdf"), nil)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "OCR extraction failed")
	})
}

func TestOCRService_ExtractTextFromImage(t *testing.T) {
	t.Run("returns error when service is disabled", func(t *testing.T) {
		service := &OCRService{
			enabled: false,
		}

		result, err := service.ExtractTextFromImage(context.Background(), []byte("image data"), nil)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not enabled")
	})

	t.Run("uses default languages when none provided", func(t *testing.T) {
		var capturedLanguages []string
		mock := newMockOCRProvider(true)
		mock.extractImageFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			capturedLanguages = languages
			return &OCRResult{Text: "text"}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		_, err := service.ExtractTextFromImage(context.Background(), []byte("img"), nil)
		require.NoError(t, err)
		assert.Equal(t, []string{"eng"}, capturedLanguages)
	})

	t.Run("uses provided languages", func(t *testing.T) {
		var capturedLanguages []string
		mock := newMockOCRProvider(true)
		mock.extractImageFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			capturedLanguages = languages
			return &OCRResult{Text: "text"}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		_, err := service.ExtractTextFromImage(context.Background(), []byte("img"), []string{"jpn"})
		require.NoError(t, err)
		assert.Equal(t, []string{"jpn"}, capturedLanguages)
	})

	t.Run("returns result on success", func(t *testing.T) {
		mock := newMockOCRProvider(true)
		mock.extractImageFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			return &OCRResult{
				Text:       "Image text",
				Pages:      1,
				Confidence: 0.88,
			}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		result, err := service.ExtractTextFromImage(context.Background(), []byte("img"), nil)
		require.NoError(t, err)
		assert.Equal(t, "Image text", result.Text)
		assert.Equal(t, 0.88, result.Confidence)
	})
}

func TestOCRService_Close(t *testing.T) {
	t.Run("closes provider", func(t *testing.T) {
		closed := false
		mock := newMockOCRProvider(true)
		mock.closeFunc = func() error {
			closed = true
			return nil
		}

		service := &OCRService{
			provider: mock,
		}

		err := service.Close()
		require.NoError(t, err)
		assert.True(t, closed)
	})

	t.Run("returns nil when provider is nil", func(t *testing.T) {
		service := &OCRService{
			provider: nil,
		}

		err := service.Close()
		require.NoError(t, err)
	})

	t.Run("returns provider close error", func(t *testing.T) {
		mock := newMockOCRProvider(true)
		mock.closeFunc = func() error {
			return assert.AnError
		}

		service := &OCRService{
			provider: mock,
		}

		err := service.Close()
		require.Error(t, err)
	})
}

func TestOCRServiceConfig_Struct(t *testing.T) {
	cfg := OCRServiceConfig{
		Enabled:          true,
		ProviderType:     OCRProviderTypeTesseract,
		DefaultLanguages: []string{"eng", "deu"},
	}

	assert.True(t, cfg.Enabled)
	assert.Equal(t, OCRProviderTypeTesseract, cfg.ProviderType)
	assert.Equal(t, []string{"eng", "deu"}, cfg.DefaultLanguages)
}

// =============================================================================
// Comprehensive OCR Tests
// =============================================================================

func TestOCRService_ConcurrentAccess(t *testing.T) {
	t.Run("concurrent PDF extractions", func(t *testing.T) {
		mock := newMockOCRProvider(true)
		mock.extractPDFFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			// Simulate some work
			return &OCRResult{Text: "PDF text", Pages: 1, Confidence: 0.9}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		// Run concurrent extractions
		errors := make(chan error, 10)
		for i := 0; i < 10; i++ {
			go func() {
				_, err := service.ExtractTextFromPDF(context.Background(), []byte("pdf"), nil)
				errors <- err
			}()
		}

		// Collect results
		for i := 0; i < 10; i++ {
			err := <-errors
			assert.NoError(t, err)
		}
	})

	t.Run("concurrent image extractions", func(t *testing.T) {
		mock := newMockOCRProvider(true)
		mock.extractImageFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			return &OCRResult{Text: "Image text", Pages: 1, Confidence: 0.85}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		// Run concurrent extractions
		errors := make(chan error, 5)
		for i := 0; i < 5; i++ {
			go func() {
				_, err := service.ExtractTextFromImage(context.Background(), []byte("img"), nil)
				errors <- err
			}()
		}

		// Collect results
		for i := 0; i < 5; i++ {
			err := <-errors
			assert.NoError(t, err)
		}
	})
}

func TestOCRService_MultipleLanguages(t *testing.T) {
	t.Run("PDF with multiple languages", func(t *testing.T) {
		var capturedLanguages []string
		mock := newMockOCRProvider(true)
		mock.extractPDFFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			capturedLanguages = languages
			return &OCRResult{Text: "Mixed language text", Pages: 1, Confidence: 0.82}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng", "fra", "deu"},
		}

		_, err := service.ExtractTextFromPDF(context.Background(), []byte("pdf"), nil)
		require.NoError(t, err)
		assert.Equal(t, []string{"eng", "fra", "deu"}, capturedLanguages)
	})

	t.Run("Image with single language", func(t *testing.T) {
		var capturedLanguages []string
		mock := newMockOCRProvider(true)
		mock.extractImageFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			capturedLanguages = languages
			return &OCRResult{Text: "Single language text", Pages: 1, Confidence: 0.91}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"jpn"},
		}

		_, err := service.ExtractTextFromImage(context.Background(), []byte("img"), nil)
		require.NoError(t, err)
		assert.Equal(t, []string{"jpn"}, capturedLanguages)
	})

	t.Run("override default languages", func(t *testing.T) {
		var capturedLanguages []string
		mock := newMockOCRProvider(true)
		mock.extractPDFFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			capturedLanguages = languages
			return &OCRResult{Text: "text"}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng", "deu"},
		}

		// Override with different languages
		_, err := service.ExtractTextFromPDF(context.Background(), []byte("pdf"), []string{"spa", "ita"})
		require.NoError(t, err)
		assert.Equal(t, []string{"spa", "ita"}, capturedLanguages)
	})
}

func TestOCRService_ErrorHandling(t *testing.T) {
	t.Run("PDF extraction timeout", func(t *testing.T) {
		mock := newMockOCRProvider(true)
		mock.extractPDFFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			// Simulate timeout by checking context
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1)
		defer cancel()

		result, err := service.ExtractTextFromPDF(ctx, []byte("pdf"), nil)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("Image extraction provider failure", func(t *testing.T) {
		mock := newMockOCRProvider(true)
		mock.extractImageFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			return nil, fmt.Errorf("image format not supported")
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		result, err := service.ExtractTextFromImage(context.Background(), []byte("img"), nil)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "image format not supported")
	})

	t.Run("empty PDF data", func(t *testing.T) {
		mock := newMockOCRProvider(true)
		mock.extractPDFFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			if len(data) == 0 {
				return nil, fmt.Errorf("empty PDF data")
			}
			return &OCRResult{Text: "text"}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		result, err := service.ExtractTextFromPDF(context.Background(), []byte{}, nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("corrupted image data", func(t *testing.T) {
		mock := newMockOCRProvider(true)
		mock.extractImageFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			return nil, fmt.Errorf("failed to decode image")
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		result, err := service.ExtractTextFromImage(context.Background(), []byte("corrupted"), nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestOCRService_MultiPageDocuments(t *testing.T) {
	t.Run("PDF with multiple pages", func(t *testing.T) {
		mock := newMockOCRProvider(true)
		mock.extractPDFFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			return &OCRResult{
				Text:       "Page 1 text\nPage 2 text\nPage 3 text",
				Pages:      3,
				Confidence: 0.92,
			}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		result, err := service.ExtractTextFromPDF(context.Background(), []byte("multipage"), nil)
		require.NoError(t, err)
		assert.Equal(t, 3, result.Pages)
		assert.Contains(t, result.Text, "Page 1")
		assert.Contains(t, result.Text, "Page 2")
		assert.Contains(t, result.Text, "Page 3")
	})

	t.Run("PDF with high confidence", func(t *testing.T) {
		mock := newMockOCRProvider(true)
		mock.extractPDFFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			return &OCRResult{
				Text:       "High confidence text",
				Pages:      1,
				Confidence: 0.98,
			}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		result, err := service.ExtractTextFromPDF(context.Background(), []byte("pdf"), nil)
		require.NoError(t, err)
		assert.Greater(t, result.Confidence, 0.95)
	})

	t.Run("PDF with low confidence warning", func(t *testing.T) {
		mock := newMockOCRProvider(true)
		mock.extractPDFFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			return &OCRResult{
				Text:       "Low quality text",
				Pages:      1,
				Confidence: 0.65,
			}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		result, err := service.ExtractTextFromPDF(context.Background(), []byte("pdf"), nil)
		require.NoError(t, err)
		assert.Less(t, result.Confidence, 0.70)
	})
}

func TestOCRService_ResultStruct(t *testing.T) {
	t.Run("OCR result with all fields", func(t *testing.T) {
		result := OCRResult{
			Text:       "Sample text",
			Pages:      2,
			Confidence: 0.89,
		}

		assert.Equal(t, "Sample text", result.Text)
		assert.Equal(t, 2, result.Pages)
		assert.Equal(t, 0.89, result.Confidence)
	})

	t.Run("OCR result with zero values", func(t *testing.T) {
		result := OCRResult{}

		assert.Empty(t, result.Text)
		assert.Zero(t, result.Pages)
		assert.Zero(t, result.Confidence)
	})

	t.Run("OCR result with very long text", func(t *testing.T) {
		longText := string(make([]byte, 10000))
		for i := range longText {
			longText = longText[:i] + "A" + longText[i+1:]
		}

		result := OCRResult{
			Text:       longText,
			Pages:      10,
			Confidence: 0.95,
		}

		assert.Equal(t, 10000, len(result.Text))
		assert.Equal(t, 10, result.Pages)
	})
}

func TestOCRService_ServiceLifecycle(t *testing.T) {
	t.Run("create and close service", func(t *testing.T) {
		closed := false
		mock := newMockOCRProvider(true)
		mock.closeFunc = func() error {
			closed = true
			return nil
		}

		service := &OCRService{
			enabled:  true,
			provider: mock,
		}

		err := service.Close()
		require.NoError(t, err)
		assert.True(t, closed)
	})

	t.Run("use service after close", func(t *testing.T) {
		mock := newMockOCRProvider(true)
		mock.extractPDFFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			return &OCRResult{Text: "text"}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		// Close the service
		_ = service.Close()

		// Try to use it (should still work as provider is not set to nil)
		// In production, you'd want to check if closed
		result, err := service.ExtractTextFromPDF(context.Background(), []byte("pdf"), nil)
		// Mock still works, so no error
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestOCRService_ProviderTypes(t *testing.T) {
	t.Run("Tesseract provider type", func(t *testing.T) {
		providerType := OCRProviderTypeTesseract
		assert.Equal(t, "tesseract", string(providerType))
	})

	t.Run("mock provider type", func(t *testing.T) {
		mock := newMockOCRProvider(true)
		assert.Equal(t, "mock", mock.Name())
		assert.Equal(t, OCRProviderTypeTesseract, mock.Type())
		assert.True(t, mock.IsAvailable())
	})

	t.Run("unavailable provider", func(t *testing.T) {
		mock := newMockOCRProvider(false)
		assert.False(t, mock.IsAvailable())
	})
}
