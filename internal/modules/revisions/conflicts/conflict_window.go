package conflicts

import (
	"encoding/json"
	"reflect"
	"slices"
	"time"

	"github.com/google/uuid"
)

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

func BuildPatchConflictWindowWithDescriptors(recordID uuid.UUID, baseRowVersion int64, currentRowVersion int64, rows []RevisionWindowRow, descriptors FieldDescriptorSet) (PatchConflictWindow, error) {
	window := PatchConflictWindow{ChangedFields: make(map[string]PatchChangedField)}
	for _, row := range rows {
		if row.RowVersion == baseRowVersion {
			baseRow, ok := DecodeRevisionRow(row.AfterJSON)
			if !ok {
				return PatchConflictWindow{}, &RevisionWindowError{RecordID: recordID, BaseRowVersion: baseRowVersion, CurrentRowVersion: currentRowVersion}
			}
			window.BaseRow = baseRow
			continue
		}
		beforeRow, beforeOK := DecodeRevisionRow(row.BeforeJSON)
		afterRow, afterOK := DecodeRevisionRow(row.AfterJSON)
		if !beforeOK || !afterOK {
			return PatchConflictWindow{}, &RevisionWindowError{RecordID: recordID, BaseRowVersion: baseRowVersion, CurrentRowVersion: currentRowVersion}
		}
		for _, fieldKey := range ChangedRevisionWritableFieldKeysWithDescriptors(descriptors, beforeRow, afterRow) {
			window.ChangedFields[fieldKey] = PatchChangedField{
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

func DecodeRevisionRow(data []byte) (map[string]any, bool) {
	if len(data) == 0 {
		return nil, false
	}
	var row map[string]any
	if err := json.Unmarshal(data, &row); err != nil {
		return nil, false
	}
	if _, ok := row["cells"].(map[string]any); !ok {
		return nil, false
	}
	return row, true
}

func ChangedRevisionWritableFieldKeysWithDescriptors(descriptors FieldDescriptorSet, beforeRow map[string]any, afterRow map[string]any) []string {
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
