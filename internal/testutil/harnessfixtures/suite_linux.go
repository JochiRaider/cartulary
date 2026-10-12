//go:build linux

package harnessfixtures

import (
	"context"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"golang.org/x/sys/unix"
)

type Suite struct {
	Environment map[string]string
	LeaseFile   string
	ScopeRef    string
	Close       func(context.Context) error
}

// SuiteOwner serializes the one run-owned physical suite. Fixture providers
// borrow its bindings; it closes only after all dependent allocations drain.
type SuiteOwner struct {
	mu                sync.Mutex
	Acquire           func(context.Context) (*Suite, error)
	suite             *Suite
	err               error
	attempted, closed bool
	closeOnce         sync.Once
	closeErr          error
}

func (s *SuiteOwner) Ensure(ctx context.Context) (*Suite, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, fixtureFailure(errors.New("suite owner closed"))
	}
	if !s.attempted {
		s.attempted = true
		s.suite, s.err = s.Acquire(ctx)
	}
	return s.suite, s.err
}
func (s *SuiteOwner) Close(ctx context.Context) error {
	s.closeOnce.Do(func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.closed = true
		if s.suite != nil && s.suite.Close != nil {
			s.closeErr = s.suite.Close(ctx)
		}
	})
	return s.closeErr
}

type ProductionOptions struct {
	Root, RunRoot, Target string
	Runtime               *harnesssession.Runtime
	Environment           map[string]string
	Run                   Command
	ContainerProof        func(map[string]any)
}

