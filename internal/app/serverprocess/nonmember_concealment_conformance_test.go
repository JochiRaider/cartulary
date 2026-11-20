package serverprocess

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	importtest "github.com/JochiRaider/cartulary/internal/modules/imports/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

// Reasons an operation addresses no resource inside an incident.
const (
	ownAccountReason        = "addresses the caller's own account, credentials or session, not an incident resource"
	enterpriseAuthReason    = "enterprise authentication flow outside any incident; addresses an auth provider"
	deploymentAdminReason   = "deployment administration of users, auth bindings, audit or extensions; gated by deployment role, never by an incident"
	referencePackReason     = "deployment-wide reference-pack administration and reads"
	viewSchemaReason        = "deployment-wide view-schema contract reads"
	incidentDirectoryReason = "lists only the incidents the caller can see and addresses none"
	incidentCreationReason  = "creates a new incident, so no existing incident is addressed"
	concealedBodyLimit      = 1 << 16
)

// notIncidentScoped classifies, with its reason, every catalog operation that
// addresses no resource inside an incident. Every other operation is
// incident scoped and has a case in concealmentCases.
var notIncidentScoped = map[string]string{
	"getCurrentAccountPreferences": ownAccountReason,
	"putCurrentAccountPreferences": ownAccountReason,
	"getCurrentAccountProfile":     ownAccountReason,
	"patchCurrentAccountProfile":   ownAccountReason,
	"getCredentialState":           ownAccountReason,
	"loginLocalUser":               ownAccountReason,
	"logoutCurrentSession":         ownAccountReason,
	"beginTOTPEnrollment":          ownAccountReason,
	"completeTOTPEnrollment":       ownAccountReason,
	"changeCurrentPassword":        ownAccountReason,
	"getCurrentSession":            ownAccountReason,

	"completeEnterpriseOIDC":      enterpriseAuthReason,
	"listEnterpriseAuthProviders": enterpriseAuthReason,
	"beginEnterpriseAuth":         enterpriseAuthReason,
	"completeEnterpriseSAML":      enterpriseAuthReason,
	"finishEnterpriseSAML":        enterpriseAuthReason,

	"listAdministrativeAuditEvents":   deploymentAdminReason,
	"listDeploymentExtensions":        deploymentAdminReason,
	"listDeploymentUsers":             deploymentAdminReason,
	"createDeploymentUser":            deploymentAdminReason,
	"getDeploymentUser":               deploymentAdminReason,
	"patchDeploymentUser":             deploymentAdminReason,
	"createEnterpriseAuthBinding":     deploymentAdminReason,
	"retireEnterpriseAuthBinding":     deploymentAdminReason,
	"rotateEnterpriseAuthBinding":     deploymentAdminReason,
	"resetDeploymentUserTOTP":         deploymentAdminReason,
	"resetDeploymentUserPassword":     deploymentAdminReason,
	"revokeAllDeploymentUserSessions": deploymentAdminReason,

	"listReferencePacks":                referencePackReason,
	"importReferencePack":               referencePackReason,
	"refreshReferencePacks":             referencePackReason,
	"getReferencePackValidationSummary": referencePackReason,
	"getReferencePackVersion":           referencePackReason,
	"activateReferencePackVersion":      referencePackReason,
	"disableReferencePackVersion":       referencePackReason,
	"removeReferencePackVersion":        referencePackReason,
	"reverifyReferencePackVersion":      referencePackReason,

	"listViewSchemas": viewSchemaReason,
	"getViewSchema":   viewSchemaReason,

	"listVisibleIncidents": incidentDirectoryReason,
	"createIncident":       incidentCreationReason,
	"importIncidentBundle": incidentCreationReason,
}

// incidentScopedByContract reports why the contract itself places an
// operation inside an incident: its path addresses the incident or a record,
// or it declares the incident's not-found response. Such an operation can never
// be classified as not incident scoped.
func incidentScopedByContract(operation httpapi.ContractOperation, recordTypes []string) string {
	switch {
	case slices.Contains(pathParameters(operation), "incident_id"):
		return "its path addresses an incident"
	case len(recordPathParameters(operation, recordTypes)) != 0:
		return "its path addresses a record"
	case declaresErrorCode(operation, http.StatusNotFound, "incident_not_found"):
		return "it declares 404 incident_not_found"
	}
	return ""
}

// concealmentPair is one live resource's request, sent by a non-member, and
// the same request with the addressed identifier replaced, sent by a member.
type concealmentPair struct {
	live    conformanceRequest
	missing conformanceRequest
}

// addressed pairs live with a copy whose path parameter name addresses a
// resource that does not exist.
func addressed(live conformanceRequest, name string) concealmentPair {
	return addressedAs(live, name, uuid.NewString())
}

func addressedAs(live conformanceRequest, name string, value string) concealmentPair {
	missing := live
	missing.params = make(map[string]string, len(live.params))
	for key, existing := range live.params {
		missing.params[key] = existing
	}
	missing.params[name] = value
	return concealmentPair{live: live, missing: missing}
}

// bodyAddressed pairs live with a copy whose body member key addresses a
// resource that does not exist.
func bodyAddressed(live conformanceRequest, key string) concealmentPair {
	missing := live
	body := map[string]any{}
	for name, value := range live.body.(map[string]any) {
		body[name] = value
	}
	body[key] = uuid.NewString()
	missing.body = body
	return concealmentPair{live: live, missing: missing}
}

func (e *conformanceEnv) incidentRequest(label string, body any) conformanceRequest {
	return conformanceRequest{label: label, params: map[string]string{"incident_id": e.incidentID.String()}, body: body}
}

func incidentAddressed(request conformanceRequest) []concealmentPair {
	return []concealmentPair{addressed(request, "incident_id")}
}

