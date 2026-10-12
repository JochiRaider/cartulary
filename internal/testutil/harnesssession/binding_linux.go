//go:build linux

package harnesssession

import (
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"path/filepath"
	"slices"
)

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
