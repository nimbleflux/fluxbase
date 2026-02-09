package storage

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	ErrInvalidDimensions  = errors.New("invalid image dimensions")
	ErrUnsupportedFormat  = errors.New("unsupported output format")
	ErrNotAnImage         = errors.New("file is not an image")
	ErrTransformFailed    = errors.New("image transformation failed")
	ErrImageTooLarge      = errors.New("image exceeds maximum allowed dimensions")
	ErrVipsNotInitialized = errors.New("vips library not initialized")
	ErrTooManyPixels      = errors.New("total pixel count exceeds maximum")
)

// MaxTransformDimension is the maximum allowed dimension for transformed images
const MaxTransformDimension = 8192

// DefaultMaxTotalPixels is the default maximum total pixel count (16 megapixels)
const DefaultMaxTotalPixels = 16_000_000

// DefaultBucketSize is the default dimension bucketing size (50px)
const DefaultBucketSize = 50

// BucketDimension rounds a dimension to the nearest bucket size
// This reduces cache key variations and provides DoS protection
func BucketDimension(dim int, bucketSize int) int {
	if dim <= 0 || bucketSize <= 0 {
		return dim
	}
	return ((dim + bucketSize/2) / bucketSize) * bucketSize
}

// SupportedOutputFormats lists the supported output formats
var SupportedOutputFormats = map[string]bool{
	"webp": true,
	"jpg":  true,
	"jpeg": true,
	"png":  true,
	"avif": true,
}

// SupportedInputMimeTypes lists MIME types that can be transformed
var SupportedInputMimeTypes = map[string]bool{
	"image/jpeg":    true,
	"image/png":     true,
	"image/webp":    true,
	"image/gif":     true,
	"image/tiff":    true,
	"image/bmp":     true,
	"image/svg+xml": true,
	"image/avif":    true,
}

// TransformerOptions configures the image transformer
type TransformerOptions struct {
	MaxWidth       int
	MaxHeight      int
	MaxTotalPixels int
	BucketSize     int
}

// FitMode defines how the image should be fit within the target dimensions
type FitMode string

const (
	FitCover   FitMode = "cover"   // Resize to cover target dimensions, cropping if needed
	FitContain FitMode = "contain" // Resize to fit within target dimensions, letterboxing if needed
	FitFill    FitMode = "fill"    // Stretch to exactly fill target dimensions
	FitInside  FitMode = "inside"  // Resize to fit within target, only scale down
	FitOutside FitMode = "outside" // Resize to be at least as large as target
)

// TransformOptions contains parameters for image transformation
type TransformOptions struct {
	Width   int     // Target width in pixels (0 = auto based on height)
	Height  int     // Target height in pixels (0 = auto based on width)
	Format  string  // Output format: webp, jpg, jpeg, png, avif (empty = same as input)
	Quality int     // Output quality 1-100 (default 80)
	Fit     FitMode // How to fit the image (default cover)
}

// TransformResult contains the result of an image transformation
type TransformResult struct {
	Data        []byte
	ContentType string
	Width       int
	Height      int
}

// ImageTransformer handles image transformations
// Fields are defined in platform-specific files (transform.go or transform_novips.go)
type ImageTransformer struct {
	initialized    bool
	maxWidth       int
	maxHeight      int
	maxTotalPixels int
	bucketSize     int
}

// TransformInterface defines the interface for image transformation operations
// This is used by the cache and other components
type TransformInterface interface {
	Transform(data io.Reader, contentType string, opts *TransformOptions) (*TransformResult, error)
	TransformReader(data io.Reader, contentType string, opts *TransformOptions) (io.ReadCloser, string, int64, error)
}

// Ensure ImageTransformer implements the interface
var _ TransformInterface = (*ImageTransformer)(nil)

// CanTransform checks if the given content type can be transformed
func CanTransform(contentType string) bool {
	// Normalize content type (remove charset, etc.)
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	return SupportedInputMimeTypes[strings.ToLower(contentType)]
}

// ParseTransformOptions parses query parameters into TransformOptions
func ParseTransformOptions(width, height int, format string, quality int, fit string) *TransformOptions {
	// Return nil if no transform options specified
	if width == 0 && height == 0 && format == "" && quality == 0 && fit == "" {
		return nil
	}

	opts := &TransformOptions{
		Width:   width,
		Height:  height,
		Format:  format,
		Quality: quality,
	}

	// Parse fit mode
	switch strings.ToLower(fit) {
	case "cover":
		opts.Fit = FitCover
	case "contain":
		opts.Fit = FitContain
	case "fill":
		opts.Fit = FitFill
	case "inside":
		opts.Fit = FitInside
	case "outside":
		opts.Fit = FitOutside
	default:
		opts.Fit = FitCover
	}

	return opts
}

// NewImageTransformerWithOptions creates a new image transformer with full options
// This is a stub that returns an error when transformation is attempted without vips support
func NewImageTransformerWithOptions(opts TransformerOptions) *ImageTransformer {
	if opts.MaxWidth <= 0 {
		opts.MaxWidth = MaxTransformDimension
	}
	if opts.MaxHeight <= 0 {
		opts.MaxHeight = MaxTransformDimension
	}
	if opts.MaxTotalPixels <= 0 {
		opts.MaxTotalPixels = DefaultMaxTotalPixels
	}
	if opts.BucketSize <= 0 {
		opts.BucketSize = DefaultBucketSize
	}

	return &ImageTransformer{
		initialized:    false, // Not initialized without vips
		maxWidth:       opts.MaxWidth,
		maxHeight:      opts.MaxHeight,
		maxTotalPixels: opts.MaxTotalPixels,
		bucketSize:     opts.BucketSize,
	}
}

// ValidateOptions validates and normalizes transform options
func (t *ImageTransformer) ValidateOptions(opts *TransformOptions) error {
	if opts == nil {
		return nil // No transformation requested
	}

	// Validate dimensions
	if opts.Width < 0 || opts.Height < 0 {
		return ErrInvalidDimensions
	}

	if opts.Width > 0 && opts.Width > t.maxWidth {
		return fmt.Errorf("%w: width %d exceeds maximum %d", ErrImageTooLarge, opts.Width, t.maxWidth)
	}

	if opts.Height > 0 && opts.Height > t.maxHeight {
		return fmt.Errorf("%w: height %d exceeds maximum %d", ErrImageTooLarge, opts.Height, t.maxHeight)
	}

	// Calculate total pixels
	totalPixels := opts.Width * opts.Height
	if totalPixels > 0 && totalPixels > t.maxTotalPixels {
		return fmt.Errorf("%w: %dx%d = %d pixels exceeds maximum %d",
			ErrTooManyPixels, opts.Width, opts.Height, totalPixels, t.maxTotalPixels)
	}

	// Validate format
	if opts.Format != "" {
		opts.Format = strings.ToLower(strings.TrimSpace(opts.Format))
		if !SupportedOutputFormats[opts.Format] {
			return ErrUnsupportedFormat
		}
	}

	// Validate quality
	if opts.Quality < 0 || opts.Quality > 100 {
		opts.Quality = 80
	}

	// Normalize fit mode
	if opts.Fit == "" {
		opts.Fit = FitCover
	}

	// Bucket dimensions for caching and DoS protection
	if t.bucketSize > 0 {
		if opts.Width > 0 {
			opts.Width = BucketDimension(opts.Width, t.bucketSize)
		}
		if opts.Height > 0 {
			opts.Height = BucketDimension(opts.Height, t.bucketSize)
		}
	}

	return nil
}
