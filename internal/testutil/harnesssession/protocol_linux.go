//go:build linux

package harnesssession

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

const RequestSchema = "cartulary.harness_runtime_session.v2"

type request struct {
	Schema         string            `json:"schema_id"`
	ID             int               `json:"request_id"`
	Operation      string            `json:"operation"`
	Repository     string            `json:"repository"`
	RunRoot        string            `json:"run_root"`
	RunID          string            `json:"run_id"`
	Root           string            `json:"root"`
	LeaseID        string            `json:"lease_id"`
	Scratch        string            `json:"scratch"`
	Parts          []string          `json:"parts"`
	Values         []string          `json:"values"`
	Environment    map[string]string `json:"environment"`
	Resource       *Resource         `json:"resource"`
	AdmissionToken string            `json:"admission_token"`
}

type response struct {
	Schema    string `json:"schema_id"`
	ID        int    `json:"request_id"`
	Operation string `json:"operation"`
	OK        bool   `json:"ok"`
	Result    any    `json:"result"`
}

func runtimeIdentity(r *Runtime) map[string]string {
	return map[string]string{"root": r.Root, "run_id": r.RunID, "lease_id": r.LeaseID}
}

// OwnRuntime admits the original owner or an explicitly proven-dead owner for
// exact recovery. Borrowing alone never grants removal authority.
func OwnRuntime(repository, runRoot, root, runID, leaseID string, caller harnessadmission.Proof, recovery bool) (*Runtime, error) {
	r, err := BorrowRuntime(repository, runRoot, root, runID, leaseID)
	if err != nil {
		return nil, err
	}
	data, err := harnesslocal.ReadFileAt(r.Directory, "runtime-process.json", 4096, false)
	var proof harnessadmission.Proof
	if err == nil {
		value, e := harnessschema.Decode(data)
		fields, ok := value.(map[string]any)
		if e != nil || !ok || len(fields) != 3 || json.Unmarshal(data, &proof) != nil || proof.Boot == "" || proof.PID < 2 || !regexpDigits.MatchString(proof.Start) {
			err = harnesslocal.ErrUnsafe
		}
	}
	if err == nil && proof != caller {
		alive, e := harnessadmission.Alive(proof)
		if !recovery || e != nil || alive {
			err = harnesslocal.ErrUnsafe
		}
	}
	if err == nil && (!strings.HasPrefix(filepath.Base(root), "suite-") || filepath.Base(filepath.Dir(root)) != "suite-runtime") {
		err = harnesslocal.ErrUnsafe
	}
	if err == nil {
		r.parent, err = harnesslocal.OpenDirectory(filepath.Dir(root), true)
	}
	if err != nil {
		r.ReleaseHandles()
		return nil, err
	}
	r.owned = true
	return r, nil
}

func execute(data []byte, repository string, caller harnessadmission.Proof) (out response, result error) {
	if len(data) == 0 || len(data) > MaximumFrame {
		return out, harnesslocal.ErrUnsafe
	}
	value, err := harnessschema.Decode(data)
	if err != nil || harnessschema.Validate(RequestSchema, value) != nil {
		return out, harnesslocal.ErrUnsafe
	}
	var q request
	if json.Unmarshal(data, &q) != nil || q.Repository != repository || !filepath.IsAbs(q.RunRoot) || q.RunID == "." || q.RunID == ".." {
		return out, harnesslocal.ErrUnsafe
	}
	out = response{Schema: RequestSchema, ID: q.ID, Operation: q.Operation}
	defer func() {
		out.OK = result == nil
		if result != nil {
			out.Result = nil
		}
	}()
	var r *Runtime
	if q.Operation == "create" {
		r, err = newRuntime(repository, q.RunRoot, q.RunID, q.Scratch, caller)
	} else if q.Operation == "close" || q.Operation == "recover_close" {
		r, err = OwnRuntime(repository, q.RunRoot, q.Root, q.RunID, q.LeaseID, caller, q.Operation == "recover_close")
	} else {
		r, err = BorrowRuntime(repository, q.RunRoot, q.Root, q.RunID, q.LeaseID)
	}
	if err != nil {
		return out, err
	}
	defer func() { result = errors.Join(result, r.ReleaseHandles()) }()
	switch q.Operation {
	case "create", "borrow":
		out.Result = runtimeIdentity(r)
	case "private_path":
		out.Result, result = r.PrivatePath(q.Parts...)
	case "recovery_path":
		out.Result, result = r.RecoveryPath(q.Parts...)
	case "claim_recovery_output":
		result = ClaimRecoveryOutput(repository, q.RunRoot, r.Root, r.RunID, r.LeaseID)
	case "register":
		for _, v := range q.Values {
			if err := r.RegisterSecret(v); err != nil {
				return out, err
			}
		}
		result = r.RegisterEnvironment(q.Environment)
	case "record":
		if q.Resource.RunID != r.RunID || q.Resource.LeaseID != r.LeaseID {
			return out, harnesslocal.ErrUnsafe
		}
		result = r.RecordDependentResource(q.Resource.Kind, q.Resource.Target, q.Resource.State, q.Resource.Dependencies, q.Resource.RecoveryFiles...)
		if result == nil && q.AdmissionToken != "" && q.Resource.State != "released" && strings.HasSuffix(q.Resource.Kind, "_process") {
			data, _ := json.Marshal(q.Resource.Target)
			var proof harnessadmission.Proof
			if json.Unmarshal(data, &proof) != nil {
				return out, harnesslocal.ErrUnsafe
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, result = harnessadmission.Transact(ctx, harnessadmission.Request{Root: harnessadmission.DefaultRoot(), Operation: "bind", Token: q.AdmissionToken, Guard: proof})
			if result != nil {
				alive, err := harnessadmission.Alive(proof)
				if err == nil && !alive {
					result = nil
				}
			}
		}
	case "close", "recover_close":
		result = r.Close()
	default:
		return out, harnesslocal.ErrUnsafe
	}
	return out, result
}

// Main is a single bounded transaction over inherited private pipes. The actual
// parent identity is resolved locally and cannot be supplied in a request.
func Main(args []string, input io.Reader, output io.Writer) int {
	if len(args) != 0 {
		return 2
	}
	repository, err := os.Getwd()
	if err != nil {
		return 11
	}
	caller, err := harnessadmission.Identity(os.Getppid())
	if err != nil {
		return 11
	}
	data, err := io.ReadAll(io.LimitReader(input, MaximumFrame+1))
	if err != nil {
		return 11
	}
	reply, err := execute(data, repository, caller)
	if err != nil {
		return 11
	}
	encoded, e := json.Marshal(reply)
	if e != nil || len(encoded)+1 > MaximumFrame {
		return 11
	}
	value, e := harnessschema.Decode(encoded)
	if e != nil || harnessschema.Validate(RequestSchema, value) != nil {
		return 11
	}
	if _, e = output.Write(append(encoded, '\n')); e != nil {
		return 11
	}
	return 0
}
