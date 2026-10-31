package networkflow_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/server"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	hc "github.com/JochiRaider/cartulary/internal/modules/networkflow/harnesscontrol"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowfixtures"
	"github.com/google/uuid"
)

// Business timestamps use Network Flow's existing injected clock. Authentication
// and Jobs keep wall time, including PostgreSQL's authoritative lease fence.
type publicationComposition struct {
	*hc.Controls
	now       time.Time
	source    *publicationSourceChange
	configure []func(ModuleDependencies) ModuleDependencies
}

func (c publicationComposition) ConfigureNetworkFlow(d ModuleDependencies) ModuleDependencies {
	d = c.Controls.ConfigureNetworkFlow(d)
	d.Now = func() time.Time { return c.now }
	c.source.ImportSourcePort = d.ImportSources
	d.ImportSources = c.source
	for _, configure := range c.configure {
		d = configure(d)
	}
	return d
}

// The fixture starts with a discovered, source-only CSV unit. Approval, plans,
// owner effects, publication and replay are produced by the serving application.
// Auth's supported actor/session fixtures pin identity without password material.
type publicationFaultCase struct {
	Boundary    string `json:"boundary"`
	Kind        string `json:"kind"`
	ExpectRetry bool   `json:"expect_retry"`
}

type publicationFixtureCase struct {
	environment    map[string]string
	ResourceLimits map[string]int64             `json:"resource_limits,omitempty"`
	QueryLimit     int                          `json:"query_limit,omitempty"`
	TxnPrefix      string                       `json:"txn_prefix,omitempty"`
	Name           string                       `json:"name"`
	Source         string                       `json:"source"`
	Filename       string                       `json:"filename"`
	ActorID        string                       `json:"actor_id"`
	IncidentID     string                       `json:"incident_id"`
	SessionID      string                       `json:"session_id"`
	UnitID         string                       `json:"unit_id"`
	TableUUID      string                       `json:"table_uuid"`
	Mapping        json.RawMessage              `json:"mapping"`
	Graph          bool                         `json:"graph"`
	Fault          *publicationFaultCase        `json:"fault,omitempty"`
	SourceChange   *publicationSourceChangeCase `json:"source_change,omitempty"`
}

func (c publicationFixtureCase) txn(action string) string {
	prefix := c.TxnPrefix
	if prefix == "" {
		prefix = "fixture"
	}
	return prefix + "-" + action
}

type publicationFixtureRuntime struct {
	graphLimits    EffectiveLimits
	currentSession string
	authSession    uuid.UUID
	t              *testing.T
	h              *appsupport.ServerHarness
	requestOptions []func(*http.Request)
	transport      []map[string]any
	controls       *hc.Controls
	started        time.Time
	source         *publicationSourceChange
}

