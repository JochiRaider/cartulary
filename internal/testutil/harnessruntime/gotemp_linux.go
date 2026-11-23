//go:build linux

package harnessruntime

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"golang.org/x/sys/unix"
)

// Go temporary work is invocation-scoped. Go build work directories and test
// TempDir roots land in GOTMPDIR, and a child that exits without running its
// cleanups (os.Exit, a kill) would otherwise leak into the shared machine path.
const (
	goTempPrefix    = "invocation-"
	goTempStaging   = ".creating-" + goTempPrefix
	goTempOwnerFile = "invocation-owner.json"
	goTempOwnerMax  = 4096
)

// newGoTemp reclaims sibling invocation directories whose recorded owner is
// proven dead, then creates this invocation's directory beneath base with the
// current process recorded as owner before any unit can see it.
func newGoTemp(base string) (string, error) {
	if !filepath.IsAbs(base) || filepath.Clean(base) != base {
		return "", errors.New("GO_TMP_DIR must be an absolute clean path")
	}
	if err := os.MkdirAll(base, 0700); err != nil {
		return "", err
	}
	reclaimGoTemp(base)
	owner, err := harnessadmission.Identity(os.Getpid())
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(owner)
	if err != nil {
		return "", err
	}
	staging, err := os.MkdirTemp(base, goTempStaging)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(staging, goTempOwnerFile), append(data, '\n'), 0600); err != nil {
		return "", errors.Join(err, os.RemoveAll(staging))
	}
	published := filepath.Join(base, strings.TrimPrefix(filepath.Base(staging), ".creating-"))
	if err := unix.Renameat2(unix.AT_FDCWD, staging, unix.AT_FDCWD, published, unix.RENAME_NOREPLACE); err != nil {
		return "", errors.Join(err, os.RemoveAll(staging))
	}
	return published, nil
}

// reclaimGoTemp removes only invocation directories whose recorded owner is
// proven dead. Unreadable, unproven, or live entries and every other entry are
// left untouched; reclamation never blocks a new invocation.
func reclaimGoTemp(base string) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return
	}
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() || !(strings.HasPrefix(name, goTempPrefix) || strings.HasPrefix(name, goTempStaging)) {
			continue
		}
		directory := filepath.Join(base, name)
		owner, ok := readGoTempOwner(directory)
		if !ok {
			continue
		}
		if alive, err := harnessadmission.Alive(owner); err != nil || alive {
			continue
		}
		_ = removeGoTemp(directory)
	}
}

func readGoTempOwner(directory string) (harnessadmission.Proof, bool) {
	var owner harnessadmission.Proof
	path := filepath.Join(directory, goTempOwnerFile)
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > goTempOwnerMax {
		return owner, false
	}
	data, err := os.ReadFile(path)
	if err != nil || json.Unmarshal(data, &owner) != nil || owner.PID <= 0 {
		return owner, false
	}
	return owner, true
}

// removeGoTemp restores owner permissions on real directories (a test may
// leave read-only trees) and removes the tree without following links.
func removeGoTemp(directory string) error {
	info, err := os.Lstat(directory)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return errors.New("go temporary work path is not a directory")
	}
	_ = filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err == nil && entry.IsDir() {
			_ = os.Chmod(path, 0700)
		}
		return nil
	})
	return os.RemoveAll(directory)
}
