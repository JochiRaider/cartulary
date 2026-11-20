package serverprocess

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	evidenceflowtest "github.com/JochiRaider/cartulary/internal/modules/evidence/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

// Shared fixtures of the AC-616 conformance matrices. A matrix keys its case
// table by operation ID and derives the expected key set from the generated
// operation catalog, so an operation without a case fails the matrix by name.
// A case prepares, through public routes only, resources and fresh requests
// that would succeed but for the condition the matrix exercises.
//
// The request shape is matrix-neutral: path parameters fill the operation's
// catalog template, and the caller is chosen at send time. The non-member
// concealment matrix reuses the same case builders, sends each request as a
// caller without visibility, and replaces the addressed resource's parameter
// with an unknown identifier to build the missing resource's request.

const conformanceTimelineView = "cartulary.view.timeline.v2"

// conformanceRequest is one fresh request to one catalog operation.
type conformanceRequest struct {
	label string
	// params fill every {name} of the operation's PathTemplate.
	params map[string]string
	// query is an encoded query string, without the leading "?".
	query string
	// body is the JSON body, or a func(*testing.T) any evaluated at send
	// time, with the sending test, for a value that must be current when sent,
	// such as an incident version.
	body any
	// raw is a non-JSON body; headers then carry its declared content type.
	raw     []byte
	headers map[string]string
}

type conformanceEnv struct {
	t          *testing.T
	harness    *appsupport.ServerHarness
	login      authcookietest.AuthCookies
	actorID    uuid.UUID
	incidentID uuid.UUID
	operations map[string]httpapi.ContractOperation
	sequence   int
}

func startConformanceEnv(t *testing.T, prefix string, incidentKey string) *conformanceEnv {
	t.Helper()
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: prefix, TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actorID := authflowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-" + incidentKey,
		"incident_key":  incidentKey,
		"title":         "Conformance " + incidentKey,
	})
	return &conformanceEnv{
		t:          t,
		harness:    harness,
		login:      login,
		actorID:    actorID,
		incidentID: appsupport.MustUUID(t, incident["incident_id"].(string)),
		operations: catalogOperationsByID(t),
	}
}

func catalogOperationsByID(t *testing.T) map[string]httpapi.ContractOperation {
	t.Helper()
	operations := map[string]httpapi.ContractOperation{}
	for _, operation := range httpapi.ContractOperations() {
		operations[operation.OperationID] = operation
	}
	if len(operations) == 0 {
		t.Fatal("generated operation catalog is empty")
	}
	return operations
}

func declaresErrorCode(operation httpapi.ContractOperation, status int, code string) bool {
	return slices.Contains(operation.ErrorCodes[status], code)
}

