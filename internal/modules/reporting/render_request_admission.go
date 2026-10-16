package reporting

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/sourceboundary"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// The descriptor is sufficient for admission. Reading a snapshot's public
// metadata or admitting a render must not transfer its case-content capture.
type snapshotDescriptor struct {
	ID             uuid.UUID
	Incident       uuid.UUID
	Actor          uuid.UUID
	Transaction    string
	At             time.Time
	Boundary       string
	BoundaryJSON   []byte
	Derivation     string
	CaptureSHA256  string
	CreateJob      uuid.UUID
	CreatedAt      time.Time
	ReferencePacks reference_data.SetBinding
}

func readSnapshotDescriptorTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (snapshotDescriptor, error) {
	var d snapshotDescriptor
	err := tx.QueryRow(ctx, `SELECT snapshot_id,incident_id,created_by_user_id,client_txn_id,snapshot_at,source_change_set_high_watermark,source_boundary_json,derivation_version,source_capture_sha256,create_job_id,created_at FROM reporting_snapshots WHERE snapshot_id=$1`, id).Scan(&d.ID, &d.Incident, &d.Actor, &d.Transaction, &d.At, &d.Boundary, &d.BoundaryJSON, &d.Derivation, &d.CaptureSHA256, &d.CreateJob, &d.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return d, ErrNotFound
	}
	if err != nil {
		return d, err
	}
	if d.Derivation != DerivationVersion {
		return d, &UnsupportedSnapshotDerivationError{DerivationVersion: d.Derivation}
	}
	admission, err := readSnapshotAdmissionTx(ctx, tx, d.CreateJob)
	if err != nil {
		return d, err
	}
	retained, err := jobs.ReadRetainedExtensionJob(ctx, tx, d.CreateJob)
	if err != nil {
		return d, err
	}
	if retained.Resource.Status != jobs.StatusSucceeded || d.ID != reportingResourceID(d.CreateJob, "snapshot") ||
		d.Incident.String() != admission.IncidentID || d.Actor.String() != admission.ActorUserID || d.Transaction != admission.ClientTxnID ||
		d.At.Before(admission.AdmittedAt) || !d.CreatedAt.Equal(d.At) || !sha256HexPattern.MatchString(d.CaptureSHA256) ||
		admission.RequestedBoundary != nil && *admission.RequestedBoundary != d.Boundary {
		return d, errors.New("reporting: invalid published snapshot descriptor")
	}
	if err := sourceboundary.ValidateRetained(sourceboundary.Boundary{Token: d.Boundary, CanonicalJSON: d.BoundaryJSON}, d.Incident); err != nil {
		return d, err
	}
	d.ReferencePacks = admission.ReferencePacks
	return d, nil
}

