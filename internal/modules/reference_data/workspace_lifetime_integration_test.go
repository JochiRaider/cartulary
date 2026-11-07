package reference_data_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	verificationtest "github.com/JochiRaider/cartulary/internal/testutil/referencetest"
)

func TestReferenceWorkspaceOperationCleanup_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	barrier := &appsupport.ReferencePackVerificationBarrier{}
	harness := startReferencePackServerWithEnv(t, runtime, "reference-workspace-lifetime", map[string]string{"CARTULARY__LIMITS__REFERENCE_PACKS__MAX_VERIFICATION_SECONDS": "60"}, barrier)
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	scratch := filepath.Join(harness.Server.Config.Roots.TemporaryWork.Path, "reference-packs", "verification")
	assertEmpty := func() {
		t.Helper()
		entries, err := os.ReadDir(scratch)
		if err != nil || len(entries) != 0 {
			t.Fatal("completed operation retained scratch", entries, err)
		}
	}
	importReferencePack(t, harness, admin, "type_registry.host", "1", "workspace-success")
	assertEmpty()
	for _, mode := range []string{"content_rejection", "cancel", "index_failure", "timeout"} {
		t.Run(mode, func(t *testing.T) {
			entered, release := make(chan struct{}), make(chan struct{})
			var fault error
			if mode == "index_failure" {
				fault = io.ErrClosedPipe
			}
			if mode != "content_rejection" {
				barrier.BlockNextIndex(entered, release, fault)
			}
			defer func() { close(release) }()
			bundle := referencePackBundle(t, bundleOptions{PackKey: "type_registry.host", PackKind: "type_registry", PackVersion: mode, BadPayloadSHA: mode == "content_rejection"})
			response := postReferencePackUpload(t, harness.Server.HTTP.URL, admin, `{"client_txn_id":"workspace-`+mode+`"}`, bundle, "workspace.zip", reference_data.MediaTypeZip)
			id := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)["job_id"].(string)
			if mode != "content_rejection" {
				select {
				case <-entered:
				case <-time.After(10 * time.Second):
					t.Fatal("no private indexing boundary")
				}
				entries, err := os.ReadDir(scratch)
				if err != nil || len(entries) == 0 {
					t.Fatal("no real scratch allocated", err)
				}
				if mode == "cancel" {
					httptestx.RequireSuccessEnvelope(t, cancelJob(t, harness, admin, id, "workspace-cancel"), http.StatusOK)
				} else if mode == "index_failure" {
					close(release)
					release = make(chan struct{})
				}
			}
			result := requireJobWithin(t, harness, admin, id, 75*time.Second)
			want := "failed"
			if mode == "cancel" {
				want = "canceled"
			}
			if result["status"] != want {
				t.Fatal(result)
			}

			if mode == "timeout" {
				summary := result["error_summary"].(map[string]any)
				if summary["code"] != "reference_pack_verification_failed" || summary["details"].(map[string]any)["reason_code"] != "verification_timeout" {
					t.Fatal("wrong owner timeout", result)
				}
			}
			assertEmpty()
			if n := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_envelopes WHERE pack_key='type_registry.host' AND pack_version=$1`, mode); n != 0 {
				t.Fatal("failed operation published content", n)
			}
		})
	}
	root, err := referenceassembly.NewRootStorage(harness.Server.Config.Roots.TemporaryWork.Path, harness.Server.Config.Roots.ReferencePackStorage.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	for _, mode := range []string{"cancel", "deadline"} {
		t.Run("historical_"+mode, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if mode == "deadline" {
				var stop context.CancelFunc
				ctx, stop = context.WithTimeout(ctx, 500*time.Millisecond)
				defer stop()
			}
			writes := 0
			storage := historicalInterruptedStorage{ArtifactStorage: root, afterWrite: func() {
				writes++
				if mode == "cancel" {
					cancel()
				} else {
					<-ctx.Done()
				}
			}}
			err := reference_data.RestoreHistoricalState(ctx, harness.Pool, storage, reference_data.DefaultLimits(), verificationtest.Lifetime(t))
			if writes == 0 || (!errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)) {
				t.Fatal("historical verification did not interrupt after extraction", writes, err)
			}

			assertEmpty()
		})
	}
}

type historicalInterruptedStorage struct {
	reference_data.ArtifactStorage
	afterWrite func()
}

func (s historicalInterruptedStorage) NewWorkspace(ctx context.Context) (reference_data.VerificationWorkspace, error) {
	w, err := s.ArtifactStorage.NewWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	return historicalInterruptedWorkspace{w, s.afterWrite}, nil
}

type historicalInterruptedWorkspace struct {
	reference_data.VerificationWorkspace
	afterWrite func()
}

func (w historicalInterruptedWorkspace) Write(ctx context.Context, name string, write func(io.Writer) error) error {
	err := w.VerificationWorkspace.Write(ctx, name, write)
	if err == nil {
		w.afterWrite()
	}
	return err
}