// concealmentEnv adds the non-member and the shared, lazily prepared fixtures.
type concealmentEnv struct {
	*conformanceEnv
	nonMember         authcookietest.AuthCookies
	candidateID       uuid.UUID
	memberID          uuid.UUID
	membershipVersion any
	report            *reportFixture
	bundleID          string
	committedID       string
	discovered        [2]string
	evidenceHandle    *handleFixture
}

type reportFixture struct {
	snapshotJobID      string
	snapshotID         string
	compositionID      string
	compositionVersion any
	releaseID          string
	redactionDigest    string
}

type handleFixture struct {
	evidenceID uuid.UUID
	token      string
}

func randomHex(t testing.TB, bytes int) string {
	t.Helper()
	value := make([]byte, bytes)
	if _, err := rand.Read(value); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(value)
}

// succeededJob admits a Job and requires it to succeed.
func (e *concealmentEnv) succeededJob(path string, body map[string]any, label string) (string, map[string]any) {
	e.t.Helper()
	job := e.require(http.MethodPost, path, body, http.StatusAccepted)
	jobID := job["job_id"].(string)
	terminal := httptestx.PollJob(e.t, e.harness.Server.HTTP.URL, jobID, httptestx.PollOptions{
		Timeout: 2 * time.Minute, Interval: 50 * time.Millisecond, Request: []func(*http.Request){httptestx.WithCookies(e.login.Session)},
	}, "succeeded", "failed", "canceled")
	if terminal["status"] != "succeeded" {
		e.t.Fatalf("%s Job did not succeed: %#v", label, terminal)
	}
	return jobID, terminal
}

// jobResource admits a Job, requires it to succeed and returns the resource
// of kind its summary references.
func (e *concealmentEnv) jobResource(path string, body map[string]any, kind string) (string, string) {
	e.t.Helper()
	jobID, terminal := e.succeededJob(path, body, kind)
	for _, raw := range terminal["result_summary"].(map[string]any)["resource_refs"].([]any) {
		if ref := raw.(map[string]any); ref["kind"] == kind {
			return jobID, ref["id"].(string)
		}
	}
	e.t.Fatalf("%s Job references no %s", kind, kind)
	return "", ""
}

func (e *concealmentEnv) reportFixtures() *reportFixture {
	e.t.Helper()
	if e.report != nil {
		return e.report
	}
	fixture := &reportFixture{}
	fixture.snapshotJobID, fixture.snapshotID = e.jobResource("/api/v1/snapshots", map[string]any{
		"incident_id": e.incidentID.String(), "client_txn_id": e.txn("snapshot"),
	}, "snapshot")
	composition := e.require(http.MethodPost, e.incidentPath()+"/report-compositions", map[string]any{
		"client_txn_id": e.txn("composition"), "template_id": reporting.DefaultTemplateID, "template_version": reporting.DefaultTemplateVersion,
	}, http.StatusCreated)
	fixture.compositionID = composition["composition_id"].(string)
	version := e.require(http.MethodPost, e.incidentPath()+"/report-compositions/"+fixture.compositionID+"/versions", map[string]any{
		"client_txn_id": e.txn("freeze"), "base_draft_version": composition["draft_version"],
	}, http.StatusCreated)
	fixture.compositionVersion = version["composition_version"]
	_, fixture.releaseID = e.jobResource("/api/v1/releases", map[string]any{
		"client_txn_id": e.txn("release"), "snapshot_id": fixture.snapshotID, "template_id": reporting.DefaultTemplateID,
		"template_version": reporting.DefaultTemplateVersion, "output_kind": reporting.OutputKindSlidev, "release_scope": reporting.ReleaseScopeInternalReview,
	}, "release")
	var capture []byte
	var digest string
	if err := e.harness.DB.QueryRow(`SELECT p.request_json->'source_capture',s.source_capture_sha256 FROM reporting_snapshots s JOIN reporting_job_payloads p ON p.job_id=s.create_job_id WHERE s.snapshot_id::text=$1`, fixture.snapshotID).Scan(&capture, &digest); err != nil {
		e.t.Fatalf("read snapshot capture: %v", err)
	}
	model, err := reporting.MaterializeSnapshotCapture(e.t.Context(), capture, digest)
	if err != nil {
		e.t.Fatalf("materialize snapshot capture: %v", err)
	}
	if _, fixture.redactionDigest, err = reporting.ResolveRedactionProfile(reporting.InternalRedactionProfileID, reporting.DefaultRedactionProfileVersion, nil, model.SnapshotContent); err != nil {
		e.t.Fatalf("resolve redaction profile: %v", err)
	}
	e.report = fixture
	return fixture
}

func (e *concealmentEnv) compositionRequest(label string, body any) conformanceRequest {
	request := e.incidentRequest(label, body)
	request.params["composition_id"] = e.reportFixtures().compositionID
	return request
}

func (e *concealmentEnv) draftVersion() any {
	e.t.Helper()
	return e.require(http.MethodGet, e.incidentPath()+"/report-compositions/"+e.reportFixtures().compositionID, nil, http.StatusOK)["draft_version"]
}

func (e *concealmentEnv) bundle() string {
	e.t.Helper()
	if e.bundleID == "" {
		_, e.bundleID = e.jobResource("/api/v1/incident-bundles/export", map[string]any{
			"incident_id": e.incidentID.String(), "client_txn_id": e.txn("bundle"),
		}, "incident_bundle")
	}
	return e.bundleID
}

// discoveredUnit is a discovered one-unit CSV session, shared by the unit
// routes; every request against it is concealed, so it stays discovered.
func (e *concealmentEnv) discoveredUnit() (string, string) {
	e.t.Helper()
	if e.discovered[0] == "" {
		e.discovered[0], e.discovered[1] = importtest.StartCSVSession(e.t, e.harness.Server.HTTP.URL, e.login, e.incidentID.String(), e.txn("csv-upload"), "summary\nconcealed row\n", "concealment.csv")
	}
	return e.discovered[0], e.discovered[1]
}

