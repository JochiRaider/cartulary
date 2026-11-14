package evidence_test

// Evidence access-handle route contracts.
import (
	"bytes"
	"context"
	"errors"
	"github.com/google/uuid"
	"io"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestEvidenceHandles_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "entity_linking-evidence-handles", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-entity_linking-handles-incident",
		"incident_key":  "entity_linking-handles",
		"title":         "Record relationships evidence handles",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	recordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
	payload := []byte("handle body")
	attachUploadedBlob(t, harness, login, incidentID, recordID, payload, "txn-entity_linking-handles-blob", "txn-entity_linking-handles-attach")
	t.Run("bounded metadata preserves state without issuing handles", func(t *testing.T) {
		unattached := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, adminID, unattached)
		if _, err := harness.DB.ExecContext(context.Background(), `UPDATE evidence SET title=NULL WHERE record_id=$1`, unattached); err != nil {
			t.Fatal(err)
		}
		url := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/evidence-metadata"
		body := map[string]any{"record_ids": []string{unattached.String(), recordID.String()}}
		before := countAccessHandles(t, harness, incidentID)
		readOnlyState := func() string {
			return structuredTableText(t, harness, "records", "evidence", "object_blobs",
				"evidence_custody_events", "route_idempotency", "change_sets",
				"change_set_mutations", "record_revisions", "evidence_access_handles",
				"evidence_grid_projection", "timeline_grid_projection")
		}
		beforeState := readOnlyState()
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, url, body, authOptions(login)...)
		data := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
		items := data["items"].([]any)
		if len(data) != 2 || data["incident_id"] != incidentID.String() || len(items) != 2 {
			t.Fatal("metadata response lost its scope or selection")
		}
		first, second := items[0].(map[string]any), items[1].(map[string]any)
		if len(first) != 7 || len(second) != 7 || first["record_id"] != unattached.String() || second["record_id"] != recordID.String() || first["title"] != nil || first["has_blob"] != false || first["upload_state"] != nil || second["has_blob"] != true || second["upload_state"] != "available" {
			t.Fatal("metadata identity, null shape, or attachment state differs", data)
		}
		for index, item := range []map[string]any{first, second} {
			if item["row_version"].(float64) < 1 {
				t.Fatal("missing current row version")
			}
			eligibility := item["preview_eligibility"].(map[string]any)
			if len(eligibility) != 2 || eligibility["eligible"] != (index == 1) {
				t.Fatal("incorrect preview metadata eligibility", eligibility)
			}
			if index == 0 && eligibility["reason_code"] != "no_visible_blob" || index == 1 && eligibility["reason_code"] != nil {
				t.Fatal("incorrect preview metadata reason", eligibility)
			}
		}
		if countAccessHandles(t, harness, incidentID) != before {
			t.Fatal("metadata read issued a handle")
		}
		if readOnlyState() != beforeState {
			t.Fatal("metadata read changed evidence, history, idempotency, or projections")
		}
		httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, url, body), http.StatusUnauthorized, "session_required")
		httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, url, body, httptestx.WithCookies(login.Session)), http.StatusForbidden, "csrf_verification_failed")
		for _, ids := range [][]string{{uuid.NewString(), recordID.String()}, {recordID.String(), recordID.String()}} {
			status, code := http.StatusNotFound, "evidence_record_not_found"
			if ids[0] == ids[1] {
				status, code = http.StatusBadRequest, "invalid_evidence_metadata_request"
			}
			httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, url, map[string]any{"record_ids": ids}, authOptions(login)...), status, code)
		}
		updateRecordDeletedState(t, harness, unattached, true)
		httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, url, body, authOptions(login)...), http.StatusNotFound, "evidence_record_not_found")
		otherIncident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-evidence-metadata-other-incident",
			"incident_key":  "evidence-metadata-other", "title": "Other incident",
		})
		otherID := appsupport.MustUUID(t, otherIncident["incident_id"].(string))
		foreignRecord := uuid.New()
		seedEvidenceRecord(t, harness, otherID, adminID, foreignRecord)
		beforeState = readOnlyState()
		denied := httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, url,
			map[string]any{"record_ids": []string{recordID.String(), foreignRecord.String()}}, authOptions(login)...),
			http.StatusNotFound, "evidence_record_not_found")
		if _, exists := denied["data"]; exists {
			t.Fatal("cross-incident query returned partial metadata")
		}
		if readOnlyState() != beforeState {
			t.Fatal("rejected metadata query changed structured state")
		}
	})

	beforeInvalidHandles := countAccessHandles(t, harness, incidentID)
	for _, endpoint := range []struct {
		name string
		path string
	}{
		{name: "preview", path: "preview-handle"},
		{name: "download", path: "download-handle"},
	} {
		for name, body := range map[string]string{
			"zero length": "",
			"null":        "null",
			"array":       "[]",
			"unknown":     `{"unexpected":true}`,
			"client txn":  `{"client_txn_id":"forbidden"}`,
		} {
			t.Run(endpoint.name+" "+name, func(t *testing.T) {
				resp := httptestx.DoRawJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/"+endpoint.path, body, authOptions(login)...)
				httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_evidence_handle_request")
			})
		}
	}
	if got := countAccessHandles(t, harness, incidentID); got != beforeInvalidHandles {
		t.Fatalf("invalid handle issuance requests wrote evidence_access_handles: got %d want %d", got, beforeInvalidHandles)
	}

	previewResp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/preview-handle", map[string]any{}, authOptions(login)...)
	previewData := httptestx.RequireSuccessEnvelope(t, previewResp, http.StatusOK)["data"].(map[string]any)
	if previewData["handle_kind"] != "preview" || previewData["single_use"] != false || previewData["disposition"] != "inline" || previewData["preview_kind"] != "text_inline" {
		t.Fatalf("unexpected preview handle payload: %#v", previewData)
	}
	downloadResp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/download-handle", map[string]any{}, authOptions(login)...)
	downloadData := httptestx.RequireSuccessEnvelope(t, downloadResp, http.StatusOK)["data"].(map[string]any)
	if downloadData["handle_kind"] != "download" || downloadData["single_use"] != true || downloadData["disposition"] != "attachment" {
		t.Fatalf("unexpected download handle payload: %#v", downloadData)
	}
	if _, ok := downloadData["preview_kind"]; ok {
		t.Fatalf("download handle must omit preview_kind, got %#v", downloadData)
	}
	afterIssuedHandles := countAccessHandles(t, harness, incidentID)
	if afterIssuedHandles != beforeInvalidHandles+2 {
		t.Fatalf("expected exactly two successful handle rows, got %d want %d", afterIssuedHandles, beforeInvalidHandles+2)
	}

	if _, err := harness.DB.ExecContext(context.Background(), `
DELETE FROM incident_memberships
 WHERE incident_id = $1
   AND user_id = $2
`, incidentID, adminID); err != nil {
		t.Fatalf("remove incident membership before handle auth re-derivation: %v", err)
	}
	for _, endpoint := range []string{"preview-handle", "download-handle"} {
		resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/"+endpoint, map[string]any{}, authOptions(login)...)
		httptestx.RequireErrorEnvelope(t, resp, http.StatusNotFound, "evidence_record_not_found")
	}
	httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/evidence-metadata",
		map[string]any{"record_ids": []string{recordID.String()}}, authOptions(login)...),
		http.StatusNotFound, "incident_not_found")
	if got := countAccessHandles(t, harness, incidentID); got != afterIssuedHandles {
		t.Fatalf("authorization failures wrote evidence_access_handles: got %d want %d", got, afterIssuedHandles)
	}
}

