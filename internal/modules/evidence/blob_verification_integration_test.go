package evidence_test

import (
	"context"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// This is a controlled I/O fault test, not product latency evidence.
func TestEvidenceVerificationCancelsAndDrainsStoredRead_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	base, err := objectstore.NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = base.Close() })
	store := &verificationBarrierStore{FilesystemStore: base, entered: make(chan struct{}), drained: make(chan struct{})}
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "evidence-verification-cancel", ObjectStore: store, TestRouteMode: httptestx.TestRouteModeDisabled})
	actor, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, harness.Server, actor, map[string]any{"incident_key": "EVIDENCE-CANCEL", "title": "Evidence verification cancellation", "client_txn_id": "incident-cancel"})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	post := func(path string, body any) *http.Response {
		t.Helper()
		request := httptestx.NewJSONRequest(t, http.MethodPost, harness.Server.HTTP.URL+path, body).WithContext(ctx)
		for _, option := range authOptions(actor) {
			option(request)
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		return response
	}
	slot := httptestx.RequireSuccessEnvelope(t, post("/api/v1/object-blobs", map[string]any{"incident_id": incident["incident_id"], "client_txn_id": "slot-cancel", "byte_size": 3, "content_type_hint": "text/plain"}), http.StatusCreated)["data"].(map[string]any)
	putObject(t, harness.Server.HTTP.URL, slot["upload_target"].(map[string]any)["href"].(string), []byte("abc"), "text/plain", actor)
	store.armed.Store(true)
	body := map[string]any{"client_txn_id": "verify-cancel"}
	route := "/api/v1/object-blobs/" + slot["object_blob_id"].(string) + "/verify"
	job := httptestx.RequireSuccessEnvelope(t, post(route, body), http.StatusAccepted)["data"].(map[string]any)
	select {
	case <-store.entered:
	case <-ctx.Done():
		t.Fatal("worker did not enter the controlled stored read")
	}
	replay := httptestx.RequireSuccessEnvelope(t, post(route, body), http.StatusAccepted)["data"].(map[string]any)
	if replay["job_id"] != job["job_id"] {
		t.Fatal("admission replay allocated another Job")
	}
	// No blob or Jobs transition lock may remain held across the byte stream.
	tx, err := harness.Pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	if _, err = tx.Exec(ctx, `SELECT object_blob_id FROM object_blobs WHERE object_blob_id=$1 FOR UPDATE`, slot["object_blob_id"]); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, `SELECT job_id FROM jobs WHERE job_id=$1 FOR UPDATE`, job["job_id"]); err != nil {
		t.Fatal(err)
	}
	if err = tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	httptestx.RequireSuccessEnvelope(t, post("/api/v1/jobs/"+job["job_id"].(string)+"/cancel", map[string]any{"client_txn_id": "cancel-read"}), http.StatusOK)
	for {
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+job["status_route"].(string), nil, authOptions(actor)...)
		observed := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
		if observed["status"] == "canceled" {
			break
		}
		if observed["status"] == "failed" || observed["status"] == "succeeded" {
			t.Fatalf("unexpected terminal verification: %v", observed["status"])
		}
		select {
		case <-ctx.Done():
			t.Fatal("verification cancellation did not complete")
		case <-time.After(10 * time.Millisecond):
		}
	}
	select {
	case <-store.drained:
	default:
		t.Fatal("Job became canceled before its read drained")
	}
	var proofs, associations int
	var state string
	if err = harness.DB.QueryRowContext(ctx, `SELECT upload_state,(SELECT count(*) FROM evidence_blob_verifications WHERE object_blob_id=$1),(SELECT count(*) FROM evidence WHERE object_blob_id=$1) FROM object_blobs WHERE object_blob_id=$1`, slot["object_blob_id"]).Scan(&state, &proofs, &associations); err != nil {
		t.Fatal(err)
	}
	if state != "pending" || proofs != 0 || associations != 0 {
		t.Fatalf("cancellation leaked state=%s proofs=%d associations=%d", state, proofs, associations)
	}
}

type verificationBarrierStore struct {
	*objectstore.FilesystemStore
	armed            atomic.Bool
	entered, drained chan struct{}
}

func (s *verificationBarrierStore) Get(ctx context.Context, request objectstore.GetObjectRequest) (io.ReadCloser, objectstore.ObjectInfo, error) {
	reader, info, err := s.FilesystemStore.Get(ctx, request)
	if err != nil || !s.armed.Load() {
		return reader, info, err
	}
	return &verificationBarrierReader{ReadCloser: reader, closed: make(chan struct{}), entered: s.entered, drained: s.drained}, info, nil
}

type verificationBarrierReader struct {
	io.ReadCloser
	enter, close             sync.Once
	closed, entered, drained chan struct{}
}

func (r *verificationBarrierReader) Read([]byte) (int, error) {
	r.enter.Do(func() { close(r.entered) })
	<-r.closed
	close(r.drained)
	return 0, io.ErrClosedPipe
}
func (r *verificationBarrierReader) Close() error {
	r.close.Do(func() { close(r.closed); _ = r.ReadCloser.Close() })
	return nil
}