func (e *concealmentEnv) unitRequest(label string, body any) conformanceRequest {
	sessionID, unitID := e.discoveredUnit()
	return conformanceRequest{label: label, params: map[string]string{"import_session_id": sessionID, "import_unit_id": unitID}, body: body}
}

func (e *concealmentEnv) sessionRequest(label string, body any) conformanceRequest {
	sessionID, _ := e.discoveredUnit()
	return conformanceRequest{label: label, params: map[string]string{"import_session_id": sessionID}, body: body}
}

// committedSource applies one Timeline import and returns its committed
// source.
func (e *concealmentEnv) committedSource() string {
	e.t.Helper()
	if e.committedID != "" {
		return e.committedID
	}
	sessionID, _ := importtest.PrepareReadyTimelineImport(e.t, e.harness.Server.HTTP.URL, e.login, e.incidentID.String(), e.txn("committed"))
	e.succeededJob("/api/v1/import-sessions/"+sessionID+"/apply", map[string]any{"client_txn_id": e.txn("apply")}, "import apply")
	sources := e.require(http.MethodGet, e.incidentPath()+"/committed-sources", nil, http.StatusOK)["committed_sources"].([]any)
	if len(sources) == 0 {
		e.t.Fatal("applied import committed no source")
	}
	e.committedID = sources[0].(map[string]any)["committed_source_id"].(string)
	return e.committedID
}

// handles attaches a verified Blob to available Evidence and issues a preview
// handle as the member.
func (e *concealmentEnv) handles() *handleFixture {
	e.t.Helper()
	if e.evidenceHandle != nil {
		return e.evidenceHandle
	}
	evidenceID := e.evidence("Concealed handle Evidence")
	blobID := e.verifiedBlob("concealed-handle")
	attach := e.attach(evidenceID, blobID)
	e.require(http.MethodPost, "/api/v1/evidence-records/"+evidenceID.String()+"/attach-blob", attach.body, http.StatusOK)
	httptestx.RequireSuccessData(e.t, e.patch(evidenceID, "cartulary.view.evidence.v1", "evidence.lifecycle_state", "available", e.version(evidenceID)), http.StatusOK)
	issued := e.require(http.MethodPost, "/api/v1/evidence-records/"+evidenceID.String()+"/preview-handle", map[string]any{}, http.StatusOK)
	token, ok := strings.CutPrefix(issued["href"].(string), "/api/v1/evidence-handles/")
	if !ok || token == "" {
		e.t.Fatalf("preview handle href %q is not a handle route", issued["href"])
	}
	e.evidenceHandle = &handleFixture{evidenceID: evidenceID, token: token}
	return e.evidenceHandle
}

