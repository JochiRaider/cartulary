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
	maxPatchChanges         = 32
	maxCollectionActions    = 64
)

var directWritableFieldKeys = map[string]struct{}{
	"timeline.date_entered_text":      {},
	"timeline.analyst_text":           {},
	"timeline.mitre_stage_text":       {},
	"timeline.device_object_text":     {},
	"timeline.ip_address_text":        {},
	"timeline.activity_utc_text":      {},
	"timeline.activity_local_text":    {},
	"timeline.raw_activity_text":      {},
	"timeline.activity_synopsis_text": {},
	"timeline.data_source_text":       {},
}

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
