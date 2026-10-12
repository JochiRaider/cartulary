package harnesssession

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

const OwnerSchema = "cartulary.harness_suite_runtime_owner.v2"
const MaximumOwner = 1048576

type runtimeOwner struct {
	Schema     string   `json:"schema_id"`
	LeaseID    string   `json:"lease_id"`
	RunID      string   `json:"run_id"`
	UID        int      `json:"owner_uid"`
	Created    string   `json:"created_at"`
	Repository string   `json:"repository"`
	RunRoot    string   `json:"run_root"`
	Outputs    []string `json:"output_roots"`
}

func canonicalPath(p string) bool { return filepath.IsAbs(p) && filepath.Clean(p) == p }
func decodeOwner(data []byte) (runtimeOwner, error) {
	var owner runtimeOwner
	value, err := harnessschema.Decode(data)
	if err != nil || harnessschema.Validate(OwnerSchema, value) != nil {
		return owner, harnesslocal.ErrUnsafe
	}
	if err = json.Unmarshal(data, &owner); err != nil {
		return owner, err
	}
	if !canonicalPath(owner.Repository) || !canonicalPath(owner.RunRoot) {
		return owner, harnesslocal.ErrUnsafe
	}
	return owner, nil
}

// ReadBinding verifies scope at every native or physical adapter boundary.
// Output directories can only be added by the native owner; they never replace
// the retained root or grant a caller a different repository's authority.
func ReadBinding(repository, runRoot, root, runID, leaseID string) (string, error) {
	if !canonicalPath(repository) || !canonicalPath(runRoot) || !canonicalPath(root) {
		return "", harnesslocal.ErrUnsafe
	}
	data, err := harnesslocal.ReadFile(filepath.Join(root, "runtime-owner.json"), MaximumOwner, false)
	if err != nil {
		return "", err
	}
	owner, err := decodeOwner(data)
	if err != nil {
		return "", err
	}
	if owner.UID != os.Getuid() || owner.Repository != repository || owner.RunID != runID || owner.LeaseID != leaseID {
		return "", harnesslocal.ErrUnsafe
	}
	for _, output := range owner.Outputs {
		relative, err := filepath.Rel(root, output)
		if !canonicalPath(output) || err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
			return "", harnesslocal.ErrUnsafe
		}
	}
	if runRoot != owner.RunRoot && !slices.Contains(owner.Outputs, runRoot) {
		return "", harnesslocal.ErrUnsafe
	}
	return owner.RunRoot, nil
}
