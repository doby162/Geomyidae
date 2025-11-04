package assets

import "embed"

// FS exposes the repository-level cmd/assets directory for other packages.
// This must be in the root of the project, because embed cannot reference sibling or parent directories.
//go:embed cmd/assets
var FS embed.FS
