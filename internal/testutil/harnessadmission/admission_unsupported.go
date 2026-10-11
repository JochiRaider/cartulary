//go:build !linux

package harnessadmission

import "io"

func Main([]string, io.Writer) int { return 2 }
