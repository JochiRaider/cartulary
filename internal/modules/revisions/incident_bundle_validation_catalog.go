package revisions

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrMissingHistoryTargetProvider    = errors.New("revisions: missing history target provider")
	ErrUnexpectedHistoryTargetProvider = errors.New("revisions: unexpected history target provider")
)

// IncidentBundleValidationCatalog is the immutable owner-composed boundary
// used by the Revisions portability validator. Source semantics stay
// owner-defined; generated projections are never treated as truth.
type IncidentBundleValidationCatalog struct {
	currentRows *DeleteRestoreSourceCatalog
	envelopes   IncidentBundleRecordEnvelopeReader
	targets     *TargetSemanticsCatalog
	schemas     map[string]string
}

type IncidentBundleRecordEnvelopeReader interface {
	RecordTypeTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (string, error)
}

// NewIncidentBundleValidationCatalog validates the complete current source
// contribution set and projects only the generic capabilities required by
// portable history validation.
func NewIncidentBundleValidationCatalog(
	envelopes IncidentBundleRecordEnvelopeReader,
	targets *TargetSemanticsCatalog,
	contributions []ProviderContribution,
) (*IncidentBundleValidationCatalog, error) {
	if envelopes == nil || targets == nil {
		return nil, ErrMissingHistoryTargetProvider
	}
	currentRows, err := buildDeleteRestoreSourceCatalog(contributions)
	if err != nil {
		return nil, fmt.Errorf("build incident-bundle validation catalog: %w", err)
	}
	schemas := make(map[string]string)
	for _, contribution := range contributions {
		for _, record := range contribution.Records {
			if strings.TrimSpace(record.SnapshotSchemaID) == "" {
				return nil, fmt.Errorf("%w: record type %q snapshot schema", ErrMissingHistoryTargetProvider, record.RecordType)
			}
			schemas[record.RecordType] = record.SnapshotSchemaID
		}
	}
	return &IncidentBundleValidationCatalog{
		currentRows: currentRows,
		envelopes:   envelopes,
		targets:     targets,
		schemas:     schemas,
	}, nil
}

func (c *IncidentBundleValidationCatalog) resolvesTargetKind(targetKind string) bool {
	if c == nil {
		return false
	}
	return c.targets.hasTargetKind(targetKind)
}

func (c *IncidentBundleValidationCatalog) validateSnapshot(recordID uuid.UUID, recordType string, value any) error {
	if c == nil {
		return ErrMissingHistoryTargetProvider
	}
	if value == nil {
		return nil
	}
	snapshot, ok := value.(map[string]any)
	if !ok {
		return ErrInvalidCapturedSnapshot
	}
	schemaID, ok := c.schemas[recordType]
	if !ok {
		return ErrMissingHistoryTargetProvider
	}
	return validatePersistedSnapshotValue(schemaID, recordID, recordType, snapshot)
}

func (c *IncidentBundleValidationCatalog) targetKinds() []string {
	if c == nil {
		return nil
	}
	return c.targets.TargetKinds()
}

func (c *IncidentBundleValidationCatalog) currentRowTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
) (map[string]any, error) {
	if c == nil || c.currentRows == nil {
		return nil, ErrMissingHistoryTargetProvider
	}
	recordType, err := c.envelopes.RecordTypeTx(ctx, tx, incidentID, recordID)
	if err != nil {
		return nil, err
	}
	source, ok := c.currentRows.Source(recordType)
	if !ok {
		return nil, fmt.Errorf("%w: current row provider", ErrUnexpectedHistoryTargetProvider)
	}
	return source.SnapshotTx(ctx, tx, recordID)
}

// currentHistoryRowMatchesTx compares a canonical retained history envelope
// with the source owner's canonical current-row snapshot.
func (c *IncidentBundleValidationCatalog) currentHistoryRowMatchesTx(
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
		sameCanonicalScalar(nestedRecord["record_id"], recordID.String()) &&
		canonicalObjectSubsetMatches(nestedRecord, currentRecord) &&
		canonicalObjectSubsetMatches(nestedSource, currentSource)
}

func canonicalObjectSubsetMatches(expected map[string]any, current map[string]any) bool {
	if len(expected) == 0 {
		return false
	}
	for key, expectedValue := range expected {
		currentValue, present := current[key]
		if !present || !canonicalValueSubsetMatches(expectedValue, currentValue) {
			return false
		}
	}
	return true
}

func canonicalValueSubsetMatches(expected any, current any) bool {
	switch expectedValue := expected.(type) {
	case map[string]any:
		currentValue, ok := current.(map[string]any)
		return ok && canonicalObjectSubsetMatches(expectedValue, currentValue)
	case []any:
		currentValue, ok := current.([]any)
		if !ok || len(expectedValue) != len(currentValue) {
			return false
		}
		for index := range expectedValue {
			if !canonicalValueSubsetMatches(expectedValue[index], currentValue[index]) {
				return false
			}
		}
		return true
	default:
		return sameCanonicalScalar(expected, current)
	}
}

func sameCanonicalScalar(left any, right any) bool {
	leftText, leftIsText := left.(string)
	rightText, rightIsText := right.(string)
	if leftIsText && rightIsText {
		leftTime, leftTimeErr := time.Parse(time.RFC3339Nano, leftText)
		rightTime, rightTimeErr := time.Parse(time.RFC3339Nano, rightText)
		if leftTimeErr == nil && rightTimeErr == nil {
			return leftTime.Truncate(time.Microsecond).Equal(rightTime.Truncate(time.Microsecond))
		}
	}
	return sameCanonicalJSON(map[string]any{"value": left}, map[string]any{"value": right})
}

func sameStoredCanonicalJSON(left map[string]any, right map[string]any) bool {
	leftNormalized, leftOK := normalizeStoredJSONValue(left).(map[string]any)
	rightNormalized, rightOK := normalizeStoredJSONValue(right).(map[string]any)
	return leftOK && rightOK && sameCanonicalJSON(leftNormalized, rightNormalized)
}

func normalizeStoredJSONValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, member := range typed {
			result[key] = normalizeStoredJSONValue(member)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for index, member := range typed {
			result[index] = normalizeStoredJSONValue(member)
		}
		return result
	case string:
		parsed, err := time.Parse(time.RFC3339Nano, typed)
		if err != nil {
			return typed
		}
		return parsed.UTC().Truncate(time.Microsecond).Format(time.RFC3339Nano)
	default:
		return value
	}
}
