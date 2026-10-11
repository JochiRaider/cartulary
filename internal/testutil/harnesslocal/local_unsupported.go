//go:build !linux

package harnesslocal

import "os"

func OpenDirectory(string, bool) (*os.File, error) { return nil, ErrUnsafe }
func ReadFile(string, int64, bool) ([]byte, error) { return nil, ErrUnsafe }
func PublishNew(string, []byte) error              { return ErrUnsafe }

func PrivateDirectory(string) (*os.File, error) { return nil, ErrUnsafe }
func ReplaceFile(string, []byte) error          { return ErrUnsafe }

func ReadFileAt(*os.File, string, int64, bool) ([]byte, error) { return nil, ErrUnsafe }
func ReplaceFileAt(*os.File, string, []byte) error             { return ErrUnsafe }
func CheckDirectory(*os.File, string) error                    { return ErrUnsafe }
