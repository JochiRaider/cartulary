package evidence

import (
	"crypto/sha256"
	"time"

	"github.com/google/uuid"
)

// CreateAdmission is an immutable, owner-validated Evidence create request.
// Its zero value is not admitted.
type CreateAdmission struct {
	request  createRequest
	hash     [sha256.Size]byte
	admitted bool
}

func (admission CreateAdmission) ClientTxnID() string  { return admission.request.ClientTxnID }
func (admission CreateAdmission) ViewSchemaID() string { return admission.request.ViewSchemaID }

func (admission CreateAdmission) valid() bool {
	return admission.admitted && admission.request.ViewSchemaID == ViewSchemaID && admission.request.ClientTxnID != ""
}

func (admission CreateAdmission) requestValue() createRequest {
	return cloneCreateRequest(admission.request)
}
func (admission CreateAdmission) requestHash() []byte {
	return append([]byte(nil), admission.hash[:]...)
}

// PatchAdmission is an immutable, owner-validated Evidence patch request. Its
// zero value is not admitted.
type PatchAdmission struct {
	request  patchRequest
	hash     [sha256.Size]byte
	admitted bool
}

func (admission PatchAdmission) ClientTxnID() string   { return admission.request.ClientTxnID }
func (admission PatchAdmission) ViewSchemaID() string  { return admission.request.ViewSchemaID }
func (admission PatchAdmission) BaseRowVersion() int64 { return admission.request.BaseRowVersion }

func (admission PatchAdmission) valid() bool {
	return admission.admitted && admission.request.ViewSchemaID == ViewSchemaID &&
		admission.request.ClientTxnID != "" && admission.request.BaseRowVersion > 0 && len(admission.request.Changes) > 0
}

func (admission PatchAdmission) requestValue() patchRequest {
	return clonePatchRequest(admission.request)
}
func (admission PatchAdmission) requestHash() []byte {
	return append([]byte(nil), admission.hash[:]...)
}

// ConflictAdmissionContext is the complete validated token context retained
// by an Evidence conflict admission.
type ConflictAdmissionContext struct {
	Version                 int64
	RecordID                uuid.UUID
	ViewSchemaID            string
	RouteKey                string
	FieldKey                string
	ConflictResolutionClass string
	BaseRowVersion          int64
	CurrentRowVersion       int64
	OriginalRequestHash     string
	IssuedAt                time.Time
	ExpiresAt               time.Time
}

// ConflictResolveAdmission is an immutable, owner-validated conflict request
// and complete token binding. Its zero value is not admitted.
type ConflictResolveAdmission struct {
	request  conflictResolveRequest
	context  ConflictAdmissionContext
	hash     [sha256.Size]byte
	admitted bool
}

func (admission ConflictResolveAdmission) ClientTxnID() string { return admission.request.ClientTxnID }

func (admission ConflictResolveAdmission) valid() bool {
	return admission.admitted && admission.request.ClientTxnID != "" && validConflictAdmissionContext(admission.context)
}

func (admission ConflictResolveAdmission) requestValue() conflictResolveRequest {
	return cloneConflictResolveRequest(admission.request)
}

func (admission ConflictResolveAdmission) contextValue() ConflictAdmissionContext {
	return admission.context
}
func (admission ConflictResolveAdmission) requestHash() []byte {
	return append([]byte(nil), admission.hash[:]...)
}

func (admission ConflictResolveAdmission) patchAdmission() PatchAdmission {
	if !admission.valid() || admission.request.Patch == nil {
		return PatchAdmission{}
	}
	return PatchAdmission{request: clonePatchRequest(*admission.request.Patch), hash: admission.hash, admitted: true}
}

func cloneCreateRequest(request createRequest) createRequest {
	cloned := createRequest{
		ViewSchemaID: request.ViewSchemaID,
		ClientTxnID:  request.ClientTxnID,
		Values:       cloneEvidenceValues(request.Values),
	}
	if request.InitialObjectBlobID != nil {
		value := *request.InitialObjectBlobID
		cloned.InitialObjectBlobID = &value
	}
	return cloned
}

func clonePatchRequest(request patchRequest) patchRequest {
	cloned := patchRequest{
		ViewSchemaID: request.ViewSchemaID, BaseRowVersion: request.BaseRowVersion,
		ClientTxnID: request.ClientTxnID, Changes: make([]patchChange, 0, len(request.Changes)),
	}
	for _, change := range request.Changes {
		clonedChange := patchChange{FieldKey: change.FieldKey, CanonicalValue: cloneAdmissionValue(change.CanonicalValue)}
		if change.Value != nil {
			value := cloneEvidenceValue(*change.Value)
			clonedChange.Value = &value
		}
		cloned.Changes = append(cloned.Changes, clonedChange)
	}
	return cloned
}

func cloneConflictResolveRequest(request conflictResolveRequest) conflictResolveRequest {
	cloned := conflictResolveRequest{
		ConflictToken: request.ConflictToken, ResolutionKind: request.ResolutionKind,
		ClientTxnID: request.ClientTxnID, CanonicalValue: cloneAdmissionValue(request.CanonicalValue),
	}
	if request.Patch != nil {
		patch := clonePatchRequest(*request.Patch)
		cloned.Patch = &patch
	}
	return cloned
}

func cloneEvidenceValues(values map[string]FieldValue) map[string]FieldValue {
	cloned := make(map[string]FieldValue, len(values))
	for key, value := range values {
		cloned[key] = cloneEvidenceValue(value)
	}
	return cloned
}

func cloneEvidenceValue(value FieldValue) FieldValue {
	cloned := FieldValue{}
	if value.Text != nil {
		copyValue := *value.Text
		cloned.Text = &copyValue
	}
	if value.Timestamp != nil {
		copyValue := *value.Timestamp
		cloned.Timestamp = &copyValue
	}
	if value.UUID != nil {
		copyValue := *value.UUID
		cloned.UUID = &copyValue
	}
	if value.Number != nil {
		copyValue := *value.Number
		cloned.Number = &copyValue
	}
	if value.Bool != nil {
		copyValue := *value.Bool
		cloned.Bool = &copyValue
	}
	return cloned
}

func cloneAdmissionValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		cloned := make(map[string]any, len(typed))
		for key, item := range typed {
			cloned[key] = cloneAdmissionValue(item)
		}
		return cloned
	case []any:
		cloned := make([]any, len(typed))
		for index, item := range typed {
			cloned[index] = cloneAdmissionValue(item)
		}
		return cloned
	default:
		return typed
	}
}
