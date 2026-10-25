//go:build linux

package boundary

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Real physical ownership through the native library, without a fixture-session
// transport or a second lifecycle coordinator in the test fixture.
func TestNativeFixturePhysicalLifecycle(t *testing.T) {
	testNativeFixturePhysicalLifecycle(t, []string{"owned", "failed_cleanup"})
}

// The persistent owner reserves its own physical capacity independently of the
// borrowing test process. Keep this case outside the installed-browser envelope.
func TestNativeBorrowedFixturePhysicalLifecycle(t *testing.T) {
	testNativeFixturePhysicalLifecycle(t, []string{"borrowed"})
}

func testNativeFixturePhysicalLifecycle(t *testing.T, variants []string) {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../.."))
	env := map[string]string{}
	for _, entry := range os.Environ() {
		k, v, ok := strings.Cut(entry, "=")
		if ok {
			env[k] = v
		}
	}
	for k := range env {
		if strings.HasPrefix(k, "CARTULARY_HARNESS_") || strings.HasPrefix(k, "CARTULARY_TEST_SERVICES_") || strings.HasPrefix(k, "CARTULARY_PGTEST_") || strings.HasPrefix(k, "CARTULARY_S3TEST_") {
			delete(env, k)
		}
	}
	// The enclosing graph prepares immutable inputs once. Pin executable bytes
	// for this owner's lifetime; no variant may rebuild a shared workspace tool.
	preparedBinaries := map[string]string{}
	for _, key := range []string{"CARTULARY_TEST_SERVICES_BIN", "CARTULARY_SERVER_HARNESS_BIN", "CARTULARY_MIGRATE_BIN"} {
		if key != "CARTULARY_TEST_SERVICES_BIN" && len(variants) == 1 && variants[0] == "borrowed" {
			continue
		}
		value := os.Getenv(key)
		if value == "" {
			t.Fatalf("graph-prepared %s is required", key)
		}
		preparedBinaries[key] = physicalFixtureExecutable(t, value)
	}
	parentPreparation := map[string]string{}
	for _, key := range []string{"CARTULARY_TEST_RESULTS_DIR", "CARTULARY_TEST_RUN_ID", "CARTULARY_HARNESS_SUITE_RUNTIME_ROOT", "CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID", "CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID"} {
		parentPreparation[key] = os.Getenv(key)
	}
	parentPreparationJSON, err := json.Marshal(parentPreparation)
	if err != nil {
		t.Fatal(err)
	}
	node := env["NODE_BIN"]
	if node == "" {
		t.Fatal("Make-owned Node required")
	}
	results := os.Getenv("CARTULARY_TEST_RESULTS_DIR")
	if !filepath.IsAbs(results) {
		results = filepath.Join(root, results)
	}
	results = filepath.Join(results, os.Getenv("CARTULARY_TEST_RUN_ID"), "physical-fixtures")
	for _, variant := range variants {
		t.Run(variant, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			run := filepath.Join(results, variant)
			if err := os.MkdirAll(run, 0700); err != nil {
				t.Fatal(err)
			}
			r, err := harnesssession.NewRuntime(root, run, variant, "")
			if err != nil {
				t.Fatal(err)
			}
			// Preparation can fail before the fixture coordinator exists.
			// Always finalize the native owner created by this test.
			defer func() {
				if err := r.Close(); err != nil {
					t.Error("final private owner close", err)
				}
			}()
			environment := map[string]string{}
			for k, v := range env {
				environment[k] = v
			}
			for k, v := range r.Environment() {
				environment[k] = v
			}
			invoke := func(source string, args ...string) ([]byte, error) {
				cmd := exec.CommandContext(ctx, node, append([]string{"--input-type=module", "-e", source}, args...)...)
				cmd.Dir = root
				for k, v := range environment {
					cmd.Env = append(cmd.Env, k+"="+v)
				}
				return cmd.Output()
			}
			// Project the private review environment without running any producers.
			source := `
import assert from 'node:assert/strict';
import {cpSync,mkdtempSync,writeFileSync,readFileSync} from 'node:fs';
import path from 'node:path';
import {createHash} from 'node:crypto';
import {borrowRuntime} from './tools/harness/runtime/session-client.mjs';
import {buildSourceSnapshot} from './tools/harness/test-catalog/index.mjs';
import {loadExecutionTopology} from './tools/harness/generated-artifacts/execution-topology.mjs';
import {resolveFrontendArtifact,sealFrontendArtifact} from './tools/harness/readiness/frontend-artifact.mjs';
import {goReadiness} from './tools/harness/readiness/installed-readiness.mjs';
import {recipeEnvironment} from './tools/harness/browser/review-recipes.mjs';
const root=process.cwd(),runRoot=process.argv[1],runtime=borrowRuntime({repoRoot:root,runRoot});
const input={preparation_policy:'installed_only',presentation:'design',profile:'default'};
const identity={source_digest:buildSourceSnapshot(root).digest,toolchain_digest:'sha256:'+createHash('sha256').update(readFileSync('tools/toolchain_pins.json')).digest('hex')};
writeFileSync(path.join(runRoot,'run-manifest.json'),JSON.stringify({purpose:'interactive_design_review',...identity}),{mode:0o600});
if(process.argv[3]!=='borrowed') {
  // Validate the graph-owned immutable artifact, then materialize it in this
  // private runtime through the existing sealing boundary. Never rebuild it.
  const prepared=resolveFrontendArtifact(root,'build-web',JSON.parse(process.argv[2]));
  for(const [key,value] of Object.entries(identity)) assert.equal(prepared.receipt[key],value,'stale prepared frontend identity');
  const profile=loadExecutionTopology({root}).browserBatch.frontend_artifacts.production;
  const staging=mkdtempSync(runtime.privatePath('frontend-staging-'));
  cpSync(prepared.directory,staging,{recursive:true});
  sealFrontendArtifact({repoRoot:root,runRoot,runtime,profile:{id:'production',...profile},staging});
}
const environment=recipeEnvironment(root,runRoot,runtime,input);
// Browser startup still launches its listener helper through Go. Resolve the
// pinned installed toolchain; this readiness check does not run any producers.
environment.GO=goReadiness(environment,root);
process.stdout.write(JSON.stringify(environment));
`
			out, err := invoke(source, run, string(parentPreparationJSON), variant)
			if err != nil {
				if exit, ok := err.(*exec.ExitError); ok {
					t.Fatalf("physical fixture preparation: %v: %s", err, exit.Stderr)
				}
				t.Fatal("physical fixture preparation", err)
			}
			// The recipe returns a complete child environment. Unmarshalling
			// into the parent map would restore keys the recipe removed.
			var prepared map[string]string
			if json.Unmarshal(out, &prepared) != nil || prepared == nil {
				t.Fatal("invalid preparation facts")
			}
			if _, inherited := prepared["CARTULARY_MAKE_INPUT_SOURCES"]; inherited {
				t.Fatal("preparation retained parent command authority")
			}
			environment = prepared
			for key, value := range preparedBinaries {
				environment[key] = value
			}
			var sessionFile string
			if variant == "borrowed" {
				sessionFile = filepath.Join(r.Root, "persistent-owner", "session.json")
				source := `import {createLocalSession} from './tools/harness/services/local-session.mjs';await createLocalSession({root:process.cwd(),binary:process.env.CARTULARY_TEST_SERVICES_BIN,sessionFile:process.argv[1],environment:{CARTULARY_HOST_ADMISSION_LEASE:process.env.CARTULARY_HOST_ADMISSION_LEASE}});`
				if _, err := invoke(source, sessionFile); err != nil {
					if exit, ok := err.(*exec.ExitError); ok {
						t.Fatalf("persistent owner start: %v: %.8192s", err, exit.Stderr)
					}
					t.Fatal("persistent owner start", err)
				}
				environment["CARTULARY_TEST_SERVICES_MODE"] = "attach"
				environment["CARTULARY_TEST_SERVICES_SESSION_FILE"] = sessionFile
			}
			target := "ui-review"
			if variant == "borrowed" {
				target = "test-slice"
				environment["SCOPE"] = "services"
			}
			options := harnessfixtures.ProductionOptions{Root: root, RunRoot: run, Target: target, Runtime: r, Environment: environment, BrowserLaneLimit: 1}
			owner := harnessfixtures.NewProduction(options, harnessfixtures.Options{})
			var blocked string
			defer func() {
				if blocked != "" {
					os.Chmod(blocked, 0600)
				}
				owner.Close(context.Background())
				if err := options.Recover(context.Background()); err != nil {
					t.Error("exact recovery", err)
					r.PreserveRecovery()
					return
				}
				if sessionFile != "" {
					if _, err := invoke(`import {stopLocalSession} from './tools/harness/services/local-session.mjs';await stopLocalSession({root:process.cwd(),binary:process.env.CARTULARY_TEST_SERVICES_BIN,sessionFile:process.argv[1]});`, sessionFile); err != nil {
						t.Error("persistent owner stop", err)
					}
				}
				if err := r.Close(); err != nil {
					t.Error("private root close", err)
				}
			}()
			capability := "browser_stack"
			if variant == "borrowed" {
				capability = "postgres_transaction"
			}
			isolation := "exclusive"
			if variant == "borrowed" {
				isolation = "shared"
			}
			lease, err := owner.Broker.Acquire(ctx, harnessfixtures.Request{PostgresIsolation: isolation, Services: []string{"object_store", "postgres"}, Capability: capability, UnitID: "boundary", Affinity: "native-physical", BrowserStage: "webserver-backed", RuntimeProfile: "default"})
			if err != nil {
				t.Fatal("acquire", err)
			}
			if variant == "borrowed" {
				if lease.Record().Ownership != "borrowed" {
					t.Fatal("borrowed ownership lost")
				}
				if _, err := invoke(`import assert from 'node:assert/strict';import {stopLocalSession,localSessionStatus} from './tools/harness/services/local-session.mjs';const options={root:process.cwd(),binary:process.env.CARTULARY_TEST_SERVICES_BIN,sessionFile:process.argv[1]};assert.equal(localSessionStatus(options).borrower_count,1);await assert.rejects(()=>stopLocalSession(options),/live borrower/);`, sessionFile); err != nil {
					t.Fatal("live borrower protection", err)
				}
			}
			if variant == "owned" {
				// Reuse the actual allocation through the native broker. A planned ordinal
				// cannot satisfy this check: each returned lease must bind the same physical
				// allocation and a newly published, ready backend generation.
				var topology struct {
					Reset map[string]int `json:"browser_reset_policy"`
				}
				readSelectedBoundaryJSON(t, filepath.Join(root, "tools/execution_topology_manifest.json"), &topology)
				resetDeadline := 0
				for name, value := range topology.Reset {
					if strings.HasSuffix(name, "_timeout_ms") {
						resetDeadline += value
					}
				}
				allocationRef := lease.AllocationRef()
				leaseFile := lease.Environment()["CARTULARY_WEB_E2E_SESSION_LEASE_FILE"]
				for generation := 2; generation <= 11; generation++ {
					stopTraffic := boundaryResetTraffic(t, ctx, lease.Environment(), generation == 2)
					if _, err := lease.Release(ctx, true, true); err != nil {
						t.Fatal("retain physical stack", err)
					}
					label := fmt.Sprintf("native-consumer-%d", generation)
					lease, err = owner.Broker.Acquire(ctx, harnessfixtures.Request{PostgresIsolation: isolation, Services: []string{"object_store", "postgres"},
						Capability: capability, UnitID: label, Affinity: "native-physical", BrowserStage: "webserver-backed", RuntimeProfile: "default",
						Preparation: &harnessfixtures.BrowserPreparation{Target: "browser-e2e-webserver-backed", Label: label, DrainTimeout: fmt.Sprint(topology.Reset["backend_drain_timeout_ms"]), Timeout: fmt.Sprint(resetDeadline)}})
					if err != nil {
						t.Fatal("consuming physical preparation", err)
					}
					if lease.AllocationRef() != allocationRef {
						t.Fatal("reuse replaced the allocation")
					}
					var physical struct {
						Generation int `json:"backend_generation"`
					}
					readSelectedBoundaryJSON(t, leaseFile, &physical)
					if physical.Generation != generation {
						t.Fatal("backend generation did not advance", physical.Generation)
					}
					var attempt struct {
						Status string `json:"status"`
						Before int    `json:"backend_generation_before"`
						After  int    `json:"backend_generation_after"`
						Ready  bool   `json:"backend_ready"`
					}
					readSelectedBoundaryJSON(t, filepath.Join(run, "browser-e2e-webserver-backed", "reset-boundary", label+".attempt.json"), &attempt)
					if attempt.Status != "pass" || attempt.Before != generation-1 || attempt.After != generation || !attempt.Ready {
						t.Fatal("incomplete consuming reset", attempt)
					}
					stopTraffic()
				}
				// Exercise real connection-owned Chromium and fresh contexts through
				// two logical groups, followed by an interrupted client. The stack
				// remains an independently owned fixture throughout these borrows.
				for key, value := range lease.Environment() {
					owner.Options.Environment[key] = value
				}
				output, err := invoke(`import {captureCapabilitySnapshot,resourceCapacities} from './tools/harness/scheduler/work-graph/capability.mjs';process.stdout.write(JSON.stringify(Object.fromEntries(resourceCapacities(captureCapabilitySnapshot({root:process.cwd()})))));`)
				if err != nil {
					t.Fatal("renderer capacity", err)
				}
				capacities := map[string]int64{}
				if json.Unmarshal(output, &capacities) != nil {
					t.Fatal("renderer capacity shape")
				}
				var parent *string
				if value := environment["CARTULARY_HOST_ADMISSION_LEASE"]; value != "" {
					parent = &value
				}
				works := []harnessfixtures.WorkRequirement{}
				for _, unit := range []string{"suite-a:one", "suite-b:two", "suite-c:interrupted"} {
					works = append(works, harnessfixtures.WorkRequirement{Request: harnessfixtures.Request{PostgresIsolation: "none", UnitID: unit, Capability: "none", RuntimeProfile: "default"}, Renderer: "pooled", Physical: map[string]map[string]int64{"renderer": {"browser_stack": 1, "memory_mb": 512, "process": 2, "port_lane": 1}}, Active: map[string]int64{"cpu": 1, "io": 1, "memory_mb": 512, "process": 1}})
				}
				owner.ConfigureAdmission("", parent, capacities)
				if err := owner.RegisterConsumers(works); err != nil {
					t.Fatal(err)
				}
				generation := ""
				for index, work := range works {
					host, finish, err := owner.Admit(ctx, work)
					if err != nil {
						t.Fatal("renderer admission", err)
					}
					borrow, err := owner.Acquire(ctx, work.Request)
					if err != nil {
						t.Fatal("renderer acquisition", err)
					}
					var attachment map[string]any
					readSelectedBoundaryJSON(t, borrow.Environment["CARTULARY_RENDERER_ATTACHMENT"], &attachment)
					current := attachment["generation"].(string)
					if generation != "" && generation != current {
						t.Fatal("compatible groups did not reuse renderer generation")
					}
					generation = current
					for key, value := range borrow.Environment {
						environment[key] = value
					}
					source := `import assert from 'node:assert/strict';import {pathToFileURL} from 'node:url';import {resolvePlaywrightPackages} from './tools/harness/readiness/playwright-packages.mjs';const packages=resolvePlaywrightPackages(process.cwd());const {chromium}=await import(pathToFileURL(packages.playwrightPath+'/index.mjs'));const browser=await chromium.connect(process.env.CARTULARY_VISUAL_RENDERER_WS_ENDPOINT,{exposeNetwork:'<loopback>',timeout:30000});const context=await browser.newContext();assert.deepEqual(await context.cookies(),[]);const page=await context.newPage();await page.goto(process.env.CARTULARY_WEB_E2E_PUBLIC_ORIGIN);assert.equal(await page.evaluate(()=>localStorage.getItem('reuse-probe')),null);await page.evaluate(()=>localStorage.setItem('reuse-probe','private'));await context.addCookies([{name:'reuse-probe',value:'private',url:process.env.CARTULARY_WEB_E2E_PUBLIC_ORIGIN}]);if(process.argv[1]==='interrupt')process.exit(0);await context.close();await browser.close();`
					mode := "close"
					if index == 2 {
						mode = "interrupt"
					}
					if _, err := invoke(source, mode); err != nil {
						if exit, ok := err.(*exec.ExitError); ok {
							t.Fatalf("renderer connection: %v: %s", err, exit.Stderr)
						}
						t.Fatal(err)
					}
					retained, err := borrow.Release(ctx, index != 2)
					if err != nil || retained != (index < 2) {
						t.Fatal("renderer settlement", retained, err)
					}
					if err := finish(ctx); err != nil {
						t.Fatal("borrow finish", err)
					}
					if err := host.Release(ctx); err != nil {
						t.Fatal("activity release", err)
					}
					if err := owner.ConsumerTerminal(work.Request.UnitID, "passed"); err != nil {
						t.Fatal(err)
					}
				}
				for _, key := range []string{"CARTULARY_RENDERER_ATTACHMENT", "CARTULARY_VISUAL_RENDERER_ATTESTED", "CARTULARY_VISUAL_RENDERER_PROFILE_ID", "CARTULARY_VISUAL_RENDERER_PROFILE_JSON", "CARTULARY_VISUAL_RENDERER_WS_ENDPOINT", "CARTULARY_RENDERER_RESOURCE_REF"} {
					delete(environment, key)
				}
			}
			if variant == "failed_cleanup" {
				resources, err := r.Resources(4096)
				if err != nil {
					t.Fatal(err)
				}
				for _, resource := range resources {
					if resource.Kind == "browser_stack" {
						blocked, _ = resource.Path()
					}
				}
				if blocked == "" {
					t.Fatal("no browser proof")
				}
				// Force the initial collision that used to delete a declared
				// recovery file. The foreign reservation survives preservation
				// and recovery of this browser's unrelated owned resources.
				foreignPort := t.TempDir()
				foreignMarker := filepath.Join(foreignPort, "acquisition.json")
				foreignBytes := []byte(`{"attempt_id":"foreign"}`)
				if err := os.WriteFile(foreignMarker, foreignBytes, 0600); err != nil {
					t.Fatal(err)
				}
				if _, err := invoke(`import assert from 'node:assert/strict';import {reserveAcquisitionPort} from './tools/harness/browser/browser-acquisition.mjs';assert.equal(reserveAcquisitionPort(process.argv[1],process.argv[2],process.pid),false);`, blocked, foreignPort); err != nil {
					t.Fatal("foreign port collision", err)
				}
				if err := os.Chmod(blocked, 0644); err != nil {
					t.Fatal(err)
				}
				if err := owner.Close(ctx); err == nil {
					t.Fatal("unsafe proof authorized cleanup")
				}
				if err := r.Close(); err == nil {
					t.Fatal("unresolved state removed")
				}
				if err := os.Chmod(blocked, 0600); err != nil {
					t.Fatal(err)
				}
				detail := filepath.Join(filepath.Dir(blocked), "private-credentials.log")
				if err := os.WriteFile(detail, []byte("unnecessary private detail"), 0600); err != nil {
					t.Fatal(err)
				}
				proofFailures := func() map[string]int {
					counts := map[string]int{}
					resources, err := r.Resources(4096)
					if err != nil {
						counts["journal"]++
						return counts
					}
					for _, resource := range resources {
						files := append([]string{}, resource.RecoveryFiles...)
						if file, err := resource.Path(); err == nil {
							files = append(files, file)
						}
						for _, file := range files {
							if _, err := harnesslocal.ReadFile(file, 1048576, false); err != nil {
								counts[resource.Kind]++
							}
						}
					}
					return counts
				}
				before := proofFailures()
				if err := r.PreserveRecovery(); err != nil {
					t.Fatal("exact dependency preservation", harnesssession.DescribeCleanupFailure(err), "unavailable proof by kind before", before, "after", proofFailures())
				}
				if _, err := os.Stat(detail); !os.IsNotExist(err) {
					t.Fatal("sibling detail retained", err)
				}
				// Minimum administrator TLS proof must still authorize a fresh
				// connection after purge; ordinary runtime keys must be gone.
				var cleanupEnvironment map[string]string
				readSelectedBoundaryJSON(t, filepath.Join(filepath.Dir(filepath.Dir(blocked)), "stack-environment.json"), &cleanupEnvironment)
				adminDSN := cleanupEnvironment["CARTULARY_PGTEST_ADMIN_DSN"]
				parsed, err := url.Parse(adminDSN)
				if err != nil || parsed.Query().Get("sslmode") != "verify-full" {
					t.Fatal("missing TLS cleanup binding")
				}
				for _, key := range []string{"sslrootcert", "sslcert", "sslkey"} {
					if _, err := os.Stat(parsed.Query().Get(key)); err != nil {
						t.Fatal("cleanup credential purged", key, err)
					}
				}
				extraKey := filepath.Join(filepath.Dir(parsed.Query().Get("sslkey")), "cartulary_runtime_login.key")
				if _, err := os.Stat(extraKey); !os.IsNotExist(err) {
					t.Fatal("ordinary credential survived recovery preservation", err)
				}
				db, err := sql.Open("pgx", adminDSN)
				if err != nil {
					t.Fatal("open fresh cleanup connection", err)
				}
				pingCtx, cancelPing := context.WithTimeout(ctx, 10*time.Second)
				pingErr := db.PingContext(pingCtx)
				cancelPing()
				closeErr := db.Close()
				if pingErr != nil || closeErr != nil {
					t.Fatal("preserved cleanup credentials unusable", pingErr, closeErr)
				}
				blocked = ""
				// A completed failed attempt cannot restart its budget. Recovery opens a new owner.
				privateRoot, leaseID := r.Root, r.LeaseID
				if err := r.ReleaseHandles(); err != nil {
					t.Fatal(err)
				}
				caller, err := harnessadmission.Identity(os.Getpid())
				if err != nil {
					t.Fatal(err)
				}
				r, err = harnesssession.OwnRuntime(root, run, privateRoot, variant, leaseID, caller, true)
				if err != nil {
					t.Fatal(err)
				}
				options.Runtime = r
				snapshot := func() map[string][32]byte {
					files := map[string][32]byte{}
					if err := filepath.WalkDir(run, func(path string, entry fs.DirEntry, err error) error {
						if err != nil {
							return err
						}
						if entry.IsDir() {
							return nil
						}
						data, err := os.ReadFile(path)
						if err != nil {
							return err
						}
						files[path] = sha256.Sum256(data)
						return nil
					}); err != nil {
						t.Fatal(err)
					}
					return files
				}
				original := snapshot()
				if err := options.Recover(ctx); err != nil {
					t.Fatal("recovery", err)
				}
				if !reflect.DeepEqual(original, snapshot()) {
					t.Fatal("physical recovery changed original evidence")
				}
				if actual, err := os.ReadFile(foreignMarker); err != nil || string(actual) != string(foreignBytes) {
					t.Fatal("foreign port changed during recovery", err)
				}
			} else {
				if err := owner.Close(ctx); err != nil {
					t.Fatal("close", err)
				}
				if err := owner.Close(ctx); err != nil {
					t.Fatal("repeated close", err)
				}
			}
			resources, err := r.Resources(4096)
			if err != nil || len(resources) != 0 {
				t.Fatal("unresolved resources", len(resources), err)
			}
			if variant == "borrowed" {
				out, err := invoke(`import {localSessionStatus} from './tools/harness/services/local-session.mjs';const state=localSessionStatus({root:process.cwd(),binary:process.env.CARTULARY_TEST_SERVICES_BIN,sessionFile:process.argv[1]});process.stdout.write(JSON.stringify({state:state.state,reason:state.reason??'',compatible:state.compatible,borrowers:state.borrower_count,services:state.services.length,running:state.services.filter(service=>service.running).length}));`, sessionFile)
				if err != nil {
					t.Fatal("borrowed suite status query", err)
				}
				var state struct {
					State, Reason                string
					Compatible                   bool
					Borrowers, Services, Running int
				}
				if json.Unmarshal(out, &state) != nil {
					t.Fatal("invalid borrowed suite status projection")
				}
				if state.State != "active" || !state.Compatible || state.Borrowers != 0 || state.Services == 0 || state.Running != state.Services {
					t.Fatalf("borrowed suite did not survive: state=%s reason=%s compatible=%t borrowers=%d services=%d running=%d", state.State, state.Reason, state.Compatible, state.Borrowers, state.Services, state.Running)
				}
			}
		})
	}
}

