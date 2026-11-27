package imports_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/app/server"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtureartifacts"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

const sourceAdmissionHeader = "Source IP Address,Destination IP Address,Source Port,Destination Port,Protocol,Bytes,Packets,Flow Start Time,Flow End Time,Input Interface,Output Interface\n"
const sourceAdmissionRow = "192.0.2.10,198.51.100.1,443,51515,TCP,1234,12,2026-07-10T12:00:00Z,2026-07-10T12:00:05Z,Gi0/1,Gi0/2\n"

// Keep failures actionable without disclosing a source value or database error.
// The existing composition seam preserves the actual facade and transaction.
type sourceAdmissionPass struct {
	Bytes     int64 `json:"bytes"`
	ReadNS    int64 `json:"read_ns"`
	ElapsedNS int64 `json:"elapsed_ns"`
}
type sourceAdmissionTiming struct {
	sync.Mutex
	passes []sourceAdmissionPass
}
type sourceAdmissionDiagnostics struct {
	t      testing.TB
	timing *sourceAdmissionTiming
}

func (d sourceAdmissionDiagnostics) WrapDatabase(db dbport.DB) dbport.DB { return db }

// Observe the real Imports-owned reader, including time between reads. This
// distinguishes source I/O from normalization and writes without logging data.
type sourceAdmissionPort struct {
	networkflow.ImportSourcePort
	t      testing.TB
	timing *sourceAdmissionTiming
}

func (p sourceAdmissionPort) OpenSourceStreamTx(ctx context.Context, tx pgx.Tx, ref string) (imports.ImportSourceStream, error) {
	stream, err := p.ImportSourcePort.OpenSourceStreamTx(ctx, tx, ref)
	if err != nil {
		return stream, err
	}
	stream.Reader = &sourceAdmissionReader{ReadCloser: stream.Reader, t: p.t, timing: p.timing, started: time.Now()}
	return stream, nil
}

type sourceAdmissionReader struct {
	io.ReadCloser
	timing   *sourceAdmissionTiming
	t        testing.TB
	started  time.Time
	readTime time.Duration
	bytes    int64
}

func (r *sourceAdmissionReader) Read(b []byte) (int, error) {
	start := time.Now()
	n, err := r.ReadCloser.Read(b)
	r.readTime += time.Since(start)
	r.bytes += int64(n)
	return n, err
}
func (r *sourceAdmissionReader) Close() error {
	err := r.ReadCloser.Close()
	if r.timing != nil {
		r.timing.Lock()
		r.timing.passes = append(r.timing.passes, sourceAdmissionPass{r.bytes, int64(r.readTime), int64(time.Since(r.started))})
		r.timing.Unlock()
	}
	r.t.Logf("bounded import source pass: bytes=%d read_time=%s pass_elapsed=%s", r.bytes, r.readTime, time.Since(r.started))
	return err
}
func (d sourceAdmissionDiagnostics) ConfigureNetworkFlow(deps networkflow.ModuleDependencies) networkflow.ModuleDependencies {
	deps.ImportSources = sourceAdmissionPort{ImportSourcePort: deps.ImportSources, t: d.t, timing: d.timing}
	return deps
}
func (d sourceAdmissionDiagnostics) WrapImportFacade(f imports.ExtensionImportFacade) imports.ExtensionImportFacade {
	return sourceAdmissionObservedFacade{ExtensionImportFacade: f, t: d.t}
}

type sourceAdmissionObservedFacade struct {
	imports.ExtensionImportFacade
	t testing.TB
}

func (f sourceAdmissionObservedFacade) TranslateImportUnitError(err error) (imports.ExtensionImportErrorTranslation, bool) {
	f.t.Logf("owner apply failure classification: transaction_timeout=%t deadline=%t canceled=%t", errors.Is(err, crossownertransaction.ErrTimeout), errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled))
	return f.ExtensionImportFacade.TranslateImportUnitError(err)
}