func startPublicationFixture(t *testing.T, input publicationFixtureCase, source []byte, configure ...func(ModuleDependencies) ModuleDependencies) *publicationFixtureRuntime {
	t.Helper()
	limitJSON, graphLimits := publicationResourceLimits(t, input)
	controls := hc.NewControls()
	sourceChange := &publicationSourceChange{}
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	environment := map[string]string{
		"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                "true",
		"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": fixtures.Path("network-flow", "key-rings.json"),
		"CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR":                "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
		"CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST":           "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
		"CARTULARY__NETWORK_FLOW_ACTIVITY__RESOURCE_LIMITS":        limitJSON,
	}
	for key, value := range input.environment {
		environment[key] = value
	}
	h := appsupport.StartRuntime(t).StartServer(t, appsupport.ServerOptions{
		Prefix: "nf-fixture-publication", Env: environment,
		TestRouteMode:    httptestx.TestRouteModeHarnessOwned,
		AdditionalRoutes: controls.Contribution().Routes,
		ConfigureRuntime: func(o *server.Options) {
			o.NetworkFlowComposition = publicationComposition{Controls: controls, now: now, source: sourceChange, configure: configure}
		},
	})
	actorID, incidentID := uuid.MustParse(input.ActorID), uuid.MustParse(input.IncidentID)
	flowtest.SeedActor(t, h.DB, flowtest.ActorAttributes{ID: actorID, Email: "nf-fixture@example.test", DisplayName: "Fixture analyst", DeploymentAdmin: true})
	keys, err := authn.LoadMasterKeys(map[string]string{authn.AuthMasterKeyEnv: strings.TrimSpace(string(fixtures.MustRead("auth", "master-key.base64")))})
	if err != nil {
		t.Fatal(err)
	}
	const sessionToken = "network-flow-publication-fixture-session"
	session := flowtest.SeedSession(t, h.DB, keys, actorID, sessionToken, time.Now().UTC(), time.Now().UTC())
	csrf := authn.CSRFTokenForSessionToken(keys, sessionToken)
	runtime := &publicationFixtureRuntime{graphLimits: graphLimits, t: t, h: h, authSession: session.ID, controls: controls, source: sourceChange, started: time.Now().UTC(), transport: []map[string]any{}, requestOptions: []func(*http.Request){
		httptestx.WithCookies(&http.Cookie{Name: authn.SessionCookieName, Value: sessionToken}, &http.Cookie{Name: authn.CSRFCookieName, Value: csrf}),
		httptestx.WithHeader(authn.CSRFHeaderName, csrf),
	}}
	if _, err := h.DB.ExecContext(context.Background(), `INSERT INTO incidents
		(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id,created_at,updated_at)
		VALUES ($1,'IR-NF-FIXTURE','ir-nf-fixture','Network Flow fixture','active',$2,$2,$3,$3)`, incidentID, actorID, now); err != nil {
		t.Fatal(err)
	}
	appsupport.SeedMembershipSQL(t, h.DB, incidentID, actorID, "admin", actorID)
	runtime.addSource(input, source)
	return runtime
}

func (r *publicationFixtureRuntime) request(method, path string, body any, status int) map[string]any {
	r.t.Helper()
	resp := httptestx.DoJSON(r.t, http.DefaultClient, method, r.h.Server.HTTP.URL+path, body, r.requestOptions...)
	raw, err := io.ReadAll(io.LimitReader(resp.Body, (4<<20)+1))
	_ = resp.Body.Close()
	if err != nil || len(raw) > 4<<20 {
		r.t.Fatalf("fixture response read failed or exceeded retention bound: %v", err)
	}
	resp.Body = io.NopCloser(bytes.NewReader(raw))
	r.transport = append(r.transport, map[string]any{"method": method, "path": path, "body": body, "status": resp.StatusCode, "response": fixtureResponseEvidence(raw)})
	response := httptestx.RequireSuccessEnvelope(r.t, resp, status)
	return response["data"].(map[string]any)
}

func (r *publicationFixtureRuntime) control(path string, body any) {
	r.t.Helper()
	response := httptestx.RequireSuccessEnvelope(r.t, httptestx.DoJSON(r.t, http.DefaultClient, http.MethodPost, r.h.Server.HTTP.URL+path, body, httptestx.WithHeader("X-Cartulary-Test-Route-Token", httptestx.TestRouteToken)), http.StatusCreated)
	r.transport = append(r.transport, map[string]any{"control": path, "body": body, "response": response})
}

func (r *publicationFixtureRuntime) jsonQuery(query string, args ...any) any {
	r.t.Helper()
	var raw []byte
	if err := r.h.DB.QueryRowContext(context.Background(), query, args...).Scan(&raw); err != nil {
		r.t.Fatal(err)
	}
	var result any
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		r.t.Fatal(err)
	}
	return result
}

