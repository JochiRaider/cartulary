package mutationadmission

// ReasonCode is the closed Entities mutation-admission failure vocabulary.
type ReasonCode string

const (
	ReasonAtLeastOneValueRequired ReasonCode = "at_least_one_value_required"
	ReasonChangeCountExceeded     ReasonCode = "change_count_exceeded"
	ReasonDuplicateFieldKey       ReasonCode = "duplicate_field_key"
	ReasonEmptyChanges            ReasonCode = "empty_changes"
	ReasonFieldForbidden          ReasonCode = "field_forbidden"
	ReasonFieldNotNullable        ReasonCode = "field_not_nullable"
	ReasonForbiddenField          ReasonCode = "forbidden_field"
	ReasonInvalidBaseRowVersion   ReasonCode = "invalid_base_row_version"
	ReasonInvalidChange           ReasonCode = "invalid_change"
	ReasonInvalidValue            ReasonCode = "invalid_value"
	ReasonInvalidViewSchemaID     ReasonCode = "invalid_view_schema_id"
	ReasonMissingFieldKey         ReasonCode = "missing_field_key"
	ReasonMissingRequiredField    ReasonCode = "missing_required_field"
	ReasonReadonlyField           ReasonCode = "readonly_field"
	ReasonRequestNotObject        ReasonCode = "request_not_object"
	ReasonUnknownField            ReasonCode = "unknown_field"
	ReasonUnknownViewSchema       ReasonCode = "unknown_view_schema"
	ReasonUnsupportedFieldKey     ReasonCode = "unsupported_field_key"
	ReasonUnsupportedViewSchema   ReasonCode = "unsupported_view_schema"
)

// Failure contains only source-owner semantic admission facts. Wire status,
// code, message, and detail-map construction belong to the consuming boundary.
type Failure struct {
	reason          ReasonCode
	field           string
	collectionField string
	requestedCount  int
	maximumCount    int
	hasCounts       bool
}

func New(field string, reason ReasonCode) *Failure {
	requireKnownReason(reason)
	return &Failure{field: field, reason: reason}
}

func NewLimit(
	field string,
	reason ReasonCode,
	requestedCount int,
	maximumCount int,
	collectionField string,
) *Failure {
	requireKnownReason(reason)
	return &Failure{
		reason:          reason,
		field:           field,
		collectionField: collectionField,
		requestedCount:  requestedCount,
		maximumCount:    maximumCount,
		hasCounts:       true,
	}
}

func (failure *Failure) Error() string {
	return "invalid mutation payload"
}

func (failure *Failure) ReasonCode() ReasonCode {
	if failure == nil {
		return ""
	}
	return failure.reason
}

func (failure *Failure) Field() (string, bool) {
	if failure == nil || failure.field == "" {
		return "", false
	}
	return failure.field, true
}

func (failure *Failure) CollectionField() (string, bool) {
	if failure == nil || failure.collectionField == "" {
		return "", false
	}
	return failure.collectionField, true
}

func (failure *Failure) RequestedCount() (int, bool) {
	if failure == nil || !failure.hasCounts {
		return 0, false
	}
	return failure.requestedCount, true
}

func (failure *Failure) MaximumCount() (int, bool) {
	if failure == nil || !failure.hasCounts {
		return 0, false
	}
	return failure.maximumCount, true
}

func requireKnownReason(reason ReasonCode) {
	switch reason {
	case ReasonAtLeastOneValueRequired,
		ReasonChangeCountExceeded,
		ReasonDuplicateFieldKey,
		ReasonEmptyChanges,
		ReasonFieldForbidden,
		ReasonFieldNotNullable,
		ReasonForbiddenField,
		ReasonInvalidBaseRowVersion,
		ReasonInvalidChange,
		ReasonInvalidValue,
		ReasonInvalidViewSchemaID,
		ReasonMissingFieldKey,
		ReasonMissingRequiredField,
		ReasonReadonlyField,
		ReasonRequestNotObject,
		ReasonUnknownField,
		ReasonUnknownViewSchema,
		ReasonUnsupportedFieldKey,
		ReasonUnsupportedViewSchema:
		return
	default:
		panic("unknown Entities mutation admission reason")
	}
}
