package incidentbundles_test

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	timelineroutetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestImportFinalPublicationRechecksSubmitterAvailability_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	sourceHarness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "extension_profile-incident-bundle-finalize-source", TestRouteMode: httptestx.TestRouteModeDisabled})
	sourceAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, sourceHarness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, sourceHarness.Server, sourceAdmin, map[string]any{
		"client_txn_id": "txn-incident-bundle-finalize-source",
		"incident_key":  "BUNDLE-FINALIZE",
		"title":         "Incident bundle finalization",
	})
	incidentID := incident["incident_id"].(string)
	timelineroutetest.CreateRow(t, sourceHarness.Server, sourceAdmin, incidentID, map[string]any{
		"client_txn_id":                   "txn-incident-bundle-finalize-row",
		"timeline.activity_synopsis_text": "Portable finalization event",
	})
	bundleBytes := exportBundleBytes(t, sourceHarness, sourceAdmin, incidentID, "txn-export-finalize-fixture")

	cases := []struct {
		name     string
		mutation string
	}{
		{name: "submitter demoted", mutation: `UPDATE users SET is_deployment_admin = false WHERE id = $1`},
		{name: "submitter inactive", mutation: `UPDATE users SET is_active = false WHERE id = $1`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			targetHarness := startIsolatedIncidentBundleServer(t, runtime, "extension_profile-incident-bundle-finalize-"+strings.ReplaceAll(tc.name, " ", "-"))
			targetAdmin, targetAdminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, targetHarness.Server.HTTP.URL)
			sequenceBefore := snapshotRecordRevisionSequence(t, targetHarness.DB)
			observerPassword := "ExtensionProfileImportObserverPass!"
			observerUser := flowtest.SeedLocalUserRecord(t, targetHarness.DB, "extension_profile-import-observer-"+strings.ReplaceAll(tc.name, " ", "-")+"@example.test", "Enterprise integration Import Observer", observerPassword, false, true, true)
			observerLogin := flowtest.LoginLocalUser(t, http.DefaultClient, targetHarness.Server.HTTP.URL, observerUser.Email, observerPassword, nil)

			barrierCtx, stopBarrier := context.WithTimeout(t.Context(), 20*time.Second)
			defer stopBarrier()
			gate, err := targetHarness.DB.BeginTx(barrierCtx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer gate.Rollback()
			// Key-share reads used by admission remain available. Only the final
			// owner's eligibility FOR UPDATE must wait for this held user row.
			var lockedUser string
			var blockerPID int
			if err := gate.QueryRowContext(barrierCtx, `SELECT id::text, pg_backend_pid() FROM users WHERE id=$1 FOR NO KEY UPDATE`, targetAdminID).Scan(&lockedUser, &blockerPID); err != nil || lockedUser != targetAdminID.String() {
				t.Fatalf("hold submitter=%q: %v", lockedUser, err)
			}
			request := newImportEnvelopeRequest(t, targetHarness.Server, targetAdmin, []uploadPart{
				jsonUploadPart("metadata", "", `{"client_txn_id":"txn-import-finalize-`+strings.ReplaceAll(tc.name, " ", "-")+`"}`),
				fileUploadPart("file", "bundle.zip", incidentBundleZIPMediaType, bundleBytes),
			})
			resp, err := http.DefaultClient.Do(request.WithContext(barrierCtx))
			if err != nil {
				t.Fatalf("admit import while submitter row is held: %v", err)
			}
			job := httptestx.RequireSuccessEnvelope(t, resp, http.StatusAccepted)["data"].(map[string]any)
			tick := time.NewTicker(10 * time.Millisecond)
			defer tick.Stop()
			for {
				var waiting int
				if err := targetHarness.DB.QueryRowContext(barrierCtx, `
SELECT count(*) FROM pg_stat_activity
WHERE datname=current_database() AND wait_event_type='Lock'
  AND query LIKE '-- name: GetIncidentBundleInitialAdminForUpdate%'
  AND $1=ANY(pg_blocking_pids(pid))`, blockerPID).Scan(&waiting); err != nil {
					t.Fatal(err)
				}
				if waiting == 1 {
					break
				}
				if waiting > 1 {
					t.Fatalf("multiple finalizers waited on one import: %d", waiting)
				}
				select {
				case <-barrierCtx.Done():
					t.Fatal("import did not reach the final submitter recheck", barrierCtx.Err())
				case <-tick.C:
				}
			}
			changed, err := gate.ExecContext(barrierCtx, tc.mutation, targetAdminID)
			if err != nil {
				t.Fatal(err)
			}
			if count, err := changed.RowsAffected(); err != nil || count != 1 {
				t.Fatalf("submitter mutation changed %d rows: %v", count, err)
			}
			if err := gate.Commit(); err != nil {
				t.Fatal(err)
			}
			stopBarrier()
			terminal := waitFailedJob(t, targetHarness.Server, observerLogin, job["job_id"].(string))
			requireFailedJobReason(t, terminal, "incident_bundle_import_rejected", "initial_admin_unavailable")
			if countRows(t, targetHarness.DB, `SELECT count(*) FROM incidents WHERE id = $1`, incidentID) != 0 {
				t.Fatalf("initial-admin-unavailable import must not make incident visible")
			}
			if sideEffects := snapshotImportFinalizationSideEffects(t, targetHarness.DB, incidentID, targetAdminID.String()); sideEffects != (importFinalizationSideEffects{}) {
				t.Fatalf("initial-admin-unavailable import left finalization side effects: %#v", sideEffects)
			}
			if sequenceAfter := snapshotRecordRevisionSequence(t, targetHarness.DB); sequenceAfter != sequenceBefore {
				t.Fatalf("failed import changed record revision sequence: before=%#v after=%#v", sequenceBefore, sequenceAfter)
			}
			assertNoIncidentBundleStaging(t, targetHarness.Server)
		})
	}
}

