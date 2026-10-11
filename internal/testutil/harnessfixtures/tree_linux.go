//go:build linux

package harnessfixtures

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"golang.org/x/sys/unix"
)

// Move the exact owned inode out of its public locator before recursive removal.
// Replacement detection preserves the foreign inode, including a replacement
// that wins the rename race. Symlinks are never followed during the deletion.
func removeOwnedTree(directory *os.File, name string) error {
	if err := harnesslocal.CheckDirectory(directory, name); err != nil {
		return err
	}
	parent, err := harnesslocal.OpenDirectory(filepath.Dir(name), true)
	if err != nil {
		return err
	}
	defer parent.Close()
	leaf := filepath.Base(name)
	quarantine := ".deleting-" + identifier()
	if err := unix.Renameat2(int(parent.Fd()), leaf, int(parent.Fd()), quarantine, unix.RENAME_NOREPLACE); err != nil {
		return err
	}
	private := fmt.Sprintf("/proc/self/fd/%d/%s", parent.Fd(), quarantine)
	original, err := directory.Stat()
	if err != nil {
		return err
	}
	current, err := os.Lstat(private)
	if err != nil || !os.SameFile(original, current) {
		restore := unix.Renameat2(int(parent.Fd()), quarantine, int(parent.Fd()), leaf, unix.RENAME_NOREPLACE)
		return errors.Join(harnesslocal.ErrUnsafe, err, restore)
	}
	return os.RemoveAll(private)
}
