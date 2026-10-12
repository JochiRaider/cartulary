//go:build linux

package harnessreview

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

func TestReviewIPC(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	environment, err := installedEnvironment(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, forged := range []bool{false, true} {
		t.Run(map[bool]string{false: "actual_private_adapter", true: "foreign_session_rejected"}[forged], func(t *testing.T) {
			run, scratch := t.TempDir(), t.TempDir()
			if err := os.Chmod(run, 0700); err != nil {
				t.Fatal(err)
			}
			runtime, err := harnesssession.NewRuntime(root, run, "fixture", scratch)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := runtime.Close(); err != nil {
					t.Error(err)
				}
			})
			session := "uireview-" + strings.Repeat("a", 32)
			worker, err := startAdapter(root, session, runtime, environment, nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			t.Cleanup(func() {
				worker.kill()
				<-worker.done
				if failures := recoverIndependent(context.Background(), root, run, runtime, environment); len(failures) > 0 {
					t.Error(failures)
				}
			})
			if forged {
				raw := `{"schema_id":"cartulary.harness_review_adapter.v1","session_id":"uireview-bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","run_id":"fixture","runtime_lease_id":"` + runtime.LeaseID + `","request_id":1,"kind":"request","operation":"close","payload":null}` + "\n"
				if _, err := worker.channel.Write([]byte(raw)); err != nil {
					t.Fatal(err)
				}
				select {
				case <-worker.done:
				case <-ctx.Done():
					t.Fatal("forged request did not close private adapter")
				}
				return
			}
			profile := map[string]any{"pins_sha256": strings.Repeat("a", 64), "lock_sha256": strings.Repeat("b", 64), "node_version": "fixture", "playwright_version": "fixture", "sharp_version": "fixture", "axe_version": "fixture"}
			reply, err := worker.call(ctx, "initialize", map[string]any{"input": reviewInput{Mode: "artifacts", Policy: "installed_only", Presentation: "ui"}, "profile": profile, "repository": root, "run_root": run, "runtime_root": runtime.Root})
			if err != nil || reply.Facts == nil || !strings.HasPrefix(reply.Facts.Source, "sha256:") {
				t.Fatal("initialize", reply.Outcome, err)
			}
			if _, err = worker.call(ctx, "validate", map[string]any{"command": "ui-review-report", "request": nil}); err != nil {
				t.Fatal(err)
			}
			var calls sync.WaitGroup
			for range 16 {
				calls.Go(func() {
					if _, err := worker.call(ctx, "validate", map[string]any{"command": "ui-review-report", "request": nil}); err != nil {
						t.Error("concurrent private request", err)
					}
				})
			}
			calls.Wait()
			if err = worker.close(ctx); err != nil {
				t.Fatal("close", err)
			}
			if failures := recoverIndependent(ctx, root, run, runtime, environment); len(failures) > 0 {
				t.Fatal(failures)
			}
			resources, err := runtime.Resources(4096)
			if err != nil || len(resources) > 0 {
				t.Fatal("producer not reaped", resources, err)
			}
		})
	}
}
