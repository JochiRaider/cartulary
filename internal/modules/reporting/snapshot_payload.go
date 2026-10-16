package reporting

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/sourceboundary"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
)

// The current snapshot Job kind owns this closed private shape. A retained
// input is admitted as a whole; decoding must not discard unknown members or
// fill omitted values, even when the attacker recomputes the model digest.
func decodeSnapshotJobPayload(raw []byte, jobID, incidentID, actorID uuid.UUID, admittedAt time.Time) (snapshotCreateJobPayload, error) {
	invalid := errors.New("reporting: invalid frozen snapshot payload")
	var payload snapshotCreateJobPayload
	canonical, err := canonicalReportingBytes(raw)
	if err != nil {
		return payload, invalid
	}
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.UseNumber()
	if decoder.Decode(&payload) != nil {
		return snapshotCreateJobPayload{}, invalid
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return snapshotCreateJobPayload{}, invalid
	}
	projected, err := canonicalReportingBytes(encoded)
	if err != nil || !bytes.Equal(canonical, projected) ||
		jobID == uuid.Nil || incidentID == uuid.Nil || actorID == uuid.Nil ||
		payload.IncidentID != incidentID.String() || payload.ActorUserID != actorID.String() ||
		!payload.JobAdmittedAt.Equal(admittedAt) || payload.SnapshotAt.Before(admittedAt) {
		return snapshotCreateJobPayload{}, invalid
	}
	capture, err := decodeSnapshotCapture(payload.SourceCapture, payload.SourceCaptureSHA256)
	if err != nil || capture.SnapshotID != reportingResourceID(jobID, "snapshot").String() ||
		capture.IncidentID != payload.IncidentID || !capture.SnapshotAt.Equal(payload.SnapshotAt) ||
		capture.SourceChangeSetHighWatermark != payload.SourceChangeSetHighWatermark {
		return snapshotCreateJobPayload{}, invalid
	}
	if err := sourceboundary.ValidateRetained(sourceboundary.Boundary{
		Token: payload.SourceChangeSetHighWatermark, CanonicalJSON: payload.SourceBoundaryJSON,
	}, incidentID); err != nil {
		return snapshotCreateJobPayload{}, invalid
	}
	normalized, err := normalizeSnapshotRequest(incidentID, payload.ClientTxnID, payload.RequestedBoundary)
	if err != nil {
		return snapshotCreateJobPayload{}, invalid
	}
	// The private hash preimage has an explicit null for an omitted selector;
	// the public route correctly rejects explicit JSON null. Reconstruct the
	// public shape without translating omission into a supplied selector.
	wire := map[string]any{"incident_id": incidentID.String(), "client_txn_id": payload.ClientTxnID}
	if payload.RequestedBoundary != nil {
		wire["source_change_set_high_watermark"] = *payload.RequestedBoundary
	}
	wireJSON, err := json.Marshal(wire)
	if err != nil {
		return snapshotCreateJobPayload{}, invalid
	}
	request, apiErr := DecodeCreateSnapshotRequest(bytes.NewReader(wireJSON))
	if apiErr != nil || request.ClientTxnID != payload.ClientTxnID || payload.RequestedBoundary != nil && *payload.RequestedBoundary != payload.SourceChangeSetHighWatermark || !bytes.Equal(request.Normalized, normalized) {
		return snapshotCreateJobPayload{}, invalid
	}
	return payload, nil
}

func validateSnapshotJobAdmission(payload snapshotCreateJobPayload, jobID uuid.UUID, retained jobs.RetainedExtensionJob) error {
	invalid := errors.New("reporting: frozen snapshot differs from Jobs admission")
	incidentID, err := uuid.Parse(payload.IncidentID)
	if err != nil {
		return invalid
	}
	actorID, err := uuid.Parse(payload.ActorUserID)
	if err != nil {
		return invalid
	}
	if retained.Resource.JobID != jobID.String() || retained.JobKind != SnapshotCreateJobKind ||
		retained.HandlerName != JobWorkerKind || retained.OwnerProfileID != ProfileID || retained.Resource.AuthPolicy != jobs.AuthPolicyIncidentMembership ||
		retained.Resource.Scope.Kind != jobs.ScopeKindIncident || retained.Resource.Scope.IncidentID == nil || *retained.Resource.Scope.IncidentID != incidentID ||
		retained.Resource.SubmittedByUserID == nil || *retained.Resource.SubmittedByUserID != payload.ActorUserID ||
		!retained.Resource.SubmittedAt.Equal(payload.JobAdmittedAt) {
		return invalid
	}
	// Jobs compacts expired public responses without deleting owner artifacts or
	// their private admission inputs. Do not resurrect erased idempotency state
	// or make a valid historical snapshot depend on public response retention.
	if retained.ExpiredAt != nil {
		if (retained.Resource.Status != jobs.StatusSucceeded && retained.Resource.Status != jobs.StatusCanceled && retained.Resource.Status != jobs.StatusFailed) || len(retained.IdempotencyIdentity) != 0 || retained.RequestSHA256 != "" || retained.RouteKey != "" || retained.ScopeKey != "" {
			return invalid
		}
		return nil
	}
	normalized, err := normalizeSnapshotRequest(incidentID, payload.ClientTxnID, payload.RequestedBoundary)
	if err != nil {
		return invalid
	}
	expected, err := jobs.NewExtensionJobAdmission(ProfileID,
		jobs.NewRouteIdempotencyKey("snapshots.create", actorID, payload.IncidentID, payload.ClientTxnID),
		jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID}, normalized)
	if err != nil || retained.RequestSHA256 != expected.NormalizedRequestSHA256 ||
		retained.RouteKey != expected.IdempotencyRouteKey || retained.ScopeKey != expected.IdempotencyScopeKey ||
		!equalFrozenJSON(retained.IdempotencyIdentity, expected.IdempotencyIdentity) {
		return invalid
	}
	return nil
}

func validateSnapshotAdmissionBinding(payload snapshotCreateJobPayload, jobID uuid.UUID, record SnapshotRecord) error {
	if _, err := decodeSnapshotRecordCapture(record); err != nil {
		return err
	}
	admitted, err := canonicalReportingBytes(payload.SourceCapture)
	if err != nil || payload.SourceCaptureSHA256 != record.SourceCaptureSHA256 ||
		!equalFrozenJSON(admitted, record.SourceCaptureJSON) || payload.IncidentID != record.IncidentID.String() ||
		payload.ActorUserID != record.CreatedByUserID.String() || payload.ClientTxnID != record.ClientTxnID ||
		!payload.SnapshotAt.Equal(record.SnapshotAt) || payload.SourceChangeSetHighWatermark != record.SourceChangeSetHighWatermark ||
		!equalFrozenJSON(payload.SourceBoundaryJSON, record.SourceBoundaryJSON) || record.CreateJobID != jobID ||
		reportingResourceID(jobID, "snapshot") != record.SnapshotID {
		return errors.New("reporting: retained snapshot differs from admission")
	}
	return nil
}
