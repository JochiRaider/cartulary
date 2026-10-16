package reporting

import (
	"context"
	"errors"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reportcomposition"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func readRenderAdmissionTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID) (renderAdmission, string, error) {
	var raw []byte
	var digest, kind, profileSHA, templateSHA, graphSHA string
	var incident, actor uuid.UUID
	var at time.Time
	err := tx.QueryRow(ctx, `SELECT request_json,render_admission_sha256,job_kind,incident_id,actor_user_id,created_at,admitted_profile_sha256,admitted_template_manifest_sha256,admitted_graph_sources_sha256 FROM reporting_job_payloads WHERE job_id=$1 AND job_kind IN ('release_create','composition_preview')`, jobID).Scan(&raw, &digest, &kind, &incident, &actor, &at, &profileSHA, &templateSHA, &graphSHA)
	if errors.Is(err, pgx.ErrNoRows) {
		return renderAdmission{}, "", ErrNotFound
	}
	if err != nil {
		return renderAdmission{}, "", err
	}
	a, err := decodeRenderAdmission(raw, digest, jobID, incident, actor, at)
	if err != nil {
		return a, "", err
	}
	if a.ProfileSHA256 != profileSHA || a.TemplateSHA256 != templateSHA || a.GraphSourcesSHA256 != graphSHA ||
		(kind == "release_create") != (a.Identity.ReleaseID != nil) {
		return a, "", errors.New("reporting: render recipe differs from independent admission")
	}
	retained, err := jobs.ReadRetainedExtensionJob(ctx, tx, jobID)
	if err != nil {
		return a, "", err
	}
	if err := validateRenderJobAdmission(a, jobID, retained); err != nil {
		return a, "", err
	}
	d, err := readSnapshotDescriptorTx(ctx, tx, uuid.MustParse(a.SnapshotID))
	if err != nil {
		return a, "", err
	}
	if d.Incident.String() != a.IncidentID || !d.At.Equal(a.SnapshotAt) || d.Boundary != a.SourceChangeSetHighWatermark ||
		d.Derivation != a.DerivationVersion || d.CaptureSHA256 != a.SourceCaptureSHA256 || !sameReportingValue(d.ReferencePacks, a.ReferencePacks) {
		return a, "", errors.New("reporting: snapshot descriptor differs from render recipe")
	}
	if a.Identity.PreviewAttemptID != nil {
		source, err := reportcomposition.PreviewSourceForRenderTx(ctx, tx, jobID)
		if err != nil {
			return a, "", err
		}
		if source.PreviewAttemptID.String() != *a.Identity.PreviewAttemptID || source.RenderAttemptID != jobID ||
			source.IncidentID != incident || source.CreatedByUserID != actor || !source.CreatedAt.Equal(at) ||
			a.CompositionID == nil || source.CompositionID.String() != *a.CompositionID ||
			source.SnapshotID != a.SnapshotID || source.DerivationVersion != a.DerivationVersion ||
			source.TemplateID != a.TemplateContract.TemplateID || source.TemplateVersion != a.TemplateContract.TemplateVersion ||
			source.RedactionProfileID != a.Profile.ProfileID || source.RedactionProfileVersion != a.Profile.Version ||
			a.ExpectedProfileSHA256 == nil || source.RedactionProfileSHA256 != *a.ExpectedProfileSHA256 ||
			source.OutputKind != a.OutputKind || source.RenderEnvironmentProfile != "cartulary.reporting.local.v1" ||
			!equalFrozenJSON(source.OutputOptions, a.OutputOptions) || !equalFrozenJSON(source.GraphProjectionRefs, a.GraphProjectionRefs) ||
			!equalFrozenJSON(source.RecipientPartitionRefs, []byte(`[]`)) || !equalFrozenJSON(source.PreviewSourceJSON, a.CompositionJSON) ||
			!equalOptionalString(source.CompositionSHA256, a.CompositionSHA256) {
			return a, "", errors.New("reporting: preview source differs from render recipe")
		}
	} else if a.CompositionID != nil {
		compositionID, err := uuid.Parse(*a.CompositionID)
		if err != nil {
			return a, "", err
		}
		composition, err := reportcomposition.ResolveReleaseTupleTx(ctx, tx, incident, a.TemplateContract.TemplateID, a.TemplateContract.TemplateVersion, reportcomposition.ReleaseTuple{CompositionID: compositionID, CompositionVersion: derefString(a.CompositionVersion), CompositionSHA256: derefString(a.CompositionSHA256)})
		if err != nil {
			return a, "", err
		}
		if !equalFrozenJSON(composition.CanonicalComposition, a.CompositionJSON) {
			return a, "", errors.New("reporting: composition differs from render recipe")
		}
	}
	return a, digest, nil
}

func validateRenderJobAdmission(a renderAdmission, id uuid.UUID, retained jobs.RetainedExtensionJob) error {
	invalid := errors.New("reporting: render recipe differs from Jobs admission")
	incident, err := uuid.Parse(a.IncidentID)
	if err != nil {
		return invalid
	}
	actor, err := uuid.Parse(a.ActorUserID)
	if err != nil {
		return invalid
	}
	at, err := time.Parse(reportingTimestampLayout, a.Identity.RenderAdmittedAt)
	if err != nil {
		return invalid
	}
	kind, route, scopeKey := ReleaseCreateJobKind, "releases.create", a.SnapshotID
	if a.Identity.PreviewAttemptID != nil {
		if a.CompositionID == nil {
			return invalid
		}
		kind, route, scopeKey = CompositionPreviewJobKind, "report_compositions.preview", *a.CompositionID
	}
	r := retained.Resource
	if r.JobID != id.String() || retained.JobKind != kind || retained.HandlerName != JobWorkerKind || retained.OwnerProfileID != ProfileID ||
		r.AuthPolicy != jobs.AuthPolicyIncidentMembership || r.Scope.Kind != jobs.ScopeKindIncident || r.Scope.IncidentID == nil || *r.Scope.IncidentID != incident ||
		r.SubmittedByUserID == nil || *r.SubmittedByUserID != a.ActorUserID || !r.SubmittedAt.Equal(at) {
		return invalid
	}
	if retained.ExpiredAt != nil {
		if r.Status != jobs.StatusSucceeded && r.Status != jobs.StatusFailed && r.Status != jobs.StatusCanceled ||
			len(retained.IdempotencyIdentity) != 0 || retained.RequestSHA256 != "" || retained.RouteKey != "" || retained.ScopeKey != "" {
			return invalid
		}
		return nil
	}
	expected, err := jobs.NewExtensionJobAdmission(ProfileID, jobs.NewRouteIdempotencyKey(route, actor, scopeKey, a.ClientTxnID), jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incident}, a.NormalizedRequest)
	if err != nil || retained.RequestSHA256 != expected.NormalizedRequestSHA256 || retained.RouteKey != expected.IdempotencyRouteKey ||
		retained.ScopeKey != expected.IdempotencyScopeKey || !equalFrozenJSON(retained.IdempotencyIdentity, expected.IdempotencyIdentity) {
		return invalid
	}
	return nil
}