func (r *publicationFixtureRuntime) durableState() map[string]any {
	return map[string]any{
		"tables":      r.jsonQuery(`SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY network_flow_table_id),'[]') FROM network_flow_tables t`),
		"rows":        r.jsonQuery(`SELECT coalesce(jsonb_agg(to_jsonb(r) ORDER BY network_flow_table_id,source_row_number),'[]') FROM network_flow_rows r`),
		"diagnostics": r.jsonQuery(`SELECT coalesce(jsonb_agg(to_jsonb(d) ORDER BY network_flow_table_id,source_row_number,diagnostic_id),'[]') FROM network_flow_rejected_row_diagnostics d`),
		"audit":       r.jsonQuery(`SELECT coalesce(jsonb_agg(to_jsonb(a) ORDER BY created_at,id),'[]') FROM deployment_admin_audit_events a WHERE event_source='network_flow'`),
		"publication": r.jsonQuery(`SELECT jsonb_build_object(
   'plans',(SELECT count(*) FROM import_apply_unit_plans),
   'outcomes',(SELECT count(*) FROM import_unit_apply_outcomes),
   'journal',(SELECT count(*) FROM import_apply_journal),
   'apply_receipts',(SELECT count(*) FROM route_idempotency WHERE route_key='imports.sessions.apply'),
   'apply_jobs',(SELECT count(*) FROM jobs WHERE job_kind='import.apply_v1'),
   'outcome_bindings',(SELECT count(*) FROM import_unit_apply_outcomes o JOIN import_apply_unit_plans p
    ON p.import_unit_id=o.import_unit_id AND p.import_session_id=o.import_session_id AND p.apply_job_id=o.apply_job_id
    JOIN jobs j ON j.job_id=o.apply_job_id JOIN import_sessions s ON s.import_session_id=o.import_session_id
    JOIN import_units u ON u.import_unit_id=o.import_unit_id
    WHERE o.source_content_sha256=p.source_content_sha256 AND o.mapping_fingerprint=p.mapping_fingerprint
     AND o.target_kind=p.target_kind AND o.owner_binding_id=p.owner_binding_id AND o.actor_user_id=j.submitted_by_user_id
     AND j.incident_id=s.incident_id AND p.approved_mapping_json=u.approved_mapping_json
     AND o.unit_commit_id='import-unit:'||o.import_session_id::text||':'||o.import_unit_id::text))`),
		"import_state": r.jsonQuery(`SELECT jsonb_build_object('session_status',s.session_status,'unit_status',u.unit_status,'selected_unit_ids',s.selected_unit_ids) FROM import_sessions s JOIN import_units u USING(import_session_id) WHERE s.import_session_id=$1`, r.currentSession),
		"outcomes":     r.jsonQuery(`SELECT coalesce(jsonb_agg(jsonb_build_object('outcome_status',outcome_status,'owner_result',owner_result_json,'resource_refs',resource_refs_json,'error_code',error_code,'reason_code',reason_code,'error_retryable',error_retryable,'error_details',error_details_json) ORDER BY import_unit_id),'[]') FROM import_unit_apply_outcomes`),
	}
}

