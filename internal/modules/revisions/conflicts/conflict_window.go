package conflicts

import (
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidSnapshotProjector = errors.New("revisions conflicts: invalid snapshot projector")

// RevisionSnapshotProjector reconstructs only the source-owned writable cells
// required by conflict calculation. Presentation rows remain live material and
// are never retained as revision history.
type RevisionSnapshotProjector struct {
	snapshotSchemaID string
	fieldSourceKeys  map[string]string
}

func NewRevisionSnapshotProjector(snapshotSchemaID string, fieldSourceKeys map[string]string) (RevisionSnapshotProjector, error) {
	if strings.TrimSpace(snapshotSchemaID) == "" || len(fieldSourceKeys) == 0 {
		return RevisionSnapshotProjector{}, ErrInvalidSnapshotProjector
	}
	copied := make(map[string]string, len(fieldSourceKeys))
	for fieldKey, sourceKey := range fieldSourceKeys {
		if strings.TrimSpace(fieldKey) == "" || strings.TrimSpace(sourceKey) == "" {
			return RevisionSnapshotProjector{}, ErrInvalidSnapshotProjector
		}
		copied[fieldKey] = sourceKey
	}
	return RevisionSnapshotProjector{snapshotSchemaID: snapshotSchemaID, fieldSourceKeys: copied}, nil
}

type PatchConflictWindow struct {
	BaseRow       map[string]any
	ChangedFields map[string]PatchChangedField
}

type PatchChangedField struct {
	ServerUpdatedBy uuid.UUID
	ServerUpdatedAt time.Time
}

type RevisionWindowError struct {
	RecordID          uuid.UUID
	BaseRowVersion    int64
	CurrentRowVersion int64
}

func (e *RevisionWindowError) Error() string {
	return "workbook conflict revision window is unavailable"
}

// BuildCanonicalPatchConflictWindow reconstructs conflict facts using canonical
// source snapshots. Its exact source-key mapping is supplied by the source
// owner, so generic Revisions code contains no source vocabulary.
func BuildCanonicalPatchConflictWindow(recordID uuid.UUID, baseRowVersion int64, currentRowVersion int64, rows []RevisionWindowRow, descriptors FieldDescriptorSet, projector RevisionSnapshotProjector) (PatchConflictWindow, error) {
	return buildPatchConflictWindow(recordID, baseRowVersion, currentRowVersion, rows, descriptors, projector.Project)
}

func buildPatchConflictWindow(recordID uuid.UUID, baseRowVersion int64, currentRowVersion int64, rows []RevisionWindowRow, descriptors FieldDescriptorSet, decode func([]byte) (map[string]any, bool)) (PatchConflictWindow, error) {
	window := PatchConflictWindow{ChangedFields: make(map[string]PatchChangedField)}
	for _, row := range rows {
		if row.RowVersion == baseRowVersion {
			baseRow, ok := decode(row.AfterJSON)
			if !ok {
				return PatchConflictWindow{}, &RevisionWindowError{RecordID: recordID, BaseRowVersion: baseRowVersion, CurrentRowVersion: currentRowVersion}
			}
			window.BaseRow = baseRow
			continue
		}
		beforeRow, beforeOK := decode(row.BeforeJSON)
		afterRow, afterOK := decode(row.AfterJSON)
		if window.BaseRow == nil || !beforeOK || !afterOK {
			return PatchConflictWindow{}, &RevisionWindowError{RecordID: recordID, BaseRowVersion: baseRowVersion, CurrentRowVersion: currentRowVersion}
		}
		for _, fieldKey := range changedRevisionWritableFieldKeys(descriptors, beforeRow, afterRow) {
			window.ChangedFields[fieldKey] = PatchChangedField{
				ServerUpdatedBy: row.ActorUserID,
				ServerUpdatedAt: row.CreatedAt.UTC(),
			}
		}
		for _, fact := range row.ConflictFacts {
			if !descriptors.Writable(fact.FieldKey) {
				continue
			}
			if _, alreadyKnown := window.ChangedFields[fact.FieldKey]; !alreadyKnown && fact.BeforePresent {
				cell, ok := decodeConflictCell(fact.BeforeValue)
				if !ok {
					return PatchConflictWindow{}, &RevisionWindowError{RecordID: recordID, BaseRowVersion: baseRowVersion, CurrentRowVersion: currentRowVersion}
				}
				cells, _ := window.BaseRow["cells"].(map[string]any)
				if cells == nil {
					cells = map[string]any{}
					window.BaseRow["cells"] = cells
				}
				if _, present := cells[fact.FieldKey]; !present {
					cells[fact.FieldKey] = cell
				}
			}
			window.ChangedFields[fact.FieldKey] = PatchChangedField{
				ServerUpdatedBy: row.ActorUserID,
				ServerUpdatedAt: row.CreatedAt.UTC(),
			}
		}
	}
	if window.BaseRow == nil {
		return PatchConflictWindow{}, &RevisionWindowError{RecordID: recordID, BaseRowVersion: baseRowVersion, CurrentRowVersion: currentRowVersion}
	}
	return window, nil
}

func decodeConflictCell(data []byte) (map[string]any, bool) {
	if len(data) == 0 {
		return nil, false
	}
	var cell map[string]any
	if err := json.Unmarshal(data, &cell); err != nil || cell == nil {
		return nil, false
	}
	return cell, true
}

func (projector RevisionSnapshotProjector) Project(data []byte) (map[string]any, bool) {
	if len(data) == 0 || projector.snapshotSchemaID == "" || len(projector.fieldSourceKeys) == 0 {
		return nil, false
	}
	var envelope map[string]any
	if err := json.Unmarshal(data, &envelope); err != nil || len(envelope) != 3 {
		return nil, false
	}
	if envelope["snapshot_schema_id"] != projector.snapshotSchemaID {
		return nil, false
	}
	source, ok := envelope["source"].(map[string]any)
	if !ok {
		return nil, false
	}
	cells := make(map[string]any, len(projector.fieldSourceKeys))
	for fieldKey, sourceKey := range projector.fieldSourceKeys {
		value, present := source[sourceKey]
		if !present {
			continue
		}
		cells[fieldKey] = map[string]any{"value": value}
	}
	return map[string]any{"cells": cells}, true
}

func changedRevisionWritableFieldKeys(descriptors FieldDescriptorSet, beforeRow map[string]any, afterRow map[string]any) []string {
	beforeCells, _ := beforeRow["cells"].(map[string]any)
	afterCells, _ := afterRow["cells"].(map[string]any)
	changed := make([]string, 0)
	for fieldKey, afterCell := range afterCells {
		if !descriptors.Writable(fieldKey) {
			continue
		}
		if !reflect.DeepEqual(beforeCells[fieldKey], afterCell) {
			changed = append(changed, fieldKey)
		}
	}
	slices.Sort(changed)
	return changed
}