func TestEvidenceHandleIssuanceReportsRegisteredUnavailableReasons(t *testing.T) {
	harness := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "entity_linking-evidence-handle-issue-reasons", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-entity_linking-handle-issue-reasons-incident",
		"incident_key":  "entity_linking-handle-issue-reasons",
		"title":         "Record relationships handle issue reasons",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	scenarios := []struct {
		name       string
		reasonCode string
		arrange    func(t *testing.T, recordID uuid.UUID)
	}{
		{
			name:       "no linked blob",
			reasonCode: "no_visible_blob",
			arrange:    func(t *testing.T, recordID uuid.UUID) {},
		},
		{
			name:       "pending blob",
			reasonCode: "blob_pending",
			arrange: func(t *testing.T, recordID uuid.UUID) {
				linkSeededBlob(t, harness, incidentID, adminID, recordID, "pending", "available", "issue/pending")
			},
		},
		{
			name:       "failed blob",
			reasonCode: "blob_failed",
			arrange: func(t *testing.T, recordID uuid.UUID) {
				linkSeededBlob(t, harness, incidentID, adminID, recordID, "failed", "available", "issue/failed")
			},
		},
		{
			name:       "missing backing object",
			reasonCode: "blob_missing",
			arrange: func(t *testing.T, recordID uuid.UUID) {
				linkSeededBlobWithCanonicalStorageKey(t, harness, incidentID, adminID, recordID, "available", "available")
			},
		},
		{
			name:       "quarantined blob",
			reasonCode: "evidence_quarantined",
			arrange: func(t *testing.T, recordID uuid.UUID) {
				linkSeededBlob(t, harness, incidentID, adminID, recordID, "quarantined", "available", "issue/quarantined")
			},
		},
		{
			name:       "inconsistent lifecycle",
			reasonCode: "evidence_inconsistent",
			arrange: func(t *testing.T, recordID uuid.UUID) {
				linkSeededBlob(t, harness, incidentID, adminID, recordID, "available", "received", "issue/inconsistent")
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			recordID := uuid.New()
			seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
			scenario.arrange(t, recordID)

			beforeHandles := countAccessHandles(t, harness, incidentID)
			for _, endpoint := range []string{"preview-handle", "download-handle"} {
				resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/"+endpoint, map[string]any{}, authOptions(login)...)
				requireEvidenceAccessUnavailableReason(t, resp, scenario.reasonCode)
			}
			if got := countAccessHandles(t, harness, incidentID); got != beforeHandles {
				t.Fatalf("blocked issuance wrote evidence_access_handles: got %d want %d", got, beforeHandles)
			}
		})
	}
}

