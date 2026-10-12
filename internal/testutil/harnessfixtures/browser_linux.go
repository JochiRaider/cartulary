//go:build linux

package harnessfixtures

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

func (p ProductionOptions) browserOperation(ctx context.Context, operation string, args []string, environment map[string]string) (map[string]any, error) {
	node := p.Environment["NODE_BIN"]
	if node == "" {
		return nil, &Failure{Class: "config", Reason: "configuration_error"}
	}
	result, err := p.command()(ctx, node, append([]string{filepath.Join(p.Root, "tools/harness/browser/fixture-lifecycle.mjs"), operation}, args...), environment)
	if err != nil {
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
	value, err := harnessschema.Decode(result.Stdout)
	if err != nil {
		return nil, artifactFailure(err)
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, artifactFailure(harnesslocal.ErrUnsafe)
	}
	return object, nil
}

type browserProvider struct {
	options             ProductionOptions
	suite               *SuiteOwner
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
	suite, err := b.suite.Ensure(ctx)
	if err != nil {
		return nil, err
	}
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
		key := profile + ":" + request.FixtureProfile + ":" + request.SnapshotKey
		b.clones[key]++
		clone = b.clones[key]
	}
	b.mu.Unlock()
	session := fmt.Sprintf("%s-allocation-%03d", affinity, ordinal)
	root, err := p.Runtime.PrivatePath("browser-stack-leases", session)
	if err != nil {
		return nil, err
	}
	directory, err := harnesslocal.PrivateDirectory(root)
	if err != nil {
		return nil, err
	}
	defer func() {
		if result != nil {
			directory.Close()
		}
	}()
	environment := p.environment()
	maps.Copy(environment, suite.Environment)
	for key, value := range map[string]string{"CARTULARY_WORK_UNIT_ID": request.UnitID, "CARTULARY_BROWSER_STAGE": request.BrowserStage, "CARTULARY_BROWSER_RUNTIME_PROFILE_ID": profile, "CARTULARY_BROWSER_SERVICE_REQUIREMENT": "test-services", "CARTULARY_BROWSER_SESSION_GROUP": session, "CARTULARY_WEB_E2E_PRIVATE_SESSION_ROOT": root, "CARTULARY_HARNESS_ALLOCATION_REF": request.AllocationRef} {
		environment[key] = value
	}
	if profiled {
		for key, value := range map[string]string{"CARTULARY_FIXTURE_PROFILE_ID": request.FixtureProfile, "CARTULARY_FIXTURE_SNAPSHOT_KEY": request.SnapshotKey, "CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID": request.BuilderUnit, "CARTULARY_FIXTURE_ROW_ID": request.RowID, "CARTULARY_FIXTURE_PREDICATE_ID": request.PredicateID, "CARTULARY_FIXTURE_CLONE_LEASE_ID": request.LeaseID, "CARTULARY_FIXTURE_CLONE_ORDINAL": fmt.Sprint(clone)} {
			environment[key] = value
		}
	}
	file := filepath.Join(root, "acquisition", "owner.json")
	lease := filepath.Join(root, "stack.lease")
	envFile := filepath.Join(root, "stack.env")
	if err := p.Runtime.RecordResource("browser_stack", file, "pending"); err != nil {
		return nil, artifactFailure(err)
	}
	var once sync.Once
	var closeErr error
	close := func(ctx context.Context) error {
		once.Do(func() {
			if err := p.Runtime.AttemptCleanup(ctx, "browser_stack", file, func(step context.Context) error {
				_, err := p.browserOperation(step, "settle", []string{file}, environment)
				return err
			}); err != nil {
				closeErr = CleanupFailure(err)
				return
			}
			if err := p.Runtime.RecordResource("browser_stack", file, "released"); err != nil {
				closeErr = artifactFailure(err)
				return
			}
			closeErr = harnesssession.RemoveOwnedTree(directory, root)
			closeErr = errors.Join(closeErr, directory.Close())
		})
		return closeErr
	}
	defer func() {
		if result != nil {
			cleanup, done := context.WithTimeout(context.Background(), 2*time.Minute)
			defer done()
			result = errors.Join(result, close(cleanup))
		}
	}()
	prepared, err := p.browserOperation(ctx, "prepare", []string{suite.LeaseFile}, environment)
	if err != nil {
		return nil, err
	}
	launch, ok := prepared["launch_id"].(string)
	if !ok || len(launch) != 36 || prepared["file"] != file || len(prepared) != 2 {
		return nil, artifactFailure(harnesslocal.ErrUnsafe)
	}
	environment["CARTULARY_BROWSER_ACQUISITION_FILE"] = file
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
	return &Allocation{Ownership: "owned", ResourceIDs: ids, Environment: attached, FixtureProfile: request.FixtureProfile, SnapshotKey: request.SnapshotKey, BuilderUnit: request.BuilderUnit, CloneOrdinal: clone, Release: func(ctx context.Context, _ bool) error { return close(ctx) }}, nil
}
