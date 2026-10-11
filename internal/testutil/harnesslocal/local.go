// Package harnesslocal owns private diagnostic and admission file access.
// It never follows symlinks and never repairs an unsafe existing entry.
package harnesslocal

import "errors"

var ErrUnsafe = errors.New("unsafe private harness file")
