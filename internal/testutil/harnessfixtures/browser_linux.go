//go:build linux

package harnessfixtures

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/JochiRaider/cartulary/internal/testutil/browserpki"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

func (p ProductionOptions) browserOperation(ctx context.Context, operation string, args []string, environment map[string]string) (map[string]any, error) {
	return p.physicalOperation(ctx, "fixture-lifecycle.mjs", operation, args, environment)
}
func (p ProductionOptions) physicalOperation(ctx context.Context, script, operation string, args []string, environment map[string]string) (value map[string]any, resultErr error) {
	defer func() { resultErr = harnesssession.AtCleanupStage("physical_command", resultErr) }()
	node := p.Environment["NODE_BIN"]
	if node == "" {
		return nil, &Failure{Class: "config", Reason: "configuration_error"}
	}
	result, err := p.command()(ctx, node, append([]string{filepath.Join(p.Root, "tools/harness/browser", script), operation}, args...), environment)
	if err != nil {
		if diagnostic := harnesssession.DecodeCleanupDiagnostic(result.Stderr); diagnostic != nil {
			err = errors.Join(err, diagnostic)
		}
		if value, decodeErr := harnessschema.Decode(result.Stdout); decodeErr == nil {
			if failure, ok := value.(map[string]any); ok && len(failure) == 2 {
				class, cok := failure["failure_class"].(string)
				reason, rok := failure["failure_reason"].(string)
				if cok && rok {
					return nil, &Failure{Class: class, Reason: reason, Cause: err}
				}
			}
		}
		if result.Exit == 11 {
			return nil, artifactFailure(err)
		}
		return nil, fixtureFailure(err)
	}
	decoded, err := harnessschema.Decode(result.Stdout)
	if err != nil {
		return nil, artifactFailure(err)
	}
	object, ok := decoded.(map[string]any)
	if !ok {
		return nil, artifactFailure(harnesslocal.ErrUnsafe)
	}
	return object, nil
}

type browserProvider struct {
	trust               browserpki.Owner
	options             ProductionOptions
	services            func(context.Context, Request) (*Suite, error)
	releaseServices     func(context.Context, Request) error
	mu                  sync.Mutex
	allocations, clones map[string]int
}

