package incidentbundles_test

import (
	"database/sql"
	"net/http"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	timelineroutetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestImportFinalPublicationRechecksSubmitterAvailability_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	sourceHarness := runtime.StartDefaultServer(t, "extension_profile-incident-bundle-finalize-source")
	sourceAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, sourceHarness.Server.HTTP.URL)
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
		name   string
		mutate func(testing.TB, *sql.DB, string)
	}{
		{
			name: "submitter demoted",
			mutate: func(t testing.TB, db *sql.DB, userID string) {
				t.Helper()
				if _, err := db.Exec(`UPDATE users SET is_deployment_admin = false WHERE id = $1`, userID); err != nil {
					t.Fatalf("demote import submitter: %v", err)
				}
			},
		},
		{
			name: "submitter inactive",
			mutate: func(t testing.TB, db *sql.DB, userID string) {
				t.Helper()
				if _, err := db.Exec(`UPDATE users SET is_active = false WHERE id = $1`, userID); err != nil {
					t.Fatalf("deactivate import submitter: %v", err)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			targetHarness := startIsolatedIncidentBundleServer(t, runtime, "extension_profile-incident-bundle-finalize-"+strings.ReplaceAll(tc.name, " ", "-"))
			targetAdmin, targetAdminID := flowtest.ProvisionBootstrapAdmin(t, targetHarness.Server.HTTP.URL)
			sequenceBefore := snapshotRecordRevisionSequence(t, targetHarness.DB)
			observerPassword := "ExtensionProfileImportObserverPass!"
			observerUser := flowtest.SeedLocalUserRecord(t, targetHarness.DB, "extension_profile-import-observer-"+strings.ReplaceAll(tc.name, " ", "-")+"@example.test", "Enterprise integration Import Observer", observerPassword, false, true, true)
			observerCookies, observerCSRF := flowtest.LoginLocalUser(t, targetHarness.Server.HTTP.URL, observerUser.Email, observerPassword, nil)
			observerLogin := flowtest.LoginResult{SessionCookie: observerCookies, CSRFCookie: observerCSRF}

			resp := postImport(t, targetHarness.Server, targetAdmin, `{"client_txn_id":"txn-import-finalize-`+strings.ReplaceAll(tc.name, " ", "-")+`"}`, bundleBytes, "bundle.zip")
			job := httptestx.RequireSuccessEnvelope(t, resp, http.StatusAccepted)["data"].(map[string]any)

			tc.mutate(t, targetHarness.DB, targetAdminID)
			terminal := waitFailedJob(t, targetHarness.Server, observerLogin, job["job_id"].(string))
			requireFailedJobReason(t, terminal, "incident_bundle_import_rejected", "initial_admin_unavailable")
			if countRows(t, targetHarness.DB, `SELECT count(*) FROM incidents WHERE id = $1`, incidentID) != 0 {
				t.Fatalf("initial-admin-unavailable import must not make incident visible")
			}
			if sideEffects := snapshotImportFinalizationSideEffects(t, targetHarness.DB, incidentID, targetAdminID); sideEffects != (importFinalizationSideEffects{}) {
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
	harness := appsupport.StartRuntime(t).StartDefaultServer(t, "extension_profile-incident-bundle-envelope-failures")
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	validFile := []byte("not a bundle but parser-valid bytes")
	cases := []struct {
		name           string
		build          func(testing.TB, *httptestx.Server, flowtest.LoginResult) *http.Request
		wantReason     string
		wantPart       string
		wantContentErr bool
	}{
		{
			name: "missing boundary",
			build: func(t testing.TB, server *httptestx.Server, login flowtest.LoginResult) *http.Request {
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
			name: "duplicate metadata part",
			build: func(t testing.TB, server *httptestx.Server, login flowtest.LoginResult) *http.Request {
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
			build: func(t testing.TB, server *httptestx.Server, login flowtest.LoginResult) *http.Request {
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
			build: func(t testing.TB, server *httptestx.Server, login flowtest.LoginResult) *http.Request {
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
			build: func(t testing.TB, server *httptestx.Server, login flowtest.LoginResult) *http.Request {
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
			build: func(t testing.TB, server *httptestx.Server, login flowtest.LoginResult) *http.Request {
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
			build: func(t testing.TB, server *httptestx.Server, login flowtest.LoginResult) *http.Request {
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
			build: func(t testing.TB, server *httptestx.Server, login flowtest.LoginResult) *http.Request {
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
