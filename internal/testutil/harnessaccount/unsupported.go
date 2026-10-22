//go:build !linux

package harnessaccount

import (
	"errors"
	"io"
)

func Main([]string, io.Writer, io.Writer) int { return 2 }
func ContainerParent() (string, error)        { return "", nil }
func InvocationRoot() (string, error)         { return "", nil }
func VerifyContainer(string, string) error    { return errors.New("invocation accounting unsupported") }
