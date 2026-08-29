package evidence

type AdmissionReasonCode string

const (
	admissionRequestNotObject      AdmissionReasonCode = "request_not_object"
	admissionUnknownViewSchema     AdmissionReasonCode = "unknown_view_schema"
	admissionUnknownField          AdmissionReasonCode = "unknown_field"
	admissionMissingRequiredField  AdmissionReasonCode = "missing_required_field"
	admissionInvalidValue          AdmissionReasonCode = "invalid_value"
	admissionInvalidViewSchemaID   AdmissionReasonCode = "invalid_view_schema_id"
	admissionInvalidBaseRowVersion AdmissionReasonCode = "invalid_base_row_version"
	admissionEmptyChanges          AdmissionReasonCode = "empty_changes"
	admissionChangeCountExceeded   AdmissionReasonCode = "change_count_exceeded"
	admissionDuplicateFieldKey     AdmissionReasonCode = "duplicate_field_key"
	admissionInvalidChange         AdmissionReasonCode = "invalid_change"
	admissionMissingFieldKey       AdmissionReasonCode = "missing_field_key"
	admissionUnsupportedFieldKey   AdmissionReasonCode = "unsupported_field_key"
	admissionFieldNotNullable      AdmissionReasonCode = "field_not_nullable"
	admissionForbiddenField        AdmissionReasonCode = "forbidden_field"
)

// AdmissionFailure is an immutable, transport-neutral Evidence admission
// failure. Workbook maps these facts to its public failure vocabulary.
type AdmissionFailure struct {
	reason          AdmissionReasonCode
	field           string
	collectionField string
	requestedCount  int
	maximumCount    int
	hasLimit        bool
}

func (*AdmissionFailure) Error() string { return "evidence: invalid mutation admission" }

func (failure *AdmissionFailure) ReasonCode() AdmissionReasonCode {
	if failure == nil {
		return ""
	}
	return failure.reason
}

func (failure *AdmissionFailure) Field() (string, bool) {
	if failure == nil || failure.field == "" {
		return "", false
	}
	return failure.field, true
}

func (failure *AdmissionFailure) CollectionField() (string, bool) {
	if failure == nil || failure.collectionField == "" {
		return "", false
	}
	return failure.collectionField, true
}

func (failure *AdmissionFailure) RequestedCount() (int, bool) {
	if failure == nil || !failure.hasLimit {
		return 0, false
	}
	return failure.requestedCount, true
}

func (failure *AdmissionFailure) MaximumCount() (int, bool) {
	if failure == nil || !failure.hasLimit {
		return 0, false
	}
	return failure.maximumCount, true
}

func newAdmissionFailure(field string, reason AdmissionReasonCode) *AdmissionFailure {
	return &AdmissionFailure{field: field, reason: reason}
}

func newAdmissionLimitFailure(field string, reason AdmissionReasonCode, requested int, maximum int) *AdmissionFailure {
	return &AdmissionFailure{field: field, reason: reason, requestedCount: requested, maximumCount: maximum, hasLimit: true}
}
