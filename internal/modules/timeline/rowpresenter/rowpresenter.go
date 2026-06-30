package rowpresenter

import (
	"time"

	"github.com/google/uuid"
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
		"timeline.date_entered_text":        map[string]any{"value": derefString(record.DateEnteredText)},
		"timeline.analyst_text":             map[string]any{"value": derefString(record.AnalystText)},
		"timeline.mitre_stage_text":         map[string]any{"value": derefString(record.MitreStageText)},
		"timeline.device_object_text":       map[string]any{"value": derefString(record.DeviceObjectText)},
		"timeline.ip_address_text":          map[string]any{"value": derefString(record.IPAddressText)},
		"timeline.activity_utc_text":        map[string]any{"value": derefString(record.ActivityUTCText)},
		"timeline.activity_local_text":      map[string]any{"value": derefString(record.ActivityLocalText)},
		"timeline.raw_activity_text":        map[string]any{"value": derefString(record.RawActivityText)},
		"timeline.activity_synopsis_text":   map[string]any{"value": derefString(record.ActivitySynopsisText)},
		"timeline.data_source_text":         map[string]any{"value": derefString(record.DataSourceText)},
		"timeline.host_refs":                map[string]any{"value": collectionValue(true, record.HostRefs)},
		"timeline.identity_refs":            map[string]any{"value": collectionValue(true, record.IdentityRefs)},
		"timeline.attached_evidence_ids":    map[string]any{"value": collectionValue(false, record.AttachedEvidence)},
		"timeline.evidence_count":           map[string]any{"value": record.EvidenceCount},
		"timeline.tags":                     map[string]any{"value": collectionValue(false, record.Tags)},
		"timeline.edited_at":                map[string]any{"value": formatTimestamp(record.EditedAt)},
		"timeline.recorded_at":              map[string]any{"value": formatTimestamp(record.RecordedAt)},
		"timeline.activity_sort_ts":         map[string]any{"value": formatTimestampPointer(record.ActivitySortTS)},
		"timeline.date_entered_sort_day":    map[string]any{"value": formatDatePointer(record.DateEnteredSortDay)},
		"timeline.activity_time_pair_state": map[string]any{"value": record.ActivityTimePairState},
		"timeline.capture_state":            map[string]any{"value": record.CaptureState},
		"timeline.replacement_record_id":    map[string]any{"value": formatUUIDPointer(record.ReplacementRecordID)},
		"timeline.has_evidence":             map[string]any{"value": record.HasEvidence},
		"timeline.has_unresolved_mentions":  map[string]any{"value": record.HasUnresolvedMentions},
	}

	row := map[string]any{
		"record_id":   record.RecordID.String(),
		"row_version": record.RowVersion,
		"cells":       cells,
	}
	row["group_values"] = map[string]any{
		"timeline.date_entered_sort_day":    formatDatePointer(record.DateEnteredSortDay),
		"timeline.activity_time_pair_state": record.ActivityTimePairState,
		"timeline.capture_state":            record.CaptureState,
		"timeline.has_evidence":             record.HasEvidence,
		"timeline.has_unresolved_mentions":  record.HasUnresolvedMentions,
	}
	return row
}

func collectionValue(ordered bool, items []map[string]any) map[string]any {
	if items == nil {
		items = []map[string]any{}
	}
	return map[string]any{
		"kind":    "collection_value_v1",
		"ordered": ordered,
		"items":   items,
	}
}

func formatTimestamp(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func formatTimestampPointer(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func formatDatePointer(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format("2006-01-02")
}

func formatUUIDPointer(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func derefString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
