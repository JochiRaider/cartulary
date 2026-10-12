//go:build linux

package harnesssession

import (
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"golang.org/x/sys/unix"
	"path/filepath"
	"slices"
)

// AllocateRecoveryOutput is native-owner-only, distinct for each physical
// operation. Original resource identity stays in the lease, not this path.
func (r *Runtime) AllocateRecoveryOutput() (string, error) {
	if !r.owned {
		return "", harnesslocal.ErrUnsafe
	}
	output, err := r.PrivatePath("recovery-output", identifier(), r.RunID)
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
	if output == original || !within(filepath.Join(root, "detail", "recovery-output"), output) || filepath.Base(output) != runID {
		return harnesslocal.ErrUnsafe
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
	if err := r.mutable(); err != nil {
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
