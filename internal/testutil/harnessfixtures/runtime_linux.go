//go:build linux

package harnessfixtures

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"golang.org/x/sys/unix"
)

type Runtime struct {
	Root, RunID, LeaseID string
	Directory            *os.File
	parent               *os.File
	owned                bool
	mu                   sync.Mutex
	secrets              map[string]bool
	closed               bool
}

var safeName = regexp.MustCompile(`[^A-Za-z0-9_.-]+`)
var component = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
var secretName = regexp.MustCompile(`(?:PASSWORD|SECRET|TOKEN|COOKIE|DSN|ACCESS_KEY|KEY_RING|PRIVATE_KEY)`)

func identifier() string {
	var v [16]byte
	if _, err := rand.Read(v[:]); err != nil {
		panic(err)
	}
	v[6] = (v[6] & 15) | 64
	v[8] = (v[8] & 63) | 128
	return fmt.Sprintf("%x-%x-%x-%x-%x", v[:4], v[4:6], v[6:8], v[8:10], v[10:])
}
func within(base, name string) bool {
	relative, err := filepath.Rel(base, name)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative)
}
func validateJSON(schema string, value any) ([]byte, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	v, err := harnessschema.Decode(bytes)
	if err == nil {
		err = harnessschema.Validate(schema, v)
	}
	return bytes, err
}
func privateJSON(name, schema string, value any, replace bool) error {
	data, err := validateJSON(schema, value)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if replace {
		return harnesslocal.ReplaceFile(name, data)
	}
	return harnesslocal.PublishNew(name, data)
}
func NewRuntime(repository, runRoot, runID, scratch string) (result *Runtime, err error) {
	if !component.MatchString(runID) || len(runID) > 96 {
		return nil, harnesslocal.ErrUnsafe
	}
	if scratch == "" {
		scratch = filepath.Join(os.TempDir(), "cartulary-harness-scratch")
	}
	for _, name := range []string{repository, runRoot, scratch} {
		if !filepath.IsAbs(name) {
			return nil, harnesslocal.ErrUnsafe
		}
	}
	if within(repository, scratch) || within(runRoot, scratch) {
		return nil, harnesslocal.ErrUnsafe
	}
	base := filepath.Join(scratch, "suite-runtime")
	parent, err := harnesslocal.PrivateDirectory(base)
	if err != nil {
		return nil, err
	}
	if _, err := CleanupStaleRoots(base, time.Now()); err != nil {
		parent.Close()
		return nil, err
	}
	name := ".creating-suite-" + runID + "-" + identifier()
	if err := unix.Mkdirat(int(parent.Fd()), name, 0700); err != nil {
		parent.Close()
		return nil, err
	}
	root := filepath.Join(base, name)
	directory, err := harnesslocal.OpenDirectory(root, true)
	if err != nil {
		parent.Close()
		return nil, err
	}
	result = &Runtime{Root: root, RunID: runID, LeaseID: identifier(), Directory: directory, parent: parent, owned: true, secrets: map[string]bool{}}
	// Incomplete initialization remains under the private root, never a retained
	// path. Only this exact created directory is eligible for rollback.
	defer func() {
		if err != nil {
			if harnesslocal.CheckDirectory(directory, root) == nil {
				_ = os.RemoveAll(fmt.Sprintf("/proc/self/fd/%d/%s", parent.Fd(), name))
			}
			directory.Close()
			parent.Close()
		}
	}()
	owner := map[string]any{"schema_id": "cartulary.harness_suite_runtime_owner.v1", "lease_id": result.LeaseID, "run_id": runID, "owner_uid": os.Getuid(), "created_at": time.Now().UTC().Format(time.RFC3339Nano)}
	if err = privateJSON(filepath.Join(root, "runtime-owner.json"), "cartulary.harness_suite_runtime_owner.v1", owner, false); err != nil {
		return nil, err
	}
	publishedName := strings.TrimPrefix(name, ".creating-")
	if err = unix.Renameat2(int(parent.Fd()), name, int(parent.Fd()), publishedName, unix.RENAME_NOREPLACE); err != nil {
		return nil, err
	}
	name = publishedName
	root = filepath.Join(base, name)
	result.Root = root
	// Preserve the pinned inode while updating its diagnostic pathname after
	// staging publication. Downstream private channels validate this exact path.
	duplicate, e := unix.FcntlInt(directory.Fd(), unix.F_DUPFD_CLOEXEC, 0)
	if e != nil {
		return nil, e
	}
	directory.Close()
	directory = os.NewFile(uintptr(duplicate), root)
	result.Directory = directory
	proof, e := harnessadmission.Identity(os.Getpid())
	if e != nil {
		return nil, e
	}
	data, e := json.Marshal(proof)
	if e != nil {
		return nil, e
	}
	if err = harnesslocal.PublishNew(filepath.Join(root, "runtime-process.json"), append(data, '\n')); err != nil {
		return nil, err
	}
	if err = harnesslocal.CheckDirectory(parent, base); err != nil {
		return nil, err
	}
	return result, nil
}
func BorrowRuntime(repository, runRoot, root, runID, leaseID string) (*Runtime, error) {
	if !filepath.IsAbs(repository) || !filepath.IsAbs(runRoot) || !filepath.IsAbs(root) || within(repository, root) || within(runRoot, root) {
		return nil, harnesslocal.ErrUnsafe
	}
	directory, err := harnesslocal.OpenDirectory(root, true)
	if err != nil {
		return nil, err
	}
	fail := func() (*Runtime, error) { directory.Close(); return nil, harnesslocal.ErrUnsafe }
	data, err := harnesslocal.ReadFileAt(directory, "runtime-owner.json", 4096, false)
	if err != nil {
		return fail()
	}
	value, err := harnessschema.Decode(data)
	if err != nil || harnessschema.Validate("cartulary.harness_suite_runtime_owner.v1", value) != nil {
		return fail()
	}
	owner := value.(map[string]any)
	if owner["lease_id"] != leaseID || owner["run_id"] != runID || owner["owner_uid"] != float64(os.Getuid()) {
		return fail()
	}
	if harnesslocal.CheckDirectory(directory, root) != nil {
		return fail()
	}
	return &Runtime{Root: root, RunID: runID, LeaseID: leaseID, Directory: directory, secrets: map[string]bool{}}, nil
}
func (r *Runtime) Environment() map[string]string {
	return map[string]string{"CARTULARY_HARNESS_SUITE_RUNTIME_ROOT": r.Root, "CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID": r.LeaseID, "CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID": r.RunID}
}
func (r *Runtime) PrivatePath(parts ...string) (string, error) {
	if len(parts) == 0 {
		return "", harnesslocal.ErrUnsafe
	}
	for _, part := range parts {
		if !component.MatchString(part) || part == "." || part == ".." {
			return "", harnesslocal.ErrUnsafe
		}
	}
	return filepath.Join(append([]string{r.Root}, parts...)...), nil
}
func (r *Runtime) RegisterSecret(value string) {
	if len(value) < 8 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.secrets[value] = true
}
func (r *Runtime) RegisterEnvironment(environment map[string]string) {
	for k, v := range environment {
		if secretName.MatchString(k) && !strings.Contains(k, "TOKENS") {
			r.RegisterSecret(v)
		}
	}
}
func (r *Runtime) ForbiddenValues() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Sorted(maps.Keys(r.secrets))
}