type recordRevisionSequenceState struct {
	LastValue int64
	IsCalled  bool
}

func snapshotRecordRevisionSequence(t testing.TB, db *sql.DB) recordRevisionSequenceState {
	t.Helper()
	var state recordRevisionSequenceState
	if err := db.QueryRow(`
SELECT last_value, is_called
  FROM public.record_revisions_revision_id_seq
`).Scan(&state.LastValue, &state.IsCalled); err != nil {
		t.Fatalf("snapshot record revision sequence: %v", err)
	}
	return state
}

func TestImportEnvelopeFailuresCreateNoDurableState_Integration(t *testing.T) {
	harness := appsupport.StartRuntime(t).StartServer(t, appsupport.ServerOptions{Prefix: "extension_profile-incident-bundle-envelope-failures", TestRouteMode: httptestx.TestRouteModeDisabled})
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	validFile := []byte("not a bundle but parser-valid bytes")
	cases := []struct {
		name           string
		build          func(testing.TB, *httptestx.Server, authcookietest.AuthCookies) *http.Request
		wantReason     string
		wantPart       string
		wantContentErr bool
	}{
		{
			name: "missing boundary",
			build: func(t testing.TB, server *httptestx.Server, login authcookietest.AuthCookies) *http.Request {
				req, err := http.NewRequest(http.MethodPost, server.HTTP.URL+"/api/v1/incident-bundles/import", strings.NewReader("not multipart"))
				if err != nil {
					t.Fatalf("create missing-boundary request: %v", err)
				}
				req.Header.Set("Content-Type", "multipart/form-data")
				addImportAuth(req, login)
				return req
			},
			wantReason: "unsupported_upload_envelope",
		},
		{
			name: "duplicate file after streamed input",
			build: func(t testing.TB, server *httptestx.Server, login authcookietest.AuthCookies) *http.Request {
				return newImportEnvelopeRequest(t, server, login, []uploadPart{
					fileUploadPart("file", "first.zip", incidentBundleZIPMediaType, validFile),
					jsonUploadPart("metadata", "", `{"client_txn_id":"txn-envelope-late-duplicate"}`),
					fileUploadPart("file", "second.zip", incidentBundleZIPMediaType, validFile),
				})
			},
			wantReason: "duplicate_part", wantPart: "file",
		},
		{
			name: "malformed metadata after streamed input",
			build: func(t testing.TB, server *httptestx.Server, login authcookietest.AuthCookies) *http.Request {
				return newImportEnvelopeRequest(t, server, login, []uploadPart{
					fileUploadPart("file", "first.zip", incidentBundleZIPMediaType, validFile),
					jsonUploadPart("metadata", "", `{"client_txn_id":`),
				})
			},
			wantReason: "malformed_metadata_json", wantPart: "metadata",
		},
		{
			name: "duplicate metadata part",
			build: func(t testing.TB, server *httptestx.Server, login authcookietest.AuthCookies) *http.Request {
				return newImportEnvelopeRequest(t, server, login, []uploadPart{
					jsonUploadPart("metadata", "", `{"client_txn_id":"txn-envelope-duplicate-metadata"}`),
					jsonUploadPart("metadata", "", `{"client_txn_id":"txn-envelope-duplicate-metadata"}`),
					fileUploadPart("file", "bundle.zip", incidentBundleZIPMediaType, validFile),
				})
			},
			wantReason: "duplicate_part",
			wantPart:   "metadata",
		},
		{
			name: "unexpected part",
			build: func(t testing.TB, server *httptestx.Server, login authcookietest.AuthCookies) *http.Request {
				return newImportEnvelopeRequest(t, server, login, []uploadPart{
					jsonUploadPart("metadata", "", `{"client_txn_id":"txn-envelope-unexpected-part"}`),
					fileUploadPart("extra", "extra.txt", "text/plain", []byte("extra")),
					fileUploadPart("file", "bundle.zip", incidentBundleZIPMediaType, validFile),
				})
			},
			wantReason: "unexpected_part",
		},
		{
			name: "malformed metadata json",
			build: func(t testing.TB, server *httptestx.Server, login authcookietest.AuthCookies) *http.Request {
				return newImportEnvelopeRequest(t, server, login, []uploadPart{
					jsonUploadPart("metadata", "", `{"client_txn_id":`),
					fileUploadPart("file", "bundle.zip", incidentBundleZIPMediaType, validFile),
				})
			},
			wantReason: "malformed_metadata_json",
			wantPart:   "metadata",
		},
		{
			name: "duplicate metadata json key",
			build: func(t testing.TB, server *httptestx.Server, login authcookietest.AuthCookies) *http.Request {
				return newImportEnvelopeRequest(t, server, login, []uploadPart{
					jsonUploadPart("metadata", "", `{"client_txn_id":"txn-a","client_txn_id":"txn-b"}`),
					fileUploadPart("file", "bundle.zip", incidentBundleZIPMediaType, validFile),
				})
			},
			wantReason: "malformed_metadata_json",
			wantPart:   "metadata",
		},
		{
			name: "non-object metadata",
			build: func(t testing.TB, server *httptestx.Server, login authcookietest.AuthCookies) *http.Request {
				return newImportEnvelopeRequest(t, server, login, []uploadPart{
					jsonUploadPart("metadata", "", `[]`),
					fileUploadPart("file", "bundle.zip", incidentBundleZIPMediaType, validFile),
				})
			},
			wantReason: "request_not_object",
			wantPart:   "metadata",
		},
		{
			name: "invalid file content type",
			build: func(t testing.TB, server *httptestx.Server, login authcookietest.AuthCookies) *http.Request {
				return newImportEnvelopeRequest(t, server, login, []uploadPart{
					jsonUploadPart("metadata", "", `{"client_txn_id":"txn-envelope-file-content-type"}`),
					fileUploadPart("file", "bundle.txt", "text/plain", validFile),
				})
			},
			wantReason:     "invalid_part_content_type",
			wantPart:       "file",
			wantContentErr: true,
		},
		{
			name: "forbidden import mode field",
			build: func(t testing.TB, server *httptestx.Server, login authcookietest.AuthCookies) *http.Request {
				return newImportEnvelopeRequest(t, server, login, []uploadPart{
					jsonUploadPart("metadata", "", `{"client_txn_id":"txn-envelope-forbidden-mode","clone_mode":"copy"}`),
					fileUploadPart("file", "bundle.zip", incidentBundleZIPMediaType, validFile),
				})
			},
			wantReason: "unknown_field",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := snapshotEnvelopeDurability(t, harness.DB)
			resp := httptestx.Do(t, http.DefaultClient, tc.build(t, harness.Server, admin))
			body := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_incident_bundle_request")
			details := httptestx.RequireErrorDetails(t, body)
			if details["reason_code"] != tc.wantReason {
				t.Fatalf("reason mismatch: got %#v want %s", details, tc.wantReason)
			}
			if tc.wantPart != "" && details["part_name"] != tc.wantPart {
				t.Fatalf("part_name mismatch: got %#v want %s", details, tc.wantPart)
			}
			if tc.wantContentErr {
				if details["received_content_type"] != "text/plain" {
					t.Fatalf("received content type missing: %#v", details)
				}
				if len(stringArray(t, details["allowed_content_types"])) == 0 {
					t.Fatalf("allowed content types missing: %#v", details)
				}
			}
			after := snapshotEnvelopeDurability(t, harness.DB)
			if after != before {
				t.Fatalf("early envelope failure created durable rows: before=%#v after=%#v", before, after)
			}
			assertNoIncidentBundleStaging(t, harness.Server)
		})
	}
}
