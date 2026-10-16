package reporting

import (
	"bytes"
	"context"
	"errors"
)

func equalFrozenJSON(left, right []byte) bool {
	a, err := canonicalReportingBytes(left)
	if err != nil {
		return false
	}
	b, err := canonicalReportingBytes(right)
	return err == nil && bytes.Equal(a, b)
}

func equalOptionalString(left, right *string) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}

// Snapshot state is retained independently of the job. Recomputing a changed
// job's own model digest must not authorize replacement source content.
func validateRenderSnapshotBinding(payload renderInput, snapshot SnapshotRecord) error {
	model, err := decodeSnapshotRecordModel(snapshot)
	if err != nil {
		return err
	}
	retained, err := canonicalJSON(model)
	if err != nil {
		return err
	}
	claimed, err := canonicalJSON(SnapshotModel{
		SourceFacts: payload.SourceFacts,
		SchemaID:    SnapshotModelSchemaID, SnapshotModelID: payload.ExportModel.SnapshotModelID,
		SnapshotContent: payload.ExportModel.SnapshotContent,
	})
	if err != nil || !bytes.Equal(retained, claimed) {
		return errors.New("reporting: frozen release source differs from retained snapshot")
	}
	return nil
}

func decodeSnapshotRecordCapture(record SnapshotRecord) (snapshotCapture, error) {
	if record.DerivationVersion != DerivationVersion {
		return snapshotCapture{}, &UnsupportedSnapshotDerivationError{DerivationVersion: record.DerivationVersion}
	}
	capture, err := decodeSnapshotCapture(record.SourceCaptureJSON, record.SourceCaptureSHA256)
	if err != nil {
		return snapshotCapture{}, err
	}
	if capture.SnapshotID != record.SnapshotID.String() || capture.IncidentID != record.IncidentID.String() ||
		!capture.SnapshotAt.Equal(record.SnapshotAt) || capture.DerivationVersion != record.DerivationVersion ||
		capture.SourceChangeSetHighWatermark != record.SourceChangeSetHighWatermark {
		return snapshotCapture{}, errors.New("reporting: retained snapshot binding mismatch")
	}
	return capture, nil
}

func decodeSnapshotRecordModel(record SnapshotRecord) (SnapshotModel, error) {
	capture, err := decodeSnapshotRecordCapture(record)
	if err != nil {
		return SnapshotModel{}, err
	}
	return materializeSnapshotCapture(context.Background(), capture)
}
