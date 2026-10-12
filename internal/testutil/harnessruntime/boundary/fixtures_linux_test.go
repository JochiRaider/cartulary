//go:build linux

package boundary

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

// Real physical ownership through the native library, without a fixture-session
// transport or a second lifecycle coordinator in the test fixture.
func TestNativeFixturePhysicalLifecycle(t *testing.T) {
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
	for _, variant := range []string{"borrowed", "owned", "failed_cleanup"} {
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
			var sessionFile string
			if variant == "borrowed" {
				sessionFile = filepath.Join(r.Root, "persistent-owner", "session.json")
				source := `import {createLocalSession} from './tools/harness/services/local-session.mjs';createLocalSession({root:process.cwd(),binary:process.env.CARTULARY_TEST_SERVICES_BIN,sessionFile:process.argv[1],environment:{}});`
				if _, err := invoke(source, sessionFile); err != nil {
					t.Fatal("persistent owner start", err)
				}
				environment["CARTULARY_TEST_SERVICES_MODE"] = "attach"
				environment["CARTULARY_TEST_SERVICES_SESSION_FILE"] = sessionFile
			} else {
				// Source compilation and build recipes remain Node responsibilities.
				source := `import {writeFileSync,readFileSync} from 'node:fs';import path from 'node:path';import {createHash} from 'node:crypto';import {borrowRuntime} from './tools/harness/runtime/session-client.mjs';import {buildSourceSnapshot} from './tools/harness/test-catalog/index.mjs';import {recipeEnvironment,buildReview} from './tools/harness/browser/review-recipes.mjs';const root=process.cwd(),runRoot=process.argv[1],runtime=borrowRuntime({repoRoot:root,runRoot});const input={preparation_policy:'installed_only',presentation:'design',profile:'default'};writeFileSync(path.join(runRoot,'run-manifest.json'),JSON.stringify({purpose:'interactive_design_review',source_digest:buildSourceSnapshot(root).digest,toolchain_digest:'sha256:'+createHash('sha256').update(readFileSync('tools/toolchain_pins.json')).digest('hex')}),{mode:0o600});const environment=recipeEnvironment(root,runRoot,runtime,input);await buildReview({root,environment,input,runtime,signal:new AbortController().signal});process.stdout.write(JSON.stringify(environment));`
				out, err := invoke(source, run)
				if err != nil {
					t.Fatal("physical fixture preparation", err)
				}
				if json.Unmarshal(out, &environment) != nil {
					t.Fatal("invalid preparation facts")
				}
			}
			target := "ui-review"
			if variant == "borrowed" {
				target = "service-backed-test-slice"
			}
			options := harnessfixtures.ProductionOptions{Root: root, RunRoot: run, Target: target, Runtime: r, Environment: environment}
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
					if _, err := invoke(`import {stopLocalSession} from './tools/harness/services/local-session.mjs';stopLocalSession({root:process.cwd(),binary:process.env.CARTULARY_TEST_SERVICES_BIN,sessionFile:process.argv[1]});`, sessionFile); err != nil {
						t.Error("persistent owner stop", err)
					}
				}
				if err := r.Close(); err != nil {
					t.Error("private root close", err)
				}
			}()
			capability := "browser_stack"
			if variant == "borrowed" {
				capability = "postgres_dedicated"
			}
			lease, err := owner.Broker.Acquire(ctx, harnessfixtures.Request{Capability: capability, UnitID: "boundary", Affinity: "native-physical", BrowserStage: "webserver-backed", RuntimeProfile: "default"})
			if err != nil {
				t.Fatal("acquire", err)
			}
			if variant == "borrowed" {
				if lease.Record().Ownership != "borrowed" {
					t.Fatal("borrowed ownership lost")
				}
				if _, err := invoke(`import assert from 'node:assert/strict';import {stopLocalSession,localSessionStatus} from './tools/harness/services/local-session.mjs';const options={root:process.cwd(),binary:process.env.CARTULARY_TEST_SERVICES_BIN,sessionFile:process.argv[1]};assert.equal(localSessionStatus(options).borrower_count,1);assert.throws(()=>stopLocalSession(options),/live borrower/);`, sessionFile); err != nil {
					t.Fatal("live borrower protection", err)
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
				blocked = ""
				if err := options.Recover(ctx); err != nil {
					t.Fatal("recovery", err)
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
				if _, err := invoke(`import assert from 'node:assert/strict';import {localSessionStatus} from './tools/harness/services/local-session.mjs';const state=localSessionStatus({root:process.cwd(),binary:process.env.CARTULARY_TEST_SERVICES_BIN,sessionFile:process.argv[1]});assert.equal(state.state,'active');assert.equal(state.borrower_count,0);assert.ok(state.services.length>0);`, sessionFile); err != nil {
					t.Fatal("borrowed suite did not survive", err)
				}
			}
		})
	}
}
