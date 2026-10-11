//go:build !linux

package harnessruntime

import (
	"context"
	"errors"
)

func Prepare(context.Context, string, []string, []string) (*Plan, error) {
	return nil, errors.New("native harness preparation requires Linux")
}
