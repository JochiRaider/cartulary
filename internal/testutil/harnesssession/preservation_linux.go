//go:build linux

package harnesssession

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices/startupproof"
	"golang.org/x/sys/unix"
)

var regexpDigits = regexp.MustCompile(`^[0-9]+$`)

// PreserveRecovery keeps only exact resource-owner proof and its containing
// private directories. It never promotes credentials or logs into retained data.
func (r *Runtime) PreserveRecovery() error {
	unlock, err := r.lock()
	if err != nil {
		return err
	}
	defer unlock()
	if !r.owned {
		return harnesslocal.ErrUnsafe
	}
	// Producer authority and fencing do not depend on the resource journal.
	// Corrupt resources cannot retain safely quiesced disposable detail.
	if err := r.fenceDetails(); err != nil {
		return err
	}
	purgeErr := errors.Join(r.purgeDetail("ordinary"), r.purgeRecoveryDetails())
	marker, err := harnesslocal.PrivateDirectory(filepath.Join(r.Root, "recovery"))
	if err != nil {
		return err
	}
	marker.Close()
	resources, err := r.resources(4096)
	if err != nil {
		return errors.Join(err, purgeErr)
	}
	if purgeErr != nil {
		return AtCleanupStage("recovery_purge", purgeErr)
	}
	if len(resources) == 0 {
		// Review terminal transactions may retain scanner authority after resources
		// have drained. Only a sealed runtime can enter this resource-free phase.
		data, err := harnesslocal.ReadFileAt(r.Directory, "finalizing", 16, false)
		if err != nil || string(data) != "sealed\n" {
			return ErrUnresolved
		}
	}
	protected := []string{}
	validationErrors := []error{purgeErr}
	for _, v := range resources {
		if p, e := v.Path(); e == nil {
			// A managed start journal is independent recovery authority even when
			// no ready lease could be published. This preserves uncertainty; it
			// does not release the resource or infer an absent container.
			journalFile := filepath.Join(filepath.Dir(p), startupproof.Filename)
			journal, je := startupproof.Read(journalFile)
			_, pe := harnesslocal.ReadFile(p, 1048576, false)
			journalOwnsMissing := v.Kind == "managed_suite" && errors.Is(pe, os.ErrNotExist) && je == nil && slices.Contains(v.RecoveryFiles, journalFile) && journal.Binding.RuntimeRoot == r.Root && journal.Binding.RuntimeLease == r.LeaseID && journal.Binding.RunRoot == r.RunRoot && journal.Binding.RunID == r.RunID && slices.Equal(journal.Binding.Dependencies, v.Dependencies)
			if !journalOwnsMissing {
				protected = append(protected, p)
			}
		}
		protected = append(protected, v.RecoveryFiles...)
	}
	for _, file := range protected {
		// Validate the entire secure path without accepting a directory or link as
		// proof. Invalid authority remains unresolved; independent detail is purged.
		if _, err := harnesslocal.ReadFile(file, 1048576, false); err != nil {
			validationErrors = append(validationErrors, AtCleanupStage("recovery_proof", err))
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
			if path == r.Root && slices.Contains([]string{"runtime-owner.json", "runtime-process.json", "recovery", "registered-secrets", "finalizing", "detail-authority.json", "recovery-attempts", "recovery-detail", "detail"}, name) {
				continue
			}
			child := filepath.Join(path, name)
			var info unix.Stat_t
			if err := unix.Fstatat(int(dir.Fd()), name, &info, unix.AT_SYMLINK_NOFOLLOW); err != nil {
				failures = append(failures, err)
				continue
			}
			if info.Uid != uint32(os.Getuid()) {
				failures = append(failures, harnesslocal.ErrUnsafe)
				continue
			}
			if slices.Contains(protected, child) && info.Mode&unix.S_IFMT != unix.S_IFDIR {
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
						e = RemoveOwnedTree(d, child)
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
	return errors.Join(append(validationErrors, AtCleanupStage("recovery_purge", purge(r.Root)))...)
}

// CleanupStaleRoots is bounded and conservative. Age alone is never deletion
// authority: live/unknown process proof and unresolved resources retain a root.
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
	removed := 0
	for _, name := range names {
		staging := strings.HasPrefix(name, ".creating-suite-")
		if !staging && !strings.HasPrefix(name, "suite-") {
			continue
		}
		if removed >= 64 {
			break
		}
		path := filepath.Join(base, name)
		d, err := harnesslocal.OpenDirectory(path, true)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return removed, err
		}
		inspect := func() error {
			data, err := harnesslocal.ReadFileAt(d, "runtime-owner.json", MaximumOwner, false)
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
					if e = RemoveOwnedTree(d, path); e == nil {
						removed++
					}
					return e
				}
			}
			if err != nil {
				return err
			}
			owner, err := rootOwner(data)
			if err != nil {
				return err
			}
			if staging {
				if e := RemoveOwnedTree(d, path); e != nil {
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
			proof, err := rootProcess(data)
			if err != nil {
				return err
			}
			alive, err := harnessadmission.Alive(proof)
			if err != nil {
				return err
			}
			if alive {
				return nil
			}
			// A dead owner does not prove that its admitted detail writers or
			// marked descendants stopped. Use the same fenced native boundary.
			runtime := &Runtime{Root: path, RunID: owner["run_id"].(string), LeaseID: owner["lease_id"].(string), Directory: d, owned: true, Repository: owner["repository"].(string), RunRoot: owner["run_root"].(string)}
			resources, err := runtime.resources(4096)
			if err != nil {
				return err
			}
			if len(resources) != 0 {
				return nil
			}
			// Resource settlement does not settle terminal publication or a
			// retained scan. PreserveRecovery deliberately keeps sealed roots
			// even with an empty journal so exact recovery retains that authority.
			sealed, err := harnesslocal.ReadFileAt(d, "finalizing", 16, false)
			if err == nil {
				if string(sealed) != "sealed\n" {
					return harnesslocal.ErrUnsafe
				}
				return nil
			}
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
			if err := runtime.settleRecoveryDetails(false); err != nil {
				return err
			}
			if err := runtime.fenceDetails(); err != nil {
				return err
			}
			if err := runtime.purgeDetail("ordinary"); err != nil {
				return err
			}
			if err = runtime.setRecoveryLocatorState("settling"); err != nil {
				return err
			}
			if err = RemoveOwnedTree(d, path); err == nil {
				removed++
				err = runtime.setRecoveryLocatorState("settled")
			}
			return err
		}
		if err = unix.Flock(int(d.Fd()), unix.LOCK_EX|unix.LOCK_NB); err == nil {
			err = inspect()
			_ = unix.Flock(int(d.Fd()), unix.LOCK_UN)
		}
		d.Close()
		if errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, ErrUnresolved) {
			continue
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return removed, err
		}
	}
	return removed, harnesslocal.CheckDirectory(dir, base)
}

