package artifacts

import (
	"crypto/sha256"
	"time"

	"github.com/google/uuid"
)

const (
	maxMutationPatchChanges      = 32
	maxMutationCollectionActions = 64
)

// CreateAdmission is an immutable, owner-validated Artifact create request.
// Its zero value is not admitted.
type CreateAdmission struct {
	request  createRequest
	hash     [sha256.Size]byte
	admitted bool
}

func (admission CreateAdmission) ClientTxnID() string  { return admission.request.ClientTxnID }
func (admission CreateAdmission) ViewSchemaID() string { return admission.request.ViewSchemaID }

func (admission CreateAdmission) valid() bool {
	return admission.admitted && admission.request.ViewSchemaID != "" && admission.request.ClientTxnID != ""
}

func (admission CreateAdmission) requestValue() createRequest {
	return cloneCreateRequest(admission.request)
}

func (admission CreateAdmission) requestHash() []byte {
	return append([]byte(nil), admission.hash[:]...)
}

// PatchAdmission is an immutable, owner-validated Artifact patch request. Its
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
	return admission.admitted && admission.request.ViewSchemaID != "" && admission.request.ClientTxnID != "" && admission.request.BaseRowVersion > 0 && len(admission.request.Changes) > 0
}

func (admission PatchAdmission) requestValue() patchRequest {
	return clonePatchRequest(admission.request)
}

func (admission PatchAdmission) requestHash() []byte {
	return append([]byte(nil), admission.hash[:]...)
}

// ContextualNoteAdmission is an immutable, owner-validated contextual note
// request. The source record is execution context and is bound into the hash
// only when the command is executed.
type ContextualNoteAdmission struct {
	request  createRequest
	admitted bool
}

func (admission ContextualNoteAdmission) ClientTxnID() string { return admission.request.ClientTxnID }

func (admission ContextualNoteAdmission) valid() bool {
	return admission.admitted && admission.request.ViewSchemaID == NotesViewSchemaID && admission.request.ClientTxnID != ""
}

func (admission ContextualNoteAdmission) requestValue() createRequest {
	return cloneCreateRequest(admission.request)
}

func (admission ContextualNoteAdmission) requestHash(sourceRecordID uuid.UUID) []byte {
	return contextualNoteAdmissionHash(sourceRecordID, admission.request)
}

// ConflictAdmissionContext is the complete validated token context retained by
// an Artifact conflict admission. Values are copied into the opaque admission.
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

// ConflictResolveAdmission is an immutable, owner-validated conflict
// resolution request and its complete token binding. Its zero value is not
// admitted.
type ConflictResolveAdmission struct {
	request  conflictResolveRequest
	context  ConflictAdmissionContext
	hash     [sha256.Size]byte
	admitted bool
}

type conflictResolveRequest struct {
	ConflictToken  string
	ResolutionKind string
	ClientTxnID    string
	Patch          *patchRequest
	CanonicalValue any
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
	return PatchAdmission{
		request:  clonePatchRequest(*admission.request.Patch),
		hash:     admission.hash,
		admitted: true,
	}
}

func cloneCreateRequest(request createRequest) createRequest {
	return createRequest{
		ViewSchemaID: request.ViewSchemaID,
		ClientTxnID:  request.ClientTxnID,
		Values:       cloneArtifactValues(request.Values),
		Collections:  cloneArtifactCollections(request.Collections),
	}
}

func clonePatchRequest(request patchRequest) patchRequest {
	cloned := patchRequest{
		ViewSchemaID:   request.ViewSchemaID,
		BaseRowVersion: request.BaseRowVersion,
		ClientTxnID:    request.ClientTxnID,
		Changes:        make([]patchChange, 0, len(request.Changes)),
	}
	for _, change := range request.Changes {
		clonedChange := patchChange{FieldKey: change.FieldKey}
		if change.Value != nil {
			value := cloneArtifactValue(*change.Value)
			clonedChange.Value = &value
			clonedChange.CanonicalValue = canonicalArtifactValue(value)
		}
		if change.Collection != nil {
			collection := cloneArtifactCollection(*change.Collection)
			clonedChange.Collection = &collection
			clonedChange.CanonicalValue = canonicalArtifactCollectionPayload(collection)
		}
		cloned.Changes = append(cloned.Changes, clonedChange)
	}
	return cloned
}

func cloneConflictResolveRequest(request conflictResolveRequest) conflictResolveRequest {
	cloned := conflictResolveRequest{
		ConflictToken:  request.ConflictToken,
		ResolutionKind: request.ResolutionKind,
		ClientTxnID:    request.ClientTxnID,
	}
	if request.Patch != nil {
		patch := clonePatchRequest(*request.Patch)
		cloned.Patch = &patch
		cloned.CanonicalValue = patch.Changes[0].CanonicalValue
	}
	return cloned
}

func cloneArtifactValues(values map[string]fieldValue) map[string]fieldValue {
	cloned := make(map[string]fieldValue, len(values))
	for key, value := range values {
		cloned[key] = cloneArtifactValue(value)
	}
	return cloned
}

func cloneArtifactValue(value fieldValue) fieldValue {
	cloned := fieldValue{}
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

func cloneArtifactCollections(collections map[string]collectionActionPayload) map[string]collectionActionPayload {
	cloned := make(map[string]collectionActionPayload, len(collections))
	for key, collection := range collections {
		cloned[key] = cloneArtifactCollection(collection)
	}
	return cloned
}

func cloneArtifactCollection(payload collectionActionPayload) collectionActionPayload {
	cloned := collectionActionPayload{Actions: make([]collectionAction, 0, len(payload.Actions))}
	for _, action := range payload.Actions {
		clonedAction := action
		if action.LinkedRecordID != nil {
			copyValue := *action.LinkedRecordID
			clonedAction.LinkedRecordID = &copyValue
		}
		if action.PartyID != nil {
			copyValue := *action.PartyID
			clonedAction.PartyID = &copyValue
		}
		cloned.Actions = append(cloned.Actions, clonedAction)
	}
	return cloned
}
