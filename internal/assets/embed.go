package assets

import "embed"

// Files contains the release payload materialized by the installer.
//
//go:embed data
var Files embed.FS