// importUpload builds createImportSession's multipart envelope.
func importUpload(t testing.TB, incidentID string, clientTxnID string) conformanceRequest {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, part := range []struct {
		header  textproto.MIMEHeader
		content string
	}{
		{textproto.MIMEHeader{"Content-Disposition": {`form-data; name="metadata"`}, "Content-Type": {"application/json"}},
			`{"client_txn_id":"` + clientTxnID + `","incident_id":"` + incidentID + `"}`},
		{textproto.MIMEHeader{"Content-Disposition": {`form-data; name="file"; filename="concealed.csv"`}, "Content-Type": {"text/csv"}},
			"summary\nconcealed upload\n"},
	} {
		partWriter, err := writer.CreatePart(part.header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(partWriter, part.content); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return conformanceRequest{label: "CSV upload", raw: body.Bytes(), headers: map[string]string{"Content-Type": writer.FormDataContentType()}}
}

func recordRequest(label string, param string, recordID uuid.UUID, body any) conformanceRequest {
	return conformanceRequest{label: label, params: map[string]string{param: recordID.String()}, body: body}
}

func one(pair concealmentPair) []concealmentPair { return []concealmentPair{pair} }

// concealmentCases holds one case per incident-scoped operation. Each pairs a
// live resource's request, valid but for the caller's visibility, with the
// same request addressing a resource that does not exist.
var concealmentCases = map[string]func(*concealmentEnv) []concealmentPair{
	// Incident lifecycle, metadata and memberships.
	"getIncident": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.incidentRequest("incident", nil))
	},
	"patchIncident": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.incidentRequest("current phase", func(t *testing.T) any {
			return map[string]any{"base_incident_version": e.incidentVersion(t), "current_phase": "Concealed phase"}
		}))
	},
	"closeIncident": func(e *concealmentEnv) []concealmentPair {
		txn := e.txn("close")
		return incidentAddressed(e.incidentRequest("close", func(t *testing.T) any {
			return map[string]any{"base_incident_version": e.incidentVersion(t), "client_txn_id": txn, "reason": "Concealed closure"}
		}))
	},
	"reopenIncident": func(e *concealmentEnv) []concealmentPair {
		txn := e.txn("reopen")
		return incidentAddressed(e.incidentRequest("reopen", func(t *testing.T) any {
			return map[string]any{"base_incident_version": e.incidentVersion(t), "client_txn_id": txn, "reason": "Concealed reopening"}
		}))
	},
	"listIncidentMemberships": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.incidentRequest("memberships", nil))
	},
	"listIncidentMembershipAuditEvents": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.incidentRequest("membership audit", nil))
	},
	"createIncidentMembership": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.incidentRequest("viewer membership", map[string]any{
			"client_txn_id": e.txn("membership"), "user_id": e.candidateID.String(), "role": "viewer",
		}))
	},
	"patchIncidentMembership": func(e *concealmentEnv) []concealmentPair {
		request := e.incidentRequest("editor role", map[string]any{"base_membership_version": e.membershipVersion, "role": "editor"})
		request.params["user_id"] = e.memberID.String()
		return incidentAddressed(request)
	},
	"deleteIncidentMembership": func(e *concealmentEnv) []concealmentPair {
		request := e.incidentRequest("viewer membership", map[string]any{"base_membership_version": e.membershipVersion})
		request.params["user_id"] = e.memberID.String()
		return incidentAddressed(request)
	},
	"getTimelineTimeConversionProfile": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.incidentRequest("profile", nil))
	},
	"putTimelineTimeConversionProfile": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.incidentRequest("local offset", map[string]any{
			"base_profile_version": 1, "enabled": true, "local_offset_minutes": -300, "local_label": "UTC-05",
		}))
	},

	// Workbook views, preferences, saved views and startup.
	"createViewRow": func(e *concealmentEnv) []concealmentPair {
		request := e.incidentRequest("Note row", map[string]any{"client_txn_id": e.txn("create"), "note.title": "Concealed note", "note.body": "Body"})
		request.params["view_schema_id"] = "cartulary.view.notes.v1"
		return incidentAddressed(request)
	},
	"applyWorkbookBulkMutation": func(e *concealmentEnv) []concealmentPair {
		recordID := e.timeline("Concealed bulk target")
		request := e.incidentRequest("Timeline fill-down", map[string]any{
			"view_schema_id": conformanceTimelineView, "client_txn_id": e.txn("bulk"), "kind": "fill_down_v1",
			"field_key": "timeline.raw_activity_text", "value": "Concealed fill",
			"targets": []map[string]any{{"record_id": recordID.String(), "base_row_version": 1}},
		})
		request.params["view_schema_id"] = conformanceTimelineView
		return incidentAddressed(request)
	},
	"pasteWorkbookClipboard": func(e *concealmentEnv) []concealmentPair {
		request := e.incidentRequest("Timeline create paste", map[string]any{
			"view_schema_id": conformanceTimelineView, "client_txn_id": e.txn("paste"),
			"clipboard_text": "concealed paste", "format": "csv", "header_mode": "none",
			"start_field_key": "timeline.raw_activity_text", "columns": []string{"timeline.raw_activity_text"},
			"targets": []map[string]any{{"kind": "create"}},
		})
		request.params["view_schema_id"] = conformanceTimelineView
		return incidentAddressed(request)
	},
	"queryWorkbookView": func(e *concealmentEnv) []concealmentPair {
		request := e.incidentRequest("Timeline query", map[string]any{})
		request.params["view_schema_id"] = conformanceTimelineView
		return incidentAddressed(request)
	},
	"locateWorkbookViewRecord": func(e *concealmentEnv) []concealmentPair {
		request := e.incidentRequest("Timeline locate", map[string]any{"record_id": e.timeline("Concealed locate target").String()})
		request.params["view_schema_id"] = conformanceTimelineView
		return incidentAddressed(request)
	},
	"getIncidentDefaultWorkbookPreferences": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.incidentRequest("default preferences", nil))
	},
	"putIncidentDefaultWorkbookPreferences": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.incidentRequest("default sheet", map[string]any{
			"default_sheet_ref": map[string]any{"kind": "view_schema", "id": conformanceTimelineView},
		}))
	},
	"getCurrentUserWorkbookPreferences": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.incidentRequest("own preferences", nil))
	},
	"putCurrentUserWorkbookPreferences": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.incidentRequest("home sheet", map[string]any{
			"home_sheet_ref": map[string]any{"kind": "view_schema", "id": conformanceTimelineView},
		}))
	},
	"getIncidentWorkbookStartup": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.incidentRequest("startup", nil))
	},
	"listIncidentSavedViews": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.incidentRequest("saved views", nil))
	},
	"createIncidentSavedView": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.incidentRequest("shared view", map[string]any{
			"view_schema_id": conformanceTimelineView, "display_name": "Concealed view", "query_json": map[string]any{},
		}))
	},
	"getIncidentSavedView": func(e *concealmentEnv) []concealmentPair {
		request, _ := e.savedView("read")
		request.body = nil
		return incidentAddressed(request)
	},
	"patchIncidentSavedView": func(e *concealmentEnv) []concealmentPair {
		request, version := e.savedView("rename")
		request.body = map[string]any{"base_saved_view_version": version, "display_name": "Concealed rename"}
		return incidentAddressed(request)
	},
	"deleteIncidentSavedView": func(e *concealmentEnv) []concealmentPair {
		request, _ := e.savedView("delete")
		request.body = nil
		return incidentAddressed(request)
	},

	// Incident-scoped reads over records, candidates, sources and Evidence.
	"listEntityCandidates": func(e *concealmentEnv) []concealmentPair {
		e.host("concealed-candidate")
		request := e.incidentRequest("Host candidates", nil)
		request.query = "entity_type=host"
		return incidentAddressed(request)
	},
	"queryEvidenceMetadata": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.incidentRequest("Evidence metadata", map[string]any{
			"record_ids": []string{e.evidence("Concealed metadata Evidence").String()},
		}))
	},
	"listCommittedImportSources": func(e *concealmentEnv) []concealmentPair {
		e.committedSource()
		return incidentAddressed(e.incidentRequest("committed sources", nil))
	},
	"getCommittedImportSource": func(e *concealmentEnv) []concealmentPair {
		request := e.incidentRequest("committed source", nil)
		request.params["committed_source_id"] = e.committedSource()
		return incidentAddressed(request)
	},
	"getCommittedImportSourceContent": func(e *concealmentEnv) []concealmentPair {
		request := e.incidentRequest("committed source content", nil)
		request.params["committed_source_id"] = e.committedSource()
		return incidentAddressed(request)
	},

	// Report compositions and retained report artifacts.
	"listReportCompositions": func(e *concealmentEnv) []concealmentPair {
		e.reportFixtures()
		return incidentAddressed(e.incidentRequest("compositions", nil))
	},
	"createReportComposition": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.incidentRequest("default template", map[string]any{
			"client_txn_id": e.txn("composition"), "template_id": reporting.DefaultTemplateID, "template_version": reporting.DefaultTemplateVersion,
		}))
	},
	"getReportComposition": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.compositionRequest("composition", nil))
	},
	"updateReportCompositionDraft": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.compositionRequest("draft texts", map[string]any{
			"client_txn_id": e.txn("draft"), "base_draft_version": e.draftVersion(), "authored_texts": []any{},
		}))
	},
	"previewReportComposition": func(e *concealmentEnv) []concealmentPair {
		fixture := e.reportFixtures()
		return incidentAddressed(e.compositionRequest("draft preview", map[string]any{
			"client_txn_id": e.txn("preview"), "source_kind": "draft", "snapshot_id": fixture.snapshotID,
			"derivation_version": reporting.DerivationVersion, "template_id": reporting.DefaultTemplateID, "template_version": reporting.DefaultTemplateVersion,
			"redaction_profile_id": reporting.InternalRedactionProfileID, "redaction_profile_version": reporting.DefaultRedactionProfileVersion,
			"redaction_profile_sha256": fixture.redactionDigest, "render_environment_profile_id": "cartulary.reporting.local.v1",
			"output_kind": reporting.OutputKindSlidev, "output_options": map[string]any{}, "recipient_partition_refs": []any{}, "graph_projection_refs": []any{},
		}))
	},
	"validateReportComposition": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.compositionRequest("draft validation", map[string]any{"source_kind": "draft"}))
	},
	"freezeReportCompositionVersion": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.compositionRequest("freeze", map[string]any{
			"client_txn_id": e.txn("freeze"), "base_draft_version": e.draftVersion(),
		}))
	},
	"getReportCompositionVersion": func(e *concealmentEnv) []concealmentPair {
		request := e.compositionRequest("frozen version", nil)
		version, err := json.Marshal(e.reportFixtures().compositionVersion)
		if err != nil {
			e.t.Fatal(err)
		}
		request.params["composition_version"] = strings.Trim(string(version), `"`)
		return incidentAddressed(request)
	},
	"retireReportComposition": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.compositionRequest("retire", map[string]any{
			"client_txn_id": e.txn("retire"), "base_draft_version": e.draftVersion(),
		}))
	},
	"getRetainedReportArtifacts": func(e *concealmentEnv) []concealmentPair {
		return incidentAddressed(e.incidentRequest("retained artifacts", nil))
	},
	// Retained artifacts exist only on an imported incident. Visibility of the
	// live incident precedes the artifact's existence (REQ-01-704).
	"getRetainedReportArtifactContent": func(e *concealmentEnv) []concealmentPair {
		request := e.incidentRequest("retained artifact content", nil)
		request.params["artifact_id"] = randomHex(e.t, 32)
		return incidentAddressed(request)
	},

	// Record mutations and reads.
	"patchRecord": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(e.notePatch(e.note("Concealed patch target")), "record_id"))
	},
	"deleteRecord": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(recordRequest("live Note", "record_id", e.note("Concealed delete target"), map[string]any{
			"base_row_version": 1, "client_txn_id": e.txn("delete"), "reason": "Concealed deletion",
		}), "record_id"))
	},
	"restoreRecord": func(e *concealmentEnv) []concealmentPair {
		noteID := e.note("Concealed restore target")
		e.softDelete(noteID)
		return one(addressed(recordRequest("deleted Note", "record_id", noteID, map[string]any{
			"base_row_version": e.version(noteID), "client_txn_id": e.txn("restore"),
		}), "record_id"))
	},
	"rollbackRecord": func(e *concealmentEnv) []concealmentPair {
		noteID := e.note("Concealed rollback target")
		httptestx.RequireSuccessData(e.t, e.patch(noteID, "cartulary.view.notes.v1", "note.title", "Concealed rollback revised", 1), http.StatusOK)
		return one(addressed(recordRequest("Note row restore", "record_id", noteID, map[string]any{
			"base_row_version": 2, "client_txn_id": e.txn("rollback"), "target": map[string]any{"kind": "row_restore", "restore_to_revision_no": 1},
		}), "record_id"))
	},
	"supersedeRecord": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(e.timelineAction(e.timeline("Concealed supersede target"), "supersede"), "record_id"))
	},
	"markTimelineRecordReviewed": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(e.timelineAction(e.timeline("Concealed review target"), "mark-reviewed"), "record_id"))
	},
	"createRecordLinkedNote": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(e.linkedNote(e.host("concealed-linked-source")), "record_id"))
	},
	"mutateNoteAssociations": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(e.noteAssociation(e.note("Concealed association"), e.note("Concealed counterpart")), "note_record_id"))
	},
	"listNoteAssociations": func(e *concealmentEnv) []concealmentPair {
		request := recordRequest("related Notes", "note_record_id", e.note("Concealed association list"), nil)
		request.query = "kind=related_note"
		return one(addressed(request, "note_record_id"))
	},
	"resolveRecordSameFieldConflict": func(e *concealmentEnv) []concealmentPair {
		recordID, token := e.timelineConflict("Concealed conflict")
		return one(addressed(e.conflictResolutions(recordID, token)[0], "record_id"))
	},
	"getRecordHistory": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(recordRequest("Note history", "record_id", e.note("Concealed history"), nil), "record_id"))
	},
	"mergeEntityRecord": func(e *concealmentEnv) []concealmentPair {
		survivorID, loserID := e.host("concealed-merge-survivor"), e.host("concealed-merge-loser")
		return one(addressed(recordRequest("Host merge", "survivor_record_id", survivorID, map[string]any{
			"loser_record_id": loserID.String(), "survivor_base_row_version": 1, "loser_base_row_version": 1,
			"client_txn_id": e.txn("merge"), "reason": "Concealed merge",
		}), "survivor_record_id"))
	},
	"resolveEntityMention": func(e *concealmentEnv) []concealmentPair {
		_, mentionID, hostID := e.unresolvedMention("concealed")
		return one(addressed(e.mentionResolution(mentionID, hostID), "entity_mention_id"))
	},

	// Indicators and their observations and state intervals.
	"createManualIndicatorObservation": func(e *concealmentEnv) []concealmentPair {
		text := "concealed-manual.conformance.test"
		return one(addressed(e.manualObservation(e.timeline(text), "timeline.raw_activity_text", text), "source_record_id"))
	},
	"listSourceRecordIndicatorObservations": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(recordRequest("source observations", "source_record_id", e.timeline("Concealed observation source"), nil), "source_record_id"))
	},
	"appendIndicatorStateInterval": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(e.stateInterval(e.indicator("concealed-interval.conformance.test"), e.timeline("Concealed interval support")), "indicator_id"))
	},
	"listIndicatorObservations": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(recordRequest("Indicator observations", "indicator_id", e.indicator("concealed-observations.conformance.test"), nil), "indicator_id"))
	},
	"listIndicatorStateIntervals": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(recordRequest("Indicator intervals", "indicator_id", e.indicator("concealed-intervals.conformance.test"), nil), "indicator_id"))
	},
	"dismissIndicatorObservation": func(e *concealmentEnv) []concealmentPair {
		observationID := e.observation(e.timeline("concealed-dismiss.conformance.test"), "timeline.raw_activity_text", "concealed-dismiss.conformance.test")
		return one(addressed(recordRequest("unresolved observation", "observation_id", observationID, map[string]any{
			"client_txn_id": e.txn("dismiss"), "base_row_version": 1,
		}), "observation_id"))
	},
	"resolveIndicatorObservation": func(e *concealmentEnv) []concealmentPair {
		observationID := e.observation(e.timeline("concealed-resolve.conformance.test"), "timeline.raw_activity_text", "concealed-resolve.conformance.test")
		indicatorID := e.indicator("concealed-resolve-target.conformance.test")
		return one(addressed(recordRequest("unresolved observation", "observation_id", observationID, map[string]any{
			"client_txn_id": e.txn("resolve"), "base_row_version": 1, "resolved_indicator_record_id": indicatorID.String(),
		}), "observation_id"))
	},
	"restoreIndicatorObservation": func(e *concealmentEnv) []concealmentPair {
		observationID := e.observation(e.timeline("concealed-restore.conformance.test"), "timeline.raw_activity_text", "concealed-restore.conformance.test")
		e.require(http.MethodPost, "/api/v1/indicator-observations/"+observationID.String()+"/dismiss",
			map[string]any{"client_txn_id": e.txn("dismiss-before-restore"), "base_row_version": 1}, http.StatusOK)
		return one(addressed(recordRequest("dismissed observation", "observation_id", observationID, map[string]any{
			"client_txn_id": e.txn("restore"), "base_row_version": 2,
		}), "observation_id"))
	},

	// Evidence records, handles, Blobs and uploads.
	"attachBlobToEvidenceRecord": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(e.attach(e.evidence("Concealed attach target"), e.verifiedBlob("concealed-attach")), "record_id"))
	},
	"issueEvidencePreviewHandle": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(recordRequest("preview handle", "record_id", e.handles().evidenceID, map[string]any{}), "record_id"))
	},
	"issueEvidenceDownloadHandle": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(recordRequest("download handle", "record_id", e.handles().evidenceID, map[string]any{}), "record_id"))
	},
	"redeemEvidenceHandle": func(e *concealmentEnv) []concealmentPair {
		token := e.handles().token
		return one(addressedAs(conformanceRequest{label: "preview handle", params: map[string]string{"handle_token": token}}, "handle_token", randomHex(e.t, len(token)/2+1)[:len(token)]))
	},
	"createObjectBlobSlot": func(e *concealmentEnv) []concealmentPair {
		return one(bodyAddressed(conformanceRequest{label: "text Blob slot", body: e.blobCreateBody("concealed-slot", []byte("concealed slot payload"))}, "incident_id"))
	},
	"verifyObjectBlob": func(e *concealmentEnv) []concealmentPair {
		slot := e.uploadedBlob("concealed-verify")
		return one(addressed(conformanceRequest{label: "uploaded Blob", params: map[string]string{"object_blob_id": slot.blobID}, body: map[string]any{"client_txn_id": e.txn("verify")}}, "object_blob_id"))
	},
	"uploadObjectBlobContent": func(e *concealmentEnv) []concealmentPair {
		return one(addressedAs(e.blobSlot("concealed-upload").upload(), "upload_token", "upl_"+randomHex(e.t, 32)))
	},

	// Import sessions and units.
	"createImportSession": func(e *concealmentEnv) []concealmentPair {
		txn := e.txn("upload")
		return one(concealmentPair{live: importUpload(e.t, e.incidentID.String(), txn), missing: importUpload(e.t, uuid.NewString(), txn)})
	},
	"getImportSession": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(e.sessionRequest("discovered session", nil), "import_session_id"))
	},
	"listImportUnits": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(e.sessionRequest("units", nil), "import_session_id"))
	},
	"getImportUnit": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(e.unitRequest("unit", nil), "import_session_id"))
	},
	"getImportUnitPreview": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(e.unitRequest("unit preview", nil), "import_session_id"))
	},
	"putImportUnitMapping": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(e.unitRequest("Timeline mapping", map[string]any{
			"client_txn_id": e.txn("mapping"), "target_view_schema_id": conformanceTimelineView,
			"header_row_ref": 1, "data_start_row_ref": 2, "unknown_column_policy": "preserve_raw_capture",
			"source_columns": []map[string]any{{
				"source_column_ordinal": 1, "source_header_text": "summary", "field_key": "timeline.activity_synopsis_text",
				"entity_binding_mode": nil, "transform_id": nil, "transform_options": map[string]any{}, "empty_value_policy": "omit_field",
			}},
		}), "import_session_id"))
	},
	"selectImportUnit": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(e.unitRequest("selection", map[string]any{"client_txn_id": e.txn("unit-selection")}), "import_session_id"))
	},
	"skipImportUnit": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(e.unitRequest("skip", map[string]any{"client_txn_id": e.txn("skip"), "reason": "Concealed skip"}), "import_session_id"))
	},
	"inspectImportUnitExtensionSource": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(e.unitRequest("Network Flow inspection", map[string]any{
			"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity",
		}), "import_session_id"))
	},
	"previewImportUnitExtensionMapping": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(e.unitRequest("Network Flow mapping preview", map[string]any{
			"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity",
			"owner_mapping_schema_id": "cartulary.network_flow.mapping_candidate.v2", "owner_mapping": map[string]any{},
		}), "import_session_id"))
	},
	"applyImportSession": func(e *concealmentEnv) []concealmentPair {
		sessionID, _ := importtest.PrepareReadyTimelineImport(e.t, e.harness.Server.HTTP.URL, e.login, e.incidentID.String(), e.txn("ready-import"))
		return one(addressed(conformanceRequest{label: "ready Timeline import", params: map[string]string{"import_session_id": sessionID}, body: map[string]any{"client_txn_id": e.txn("apply")}}, "import_session_id"))
	},
	"createImportUnitRegion": func(e *concealmentEnv) []concealmentPair {
		sessionID, unitIDs := importtest.StartXLSXSession(e.t, e.harness.Server.HTTP.URL, e.login, e.incidentID.String(), e.txn("xlsx-upload"), "concealed-regions.xlsx")
		return one(addressed(conformanceRequest{
			label: "XLSX operator region", params: map[string]string{"import_session_id": sessionID, "base_unit_id": unitIDs[0]},
			body: map[string]any{"client_txn_id": e.txn("region"), "source_rect": map[string]any{"start_row": 1, "start_column": 1, "end_row": 2, "end_column": 2}},
		}, "import_session_id"))
	},

	// Jobs, Reporting snapshots and releases, and incident bundles.
	"getJob": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(conformanceRequest{label: "snapshot Job", params: map[string]string{"job_id": e.reportFixtures().snapshotJobID}}, "job_id"))
	},
	"cancelJob": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(conformanceRequest{
			label: "snapshot Job", params: map[string]string{"job_id": e.reportFixtures().snapshotJobID},
			body: map[string]any{"client_txn_id": e.txn("cancel"), "reason": "Concealed cancellation"},
		}, "job_id"))
	},
	"createSnapshot": func(e *concealmentEnv) []concealmentPair {
		return one(bodyAddressed(conformanceRequest{label: "incident snapshot", body: map[string]any{
			"incident_id": e.incidentID.String(), "client_txn_id": e.txn("snapshot"),
		}}, "incident_id"))
	},
	"getSnapshot": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(conformanceRequest{label: "snapshot", params: map[string]string{"snapshot_id": e.reportFixtures().snapshotID}}, "snapshot_id"))
	},
	"createRelease": func(e *concealmentEnv) []concealmentPair {
		return one(bodyAddressed(conformanceRequest{label: "snapshot release", body: map[string]any{
			"client_txn_id": e.txn("release"), "snapshot_id": e.reportFixtures().snapshotID, "template_id": reporting.DefaultTemplateID,
			"template_version": reporting.DefaultTemplateVersion, "output_kind": reporting.OutputKindSlidev, "release_scope": reporting.ReleaseScopeInternalReview,
		}}, "snapshot_id"))
	},
	"getRelease": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(conformanceRequest{label: "release", params: map[string]string{"release_id": e.reportFixtures().releaseID}}, "release_id"))
	},
	"approveRelease": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(conformanceRequest{label: "approval", params: map[string]string{"release_id": e.reportFixtures().releaseID}, body: map[string]any{"client_txn_id": e.txn("approve"), "reason": "Concealed approval"}}, "release_id"))
	},
	"invalidateRelease": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(conformanceRequest{label: "invalidation", params: map[string]string{"release_id": e.reportFixtures().releaseID}, body: map[string]any{"client_txn_id": e.txn("invalidate"), "reason": "Concealed invalidation"}}, "release_id"))
	},
	"publishRelease": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(conformanceRequest{label: "publication", params: map[string]string{"release_id": e.reportFixtures().releaseID}, body: map[string]any{"client_txn_id": e.txn("publish"), "reason": "Concealed publication"}}, "release_id"))
	},
	"exportIncidentBundle": func(e *concealmentEnv) []concealmentPair {
		return one(bodyAddressed(conformanceRequest{label: "incident export", body: map[string]any{
			"incident_id": e.incidentID.String(), "client_txn_id": e.txn("export"),
		}}, "incident_id"))
	},
	"getIncidentBundleDescriptor": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(conformanceRequest{label: "bundle descriptor", params: map[string]string{"bundle_id": e.bundle()}}, "bundle_id"))
	},
	"getIncidentBundleContent": func(e *concealmentEnv) []concealmentPair {
		return one(addressed(conformanceRequest{label: "bundle content", params: map[string]string{"bundle_id": e.bundle()}}, "bundle_id"))
	},
}

