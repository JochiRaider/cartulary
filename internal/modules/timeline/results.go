package timeline

import (
	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

type MutationResult struct {
	Payload          map[string]any
	StatusCode       int
	Replayed         bool
	IncidentID       uuid.UUID
	RecordID         uuid.UUID
	ChangeSetID      uuid.UUID
	ClientTxnID      string
	RowVersion       int64
	ChangedFieldKeys []string
	Row              workbookprojection.DerivedRecord
}

type ClipboardPasteResult struct {
	Payload     map[string]any
	StatusCode  int
	Replayed    bool
	IncidentID  uuid.UUID
	ChangeSetID uuid.UUID
	ClientTxnID string
	Rows        []ClipboardPasteRowResult
}

type ClipboardPasteRowResult struct {
	RecordID         uuid.UUID
	RowVersion       int64
	ChangedFieldKeys []string
	Row              map[string]any
}