func sourceAdmissionHarness(t *testing.T, overrides map[string]string) (*appsupport.ServerHarness, authcookietest.AuthCookies, string) {
	t.Helper()
	env := map[string]string{
		"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                "true",
		"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": fixtures.Path("network-flow", "key-rings.json"),
		"CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR":                "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
		"CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST":           "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
	}
	for k, v := range overrides {
		env[k] = v
	}
	timing := &sourceAdmissionTiming{}
	if t.Name() == networkFlowDefaultLimitTestName {
		t.Cleanup(func() {
			root, err := fixtureartifacts.SelectedUnitRoot()
			if err != nil {
				t.Error(err)
				return
			}
			timing.Lock()
			defer timing.Unlock()
			raw, err := json.MarshalIndent(timing.passes, "", "  ")
			if err != nil {
				t.Error(err)
				return
			}
			if err := os.WriteFile(filepath.Join(root, "network-flow-import-source-passes.json"), append(raw, '\n'), 0600); err != nil {
				t.Error(err)
			}
		})
	}
	runtime := appsupport.StartRuntime(t)
	h := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "target-neutral-source", Env: env, TestRouteMode: httptestx.TestRouteModeDisabled,
		ConfigureRuntime: func(options *server.Options) {
			options.NetworkFlowComposition = sourceAdmissionDiagnostics{t: t, timing: timing}
		},
	})
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "source-admission-incident", "incident_key": "SOURCE-ADMISSION", "title": "Source admission"})
	return h, login, incident["incident_id"].(string)
}

func TestNetworkFlowPublicSourceAdmission(t *testing.T) {
	h, login, incident := sourceAdmissionHarness(t, map[string]string{
		"CARTULARY__LIMITS__IMPORTS__MAX_ROWS":             "1",
		"CARTULARY__LIMITS__IMPORTS__MAX_COLUMNS":          "1",
		"CARTULARY__LIMITS__IMPORTS__MAX_CELLS":            "1",
		"CARTULARY__LIMITS__IMPORTS__MAX_CSV_SOURCE_BYTES": "32768",
	})
	for _, tc := range []struct {
		name, body                       string
		previewOK, applyOK, genericFails bool
		accepted                         int
	}{
		{"literal_quotes", strings.Replace(sourceAdmissionRow, "Gi0/1", "Gi\"0/1", 1), true, true, true, 1},
		{"generic_limits_do_not_gate_csv", strings.Repeat(sourceAdmissionRow, 2), true, true, false, 2},
		{"early_bad_quote", "\"unclosed\n", false, false, true, 0},
		{"early_bad_encoding", strings.Replace(sourceAdmissionRow, "Gi0/1", "\xff", 1), false, false, true, 0},
		{"late_bad_quote", strings.Repeat(sourceAdmissionRow, 50) + "\"unclosed\n", true, false, true, 0},
		{"late_bad_encoding", strings.Repeat(sourceAdmissionRow, 50) + strings.Replace(sourceAdmissionRow, "Gi0/1", "\xff", 1), true, false, true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			session, unit := startCSVImportSession(t, h.Server.HTTP.URL, login, incident, "upload-"+tc.name, sourceAdmissionHeader+tc.body, tc.name+".csv")
			base := "/api/v1/import-sessions/" + session + "/units/" + unit
			if tc.genericFails {
				resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, h.Server.HTTP.URL+base+"/preview", nil, httptestx.WithCookies(login.Session))
				httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "import_source_unsupported")
				mapping := map[string]any{"client_txn_id": "generic-reject-" + tc.name, "target_view_schema_id": "cartulary.view.timeline.v2", "header_row_ref": 1, "data_start_row_ref": 2, "unknown_column_policy": "preserve_raw_capture", "source_columns": []map[string]any{{"source_column_ordinal": 1, "source_header_text": "Source IP Address", "field_key": "timeline.activity_synopsis_text", "entity_binding_mode": nil, "transform_id": nil, "transform_options": map[string]any{}, "empty_value_policy": "omit_field"}}}
				approval := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPut, base+"/mapping", mapping)
				httptestx.RequireErrorEnvelope(t, approval, http.StatusConflict, "import_source_unsupported")
				if dbassert.CountSQL(t, h.DB, `SELECT count(*) FROM import_unit_projection_rows WHERE import_unit_id::text=$1`, unit) != 0 || dbassert.CountSQL(t, h.DB, `SELECT count(*) FROM import_units WHERE import_unit_id::text=$1 AND approved_mapping_json IS NOT NULL`, unit) != 0 {
					t.Fatal("invalid interpretation published mapping or staged rows")
				}
			} else if tc.name == "generic_limits_do_not_gate_csv" {
				resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, h.Server.HTTP.URL+base+"/preview", nil, httptestx.WithCookies(login.Session))
				httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
			}
			assertPublicAnalyticalSource(t, h, login, session, unit, sourceAdmissionHeader+tc.body, tc.previewOK, tc.applyOK, tc.accepted)
		})
	}
	t.Run("owner_column_limit", func(t *testing.T) {
		session, unit := startCSVImportSession(t, h.Server.HTTP.URL, login, incident, "owner-column-limit", strings.Repeat("column,", 256)+"last\n", "columns.csv")
		resp := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, "/api/v1/import-sessions/"+session+"/units/"+unit+"/source-inspection", map[string]any{"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity"})
		httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_import_request")
	})
	t.Run("byte_overflow_has_no_durable_admission", func(t *testing.T) {
		counts := []int{}
		for _, table := range []string{"import_sessions", "jobs", "route_idempotency"} {
			counts = append(counts, dbassert.CountSQL(t, h.DB, "SELECT count(*) FROM "+table))
		}
		metadata := fmt.Sprintf(`{"client_txn_id":"byte-overflow","incident_id":%q}`, incident)
		resp := postImportUpload(t, h.Server.HTTP.URL, login, metadata, strings.Repeat("x", 32769), "overflow.csv", false)
		httptestx.RequireErrorEnvelope(t, resp, http.StatusRequestEntityTooLarge, "import_source_rejected")
		for i, table := range []string{"import_sessions", "jobs", "route_idempotency"} {
			if got := dbassert.CountSQL(t, h.DB, "SELECT count(*) FROM "+table); got != counts[i] {
				t.Fatalf("overflow persisted %s", table)
			}
		}
	})
}

