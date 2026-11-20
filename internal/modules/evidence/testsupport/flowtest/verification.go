// Package flowtest exercises Evidence's public upload-verification workflow.
package flowtest

import (
	"context"
	"net/http"
	"testing"
	"time"

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
	// One 10 s budget covers admission and observation.
	deadline := time.Now().Add(10 * time.Second)
	ctx, cancel := context.WithDeadline(t.Context(), deadline)
	defer cancel()
	options := []func(*http.Request){httptestx.WithCookies(sessionCookie, csrfCookie), httptestx.WithHeader(authn.CSRFHeaderName, csrfCookie.Value)}
	withDeadline := func(req *http.Request) { *req = *req.WithContext(ctx) }
	response := httptestx.DoJSON(t, client, http.MethodPost, origin+"/api/v1/object-blobs/"+blobID+"/verify", map[string]any{"client_txn_id": transactionID}, append(options, withDeadline)...)
	job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
	jobID, ok := job["job_id"].(string)
	if !ok {
		t.Fatal("verification fixture Job identity missing")
	}
	if _, err := uuid.Parse(jobID); err != nil {
		t.Fatal("verification fixture Job identity invalid")
	}
	// The shared poll helper reports a failed or canceled Job with its error
	// summary and the last response at the deadline (TH-HARNESS-REQ-864).
	httptestx.PollJob(t, origin, jobID, httptestx.PollOptions{
		Timeout: max(time.Until(deadline), time.Millisecond), Interval: 10 * time.Millisecond, Client: client, Request: options,
	}, "succeeded")
}