// requireCaseKeySet fails, naming every difference, unless the case table is
// keyed by exactly the catalog-derived operation set.
func requireCaseKeySet(t *testing.T, matrix string, cases []string, want []string) {
	t.Helper()
	var missing, extra []string
	for _, id := range want {
		if !slices.Contains(cases, id) {
			missing = append(missing, id)
		}
	}
	for _, id := range cases {
		if !slices.Contains(want, id) {
			extra = append(extra, id)
		}
	}
	if len(missing) != 0 || len(extra) != 0 {
		slices.Sort(missing)
		slices.Sort(extra)
		t.Fatalf("%s matrix key set differs from the operation catalog: operations without a case %v; cases for operations outside the catalog-derived set %v", matrix, missing, extra)
	}
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

var pathParameterPattern = regexp.MustCompile(`\{([a-z_]+)\}`)

func pathParameters(operation httpapi.ContractOperation) []string {
	var names []string
	for _, match := range pathParameterPattern.FindAllStringSubmatch(operation.PathTemplate, -1) {
		names = append(names, match[1])
	}
	return names
}

// recordPathParameters returns the path parameters that identify a record:
// record_id, a role-qualified <role>_record_id, or <record_type>_id for a
// record type a public view sources. Child objects (mentions, observations,
// Blobs), incidents, jobs and other resources are not records.
func recordPathParameters(operation httpapi.ContractOperation, recordTypes []string) []string {
	var names []string
	for _, name := range pathParameters(operation) {
		if name == "record_id" || strings.HasSuffix(name, "_record_id") ||
			slices.Contains(recordTypes, strings.TrimSuffix(name, "_id")) {
			names = append(names, name)
		}
	}
	return names
}

func publicRecordTypes() []string {
	var types []string
	for _, resource := range viewschema.ListPublicResources() {
		for _, recordType := range resource.SourceRecordTypes {
			if !slices.Contains(types, recordType) {
				types = append(types, recordType)
			}
		}
	}
	slices.Sort(types)
	return types
}

func createCapableViews() []string {
	var views []string
	for _, resource := range viewschema.ListPublicResources() {
		if resource.CreateCapable {
			views = append(views, resource.ViewSchemaID)
		}
	}
	slices.Sort(views)
	return views
}

func (e *conformanceEnv) txn(label string) string {
	e.sequence++
	return fmt.Sprintf("conformance-%03d-%s", e.sequence, label)
}

func (e *conformanceEnv) authOptions(caller authcookietest.AuthCookies) []func(*http.Request) {
	return []func(*http.Request){httptestx.WithCookies(caller.Session, caller.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, caller.CSRF.Value)}
}

func (e *conformanceEnv) do(method string, path string, body any) *http.Response {
	e.t.Helper()
	return httptestx.DoJSON(e.t, http.DefaultClient, method, e.harness.Server.HTTP.URL+path, body, e.authOptions(e.login)...)
}

func (e *conformanceEnv) require(method string, path string, body any, status int) map[string]any {
	e.t.Helper()
	return httptestx.RequireSuccessData(e.t, e.do(method, path, body), status)
}

// send issues request to operation as caller. The path is the operation's
// catalog template, so a request can only address the operation it is keyed by.
func (e *conformanceEnv) send(t *testing.T, operation httpapi.ContractOperation, caller authcookietest.AuthCookies, request conformanceRequest) *http.Response {
	t.Helper()
	used := map[string]bool{}
	path := pathParameterPattern.ReplaceAllStringFunc(operation.PathTemplate, func(segment string) string {
		name := segment[1 : len(segment)-1]
		value, ok := request.params[name]
		if !ok {
			t.Fatalf("%s request %q lacks path parameter %s", operation.OperationID, request.label, name)
		}
		used[name] = true
		return url.PathEscape(value)
	})
	for name := range request.params {
		if !used[name] {
			t.Fatalf("%s request %q names path parameter %s outside %s", operation.OperationID, request.label, name, operation.PathTemplate)
		}
	}
	options := e.authOptions(caller)
	for name, value := range request.headers {
		options = append(options, httptestx.WithHeader(name, value))
	}
	target := e.harness.Server.HTTP.URL + path
	if request.query != "" {
		target += "?" + request.query
	}
	if request.raw != nil {
		req, err := http.NewRequest(operation.Method, target, bytes.NewReader(request.raw))
		if err != nil {
			t.Fatalf("build %s request: %v", operation.OperationID, err)
		}
		for _, option := range options {
			option(req)
		}
		return httptestx.Do(t, http.DefaultClient, req)
	}
	body := request.body
	if late, ok := body.(func(*testing.T) any); ok {
		body = late(t)
	}
	return httptestx.DoJSON(t, http.DefaultClient, operation.Method, target, body, options...)
}

// requireRejected sends request and requires the error envelope and no
// committed effect of any AC-616 class.
func (e *conformanceEnv) requireRejected(t *testing.T, operation httpapi.ContractOperation, request conformanceRequest, status int, code string) {
	t.Helper()
	before := workbookscenariotest.SnapshotIncidentEffects(t, e.harness, e.incidentID)
	response := e.send(t, operation, e.login, request)
	if response.StatusCode != status {
		t.Fatalf("%s %q: status %d, want %d %s; body=%#v", operation.OperationID, request.label, response.StatusCode, status, code, httptestx.ReadJSONBody(t, response))
	}
	httptestx.RequireErrorEnvelope(t, response, status, code)
	workbookscenariotest.RequireIncidentEffectsUnchanged(t, operation.OperationID+" "+request.label, before, workbookscenariotest.SnapshotIncidentEffects(t, e.harness, e.incidentID))
}

// quiesce stops asynchronous Collaboration delivery so background sequencing
// cannot race effect snapshots; transactional publication stays observable.
func (e *conformanceEnv) quiesce() {
	e.t.Helper()
	if err := e.harness.Collaboration.CloseDispatcher(context.Background()); err != nil {
		e.t.Fatalf("close Collaboration dispatcher: %v", err)
	}
}

func (e *conformanceEnv) incidentPath() string {
	return "/api/v1/incidents/" + e.incidentID.String()
}

func (e *conformanceEnv) incidentVersion(t testing.TB) int64 {
	t.Helper()
	return int64(appsupport.QueryCount(t, e.harness.DB, `SELECT incident_version FROM incidents WHERE id = $1`, e.incidentID))
}

func (e *conformanceEnv) closeIncident() {
	e.t.Helper()
	e.require(http.MethodPost, e.incidentPath()+"/close", map[string]any{
		"base_incident_version": e.incidentVersion(e.t), "client_txn_id": e.txn("close"), "reason": "Conformance closure",
	}, http.StatusOK)
}

func (e *conformanceEnv) version(recordID uuid.UUID) int64 {
	e.t.Helper()
	return int64(appsupport.QueryCount(e.t, e.harness.DB, `SELECT row_version FROM records WHERE record_id = $1`, recordID))
}

func (e *conformanceEnv) createRow(view string, fields map[string]any) map[string]any {
	e.t.Helper()
	body := map[string]any{"client_txn_id": e.txn("create")}
	for key, value := range fields {
		body[key] = value
	}
	return e.require(http.MethodPost, e.incidentPath()+"/views/"+view+"/rows", body, http.StatusCreated)["row"].(map[string]any)
}

func (e *conformanceEnv) createRecord(view string, fields map[string]any) uuid.UUID {
	e.t.Helper()
	return appsupport.MustUUID(e.t, e.createRow(view, fields)["record_id"].(string))
}

func (e *conformanceEnv) timeline(text string) uuid.UUID {
	e.t.Helper()
	return e.createRecord(conformanceTimelineView, map[string]any{
		"timeline.activity_synopsis_text": text, "timeline.raw_activity_text": text,
	})
}

func (e *conformanceEnv) note(title string) uuid.UUID {
	e.t.Helper()
	return e.createRecord("cartulary.view.notes.v1", map[string]any{"note.title": title, "note.body": "Conformance body"})
}

func (e *conformanceEnv) host(name string) uuid.UUID {
	e.t.Helper()
	return e.createRecord("cartulary.view.hosts.v1", map[string]any{"host.display_name": "Conformance " + name, "host.hostname": name + ".conformance.test"})
}

func (e *conformanceEnv) indicator(value string) uuid.UUID {
	e.t.Helper()
	return e.createRecord("cartulary.view.indicators.v1", map[string]any{
		"indicator.indicator_type": "domain_name", "indicator.value_kind": "atomic", "indicator.display_value": value,
	})
}

func (e *conformanceEnv) evidence(title string) uuid.UUID {
	e.t.Helper()
	return e.createRecord("cartulary.view.evidence.v1", map[string]any{"evidence.title": title})
}

func (e *conformanceEnv) patch(recordID uuid.UUID, view string, field string, value any, base int64) *http.Response {
	e.t.Helper()
	return e.do(http.MethodPatch, "/api/v1/records/"+recordID.String(), map[string]any{
		"view_schema_id": view, "base_row_version": base, "client_txn_id": e.txn("patch"),
		"changes": []map[string]any{{"field_key": field, "value": value}},
	})
}

// softDelete deletes recordID through the public soft-delete route.
func (e *conformanceEnv) softDelete(recordID uuid.UUID) {
	e.t.Helper()
	e.require(http.MethodDelete, "/api/v1/records/"+recordID.String(), map[string]any{
		"base_row_version": e.version(recordID), "client_txn_id": e.txn("delete"), "reason": "Conformance deletion",
	}, http.StatusOK)
}

// timelineConflict saves a Timeline synopsis, then submits a stale one, and
// returns the record and its same-field conflict token.
func (e *conformanceEnv) timelineConflict(text string) (uuid.UUID, string) {
	e.t.Helper()
	recordID := e.timeline(text)
	httptestx.RequireSuccessData(e.t, e.patch(recordID, conformanceTimelineView, "timeline.activity_synopsis_text", text+" saved", 1), http.StatusOK)
	stale := httptestx.RequireErrorEnvelope(e.t, e.patch(recordID, conformanceTimelineView, "timeline.activity_synopsis_text", text+" unsaved", 1), http.StatusConflict, "same_field_conflict")
	return recordID, stale["error"].(map[string]any)["conflict"].(map[string]any)["conflict_token"].(string)
}

// conflictResolutions resolves token both ways: keep_saved, which commits only
// a receipt, and use_unsaved, which commits a value.
func (e *conformanceEnv) conflictResolutions(recordID uuid.UUID, token string) []conformanceRequest {
	params := map[string]string{"record_id": recordID.String(), "conflict_token": token}
	return []conformanceRequest{
		{label: "Timeline keep_saved", params: params, body: map[string]any{
			"conflict_token": token, "resolution_kind": "keep_saved", "client_txn_id": e.txn("keep-saved"),
		}},
		{label: "Timeline use_unsaved", params: params, body: map[string]any{
			"conflict_token": token, "resolution_kind": "use_unsaved", "client_txn_id": e.txn("use-unsaved"), "resolved_value": "Conformance resolved value",
		}},
	}
}

// unresolvedMention captures a Timeline host token and a Host to resolve it
// to, created after the token so the mention stays unresolved.
func (e *conformanceEnv) unresolvedMention(label string) (sourceID uuid.UUID, mentionID uuid.UUID, hostID uuid.UUID) {
	e.t.Helper()
	row := e.createRow(conformanceTimelineView, map[string]any{
		"timeline.activity_synopsis_text": "Mention source " + label,
		"timeline.host_refs": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{
			{"op": "add_token", "raw_text": "unmatched " + label + " token"},
		}},
	})
	items := row["cells"].(map[string]any)["timeline.host_refs"].(map[string]any)["value"].(map[string]any)["items"].([]any)
	if len(items) != 1 {
		e.t.Fatalf("mention fixture %s captured %d host items", label, len(items))
	}
	return appsupport.MustUUID(e.t, row["record_id"].(string)),
		entitytest.MentionIDFromItemRef(e.t, items[0].(map[string]any)["item_ref"].(string)),
		e.host("mention-target-" + label)
}

