package ownerfacade

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type ImportScalarKind string

const (
	ImportScalarNull            ImportScalarKind = "null"
	ImportScalarText            ImportScalarKind = "text"
	ImportScalarTimestamp       ImportScalarKind = "timestamp"
	ImportScalarUUID            ImportScalarKind = "uuid"
	ImportScalarNumber          ImportScalarKind = "number"
	ImportScalarBool            ImportScalarKind = "bool"
	ImportScalarCollectionToken ImportScalarKind = "collection_token"
)

type ImportScalarValue struct {
	kind            ImportScalarKind
	valid           bool
	text            string
	timestamp       time.Time
	uuid            uuid.UUID
	number          int64
	boolean         bool
	collectionToken ImportCollectionToken
}

type ImportCollectionToken struct {
	RawText        string
	NormalizedText string
}

func NewNullImportScalar() ImportScalarValue {
	return ImportScalarValue{kind: ImportScalarNull, valid: true}
}

func NewTextImportScalar(value string) ImportScalarValue {
	return ImportScalarValue{kind: ImportScalarText, valid: true, text: value}
}

func NewTimestampImportScalar(value time.Time) ImportScalarValue {
	return ImportScalarValue{kind: ImportScalarTimestamp, valid: true, timestamp: value}
}

func NewUUIDImportScalar(value uuid.UUID) ImportScalarValue {
	return ImportScalarValue{kind: ImportScalarUUID, valid: true, uuid: value}
}

func NewNumberImportScalar(value int64) ImportScalarValue {
	return ImportScalarValue{kind: ImportScalarNumber, valid: true, number: value}
}

func NewBoolImportScalar(value bool) ImportScalarValue {
	return ImportScalarValue{kind: ImportScalarBool, valid: true, boolean: value}
}

func NewCollectionTokenImportScalar(value ImportCollectionToken) ImportScalarValue {
	return ImportScalarValue{
		kind:            ImportScalarCollectionToken,
		valid:           true,
		collectionToken: value,
	}
}

func (v ImportScalarValue) Kind() ImportScalarKind {
	return v.kind
}

func (v ImportScalarValue) IsValid() bool {
	if !v.valid {
		return false
	}
	switch v.kind {
	case ImportScalarNull,
		ImportScalarText,
		ImportScalarTimestamp,
		ImportScalarUUID,
		ImportScalarNumber,
		ImportScalarBool,
		ImportScalarCollectionToken:
		return true
	default:
		return false
	}
}

func (v ImportScalarValue) Text() (string, bool) {
	return v.text, v.IsValid() && v.kind == ImportScalarText
}

func (v ImportScalarValue) Timestamp() (time.Time, bool) {
	return v.timestamp, v.IsValid() && v.kind == ImportScalarTimestamp
}

func (v ImportScalarValue) UUID() (uuid.UUID, bool) {
	return v.uuid, v.IsValid() && v.kind == ImportScalarUUID
}

func (v ImportScalarValue) Number() (int64, bool) {
	return v.number, v.IsValid() && v.kind == ImportScalarNumber
}

func (v ImportScalarValue) Bool() (bool, bool) {
	return v.boolean, v.IsValid() && v.kind == ImportScalarBool
}

func (v ImportScalarValue) CollectionToken() (ImportCollectionToken, bool) {
	return v.collectionToken, v.IsValid() && v.kind == ImportScalarCollectionToken
}

var errInvalidImportFieldValues = errors.New("invalid import field values")