const networkFlowDefaultLimitTestName = "TestNetworkFlowPublicDefaultLimitsCommitAcceptedRowLimit"

// This is NF-REQ-171a's retained capacity evidence: with every limit and
// timeout at its default, a unit of exactly the default accepted-row limit
// commits within the default participant budget. The harness runs it in the
// standard placement, alongside check's other concurrent work. 250,000 rows
// also exceed Workbook's former 100,000-row ceiling, so the public
// source-admission assertions stay. The poll's own deadline only bounds the
// observation; the unit fails with a retryable timeout when it overruns the
// participant budget.
func TestNetworkFlowPublicDefaultLimitsCommitAcceptedRowLimit(t *testing.T) {
	h, login, incident := sourceAdmissionHarness(t, nil)
	limits := h.Server.Config.NetworkFlowActivity.EffectiveResourceLimits()
	if limits.MaxAcceptedRowsPerTable != 250000 || limits.MaxRowsPerCSV != 250000 || h.Server.Config.Timeouts.Extensions.TransactionParticipantSeconds != 30 {
		t.Fatalf("defaults are not in effect: accepted=%d rows=%d participant_seconds=%d", limits.MaxAcceptedRowsPerTable, limits.MaxRowsPerCSV, h.Server.Config.Timeouts.Extensions.TransactionParticipantSeconds)
	}
	rows := int(limits.MaxAcceptedRowsPerTable)
	content := sourceAdmissionHeader + strings.Repeat(sourceAdmissionRow, rows)
	if int64(len(content)) > h.Server.Config.Limits.Imports.MaxCSVSourceBytes {
		t.Fatalf("default-limit source of %d bytes exceeds the default CSV source bound", len(content))
	}
	session, unit := startCSVImportSession(t, h.Server.HTTP.URL, login, incident, "large-source", content, "large.csv")
	assertPublicAnalyticalSource(t, h, login, session, unit, content, true, true, rows)
}