func (e *conformanceEnv) mentionResolution(mentionID uuid.UUID, hostID uuid.UUID) conformanceRequest {
	return conformanceRequest{
		label:  "resolve to a Host",
		params: map[string]string{"entity_mention_id": mentionID.String()},
		body:   entitytest.MentionResolveRoutePayload(1, e.txn("mention"), entitytest.MentionActionResolve, &hostID, nil),
	}
}

func (e *conformanceEnv) manualObservation(sourceID uuid.UUID, field string, text string) conformanceRequest {
	return conformanceRequest{
		label:  "manual observation",
		params: map[string]string{"source_record_id": sourceID.String()},
		body: map[string]any{
			"client_txn_id": e.txn("observation"), "base_row_version": e.version(sourceID),
			"source_field_key": field, "span_start_byte": 0, "span_end_byte": len(text),
		},
	}
}

// observation commits a manual observation over the whole of field's text.
func (e *conformanceEnv) observation(sourceID uuid.UUID, field string, text string) uuid.UUID {
	e.t.Helper()
	request := e.manualObservation(sourceID, field, text)
	created := e.require(http.MethodPost, "/api/v1/records/"+sourceID.String()+"/indicator-observations", request.body, http.StatusCreated)
	return appsupport.MustUUID(e.t, created["observation"].(map[string]any)["observation_id"].(string))
}

