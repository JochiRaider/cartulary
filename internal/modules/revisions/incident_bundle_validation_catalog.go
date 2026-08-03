package revisions

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrDuplicateHistoryTargetProvider  = errors.New("revisions: duplicate history target provider")
	ErrMissingHistoryTargetProvider    = errors.New("revisions: missing history target provider")
	ErrUnexpectedHistoryTargetProvider = errors.New("revisions: unexpected history target provider")
)

// IncidentBundleValidationCatalog is the immutable owner-composed boundary
// used by the Revisions portability validator. Source semantics stay
// owner-defined; generated projections are never treated as truth.
type IncidentBundleValidationCatalog struct {
	currentRows *DeleteRestoreSourceCatalog
	envelopes   IncidentBundleRecordEnvelopeReader
	targets     map[string]SourceOwnerModule
}

type IncidentBundleRecordEnvelopeReader interface {
	RecordTypeTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (string, error)
}

// NewIncidentBundleValidationCatalog validates the complete current source
// contribution set and projects only the generic capabilities required by
// portable history validation.
func NewIncidentBundleValidationCatalog(
	envelopes IncidentBundleRecordEnvelopeReader,
	contributions []ProviderContribution,
) (*IncidentBundleValidationCatalog, error) {
	if envelopes == nil {
		return nil, ErrMissingHistoryTargetProvider
	}
	currentRows, _, _, err := buildProviderCatalogs(contributions)
	if err != nil {
		return nil, fmt.Errorf("build incident-bundle validation catalog: %w", err)
	}
	targets := map[string]SourceOwnerModule{"record": "records"}
	for _, contribution := range contributions {
		for _, record := range contribution.Records {
			targetKinds := record.HistoryTargetKinds
			if len(targetKinds) == 0 {
				targetKinds = []string{record.RecordType}
			}
			for _, targetKind := range targetKinds {
				if targetKind == "" {
					return nil, ErrMissingHistoryTargetProvider
				}
				if prior, duplicate := targets[targetKind]; duplicate {
					return nil, fmt.Errorf(
						"%w: %s claimed by %s and %s",
						ErrDuplicateHistoryTargetProvider,
						targetKind,
						prior,
						contribution.SourceOwnerModule,
					)
				}
				targets[targetKind] = contribution.SourceOwnerModule
			}
		}
		for _, target := range contribution.NonRowTargets {
			if target.TargetKind == "" {
				return nil, ErrMissingHistoryTargetProvider
			}
			if prior, duplicate := targets[target.TargetKind]; duplicate {
				return nil, fmt.Errorf(
					"%w: %s claimed by %s and %s",
					ErrDuplicateHistoryTargetProvider,
					target.TargetKind,
					prior,
					contribution.SourceOwnerModule,
				)
			}
			targets[target.TargetKind] = contribution.SourceOwnerModule
		}
	}
	return &IncidentBundleValidationCatalog{
		currentRows: currentRows,
		envelopes:   envelopes,
		targets:     targets,
	}, nil
}

func (c *IncidentBundleValidationCatalog) resolvesTargetKind(targetKind string) bool {
	if c == nil {
		return false
	}
	_, ok := c.targets[targetKind]
	return ok
}

func (c *IncidentBundleValidationCatalog) targetKinds() []string {
	if c == nil {
		return nil
	}
	result := make([]string, 0, len(c.targets))
	for targetKind := range c.targets {
		result = append(result, targetKind)
	}
	sort.Strings(result)
	return result
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

// currentHistoryRowMatchesTx compares an owner-authored history row with the
// source owner's canonical current-row snapshot. History rows may use either
// the canonical {record,source} form emitted by rollback or the workbook row
// form emitted by interactive mutations. The latter is matched only through
// source fields that the owner snapshot actually exposes; derived cells are
// deliberately ignored rather than treating a generated projection as truth.
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
	if nestedRecord, ok := historyRow["record"].(map[string]any); ok {
		nestedSource, sourcePresent := historyRow["source"].(map[string]any)
		matches := sourcePresent &&
			canonicalObjectSubsetMatches(nestedRecord, currentRecord) &&
			canonicalObjectSubsetMatches(nestedSource, currentSource)
		return matches
	}
	if !sameCanonicalScalar(historyRow["record_id"], recordID.String()) ||
		!sameCanonicalScalar(historyRow["row_version"], currentRecord["row_version"]) {
		return false
	}

	matchedSourceField := false
	if cells, ok := historyRow["cells"].(map[string]any); ok {
		for fieldKey, rawCell := range cells {
			cell, ok := rawCell.(map[string]any)
			if !ok {
				return false
			}
			separator := strings.LastIndexByte(fieldKey, '.')
			if separator < 0 || separator == len(fieldKey)-1 {
				continue
			}
			sourceValue, represented := currentSource[fieldKey[separator+1:]]
			if !represented {
				continue
			}
			value, present := cell["value"]
			if !present || !sameCanonicalScalar(value, sourceValue) {
				return false
			}
			matchedSourceField = true
		}
	}
	for key, value := range historyRow {
		if key == "record_id" || key == "row_version" || key == "cells" || key == "group_values" {
			continue
		}
		if sourceValue, represented := currentSource[key]; represented {
			if !sameCanonicalScalar(value, sourceValue) {
				return false
			}
			matchedSourceField = true
		}
	}
	return matchedSourceField
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