func assertPublicAnalyticalSource(t *testing.T, h *appsupport.ServerHarness, login authcookietest.AuthCookies, session, unit, content string, previewOK, applyOK bool, accepted int) {
	t.Helper()
	base := "/api/v1/import-sessions/" + session + "/units/" + unit
	inspection := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, base+"/source-inspection", map[string]any{"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity"})
	owner := httptestx.RequireSuccessEnvelope(t, inspection, http.StatusOK)["data"].(map[string]any)["owner_result"].(map[string]any)
	wantSHA := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
	if owner["source_content_sha256"] != wantSHA {
		t.Fatalf("inspection source digest = %v, want %s", owner["source_content_sha256"], wantSHA)
	}
	wantHeaders := strings.Split(strings.TrimSuffix(sourceAdmissionHeader, "\n"), ",")
	columns := owner["source_columns"].([]any)
	if len(columns) != len(wantHeaders) {
		t.Fatalf("inspection columns = %d, want %d", len(columns), len(wantHeaders))
	}
	for i, raw := range columns {
		column := raw.(map[string]any)
		if column["source_column_ordinal"] != float64(i+1) || column["raw_header_text"] != wantHeaders[i] {
			t.Fatalf("inspection column %d has incorrect ordinal or raw header", i+1)
		}
	}
	preview := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, base+"/mapping-preview", networkFlowMappingPreviewPayload())
	if !previewOK {
		httptestx.RequireErrorEnvelope(t, preview, http.StatusBadRequest, "invalid_import_request")
		return
	}
	previewResult := httptestx.RequireSuccessEnvelope(t, preview, http.StatusOK)["data"].(map[string]any)["owner_result"].(map[string]any)
	if previewResult["source_content_sha256"] != wantSHA {
		t.Fatalf("preview source digest = %v, want %s", previewResult["source_content_sha256"], wantSHA)
	}
	fingerprint, ok := previewResult["mapping_fingerprint"].(string)
	decoded, err := hex.DecodeString(fingerprint)
	if !ok || err != nil || len(decoded) != sha256.Size || fingerprint != strings.ToLower(fingerprint) {
		t.Fatal("preview mapping fingerprint is not lower-hex SHA-256")
	}
	mapping := networkFlowMappingPayload("map-" + unit)
	echoes := mapping["source_columns"].([]map[string]any)
	for i, raw := range columns {
		column := raw.(map[string]any)
		echoes[i]["source_header_text"] = column["raw_header_text"]
		echoes[i]["source_column_ordinal"] = column["source_column_ordinal"]
	}
	resp := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPut, base+"/mapping", mapping)
	approved := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	if approved["mapping_fingerprint"] != fingerprint {
		t.Fatal("approval differs from preview")
	}
	resp = doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, base+"/select", map[string]any{"client_txn_id": "select-" + unit})
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	applyBody := map[string]any{"client_txn_id": "apply-" + unit}
	resp = doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, "/api/v1/import-sessions/"+session+"/apply", applyBody)
	jobID := httptestx.RequireSuccessEnvelope(t, resp, http.StatusAccepted)["data"].(map[string]any)["job_id"].(string)
	// The declared large-input bound is the poll's own deadline; if it
	// expires, the poll records the database's lock waits.
	job := httptestx.PollJob(t, h.Server.HTTP.URL, jobID, httptestx.PollOptions{
		Timeout: 2 * time.Minute, Interval: 50 * time.Millisecond, Request: []func(*http.Request){httptestx.WithCookies(login.Session)}, LockWaits: h.DB,
	}, "succeeded", "failed", "canceled")
	wantStatus := "failed"
	if applyOK {
		wantStatus = "succeeded"
	}
	if job["status"] != wantStatus {
		var code, reason sql.NullString
		err := h.DB.QueryRowContext(t.Context(), `SELECT error_code,reason_code FROM import_unit_apply_outcomes WHERE import_unit_id::text=$1`, unit).Scan(&code, &reason)
		t.Fatalf("apply status=%v want=%s code=%q reason=%q outcome_read=%v", job["status"], wantStatus, code.String, reason.String, err)
	}
	wantTables := 0
	if applyOK {
		wantTables = 1
	}
	if got := dbassert.CountSQL(t, h.DB, `SELECT count(*) FROM network_flow_tables t JOIN import_committed_sources s USING(committed_source_id) WHERE s.origin_import_session_id::text=$1`, session); got != wantTables {
		t.Fatalf("atomic table count=%d want=%d", got, wantTables)
	}
	if got := dbassert.CountSQL(t, h.DB, `SELECT count(*) FROM network_flow_rows r JOIN network_flow_tables t USING(network_flow_table_id) JOIN import_committed_sources s ON s.committed_source_id=t.committed_source_id WHERE s.origin_import_session_id::text=$1`, session); got != accepted {
		t.Fatalf("accepted rows=%d want=%d", got, accepted)
	}
	var materialized bool
	if err := h.DB.QueryRowContext(context.Background(), `SELECT source_rows_json IS NOT NULL FROM import_units WHERE import_unit_id::text=$1`, unit).Scan(&materialized); err != nil || materialized {
		t.Fatalf("analytical workflow materialized generic content: %v", err)
	}
	resp = doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, "/api/v1/import-sessions/"+session+"/apply", applyBody)
	replayed := httptestx.RequireSuccessEnvelope(t, resp, http.StatusAccepted)["data"].(map[string]any)
	if replayed["job_id"] != jobID {
		t.Fatal("apply replay replaced job identity")
	}
}

