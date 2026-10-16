package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"sort"
	"time"
)

// RenderCandidate is an owner-validated publication capability. Its contents
// are private and detached from caller buffers. Validation happens before the
// finalizer locks Jobs; commit rechecks the independently retained input seal.
type RenderCandidate struct {
	jobID       uuid.UUID
	preparation renderPreparation
	rendered    RenderedRelease
}

func (s *Store) PrepareRenderCandidate(ctx context.Context, execution jobs.Execution, input renderInput, rendered RenderedRelease) (RenderCandidate, error) {
	return prepareRenderCandidate(ctx, execution.JobID(), input, rendered)
}

// ValidateRenderCandidate supports publication/replay clients that do not hold
// a worker's prepared inputs. It uses a separate consistent read, never the
// finalizer transaction, and does not modify history or invoke a renderer.
func ValidateRenderCandidate(ctx context.Context, pool *pgxpool.Pool, jobID uuid.UUID, rendered RenderedRelease) (RenderCandidate, error) {
	if pool == nil {
		return RenderCandidate{}, errors.New("reporting: candidate validation requires storage")
	}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return RenderCandidate{}, err
	}
	defer rollbackSnapshotTx(tx)
	_, input, err := materializeRetainedRenderTx(ctx, tx, jobID)
	if err != nil {
		return RenderCandidate{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return RenderCandidate{}, err
	}
	return prepareRenderCandidate(ctx, jobID, input, rendered)
}

func prepareRenderCandidate(ctx context.Context, jobID uuid.UUID, input renderInput, rendered RenderedRelease) (RenderCandidate, error) {
	if err := ctx.Err(); err != nil {
		return RenderCandidate{}, err
	}
	if jobID == uuid.Nil || input.preparation == nil {
		return RenderCandidate{}, errors.New("reporting: candidate lacks preparation")
	}
	// Detach before validation: external callers cannot mutate a validated proof.
	raw, err := json.Marshal(retainedRenderArtifacts(rendered))
	if err != nil {
		return RenderCandidate{}, err
	}
	var detached RenderedRelease
	if err := json.Unmarshal(raw, &detached); err != nil {
		return RenderCandidate{}, err
	}
	if err := validateRenderCandidateBindingContext(ctx, input, detached); err != nil {
		return RenderCandidate{}, err
	}
	if err := ctx.Err(); err != nil {
		return RenderCandidate{}, err
	}
	return RenderCandidate{jobID: jobID, preparation: *input.preparation, rendered: detached}, nil
}

func renderCandidateForCommitTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID, candidate RenderCandidate) (renderAdmission, renderInput, RenderedRelease, error) {
	var noInput renderInput
	var noOutput RenderedRelease
	if candidate.jobID != jobID || jobID == uuid.Nil {
		return renderAdmission{}, noInput, noOutput, errors.New("reporting: invalid publication candidate")
	}
	a, digest, err := readRenderAdmissionTx(ctx, tx, jobID)
	if err != nil {
		return a, noInput, noOutput, err
	}
	seal, _, err := readRenderPreparationTx(ctx, tx, jobID, a, digest)
	if err != nil {
		return a, noInput, noOutput, err
	}
	if seal == nil || *seal != candidate.preparation {
		return a, noInput, noOutput, errors.New("reporting: candidate preparation changed")
	}
	input, err := renderAdmissionMetadata(a, seal)
	if err != nil {
		return a, noInput, noOutput, err
	}
	input.RedactionProfile = candidate.rendered.Profile
	return a, input, candidate.rendered, nil
}

// Metadata is sufficient for failure records and committing validated output.
// An absent preparation has no model/profile digest; callers must not fabricate
// a release row from that state.
func renderAdmissionMetadata(a renderAdmission, seal *renderPreparation) (renderInput, error) {
	at, err := time.Parse(reportingTimestampLayout, a.Identity.RenderAdmittedAt)
	if err != nil {
		return renderInput{}, err
	}
	input := renderInput{ActorUserID: a.ActorUserID, IncidentID: a.IncidentID, SnapshotID: a.SnapshotID, SnapshotAt: a.SnapshotAt,
		SourceChangeSetHighWatermark: a.SourceChangeSetHighWatermark, DerivationVersion: a.DerivationVersion,
		TemplateID: a.TemplateContract.TemplateID, TemplateVersion: a.TemplateContract.TemplateVersion, TemplateContract: a.TemplateContract,
		RedactionProfileID: a.Profile.ProfileID, RedactionProfileVersion: a.Profile.Version, RedactionProfile: a.Profile,
		OutputKind: a.OutputKind, OutputOptions: a.OutputOptions, ReleaseScope: a.ReleaseScope, RecipientPartitionRefs: a.RecipientPartitionRefs,
		GraphProjectionRefs: a.GraphProjectionRefs, CompositionJSON: a.CompositionJSON, CompositionID: a.CompositionID, CompositionVersion: a.CompositionVersion, CompositionSHA256: a.CompositionSHA256,
		RenderAdmittedAt: at, ExportModel: ExportModel{RenderIdentity: a.Identity, SnapshotContent: SnapshotContent{ReferencePacks: a.ReferencePacks}}}
	if seal != nil {
		input.ExportModelSHA256 = seal.ExportModelSHA256
		input.RedactionProfileSHA256 = seal.ProfileSHA256
		input.ExportModel.ExportModelID = seal.ExportModelID
		input.ExportModel.SnapshotModelID = seal.SnapshotModelID
	}
	return input, nil
}

// Compare the exact durable artifact projection, including member bytes and
// private reveal-map bytes. Transient renderer models are not publication data.
func retainedRenderArtifacts(value RenderedRelease) RenderedRelease {
	r := value.Redaction
	value.Redaction = RedactionResult{TokenManifestJSON: r.TokenManifestJSON, TokenManifestSHA256: r.TokenManifestSHA256, RevealMapJSON: r.RevealMapJSON, RevealMapSHA256: r.RevealMapSHA256}
	value.RenderBundle.Files = append([]RenderBundleFile{}, value.RenderBundle.Files...)
	sort.Slice(value.RenderBundle.Files, func(i, j int) bool { return value.RenderBundle.Files[i].Path < value.RenderBundle.Files[j].Path })
	return value
}

func sameRetainedRenderArtifacts(left, right RenderedRelease) bool {
	a, err := json.Marshal(retainedRenderArtifacts(left))
	if err != nil {
		return false
	}
	b, err := json.Marshal(retainedRenderArtifacts(right))
	return err == nil && bytes.Equal(a, b)
}