func TestEvidenceHandleRedemptionReportsRegisteredUnavailableReasons(t *testing.T) {
	harness := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "entity_linking-evidence-handle-redeem-reasons", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-entity_linking-handle-redeem-reasons-incident",
		"incident_key":  "entity_linking-handle-redeem-reasons",
		"title":         "Record relationships handle redeem reasons",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	scenarios := []struct {
		name       string
		reasonCode string
		mutate     func(t *testing.T, recordID uuid.UUID, objectBlobID uuid.UUID)
	}{
		{
			name:       "detached blob",
			reasonCode: "no_visible_blob",
			mutate: func(t *testing.T, recordID uuid.UUID, objectBlobID uuid.UUID) {
				updateEvidenceBlobLink(t, harness, recordID, nil)
			},
		},
		{
			name:       "pending blob",
			reasonCode: "blob_pending",
			mutate: func(t *testing.T, recordID uuid.UUID, objectBlobID uuid.UUID) {
				updateBlobState(t, harness, objectBlobID, "pending")
			},
		},
		{
			name:       "failed blob",
			reasonCode: "blob_failed",
			mutate: func(t *testing.T, recordID uuid.UUID, objectBlobID uuid.UUID) {
				updateBlobState(t, harness, objectBlobID, "failed")
			},
		},
		{
			name:       "missing backing object",
			reasonCode: "blob_missing",
			mutate: func(t *testing.T, recordID uuid.UUID, objectBlobID uuid.UUID) {
				deleteBlobObject(t, harness, blobStorageKey(t, harness, objectBlobID))
			},
		},
		{
			name:       "quarantined blob",
			reasonCode: "evidence_quarantined",
			mutate: func(t *testing.T, recordID uuid.UUID, objectBlobID uuid.UUID) {
				updateBlobState(t, harness, objectBlobID, "quarantined")
			},
		},
		{
			name:       "inconsistent lifecycle",
			reasonCode: "evidence_inconsistent",
			mutate: func(t *testing.T, recordID uuid.UUID, objectBlobID uuid.UUID) {
				updateEvidenceLifecycle(t, harness, recordID, "received")
			},
		},
	}

	for _, scenario := range scenarios {
		for _, endpoint := range []string{"preview-handle", "download-handle"} {
			t.Run(endpoint+" "+scenario.name, func(t *testing.T) {
				recordID := uuid.New()
				seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
				attachData := attachUploadedBlob(t, harness, login, incidentID, recordID, []byte("redeem body"), "txn-"+recordID.String()+"-blob", "txn-"+recordID.String()+"-attach")
				objectBlobID := appsupport.MustUUID(t, attachData["object_blob_id"].(string))

				issueResp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/"+endpoint, map[string]any{}, authOptions(login)...)
				issueData := httptestx.RequireSuccessEnvelope(t, issueResp, http.StatusOK)["data"].(map[string]any)
				scenario.mutate(t, recordID, objectBlobID)

				redeemResp := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+issueData["href"].(string), nil, httptestx.WithCookies(login.Session))
				requireEvidenceAccessUnavailableReason(t, redeemResp, scenario.reasonCode)
			})
		}
	}
}