func TestAnalyticalReadOnlyPOSTAuthentication(t *testing.T) {
	h, login, incident := sourceAdmissionHarness(t, nil)
	session, unit := startCSVImportSession(t, h.Server.HTTP.URL, login, incident, "read-only-auth", sourceAdmissionHeader+sourceAdmissionRow, "read-only.csv")
	before := importsSecurityState(t, h.DB)
	for _, suffix := range []string{"source-inspection", "mapping-preview"} {
		body := map[string]any{"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity"}
		if suffix == "mapping-preview" {
			body = networkFlowMappingPreviewPayload()
		}
		path := h.Server.HTTP.URL + "/api/v1/import-sessions/" + session + "/units/" + unit + "/" + suffix
		unauthorized := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, path, body)
		httptestx.RequireErrorEnvelope(t, unauthorized, http.StatusUnauthorized, "session_required")
		cookie := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, path, body, httptestx.WithCookies(login.Session))
		httptestx.RequireSuccessEnvelope(t, cookie, http.StatusOK)
		bearer := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, path, body, httptestx.WithHeader("Authorization", "Bearer "+login.Session.Value))
		httptestx.RequireSuccessEnvelope(t, bearer, http.StatusOK)
	}
	after := importsSecurityState(t, h.DB)
	for key, want := range before {
		if after[key] != want {
			t.Fatalf("read-only POST changed %s", key)
		}
	}
}

