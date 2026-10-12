//go:build linux

package harnessreview

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"testing/synctest"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

func preparationStallHelper(t *testing.T, stage string) {
	signal.Ignore(syscall.SIGTERM, syscall.SIGINT)
	if stage != "descendant" {
		executable, _ := os.Executable()
		child := exec.Command(executable, "-test.run=^TestReviewIPC$")
		child.Env = append(os.Environ(), "CARTULARY_REVIEW_STALL_HELPER=descendant")
		if err := child.Start(); err != nil {
			t.Fatal(err)
		}
	}
	proof, err := harnessadmission.Identity(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(proof)
	if err := os.WriteFile(filepath.Join(os.Getenv("CARTULARY_REVIEW_STALL_DIRECTORY"), stage+".json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	// A live, wholly unresponsive adapter, including an inherited descendant.
	for {
		time.Sleep(time.Hour)
	}
}

func testPreparationDeadlines(t *testing.T) {
	t.Run("independent_producer_clock", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			d := &productionDriver{}
			operations := append([]string{"initialize"}, preparationOperations("ensure")...)
			operations = append(operations, "seed", "prepare_browser")
			for _, operation := range operations {
				before := time.Now()
				ctx, cancel := d.producerContext(context.Background())
				<-ctx.Done()
				if time.Since(before) != 300*time.Second {
					t.Fatal("producer did not receive its own 300 seconds", operation)
				}
				var classified *ReviewError
				if !errors.As(preparationCancellation(ctx, operation), &classified) || classified.Record.Code != "operation_expired" || (classified.Record.Condition == nil || *classified.Record.Condition != "timeout") {
					t.Fatal("deadline classification", operation)
				}
				cancel()
			}
		})
	})
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	environment, err := installedEnvironment(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"seeded", "dev", "artifacts", "design"} {
		t.Run("native_initialization_deadline_"+mode, func(t *testing.T) {
			run, scratch := t.TempDir(), t.TempDir()
			runtime, err := harnesssession.NewRuntime(root, run, "fixture", scratch)
			if err != nil {
				t.Fatal(err)
			}
			executable, _ := os.Executable()
			shim := filepath.Join(t.TempDir(), "node")
			quoted := "'" + strings.ReplaceAll(executable, "'", "'\"'\"'") + "'"
			if err := os.WriteFile(shim, []byte("#!/bin/sh\nexec "+quoted+" -test.run='^TestReviewIPC$'\n"), 0700); err != nil {
				t.Fatal(err)
			}
			env := map[string]string{}
			for k, v := range environment {
				env[k] = v
			}
			env["NODE_BIN"] = shim
			env["CARTULARY_REVIEW_STALL_HELPER"] = "adapter"
			env["CARTULARY_REVIEW_STALL_DIRECTORY"] = t.TempDir()
			input := reviewInput{Mode: mode, Policy: "installed_only", Presentation: "ui"}
			if mode == "design" {
				input.Mode = "seeded"
				input.Policy = "ensure"
				input.Presentation = "design"
			}
			d := &productionDriver{root: root, runRoot: run, session: "uireview-" + strings.Repeat("a", 32), runtime: runtime, environment: env, input: input,
				profile: map[string]any{"pins_sha256": strings.Repeat("a", 64), "lock_sha256": strings.Repeat("b", 64), "node_version": "fixture", "playwright_version": "fixture", "sharp_version": "fixture", "axe_version": "fixture"},
				deadline: func(ctx context.Context, budget time.Duration) (context.Context, context.CancelFunc) {
					if budget != 300*time.Second {
						t.Error("wrong native producer policy")
					}
					return context.WithTimeout(ctx, 100*time.Millisecond)
				}}
			_, err = d.Prepare(context.Background())
			var classified *ReviewError
			if !errors.As(err, &classified) || classified.Record.Code != "operation_expired" {
				t.Error("initialization escaped native deadline", err)
			}
			d.environment = environment
			if failures := d.Close(context.Background()); len(failures) > 0 {
				t.Error("intentional termination became cleanup failure", failures)
			}
			if err := runtime.Close(); err != nil {
				t.Error(err)
			}
		})
	}
	for _, operation := range []string{"initialize", "build-web", "seed", "prepare_browser"} {
		t.Run("unresponsive_"+operation, func(t *testing.T) {
			run, scratch := t.TempDir(), t.TempDir()
			runtime, err := harnesssession.NewRuntime(root, run, "fixture", scratch)
			if err != nil {
				t.Fatal(err)
			}
			executable, _ := os.Executable()
			shim := filepath.Join(t.TempDir(), "node")
			quoted := "'" + strings.ReplaceAll(executable, "'", "'\"'\"'") + "'"
			if err := os.WriteFile(shim, []byte("#!/bin/sh\nexec "+quoted+" -test.run='^TestReviewIPC$'\n"), 0700); err != nil {
				t.Fatal(err)
			}
			env := map[string]string{}
			for k, v := range environment {
				env[k] = v
			}
			env["NODE_BIN"] = shim
			env["CARTULARY_REVIEW_STALL_HELPER"] = "adapter"
			env["CARTULARY_REVIEW_STALL_DIRECTORY"] = t.TempDir()
			a, err := startAdapter(root, "uireview-"+strings.Repeat("a", 32), runtime, env, nil)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				a.kill()
				<-a.done
				if errs := (harnessfixtures.ProductionOptions{Root: root, RunRoot: run, Runtime: runtime, Environment: environment}).Recover(context.Background()); errs != nil {
					t.Error(errs)
				}
				if err := runtime.Close(); err != nil {
					t.Error(err)
				}
			})
			deadline := time.Now().Add(10 * time.Second)
			var descendant harnessadmission.Proof
			for {
				data, e := os.ReadFile(filepath.Join(env["CARTULARY_REVIEW_STALL_DIRECTORY"], "descendant.json"))
				if e == nil && json.Unmarshal(data, &descendant) == nil {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("helper did not initialize")
				}
				time.Sleep(10 * time.Millisecond)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
			defer cancel()
			payload := any(map[string]any{"admission_token": strings.Repeat("a", 32)})
			if operation == "initialize" {
				payload = map[string]any{"input": reviewInput{Mode: "artifacts", Policy: "installed_only", Presentation: "ui"}, "profile": map[string]any{}, "repository": root, "run_root": run, "runtime_root": runtime.Root}
			}
			if operation == "seed" {
				payload = map[string]any{"environment": map[string]string{}}
			}
			// initialize's profile must satisfy the wire schema even though it is never read.
			if operation == "initialize" {
				payload.(map[string]any)["profile"] = map[string]any{"pins_sha256": strings.Repeat("a", 64), "lock_sha256": strings.Repeat("b", 64), "node_version": "fixture", "playwright_version": "fixture", "sharp_version": "fixture", "axe_version": "fixture"}
			}
			started := time.Now()
			_, err = a.call(ctx, operation, payload)
			var classified *ReviewError
			if !errors.As(err, &classified) || classified.Record.Code != "operation_expired" {
				t.Fatal("unresponsive adapter misclassified", err)
			}
			if time.Since(started) > 6*time.Second {
				t.Fatal("native grace exceeded")
			}
			for {
				alive, e := harnessadmission.Alive(descendant)
				if e != nil {
					t.Fatal(e)
				}
				if !alive {
					break
				}
				if time.Since(started) > 6*time.Second {
					t.Fatal("descendant survived native termination")
				}
				time.Sleep(10 * time.Millisecond)
			}
			select {
			case <-a.done:
			default:
				t.Fatal("adapter not reaped")
			}
		})
	}
}
