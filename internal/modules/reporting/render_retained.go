package reporting

import (
	"bytes"
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func releasePayloadFromPrepared(a renderAdmission, input renderInput) (releaseCreateJobPayload, error) {
	if a.Identity.ReleaseID == nil {
		return releaseCreateJobPayload{}, errors.New("reporting: release requires release admission")
	}
	request, apiErr := DecodeCreateReleaseRequest(bytes.NewReader(a.NormalizedRequest))
	if apiErr != nil {
		return releaseCreateJobPayload{}, errors.New("reporting: invalid retained release request")
	}
	return releaseCreateJobPayload{renderInput: input, ClientTxnID: a.ClientTxnID, NormalizedRequest: bytes.Clone(a.NormalizedRequest), Request: request}, nil
}

func previewPayloadFromPrepared(a renderAdmission, input renderInput) (compositionPreviewJobPayload, error) {
	if a.Identity.PreviewAttemptID == nil {
		return compositionPreviewJobPayload{}, errors.New("reporting: preview requires preview admission")
	}
	id, err := uuid.Parse(*a.Identity.PreviewAttemptID)
	if err != nil {
		return compositionPreviewJobPayload{}, err
	}
	return compositionPreviewJobPayload{renderInput: input, PreviewAttemptID: id, NormalizedRequest: bytes.Clone(a.NormalizedRequest)}, nil
}

// Recovery derives against retained sources in a consistent read transaction.
// Live publication must use a previously validated candidate instead of doing
// this incident-sized work under a Jobs transition lock.
func materializeRetainedRenderTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID) (renderAdmission, renderInput, error) {
	a, digest, err := readRenderAdmissionTx(ctx, tx, jobID)
	if err != nil {
		return a, renderInput{}, err
	}
	seal, _, err := readRenderPreparationTx(ctx, tx, jobID, a, digest)
	if err != nil {
		return a, renderInput{}, err
	}
	if seal == nil {
		return a, renderInput{}, errors.New("reporting: render preparation is absent")
	}
	record, err := getSnapshotRecordTx(ctx, tx, uuid.MustParse(a.SnapshotID))
	if err != nil {
		return a, renderInput{}, err
	}
	captured, err := getSnapshotPayloadTx(ctx, tx, record.CreateJobID)
	if err != nil {
		return a, renderInput{}, err
	}
	if err := validateSnapshotAdmissionBinding(captured, record.CreateJobID, record); err != nil {
		return a, renderInput{}, err
	}
	input, observed, err := materializeRenderAdmission(ctx, a, digest, record)
	if err != nil {
		return a, renderInput{}, err
	}
	if observed != *seal {
		return a, renderInput{}, errors.New("reporting: derived inputs differ from retained preparation")
	}
	return a, input, nil
}