func TestGenericCSVRaggedPublicLifecycle(t *testing.T) {
	h, login, incident := sourceAdmissionHarness(t, nil)
	session, unit := startCSVImportSession(t, h.Server.HTTP.URL, login, incident, "ragged-upload", "summary,optional,unmapped\nshort\nexplicit,,\n", "ragged.csv")
	base := "/api/v1/import-sessions/" + session + "/units/" + unit
	preview := func() {
		t.Helper()
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, h.Server.HTTP.URL+base+"/preview", nil, httptestx.WithCookies(login.Session))
		result := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
		rows := result["preview_rows"].([]any)
		if len(result["columns"].([]any)) != 3 || len(rows) != 2 {
			t.Fatal("wrong ragged preview geometry")
		}
		for i, want := range []int{1, 3} {
			if len(rows[i].(map[string]any)["cells"].([]any)) != want {
				t.Fatalf("row %d was padded", i)
			}
		}
	}
	preview()
	if dbassert.CountSQL(t, h.DB, `SELECT count(*) FROM import_units WHERE import_unit_id::text=$1 AND source_rows_json IS NULL AND approved_mapping_json IS NULL`, unit) != 1 {
		t.Fatal("preview wrote projection")
	}
	for range 2 {
		approveTimelineImportMapping(t, h.Server.HTTP.URL, login, session, unit, []string{"summary", "optional", "unmapped"}, "ragged")
	}
	preview()
	var count int
	if err := h.DB.QueryRowContext(t.Context(), `SELECT sum(jsonb_array_length(convert_from(row_json,'UTF8')::jsonb->'cells')) FROM import_unit_projection_rows WHERE import_unit_id::text=$1`, unit).Scan(&count); err != nil || count != 4 {
		t.Fatalf("materialized cell count %d: %v", count, err)
	}
	selected := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, base+"/select", map[string]any{"client_txn_id": "ragged-select"})
	httptestx.RequireSuccessEnvelope(t, selected, http.StatusOK)
	var jobID string
	for range 2 {
		response := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, "/api/v1/import-sessions/"+session+"/apply", map[string]any{"client_txn_id": "ragged-apply"})
		job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
		id := job["job_id"].(string)
		if jobID != "" && id != jobID {
			t.Fatal("apply replay created another job")
		}
		jobID = id
		terminal := waitImportJobTerminal(t, h.Server.HTTP.URL, login, id)
		if terminal["status"] != "succeeded" {
			t.Fatalf("ragged apply: %#v", terminal)
		}
	}
	if got := dbassert.CountSQL(t, h.DB, `SELECT count(*) FROM timeline_events WHERE incident_id::text=$1`, incident); got != 2 {
		t.Fatalf("owner rows %d", got)
	}
	preview()
}

// crossOwnerDeadlineComposition holds the apply source read until the unit's
// cross-owner transaction context ends. With a one-second transaction budget,
// the owner write is interrupted before commit without depending on load.
type crossOwnerDeadlineComposition struct{}

func (crossOwnerDeadlineComposition) WrapDatabase(db dbport.DB) dbport.DB { return db }
func (crossOwnerDeadlineComposition) ConfigureNetworkFlow(deps networkflow.ModuleDependencies) networkflow.ModuleDependencies {
	deps.ImportSources = crossOwnerDeadlinePort{ImportSourcePort: deps.ImportSources}
	return deps
}
func (crossOwnerDeadlineComposition) WrapImportFacade(f imports.ExtensionImportFacade) imports.ExtensionImportFacade {
	return f
}

type crossOwnerDeadlinePort struct{ networkflow.ImportSourcePort }

func (p crossOwnerDeadlinePort) OpenSourceStreamTx(ctx context.Context, tx pgx.Tx, ref string) (imports.ImportSourceStream, error) {
	stream, err := p.ImportSourcePort.OpenSourceStreamTx(ctx, tx, ref)
	if err != nil {
		return stream, err
	}
	stream.Reader = &crossOwnerDeadlineReader{ReadCloser: stream.Reader, ctx: ctx}
	return stream, nil
}

type crossOwnerDeadlineReader struct {
	io.ReadCloser
	ctx context.Context
}

func (r *crossOwnerDeadlineReader) Read([]byte) (int, error) {
	<-r.ctx.Done()
	return 0, r.ctx.Err()
}

