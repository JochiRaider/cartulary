//go:build !linux

package harnessreview

import "io"

func Main([]string, io.Reader, io.Writer) int { return 2 }
