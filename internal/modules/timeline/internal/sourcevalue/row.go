package sourcevalue

import (
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	shape "github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mutationpolicy"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/timecontract"
)

// Row is the logical Timeline source; envelope mirrors belong to Records.
type Row struct {
	RecordID               uuid.UUID  `json:"record_id"`
	IncidentID             uuid.UUID  `json:"incident_id"`
	CaptureState           string     `json:"capture_state"`
	ReviewedByUserID       *uuid.UUID `json:"reviewed_by_user_id"`
	ReviewedAt             *time.Time `json:"reviewed_at"`
	SupersededByUserID     *uuid.UUID `json:"superseded_by_user_id"`
	SupersededAt           *time.Time `json:"superseded_at"`
	DateEnteredText        *string    `json:"date_entered_text"`
	AnalystText            *string    `json:"analyst_text"`
	MitreStageText         *string    `json:"mitre_stage_text"`
	DeviceObjectText       *string    `json:"device_object_text"`
	IPAddressText          *string    `json:"ip_address_text"`
	ActivityUTCText        *string    `json:"activity_utc_text"`
	ActivityLocalText      *string    `json:"activity_local_text"`
	RawActivityText        *string    `json:"raw_activity_text"`
	ActivitySynopsisText   *string    `json:"activity_synopsis_text"`
	DataSourceText         *string    `json:"data_source_text"`
	ActivityUTCGenerated   bool       `json:"activity_utc_generated"`
	ActivityLocalGenerated bool       `json:"activity_local_generated"`
	ActivityTimePairState  string     `json:"activity_time_pair_state"`
}

func Parse(raw map[string]any) (Row, bool) {
	var row Row
	if err := shape.ValidateSourceObject(raw, raw,
		shape.SourceFieldGroup{Names: "record_id incident_id", Kind: shape.SourceUUID},
		shape.SourceFieldGroup{Names: "capture_state activity_time_pair_state", Kind: shape.SourceString},
		shape.SourceFieldGroup{Names: "reviewed_by_user_id superseded_by_user_id", Kind: shape.SourceUUID, Nullable: true},
		shape.SourceFieldGroup{Names: "reviewed_at superseded_at", Kind: shape.SourceTimestamp, Nullable: true},
		shape.SourceFieldGroup{Names: "date_entered_text analyst_text mitre_stage_text device_object_text ip_address_text activity_utc_text activity_local_text raw_activity_text activity_synopsis_text data_source_text", Kind: shape.SourceString, Nullable: true},
		shape.SourceFieldGroup{Names: "activity_utc_generated activity_local_generated", Kind: shape.SourceBoolean},
	); err != nil {
		return row, false
	}
	row = ReadAdmittedFields(raw)
	if !row.VisibleTextLegal() {
		return Row{}, false
	}
	return row, true
}

// ReadAdmittedFields projects the logical fields after the caller has checked
// its complete closed scalar shape. Historical snapshots also carry envelope
// mirrors; Parse admits only the logical shape. Direct values preserve invalid
// UTF-8 for rejection rather than replacing it through JSON re-encoding.
func ReadAdmittedFields(raw map[string]any) Row {
	id := func(name string) *uuid.UUID {
		if raw[name] == nil {
			return nil
		}
		value := uuid.MustParse(raw[name].(string))
		return &value
	}
	stamp := func(name string) *time.Time {
		if raw[name] == nil {
			return nil
		}
		value, _ := time.Parse(time.RFC3339Nano, raw[name].(string))
		value = value.UTC()
		return &value
	}
	text := func(name string) *string {
		if raw[name] == nil {
			return nil
		}
		value := raw[name].(string)
		return &value
	}
	return Row{
		RecordID: *id("record_id"), IncidentID: *id("incident_id"), CaptureState: raw["capture_state"].(string),
		ReviewedByUserID: id("reviewed_by_user_id"), ReviewedAt: stamp("reviewed_at"), SupersededByUserID: id("superseded_by_user_id"), SupersededAt: stamp("superseded_at"),
		DateEnteredText: text("date_entered_text"), AnalystText: text("analyst_text"), MitreStageText: text("mitre_stage_text"), DeviceObjectText: text("device_object_text"), IPAddressText: text("ip_address_text"),
		ActivityUTCText: text("activity_utc_text"), ActivityLocalText: text("activity_local_text"), RawActivityText: text("raw_activity_text"), ActivitySynopsisText: text("activity_synopsis_text"), DataSourceText: text("data_source_text"),
		ActivityUTCGenerated: raw["activity_utc_generated"].(bool), ActivityLocalGenerated: raw["activity_local_generated"].(bool), ActivityTimePairState: raw["activity_time_pair_state"].(string),
	}
}

// NormalizeClocks changes only time.Time's zone representation after admission.
func (r *Row) NormalizeClocks() {
	if r.ReviewedAt != nil {
		value := r.ReviewedAt.UTC()
		r.ReviewedAt = &value
	}
	if r.SupersededAt != nil {
		value := r.SupersededAt.UTC()
		r.SupersededAt = &value
	}
}

func (r Row) VisibleTextLegal() bool {
	for _, value := range []*string{r.DateEnteredText, r.AnalystText, r.MitreStageText, r.DeviceObjectText, r.IPAddressText, r.ActivityUTCText, r.ActivityLocalText, r.RawActivityText, r.ActivitySynopsisText, r.DataSourceText} {
		if value != nil && (!utf8.ValidString(*value) || !mutationpolicy.IsValidVisibleText(*value)) {
			return false
		}
	}
	return true
}

func (r Row) LifecycleLegal() bool {
	reviewed, superseded := r.ReviewedAt != nil, r.SupersededAt != nil
	if reviewed != (r.ReviewedByUserID != nil) || superseded != (r.SupersededByUserID != nil) {
		return false
	}
	switch r.CaptureState {
	case "rough", "enriched":
		return !reviewed && !superseded
	case "reviewed":
		return reviewed && !superseded
	case "superseded":
		return superseded
	default:
		return false
	}
}

func (r Row) GeneratedTimeLegal() bool {
	if r.ActivityUTCGenerated && r.ActivityLocalGenerated {
		return false
	}
	if r.ActivityUTCGenerated {
		parsed, ok := timecontract.ParseUTC(r.ActivityUTCText)
		if !ok || r.ActivityUTCText == nil || *r.ActivityUTCText != timecontract.FormatUTC(parsed) {
			return false
		}
	}
	if r.ActivityLocalGenerated {
		parsed, offset, ok := timecontract.ParseLocalOffset(r.ActivityLocalText)
		if !ok || r.ActivityLocalText == nil || *r.ActivityLocalText != timecontract.FormatLocal(parsed, offset/60) {
			return false
		}
	}
	return true
}

func (r Row) PairStateLegal() bool {
	utc, local := "", ""
	if r.ActivityUTCText != nil {
		utc = *r.ActivityUTCText
	}
	if r.ActivityLocalText != nil {
		local = *r.ActivityLocalText
	}
	generated := r.ActivityUTCGenerated || r.ActivityLocalGenerated
	switch r.ActivityTimePairState {
	case "disabled", "conversion_unavailable":
		return true
	case "empty":
		return utc == "" && local == "" && !generated
	case "paired_generated":
		return utc != "" && local != "" && generated
	case "paired_user_preserved":
		return utc != "" && local != "" && !generated
	case "paired_mismatch":
		return utc != "" && local != ""
	default:
		return false
	}
}