func (s *Store) admitReleaseJob(ctx context.Context, params CreateReleaseParams) (CreateReleaseResult, error) {
	params.Now = params.Now.UTC().Truncate(time.Microsecond)
	key := authn.RouteIdempotencyKey{RouteKey: "releases.create", ActorUserID: params.ActorUserID, ScopeKey: params.Request.SnapshotID.String(), ClientTxnID: params.Request.ClientTxnID}
	sum := sha256.Sum256(params.Request.Normalized)
	tx, release, err := authn.BeginRepeatableRouteIdempotencyTx(ctx, s.pool, key)
	if err != nil {
		return CreateReleaseResult{}, err
	}
	defer release()
	existing, err := lookupRouteIdempotencyTx(ctx, tx, key)
	if err == nil {
		if !bytes.Equal(existing.RequestHash, sum[:]) {
			return CreateReleaseResult{}, authn.ErrClientTxnConflict
		}
		var resource jobs.Resource
		if err := json.Unmarshal(existing.ResponseJSON, &resource); err != nil {
			return CreateReleaseResult{}, err
		}
		jobID, err := uuid.Parse(resource.JobID)
		if err != nil {
			return CreateReleaseResult{}, err
		}
		if _, _, err := readRenderAdmissionTx(ctx, tx, jobID); err != nil {
			return CreateReleaseResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return CreateReleaseResult{}, err
		}
		return CreateReleaseResult{Job: resource, JobID: jobID, Replayed: true}, nil
	}
	if !errors.Is(err, authn.ErrNotFound) {
		return CreateReleaseResult{}, err
	}
	descriptor, err := readSnapshotDescriptorTx(ctx, tx, params.Request.SnapshotID)
	if err != nil {
		return CreateReleaseResult{}, err
	}
	profile, err := selectedRedactionProfile(params.Request.RedactionProfileID, params.Request.RedactionProfileVersion, params.Request.RecipientPartitionRefs)
	if err != nil || !profileSupportsScope(RedactionProfileView{ProfileStatus: profile.ProfileStatus, AllowedReleaseScopes: profile.AllowedReleaseScopes}, params.Request.ReleaseScope) {
		return CreateReleaseResult{}, &InvalidReleaseRequestError{Field: "redaction_profile_id", ReasonCode: "unsupported_redaction_profile"}
	}
	profileSHA, err := ValidateRedactionProfile(profile)
	if err != nil {
		return CreateReleaseResult{}, err
	}
	if s.referencePacks == nil {
		return CreateReleaseResult{}, errors.New("reporting: Reference Pack retention is required")
	}
	if err := s.referencePacks.ValidateBindingTx(ctx, tx, descriptor.ReferencePacks); err != nil {
		var missing *reference_data.ConsumerError
		if errors.As(err, &missing) {
			return CreateReleaseResult{}, ErrRequiredReferencePackUnavailable
		}
		return CreateReleaseResult{}, err
	}
	resolved, err := resolveReleaseCompositionTupleTx(ctx, tx, descriptor.Incident, params.Request)
	if err != nil {
		return CreateReleaseResult{}, err
	}
	refs, err := decodeSourceProjectionRefs(params.Request.GraphProjectionRefs)
	if err != nil {
		return CreateReleaseResult{}, &InvalidReleaseRequestError{Field: "graph_projection_refs", ReasonCode: "invalid_value"}
	}
	scope := jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &descriptor.Incident}
	extension, err := jobs.NewExtensionJobAdmission(ProfileID, jobs.NewRouteIdempotencyKey(key.RouteKey, key.ActorUserID, key.ScopeKey, key.ClientTxnID), scope, params.Request.Normalized)
	if err != nil {
		return CreateReleaseResult{}, err
	}
	job, err := s.jobTransactions.CreateQueuedTx(ctx, tx, jobs.EnqueueParams{JobKind: ReleaseCreateJobKind, Scope: scope, SubmittedByUserID: params.ActorUserID, Cancelable: true, Progress: jobs.Progress{Completed: 0}, Extension: extension}, params.Now)
	if err != nil {
		return CreateReleaseResult{}, err
	}
	jobID := uuid.MustParse(job.JobID)
	graphs, err := s.graphSources.FreezeTx(ctx, tx, descriptor.Incident, jobID, refs, params.Now)
	if err != nil {
		return CreateReleaseResult{}, err
	}
	graphSHA, err := graphSourcesDigest(graphs)
	if err != nil {
		return CreateReleaseResult{}, err
	}
	templateSHA, err := templateManifestDigest(params.TemplateContract)
	if err != nil {
		return CreateReleaseResult{}, err
	}
	releaseID := reportingResourceID(jobID, "release").String()
	stamp, err := normalizedReportingTimestamp(params.Now)
	if err != nil {
		return CreateReleaseResult{}, err
	}
	a := renderAdmission{SchemaID: renderAdmissionSchema, Algorithm: renderPreparationAlgorithm, Identity: RenderIdentity{ReleaseID: &releaseID, RenderAdmittedAt: stamp},
		ActorUserID: params.ActorUserID.String(), IncidentID: descriptor.Incident.String(), SnapshotID: descriptor.ID.String(), SnapshotAt: descriptor.At,
		SourceChangeSetHighWatermark: descriptor.Boundary, DerivationVersion: descriptor.Derivation, SourceCaptureSHA256: descriptor.CaptureSHA256, ReferencePacks: descriptor.ReferencePacks,
		TemplateContract: params.TemplateContract, TemplateFiles: append([]TemplatePackFile{}, params.TemplateFiles...), TemplateSHA256: templateSHA,
		Profile: profile, ProfileSHA256: profileSHA, ReleaseScope: params.Request.ReleaseScope, RecipientPartitionRefs: append([]string{}, params.Request.RecipientPartitionRefs...),
		OutputKind: params.Request.OutputKind, OutputOptions: bytes.Clone(params.Request.OutputOptions), GraphProjectionRefs: bytes.Clone(params.Request.GraphProjectionRefs), GraphSources: graphs, GraphSourcesSHA256: graphSHA,
		CompositionID: optionalUUIDStringPointer(params.Request.CompositionID), CompositionVersion: cloneStringPtr(params.Request.CompositionVersion), CompositionSHA256: cloneStringPtr(params.Request.CompositionSHA256),
		ClientTxnID: params.Request.ClientTxnID, NormalizedRequest: bytes.Clone(params.Request.Normalized)}
	if resolved != nil {
		a.CompositionJSON = bytes.Clone(resolved.CanonicalComposition)
	}
	if err := insertRenderAdmissionTx(ctx, tx, jobID, "release_create", params.Now, a); err != nil {
		return CreateReleaseResult{}, err
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, sum[:], http.StatusAccepted, job); err != nil {
		return CreateReleaseResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CreateReleaseResult{}, err
	}
	return CreateReleaseResult{Job: job, JobID: jobID}, nil
}

func insertRenderAdmissionTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID, kind string, at time.Time, a renderAdmission) error {
	raw, err := marshalRenderAdmission(a)
	if err != nil {
		return err
	}
	digest := reportingObjectDigest(renderAdmissionSchema, raw)
	incident, err := uuid.Parse(a.IncidentID)
	if err != nil {
		return err
	}
	actor, err := uuid.Parse(a.ActorUserID)
	if err != nil {
		return err
	}
	if _, err := decodeRenderAdmission(raw, digest, jobID, incident, actor, at); err != nil {
		return err
	}
	if (kind == "release_create") != (a.Identity.ReleaseID != nil) || kind != "release_create" && kind != "composition_preview" {
		return errors.New("reporting: render admission kind mismatch")
	}
	_, err = tx.Exec(ctx, `INSERT INTO reporting_job_payloads(job_id,job_kind,incident_id,actor_user_id,request_json,created_at,updated_at,admitted_profile_sha256,admitted_graph_sources_sha256,admitted_template_manifest_sha256,render_admission_sha256) VALUES($1,$2,$3,$4,$9,$5,$5,$6,$7,$8,$10)`, jobID, kind, incident, actor, at, a.ProfileSHA256, a.GraphSourcesSHA256, a.TemplateSHA256, raw, digest)
	return err
}
