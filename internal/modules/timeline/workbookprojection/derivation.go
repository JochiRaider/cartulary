package workbookprojection

import (
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/rowpresenter"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/timecontract"
)

type DerivedRecord struct {
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
	HostRefs              []MentionRef
	IdentityRefs          []MentionRef
	AttachedEvidence      []EvidenceRef
	Tags                  []TagRef
}

func Derive(snapshot sourcerepository.Snapshot, replacementRecordID *uuid.UUID) DerivedRecord {
	return DerivedRecord{
		RecordID:              snapshot.RecordID,
		IncidentID:            snapshot.IncidentID,
		RowVersion:            snapshot.RowVersion,
		DateEnteredText:       cloneStringPointer(snapshot.DateEnteredText),
		AnalystText:           cloneStringPointer(snapshot.AnalystText),
		MitreStageText:        cloneStringPointer(snapshot.MitreStageText),
		DeviceObjectText:      cloneStringPointer(snapshot.DeviceObjectText),
		IPAddressText:         cloneStringPointer(snapshot.IPAddressText),
		ActivityUTCText:       cloneStringPointer(snapshot.ActivityUTCText),
		ActivityLocalText:     cloneStringPointer(snapshot.ActivityLocalText),
		RawActivityText:       cloneStringPointer(snapshot.RawActivityText),
		ActivitySynopsisText:  cloneStringPointer(snapshot.ActivitySynopsisText),
		DataSourceText:        cloneStringPointer(snapshot.DataSourceText),
		RecordedAt:            snapshot.RecordedAt.UTC(),
		EditedAt:              snapshot.EditedAt.UTC(),
		ActivitySortTS:        deriveActivitySortTS(snapshot.ActivityUTCText, snapshot.ActivityLocalText),
		DateEnteredSortDay:    deriveDateEnteredSortDay(snapshot.DateEnteredText),
		ActivityTimePairState: snapshot.ActivityTimePairState,
		CaptureState:          snapshot.CaptureState,
		ReplacementRecordID:   cloneUUIDPointer(replacementRecordID),
		HostRefs:              []MentionRef{},
		IdentityRefs:          []MentionRef{},
		AttachedEvidence:      []EvidenceRef{},
		Tags:                  []TagRef{},
	}
}

func (record DerivedRecord) ProjectionInput() ProjectionInput {
	return ProjectionInput{
		RecordID:              record.RecordID,
		IncidentID:            record.IncidentID,
		RowVersion:            record.RowVersion,
		DateEnteredText:       record.DateEnteredText,
		AnalystText:           record.AnalystText,
		MitreStageText:        record.MitreStageText,
		DeviceObjectText:      record.DeviceObjectText,
		IPAddressText:         record.IPAddressText,
		ActivityUTCText:       record.ActivityUTCText,
		ActivityLocalText:     record.ActivityLocalText,
		RawActivityText:       record.RawActivityText,
		ActivitySynopsisText:  record.ActivitySynopsisText,
		DataSourceText:        record.DataSourceText,
		RecordedAt:            record.RecordedAt,
		EditedAt:              record.EditedAt,
		ActivitySortTS:        record.ActivitySortTS,
		DateEnteredSortDay:    record.DateEnteredSortDay,
		ActivityTimePairState: record.ActivityTimePairState,
		CaptureState:          record.CaptureState,
		ReplacementRecordID:   record.ReplacementRecordID,
		EvidenceCount:         record.EvidenceCount,
		HasEvidence:           record.HasEvidence,
		HasUnresolvedMentions: record.HasUnresolvedMentions,
		HostRefs:              append(make([]MentionRef, 0, len(record.HostRefs)), record.HostRefs...),
		IdentityRefs:          append(make([]MentionRef, 0, len(record.IdentityRefs)), record.IdentityRefs...),
		AttachedEvidence:      append(make([]EvidenceRef, 0, len(record.AttachedEvidence)), record.AttachedEvidence...),
		Tags:                  append(make([]TagRef, 0, len(record.Tags)), record.Tags...),
	}
}

func (record DerivedRecord) PresenterRecord() rowpresenter.Record {
	return rowpresenter.Record{
		RecordID:              record.RecordID,
		IncidentID:            record.IncidentID,
		RowVersion:            record.RowVersion,
		DateEnteredText:       record.DateEnteredText,
		AnalystText:           record.AnalystText,
		MitreStageText:        record.MitreStageText,
		DeviceObjectText:      record.DeviceObjectText,
		IPAddressText:         record.IPAddressText,
		ActivityUTCText:       record.ActivityUTCText,
		ActivityLocalText:     record.ActivityLocalText,
		RawActivityText:       record.RawActivityText,
		ActivitySynopsisText:  record.ActivitySynopsisText,
		DataSourceText:        record.DataSourceText,
		RecordedAt:            record.RecordedAt,
		EditedAt:              record.EditedAt,
		ActivitySortTS:        record.ActivitySortTS,
		DateEnteredSortDay:    record.DateEnteredSortDay,
		ActivityTimePairState: record.ActivityTimePairState,
		CaptureState:          record.CaptureState,
		ReplacementRecordID:   record.ReplacementRecordID,
		EvidenceCount:         record.EvidenceCount,
		HasEvidence:           record.HasEvidence,
		HasUnresolvedMentions: record.HasUnresolvedMentions,
		HostRefs:              mentionRefsToMaps(record.HostRefs),
		IdentityRefs:          mentionRefsToMaps(record.IdentityRefs),
		AttachedEvidence:      evidenceRefsToMaps(record.AttachedEvidence),
		Tags:                  tagRefsToMaps(record.Tags),
	}
}

func deriveActivitySortTS(utcText *string, localText *string) *time.Time {
	if parsed, ok := timecontract.ParseUTC(utcText); ok {
		return &parsed
	}
	if parsed, _, ok := timecontract.ParseLocalOffset(localText); ok {
		return &parsed
	}
	return nil
}

func deriveDateEnteredSortDay(text *string) *time.Time {
	if text == nil || *text == "" {
		return nil
	}
	if parsed, err := time.Parse("2006-01-02", *text); err == nil {
		day := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
		return &day
	}
	if parsed, ok := timecontract.ParseUTC(text); ok {
		day := time.Date(parsed.UTC().Year(), parsed.UTC().Month(), parsed.UTC().Day(), 0, 0, 0, 0, time.UTC)
		return &day
	}
	if parsed, _, ok := timecontract.ParseLocalOffset(text); ok {
		day := time.Date(parsed.UTC().Year(), parsed.UTC().Month(), parsed.UTC().Day(), 0, 0, 0, 0, time.UTC)
		return &day
	}
	return nil
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneUUIDPointer(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
