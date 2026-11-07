// Package attackinputs exposes producer-only machine inputs to source-validation
// tests. The normal Go import closure binds these exact JSON bytes into both Go
// and harness cache identities. Runtime application assembly must not import it.
package attackinputs

import "embed"

// Files contains no human documentation or deployment credentials.
//
//go:embed *.json fixtures/*.json
var Files embed.FS