type Resource struct {
	Kind    string          `json:"kind"`
	Target  json.RawMessage `json:"target"`
	State   string          `json:"state"`
	LeaseID string          `json:"lease_id"`
	RunID   string          `json:"run_id"`
}

func (r Resource) Path() (string, error) {
	var path string
	err := json.Unmarshal(r.Target, &path)
	return path, err
}
func resourceName(kind string, target json.RawMessage) (string, error) {
	var compact bytes.Buffer
	if err := json.Compact(&compact, target); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%x.json", kind, sha256.Sum256(compact.Bytes())), nil
}
func (r *Runtime) checkResource(v Resource) error {
	if v.LeaseID != r.LeaseID || v.RunID != r.RunID || v.State != "pending" && v.State != "acquired" && v.State != "released" {
		return harnesslocal.ErrUnsafe
	}
	switch v.Kind {
	case "managed_suite", "borrowed_suite", "browser_stack", "object_store_proxy":
		target, err := v.Path()
		if err != nil || target == r.Root || !within(r.Root, target) {
			return harnesslocal.ErrUnsafe
		}
		relative, _ := filepath.Rel(r.Root, target)
		for _, part := range strings.Split(relative, string(filepath.Separator)) {
			if part == "recovery" || part == "." || part == ".." {
				return harnesslocal.ErrUnsafe
			}
		}
	case "preparation_process", "browser_process", "helper_process":
		value, err := harnessschema.Decode(v.Target)
		if err != nil {
			return err
		}
		o, ok := value.(map[string]any)
		if !ok || len(o) != 4 {
			return harnesslocal.ErrUnsafe
		}
		var proof struct {
			Boot  string `json:"boot"`
			PID   int    `json:"pid"`
			Start string `json:"start"`
			Group *bool  `json:"group"`
		}
		if json.Unmarshal(v.Target, &proof) != nil || proof.Boot == "" || proof.PID < 2 || proof.Group == nil || !regexp.MustCompile(`^[0-9]+$`).MatchString(proof.Start) {
			return harnesslocal.ErrUnsafe
		}
	case "browser_renderer", "diagnostic_scope":
		// These private proofs are interpreted only by their existing owners. The
		// runtime preserves them and refuses root removal while either remains.
		if value, err := harnessschema.Decode(v.Target); err != nil || value == nil {
			return harnesslocal.ErrUnsafe
		}
	default:
		return harnesslocal.ErrUnsafe
	}
	return nil
}
func (r *Runtime) RecordResource(kind string, target any, state string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed || harnesslocal.CheckDirectory(r.Directory, r.Root) != nil {
		return harnesslocal.ErrUnsafe
	}
	raw, err := json.Marshal(target)
	if err != nil {
		return err
	}
	v := Resource{Kind: kind, Target: raw, State: state, LeaseID: r.LeaseID, RunID: r.RunID}
	if err := r.checkResource(v); err != nil {
		return err
	}
	directory, err := harnesslocal.PrivateDirectory(filepath.Join(r.Root, "recovery"))
	if err != nil {
		return err
	}
	defer directory.Close()
	name, err := resourceName(kind, raw)
	if err != nil {
		return err
	}
	if state == "released" {
		if _, err := harnesslocal.ReadFileAt(directory, name, 1048576, false); errors.Is(err, os.ErrNotExist) {
			return nil
		} else if err != nil {
			return err
		}
		return unix.Unlinkat(int(directory.Fd()), name, 0)
	}
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return harnesslocal.ReplaceFileAt(directory, name, append(data, '\n'))
}
func (r *Runtime) Resources(maximum int) ([]Resource, error) {
	if harnesslocal.CheckDirectory(r.Directory, r.Root) != nil {
		return nil, harnesslocal.ErrUnsafe
	}
	directory, err := harnesslocal.OpenDirectory(filepath.Join(r.Root, "recovery"), true)
	if errors.Is(err, os.ErrNotExist) {
		return []Resource{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer directory.Close()
	names, err := directory.Readdirnames(maximum + 1)
	if err != nil && !errors.Is(err, io.EOF) && len(names) == 0 {
		return nil, err
	}
	if len(names) > maximum {
		return nil, harnesslocal.ErrUnsafe
	}
	slices.Sort(names)
	resources := []Resource{}
	for _, name := range names {
		data, err := harnesslocal.ReadFileAt(directory, name, 1048576, false)
		if err != nil {
			return nil, err
		}
		value, err := harnessschema.Decode(data)
		if err != nil {
			return nil, err
		}
		o, ok := value.(map[string]any)
		if !ok || len(o) != 5 {
			return nil, harnesslocal.ErrUnsafe
		}
		var v Resource
		if json.Unmarshal(data, &v) != nil || r.checkResource(v) != nil || v.State == "released" {
			return nil, harnesslocal.ErrUnsafe
		}
		expected, err := resourceName(v.Kind, v.Target)
		if err != nil || expected != name {
			return nil, harnesslocal.ErrUnsafe
		}
		resources = append(resources, v)
	}
	return resources, nil
}
func (r *Runtime) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil
	}
	if !r.owned {
		r.closed = true
		return r.Directory.Close()
	}
	resources, err := r.Resources(4096)
	if err != nil {
		return err
	}
	if len(resources) > 0 {
		return CleanupFailure(errors.New("runtime retains unresolved resource ownership"))
	}
	if harnesslocal.CheckDirectory(r.parent, filepath.Dir(r.Root)) != nil || harnesslocal.CheckDirectory(r.Directory, r.Root) != nil {
		return harnesslocal.ErrUnsafe
	}
	if err := removeOwnedTree(r.Directory, r.Root); err != nil {
		return err
	}
	r.closed = true
	r.secrets = nil
	return errors.Join(r.Directory.Close(), r.parent.Close())
}

// ReleaseHandles ends a supervisor's access after either successful removal or
// preserved recovery. It does not delete proof or assert resource cleanup.
func (r *Runtime) ReleaseHandles() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil
	}
	r.closed = true
	r.secrets = nil
	var parentErr error
	if r.parent != nil {
		parentErr = r.parent.Close()
	}
	return errors.Join(r.Directory.Close(), parentErr)
}