func (p ProductionOptions) command() Command {
	if p.Run != nil {
		return p.Run
	}
	return commandRunner(p.Root, p.Runtime)
}
func (p ProductionOptions) environment() map[string]string {
	v := maps.Clone(p.Environment)
	if v == nil {
		v = map[string]string{}
	}
	maps.Copy(v, p.Runtime.Environment())
	v["CARTULARY_TEST_RESULTS_DIR"] = filepath.Dir(p.RunRoot)
	v["CARTULARY_TEST_RUN_ID"] = p.Runtime.RunID
	v["CARTULARY_TEST_TARGET"] = p.Target
	return v
}
func (p ProductionOptions) serviceBinary() string {
	for _, name := range []string{"CARTULARY_TEST_SERVICES_BIN", "TEST_SERVICES_BIN"} {
		if v := p.Environment[name]; v != "" {
			if filepath.IsAbs(v) {
				return v
			}
			return filepath.Join(p.Root, v)
		}
	}
	return ""
}
func (p ProductionOptions) containerProof(value map[string]any) {
	if p.ContainerProof == nil {
		return
	}
	defer func() { _ = recover() }()
	p.ContainerProof(value)
}
func readObject(name, schema string, maximum int64) (map[string]any, error) {
	data, err := harnesslocal.ReadFile(name, maximum, false)
	if err != nil {
		return nil, err
	}
	value, err := harnessschema.Decode(data)
	if err != nil || harnessschema.Validate(schema, value) != nil {
		return nil, harnesslocal.ErrUnsafe
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, harnesslocal.ErrUnsafe
	}
	return object, nil
}
func removePrivate(name string) error {
	parent, err := harnesslocal.OpenDirectory(filepath.Dir(name), true)
	if err != nil {
		return err
	}
	defer parent.Close()
	if _, err := harnesslocal.ReadFileAt(parent, filepath.Base(name), 1048576, false); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	return unix.Unlinkat(int(parent.Fd()), filepath.Base(name), 0)
}
func (p ProductionOptions) stopSuite(ctx context.Context, lease string, environment map[string]string) error {
	bytes, err := harnesslocal.ReadFile(lease, 1048576, false)
	if err != nil {
		return err
	}
	snapshot := lease + ".cleanup-" + identifier()
	if err := harnesslocal.PublishNew(snapshot, bytes); err != nil {
		return err
	}
	_, stopErr := p.command()(ctx, p.serviceBinary(), []string{"terminate-suite", "--lease", snapshot}, environment)
	cleanupErr := removePrivate(snapshot)
	if stopErr != nil || cleanupErr != nil {
		return CleanupFailure(errors.Join(stopErr, cleanupErr))
	}
	if err := p.Runtime.RecordResource("managed_suite", lease, "released"); err != nil {
		return artifactFailure(err)
	}
	return removePrivate(lease)
}
func (p ProductionOptions) StartSuite(ctx context.Context) (suite *Suite, result error) {
	root, err := p.Runtime.PrivatePath("test-services")
	if err != nil {
		return nil, err
	}
	directory, err := harnesslocal.PrivateDirectory(root)
	if err != nil {
		return nil, err
	}
	directory.Close()
	environmentFile, leaseFile, resultFile := filepath.Join(root, "suite-environment.json"), filepath.Join(root, "suite-lease.json"), filepath.Join(root, "suite-start-result.json")
	environment := p.environment()
	environment["CARTULARY_TEST_SUITE_ID"] = regexp.MustCompile(`[^A-Za-z0-9_.-]+`).ReplaceAllString("work-graph-"+p.Target, "-")
	delete(environment, "CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT")
	if err := p.Runtime.RecordResource("managed_suite", leaseFile, "pending"); err != nil {
		return nil, artifactFailure(err)
	}
	startup, commandErr := p.command()(ctx, p.serviceBinary(), []string{"start-suite", "--env-file", environmentFile, "--lease-file", leaseFile, "--result-file", resultFile}, environment)
	// Successful helper exit is insufficient: exact start/scope evidence must
	// agree before bindings or container proof become available to consumers.
	settled := false
	defer func() {
		if result == nil {
			return
		}
		if _, err := harnesslocal.ReadFile(leaseFile, 1048576, false); err == nil {
			cleanupCtx, done := context.WithTimeout(context.Background(), 2*time.Minute)
			cleanup := p.stopSuite(cleanupCtx, leaseFile, environment)
			done()
			result = errors.Join(result, cleanup)
		} else if !startup.Started && !settled {
			result = errors.Join(result, p.Runtime.RecordResource("managed_suite", leaseFile, "released"))
		}
	}()
	start, err := readObject(resultFile, "cartulary.test_services.start_result.v1", 1048576)
	if err != nil {
		if !startup.Started {
			return nil, &Failure{Class: "config", Reason: "configuration_error", Cause: commandErr}
		}
		if errors.Is(err, os.ErrNotExist) && startup.Exit == 2 {
			return nil, &Failure{Class: "config", Reason: "configuration_error", Cause: commandErr}
		}
		return nil, artifactFailure(err)
	}
	defer func() { result = errors.Join(result, removePrivate(resultFile)) }()
	if start["run_id"] != p.Runtime.RunID || start["target"] != p.Target {
		return nil, artifactFailure(errors.New("start identity mismatch"))
	}
	ref, _ := start["service_scope_ref"].(string)
	suiteID, _ := start["suite_id"].(string)
	if !component.MatchString(suiteID) || ref != "_shared/test-services/"+suiteID+"/service-scope.json" {
		return nil, artifactFailure(errors.New("scope reference mismatch"))
	}
	scope, err := readObject(filepath.Join(p.RunRoot, ref), "cartulary.test_services.scope.v2", 1048576)
	if err != nil {
		return nil, artifactFailure(err)
	}
	if scope["run_id"] != start["run_id"] || scope["target"] != start["target"] || scope["suite_id"] != suiteID {
		return nil, artifactFailure(errors.New("scope identity mismatch"))
	}
	if startup.Exit != 0 || commandErr != nil {
		failure, ok := scope["failure"].(map[string]any)
		if !ok || start["status"] != "failed" || failure["failure_class"] != start["failure_class"] || failure["failure_reason"] != start["failure_reason"] {
			return nil, artifactFailure(errors.New("startup failure evidence mismatch"))
		}
		if cleanup, ok := scope["cleanup"].(map[string]any); ok && cleanup["status"] == "startup_failed" {
			if err := p.Runtime.RecordResource("managed_suite", leaseFile, "released"); err != nil {
				return nil, artifactFailure(err)
			}
			settled = true
		}
		class, _ := start["failure_class"].(string)
		reason, _ := start["failure_reason"].(string)
		return nil, &Failure{Class: class, Reason: reason, Refs: []string{ref}, Cause: commandErr}
	}
	if start["status"] != "ready" || scope["failure"] != nil {
		return nil, artifactFailure(errors.New("ready evidence mismatch"))
	}
	data, err := harnesslocal.ReadFile(environmentFile, 1048576, false)
	if err != nil {
		return nil, artifactFailure(err)
	}
	value, err := harnessschema.Decode(data)
	if err != nil {
		return nil, artifactFailure(err)
	}
	raw, ok := value.(map[string]any)
	if !ok {
		return nil, artifactFailure(harnesslocal.ErrUnsafe)
	}
	admitted := map[string]string{}
	for name, value := range raw {
		if name == "CARTULARY_TEST_SUITE_ID" || name == "SSL_CERT_FILE" || strings.HasPrefix(name, "CARTULARY_HARNESS_SUITE_RUNTIME_") || strings.HasPrefix(name, "CARTULARY_TEST_SERVICES_") || strings.HasPrefix(name, "CARTULARY_PGTEST_") || strings.HasPrefix(name, "CARTULARY_S3TEST_") {
			text, ok := value.(string)
			if !ok {
				return nil, artifactFailure(harnesslocal.ErrUnsafe)
			}
			admitted[name] = text
		}
	}
	if err := p.Runtime.RegisterEnvironment(admitted); err != nil {
		return nil, artifactFailure(err)
	}
	maps.Copy(environment, admitted)
	if err := removePrivate(environmentFile); err != nil {
		return nil, artifactFailure(err)
	}
	if err := p.Runtime.RecordResource("managed_suite", leaseFile, "acquired"); err != nil {
		return nil, artifactFailure(err)
	}
	if proof, err := readObject(leaseFile, "cartulary.test_services.lease.v1", 1048576); err == nil {
		p.containerProof(map[string]any{"lease": proof, "ownership": "owned", "borrower_file": "", "borrower_id": ""})
	}
	var once sync.Once
	var closeErr error
	suite = &Suite{Environment: admitted, LeaseFile: leaseFile, ScopeRef: ref, Close: func(ctx context.Context) error {
		once.Do(func() { p.containerProof(nil); closeErr = p.stopSuite(ctx, leaseFile, environment) })
		return closeErr
	}}
	return suite, nil
}
