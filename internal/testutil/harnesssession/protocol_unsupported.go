//go:build !linux

package harnesssession

import "io"

func Main([]string, io.Reader, io.Writer) int { return 2 }
