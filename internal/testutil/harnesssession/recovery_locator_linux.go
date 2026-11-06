//go:build linux

package harnesssession

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

const recoveryLocatorSchema = "cartulary.harness_recovery_locator.v1"

// RecoveryLocator is private owner-issued location evidence, not permission to
// remove resources. OwnRuntime and each resource owner still admit cleanup.
type RecoveryLocator struct {
	Schema     string `json:"schema_id"`
	Repository string `json:"repository"`
	RunRoot    string `json:"run_root"`
	RunID      string `json:"run_id"`
	Root       string `json:"runtime_root"`
	LeaseID    string `json:"lease_id"`
	UID        int    `json:"owner_uid"`
	State      string `json:"state"`
}

func resolveScratch(scratch string) (string, error) {
	if scratch == "" {
		if configured, ok := os.LookupEnv("CARTULARY_HARNESS_SCRATCH_ROOT"); ok {
			scratch = configured
		} else {
			cache, configured := os.LookupEnv("CARTULARY_MACHINE_CACHE_DIR")
			if !configured {
				if xdg := os.Getenv("XDG_CACHE_HOME"); filepath.IsAbs(xdg) {
					cache = filepath.Join(xdg, "cartulary")
				} else if home := os.Getenv("HOME"); filepath.IsAbs(home) {
					cache = filepath.Join(home, ".cache", "cartulary")
				}
			}
			if !canonicalPath(cache) {
				return "", harnesslocal.ErrUnsafe
			}
			scratch = filepath.Join(cache, "harness-scratch")
		}
	}
	if !canonicalPath(scratch) {
		return "", harnesslocal.ErrUnsafe
	}
	return scratch, nil
}

func locatorName(repository, runRoot, runID string) string {
	return fmt.Sprintf("%x.json", sha256.Sum256([]byte(repository+"\x00"+runRoot+"\x00"+runID)))
}
func readLocator(directory, repository, runRoot, runID string) (RecoveryLocator, error) {
	var v RecoveryLocator
	data, err := harnesslocal.ReadFile(filepath.Join(directory, locatorName(repository, runRoot, runID)), 32768, false)
	if err != nil {
		return v, err
	}
	decoded, err := harnessschema.Decode(data)
	if err != nil || harnessschema.Validate(recoveryLocatorSchema, decoded) != nil || json.Unmarshal(data, &v) != nil {
		return v, harnesslocal.ErrUnsafe
	}
	if v.Repository != repository || v.RunRoot != runRoot || v.RunID != runID || v.UID != os.Getuid() || !canonicalPath(v.Root) || filepath.Dir(filepath.Dir(v.Root)) != filepath.Dir(directory) || filepath.Base(filepath.Dir(v.Root)) != "suite-runtime" {
		return v, harnesslocal.ErrUnsafe
	}
	return v, nil
}
func (r *Runtime) publishRecoveryLocator() error {
	directory := filepath.Join(filepath.Dir(filepath.Dir(r.Root)), "recovery-index")
	d, err := harnesslocal.PrivateDirectory(directory)
	if err != nil {
		return err
	}
	defer d.Close()
	v := RecoveryLocator{Schema: recoveryLocatorSchema, Repository: r.Repository, RunRoot: r.RunRoot, RunID: r.RunID, Root: r.Root, LeaseID: r.LeaseID, UID: os.Getuid(), State: "active"}
	return privateJSON(filepath.Join(directory, locatorName(r.Repository, r.RunRoot, r.RunID)), recoveryLocatorSchema, v, false)
}
func (r *Runtime) setRecoveryLocatorState(state string) error {
	directory := filepath.Join(filepath.Dir(filepath.Dir(r.Root)), "recovery-index")
	v, err := readLocator(directory, r.Repository, r.RunRoot, r.RunID)
	// Historical runtimes predate the index and retain their explicit-tuple path.
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if v.Root != r.Root || v.LeaseID != r.LeaseID {
		return harnesslocal.ErrUnsafe
	}
	v.State = state
	return privateJSON(filepath.Join(directory, locatorName(r.Repository, r.RunRoot, r.RunID)), recoveryLocatorSchema, v, true)
}

// ResolveRecovery performs one exact keyed read; it neither scans processes nor
// infers ownership from directory names. It never changes evidence or resources.
func ResolveRecovery(repository, runRoot, runID, scratch string) (RecoveryLocator, error) {
	var empty RecoveryLocator
	if !canonicalPath(repository) || !canonicalPath(runRoot) || !component.MatchString(runID) || runID == "." || runID == ".." {
		return empty, harnesslocal.ErrUnsafe
	}
	for _, bound := range []string{repository, runRoot} {
		real, err := filepath.EvalSymlinks(bound)
		if err != nil || real != bound {
			return empty, harnesslocal.ErrUnsafe
		}
	}
	scratch, err := resolveScratch(scratch)
	if err != nil {
		return empty, err
	}
	if within(repository, scratch) || within(runRoot, scratch) {
		return empty, harnesslocal.ErrUnsafe
	}
	v, err := readLocator(filepath.Join(scratch, "recovery-index"), repository, runRoot, runID)
	if err != nil {
		return empty, err
	}
	d, err := harnesslocal.OpenDirectory(v.Root, true)
	if errors.Is(err, os.ErrNotExist) && (v.State == "settling" || v.State == "settled") {
		v.State = "settled"
		return v, nil
	}
	if err != nil {
		return empty, err
	}
	d.Close()
	if v.State == "settled" {
		return empty, harnesslocal.ErrUnsafe
	}
	if _, err := ReadBinding(repository, runRoot, v.Root, runID, v.LeaseID); err != nil {
		return empty, err
	}
	return v, nil
}