func (b *browserProvider) Close(context.Context) error { return nil }
func (b *browserProvider) Acquire(ctx context.Context, request Request) (allocation *Allocation, result error) {
	p := b.options
	profiled := request.FixtureProfile != "" || request.SnapshotKey != "" || request.BuilderUnit != ""
	if profiled && (request.FixtureProfile == "" || request.SnapshotKey == "" || request.BuilderUnit == "" || request.RowID == "" || request.PredicateID == "" || request.LeaseID == "") {
		return nil, fixtureFailure(errors.New("incomplete snapshot identity"))
	}
	// Transfer acquired ownership even on startup failure. The broker must
	// retain this allocation until its single cleanup attempt settles it,
	// including a service generation that failed before browser preparation.
	var closeBrowser func(context.Context) error
	partial := &Allocation{Ownership: "owned", ResourceIDs: []string{}, Release: func(ctx context.Context, _ bool) error {
		if closeBrowser != nil {
			if err := closeBrowser(ctx); err != nil {
				return err
			}
		}
		if b.releaseServices != nil {
			return b.releaseServices(ctx, request)
		}
		return nil
	}}
	defer func() {
		if result != nil {
			allocation = partial
		}
	}()
	suite, err := b.services(ctx, request)
	if err != nil {
		return nil, err
	}
	p.Dependencies = slices.Clone(suite.ResourceRefs)
	affinity := safeName.ReplaceAllString(request.Affinity, "-")
	if affinity == "" {
		affinity = "browser"
	}
	profile := request.RuntimeProfile
	if profile == "" {
		profile = "default"
	}
	b.mu.Lock()
	b.allocations[affinity]++
	ordinal := b.allocations[affinity]
	clone := 0
	if profiled {
		key := clusterKey(request)
		b.clones[key]++
		clone = b.clones[key]
	}
	b.mu.Unlock()
	session := fmt.Sprintf("%s-allocation-%03d", affinity, ordinal)
	root, err := p.Runtime.RecoveryPath("browser-stack-leases", session)
	if err != nil {
		return nil, err
	}
	directory, err := harnesslocal.PrivateDirectory(root)
	if err != nil {
		return nil, err
	}
	closeBrowser = func(context.Context) error {
		return errors.Join(harnesssession.RemoveOwnedTree(directory, root), directory.Close())
	}
	environment := serviceEnvironment(p.environment(), suite.Environment)
	for key, value := range map[string]string{"CARTULARY_WORK_UNIT_ID": request.UnitID, "CARTULARY_BROWSER_STAGE": request.BrowserStage, "CARTULARY_BROWSER_RUNTIME_PROFILE_ID": profile, "CARTULARY_BROWSER_SERVICE_REQUIREMENT": "test-services", "CARTULARY_BROWSER_SESSION_GROUP": session, "CARTULARY_WEB_E2E_PRIVATE_SESSION_ROOT": root, "CARTULARY_HARNESS_ALLOCATION_REF": request.AllocationRef} {
		environment[key] = value
	}
	if profiled {
		for key, value := range map[string]string{"CARTULARY_FIXTURE_PROFILE_ID": request.FixtureProfile, "CARTULARY_FIXTURE_SNAPSHOT_KEY": request.SnapshotKey, "CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID": request.BuilderUnit, "CARTULARY_FIXTURE_ROW_ID": request.RowID, "CARTULARY_FIXTURE_PREDICATE_ID": request.PredicateID, "CARTULARY_FIXTURE_CLONE_LEASE_ID": request.LeaseID, "CARTULARY_FIXTURE_CLONE_ORDINAL": fmt.Sprint(clone)} {
			environment[key] = value
		}
	}
	file := filepath.Join(root, "acquisition", "owner.json")
	resourceRef, err := p.Runtime.ResourceReference("browser_stack", file)
	if err != nil {
		return nil, err
	}
	p.Dependencies = append(slices.Clone(suite.ResourceRefs), resourceRef)
	lease := filepath.Join(root, "stack.lease")
	envFile := filepath.Join(root, "stack.env")
	dependencies := append(slices.Clone(suite.ResourceRefs), p.residency.reference(stackKey(request))...)
	recoveryEnvironment := filepath.Join(root, "stack-environment.json")
	if err := saveRecoveryEnvironment(recoveryEnvironment, environment); err != nil {
		return nil, err
	}
	var once sync.Once
	var closeErr error
	closeBrowser = func(ctx context.Context) error {
		once.Do(func() {
			closeErr = p.closeResource(ctx, "browser_stack", file, func(ctx context.Context) error {
				if err := p.Runtime.AttemptCleanup(ctx, "browser_stack", file, func(step context.Context) error {
					_, err := p.browserOperation(step, "settle", []string{file}, environment)
					return err
				}); err != nil {
					return CleanupFailure(err)
				}
				if err := p.Runtime.RecordResource("browser_stack", file, "released"); err != nil {
					return artifactFailure(err)
				}
				return errors.Join(harnesssession.RemoveOwnedTree(directory, root), p.residency.close(ctx, stackKey(request)))
			})
			// A failed cleanup is terminal for this handle. Disk proof remains
			// journal-owned for explicit recovery; no descriptor needs to leak.
			closeErr = errors.Join(closeErr, directory.Close())
		})
		return closeErr
	}
	if err := p.Runtime.RecordDependentResource("browser_stack", file, "pending", dependencies, suite.LeaseFile, recoveryEnvironment); err != nil {
		return nil, artifactFailure(err)
	}
	prepared, err := p.browserOperation(ctx, "prepare", []string{suite.LeaseFile}, environment)
	if err != nil {
		return nil, err
	}
	launch, ok := prepared["launch_id"].(string)
	if !ok || len(launch) != 36 || prepared["file"] != file || len(prepared) != 2 {
		return nil, artifactFailure(harnesslocal.ErrUnsafe)
	}
	environment["CARTULARY_BROWSER_ACQUISITION_FILE"] = file
	// Issuance belongs to the native physical owner, before any application
	// child starts. The shell consumes these files but cannot mint identities.
	tlsParent, err := harnesslocal.PrivateDirectory(filepath.Join(root, "runtime-root"))
	if err != nil {
		return nil, err
	}
	defer tlsParent.Close()
	objectRoot, err := harnesslocal.ReadFile(suite.Environment["SSL_CERT_FILE"], 256*1024, false)
	if err != nil {
		return nil, err
	}
	trustScope := profile
	if profiled || p.Target == "ui-review" {
		trustScope += ":dedicated:" + session
	}
	if err := b.trust.Provision(filepath.Join(root, "runtime-root", "tls"), trustScope, objectRoot); err != nil {
		return nil, err
	}
	if _, err := p.browserOperation(ctx, "start", []string{file, launch}, environment); err != nil {
		return nil, err
	}

	values, err := p.browserOperation(ctx, "validate", []string{envFile}, environment)
	if err != nil {
		return nil, err
	}
	attached := map[string]string{}
	for key, value := range suite.Environment {
		if key == "CARTULARY_TEST_SERVICES_CALL_MODE" || key == "CARTULARY_TEST_SERVICES_LIFECYCLE_MODE" || key == "CARTULARY_TEST_SUITE_ID" || key == "SSL_CERT_FILE" || strings.HasPrefix(key, "CARTULARY_PGTEST_") || strings.HasPrefix(key, "CARTULARY_S3TEST_") {
			attached[key] = value
		}
	}
	for key, value := range values {
		text, ok := value.(string)
		if !ok {
			return nil, artifactFailure(harnesslocal.ErrUnsafe)
		}
		attached[key] = text
	}
	attached["CARTULARY_BROWSER_STAGE"] = request.BrowserStage
	attached["CARTULARY_BROWSER_SESSION_GROUP"] = session
	attached["CARTULARY_WEB_E2E_SESSION_LEASE_FILE"] = lease
	attached["CARTULARY_HARNESS_RESOURCE_DEPENDENCIES"] = resourceDependencies([]string{resourceRef})
	if err := p.Runtime.RegisterEnvironment(attached); err != nil {
		return nil, artifactFailure(err)
	}
	if err := removePrivate(envFile); err != nil {
		return nil, artifactFailure(err)
	}
	if err := p.Runtime.RecordResource("browser_stack", file, "acquired"); err != nil {
		return nil, artifactFailure(err)
	}
	ids := []string{"browser-stack:" + session}
	if profiled {
		ids = append(ids, fmt.Sprintf("fixture-clone:%s:%d", request.SnapshotKey, clone))
	}
	partial.ResourceIDs, partial.Environment = ids, attached
	partial.FixtureProfile, partial.SnapshotKey, partial.BuilderUnit, partial.CloneOrdinal = request.FixtureProfile, request.SnapshotKey, request.BuilderUnit, clone
	return partial, nil
}