// savedView creates a shared saved view and returns its addressed request and
// version.
func (e *concealmentEnv) savedView(label string) (conformanceRequest, any) {
	e.t.Helper()
	created := e.require(http.MethodPost, e.incidentPath()+"/saved-views", map[string]any{
		"view_schema_id": conformanceTimelineView, "display_name": "Concealed " + label, "query_json": map[string]any{},
	}, http.StatusCreated)
	request := e.incidentRequest("saved view "+label, nil)
	request.params["saved_view_id"] = created["saved_view_id"].(string)
	return request, created["saved_view_version"]
}

// concealedResponse is an error response without its request identifier.
type concealedResponse struct {
	Status  int
	Code    any
	Details any
}

func readConcealedResponse(t *testing.T, caller string, response *http.Response) concealedResponse {
	t.Helper()
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, concealedBodyLimit))
	if err != nil {
		t.Fatalf("read %s response: %v", caller, err)
	}
	var envelope map[string]any
	if err := json.Unmarshal(body, &envelope); err != nil || response.StatusCode < 400 {
		excerpt := string(body)
		if len(excerpt) > 512 {
			excerpt = excerpt[:512]
		}
		t.Fatalf("%s response was not an error envelope: status %d body=%s", caller, response.StatusCode, excerpt)
	}
	errorObject, ok := envelope["error"].(map[string]any)
	if !ok {
		t.Fatalf("%s response has no error object: status %d body=%s", caller, response.StatusCode, body)
	}
	return concealedResponse{Status: response.StatusCode, Code: errorObject["code"], Details: errorObject["details"]}
}