// Opaque audit identities/times are retained in full and checked for stable
// replay; the frozen comparison covers every semantic audit member.
func (r *publicationFixtureRuntime) comparableState(state map[string]any) map[string]any {
	result := map[string]any{}
	for k, v := range state {
		result[k] = v
	}
	audit := []map[string]any{}
	for _, entry := range state["audit"].([]any) {
		raw := entry.(map[string]any)
		if _, err := uuid.Parse(raw["id"].(string)); err != nil {
			r.t.Fatal(err)
		}
		at, err := time.Parse(time.RFC3339Nano, raw["created_at"].(string))
		if err != nil || at.Before(r.started) || at.After(time.Now().UTC().Add(time.Second)) {
			r.t.Fatal("audit timestamp is outside the operation interval")
		}
		if requestID := raw["request_id"]; requestID != nil {
			if text, ok := requestID.(string); !ok || text == "" {
				r.t.Fatal("invalid audit request identity")
			}
		}
		item := map[string]any{}
		for k, v := range raw {
			if k != "id" && k != "created_at" && k != "request_id" {
				item[k] = v
			}
		}
		audit = append(audit, item)
	}
	result["audit"] = audit
	return result
}
func (r *publicationFixtureRuntime) interruptedAttempt(jobID string) map[string]any {
	r.t.Helper()
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		var failures int
		if err := r.h.DB.QueryRowContext(context.Background(), `SELECT handler_failure_count FROM jobs WHERE job_id=$1`, jobID).Scan(&failures); err != nil {
			r.t.Fatal(err)
		}
		if failures > 0 {
			if failures != 1 {
				r.t.Fatal("fault did not produce exactly one interrupted attempt")
			}
			return r.jsonQuery(`SELECT jsonb_build_object('tables',(SELECT count(*) FROM network_flow_tables),'rows',(SELECT count(*) FROM network_flow_rows),'diagnostics',(SELECT count(*) FROM network_flow_rejected_row_diagnostics),'bindings',(SELECT count(*) FROM network_flow_indicator_bindings),'applied_outcomes',(SELECT count(*) FROM import_unit_apply_outcomes WHERE outcome_status='applied'),'audit',(SELECT count(*) FROM deployment_admin_audit_events WHERE event_source='network_flow'))`).(map[string]any)
		}
		time.Sleep(10 * time.Millisecond)
	}
	r.t.Fatal("expected interrupted apply attempt did not occur")
	return nil
}

func (r *publicationFixtureRuntime) waitJob(id string) map[string]any {
	r.t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		response := httptestx.RequireSuccessEnvelope(r.t, httptestx.DoJSON(r.t, http.DefaultClient, http.MethodGet, r.h.Server.HTTP.URL+"/api/v1/jobs/"+id, nil, r.requestOptions...), http.StatusOK)
		job := response["data"].(map[string]any)
		if job["status"] == "succeeded" || job["status"] == "failed" || job["status"] == "canceled" {
			return job
		}
		time.Sleep(25 * time.Millisecond)
	}
	state := r.jsonQuery(`SELECT jsonb_build_object('status',status,'failure_count',handler_failure_count,'last_error',handler_last_error,'last_attempt',handler_last_attempted_at,'next_attempt',handler_next_attempt_at) FROM jobs WHERE job_id=$1`, id)
	r.t.Fatalf("fixture apply job did not reach a terminal state: %v", state)
	return nil
}

