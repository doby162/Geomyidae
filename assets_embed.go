package assets

import "embed"

// FS exposes the repository-level cmd/assets directory for other packages.
//go:embed cmd/assets
var FS embed.FS