// A unit whose cross-owner transaction times out before commit gets the Core
// unit outcome of Core 01 REQ-01-620c, never the owner-translation fallback,
// and leaves no unit effect.
func assertCrossOwnerTimeoutBeforeCommitUnitOutcome(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	h := runtime.StartServer(t, appsupport.ServerOptions{
		Prefix: "network-flow-import-transaction-timeout",
		Env: map[string]string{
			"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                        "true",
			"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH":         fixtures.Path("network-flow", "key-rings.json"),
			"CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR":                        "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
			"CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST":                   "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
			"CARTULARY__TIMEOUTS__EXTENSIONS__TRANSACTION_PARTICIPANT_SECONDS": "1",
		},
		TestRouteMode: httptestx.TestRouteModeDisabled,
		ConfigureRuntime: func(options *server.Options) {
			options.NetworkFlowComposition = crossOwnerDeadlineComposition{}
		},
	})
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, h.Server, login, map[string]any{
		"client_txn_id": "txn-network-flow-import-timeout-incident", "incident_key": "IR-NF-IMPORT-TIMEOUT", "title": "Network Flow transaction timeout",
	})
	incidentID := incident["incident_id"].(string)
	session, unit := startCSVImportSession(t, h.Server.HTTP.URL, login, incidentID, "txn-network-flow-import-timeout-upload", sourceAdmissionHeader+sourceAdmissionRow, "timeout.csv")
	base := "/api/v1/import-sessions/" + session
	httptestx.RequireSuccessEnvelope(t, doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPut, base+"/units/"+unit+"/mapping", networkFlowMappingPayload("txn-network-flow-import-timeout-mapping")), http.StatusOK)
	httptestx.RequireSuccessEnvelope(t, doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, base+"/units/"+unit+"/select", map[string]any{"client_txn_id": "txn-network-flow-import-timeout-select"}), http.StatusOK)
	applied := httptestx.RequireSuccessEnvelope(t, doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, base+"/apply", map[string]any{"client_txn_id": "txn-network-flow-import-timeout-apply"}), http.StatusAccepted)["data"].(map[string]any)
	job := waitImportJobTerminal(t, h.Server.HTTP.URL, login, applied["job_id"].(string))
	summary, _ := job["error_summary"].(map[string]any)
	wantDetails := map[string]any{
		"reason_code":     "extension_transaction_timeout",
		"operation_id":    "import-unit:" + session + ":" + unit,
		"timeout_seconds": float64(1),
	}
	if job["status"] != "failed" || summary["code"] != "service_unavailable" || summary["retryable"] != true || !reflect.DeepEqual(summary["details"], wantDetails) {
		t.Fatalf("timed-out unit job = status %v, error_summary %#v", job["status"], summary)
	}
	if got := dbassert.CountSQL(t, h.DB, `
SELECT count(*) FROM import_unit_apply_outcomes
 WHERE import_session_id::text = $1 AND import_unit_id::text = $2 AND outcome_status = 'failed'
   AND error_code = 'service_unavailable' AND reason_code = 'extension_transaction_timeout' AND error_retryable`, session, unit); got != 1 {
		t.Fatalf("durable timed-out unit outcome count = %d, want 1", got)
	}
	for query, arg := range map[string]string{
		`SELECT count(*) FROM network_flow_tables WHERE incident_id::text = $1`:                   incidentID,
		`SELECT count(*) FROM network_flow_rows WHERE incident_id::text = $1`:                     incidentID,
		`SELECT count(*) FROM import_committed_sources WHERE origin_import_session_id::text = $1`: session,
	} {
		if got := dbassert.CountSQL(t, h.DB, query, arg); got != 0 {
			t.Fatalf("timed-out unit left %d effects: %s", got, query)
		}
	}
	// A fatal process answers every request with 503 extension_integrity_failure.
	httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, h.Server.HTTP.URL+"/api/v1/jobs/"+applied["job_id"].(string), nil, httptestx.WithCookies(login.Session)), http.StatusOK)
}

// unitCommitHoldComposition holds a view-schema unit transaction's COMMIT
// until the attempt context that began the transaction ends, as when the job
// runner stops or loses the execution while the commit is issued.
type unitCommitHoldComposition struct {
	entered chan struct{}
	once    *sync.Once
}