func (e *conformanceEnv) stateInterval(indicatorID uuid.UUID, supportID uuid.UUID) conformanceRequest {
	return conformanceRequest{
		label:  "state interval",
		params: map[string]string{"indicator_id": indicatorID.String()},
		body: map[string]any{
			"client_txn_id": e.txn("interval"), "base_row_version": e.version(indicatorID),
			"lifecycle_state": "false_positive", "valid_from": "2026-08-03T20:00:00Z", "valid_to": nil,
			"confidence": 80, "rationale": "Conformance interval", "support_refs": []any{supportID.String()}, "assessor": "Conformance reviewer",
		},
	}
}

func (e *conformanceEnv) linkedNote(sourceID uuid.UUID) conformanceRequest {
	return conformanceRequest{
		label:  "linked Note under a Host",
		params: map[string]string{"record_id": sourceID.String()},
		body:   map[string]any{"client_txn_id": e.txn("linked-note"), "note.title": "Conformance linked Note", "note.body": "Linked body"},
	}
}

func (e *conformanceEnv) noteAssociation(noteID uuid.UUID, counterpartID uuid.UUID) conformanceRequest {
	return conformanceRequest{
		label:  "related Note",
		params: map[string]string{"note_record_id": noteID.String()},
		body: map[string]any{
			"kind": "related_note", "base_row_version": e.version(noteID), "client_txn_id": e.txn("association"),
			"actions": []map[string]any{{"op": "add", "counterpart_record_id": counterpartID.String()}},
		},
	}
}

