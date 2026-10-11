//go:build !linux

package harnesscollect

import "io"

func Main(_ io.Reader, out io.Writer) int { io.WriteString(out, "{\"type\":\"failed\"}\n"); return 1 }
