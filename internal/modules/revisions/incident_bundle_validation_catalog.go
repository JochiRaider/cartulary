package revisions

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrMissingHistoryTargetProvider    = errors.New("revisions: missing history target provider")
	ErrUnexpectedHistoryTargetProvider = errors.New("revisions: unexpected history target provider")
)

// incidentBundleValidationCatalog reuses the immutable catalogs built for the
// rest of the Revisions runtime. Source semantics stay owner-defined;
// generated projections are never treated as truth.
type incidentBundleValidationCatalog struct {
	envelopes IncidentBundleRecordEnvelopeReader
	snapshots *RecordSnapshotCaptureCatalog
	targets   *TargetSemanticsCatalog
}

type IncidentBundleRecordEnvelopeReader interface {
	RecordTypeTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (string, error)
	VisitIncidentRecordIDsTx(context.Context, pgx.Tx, uuid.UUID, func(uuid.UUID) error) error
}

func newIncidentBundleValidationCatalog(
	envelopes IncidentBundleRecordEnvelopeReader,
	snapshots *RecordSnapshotCaptureCatalog,
	targets *TargetSemanticsCatalog,
) (*incidentBundleValidationCatalog, error) {
	if envelopes == nil || snapshots == nil || targets == nil {
		return nil, ErrMissingHistoryTargetProvider
	}
	return &incidentBundleValidationCatalog{
		envelopes: envelopes,
		snapshots: snapshots,
		targets:   targets,
	}, nil
}

func (c *incidentBundleValidationCatalog) validateContract() error {
	if c == nil || c.envelopes == nil || c.snapshots == nil || c.targets == nil ||
		len(c.snapshots.byRecordType) == 0 || len(c.targets.byTargetKind) == 0 {
		return ErrMissingHistoryTargetProvider
	}
	return nil
}

func (c *incidentBundleValidationCatalog) resolvesTargetKind(targetKind string) bool {
	if c == nil {
		return false
	}
	return c.targets.hasTargetKind(targetKind)
}

func (c *incidentBundleValidationCatalog) validateSnapshot(incidentID, recordID uuid.UUID, recordType string, value any) error {
	if c == nil {
		return ErrMissingHistoryTargetProvider
	}
	if value == nil {
		return nil
	}
	snapshot, ok := value.(map[string]any)
	if !ok {
		return ErrInvalidRecordSnapshot
	}
	return c.snapshots.validatePersisted(incidentID, recordID, recordType, snapshot)
}

func (c *incidentBundleValidationCatalog) targetKinds() []string {
	if c == nil {
		return nil
	}
	return c.targets.targetKinds()
}

func (c *incidentBundleValidationCatalog) currentRowTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
) (map[string]any, error) {
	if c == nil || c.snapshots == nil {
		return nil, ErrMissingHistoryTargetProvider
	}
	recordType, err := c.envelopes.RecordTypeTx(ctx, tx, incidentID, recordID)
	if err != nil {
		return nil, err
	}
	return c.currentSourceRowTx(ctx, tx, recordID, recordType)
}

func (c *incidentBundleValidationCatalog) currentSourceRowTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	recordType string,
) (map[string]any, error) {
	capture, ok := c.snapshots.byRecordType[recordType]
	if !ok {
		return nil, fmt.Errorf("%w: current row provider", ErrUnexpectedHistoryTargetProvider)
	}
	return capture.source.SnapshotTx(ctx, tx, recordID)
}

// currentHistoryRowMatchesTx compares a canonical retained history envelope
// with the source owner's canonical current-row snapshot.
func (c *incidentBundleValidationCatalog) currentHistoryRowMatchesTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
	historyRow map[string]any,
) (bool, error) {
	current, err := c.currentRowTx(ctx, tx, incidentID, recordID)
	if err != nil {
		return false, err
	}
	return canonicalHistoryRowMatchesSnapshot(recordID, historyRow, current), nil
}

func canonicalHistoryRowMatchesSnapshot(
	recordID uuid.UUID,
	historyRow map[string]any,
	current map[string]any,
) bool {
	currentRecord, recordOK := current["record"].(map[string]any)
	currentSource, sourceOK := current["source"].(map[string]any)
	if !recordOK || !sourceOK {
		return false
	}
	nestedRecord, recordPresent := historyRow["record"].(map[string]any)
	nestedSource, sourcePresent := historyRow["source"].(map[string]any)
	return recordPresent && sourcePresent &&
		nestedRecord["record_id"] == recordID.String() &&
		sameCanonicalJSON(nestedRecord, currentRecord) &&
		sameCanonicalJSON(nestedSource, currentSource)
}