func (e *conformanceEnv) timelineAction(recordID uuid.UUID, label string) conformanceRequest {
	return conformanceRequest{
		label:  label,
		params: map[string]string{"record_id": recordID.String()},
		body:   map[string]any{"base_row_version": e.version(recordID), "client_txn_id": e.txn(label), "reason": "Conformance " + label},
	}
}

func (e *conformanceEnv) notePatch(noteID uuid.UUID) conformanceRequest {
	return conformanceRequest{
		label:  "Note title",
		params: map[string]string{"record_id": noteID.String()},
		body: map[string]any{
			"view_schema_id": "cartulary.view.notes.v1", "base_row_version": e.version(noteID), "client_txn_id": e.txn("note-patch"),
			"changes": []map[string]any{{"field_key": "note.title", "value": "Conformance patched title"}},
		},
	}
}

// uploadSlot is a created Blob slot and its upload capability.
type uploadSlot struct {
	blobID  string
	token   string
	headers map[string]string
	payload []byte
}

func (e *conformanceEnv) blobCreateBody(label string, payload []byte) map[string]any {
	sum := sha256.Sum256(payload)
	return map[string]any{
		"incident_id": e.incidentID.String(), "client_txn_id": e.txn(label), "byte_size": len(payload),
		"filename_hint": label + ".txt", "content_type_hint": "text/plain", "sha256_hex": fmt.Sprintf("%x", sum[:]),
	}
}

func (e *conformanceEnv) blobSlot(label string) uploadSlot {
	e.t.Helper()
	payload := []byte("conformance Blob " + label)
	created := e.require(http.MethodPost, "/api/v1/object-blobs", e.blobCreateBody(label, payload), http.StatusCreated)
	target := created["upload_target"].(map[string]any)
	href, err := url.Parse(target["href"].(string))
	if err != nil {
		e.t.Fatalf("parse upload target: %v", err)
	}
	token, ok := strings.CutPrefix(href.Path, "/api/v1/object-uploads/")
	if !ok || token == "" {
		e.t.Fatalf("upload target %q is not an object upload", href.Path)
	}
	headers := map[string]string{}
	for name, value := range target["headers"].(map[string]any) {
		headers[name] = value.(string)
	}
	return uploadSlot{blobID: created["object_blob_id"].(string), token: token, headers: headers, payload: payload}
}

func (slot uploadSlot) upload() conformanceRequest {
	return conformanceRequest{label: "Blob content", params: map[string]string{"upload_token": slot.token}, raw: slot.payload, headers: slot.headers}
}

