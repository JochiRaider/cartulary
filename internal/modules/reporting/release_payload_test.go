package reporting

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
)

func TestReportingFrozenReleasePayload(t *testing.T) {
	jobID := uuid.MustParse("00000000-0000-0000-0000-000000000410")
	incidentID := uuid.MustParse("00000000-0000-0000-0000-000000000411")
	actorID := uuid.MustParse("00000000-0000-0000-0000-000000000412")
	admittedAt := time.Date(2026, 10, 14, 13, 0, 0, 0, time.UTC)
	snapshot, _, err := BuildSnapshotModel(IncidentMetadataSnapshot{ID: incidentID.String(), Title: "Frozen", Status: "active", Version: 1}, fixtureSnapshotID, admittedAt.Add(-time.Hour), "boundary", fixtureSnapshotSource(nil))
	if err != nil {
		t.Fatal(err)
	}
	releaseID := reportingResourceID(jobID, "release").String()
	stamp, _ := normalizedReportingTimestamp(admittedAt)
	model, digest, err := bindRenderModel(snapshot, RenderIdentity{ReleaseID: &releaseID, RenderAdmittedAt: stamp}, ReleaseScopeInternalDraft, nil)
	if err != nil {
		t.Fatal(err)
	}
	request, apiErr := DecodeCreateReleaseRequest(strings.NewReader(`{"snapshot_id":"` + fixtureSnapshotID + `","client_txn_id":"frozen-release","template_id":"cartulary.report.default","template_version":"` + DefaultTemplateVersion + `","output_kind":"slidev"}`))
	if apiErr != nil {
		t.Fatal(apiErr)
	}
	if request.RedactionProfileID != InternalRedactionProfileID || request.RedactionProfileVersion != DefaultRedactionProfileVersion {
		t.Fatal("internal omission was not normalized before freezing")
	}
	template, _ := ResolveTemplateContract(request.TemplateID, request.TemplateVersion)
	profile, profileSHA, err := ResolveRedactionProfile(request.RedactionProfileID, request.RedactionProfileVersion, request.RecipientPartitionRefs, model.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	payload := releaseCreateJobPayload{
		renderInput: renderInput{
			GraphSources: frozenGraphSources{SchemaID: frozenGraphSourcesSchema, Results: []resolvedGraphResult{}},
			SourceFacts:  snapshot.SourceFacts.Clone(), SourceFactsSHA256: fixtureSourceFactsDigest(snapshot.SnapshotContent),
			ActorUserID: actorID.String(), IncidentID: incidentID.String(), SnapshotID: snapshot.SnapshotID,
			SnapshotAt: snapshot.SnapshotAt, SourceChangeSetHighWatermark: snapshot.SourceChangeSetHighWatermark, DerivationVersion: snapshot.DerivationVersion,
			ExportModel: model, ExportModelSHA256: digest, TemplateID: request.TemplateID, TemplateVersion: request.TemplateVersion, TemplateContract: template,
			RedactionProfileID: request.RedactionProfileID, RedactionProfileVersion: request.RedactionProfileVersion,
			RedactionProfile: profile, RedactionProfileSHA256: profileSHA,
			OutputKind: request.OutputKind, OutputOptions: request.OutputOptions, ReleaseScope: request.ReleaseScope,
			RecipientPartitionRefs: []string{}, GraphProjectionRefs: request.GraphProjectionRefs, RenderAdmittedAt: admittedAt,
		},
		ClientTxnID: request.ClientTxnID, NormalizedRequest: request.Normalized,
	}
	var recipe renderAdmission
	var preparedInput renderInput
	t.Run("compact render admission", func(t *testing.T) {
		recipe, preparedInput = testCompactRenderAdmission(t, payload, jobID, incidentID, actorID, admittedAt)
	})
	payload.Request = request
	assertFrozenReleaseBindings(t, payload, jobID)
	raw, err := marshalRenderAdmission(recipe)
	if err != nil {
		t.Fatal(err)
	}
	admissionSHA := reportingObjectDigest(renderAdmissionSchema, raw)
	var formatted bytes.Buffer
	if err := json.Indent(&formatted, raw, "", "  "); err != nil {
		t.Fatal(err)
	}
	if _, err := decodeRenderAdmission(formatted.Bytes(), admissionSHA, jobID, incidentID, actorID, admittedAt); err != nil {
		t.Fatal("JSONB formatting changed admission", err)
	}
	if _, err := decodeRenderAdmission(raw, admissionSHA, uuid.New(), incidentID, actorID, admittedAt); err == nil {
		t.Fatal("recipe admitted under another Job")
	}
	retired, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decodeRenderAdmission(retired, reportingObjectDigest(renderAdmissionSchema, retired), jobID, incidentID, actorID, admittedAt); err == nil {
		t.Fatal("retired full-model Job input admitted")
	}
	if preparedInput.preparation == nil {
		t.Fatal("missing preparation seal")
	}
	snapshotBytes, captureDigest, err := freezeSnapshotCapture(IncidentMetadataSnapshot{ID: incidentID.String(), Title: "Frozen", Status: "active", Version: 1}, fixtureSnapshotID, admittedAt.Add(-time.Hour), "boundary", reference_data.SetBinding{}, fixtureSnapshotSource(nil))
	if err != nil {
		t.Fatal(err)
	}
	retained := SnapshotRecord{
		SnapshotID: request.SnapshotID, IncidentID: incidentID, SnapshotAt: snapshot.SnapshotAt,
		DerivationVersion: snapshot.DerivationVersion, SourceChangeSetHighWatermark: snapshot.SourceChangeSetHighWatermark,
		SourceCaptureJSON: snapshotBytes, SourceCaptureSHA256: captureDigest,
	}
	if err := validateRenderSnapshotBinding(payload.renderInput, retained); err != nil {
		t.Fatal("release did not retain its immutable snapshot", err)
	}
	// An altered job can be internally consistent after recomputing its own
	// digest. The separately retained snapshot still rejects substituted source.
	substituted := payload
	substituted.ExportModel.SourceChangeSetHighWatermark = "substituted"
	substituted.SourceChangeSetHighWatermark = "substituted"
	changedBytes, err := canonicalJSON(substituted.ExportModel)
	if err != nil {
		t.Fatal(err)
	}
	substituted.ExportModelSHA256 = reportingObjectDigest(substituted.ExportModel.SchemaID, changedBytes)
	if err := validateReleasePayloadIdentity(substituted, jobID); err != nil {
		t.Fatal("substitution fixture must have a self-consistent digest", err)
	}
	if err := validateRenderSnapshotBinding(substituted.renderInput, retained); err == nil {
		t.Fatal("recomputed job digest authorized substituted snapshot content")
	}
	for _, key := range []string{"snapshot", "incident", "time", "boundary", "derivation", "digest"} {
		t.Run("retained snapshot "+key, func(t *testing.T) {
			changed := retained
			switch key {
			case "snapshot":
				changed.SnapshotID = jobID
			case "incident":
				changed.IncidentID = actorID
			case "time":
				changed.SnapshotAt = admittedAt
			case "boundary":
				changed.SourceChangeSetHighWatermark = "changed"
			case "derivation":
				changed.DerivationVersion = "retired"
			case "digest":
				changed.SourceCaptureSHA256 = strings.Repeat("0", 64)
			}
			if _, err := decodeSnapshotRecordModel(changed); err == nil {
				t.Fatal("snapshot content disagreed with its retained row binding")
			}
		})
	}
	admission, err := jobs.NewExtensionJobAdmission(ProfileID, jobs.NewRouteIdempotencyKey("releases.create", actorID, payload.SnapshotID, payload.ClientTxnID), jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID}, payload.NormalizedRequest)
	if err != nil {
		t.Fatal(err)
	}
	retainedJob := jobs.RetainedExtensionJob{
		Resource: jobs.Resource{JobID: jobID.String(), Scope: jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID}, AuthPolicy: jobs.AuthPolicyIncidentMembership, SubmittedByUserID: &payload.ActorUserID, SubmittedAt: admittedAt},
		JobKind:  ReleaseCreateJobKind, HandlerName: JobWorkerKind, OwnerProfileID: ProfileID,
		IdempotencyIdentity: admission.IdempotencyIdentity, RouteKey: admission.IdempotencyRouteKey, ScopeKey: admission.IdempotencyScopeKey, RequestSHA256: admission.NormalizedRequestSHA256,
	}
	if err := validateRenderJobAdmission(recipe, jobID, retainedJob); err != nil {
		t.Fatal("release lost its Jobs admission", err)
	}
	changedProfile := recipe
	changedProfile.Profile.DefaultAction = RedactionActionSpec{Type: ActionDrop}
	changedProfile.ProfileSHA256, err = ValidateRedactionProfile(changedProfile.Profile)
	if err != nil {
		t.Fatal(err)
	}
	changedRaw, err := marshalRenderAdmission(changedProfile)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decodeRenderAdmission(changedRaw, admissionSHA, jobID, incidentID, actorID, admittedAt); err == nil {
		t.Fatal("rehashed profile replaced independent admission")
	}
	if _, err := decodeRenderAdmission(raw, "", jobID, incidentID, actorID, admittedAt); err == nil {
		t.Fatal("missing admission was inferred")
	}
	legacyJob := retainedJob
	legacyJob.JobKind = "snapshot_reporting.release_create_v2"
	if validateRenderJobAdmission(recipe, jobID, legacyJob) == nil {
		t.Fatal("retired v2 release Job admitted")
	}
	for _, state := range []string{jobs.StatusSucceeded, jobs.StatusFailed, jobs.StatusCanceled} {
		expired := retainedJob
		expired.Resource.Status = state
		expired.ExpiredAt = &admittedAt
		expired.IdempotencyIdentity = nil
		expired.RouteKey, expired.ScopeKey, expired.RequestSHA256 = "", "", ""
		if err := validateRenderJobAdmission(recipe, jobID, expired); err != nil {
			t.Fatal("expired native history rejected", err)
		}
		expired.Resource.Status = jobs.StatusRunning
		if validateRenderJobAdmission(recipe, jobID, expired) == nil {
			t.Fatal("unfinished Job admitted as expired release")
		}
	}
	changedRequest := request
	changedRequest.OutputOptions = bytes.ReplaceAll(changedRequest.OutputOptions, []byte(`"pdf":true`), []byte(`"pdf":false`))
	changedRequest.Normalized, err = normalizeReleaseRequest(changedRequest)
	if err != nil {
		t.Fatal(err)
	}
	changedRecipe := recipe
	changedRecipe.NormalizedRequest = changedRequest.Normalized
	changedRecipe.OutputOptions = changedRequest.OutputOptions
	if err := validateRenderAdmissionRequest(changedRecipe); err != nil {
		t.Fatal("changed request fixture is inconsistent", err)
	}
	if err := validateRenderJobAdmission(changedRecipe, jobID, retainedJob); err == nil {
		t.Fatal("rewritten request replaced independently admitted options")
	}
	for _, key := range []string{"job", "kind", "handler", "owner", "policy", "scope", "incident", "actor", "time", "route", "route scope", "request hash", "identity", "expired"} {
		t.Run("Jobs admission "+key, func(t *testing.T) {
			changed := retainedJob
			switch key {
			case "job":
				changed.Resource.JobID = actorID.String()
			case "kind":
				changed.JobKind = SnapshotCreateJobKind
			case "handler":
				changed.HandlerName = "retired"
			case "owner":
				changed.OwnerProfileID = "other"
			case "policy":
				changed.Resource.AuthPolicy = "other"
			case "scope":
				changed.Resource.Scope.Kind = jobs.ScopeKindDeployment
			case "incident":
				changed.Resource.Scope.IncidentID = &actorID
			case "actor":
				changed.Resource.SubmittedByUserID = nil
			case "time":
				changed.Resource.SubmittedAt = admittedAt.Add(time.Second)
			case "route":
				changed.RouteKey = "snapshots.create"
			case "route scope":
				changed.ScopeKey = actorID.String()
			case "request hash":
				changed.RequestSHA256 = strings.Repeat("0", 64)
			case "identity":
				changed.IdempotencyIdentity = json.RawMessage(`{}`)
			case "expired":
				changed.ExpiredAt = &admittedAt
			}
			if err := validateRenderJobAdmission(recipe, jobID, changed); err == nil {
				t.Fatal("retained job disagreed with release admission")
			}
		})
	}
}
