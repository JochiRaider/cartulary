// Package verisinputs holds immutable producer policy and independent JSON expectations.
// Runtime consumers never import these offline inputs.
package verisinputs

import "embed"

// Files contains only machine inputs and distribution notices.
//
//go:embed *.json *.txt
var Files embed.FS
