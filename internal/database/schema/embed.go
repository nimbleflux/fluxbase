package schema

import "embed"

// SchemasFS holds the embedded Fluxbase internal schema SQL files.
// These files are baked into the binary and extracted to a temporary
// directory at runtime so they work regardless of deployment path.
//
//go:embed schemas/*.sql
var SchemasFS embed.FS
