package tabularingest

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type ImportScalarValue struct {
	Kind      string
	Text      *string
	Timestamp *time.Time
	UUID      *uuid.UUID
	Number    *int64
	Bool      *bool
}

type ImportFieldValue struct {
	FieldKey            string
	NormalizedValue     ImportScalarValue
	SourceColumnOrdinal int
	SourceHeaderText    any
	RawValue            string
	CellKind            string
	TransformID         *string
	EmptyValuePolicy    string
	EntityBindingMode   *string
}

type ImportUnknownValue struct {
	SourceColumnOrdinal int
	SourceHeaderText    any
	RawValue            string
	CellKind            string
}

type ImportSourceRowProvenance struct {
	SourceRowRef int
}

type ImportOwnerCreateRequest struct {
	IncidentID          uuid.UUID
	ActorUserID         uuid.UUID
	TargetViewSchemaID  string
	ImportSessionID     uuid.UUID
	ImportUnitID        uuid.UUID
	MappingFingerprint  string
	SourceFileKind      string
	SourceContentSHA256 string
	ParserProfileID     string
	ParserVersion       string
	LocatorKind         string
	Locator             string
	SourceRectA1        string
	SourceRowRef        int
	ClientTxnID         string
	FieldValues         []ImportFieldValue
	UnknownValues       []ImportUnknownValue
	SourceRowProvenance ImportSourceRowProvenance
}

type ImportOwnerCreateResponse struct {
	RecordID             uuid.UUID
	RowVersion           int64
	ChangeSetMutationRef string
	CreatedOrReused      string
	OwnerResultCode      string
	RowRefresh           map[string]any
}

func NormalizeImportScalar(viewSchemaID string, fieldKey string, raw string, emptyValuePolicy string) (ImportScalarValue, bool, error) {
	field, ok := viewschema.LookupField(viewSchemaID, fieldKey)
	if !ok || (!field.Writable && !field.CreateWritable) {
		return ImportScalarValue{}, false, fmt.Errorf("field %q is not import writable", fieldKey)
	}
	if field.ConflictResolutionClass == "collection_review" {
		return ImportScalarValue{}, false, fmt.Errorf("field %q requires collection owner import support", fieldKey)
	}
	if raw == "" {
		switch emptyValuePolicy {
		case "omit_field", "":
			return ImportScalarValue{}, false, nil
		case "use_null":
			if !field.Clearable {
				return ImportScalarValue{}, false, fmt.Errorf("field %q is not nullable", fieldKey)
			}
			return ImportScalarValue{Kind: "null"}, true, nil
		default:
			return ImportScalarValue{}, false, fmt.Errorf("unsupported empty value policy %q", emptyValuePolicy)
		}
	}
	if field.DirectScalarContractID != nil && *field.DirectScalarContractID == "timestamp_instant_v1" {
		utc, ok := fieldnorm.NormalizeTimestampInstant(raw)
		if !ok {
			return ImportScalarValue{}, false, fmt.Errorf("invalid timestamp for %q", fieldKey)
		}
		return ImportScalarValue{Kind: "timestamp", Timestamp: &utc}, true, nil
	}
	if field.DirectReferenceContractID != nil || isUUIDImportField(fieldKey, field) {
		parsed, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil {
			return ImportScalarValue{}, false, fmt.Errorf("invalid uuid for %q", fieldKey)
		}
		if field.DirectReferenceContractID != nil && parsed.String() != raw {
			return ImportScalarValue{}, false, fmt.Errorf("invalid exact uuid for %q", fieldKey)
		}
		return ImportScalarValue{Kind: "uuid", UUID: &parsed}, true, nil
	}
	if field.ReadKind == "number" {
		parsed, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		if err != nil {
			return ImportScalarValue{}, false, fmt.Errorf("invalid integer for %q", fieldKey)
		}
		return ImportScalarValue{Kind: "number", Number: &parsed}, true, nil
	}
	if field.ReadKind == "boolean" {
		parsed, err := strconv.ParseBool(strings.TrimSpace(raw))
		if err != nil {
			return ImportScalarValue{}, false, fmt.Errorf("invalid boolean for %q", fieldKey)
		}
		return ImportScalarValue{Kind: "bool", Bool: &parsed}, true, nil
	}
	normalized, ok := normalizeStringContract(field, raw)
	if !ok {
		return ImportScalarValue{}, false, fmt.Errorf("invalid text for %q", fieldKey)
	}
	return ImportScalarValue{Kind: "text", Text: &normalized}, true, nil
}

func normalizeStringContract(field viewschema.Field, raw string) (string, bool) {
	if field.StringContractID != nil && *field.StringContractID == "multiline_body_v1" {
		return fieldnorm.NormalizeNote(raw)
	}
	return fieldnorm.NormalizeLine(raw)
}

func isUUIDImportField(fieldKey string, field viewschema.Field) bool {
	if field.ReadKind == "uuid" {
		return true
	}
	return strings.HasSuffix(fieldKey, "_user_id") || strings.HasSuffix(fieldKey, "_party_id") || strings.HasSuffix(fieldKey, "_record_id") || strings.HasSuffix(fieldKey, "_ref")
}