func IndexImportFieldValues(fields []ImportFieldValue) (map[string]ImportScalarValue, error) {
	values := make(map[string]ImportScalarValue, len(fields))
	for _, field := range fields {
		if strings.TrimSpace(field.FieldKey) == "" || !field.NormalizedValue.IsValid() {
			return nil, errInvalidImportFieldValues
		}
		if _, duplicate := values[field.FieldKey]; duplicate {
			return nil, errInvalidImportFieldValues
		}
		values[field.FieldKey] = field.NormalizedValue
	}
	return values, nil
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

const ImportOwnerCreateValidationFailed = "owner_create_validation_failed"

type ImportOwnerCreateError struct {
	OwnerCode            string
	ReasonCode           string
	Field                string
	Guard                string
	PartyReasonCode      string
	ConflictingFieldKeys []string
	Retryable            bool
	cause                error
}

func NewPartyMatchConflictError(
	partyReasonCode string,
	conflictingFieldKeys []string,
	cause error,
) *ImportOwnerCreateError {
	return &ImportOwnerCreateError{
		OwnerCode:            ImportOwnerCreateValidationFailed,
		ReasonCode:           "party_match_conflict",
		PartyReasonCode:      partyReasonCode,
		ConflictingFieldKeys: append([]string(nil), conflictingFieldKeys...),
		Retryable:            false,
		cause:                cause,
	}
}

func (e *ImportOwnerCreateError) Error() string {
	return "import owner create validation failed"
}

func (e *ImportOwnerCreateError) Unwrap() error {
	return e.cause
}

func NewImportOwnerCreateValidationError(
	reasonCode string,
	field string,
	guard string,
	cause error,
) *ImportOwnerCreateError {
	return &ImportOwnerCreateError{
		OwnerCode:  ImportOwnerCreateValidationFailed,
		ReasonCode: reasonCode,
		Field:      field,
		Guard:      guard,
		Retryable:  false,
		cause:      cause,
	}
}

func ImportOwnerCreateErrorDetail(err error) (map[string]any, bool) {
	var ownerErr *ImportOwnerCreateError
	if !errors.As(err, &ownerErr) ||
		ownerErr.OwnerCode != ImportOwnerCreateValidationFailed ||
		!validImportOwnerCreateReason(ownerErr.ReasonCode) ||
		ownerErr.Retryable {
		return nil, false
	}
	if ownerErr.ReasonCode == "party_match_conflict" && !validPartyMatchDetail(ownerErr) {
		return nil, false
	}
	safeDetails := map[string]any{"reason_code": ownerErr.ReasonCode}
	if ownerErr.Field != "" {
		safeDetails["field"] = ownerErr.Field
	}
	if ownerErr.Guard != "" {
		safeDetails["guard"] = ownerErr.Guard
	}
	if ownerErr.ReasonCode == "party_match_conflict" {
		safeDetails["party_reason_code"] = ownerErr.PartyReasonCode
		safeDetails["conflicting_field_keys"] = append([]string(nil), ownerErr.ConflictingFieldKeys...)
	}
	return map[string]any{
		"owner_code":   ownerErr.OwnerCode,
		"retryable":    ownerErr.Retryable,
		"safe_details": safeDetails,
	}, true
}

func validPartyMatchDetail(ownerErr *ImportOwnerCreateError) bool {
	switch ownerErr.PartyReasonCode {
	case "cross_key_exact_match", "exact_match_key_claimed":
	default:
		return false
	}
	fields := ownerErr.ConflictingFieldKeys
	if len(fields) == 0 || len(fields) > 2 || !slices.IsSorted(fields) {
		return false
	}
	previous := ""
	for _, field := range fields {
		if (field != "party.external_ref" && field != "party.primary_email") || field == previous {
			return false
		}
		previous = field
	}
	return true
}

func NormalizeImportScalar(viewSchemaID string, fieldKey string, raw string, emptyValuePolicy string) (ImportScalarValue, bool, error) {
	field, ok := viewschema.LookupField(viewSchemaID, fieldKey)
	if !ok || (!field.Writable && !field.CreateWritable) {
		return ImportScalarValue{}, false, NewImportOwnerCreateValidationError(
			"field_not_import_writable",
			fieldKey,
			"create_writable",
			fmt.Errorf("field %q is not import writable", fieldKey),
		)
	}
	if field.ConflictResolutionClass == "collection_review" {
		return ImportScalarValue{}, false, NewImportOwnerCreateValidationError(
			"collection_owner_support_required",
			fieldKey,
			"collection_review",
			fmt.Errorf("field %q requires collection owner import support", fieldKey),
		)
	}
	if raw == "" {
		switch emptyValuePolicy {
		case "omit_field", "":
			return ImportScalarValue{}, false, nil
		case "write_null":
			if !field.Clearable {
				return ImportScalarValue{}, false, NewImportOwnerCreateValidationError(
					"field_not_nullable",
					fieldKey,
					"clearable",
					fmt.Errorf("field %q is not nullable", fieldKey),
				)
			}
			return NewNullImportScalar(), true, nil
		default:
			return ImportScalarValue{}, false, NewImportOwnerCreateValidationError(
				"invalid_empty_value_policy",
				fieldKey,
				"empty_value_policy",
				fmt.Errorf("unsupported empty value policy %q", emptyValuePolicy),
			)
		}
	}
	if field.DirectScalarContractID != nil && *field.DirectScalarContractID == "timestamp_instant_v1" {
		utc, ok := fieldnorm.NormalizeTimestampInstant(raw)
		if !ok {
			return ImportScalarValue{}, false, importValueError(
				"invalid_timestamp",
				fieldKey,
				"timestamp_instant_v1",
			)
		}
		return NewTimestampImportScalar(utc), true, nil
	}
	if field.DirectReferenceContractID != nil || isUUIDImportField(fieldKey, field) {
		parsed, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil {
			return ImportScalarValue{}, false, importValueError(
				"invalid_uuid",
				fieldKey,
				"uuid",
			)
		}
		if field.DirectReferenceContractID != nil && parsed.String() != raw {
			return ImportScalarValue{}, false, importValueError(
				"invalid_exact_uuid",
				fieldKey,
				*field.DirectReferenceContractID,
			)
		}
		return NewUUIDImportScalar(parsed), true, nil
	}
	if field.ReadKind == "number" {
		parsed, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		if err != nil {
			return ImportScalarValue{}, false, importValueError(
				"invalid_integer",
				fieldKey,
				"number",
			)
		}
		return NewNumberImportScalar(parsed), true, nil
	}
	if field.ReadKind == "boolean" {
		parsed, err := strconv.ParseBool(strings.TrimSpace(raw))
		if err != nil {
			return ImportScalarValue{}, false, importValueError(
				"invalid_boolean",
				fieldKey,
				"boolean",
			)
		}
		return NewBoolImportScalar(parsed), true, nil
	}
	normalized, ok := normalizeStringContract(field, raw)
	if !ok {
		guard := "line_v1"
		if field.StringContractID != nil {
			guard = *field.StringContractID
		}
		return ImportScalarValue{}, false, importValueError("invalid_text", fieldKey, guard)
	}
	return NewTextImportScalar(normalized), true, nil
}

func importValueError(reasonCode string, fieldKey string, guard string) error {
	return NewImportOwnerCreateValidationError(
		reasonCode,
		fieldKey,
		guard,
		fmt.Errorf("invalid imported value for %q", fieldKey),
	)
}

func validImportOwnerCreateReason(reasonCode string) bool {
	switch reasonCode {
	case "field_not_import_writable",
		"collection_owner_support_required",
		"field_not_nullable",
		"invalid_empty_value_policy",
		"invalid_timestamp",
		"invalid_uuid",
		"invalid_exact_uuid",
		"invalid_integer",
		"invalid_boolean",
		"invalid_text",
		"party_match_conflict":
		return true
	default:
		return false
	}
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
