//go:build linux

package harnessruntime

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"golang.org/x/sys/unix"
)

// The caller supplies its already validated, run-owned private runtime handle.
// The channel retains descriptor identity and can remove only its own closed
// set of files. It never accepts a path from a retained document.
type privateFailureScope struct {
	parent, directory *os.File
	name              string
	expected          map[string]any
	closed            bool
}

func createFailureScope(runtime *os.File, identity map[string]any) (failureScope, error) {
	if err := harnesslocal.CheckDirectory(runtime, runtime.Name()); err != nil {
		return nil, err
	}
	expected := map[string]any{"schema_id": "cartulary.harness_command_failure.v1"}
	for _, key := range []string{"run_id", "unit_id", "command_id", "invocation_id"} {
		expected[key] = identity[key]
	}
	probe := map[string]any{}
	for k, v := range expected {
		probe[k] = v
	}
	probe["failure_class"] = "harness"
	probe["failure_reason"] = "scheduler_accounting_error"
	if _, err := documentBytes("cartulary.harness_command_failure.v1", probe); err != nil {
		return nil, err
	}
	detail, err := harnesslocal.PrivateDirectory(filepath.Join(runtime.Name(), "detail"))
	if err != nil {
		return nil, err
	}
	runtime = detail
	name := "command-failure-" + identity["invocation_id"].(string)
	if err := unix.Mkdirat(int(runtime.Fd()), name, 0700); err != nil {
		detail.Close()
		return nil, err
	}
	fd, err := unix.Openat(int(runtime.Fd()), name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		_ = unix.Unlinkat(int(runtime.Fd()), name, unix.AT_REMOVEDIR)
		detail.Close()
		return nil, err
	}
	return &privateFailureScope{parent: runtime, directory: os.NewFile(uintptr(fd), filepath.Join(runtime.Name(), name)), name: name, expected: expected}, nil
}
func (s *privateFailureScope) Environment() map[string]string {
	bytes, _ := json.Marshal(s.expected)
	return map[string]string{failureContextKey: string(bytes)}
}
func (s *privateFailureScope) Read() *failure {
	accounting := func() *failure { return failed("harness", "scheduler_accounting_error") }
	if s.closed || harnesslocal.CheckDirectory(s.parent, s.parent.Name()) != nil || harnesslocal.CheckDirectory(s.directory, filepath.Join(s.parent.Name(), s.name)) != nil {
		return accounting()
	}
	var stat unix.Stat_t
	if err := unix.Fstatat(int(s.directory.Fd()), "conflict", &stat, unix.AT_SYMLINK_NOFOLLOW); err == nil || !errors.Is(err, unix.ENOENT) {
		return accounting()
	}
	bytes, err := harnesslocal.ReadFileAt(s.directory, "failure.json", 4096, false)
	if errors.Is(err, os.ErrNotExist) {
		if e := unix.Fstatat(int(s.directory.Fd()), "writer", &stat, unix.AT_SYMLINK_NOFOLLOW); errors.Is(e, unix.ENOENT) {
			return nil
		}
		return accounting()
	}
	if err != nil {
		return accounting()
	}
	value, err := harnessschema.Decode(bytes)
	if err != nil || harnessschema.Validate("cartulary.harness_command_failure.v1", value) != nil {
		return accounting()
	}
	record := value.(map[string]any)
	for k, v := range s.expected {
		if record[k] != v {
			return accounting()
		}
	}
	class, reason := record["failure_class"].(string), record["failure_reason"].(string)
	if failureReasonClasses[reason] != class {
		return accounting()
	}
	return failed(class, reason)
}
func (s *privateFailureScope) Close() error {
	if s.closed {
		return nil
	}
	if harnesslocal.CheckDirectory(s.parent, s.parent.Name()) != nil || harnesslocal.CheckDirectory(s.directory, filepath.Join(s.parent.Name(), s.name)) != nil {
		return harnesslocal.ErrUnsafe
	}
	for _, name := range []string{"failure.json", "pending.json", "conflict"} {
		if err := unix.Unlinkat(int(s.directory.Fd()), name, 0); err != nil && !errors.Is(err, unix.ENOENT) {
			return err
		}
	}
	if err := unix.Unlinkat(int(s.directory.Fd()), "writer", unix.AT_REMOVEDIR); err != nil && !errors.Is(err, unix.ENOENT) {
		return err
	}
	if err := unix.Unlinkat(int(s.parent.Fd()), s.name, unix.AT_REMOVEDIR); err != nil {
		return fmt.Errorf("command failure channel cleanup: %w", err)
	}
	s.closed = true
	return errors.Join(s.directory.Close(), s.parent.Close())
}