// rootOwner validates a runtime root's private ownership proof. Invalid or
// foreign proof fails closed.
func rootOwner(data []byte) (map[string]any, error) {
	v, err := harnessschema.Decode(data)
	if err != nil || harnessschema.Validate("cartulary.harness_suite_runtime_owner.v2", v) != nil {
		return nil, harnesslocal.ErrUnsafe
	}
	owner := v.(map[string]any)
	if owner["owner_uid"] != float64(os.Getuid()) {
		return nil, harnesslocal.ErrUnsafe
	}
	return owner, nil
}

// rootProcess validates a runtime root's original owner process proof.
func rootProcess(data []byte) (harnessadmission.Proof, error) {
	var proof harnessadmission.Proof
	v, err := harnessschema.Decode(data)
	m, ok := v.(map[string]any)
	if err != nil || !ok || len(m) != 3 || json.Unmarshal(data, &proof) != nil || proof.PID < 1 || proof.Boot == "" || !regexpDigits.MatchString(proof.Start) {
		return harnessadmission.Proof{}, harnesslocal.ErrUnsafe
	}
	return proof, nil
}

// DeadRuntime is one retained runtime root whose original owner process is
// proven dead: the exact identity explicit recovery admits.
type DeadRuntime struct {
	Root, RunRoot, LeaseID string
}

// DeadRuntimes lists, in name order, the retained runtime roots bound to the
// canonical repository whose original owner process is proven dead, with no
// age requirement. It shares stale-root maintenance's enumeration bound and
// grants no cleanup authority: recovery still admits each root through its
// exact ownership proof. Busy and sealed roots, live or unproven owners, roots
// of other repositories and roots with invalid ownership records are omitted.
func DeadRuntimes(repository, scratch string) ([]DeadRuntime, error) {
	scratch, err := resolveScratch(scratch)
	if err != nil {
		return nil, err
	}
	base := filepath.Join(scratch, "suite-runtime")
	dir, err := harnesslocal.OpenDirectory(base, true)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	names, err := dir.Readdirnames(8193)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if len(names) > 8192 {
		return nil, harnesslocal.ErrUnsafe
	}
	slices.Sort(names)
	dead := []DeadRuntime{}
	for _, name := range names {
		if !strings.HasPrefix(name, "suite-") {
			continue
		}
		path := filepath.Join(base, name)
		d, err := harnesslocal.OpenDirectory(path, true)
		if err != nil {
			continue
		}
		inspect := func() (*DeadRuntime, error) {
			data, err := harnesslocal.ReadFileAt(d, "runtime-owner.json", MaximumOwner, false)
			if err != nil {
				return nil, err
			}
			owner, err := rootOwner(data)
			if err != nil || owner["repository"] != repository {
				return nil, err
			}
			if data, err = harnesslocal.ReadFileAt(d, "runtime-process.json", 4096, false); err != nil {
				return nil, err
			}
			proof, err := rootProcess(data)
			if err != nil {
				return nil, err
			}
			if alive, err := harnessadmission.Alive(proof); err != nil || alive {
				return nil, err
			}
			// A sealed root drained its resources before sealing; it retains only
			// terminal publication authority for exact explicit recovery.
			if _, err := harnesslocal.ReadFileAt(d, "finalizing", 16, false); !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
			return &DeadRuntime{Root: path, RunRoot: owner["run_root"].(string), LeaseID: owner["lease_id"].(string)}, nil
		}
		if unix.Flock(int(d.Fd()), unix.LOCK_EX|unix.LOCK_NB) == nil {
			if runtime, err := inspect(); err == nil && runtime != nil {
				dead = append(dead, *runtime)
			}
			_ = unix.Flock(int(d.Fd()), unix.LOCK_UN)
		}
		d.Close()
	}
	return dead, harnesslocal.CheckDirectory(dir, base)
}
