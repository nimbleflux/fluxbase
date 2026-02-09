//go:build !vips
// +build !vips

package storage

import (
	"errors"
	"io"
)

// InitVips is a no-op when vips is not available
func InitVips() {
	// No-op when vips is not compiled in
}

// ShutdownVips is a no-op when vips is not available
func ShutdownVips() {
	// No-op when vips is not compiled in
}

// Transform returns an error when vips is not available
func (t *ImageTransformer) Transform(data io.Reader, contentType string, opts *TransformOptions) (*TransformResult, error) {
	if opts == nil || (opts.Width == 0 && opts.Height == 0 && opts.Format == "") {
		// No transformation requested, return as-is
		return nil, nil
	}

	// Image transformation requested but vips is not available
	return nil, errors.New("image transformation requires vips build tag")
}

// TransformReader returns an error when vips is not available
func (t *ImageTransformer) TransformReader(data io.Reader, contentType string, opts *TransformOptions) (io.ReadCloser, string, int64, error) {
	if opts == nil || (opts.Width == 0 && opts.Height == 0 && opts.Format == "") {
		// No transformation requested, return as-is
		return nil, "", 0, nil
	}

	// Image transformation requested but vips is not available
	return nil, "", 0, errors.New("image transformation requires vips build tag")
}