func physicalFixtureExecutable(t *testing.T, source string) string {
	t.Helper()
	input, err := os.Open(source)
	if err != nil {
		t.Fatal("open prepared fixture executable", err)
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
		t.Fatal("prepared fixture executable is not a regular executable", err)
	}
	name := filepath.Join(t.TempDir(), filepath.Base(source))
	output, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0700)
	if err != nil {
		t.Fatal("create owned fixture executable", err)
	}
	size, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil || closeErr != nil || size != info.Size() {
		t.Fatal("copy prepared fixture executable", copyErr, closeErr)
	}
	return name
}

// Keep all eight request workers occupied across each backend replacement. This
// bounded lifecycle probe tests drain/readiness under traffic; it is not a
// throughput measurement and does not modify host capacity or reset deadlines.
func boundaryResetTraffic(t *testing.T, parent context.Context, environment map[string]string, initial bool) func() {
	t.Helper()
	certificate, err := harnesslocal.ReadFile(environment["CARTULARY_WEB_E2E_TLS_ROOT_CERTIFICATE"], 256*1024, false)
	if err != nil {
		t.Fatal("reset traffic trust", err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(certificate) {
		t.Fatal("reset traffic trust is invalid")
	}
	token, err := harnesslocal.ReadFile(environment["CARTULARY_TEST_ROUTE_TOKEN_FILE"], 4096, false)
	if err != nil {
		t.Fatal("reset traffic token", err)
	}
	transport := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS13}, MaxConnsPerHost: 8, MaxIdleConnsPerHost: 8}
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}
	ctx, cancel := context.WithCancel(parent)
	request := func(ctx context.Context) bool {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, environment["CARTULARY_WEB_E2E_API_ORIGIN"]+"/api/v1/test/runtime/identity", nil)
		if err != nil {
			return false
		}
		req.Header.Set("Origin", environment["CARTULARY_WEB_E2E_PUBLIC_ORIGIN"])
		req.Header.Set("X-Cartulary-Test-Route-Token", strings.TrimSpace(string(token)))
		response, err := client.Do(req)
		if err != nil {
			return false
		}
		_, readErr := io.Copy(io.Discard, io.LimitReader(response.Body, 65536))
		closeErr := response.Body.Close()
		return response.StatusCode == http.StatusOK && readErr == nil && closeErr == nil
	}
	var workers sync.WaitGroup
	var once sync.Once
	stop := func() { once.Do(func() { cancel(); workers.Wait(); transport.CloseIdleConnections() }) }
	t.Cleanup(stop)
	// Arm through the installed HTTP boundary. A duplicate must conflict before
	// reset; the identical request must succeed only after backend replacement.
	// Leave it armed to prove isolation again on the next consuming reset.
	controls := []struct{ route, body string }{
		{"network-flow-faults", `{"boundary":"network_flow.import.before_transaction_commit","fault_kind":"return_error","error_code":"network_flow_reset_probe","consume_once":true}`},
		{"network-flow-randomness", `{"stream":"network_flow.table_id","value_kind":"uuid","values":["01234567-89ab-cdef-0123-456789abcdef"],"consume_once":true,"exhaustion":"fail_closed"}`},
	}
	armControls := func(fresh bool) {
		t.Helper()
		for _, control := range controls {
			for attempt := range 2 {
				req, err := http.NewRequestWithContext(parent, http.MethodPost, environment["CARTULARY_WEB_E2E_API_ORIGIN"]+"/api/v1/test/runtime/"+control.route, strings.NewReader(control.body))
				if err != nil {
					t.Fatal("reset control request", err)
				}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Origin", environment["CARTULARY_WEB_E2E_PUBLIC_ORIGIN"])
				req.Header.Set("X-Cartulary-Test-Route-Token", strings.TrimSpace(string(token)))
				response, err := client.Do(req)
				if err != nil {
					t.Fatal("reset control transport", err)
				}
				_, readErr := io.Copy(io.Discard, io.LimitReader(response.Body, 65536))
				closeErr := response.Body.Close()
				expected := http.StatusConflict
				if fresh && attempt == 0 {
					expected = http.StatusCreated
				}
				if response.StatusCode != expected || readErr != nil || closeErr != nil {
					t.Fatalf("reset control %s: status=%d want=%d read=%v close=%v", control.route, response.StatusCode, expected, readErr, closeErr)
				}
			}
		}
	}
	armControls(initial)
	started := make(chan bool, 8)
	for range 8 {
		workers.Go(func() {
			first := true
			for ctx.Err() == nil {
				ok := request(ctx)
				if first {
					started <- ok
					first = false
				}
				if !ok {
					select {
					case <-ctx.Done():
						return
					case <-time.After(10 * time.Millisecond):
					}
				}
			}
		})
	}
	for range 8 {
		select {
		case ok := <-started:
			if !ok {
				t.Fatal("reset traffic did not reach the ready backend")
			}
		case <-ctx.Done():
			t.Fatal("reset traffic startup cancelled")
		}
	}
	return func() {
		// Check the replacement before disposing the load clients. Transient errors
		// during intentional termination are load observations, never reset retries.
		ready := request(parent)
		armControls(true)
		stop()
		if !ready {
			t.Fatal("reset replacement is unavailable under bounded traffic")
		}
	}
}
