//go:build linux

package harnesssession

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	Repository, RunRoot  string
	Directory            *os.File
	parent               *os.File
	owned                bool
	mu                   sync.Mutex
	closed               bool
	cleanupMu            sync.Mutex
	cleanupAttempts      map[string]*cleanupAttempt
	actor                harnessadmission.Proof
	attempt              string
}

var component = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
var uuidPattern = regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`)
var secretName = regexp.MustCompile(`(?i)(?:PASSWORD|SECRET|TOKEN(?:[^S]|$)|COOKIE|DSN|ACCESS_KEY|KEY_RING|PRIVATE_KEY)`)

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
func NewRuntime(repository, runRoot, runID, scratch string) (*Runtime, error) {
	owner, err := harnessadmission.Identity(os.Getpid())
	if err != nil {
		return nil, err
	}
	return newRuntime(repository, runRoot, runID, scratch, owner)
}
func newRuntime(repository, runRoot, runID, scratch string, ownerProcess harnessadmission.Proof) (result *Runtime, err error) {
	if !component.MatchString(runID) || runID == "." || runID == ".." || len(runID) > 96 {
		return nil, harnesslocal.ErrUnsafe
	}
	if scratch == "" {
		scratch = filepath.Join(os.TempDir(), "cartulary-harness-scratch")
	}
	for _, name := range []string{repository, runRoot, scratch} {
		if !canonicalPath(name) {
			return nil, harnesslocal.ErrUnsafe
		}
	}
	for _, bound := range []string{repository, runRoot} {
		canonical, err := filepath.EvalSymlinks(bound)
		if err != nil || canonical != bound {
			return nil, harnesslocal.ErrUnsafe
		}
		dir, err := harnesslocal.OpenDirectory(bound, false)
		if err != nil {
			return nil, err
		}
		dir.Close()
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
	result = &Runtime{Root: root, RunID: runID, LeaseID: identifier(), Repository: repository, RunRoot: runRoot, Directory: directory, parent: parent, owned: true, actor: ownerProcess}
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
	owner := runtimeOwner{Schema: OwnerSchema, LeaseID: result.LeaseID, RunID: runID, UID: os.Getuid(), Created: time.Now().UTC().Format(time.RFC3339Nano), Repository: repository, RunRoot: runRoot, Outputs: []string{}}

	if err = privateJSON(filepath.Join(root, "runtime-owner.json"), "cartulary.harness_suite_runtime_owner.v2", owner, false); err != nil {
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
	data, e := json.Marshal(ownerProcess)
	if e != nil {
		return nil, e
	}
	if err = harnesslocal.PublishNew(filepath.Join(root, "runtime-process.json"), append(data, '\n')); err != nil {
		return nil, err
	}
	if _, err = result.createDetail("ordinary", ownerProcess); err != nil {
		return nil, err
	}
	if err = harnesslocal.CheckDirectory(parent, base); err != nil {
		return nil, err
	}
	return result, nil
}
func BorrowRuntime(repository, runRoot, root, runID, leaseID string) (*Runtime, error) {
	if !canonicalPath(repository) || !canonicalPath(runRoot) || !canonicalPath(root) || within(repository, root) || within(runRoot, root) {
		return nil, harnesslocal.ErrUnsafe
	}
	directory, err := harnesslocal.OpenDirectory(root, true)
	if err != nil {
		return nil, err
	}
	fail := func() (*Runtime, error) { directory.Close(); return nil, harnesslocal.ErrUnsafe }
	if _, err := ReadBinding(repository, runRoot, root, runID, leaseID); err != nil {
		return fail()
	}
	data, err := harnesslocal.ReadFileAt(directory, "runtime-owner.json", MaximumOwner, false)
	if err != nil {
		return fail()
	}
	owner, err := decodeOwner(data)
	if err != nil {
		return fail()
	}

	if harnesslocal.CheckDirectory(directory, root) != nil {
		return fail()
	}
	return &Runtime{Root: root, RunID: runID, LeaseID: leaseID, Directory: directory, Repository: owner.Repository, RunRoot: owner.RunRoot}, nil
}
func (r *Runtime) Environment() map[string]string {
	v := map[string]string{"CARTULARY_HARNESS_SUITE_RUNTIME_ROOT": r.Root, "CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID": r.LeaseID, "CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID": r.RunID}
	id := "ordinary"
	if r.attempt != "" {
		id = r.attempt
		v["CARTULARY_HARNESS_RECOVERY_ATTEMPT"] = id
	}
	if authority, err := r.readDetail(id); err == nil {
		v[detailMarker] = authority.Scope.Token
	}
	return v
}
func (r *Runtime) PrivatePath(parts ...string) (string, error) {
	return r.detailPath("ordinary", parts...)
}

func (r *Runtime) RecoveryPath(parts ...string) (string, error) {
	if len(parts) == 0 {
		return "", harnesslocal.ErrUnsafe
	}
	for _, part := range parts {
		if !component.MatchString(part) || part == "." || part == ".." {
			return "", harnesslocal.ErrUnsafe
		}
	}
	if slices.Contains([]string{"detail", "recovery", "registered-secrets", "runtime-owner.json", "runtime-process.json", "finalizing", "detail-authority.json", "recovery-attempts", "recovery-detail"}, parts[0]) {
		return "", harnesslocal.ErrUnsafe
	}
	return filepath.Join(append([]string{r.Root}, parts...)...), nil
}

// Registration is durable across short-lived producer clients. The directory
// lock also serializes other handles/processes against close and preservation.
func (r *Runtime) RegisterSecret(value string) error {
	if len(value) < 8 {
		return nil
	}
	if len(value) > 1048576 {
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
	dir, err := harnesslocal.PrivateDirectory(filepath.Join(r.Root, "registered-secrets"))
	if err != nil {
		return err
	}
	defer dir.Close()
	name := fmt.Sprintf("%x", sha256.Sum256([]byte(value)))
	if existing, err := harnesslocal.ReadFileAt(dir, name, 1048576, false); err == nil {
		if string(existing) != value {
			return harnesslocal.ErrUnsafe
		}
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	names, err := dir.Readdirnames(4097)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	if len(names) >= 4096 {
		return harnesslocal.ErrUnsafe
	}
	return harnesslocal.ReplaceFileAt(dir, name, []byte(value))
}
func (r *Runtime) RegisterEnvironment(environment map[string]string) error {
	for k, v := range environment {
		if secretName.MatchString(k) {
			if err := r.RegisterSecret(v); err != nil {
				return err
			}
		}
	}
	return nil
}
func (r *Runtime) ForbiddenValues() ([]string, error) {
	unlock, err := r.lock()
	if err != nil {
		return nil, err
	}
	defer unlock()
	return r.forbiddenValues()
}
func (r *Runtime) forbiddenValues() ([]string, error) {
	dir, err := harnesslocal.OpenDirectory(filepath.Join(r.Root, "registered-secrets"), true)
	if errors.Is(err, os.ErrNotExist) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	names, err := dir.Readdirnames(4097)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	if len(names) > 4096 {
		return nil, harnesslocal.ErrUnsafe
	}
	values := []string{}
	for _, name := range names {
		data, err := harnesslocal.ReadFileAt(dir, name, 1048576, false)
		if err != nil {
			return nil, err
		}
		if len(data) < 8 || fmt.Sprintf("%x", sha256.Sum256(data)) != name {
			return nil, harnesslocal.ErrUnsafe
		}
		values = append(values, string(data))
	}
	slices.Sort(values)
	return values, nil
}

type Resource struct {
	Dependencies  []string        `json:"depends_on"`
	RecoveryFiles []string        `json:"recovery_files,omitempty"`
	Kind          string          `json:"kind"`
	Target        json.RawMessage `json:"target"`
	State         string          `json:"state"`
	LeaseID       string          `json:"lease_id"`
	RunID         string          `json:"run_id"`
}

func (r Resource) Path() (string, error) {
	var path string
	err := json.Unmarshal(r.Target, &path)
	return path, err
}
func resourceName(kind string, target json.RawMessage) (string, error) {
	var value any
	if err := json.Unmarshal(target, &value); err != nil {
		return "", err
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%x.json", kind, sha256.Sum256(canonical)), nil
}

func (v Resource) Reference() string {
	name, err := resourceName(v.Kind, v.Target)
	if err != nil {
		return ""
	}
	return "resource:" + v.RunID + ":" + v.LeaseID + ":" + strings.TrimSuffix(name, ".json")
}

func (r *Runtime) ResourceReference(kind string, target any) (string, error) {
	raw, err := json.Marshal(target)
	if err != nil {
		return "", err
	}
	v := Resource{Kind: kind, Target: raw, State: "pending", RunID: r.RunID, LeaseID: r.LeaseID}
	if err := r.checkResource(v); err != nil {
		return "", err
	}
	return v.Reference(), nil
}

func (r *Runtime) referenceName(reference string) (string, error) {
	prefix := "resource:" + r.RunID + ":" + r.LeaseID + ":"
	if !strings.HasPrefix(reference, prefix) {
		return "", harnesslocal.ErrUnsafe
	}
	name := strings.TrimPrefix(reference, prefix)
	if !regexp.MustCompile(`^[a-z_]+-[a-f0-9]{64}$`).MatchString(name) {
		return "", harnesslocal.ErrUnsafe
	}
	return name + ".json", nil
}
func (r *Runtime) checkResource(v Resource) error {
	if _, err := validateJSON("cartulary.harness_runtime_resource.v2", v); err != nil {
		return harnesslocal.ErrUnsafe
	}
	if v.LeaseID != r.LeaseID || v.RunID != r.RunID {
		return harnesslocal.ErrUnsafe
	}
	for _, file := range v.RecoveryFiles {
		if !filepath.IsAbs(file) || filepath.Clean(file) != file || file == r.Root || !within(r.Root, file) || within(filepath.Join(r.Root, "recovery"), file) || (within(filepath.Join(r.Root, "detail"), file) || within(filepath.Join(r.Root, "recovery-detail"), file) || within(filepath.Join(r.Root, "recovery-attempts"), file) || file == filepath.Join(r.Root, "detail-authority.json")) {
			return harnesslocal.ErrUnsafe
		}
	}
	for _, dependency := range v.Dependencies {
		if _, err := r.referenceName(dependency); err != nil || dependency == v.Reference() {
			return harnesslocal.ErrUnsafe
		}
	}
	switch v.Kind {
	case "managed_suite", "borrowed_suite", "browser_stack", "object_store_proxy", "resident_allocation":
		target, err := v.Path()
		if err != nil || !filepath.IsAbs(target) || filepath.Clean(target) != target || target == r.Root || !within(r.Root, target) || (within(filepath.Join(r.Root, "detail"), target) || within(filepath.Join(r.Root, "recovery-detail"), target) || within(filepath.Join(r.Root, "recovery-attempts"), target) || target == filepath.Join(r.Root, "detail-authority.json")) {
			return harnesslocal.ErrUnsafe
		}
		relative, _ := filepath.Rel(r.Root, target)
		for _, part := range strings.Split(relative, string(filepath.Separator)) {
			if part == "recovery" || part == "." || part == ".." {
				return harnesslocal.ErrUnsafe
			}
		}
	case "browser_process", "helper_process":
		// The projected schema validates exact process proof fields.
	case "diagnostic_scope":
		var proof struct {
			UID int `json:"uid"`
		}
		if json.Unmarshal(v.Target, &proof) != nil || proof.UID != os.Getuid() {
			return harnesslocal.ErrUnsafe
		}
	case "browser_renderer":
		var proof struct {
			Name  string `json:"name"`
			Token string `json:"token"`
		}
		if json.Unmarshal(v.Target, &proof) != nil || !strings.HasSuffix(proof.Name, "-"+proof.Token[:12]) {
			return harnesslocal.ErrUnsafe
		}
	default:
		return harnesslocal.ErrUnsafe
	}
	return nil
}
func (r *Runtime) RecordResource(kind string, target any, state string, files ...string) error {
	return r.RecordDependentResource(kind, target, state, nil, files...)
}

// Nil dependencies mean unknown; an empty list explicitly proves independence.
// Retention files do not imply lifecycle edges. Only the fixture owner supplies
// those edges, before starting a consumer of the dependency.
func (r *Runtime) RecordDependentResource(kind string, target any, state string, dependencies []string, files ...string) error {
	unlock, err := r.lock()
	if err != nil {
		return err
	}
	defer unlock()
	if state != "released" {
		if err := r.productionMutable(); err != nil {
			return err
		}
	}
	raw, err := json.Marshal(target)
	if err != nil {
		return err
	}
	v := Resource{Kind: kind, Target: raw, State: state, LeaseID: r.LeaseID, RunID: r.RunID, Dependencies: dependencies}
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
	if prior, err := harnesslocal.ReadFileAt(directory, name, 1048576, false); err == nil {
		var previous Resource
		if json.Unmarshal(prior, &previous) != nil || r.checkResource(previous) != nil {
			return harnesslocal.ErrUnsafe
		}
		files = append(append([]string{}, previous.RecoveryFiles...), files...)
		if dependencies == nil {
			v.Dependencies = previous.Dependencies
		} else {
			v.Dependencies = slices.Clone(dependencies)
			slices.Sort(v.Dependencies)
			old := slices.Clone(previous.Dependencies)
			slices.Sort(old)
			if previous.Dependencies != nil && !slices.Equal(old, v.Dependencies) {
				return harnesslocal.ErrUnsafe
			}
		}
		if previous.State == "acquired" {
			v.State = "acquired"
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if v.Dependencies == nil {
		return harnesslocal.ErrUnsafe
	}
	for _, reference := range v.Dependencies {
		dependencyName, err := r.referenceName(reference)
		if err != nil {
			return err
		}
		data, err := harnesslocal.ReadFileAt(directory, dependencyName, 1048576, false)
		var dependency Resource
		if err != nil || json.Unmarshal(data, &dependency) != nil || r.checkResource(dependency) != nil || dependency.Reference() != reference || dependency.State == "released" {
			return harnesslocal.ErrUnsafe
		}
	}
	v.RecoveryFiles = slices.Clone(files)
	slices.Sort(v.RecoveryFiles)
	v.RecoveryFiles = slices.Compact(v.RecoveryFiles)
	if v.Dependencies != nil {
		slices.Sort(v.Dependencies)
		v.Dependencies = slices.Compact(v.Dependencies)
	}
	if err := r.checkResource(v); err != nil {
		return err
	}
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return harnesslocal.ReplaceFileAt(directory, name, append(data, '\n'))
}
func (r *Runtime) Resources(maximum int) ([]Resource, error) {
	unlock, err := r.lock()
	if err != nil {
		return nil, err
	}
	defer unlock()
	if maximum < 1 || maximum > 4096 {
		return nil, harnesslocal.ErrUnsafe
	}
	return r.resources(maximum)
}
func (r *Runtime) resources(maximum int) ([]Resource, error) {
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
		if harnessschema.Validate("cartulary.harness_runtime_resource.v2", value) != nil {
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
	if err := unix.Flock(int(r.Directory.Fd()), unix.LOCK_EX); err != nil {
		return err
	}
	defer unix.Flock(int(r.Directory.Fd()), unix.LOCK_UN)
	if harnesslocal.CheckDirectory(r.Directory, r.Root) != nil {
		return harnesslocal.ErrUnsafe
	}
	if _, err := ReadBinding(r.Repository, r.RunRoot, r.Root, r.RunID, r.LeaseID); err != nil {
		return err
	}
	if err := r.fenceDetails(); err != nil {
		return err
	}
	if err := r.purgeDetail("ordinary"); err != nil {
		return err
	}
	if err := r.purgeRecoveryDetails(); err != nil {
		return err
	}
	resources, err := r.resources(4096)
	if err != nil {
		return err
	}
	if len(resources) > 0 {
		return ErrUnresolved
	}
	if harnesslocal.CheckDirectory(r.parent, filepath.Dir(r.Root)) != nil || harnesslocal.CheckDirectory(r.Directory, r.Root) != nil {
		return harnesslocal.ErrUnsafe
	}
	if err := RemoveOwnedTree(r.Directory, r.Root); err != nil {
		return err
	}
	r.closed = true
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
	var parentErr error
	if r.parent != nil {
		parentErr = r.parent.Close()
	}
	return errors.Join(r.Directory.Close(), parentErr)
}

// BeginFinalization first stops ordinary production, independently of external
// cleanup. Successful completion additionally requires settled resources.
func (r *Runtime) BeginFinalization() ([]string, error) {
	unlock, err := r.lock()
	if err != nil {
		return nil, err
	}
	defer unlock()
	if !r.owned {
		return nil, harnesslocal.ErrUnsafe
	}
	if err := r.fenceDetails(); err != nil {
		return nil, err
	}
	if err := r.purgeDetail("ordinary"); err != nil {
		return nil, err
	}
	resources, err := r.resources(4096)
	if err != nil {
		return nil, err
	}
	if len(resources) != 0 {
		return nil, ErrUnresolved
	}
	values, err := r.forbiddenValues()
	if err != nil {
		return nil, err
	}
	return values, nil
}