func (c unitCommitHoldComposition) WrapDatabase(db dbport.DB) dbport.DB {
	return unitCommitHoldDB{DB: db, composition: c}
}
func (unitCommitHoldComposition) ConfigureNetworkFlow(deps networkflow.ModuleDependencies) networkflow.ModuleDependencies {
	return deps
}
func (unitCommitHoldComposition) WrapImportFacade(f imports.ExtensionImportFacade) imports.ExtensionImportFacade {
	return f
}

type unitCommitHoldDB struct {
	dbport.DB
	composition unitCommitHoldComposition
}

func (d unitCommitHoldDB) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := d.DB.BeginTx(ctx, options)
	if err != nil {
		return tx, err
	}
	return &unitCommitHoldTx{Tx: tx, begun: ctx, composition: d.composition}, nil
}

type unitCommitHoldTx struct {
	pgx.Tx
	begun       context.Context
	composition unitCommitHoldComposition
	unitOutcome bool
}

func (t *unitCommitHoldTx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	if strings.Contains(sql, "INSERT INTO import_unit_apply_outcomes") {
		t.unitOutcome = true
	}
	return t.Tx.Exec(ctx, sql, arguments...)
}

func (t *unitCommitHoldTx) Commit(ctx context.Context) error {
	if t.unitOutcome {
		t.composition.once.Do(func() { close(t.composition.entered) })
		<-t.begun.Done()
	}
	return t.Tx.Commit(ctx)
}

// The job runner stops while the unit transaction's COMMIT is issued: it
// cancels the attempt and conditionally releases the execution. The unit
// transaction's Jobs-locked validation already decided that race, so the
// issued commit is awaited and its applied outcome is retained (REQ-01-620c).
func assertIssuedUnitCommitSurvivesAttemptCancellation(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	database := runtime.PrepareIsolatedDatabase(t, "imports-issued-unit-commit")
	durableObjects, err := objectstore.NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatalf("create durable import object store: %v", err)
	}
	hold := unitCommitHoldComposition{entered: make(chan struct{}), once: &sync.Once{}}
	h := runtime.StartServer(t, appsupport.ServerOptions{
		Prefix: "imports-issued-unit-commit", Database: database, ObjectStore: durableObjects, TestRouteMode: httptestx.TestRouteModeDisabled,
		ConfigureRuntime: func(options *server.Options) { options.NetworkFlowComposition = hold },
	})
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, h.Server, login, map[string]any{
		"client_txn_id": "txn-imports-issued-commit-incident", "incident_key": "IR-IMPORT-ISSUED-COMMIT", "title": "Issued unit commit",
	})
	sessionID, unitID := prepareReadyTimelineImport(t, h, login, incident["incident_id"].(string), "imports-issued-commit")
	applied := httptestx.RequireSuccessEnvelope(t, doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/apply", map[string]any{"client_txn_id": "txn-imports-issued-commit-apply"}), http.StatusAccepted)["data"].(map[string]any)
	select {
	case <-hold.entered:
	case <-time.After(30 * time.Second):
		t.Fatalf("apply job %v never issued its unit commit", applied["job_id"])
	}
	closed := make(chan struct{})
	go func() {
		defer close(closed)
		h.Server.Close()
	}()
	select {
	case <-closed:
	case <-time.After(60 * time.Second):
		t.Fatal("server did not stop after the unit commit was released")
	}
	if got := dbassert.CountSQL(t, h.DB, `
SELECT count(*)
  FROM import_unit_apply_outcomes o
  JOIN import_units u USING (import_session_id, import_unit_id)
 WHERE o.import_session_id::text = $1 AND o.import_unit_id::text = $2
   AND o.outcome_status = 'applied' AND u.unit_status = 'applied'`, sessionID, unitID); got != 1 {
		t.Fatalf("issued unit commit lost to attempt cancellation: applied outcomes = %d", got)
	}
}
