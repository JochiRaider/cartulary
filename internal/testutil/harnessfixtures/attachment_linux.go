//go:build linux

package harnessfixtures

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

func (p ProductionOptions) attachmentOperation(ctx context.Context, op string) (map[string]any, error) {
	env := p.environment()
	proof, err := harnessadmission.Identity(os.Getpid())
	if err != nil {
		return nil, err
	}
	bytes, _ := json.Marshal(proof)
	env["CARTULARY_NATIVE_FIXTURE_OWNER"] = string(bytes)
	response, err := p.command()(ctx, p.Environment["NODE_BIN"], []string{filepath.Join(p.Root, "tools/harness/services/fixture-attachment.mjs"), op}, env)
	if err != nil {
		return nil, artifactFailure(err)
	}
	value, err := harnessschema.Decode(response.Stdout)
	if err != nil {
		return nil, artifactFailure(err)
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, artifactFailure(harnesslocal.ErrUnsafe)
	}
	return object, nil
}
func (p ProductionOptions) AttachSuite(ctx context.Context) (suite *Suite, result error) {
	file := filepath.Join(p.Runtime.Root, "borrowed-suite.json")
	leaseFile := filepath.Join(p.Runtime.Root, "borrowed-service-lease.json")
	if err := p.Runtime.RecordResource("borrowed_suite", file, "pending"); err != nil {
		return nil, err
	}
	var once sync.Once
	var closeErr error
	close := func(ctx context.Context) error {
		once.Do(func() {
			p.containerProof(nil)
			_, closeErr = p.attachmentOperation(ctx, "detach")
			if closeErr == nil {
				closeErr = p.Runtime.RecordResource("borrowed_suite", file, "released")
				if closeErr == nil {
					closeErr = removePrivate(leaseFile)
				}
			}
		})
		return closeErr
	}
	defer func() {
		if result != nil {
			cleanup, done := context.WithTimeout(context.Background(), 30*time.Second)
			defer done()
			if _, err := harnesslocal.ReadFile(file, 1048576, false); err == nil {
				result = errors.Join(result, close(cleanup))
			}
		}
	}()
	values, err := p.attachmentOperation(ctx, "attach")
	if err != nil {
		return nil, err
	}
	if len(values) != 2 {
		return nil, artifactFailure(harnesslocal.ErrUnsafe)
	}
	raw, ok := values["environment"].(map[string]any)
	if !ok {
		return nil, artifactFailure(harnesslocal.ErrUnsafe)
	}
	environment := map[string]string{}
	for k, v := range raw {
		text, ok := v.(string)
		if !ok {
			return nil, artifactFailure(harnesslocal.ErrUnsafe)
		}
		environment[k] = text
	}
	if err := p.Runtime.RegisterEnvironment(environment); err != nil {
		return nil, artifactFailure(err)
	}
	if err := p.Runtime.RecordResource("borrowed_suite", file, "acquired"); err != nil {
		return nil, err
	}
	proof, ok := values["container_proof"].(map[string]any)
	if !ok {
		return nil, artifactFailure(harnesslocal.ErrUnsafe)
	}
	source, ok := proof["lease"].(map[string]any)
	if !ok || harnessschema.Validate("cartulary.test_services.lease.v1", source) != nil {
		return nil, artifactFailure(harnesslocal.ErrUnsafe)
	}
	// This run-scoped attachment proof permits browser fixture cleanup only. The
	// borrowed-suite owner detaches; it never dispatches terminate-suite with it.
	lease := maps.Clone(source)
	for key, value := range map[string]any{"lease_id": p.Runtime.LeaseID, "suite_id": environment["CARTULARY_TEST_SUITE_ID"], "run_id": p.Runtime.RunID, "result_root": filepath.Dir(p.RunRoot), "run_root": p.RunRoot, "target": p.Target, "mode": "attach", "ownership_mode": "attach", "owner_pid": os.Getpid(), "created_at": time.Now().UTC().Format(time.RFC3339Nano), "cleanup_state": "not_started"} {
		lease[key] = value
	}
	if err := privateJSON(leaseFile, "cartulary.test_services.lease.v1", lease, false); err != nil {
		return nil, artifactFailure(err)
	}
	p.containerProof(proof)
	return &Suite{Environment: environment, LeaseFile: leaseFile, Close: close}, nil
}
