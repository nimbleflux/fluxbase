package schema

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// ExtractSchemas writes all embedded schema SQL files to a temporary directory
// and returns the path to that directory. The caller should call Cleanup when done.
func ExtractSchemas() (string, error) {
	tmpDir, err := os.MkdirTemp("", "fluxbase-schemas-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory for schemas: %w", err)
	}

	// Read all files from the embedded "schemas" subdirectory
	entries, err := fs.ReadDir(SchemasFS, "schemas")
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to read embedded schemas: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		data, err := fs.ReadFile(SchemasFS, "schemas/"+entry.Name())
		if err != nil {
			_ = os.RemoveAll(tmpDir)
			return "", fmt.Errorf("failed to read embedded schema file %s: %w", entry.Name(), err)
		}

		outPath := filepath.Join(tmpDir, entry.Name())
		if err := os.WriteFile(outPath, data, 0o644); err != nil {
			_ = os.RemoveAll(tmpDir)
			return "", fmt.Errorf("failed to write schema file %s: %w", entry.Name(), err)
		}
	}

	log.Info().Str("dir", tmpDir).Int("files", len(entries)).Msg("Extracted embedded schema files")
	return tmpDir, nil
}
