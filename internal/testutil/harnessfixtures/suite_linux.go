//go:build linux

package harnessfixtures

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"github.com/JochiRaider/cartulary/internal/testutil/pgschema"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"golang.org/x/sys/unix"
)

type Suite struct {
	Environment  map[string]string
	LeaseFile    string
	ResourceRef  string
	ResourceRefs []string
	ScopeRef     string
	Close        func(context.Context) error
}

// SuiteOwner serializes one physical service generation. Fixture providers
// borrow its bindings; it closes only after all dependent allocations drain.
type SuiteOwner struct {
	mu                sync.Mutex
	Acquire           func(context.Context) (*Suite, error)
	ReleaseResidency  func(context.Context) error
	suite             *Suite
	binding           *Suite
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
		if s.closeErr == nil && s.ReleaseResidency != nil {
			s.closeErr = s.ReleaseResidency(ctx)
		}
	})
	return s.closeErr
}

type ProductionOptions struct {
	ServiceKey, ServiceID, PostgresIsolation, PostgresAllocation string
	ServiceDependencies                                          []string
	BrowserLaneLimit                                             int
	residency                                                    *residency
	Root, RunRoot, Target                                        string
	Runtime                                                      *harnesssession.Runtime
	Environment                                                  map[string]string
	Run                                                          Command
	ContainerProof                                               func(context.Context, map[string]any, bool)
	Dependencies                                                 []string
	ReportCleanup                                                func(CleanupOutcome)
	RecoveryOutput                                               bool
}

func resourceDependencies(references []string) string {
	if references == nil {
		references = []string{}
	}
	data, _ := json.Marshal(references)
	return string(data)
}

