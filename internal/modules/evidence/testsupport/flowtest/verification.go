// Package flowtest exercises Evidence's public upload-verification workflow.
package flowtest

import (
	"context"
	"net/http"
	"testing"
	"time"

	authflow "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

// VerifyUploadedBlob admits and observes real stored-byte verification before a
// positive fixture finalizes it. It never replaces a failed Job or retries work.
func VerifyUploadedBlob(t testing.TB, client *http.Client, origin string, sessionCookie, csrfCookie *http.Cookie, blobID, transactionID string) {
	t.Helper()
	if _, err := uuid.Parse(blobID); err != nil {
		t.Fatal("invalid verification fixture blob identity")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	options := []func(*http.Request){httptestx.WithCookies(sessionCookie, csrfCookie), httptestx.WithHeader(authn.CSRFHeaderName, csrfCookie.Value), func(req *http.Request) { *req = *req.WithContext(ctx) }}
	response := authflow.DoJSON(t, client, http.MethodPost, origin+"/api/v1/object-blobs/"+blobID+"/verify", map[string]any{"client_txn_id": transactionID}, options...)
	job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
	jobID, ok := job["job_id"].(string)
	if !ok {
		t.Fatal("verification fixture Job identity missing")
	}
	if _, err := uuid.Parse(jobID); err != nil {
		t.Fatal("verification fixture Job identity invalid")
	}
	for job["status"] == "queued" || job["status"] == "running" || job["status"] == "cancel_requested" {
		select {
		case <-ctx.Done():
			t.Fatal("verification fixture Job did not finish")
		case <-time.After(10 * time.Millisecond):
		}
		response = authflow.DoJSON(t, client, http.MethodGet, origin+"/api/v1/jobs/"+jobID, nil, options...)
		job = httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
	}
	if job["status"] != "succeeded" {
		t.Fatalf("verification fixture status=%v error=%v", job["status"], job["error_summary"])
	}
}
