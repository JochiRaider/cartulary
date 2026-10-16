//go:build !linux

package harnessfixtures

import (
	"context"
	"io"
)

func RecoverRun(context.Context, string, map[string]string, io.Writer) int { return 2 }
