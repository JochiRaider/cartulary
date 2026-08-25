package tasksdecisions

// AdmissionFailure is Tasks/Decisions' closed, transport-neutral mutation
// admission failure. Workbook owns the translation to public HTTP errors.
type AdmissionFailure struct {
	field              string
	reasonCode         string
	collectionFieldKey string
	requestedCount     int
	maxCount           int
	hasCountLimit      bool
}

func (*AdmissionFailure) Error() string {
	return "tasksdecisions: invalid mutation admission"
}

func (failure *AdmissionFailure) Field() string {
	if failure == nil {
		return ""
	}
	return failure.field
}

func (failure *AdmissionFailure) ReasonCode() string {
	if failure == nil {
		return ""
	}
	return failure.reasonCode
}

func (failure *AdmissionFailure) CollectionFieldKey() (string, bool) {
	if failure == nil || failure.collectionFieldKey == "" {
		return "", false
	}
	return failure.collectionFieldKey, true
}

func (failure *AdmissionFailure) CountLimit() (requestedCount int, maxCount int, ok bool) {
	if failure == nil || !failure.hasCountLimit {
		return 0, 0, false
	}
	return failure.requestedCount, failure.maxCount, true
}

func invalidAdmission(field string, reasonCode string) *AdmissionFailure {
	return &AdmissionFailure{field: field, reasonCode: reasonCode}
}

func invalidCollectionAdmission(field string, reasonCode string, collectionFieldKey string) *AdmissionFailure {
	return &AdmissionFailure{
		field: field, reasonCode: reasonCode, collectionFieldKey: collectionFieldKey,
	}
}

func invalidCountAdmission(
	field string,
	reasonCode string,
	requestedCount int,
	maxCount int,
	collectionFieldKey string,
) *AdmissionFailure {
	return &AdmissionFailure{
		field: field, reasonCode: reasonCode, collectionFieldKey: collectionFieldKey,
		requestedCount: requestedCount, maxCount: maxCount, hasCountLimit: true,
	}
}
