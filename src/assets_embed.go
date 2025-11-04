package src

import "embed"

// FS exposes the repository-level src/assets directory for other packages.
// This must be in the root of the project, because embed cannot reference sibling or parent directories.
//
//go:embed assets
var FS embed.FS
