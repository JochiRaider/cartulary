package conflictwindows

import (
	"encoding/json"
	"reflect"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/historyquery"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type PatchConflictWindow struct {
	BaseRow       map[string]any
	ChangedFields map[string]PatchChangedField
}

type PatchChangedField struct {
	ServerUpdatedBy uuid.UUID
	ServerUpdatedAt time.Time
}

type FieldDescriptor struct {
	FieldKey                string
	Writable                bool
	ConflictResolutionClass string
}

type FieldDescriptorSet struct {
	fields map[string]FieldDescriptor
}

type RevisionWindowError struct {
	RecordID          uuid.UUID
	BaseRowVersion    int64
	CurrentRowVersion int64
}

func (e *RevisionWindowError) Error() string {
	return "workbook conflict revision window is unavailable"
}

func BuildPatchConflictWindow(recordID uuid.UUID, viewSchemaID string, baseRowVersion int64, currentRowVersion int64, rows []historyquery.RevisionWindowRow) (PatchConflictWindow, error) {
	descriptors := ViewSchemaFieldDescriptors(viewSchemaID)
	return BuildPatchConflictWindowWithDescriptors(recordID, baseRowVersion, currentRowVersion, rows, descriptors)
}

func BuildPatchConflictWindowWithDescriptors(recordID uuid.UUID, baseRowVersion int64, currentRowVersion int64, rows []historyquery.RevisionWindowRow, descriptors FieldDescriptorSet) (PatchConflictWindow, error) {
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

func NewFieldDescriptorSet(fields []FieldDescriptor) FieldDescriptorSet {
	descriptors := FieldDescriptorSet{fields: make(map[string]FieldDescriptor, len(fields))}
	for _, field := range fields {
		if field.FieldKey == "" {
			continue
		}
		descriptors.fields[field.FieldKey] = field
	}
	return descriptors
}

func ViewSchemaFieldDescriptors(viewSchemaID string) FieldDescriptorSet {
	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok {
		return FieldDescriptorSet{}
	}
	fields := make([]FieldDescriptor, 0, len(schema.Fields()))
	for _, field := range schema.Fields() {
		fields = append(fields, FieldDescriptor{
			FieldKey:                field.FieldKey,
			Writable:                field.Writable,
			ConflictResolutionClass: field.ConflictResolutionClass,
		})
	}
	return NewFieldDescriptorSet(fields)
}

func (s FieldDescriptorSet) Writable(fieldKey string) bool {
	if s.fields == nil || isReadOnlySystemField(fieldKey) {
		return false
	}
	field, ok := s.fields[fieldKey]
	return ok && field.Writable
}

func (s FieldDescriptorSet) ConflictResolutionClass(fieldKey string) string {
	if s.fields == nil {
		return ""
	}
	field, ok := s.fields[fieldKey]
	if !ok {
		return ""
	}
	return field.ConflictResolutionClass
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

func ChangedRevisionWritableFieldKeys(viewSchemaID string, beforeRow map[string]any, afterRow map[string]any) []string {
	return ChangedRevisionWritableFieldKeysWithDescriptors(ViewSchemaFieldDescriptors(viewSchemaID), beforeRow, afterRow)
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

func isReadOnlySystemField(fieldKey string) bool {
	switch fieldKey {
	case "record_id", "row_version", "version_id", "updated_at", "created_at", "created_by_user_id", "updated_by_user_id":
		return true
	default:
		return false
	}
}
