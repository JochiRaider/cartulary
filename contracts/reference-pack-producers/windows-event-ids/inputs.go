// Package windowseventinputs contains offline producer contracts only. Runtime
// assembly never imports these inputs or accesses upstream documentation.
package windowseventinputs

import "embed"

// Files binds executable producer policy and fixtures, never Markdown.
//
//go:embed *.json
var Files embed.FS