// uploadedBlob is a slot whose content is stored but not yet verified.
func (e *conformanceEnv) uploadedBlob(label string) uploadSlot {
	e.t.Helper()
	slot := e.blobSlot(label)
	response := e.send(e.t, e.operations["uploadObjectBlobContent"], e.login, slot.upload())
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		e.t.Fatalf("upload Blob fixture %s: status %d body=%#v", label, response.StatusCode, httptestx.ReadJSONBody(e.t, response))
	}
	response.Body.Close()
	return slot
}

func (e *conformanceEnv) verifiedBlob(label string) string {
	e.t.Helper()
	slot := e.uploadedBlob(label)
	evidenceflowtest.VerifyUploadedBlob(e.t, http.DefaultClient, e.harness.Server.HTTP.URL, e.login.Session, e.login.CSRF, slot.blobID, e.txn(label+"-verify"))
	return slot.blobID
}

func (e *conformanceEnv) attach(evidenceID uuid.UUID, blobID string) conformanceRequest {
	return conformanceRequest{
		label:  "verified Blob",
		params: map[string]string{"record_id": evidenceID.String()},
		body:   map[string]any{"object_blob_id": blobID, "base_row_version": e.version(evidenceID), "client_txn_id": e.txn("attach")},
	}
}

// createViewPayloads holds one valid create per create-capable view.
func (e *conformanceEnv) createViewPayloads(subjectHostID uuid.UUID, supportID uuid.UUID) map[string]map[string]any {
	return map[string]map[string]any{
		conformanceTimelineView:           {"timeline.activity_synopsis_text": "Conformance created timeline"},
		"cartulary.view.hosts.v1":         {"host.hostname": "created.conformance.test", "host.display_name": "Conformance created host"},
		"cartulary.view.identities.v1":    {"identity.email": "created@conformance.test", "identity.display_name": "Conformance identity"},
		"cartulary.view.parties.v1":       {"party.party_kind": "person", "party.display_name": "Conformance party"},
		"cartulary.view.indicators.v1":    {"indicator.indicator_type": "domain_name", "indicator.value_kind": "atomic", "indicator.display_value": "created.conformance.test"},
		"cartulary.view.task_requests.v1": {"task.task_kind": "question", "task.title": "Conformance task"},
		"cartulary.view.decisions.v1":     {"decision.decision_type": "scope", "decision.rationale": "Conformance", "decision.summary": "Conformance decision"},
		"cartulary.view.evidence.v1":      {"evidence.title": "Conformance created evidence"},
		"cartulary.view.findings.v1":      {"finding.statement": "Conformance finding"},
		"cartulary.view.notes.v1":         {"note.title": "Conformance created note", "note.body": "Body"},
		"cartulary.view.comm_log.v1": {
			"comm_log.comm_type": "briefing", "comm_log.audience": "Operations", "comm_log.channel_or_meeting": "Bridge", "comm_log.summary": "Conformance communication",
		},
		"cartulary.view.lesson.v1":        {"lesson.summary": "Conformance lesson"},
		"cartulary.view.handoff.v1":       {"handoff.incoming_owner_user_id": e.actorID.String(), "handoff.current_state_summary": "Conformance handoff"},
		"cartulary.view.status_review.v1": {"status_review.current_state_summary": "Conformance status review"},
		"cartulary.view.forensic_keywords.v1": {
			"forensic_keyword.reason": "Conformance fixture", "forensic_keyword.pattern": "conformance-keyword",
		},
		"cartulary.view.investigative_queries.v1": {
			"investigative_query.platform": "SQL", "investigative_query.query_text": "select value", "investigative_query.purpose": "Conformance query",
		},
		"cartulary.view.assessments.v1": {
			"assessment.subject_ref": subjectHostID.String(), "assessment.subject_type": "host", "assessment.assessment_state": "confirmed",
			"assessment.confidence_score": 85, "assessment.assessed_at": "2026-09-10T00:00:00Z", "assessment.rationale": "Conformance assessment",
			"assessment.support_refs": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{
				{"op": "add_record_ref", "linked_record_id": supportID.String()},
			}},
		},
	}
}
