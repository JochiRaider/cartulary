package rowpresenter

import (
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/valuecodec"
)

type Record struct {
	RecordID              uuid.UUID
	IncidentID            uuid.UUID
	RowVersion            int64
	DateEnteredText       *string
	AnalystText           *string
	MitreStageText        *string
	DeviceObjectText      *string
	IPAddressText         *string
	ActivityUTCText       *string
	ActivityLocalText     *string
	RawActivityText       *string
	ActivitySynopsisText  *string
	DataSourceText        *string
	RecordedAt            time.Time
	EditedAt              time.Time
	ActivitySortTS        *time.Time
	DateEnteredSortDay    *time.Time
	ActivityTimePairState string
	CaptureState          string
	ReplacementRecordID   *uuid.UUID
	EvidenceCount         int
	HasEvidence           bool
	HasUnresolvedMentions bool
	HostRefs              []map[string]any
	IdentityRefs          []map[string]any
	AttachedEvidence      []map[string]any
	Tags                  []map[string]any
}

func BuildRow(record Record) map[string]any {
	cells := map[string]any{
		"timeline.date_entered_text":        map[string]any{"value": valuecodec.OptionalString(record.DateEnteredText)},
		"timeline.analyst_text":             map[string]any{"value": valuecodec.OptionalString(record.AnalystText)},
		"timeline.mitre_stage_text":         map[string]any{"value": valuecodec.OptionalString(record.MitreStageText)},
		"timeline.device_object_text":       map[string]any{"value": valuecodec.OptionalString(record.DeviceObjectText)},
		"timeline.ip_address_text":          map[string]any{"value": valuecodec.OptionalString(record.IPAddressText)},
		"timeline.activity_utc_text":        map[string]any{"value": valuecodec.OptionalString(record.ActivityUTCText)},
		"timeline.activity_local_text":      map[string]any{"value": valuecodec.OptionalString(record.ActivityLocalText)},
		"timeline.raw_activity_text":        map[string]any{"value": valuecodec.OptionalString(record.RawActivityText)},
		"timeline.activity_synopsis_text":   map[string]any{"value": valuecodec.OptionalString(record.ActivitySynopsisText)},
		"timeline.data_source_text":         map[string]any{"value": valuecodec.OptionalString(record.DataSourceText)},
		"timeline.host_refs":                map[string]any{"value": valuecodec.Collection(true, record.HostRefs)},
		"timeline.identity_refs":            map[string]any{"value": valuecodec.Collection(true, record.IdentityRefs)},
		"timeline.attached_evidence_ids":    map[string]any{"value": valuecodec.Collection(false, record.AttachedEvidence)},
		"timeline.evidence_count":           map[string]any{"value": record.EvidenceCount},
		"timeline.tags":                     map[string]any{"value": valuecodec.Collection(false, record.Tags)},
		"timeline.edited_at":                map[string]any{"value": valuecodec.Timestamp(record.EditedAt)},
		"timeline.recorded_at":              map[string]any{"value": valuecodec.Timestamp(record.RecordedAt)},
		"timeline.activity_sort_ts":         map[string]any{"value": valuecodec.OptionalTimestamp(record.ActivitySortTS)},
		"timeline.date_entered_sort_day":    map[string]any{"value": valuecodec.OptionalDate(record.DateEnteredSortDay)},
		"timeline.activity_time_pair_state": map[string]any{"value": record.ActivityTimePairState},
		"timeline.capture_state":            map[string]any{"value": record.CaptureState},
		"timeline.replacement_record_id":    map[string]any{"value": valuecodec.OptionalUUID(record.ReplacementRecordID)},
		"timeline.has_evidence":             map[string]any{"value": record.HasEvidence},
		"timeline.has_unresolved_mentions":  map[string]any{"value": record.HasUnresolvedMentions},
	}

	row := map[string]any{
		"record_id":   record.RecordID.String(),
		"row_version": record.RowVersion,
		"cells":       cells,
	}
	row["group_values"] = map[string]any{
		"timeline.date_entered_sort_day":    valuecodec.OptionalDate(record.DateEnteredSortDay),
		"timeline.activity_time_pair_state": record.ActivityTimePairState,
		"timeline.capture_state":            record.CaptureState,
		"timeline.has_evidence":             record.HasEvidence,
		"timeline.has_unresolved_mentions":  record.HasUnresolvedMentions,
	}
	return row
}
