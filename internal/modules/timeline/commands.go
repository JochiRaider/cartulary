package timeline

import (
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicttokens"
	"github.com/JochiRaider/cartulary/internal/modules/tabularingest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

const (
	OwnerBatchOperationClipboardPasteV1        = "clipboard_paste_v1"
	OwnerBatchOperationFillDownV1              = "fill_down_v1"
	OwnerBatchOperationMultiRowTagAssignmentV1 = "multi_row_tag_assignment_v1"
)

type OwnerBatchTargetV1 struct {
	Kind           string
	RecordID       uuid.UUID
	BaseRowVersion int64
}

type CreateRowCommand struct {
	Actor       authn.UserRecord
	IncidentID  uuid.UUID
	Request     CreateRequest
	RequestHash []byte
	RequestID   string
	Now         time.Time
}

type PatchRowCommand struct {
	Actor       authn.UserRecord
	RecordID    uuid.UUID
	Request     PatchRequest
	RequestHash []byte
	RequestID   string
	Now         time.Time
}

type ConflictResolveCommand struct {
	Actor       authn.UserRecord
	RecordID    uuid.UUID
	Claims      conflicttokens.ConflictTokenClaims
	Request     ConflictResolveRequest
	RequestHash []byte
	RequestID   string
	Now         time.Time
}

type ClipboardPasteCommand struct {
	Actor       authn.UserRecord
	IncidentID  uuid.UUID
	ClientTxnID string
	Plan        tabularingest.TabularRowPlanV1
	Targets     []OwnerBatchTargetV1
	RequestHash []byte
	RequestID   string
	Now         time.Time
}

type FillDownCommand struct {
	Actor       authn.UserRecord
	IncidentID  uuid.UUID
	ClientTxnID string
	FieldKey    string
	Value       string
	Targets     []OwnerBatchTargetV1
	RequestHash []byte
	RequestID   string
	Now         time.Time
}

type MultiRowTagAssignmentCommand struct {
	Actor         authn.UserRecord
	IncidentID    uuid.UUID
	ClientTxnID   string
	TagName       string
	NormalizedTag string
	Targets       []OwnerBatchTargetV1
	RequestHash   []byte
	RequestID     string
	Now           time.Time
}

type MarkReviewedCommand struct {
	Actor       authn.UserRecord
	RecordID    uuid.UUID
	Request     ActionRequest
	RequestHash []byte
	RequestID   string
	Now         time.Time
}

type SupersedeCommand struct {
	Actor       authn.UserRecord
	RecordID    uuid.UUID
	Request     SupersedeRequest
	RequestHash []byte
	RequestID   string
	Now         time.Time
}
