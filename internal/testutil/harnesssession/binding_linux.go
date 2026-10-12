//go:build linux

package harnesssession

import (
	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// AllocateRecoveryOutput is native-owner-only, distinct for each physical
// operation. Original resource identity stays in the lease, not this path.
func (r *Runtime) AllocateRecoveryOutput() (string, error) {
	if !r.owned || r.attempt == "" {
		return "", harnesslocal.ErrUnsafe
	}
	output, err := r.RecoveryDetailPath("output", identifier(), r.RunID)
	if err != nil {
		return "", err
	}
	dir, err := harnesslocal.PrivateDirectory(output)
	if err != nil {
		return "", err
	}
	dir.Close()
	if err = r.AuthorizeOutputRoot(output); err != nil {
		return "", err
	}
	return output, nil
}

// ClaimRecoveryOutput rejects replay and concurrent writers before physical
// cleanup begins. Ambient environment alone cannot authorize an output root.
func ClaimRecoveryOutput(repository, output, root, runID, leaseID string) error {
	authority, err := harnesslocal.OpenDirectory(root, true)
	if err != nil {
		return err
	}
	defer authority.Close()
	if err = unix.Flock(int(authority.Fd()), unix.LOCK_EX); err != nil {
		return err
	}
	defer unix.Flock(int(authority.Fd()), unix.LOCK_UN)
	original, err := ReadBinding(repository, output, root, runID, leaseID)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(filepath.Join(root, "recovery-detail"), output)
	parts := strings.Split(relative, string(filepath.Separator))
	if err != nil || len(parts) != 4 || !uuidPattern.MatchString(parts[0]) || parts[1] != "output" || !uuidPattern.MatchString(parts[2]) || parts[3] != runID || output == original {
		return harnesslocal.ErrUnsafe
	}
	r := &Runtime{Root: root, RunID: runID, LeaseID: leaseID}
	v, err := r.readDetail(parts[0])
	if err != nil {
		return err
	}
	actor, err := harnessadmission.Identity(os.Getpid())
	if err != nil {
		return err
	}
	if err := attemptCaller(v, actor); err != nil {
		return err
	}
	dir, err := harnesslocal.OpenDirectory(output, true)
	if err != nil {
		return err
	}
	defer dir.Close()
	if err = unix.Flock(int(dir.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		return err
	}
	defer unix.Flock(int(dir.Fd()), unix.LOCK_UN)
	names, err := dir.Readdirnames(-1)
	if err != nil {
		return err
	}
	if len(names) != 0 {
		return harnesslocal.ErrUnsafe
	}
	data, err := harnesslocal.ReadFileAt(authority, "runtime-owner.json", MaximumOwner, false)
	if err != nil {
		return err
	}
	owner, err := decodeOwner(data)
	if err != nil {
		return err
	}
	if slices.Contains(owner.ClaimedOutputs, output) {
		return harnesslocal.ErrUnsafe
	}
	owner.ClaimedOutputs = append(owner.ClaimedOutputs, output)
	slices.Sort(owner.ClaimedOutputs)
	data, err = validateJSON(OwnerSchema, owner)
	if err != nil {
		return err
	}
	return harnesslocal.ReplaceFileAt(authority, "runtime-owner.json", append(data, '\n'))
}

// AuthorizeOutputRoot binds one exact private physical lifecycle output. It is
// deliberately absent from the producer wire protocol.
func (r *Runtime) AuthorizeOutputRoot(output string) error {
	if !r.owned || !canonicalPath(output) {
		return harnesslocal.ErrUnsafe
	}
	if output == r.RunRoot {
		return nil
	}
	if output == r.Root || !within(r.Root, output) || within(filepath.Join(r.Root, "recovery"), output) {
		return harnesslocal.ErrUnsafe
	}
	unlock, err := r.lock()
	if err != nil {
		return err
	}
	defer unlock()
	if err := r.productionMutable(); err != nil {
		return err
	}
	data, err := harnesslocal.ReadFileAt(r.Directory, "runtime-owner.json", MaximumOwner, false)
	if err != nil {
		return err
	}
	owner, err := decodeOwner(data)
	if err != nil {
		return err
	}
	if owner.Repository != r.Repository || owner.RunRoot != r.RunRoot || owner.LeaseID != r.LeaseID || owner.RunID != r.RunID {
		return harnesslocal.ErrUnsafe
	}
	if slices.Contains(owner.Outputs, output) {
		return nil
	}
	owner.Outputs = append(owner.Outputs, output)
	slices.Sort(owner.Outputs)
	data, err = validateJSON(OwnerSchema, owner)
	if err != nil {
		return err
	}
	return harnesslocal.ReplaceFileAt(r.Directory, "runtime-owner.json", append(data, '\n'))
}
