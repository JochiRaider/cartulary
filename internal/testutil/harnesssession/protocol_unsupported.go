//go:build !linux

package harnesssession

import "io"

func Main([]string, io.Reader, io.Writer) int { return 2 }
func ProcessMain([]string, io.Reader) int     { return 2 }