func (p ProductionOptions) command() Command {
	dependencies := p.Dependencies
	if dependencies == nil {
		dependencies = []string{}
	}
	run := p.Run
	if run == nil {
		run = commandRunner(p.Root, p.Runtime, dependencies)
	}
	return func(ctx context.Context, command string, args []string, environment map[string]string) (CommandResult, error) {
		if token, _ := ctx.Value(activityKey{}).(string); token != "" {
			environment = maps.Clone(environment)
			environment["CARTULARY_HOST_ADMISSION_LEASE"] = token
		}
		return run(ctx, command, args, environment)
	}
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
	v["CARTULARY_HARNESS_RESOURCE_DEPENDENCIES"] = resourceDependencies(p.Dependencies)
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
func (p ProductionOptions) containerProof(ctx context.Context, value map[string]any, retiring bool) {
	if p.ContainerProof == nil {
		return
	}
	defer func() { _ = recover() }()
	p.ContainerProof(ctx, value, retiring)
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
	return p.closeResource(ctx, "managed_suite", lease, func(step context.Context) error { return p.stopSuitePhysical(step, lease, environment) })
}
func (p ProductionOptions) stopSuitePhysical(ctx context.Context, lease string, environment map[string]string) error {
	return p.Runtime.AttemptCleanup(ctx, "managed_suite", lease, func(step context.Context) error {
		// The dependency evaluator may release the physical suite before the
		// logical Suite.Close handle. Capture while the owned container is live.
		if proof, err := readObject(lease, "cartulary.test_services.lease.v3", 1048576); err == nil {
			p.containerProof(step, map[string]any{"lease": proof, "ownership": "owned", "borrower_file": "", "borrower_id": ""}, true)
		}
		return p.stopSuiteAttempt(step, lease, environment)
	})
}
func (p ProductionOptions) stopSuiteAttempt(ctx context.Context, lease string, environment map[string]string) (result error) {
	defer func() { result = harnesssession.AtCleanupStage("managed_service_settlement", result) }()
	reference, err := p.Runtime.ResourceReference("managed_suite", lease)
	if err != nil {
		return err
	}
	p.Dependencies = []string{reference}
	bytes, err := harnesslocal.ReadFile(lease, 1048576, false)
	if err != nil {
		return err
	}
	snapshot := lease + ".cleanup-" + identifier()
	if err := harnesslocal.PublishNew(snapshot, bytes); err != nil {
		return err
	}
	args := []string{"terminate-suite", "--lease", snapshot}
	if p.RecoveryOutput {
		output, err := p.Runtime.AllocateRecoveryOutput()
		if err != nil {
			return errors.Join(err, removePrivate(snapshot))
		}
		args = append(args, "--recovery-output", output)
	}
	_, stopErr := p.command()(ctx, p.serviceBinary(), args, environment)
	cleanupErr := removePrivate(snapshot)
	if stopErr != nil || cleanupErr != nil {
		return CleanupFailure(errors.Join(stopErr, cleanupErr))
	}
	if err := p.Runtime.RecordResource("managed_suite", lease, "released"); err != nil {
		return artifactFailure(harnesssession.AtCleanupStage("resource_release", err))
	}
	return errors.Join(removePrivate(lease), removePrivate(filepath.Join(filepath.Dir(lease), "suite-environment.json")))
}
func (p ProductionOptions) StartSuite(ctx context.Context) (suite *Suite, result error) {
	serviceID := p.ServiceID
	if serviceID == "" {
		serviceID = "suite-" + identifier()
	}
	root, err := p.Runtime.RecoveryPath("service-owners", serviceID)
	if err != nil {
		return nil, err
	}
	directory, err := harnesslocal.PrivateDirectory(root)
	if err != nil {
		return nil, err
	}
	directory.Close()
	environmentFile, leaseFile := filepath.Join(root, "suite-environment.json"), filepath.Join(root, "suite-lease.json")
	resultFile, err := p.Runtime.PrivatePath("suite-start-" + serviceID + ".json")
	if err != nil {
		return nil, err
	}
	bindingsFile, err := p.Runtime.PrivatePath("suite-bindings-" + serviceID + ".json")
	if err != nil {
		return nil, err
	}
	defer func() { result = errors.Join(result, removePrivate(bindingsFile)) }()
	resourceRef, err := p.Runtime.ResourceReference("managed_suite", leaseFile)
	if err != nil {
		return nil, err
	}
	p.Dependencies = []string{resourceRef}
	environment := serviceEnvironment(p.environment(), nil)
	environment["CARTULARY_TEST_SUITE_ID"] = strings.ReplaceAll(identifier(), "-", "")[:24]
	delete(environment, "CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT")
	if p.ServiceDependencies != nil {
		environment["CARTULARY_HARNESS_SERVICE_DEPENDENCIES"] = strings.Join(p.ServiceDependencies, ",")
	}
	environment[suiteservices.PGIsolationEnv] = p.PostgresIsolation
	environment[suiteservices.PGAllocationRefEnv] = p.PostgresAllocation
	dependencies := append([]string{}, p.residency.reference(p.ServiceKey)...)

	// Recovery authority belongs to the broker and must survive failure before
	// readiness. The helper's success-only bindings are disposable until validated;
	// partial publication must never truncate or replace this recovery record.
	if err := saveRecoveryEnvironment(environmentFile, environment); err != nil {
		return nil, artifactFailure(err)
	}
	if err := p.Runtime.RecordDependentResource("managed_suite", leaseFile, "pending", dependencies, environmentFile); err != nil {
		return nil, artifactFailure(err)
	}
	startup, commandErr := p.command()(ctx, p.serviceBinary(), []string{"start-suite", "--env-file", bindingsFile, "--lease-file", leaseFile, "--result-file", resultFile}, environment)
	// Successful helper exit is insufficient: exact start/scope evidence must
	// agree before bindings or container proof become available to consumers.
	settled := false
	settlementAttempted := false
	defer func() {
		if result == nil {
			return
		}
		var settlement error
		if settled {
			return
		}
		if settlementAttempted {
			settlement = harnesssession.ErrUnresolved
		} else if _, err := harnesslocal.ReadFile(leaseFile, 1048576, false); err == nil {
			cleanupCtx, done := context.WithTimeout(context.Background(), 2*time.Minute)
			settlement = p.stopSuite(cleanupCtx, leaseFile, environment)
			done()
		} else if !startup.Started && !settled {
			settlement = p.Runtime.RecordResource("managed_suite", leaseFile, "released")
		} else if !settled {
			settlement = harnesssession.ErrUnresolved
		}
		if settlement != nil {
			// Acquisition already spent its one cleanup attempt. Carry the
			// unresolved ownership into SuiteOwner so closing the generation
			// cannot release resident capacity or retry physical cleanup.
			suite = &Suite{Close: func(context.Context) error { return settlement }}
			result = errors.Join(result, CleanupFailure(settlement))
		}
	}()
	start, err := readObject(resultFile, "cartulary.test_services.start_result.v2", 1048576)
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
	if suiteID != environment["CARTULARY_TEST_SUITE_ID"] || !component.MatchString(suiteID) || ref != "_shared/test-services/"+suiteID+"/service-scope.json" {
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
		if cleanup, ok := scope["cleanup"].(map[string]any); ok {
			settlementAttempted = cleanup["physical_settlement"] == "completed" || cleanup["physical_settlement"] == "unresolved"
			if cleanup["physical_settlement"] == "completed" {
				if err := p.Runtime.RecordResource("managed_suite", leaseFile, "released"); err != nil {
					return nil, artifactFailure(err)
				}
				settled = true
			} else if settlementAttempted {
				files := []string{environmentFile}
				for _, file := range start["recovery_files"].([]any) {
					files = append(files, file.(string))
				}
				if err := p.Runtime.RecordResource("managed_suite", leaseFile, "acquired", files...); err != nil {
					return nil, artifactFailure(err)
				}
			}
		}
		class, _ := start["failure_class"].(string)
		reason, _ := start["failure_reason"].(string)
		return nil, &Failure{Class: class, Reason: reason, Refs: []string{ref}, Cause: commandErr}
	}
	if start["status"] != "ready" || scope["failure"] != nil {
		return nil, artifactFailure(errors.New("ready evidence mismatch"))
	}
	data, err := harnesslocal.ReadFile(bindingsFile, 1048576, false)
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
	if err := saveRecoveryEnvironment(environmentFile, admitted); err != nil {
		return nil, artifactFailure(err)
	}
	recoveryFiles := []string{environmentFile}
	for _, file := range start["recovery_files"].([]any) {
		recoveryFiles = append(recoveryFiles, file.(string))
	}
	if err := p.Runtime.RecordResource("managed_suite", leaseFile, "acquired", recoveryFiles...); err != nil {
		return nil, artifactFailure(err)
	}
	for _, file := range recoveryFiles {
		if _, err := harnesslocal.ReadFile(file, 1048576, false); err != nil {
			return nil, artifactFailure(err)
		}
	}
	proof, err := readObject(leaseFile, "cartulary.test_services.lease.v3", 1048576)
	if err != nil {
		return nil, artifactFailure(err)
	}
	if postgres, ok := proof["postgres_provisioning"].(map[string]any); ok && (postgres["postgres_isolation"] != p.PostgresIsolation || postgres["allocation_ref"] != p.PostgresAllocation) {
		return nil, artifactFailure(errors.New("PostgreSQL binding does not match requested ownership"))
	}
	if err := validatePostgresProvisioning(proof, admitted); err != nil {
		return nil, artifactFailure(err)
	}
	p.containerProof(ctx, map[string]any{"lease": proof, "ownership": "owned", "borrower_file": "", "borrower_id": ""}, false)
	var once sync.Once
	var closeErr error
	admitted["CARTULARY_HARNESS_RESOURCE_DEPENDENCIES"] = resourceDependencies([]string{resourceRef})
	suite = &Suite{Environment: admitted, LeaseFile: leaseFile, ResourceRef: resourceRef, ResourceRefs: []string{resourceRef}, ScopeRef: ref, Close: func(ctx context.Context) error {
		once.Do(func() { closeErr = p.stopSuite(ctx, leaseFile, environment) })
		return closeErr
	}}

	return suite, nil
}

// Only the native prerequisite owner computes policy identity. Attachments
// carry that capability; consumers never repair a service in place.
func validatePostgresProvisioning(lease map[string]any, env map[string]string) error {
	hasPostgres := false
	for _, resource := range lease["resources"].([]any) {
		if resource.(map[string]any)["service"] == "postgres" {
			hasPostgres = true
		}
	}
	if !hasPostgres {
		if lease["postgres_provisioning"] != nil || env[suiteservices.PGAdminDSNEnv] != "" {
			return harnesslocal.ErrUnsafe
		}
		return nil
	}
	proof, ok := lease["postgres_provisioning"].(map[string]any)
	if !ok || proof["generation"] != env[suiteservices.PGGenerationEnv] || env[suiteservices.PGGenerationEnv] == "" ||
		proof["postgres_isolation"] != env[suiteservices.PGIsolationEnv] || proof["allocation_ref"] != env[suiteservices.PGAllocationRefEnv] ||
		proof["policy_digest"] != pgtest.ProvisioningDigest() || proof["policy_digest"] != env[suiteservices.PGProvisioningDigestEnv] ||
		proof["template_digest"] != pgtest.TemplateDigest(pgschema.MustHash()) || proof["template_digest"] != env[suiteservices.PGTemplateDigestEnv] {
		return errors.New("incompatible PostgreSQL provisioning capability")
	}
	return nil
}

// Recovery needs only physical cleanup bindings, not build or application secrets.
func saveRecoveryEnvironment(file string, environment map[string]string) error {
	values := map[string]string{}
	for name, value := range environment {
		if name == "SSL_CERT_FILE" || name == "CARTULARY_TEST_SUITE_ID" || strings.HasPrefix(name, "CARTULARY_PGTEST_") || strings.HasPrefix(name, "CARTULARY_S3TEST_") {
			if strings.Contains(name, "PROXY_CERTIFICATE") || strings.Contains(name, "PROXY_PRIVATE_KEY") {
				continue
			}
			values[name] = value
		}
	}
	data, err := json.Marshal(values)
	if err != nil {
		return err
	}
	return harnesslocal.ReplaceFile(file, data)
}
func readRecoveryEnvironment(file string) (map[string]string, error) {
	data, err := harnesslocal.ReadFile(file, 1048576, false)
	if err != nil {
		return nil, err
	}
	values := map[string]string{}
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, err
	}
	for name := range values {
		if name != "SSL_CERT_FILE" && name != "CARTULARY_TEST_SUITE_ID" && !strings.HasPrefix(name, "CARTULARY_PGTEST_") && !strings.HasPrefix(name, "CARTULARY_S3TEST_") {
			return nil, harnesslocal.ErrUnsafe
		}
	}
	return values, nil
}
