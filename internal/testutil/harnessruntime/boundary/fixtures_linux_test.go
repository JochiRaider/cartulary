//go:build linux

package boundary

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"io"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
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
	env["CARTULARY_TEST_SERVICES_BIN"] = filepath.Join(root, ".cache/cartulary/toolbin/cartulary-test-services")
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
			// Source compilation and build recipes remain Node responsibilities.
			source := `import {writeFileSync,readFileSync} from 'node:fs';import path from 'node:path';import {createHash} from 'node:crypto';import {borrowRuntime} from './tools/harness/runtime/session-client.mjs';import {buildSourceSnapshot} from './tools/harness/test-catalog/index.mjs';import {goReadiness} from './tools/harness/readiness/installed-readiness.mjs';import {recipeEnvironment,buildReviewProducer} from './tools/harness/browser/review-recipes.mjs';const root=process.cwd(),runRoot=process.argv[1],runtime=borrowRuntime({repoRoot:root,runRoot});const input={preparation_policy:'installed_only',presentation:'design',profile:'default'};writeFileSync(path.join(runRoot,'run-manifest.json'),JSON.stringify({purpose:'interactive_design_review',source_digest:buildSourceSnapshot(root).digest,toolchain_digest:'sha256:'+createHash('sha256').update(readFileSync('tools/toolchain_pins.json')).digest('hex')}),{mode:0o600});const environment=recipeEnvironment(root,runRoot,runtime,input);environment.GO=goReadiness(environment,root);for(const target of (process.argv[2]==='borrowed'?['testservices-build']:['build-web','embedded-web-assets','build-server-harness','build-migrate','testservices-build'])) await buildReviewProducer({target,root,environment,input,runtime,signal:new AbortController().signal});process.stdout.write(JSON.stringify(environment));`
			out, err := invoke(source, run, variant)
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
			var sessionFile string
			if variant == "borrowed" {
				// A persistent suite binds exact tool bytes. Other preparation
				// owners may rebuild the workspace tool for a different Go build
				// context, so this fixture owns its executable for the whole lease.
				environment["CARTULARY_TEST_SERVICES_BIN"] = physicalFixtureExecutable(t, environment["CARTULARY_TEST_SERVICES_BIN"])
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
				owner.ConfigureAdmission("", parent, capacities, works)
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
					retained, err := borrow.Release(ctx, index != 2, false)
					if err != nil || retained != (index < 2) {
						t.Fatal("renderer settlement", retained, err)
					}
					if err := finish(ctx); err != nil {
						t.Fatal("borrow finish", err)
					}
					if err := host.Release(ctx); err != nil {
						t.Fatal("activity release", err)
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
	name := filepath.Join(t.TempDir(), "cartulary-test-services")
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
