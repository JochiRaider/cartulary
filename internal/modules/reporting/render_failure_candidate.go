package reporting

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// All content work ends before this capability crosses the terminal boundary.
// Private fields prevent an external caller from fabricating validated facts.
type renderFailureCandidate struct {
	owner        *Store
	jobID        uuid.UUID
	inputVersion string
	inputPresent bool
	record       renderFailureRecord
	raw          []byte
	digest       string
	failure      RenderFailure
	release      *releaseCreateJobPayload
}

func (s *Store) prepareRenderFailureCandidate(ctx context.Context, execution jobs.Execution, failure RenderFailure) (renderFailureCandidate, error) {
	var none renderFailureCandidate
	if s == nil || s.pool == nil || s.jobTransactions == nil {
		return none, jobs.ErrNotConfigured
	}
	gate, err := s.pool.Begin(ctx)
	if err != nil {
		return none, err
	}
	defer rollbackSnapshotTx(gate)
	if _, err := s.jobTransactions.ExtensionCancellationContextTx(ctx, gate, execution); err != nil {
		return none, err
	}
	retained, err := s.jobTransactions.ReadRetainedExtensionJob(ctx, gate, execution.JobID())
	if err != nil {
		return none, err
	}
	if err := gate.Commit(ctx); err != nil {
		return none, err
	}
	return s.prepareRetainedRenderFailure(ctx, execution.JobID(), retained, failure)
}

func (s *Store) prepareRetainedRenderFailure(ctx context.Context, id uuid.UUID, retained jobs.RetainedExtensionJob, failure RenderFailure) (renderFailureCandidate, error) {
	var none renderFailureCandidate
	if err := validateRenderFailure(failure); err != nil {
		return none, err
	}
	source, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return none, err
	}
	defer rollbackSnapshotTx(source)
	candidate := renderFailureCandidate{owner: s, jobID: id, failure: failure}
	err = source.QueryRow(ctx, `SELECT xmin::text FROM reporting_job_payloads WHERE job_id=$1`, id).Scan(&candidate.inputVersion)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return none, err
	}
	candidate.inputPresent = err == nil
	var summary []byte
	var binding *renderFailureInputBinding
	// Authority denial may be retained without reading private case content.
	// An already classified integrity failure likewise has no trusted identity.
	if failure.FailureCode != "authorization_denied" && failure.FailureCode != "retained_input_invalid" {
		var integrity *retainedRenderInputError
		if !candidate.inputPresent {
			integrity = &retainedRenderInputError{missing: true}
		} else {
			a, admissionSHA, readErr := readRenderAdmissionTx(ctx, s.jobTransactions, source, id)
			if readErr != nil {
				if !errors.As(readErr, &integrity) {
					return none, readErr
				}
			} else {
				seal, preparationSHA, readErr := readRenderPreparationTx(ctx, source, id, a, admissionSHA)
				if readErr != nil {
					if !errors.As(readErr, &integrity) {
						return none, readErr
					}
				} else {
					input, readErr := renderAdmissionMetadata(a, seal)
					if readErr != nil {
						return none, readErr
					}
					binding = &renderFailureInputBinding{AdmissionSHA256: admissionSHA}
					if seal != nil {
						binding.PreparationSHA256 = &preparationSHA
					}
					summary, err = failedRenderSummary(input, failure)
					if err != nil {
						return none, err
					}
					if seal != nil && retained.JobKind == ReleaseCreateJobKind {
						payload, err := releasePayloadFromPrepared(a, input)
						if err != nil {
							return none, err
						}
						metadata := failureReleaseMetadata(payload)
						candidate.release = &metadata
					}
				}
			}
		}
		if integrity != nil && failure.FailureCode != "render_canceled" && failure.FailureCode != "render_timeout" && failure.ReasonCode != "partial_output_cleanup_failed" {
			candidate.failure = renderFailureDetails(integrity)
		}
	}
	at, err := normalizedReportingTimestamp(retained.Resource.SubmittedAt)
	if err != nil {
		return none, err
	}
	if summary == nil {
		summary, err = encodeFailedRenderSummary(nil, nil, at, candidate.failure)
		if err != nil {
			return none, err
		}
	}
	candidate.record, err = bindRenderFailureRecord(retained, summary)
	if err != nil {
		return none, err
	}
	candidate.record.InputBinding = binding
	candidate.raw, candidate.digest, err = encodeRenderFailureRecord(candidate.record)
	if err != nil {
		return none, err
	}
	if _, _, err := decodeRenderFailureRecord(candidate.raw, candidate.digest, retained); err != nil {
		return none, err
	}
	if err := ctx.Err(); err != nil {
		return none, err
	}
	if err := source.Commit(ctx); err != nil {
		return none, err
	}
	return candidate, nil
}

// Keep only the bounded values required to persist a justified failed release.
// Case content, authored policies, template files and expanded models stay out
// of the terminal transaction.
func failureReleaseMetadata(p releaseCreateJobPayload) releaseCreateJobPayload {
	return releaseCreateJobPayload{ClientTxnID: p.ClientTxnID, renderInput: renderInput{
		ActorUserID: p.ActorUserID, IncidentID: p.IncidentID, SnapshotID: p.SnapshotID, SnapshotAt: p.SnapshotAt,
		SourceChangeSetHighWatermark: p.SourceChangeSetHighWatermark, DerivationVersion: p.DerivationVersion,
		TemplateID: p.TemplateID, TemplateVersion: p.TemplateVersion, RedactionProfileID: p.RedactionProfileID,
		RedactionProfileVersion: p.RedactionProfileVersion, RedactionProfileSHA256: p.RedactionProfileSHA256,
		ExportModelSHA256: p.ExportModelSHA256, OutputKind: p.OutputKind, OutputOptions: append([]byte(nil), p.OutputOptions...),
		ReleaseScope: p.ReleaseScope, RecipientPartitionRefs: append([]string(nil), p.RecipientPartitionRefs...),
		GraphProjectionRefs: append([]byte(nil), p.GraphProjectionRefs...), CompositionID: p.CompositionID,
		CompositionVersion: p.CompositionVersion, CompositionSHA256: p.CompositionSHA256, RenderAdmittedAt: p.RenderAdmittedAt,
		ExportModel: ExportModel{RenderIdentity: p.ExportModel.RenderIdentity, SnapshotContent: SnapshotContent{ReferencePacks: p.ExportModel.ReferencePacks}},
	}}
}