// TestNonMemberConcealmentConformance verifies Core 04 AC-616's concealment
// family (REQ-01-704). Every catalog operation is classified as incident
// scoped, with a case, or not incident scoped, with a reason; the contract
// itself forces the classification of any operation whose path addresses an
// incident or a record or which declares incident_not_found. For each
// incident-scoped operation an authenticated deployment user without
// membership sends a valid request for a live resource and receives the
// response a member receives for the same request addressing a resource that
// does not exist, identical in status, code and details; the request commits
// no effect of any AC-616 class.
func TestNonMemberConcealmentConformance(t *testing.T) {
	catalog := catalogOperationsByID(t)
	classified := sortedKeys(concealmentCases)
	for id, reason := range notIncidentScoped {
		if _, scoped := concealmentCases[id]; scoped {
			t.Errorf("%s is classified both incident scoped and not incident scoped (%s)", id, reason)
		}
		classified = append(classified, id)
	}
	recordTypes := publicRecordTypes()
	for id, reason := range notIncidentScoped {
		if operation, ok := catalog[id]; ok {
			if derived := incidentScopedByContract(operation, recordTypes); derived != "" {
				t.Errorf("%s is classified not incident scoped (%s), but %s", id, reason, derived)
			}
		}
	}
	if t.Failed() {
		t.FailNow()
	}
	requireCaseKeySet(t, "non-member concealment", classified, sortedKeys(catalog))

	base := startConformanceEnv(t, "nonmember-concealment-conformance", "IR-CONCEALMENT-CONFORMANCE")
	e := &concealmentEnv{conformanceEnv: base}
	candidate := authflowtest.SeedLocalUserRecord(t, base.harness.DB, "concealment-candidate@example.test", "Concealment Candidate", "ConcealmentCandidate1!", false, false, true)
	member := authflowtest.SeedLocalUserRecord(t, base.harness.DB, "concealment-member@example.test", "Concealment Member", "ConcealmentMember1!", false, false, true)
	outsider := authflowtest.SeedLocalUserRecord(t, base.harness.DB, "concealment-outsider@example.test", "Concealment Outsider", "ConcealmentOutsider1!", false, false, true)
	e.candidateID, e.memberID = candidate.ID, member.ID
	membership := e.require(http.MethodPost, e.incidentPath()+"/memberships", map[string]any{
		"client_txn_id": e.txn("member"), "user_id": member.ID.String(), "role": "viewer",
	}, http.StatusCreated)
	e.membershipVersion = membership["membership_version"]
	e.nonMember = authflowtest.LoginLocalUser(t, http.DefaultClient, base.harness.Server.HTTP.URL, outsider.Email, "ConcealmentOutsider1!", nil)

	operations := sortedKeys(concealmentCases)
	prepared := make(map[string][]concealmentPair, len(operations))
	for _, id := range operations {
		prepared[id] = concealmentCases[id](e)
		if len(prepared[id]) == 0 {
			t.Fatalf("%s case prepared no request", id)
		}
	}
	e.quiesce()

	for _, id := range operations {
		operation := e.operations[id]
		t.Run(id, func(t *testing.T) {
			for _, pair := range prepared[id] {
				t.Run(pair.live.label, func(t *testing.T) {
					before := workbookscenariotest.SnapshotIncidentEffects(t, e.harness, e.incidentID)
					concealed := readConcealedResponse(t, "non-member", e.send(t, operation, e.nonMember, pair.live))
					workbookscenariotest.RequireIncidentEffectsUnchanged(t, id+" non-member "+pair.live.label, before, workbookscenariotest.SnapshotIncidentEffects(t, e.harness, e.incidentID))
					missing := readConcealedResponse(t, "member missing-resource", e.send(t, operation, e.login, pair.missing))
					if !reflect.DeepEqual(concealed, missing) {
						t.Fatalf("%s %q: non-member response %+v differs from the missing resource's %+v", id, pair.live.label, concealed, missing)
					}
				})
			}
		})
	}
}
