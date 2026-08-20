package timeline

import "github.com/google/uuid"

const (
	TimelineViewSchemaID    = "cartulary.view.timeline.v2"
	timelineQueryRouteKey   = "timeline.query"
	createRouteKey          = "timeline.rows.create"
	patchRouteKey           = "timeline.records.patch"
	ConflictResolveRouteKey = "timeline.records.conflicts.resolve"
	conflictResolveRouteKey = ConflictResolveRouteKey
	reviewRouteKey          = "timeline.records.mark_reviewed"
	supersedeRouteKey       = "timeline.records.supersede"
)

type CreateRequest struct {
	ClientTxnID          string
	DateEnteredText      *string
	AnalystText          *string
	MitreStageText       *string
	DeviceObjectText     *string
	IPAddressText        *string
	ActivityUTCText      *string
	ActivityLocalText    *string
	RawActivityText      *string
	ActivitySynopsisText *string
	DataSourceText       *string
	HostRefs             *CollectionActionPayload
	IdentityRefs         *CollectionActionPayload
	Tags                 *CollectionActionPayload
	AttachedEvidence     *CollectionActionPayload
	RawCaptureColumns    []ClipboardRawImportColumn
}

type PatchRequest struct {
	ViewSchemaID    string
	BaseRowVersion  int64
	ClientTxnID     string
	CanonicalChange []PatchChange
}

type ConflictResolveRequest struct {
	ConflictToken  string
	ResolutionKind string
	ClientTxnID    string
	ResolvedChange *PatchChange
	CanonicalAny   any
}

type PatchChange struct {
	FieldKey      string
	TextValue     *string
	ActionPayload *CollectionActionPayload
	CanonicalAny  any
}

type CollectionActionPayload struct {
	Actions []CollectionAction
}

type CollectionAction struct {
	Op             string
	RawText        string
	NormalizedText string
	ItemRef        string
	ResolvedRecord *uuid.UUID
	LinkedRecordID *uuid.UUID
}

type ActionRequest struct {
	BaseRowVersion int64
	ClientTxnID    string
	Reason         *string
}

type SupersedeRequest struct {
	BaseRowVersion      int64
	ClientTxnID         string
	Reason              string
	ReplacementRecordID *uuid.UUID
}

type TimeConversionProfilePutRequest struct {
	BaseProfileVersion int64
	Enabled            bool
	LocalOffsetMinutes *int
	LocalLabel         *string
}