func TestDownloadHandleBlobMissingDoesNotConsumeHandle_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	base, err := objectstore.NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = base.Close() })
	store := &firstReadFailureStore{FilesystemStore: base}
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "entity_linking-download-handle-no-byte-failure", ObjectStore: store, TestRouteMode: httptestx.TestRouteModeDisabled})
	login, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-entity_linking-download-no-byte-incident",
		"incident_key":  "entity_linking-download-no-byte",
		"title":         "Record relationships download no-byte failure",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	recordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
	payload := []byte("download retry body")
	attachData := attachUploadedBlob(t, harness, login, incidentID, recordID, payload, "txn-entity_linking-download-no-byte-blob", "txn-entity_linking-download-no-byte-attach")
	objectBlobID := appsupport.MustUUID(t, attachData["object_blob_id"].(string))
	originalStorageKey := blobStorageKey(t, harness, objectBlobID)

	downloadResp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/download-handle", map[string]any{}, authOptions(login)...)
	downloadData := httptestx.RequireSuccessEnvelope(t, downloadResp, http.StatusOK)["data"].(map[string]any)
	downloadURL := harness.Server.HTTP.URL + downloadData["href"].(string)

	deleteBlobObject(t, harness, originalStorageKey)
	missingResp := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, downloadURL, nil, httptestx.WithCookies(login.Session))
	requireEvidenceAccessUnavailableReason(t, missingResp, "blob_missing")

	if err := harness.ObjectStore.PutObject(context.Background(), originalStorageKey, bytes.NewReader(payload), int64(len(payload)), "text/plain"); err != nil {
		t.Fatalf("restore blob object: %v", err)
	}
	for _, fault := range []struct {
		name, code, reason string
		err                error
		status             int
		retryable          bool
	}{
		{"unavailable", "object_store_unavailable", "endpoint_unreachable", &objectstore.AdapterError{Code: objectstore.ErrorCodeUnavailable, Reason: objectstore.ReasonEndpointUnreachable, Message: originalStorageKey}, http.StatusServiceUnavailable, true},
		{"access rejected", "object_store_access_rejected", "capability_missing", &objectstore.AdapterError{Code: objectstore.ErrorCodeAccessRejected, Reason: objectstore.ReasonCapabilityMissing, Message: originalStorageKey}, http.StatusServiceUnavailable, false},
		{"premature EOF", "evidence_access_unavailable", "evidence_inconsistent", io.EOF, http.StatusConflict, false},
		{"untyped read failure", "object_store_unavailable", "endpoint_unreachable", errors.New(originalStorageKey), http.StatusServiceUnavailable, true},
	} {
		t.Run(fault.name, func(t *testing.T) {
			failure := &firstReadFailure{err: fault.err}
			store.failure.Store(failure)
			response := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, downloadURL, nil, httptestx.WithCookies(login.Session))
			body := httptestx.RequireErrorEnvelope(t, response, fault.status, fault.code)
			httptestx.RequireErrorDetail(t, body, "reason_code", fault.reason)
			if body["error"].(map[string]any)["retryable"] != fault.retryable {
				t.Fatal("incorrect retry policy", body)
			}
			requireNoPublicEvidenceLeak(t, "first read failure", body, []string{originalStorageKey})
			if !failure.closed.Load() {
				t.Fatal("failed stored reader was not closed")
			}
		})
	}
	downloadBody := redeemHandle(t, downloadURL, login)
	if string(downloadBody) != string(payload) {
		t.Fatalf("download body mismatch after no-byte failure: got %q", string(downloadBody))
	}
	second := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, downloadURL, nil, httptestx.WithCookies(login.Session))
	httptestx.RequireErrorEnvelope(t, second, http.StatusGone, "handle_consumed")

	for _, kind := range []string{"download-handle", "preview-handle"} {
		t.Run(kind+" range", func(t *testing.T) {
			issued := issueEvidenceHandle(t, harness, login, recordID, kind)
			url := harness.Server.HTTP.URL + issued["href"].(string)
			failure := &firstReadFailure{err: io.ErrUnexpectedEOF}
			store.failure.Store(failure)
			request := httptestx.NewJSONRequest(t, http.MethodGet, url, nil)
			request.Header.Set("Range", "bytes=2-5")
			httptestx.WithCookies(login.Session)(request)
			response := httptestx.Do(t, http.DefaultClient, request)
			requireEvidenceAccessUnavailableReason(t, response, "evidence_inconsistent")
			if !failure.closed.Load() || response.Header.Get("Content-Range") != "" {
				t.Fatal("failed range retained a reader or committed range headers")
			}
			response = httptestx.Do(t, http.DefaultClient, request.Clone(t.Context()))
			body := httptestx.ReadResponseBody(t, response, http.StatusPartialContent)
			if !bytes.Equal(body, payload[2:6]) || response.Header.Get("Content-Range") != "bytes 2-5/19" {
				t.Fatalf("range result: %q %q", body, response.Header.Get("Content-Range"))
			}
			if kind == "download-handle" {
				httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, url, nil, httptestx.WithCookies(login.Session)), http.StatusGone, "handle_consumed")
			} else if !bytes.Equal(redeemHandle(t, url, login), payload) {
				t.Fatal("preview was consumed or lost bytes")
			}
		})
	}

	t.Run("empty download remains a successful single use", func(t *testing.T) {
		emptyID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, adminID, emptyID)
		attachUploadedBlob(t, harness, login, incidentID, emptyID, nil, "empty-blob", "empty-attach")
		issued := issueEvidenceHandle(t, harness, login, emptyID, "download-handle")
		url := harness.Server.HTTP.URL + issued["href"].(string)
		if body := redeemHandle(t, url, login); len(body) != 0 {
			t.Fatalf("empty download contained %q", body)
		}
		httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, url, nil, httptestx.WithCookies(login.Session)), http.StatusGone, "handle_consumed")
	})
}

type firstReadFailureStore struct {
	*objectstore.FilesystemStore
	failure atomic.Pointer[firstReadFailure]
}

type firstReadFailure struct {
	err    error
	closed atomic.Bool
}

func (s *firstReadFailureStore) Get(ctx context.Context, request objectstore.GetObjectRequest) (io.ReadCloser, objectstore.ObjectInfo, error) {
	reader, info, err := s.FilesystemStore.Get(ctx, request)
	if err != nil {
		return reader, info, err
	}
	if failure := s.failure.Swap(nil); failure != nil {
		return &firstReadFailureReader{ReadCloser: reader, failure: failure}, info, nil
	}
	return reader, info, nil
}

type firstReadFailureReader struct {
	io.ReadCloser
	failure *firstReadFailure
}

func (r *firstReadFailureReader) Read([]byte) (int, error) { return 0, r.failure.err }
func (r *firstReadFailureReader) Close() error {
	err := r.ReadCloser.Close()
	r.failure.closed.Store(true)
	return err
}
