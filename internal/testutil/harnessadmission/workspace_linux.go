//go:build linux

package harnessadmission

import (
	"errors"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"strings"
)

// InheritWorkspace preserves the outer Make wrapper's shared descriptor 9. It
// neither acquires nor unlocks the lease. Existing private descriptors 3–8 stay
// in place, and temporary parent duplicates are close-on-exec.
func InheritWorkspace(cmd *exec.Cmd) (func(), error) {
	close := func() {}
	bytes, err := os.ReadFile("/proc/self/fdinfo/9")
	if err != nil || !strings.Contains(string(bytes), "FLOCK") || !strings.Contains(string(bytes), "READ") {
		return close, nil
	}
	if len(cmd.ExtraFiles) > 6 {
		return close, errors.New("reserved workspace descriptor is occupied")
	}
	fd, err := unix.FcntlInt(9, unix.F_DUPFD_CLOEXEC, 0)
	if err != nil {
		return close, err
	}
	admission := os.NewFile(uintptr(fd), "workspace-admission")
	null, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		admission.Close()
		return close, err
	}
	for len(cmd.ExtraFiles) < 6 {
		cmd.ExtraFiles = append(cmd.ExtraFiles, null)
	}
	cmd.ExtraFiles = append(cmd.ExtraFiles, admission)
	return func() { admission.Close(); null.Close() }, nil
}