func publicationFixtureOutcomes(t *testing.T, bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
	data, err := bundle.Read("source/cases.json")
	if err != nil {
		return nil, err
	}
	var inputs []publicationFixtureCase
	if err := json.Unmarshal(data, &inputs); err != nil {
		return nil, err
	}
	outcomes := make([]map[string]any, 0, len(inputs))
	for _, input := range inputs {
		var outcome map[string]any
		var caseErr error
		passed := t.Run(input.Name, func(t *testing.T) { outcome, caseErr = publicationFixtureOutcome(t, bundle, input) })
		if caseErr != nil {
			return nil, caseErr
		}
		if !passed {
			return nil, fmt.Errorf("%s: serving adapter failed", input.Name)
		}
		outcomes = append(outcomes, outcome)
	}
	return outcomes, nil
}
func publicationFixtureOutcome(t *testing.T, bundle *networkflowfixtures.Bundle, input publicationFixtureCase) (map[string]any, error) {
	return runPublicationFixture(t, bundle, input, nil, nil)
}
func runPublicationFixture(t *testing.T, bundle *networkflowfixtures.Bundle, input publicationFixtureCase, r *publicationFixtureRuntime, after func(*publicationFixtureRuntime, map[string]any) error, configure ...func(ModuleDependencies) ModuleDependencies) (result map[string]any, resultErr error) {

	source, err := bundle.Read(input.Source)
	if err != nil {
		return nil, err
	}
	if r != nil && len(configure) != 0 {
		return nil, errors.New("fixture configuration requires a new isolated runtime")
	}
	if r == nil {
		r = startPublicationFixture(t, input, source, configure...)
	} else {
		r.addSource(input, source)
	}
	defer func() {
		if result != nil && resultErr == nil && after != nil {
			resultErr = after(r, result)
		}
		encoded, encodeErr := retainedPublicationTransport(r.transport)
		if encodeErr != nil {
			resultErr = errors.Join(resultErr, encodeErr)
			return
		}
		resultErr = errors.Join(resultErr, bundle.Record(input.Name+"-transport.json", append(encoded, '\n')))
	}()
	unitPath := "/api/v1/import-sessions/" + input.SessionID + "/units/" + input.UnitID
	root := "/api/v1/incidents/" + input.IncidentID + "/network-flow"
	before := r.durableState()
	inspection := r.request(http.MethodPost, unitPath+"/source-inspection", map[string]any{"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity"}, http.StatusOK)
	previewBody := map[string]any{"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity", "owner_mapping_schema_id": "cartulary.network_flow.mapping_candidate.v1", "owner_mapping": input.Mapping}
	preview := r.request(http.MethodPost, unitPath+"/mapping-preview", previewBody, http.StatusOK)
	if repeated := r.request(http.MethodPost, unitPath+"/mapping-preview", previewBody, http.StatusOK); !reflect.DeepEqual(preview, repeated) {
		return nil, fmt.Errorf("%s: preview changed on repetition", input.Name)
	}
	if !reflect.DeepEqual(before, r.durableState()) {
		return nil, fmt.Errorf("%s: preview published durable effects", input.Name)
	}
	approvalBody := map[string]any{}
	for key, value := range previewBody {
		approvalBody[key] = value
	}
	approvalBody["client_txn_id"] = input.txn("approve")
	approvalBody["header_row_ref"], approvalBody["data_start_row_ref"] = 1, 2
	columns := []map[string]any{}
	for _, raw := range inspection["owner_result"].(map[string]any)["source_columns"].([]any) {
		column := raw.(map[string]any)
		columns = append(columns, map[string]any{"source_column_ordinal": column["source_column_ordinal"], "source_header_text": column["raw_header_text"], "field_key": nil, "entity_binding_mode": nil, "transform_id": nil, "transform_options": map[string]any{}, "empty_value_policy": "omit_field"})
	}
	approvalBody["source_columns"] = columns
	approval := r.request(http.MethodPut, unitPath+"/mapping", approvalBody, http.StatusOK)
	selected := r.request(http.MethodPost, unitPath+"/select", map[string]any{"client_txn_id": input.txn("select")}, http.StatusOK)
	applyPath := "/api/v1/import-sessions/" + input.SessionID + "/apply"
	applyBody := map[string]any{"client_txn_id": input.txn("apply")}
	if input.Fault != nil {
		r.control("/api/v1/test/runtime/network-flow-faults", map[string]any{"boundary": input.Fault.Boundary, "fault_kind": input.Fault.Kind, "consume_once": true, "correlation_key": input.UnitID, "error_code": "harness_import_failure"})
	}
	if input.SourceChange != nil {
		if err := r.armSourceChange(bundle, input.SourceChange); err != nil {
			return nil, err
		}
	}
	job := r.request(http.MethodPost, applyPath, applyBody, http.StatusAccepted)
	jobID, ok := job["job_id"].(string)
	if !ok {
		return nil, fmt.Errorf("%s: missing apply job identity", input.Name)
	}
	if _, err := uuid.Parse(jobID); err != nil {
		return nil, err
	}
	var interrupted map[string]any
	if input.Fault != nil && input.Fault.ExpectRetry {
		interrupted = r.interruptedAttempt(jobID)
	}
	terminal := r.waitJob(jobID)
	if input.Fault != nil {
		if _, pending := r.controls.Faults.ConsumeNetworkFlowFaultFor(input.Fault.Boundary, input.UnitID); pending {
			return nil, fmt.Errorf("%s: real apply did not consume the registered fault", input.Name)
		}
	}
	r.transport = append(r.transport, map[string]any{"terminal_job": terminal})
	if terminal["status"] != "succeeded" && terminal["status"] != "failed" {
		return nil, fmt.Errorf("%s: unexpected apply result: %v", input.Name, terminal)
	}
	durable := r.durableState()
	replay := r.request(http.MethodPost, applyPath, applyBody, http.StatusAccepted)
	// Success finalization publishes the terminal resource in the receipt;
	// failure retains the original admission receipt. GET remains authoritative.
	wantReplay := job
	if terminal["status"] == "succeeded" {
		wantReplay = terminal
	}
	current := r.request(http.MethodGet, "/api/v1/jobs/"+jobID, nil, http.StatusOK)
	if !reflect.DeepEqual(wantReplay, replay) || !reflect.DeepEqual(terminal, current) || !reflect.DeepEqual(durable, r.durableState()) {
		return nil, fmt.Errorf("%s: replay changed response or durable effects", input.Name)
	}
	tableID := "nft_" + strings.ReplaceAll(input.TableUUID, "-", "")
	list := r.request(http.MethodGet, root+"/tables", nil, http.StatusOK)
	outcome := map[string]any{"case": input.Name, "request": input, "inspection": inspection, "preview": preview, "approval": approval, "selected": selected, "job_status": terminal["status"], "job_result": terminal["result_summary"], "durable": r.comparableState(durable), "table_list": list, "replay_unchanged": true}
	r.transport = append(r.transport, map[string]any{"durable_state": durable})
	if input.SourceChange != nil {
		outcome["source_change"] = r.source.outcome()
		if err := r.sourceChangePrivacy(input.SourceChange); err != nil {
			return nil, err
		}
	}
	if interrupted != nil {
		outcome["interrupted_attempt"] = interrupted
	}
	if terminal["status"] == "failed" {
		outcome["job_error"] = terminal["error_summary"]
		return outcome, nil
	}
	queryLimit := input.QueryLimit
	if queryLimit == 0 {
		queryLimit = 10
	}
	outcome["rows"] = r.request(http.MethodPost, root+"/tables/"+tableID+"/query", map[string]any{"schema_id": "cartulary.network_flow.table_query_request.v1", "limit": queryLimit}, http.StatusOK)
	outcome["diagnostics"] = r.request(http.MethodPost, root+"/tables/"+tableID+"/rejected-rows/query", map[string]any{"schema_id": "cartulary.network_flow.rejected_rows_query_request.v1", "limit": queryLimit}, http.StatusOK)

	if input.Graph {
		outcome["graph"] = r.request(http.MethodPost, root+"/graphs/query", map[string]any{"schema_id": "cartulary.network_flow.graph_query_request.v2", "table_scope": map[string]any{"mode": "active_table", "active_table_id": tableID}, "aggregation": map[string]any{"mode": "default_flow_edge_v1", "include_example_row_refs": true}}, http.StatusOK)
	}
	return outcome, nil
}

