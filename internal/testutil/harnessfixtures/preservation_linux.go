//go:build linux

package harnessfixtures

import (
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

var regexpDigits = regexp.MustCompile(`^[0-9]+$`)

// PreserveRecovery keeps only exact resource-owner proof and its containing
// private directories. It never promotes credentials or logs into retained data.
func (r *Runtime) PreserveRecovery() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed || !r.owned || harnesslocal.CheckDirectory(r.Directory, r.Root) != nil {
		return harnesslocal.ErrUnsafe
	}
	marker, err := harnesslocal.PrivateDirectory(filepath.Join(r.Root, "recovery"))
	if err != nil {
		return err
	}
	marker.Close()
	resources, err := r.Resources(4096)
	if err != nil {
		return err
	}
	if len(resources) == 0 {
		return CleanupFailure(errors.New("unresolved cleanup has no recovery proof"))
	}
	protected := []string{}
	for _, v := range resources {
		if p, e := v.Path(); e == nil {
			protected = append(protected, filepath.Dir(p))
		}
	}
	var purge func(string) error
	purge = func(path string) error {
		dir, err := harnesslocal.OpenDirectory(path, true)
		if err != nil {
			return err
		}
		defer dir.Close()
		names, err := dir.Readdirnames(-1)
		if err != nil {
			return err
		}
		slices.Sort(names)
		failures := []error{}
		for _, name := range names {
			if path == r.Root && slices.Contains([]string{"runtime-owner.json", "runtime-process.json", "recovery"}, name) {
				continue
			}
			child := filepath.Join(path, name)
			if slices.Contains(protected, child) {
				continue
			}
			var info unix.Stat_t
			if err := unix.Fstatat(int(dir.Fd()), name, &info, unix.AT_SYMLINK_NOFOLLOW); err != nil {
				failures = append(failures, err)
				continue
			}
			if info.Uid != uint32(os.Getuid()) {
				failures = append(failures, harnesslocal.ErrUnsafe)
				continue
			}
			if info.Mode&unix.S_IFMT == unix.S_IFDIR {
				contains := false
				for _, p := range protected {
					contains = contains || within(child, p)
				}
				if contains {
					failures = append(failures, purge(child))
				} else {
					d, e := harnesslocal.OpenDirectory(child, true)
					if e == nil {
						e = removeOwnedTree(d, child)
						d.Close()
					}
					failures = append(failures, e)
				}
			} else {
				failures = append(failures, unix.Unlinkat(int(dir.Fd()), name, 0))
			}
		}
		return errors.Join(failures...)
	}
	return purge(r.Root)
}

// CleanupStaleRoots is bounded and conservative. Age alone is never deletion
// authority: live/unknown process proof and any recovery marker retain a root.
func CleanupStaleRoots(base string, now time.Time) (int, error) {
	dir, err := harnesslocal.OpenDirectory(base, true)
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	defer dir.Close()
	names, err := dir.Readdirnames(8193)
	if err != nil && err != io.EOF {
		return 0, err
	}
	if len(names) > 8192 {
		return 0, harnesslocal.ErrUnsafe
	}
	slices.Sort(names)
	scanned, removed := 0, 0
	for _, name := range names {
		staging := strings.HasPrefix(name, ".creating-suite-")
		if !staging && !strings.HasPrefix(name, "suite-") {
			continue
		}
		if scanned >= 64 {
			break
		}
		scanned++
		path := filepath.Join(base, name)
		d, err := harnesslocal.OpenDirectory(path, true)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return removed, err
		}
		inspect := func() error {
			data, err := harnesslocal.ReadFileAt(d, "runtime-owner.json", 4096, false)
			if staging {
				info, e := d.Stat()
				if e != nil {
					return e
				}
				if now.Sub(info.ModTime()) < 24*time.Hour {
					return nil
				}
				entries, e := d.Readdirnames(2)
				if e != nil && e != io.EOF {
					return e
				}
				if len(entries) > 1 || len(entries) == 1 && entries[0] != "runtime-owner.json" {
					return harnesslocal.ErrUnsafe
				}
				if len(entries) == 0 {
					if e = removeOwnedTree(d, path); e == nil {
						removed++
					}
					return e
				}
			}
			if err != nil {
				return err
			}
			v, err := harnessschema.Decode(data)
			if err != nil || harnessschema.Validate("cartulary.harness_suite_runtime_owner.v1", v) != nil {
				return harnesslocal.ErrUnsafe
			}
			owner := v.(map[string]any)
			if owner["owner_uid"] != float64(os.Getuid()) {
				return harnesslocal.ErrUnsafe
			}
			if staging {
				if e := removeOwnedTree(d, path); e != nil {
					return e
				}
				removed++
				return nil
			}
			created, err := time.Parse(time.RFC3339Nano, owner["created_at"].(string))
			if err != nil {
				return err
			}
			if now.Sub(created) < 24*time.Hour {
				return nil
			}
			data, err = harnesslocal.ReadFileAt(d, "runtime-process.json", 4096, false)
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			if err != nil {
				return err
			}
			v, err = harnessschema.Decode(data)
			m, ok := v.(map[string]any)
			if err != nil || !ok || len(m) != 3 {
				return harnesslocal.ErrUnsafe
			}
			var proof harnessadmission.Proof
			if json.Unmarshal(data, &proof) != nil || proof.PID < 1 || proof.Boot == "" || !regexpDigits.MatchString(proof.Start) {
				return harnesslocal.ErrUnsafe
			}
			alive, err := harnessadmission.Alive(proof)
			if err != nil {
				return err
			}
			if alive {
				return nil
			}
			var recovery unix.Stat_t
			err = unix.Fstatat(int(d.Fd()), "recovery", &recovery, unix.AT_SYMLINK_NOFOLLOW)
			if err == nil {
				return nil
			}
			if err != unix.ENOENT {
				return err
			}
			if err = removeOwnedTree(d, path); err == nil {
				removed++
			}
			return err
		}
		err = inspect()
		d.Close()
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return removed, err
		}
	}
	return removed, harnesslocal.CheckDirectory(dir, base)
}
