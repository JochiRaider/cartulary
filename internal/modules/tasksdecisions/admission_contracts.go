package tasksdecisions

import "github.com/google/uuid"

const (
	maxMutationPatchChanges      = 32
	maxMutationCollectionActions = 64
)

type ConflictClaims struct {
	RecordID          uuid.UUID
	ViewSchemaID      string
	FieldKey          string
	CurrentRowVersion int64
}

type ConflictResolveRequest struct {
	ConflictToken  string
	ResolutionKind string
	ClientTxnID    string
	Patch          *PatchRequest
	CanonicalValue any
}
