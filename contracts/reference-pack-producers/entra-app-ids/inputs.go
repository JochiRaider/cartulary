// Package entrainputs contains the immutable producer policy and independent
// factual expectations. Runtime consumers do not import producer inputs.
package entrainputs

import "embed"

// Files contains JSON policy and project notices, never upstream Markdown.
//
//go:embed *.json *.txt
var Files embed.FS
