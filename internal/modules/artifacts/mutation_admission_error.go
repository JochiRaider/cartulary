package artifacts

type AdmissionReasonCode string

const (
	admissionRequestNotObject              AdmissionReasonCode = "request_not_object"
	admissionUnknownViewSchema             AdmissionReasonCode = "unknown_view_schema"
	admissionUnsupportedViewSchema         AdmissionReasonCode = "unsupported_view_schema"
	admissionInvalidViewSchemaID           AdmissionReasonCode = "invalid_view_schema_id"
	admissionUnknownField                  AdmissionReasonCode = "unknown_field"
	admissionMissingRequiredField          AdmissionReasonCode = "missing_required_field"
	admissionInvalidValue                  AdmissionReasonCode = "invalid_value"
	admissionMissingMinimumCreateSignal    AdmissionReasonCode = "missing_minimum_create_signal"
	admissionInvalidBaseRowVersion         AdmissionReasonCode = "invalid_base_row_version"
	admissionEmptyChanges                  AdmissionReasonCode = "empty_changes"
	admissionChangeCountExceeded           AdmissionReasonCode = "change_count_exceeded"
	admissionDuplicateFieldKey             AdmissionReasonCode = "duplicate_field_key"
	admissionInvalidChange                 AdmissionReasonCode = "invalid_change"
	admissionMissingFieldKey               AdmissionReasonCode = "missing_field_key"
	admissionUnsupportedFieldKey           AdmissionReasonCode = "unsupported_field_key"
	admissionFieldNotNullable              AdmissionReasonCode = "field_not_nullable"
	admissionForbiddenField                AdmissionReasonCode = "forbidden_field"
	admissionEmptyCollectionActions        AdmissionReasonCode = "empty_collection_actions"
	admissionCollectionActionCountExceeded AdmissionReasonCode = "collection_action_count_exceeded"
)

// AdmissionError is an immutable owner error. Workbook maps these facts to its
// public MutationFailure and transport response.
type AdmissionError struct {
	reason          AdmissionReasonCode
	field           string
	collectionField string
	requestedCount  int
	maximumCount    int
	hasLimit        bool
}

func (err *AdmissionError) Error() string { return "artifacts: invalid mutation admission" }

func (err *AdmissionError) ReasonCode() AdmissionReasonCode {
	if err == nil {
		return ""
	}
	return err.reason
}

func (err *AdmissionError) Field() (string, bool) {
	if err == nil || err.field == "" {
		return "", false
	}
	return err.field, true
}

func (err *AdmissionError) CollectionField() (string, bool) {
	if err == nil || err.collectionField == "" {
		return "", false
	}
	return err.collectionField, true
}

func (err *AdmissionError) RequestedCount() (int, bool) {
	if err == nil || !err.hasLimit {
		return 0, false
	}
	return err.requestedCount, true
}

func (err *AdmissionError) MaximumCount() (int, bool) {
	if err == nil || !err.hasLimit {
		return 0, false
	}
	return err.maximumCount, true
}

func newAdmissionError(field string, reason AdmissionReasonCode) *AdmissionError {
	return &AdmissionError{field: field, reason: reason}
}

func newAdmissionLimitError(field string, collectionField string, reason AdmissionReasonCode, requested int, maximum int) *AdmissionError {
	return &AdmissionError{
		field:           field,
		collectionField: collectionField,
		reason:          reason,
		requestedCount:  requested,
		maximumCount:    maximum,
		hasLimit:        true,
	}
}

func knownAdmissionReason(reason string) AdmissionReasonCode {
	candidate := AdmissionReasonCode(reason)
	switch candidate {
	case admissionRequestNotObject,
		admissionUnknownViewSchema,
		admissionUnsupportedViewSchema,
		admissionInvalidViewSchemaID,
		admissionUnknownField,
		admissionMissingRequiredField,
		admissionInvalidValue,
		admissionMissingMinimumCreateSignal,
		admissionInvalidBaseRowVersion,
		admissionEmptyChanges,
		admissionChangeCountExceeded,
		admissionDuplicateFieldKey,
		admissionInvalidChange,
		admissionMissingFieldKey,
		admissionUnsupportedFieldKey,
		admissionFieldNotNullable,
		admissionForbiddenField,
		admissionEmptyCollectionActions,
		admissionCollectionActionCountExceeded:
		return candidate
	default:
		return admissionInvalidValue
	}
}