func TestNetworkFlowMinimalImportFixtureComparison(t *testing.T) {
	RunFixtureComparison(t, "NF-FIX-001-cisco-sna-minimal", 8, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		return publicationFixtureOutcomes(t, bundle)
	})
}

func TestNetworkFlowInterfacesImportFixtureComparison(t *testing.T) {
	RunFixtureComparison(t, "NF-FIX-002-cisco-sna-interface-fields", 5, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		return publicationFixtureOutcomes(t, bundle)
	})
}

func TestNetworkFlowRejectedImportFixtureComparison(t *testing.T) {
	RunFixtureComparison(t, "NF-FIX-004-rejected-rows", 6, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		return publicationFixtureOutcomes(t, bundle)
	})
}

func TestNetworkFlowAtomicImportFixtureComparison(t *testing.T) {
	RunFixtureComparison(t, "NF-FIX-020-atomic-import-commit", 3, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		return publicationFixtureOutcomes(t, bundle)
	})
}

func (r *publicationFixtureRuntime) addSource(input publicationFixtureCase, source []byte) {
	t := r.t
	r.currentSession = input.SessionID
	actorID, incidentID := uuid.MustParse(input.ActorID), uuid.MustParse(input.IncidentID)
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	digest := sha256.Sum256(source)
	sourceHash := hex.EncodeToString(digest[:])
	streamRef := "impsrc_" + strings.ReplaceAll(input.UnitID, "-", "")
	tx, err := r.h.DB.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`INSERT INTO import_sessions
		(import_session_id,incident_id,created_by_user_id,client_txn_id,assistant_profile,source_file_kind,original_filename,
		source_content_sha256,source_media_type,source_byte_size,parser_profile_id,parser_version,session_status,created_at,updated_at)
		VALUES ($1,$2,$3,$8,'workbook_import_v1','csv',$4,$5,'text/csv',$6,
		'cartulary.import.workbook.v1','workbook_import_adapter_v1','discovered',$7,$7)`, input.SessionID, incidentID, actorID, input.Filename, sourceHash, len(source), now, input.txn("source")); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`INSERT INTO import_units
		(import_unit_id,import_session_id,unit_status,locator_kind,locator,source_stream_ref,discovery_sequence,created_at,updated_at)
		VALUES ($1,$2,'discovered','csv_file','file',$3,1,$4,$4)`, input.UnitID, input.SessionID, streamRef, now); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`INSERT INTO import_source_streams
		(source_stream_ref,import_session_id,import_unit_id,source_content_sha256,source_media_type,source_byte_size,source_bytes,created_at)
		VALUES ($1,$2,$3,$4,'text/csv',$5,$6,$7)`, streamRef, input.SessionID, input.UnitID, sourceHash, len(source), source, now); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	values := []string{input.TableUUID}
	if input.Fault != nil && input.Fault.ExpectRetry {
		values = append(values, input.TableUUID)
	}
	// A completed prior import may leave an exhausted or unused fixture stream.
	r.controls.Randomness.Clear()
	r.control("/api/v1/test/runtime/network-flow-randomness", map[string]any{
		"stream": hc.NetworkFlowRandomStreamTableID, "value_kind": "uuid", "values": values,
		"consume_once": true, "exhaustion": "fail_closed",
	})
}

// Cursor values are authenticated capabilities. Compare them in memory, then
// retain a bound digest under the harness's existing secret-redaction policy.
func retainedPublicationTransport(value any) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var copy any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&copy); err != nil {
		return nil, err
	}
	var redact func(any)
	redact = func(v any) {
		switch item := v.(type) {
		case map[string]any:
			for k, child := range item {
				if k == "cursor_token" || k == "next_cursor_token" {
					if token, ok := child.(string); ok && token != "" {
						digest := sha256.Sum256([]byte(token))
						item[k] = "<redacted>:cursor:sha256:" + hex.EncodeToString(digest[:])
						continue
					}
				}
				redact(child)
			}
		case []any:
			for _, child := range item {
				redact(child)
			}
		}
	}
	redact(copy)
	return EncodeFixtureEvidence(copy)
}

// Preserve unexpected non-JSON bodies byte-for-byte so an envelope failure
// cannot prevent the surrounding transport evidence from being published.
func fixtureResponseEvidence(raw []byte) any {
	if json.Valid(raw) {
		return json.RawMessage(raw)
	}
	return map[string]any{"non_json_body_base64": base64.StdEncoding.EncodeToString(raw)}
}
