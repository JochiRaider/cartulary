package timeline

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mutationpolicy"
)

type timelinePreparationBinding struct{ contractMajor int }

type preparedTimelineImport struct {
	binding             *timelinePreparationBinding
	members             *incidentportability.BoundMembers
	incidentID, actorID uuid.UUID
	operationID         string
	version             int
	profile             *portableTimelineProfile
}

type portableTimelineProfile struct {
	IncidentID         uuid.UUID  `json:"incident_id"`
	Enabled            bool       `json:"enabled"`
	LocalOffsetMinutes *int64     `json:"local_offset_minutes"`
	LocalLabel         *string    `json:"local_label"`
	ProfileVersion     int64      `json:"profile_version"`
	UpdatedAt          time.Time  `json:"updated_at"`
	UpdatedByUserID    *uuid.UUID `json:"updated_by_user_id"`
}

func (p preparedTimelineImport) matches(input sourceport.ImportContext, binding *timelinePreparationBinding) bool {
	return binding != nil && p.binding == binding && binding.contractMajor == sourceport.ContractMajor && p.members != nil && p.incidentID == input.IncidentID && p.actorID == input.ActorUserID && p.operationID == input.OperationID && p.version == input.BundleVersion && input.Attributions != nil
}

func prepareTimelineImport(ctx context.Context, bundle sourceport.Bundle, input sourceport.ImportContext, d sourceport.Descriptor, binding *timelinePreparationBinding) (preparedTimelineImport, error) {
	if input.IncidentID == uuid.Nil || input.ActorUserID == uuid.Nil || input.BundleVersion != 8 || input.OperationID == "" {
		return preparedTimelineImport{}, sourceport.ErrPreparedBinding
	}
	paths := []string{timelineBundleProfilesPath, timelineBundleRecordsPath, timelineBundleProvenancePath}
	members, err := incidentportability.BindMembers(bundle, paths)
	if err != nil {
		return preparedTimelineImport{}, err
	}
	// Finish syntax admission for every immutable member before identity or
	// owner semantic phases can choose a failure.
	for _, path := range paths {
		err := incidentportability.WalkPath(ctx, members, path, func(map[string]any) error { return nil })
		var malformed *incidentportability.StrictNDJSONDecodeError
		if errors.As(err, &malformed) {
			return preparedTimelineImport{}, d.DeclaredFailure("timeline.version_shape_exact")
		}
		if err != nil {
			return preparedTimelineImport{}, err
		}
	}
	p := preparedTimelineImport{binding: binding, members: members, incidentID: input.IncidentID, actorID: input.ActorUserID, operationID: input.OperationID, version: input.BundleVersion}
	err = incidentportability.WalkPath(ctx, members, timelineBundleProfilesPath, func(raw map[string]any) error {
		if !portableTimelineUUID(raw["incident_id"]) || p.profile != nil {
			return d.DeclaredFailure("timeline.source_identity_admitted")
		}
		row, ok := parsePortableTimelineProfile(raw)
		if !ok {
			return d.DeclaredFailure("timeline.version_shape_exact")
		}
		if row.IncidentID != input.IncidentID || row.Enabled && row.LocalOffsetMinutes == nil {
			return d.DeclaredFailure("timeline.profile_coherent")
		}
		if row.UpdatedByUserID != nil {
			if _, ok := input.Actors.Lookup(row.UpdatedByUserID.String()); !ok {
				return d.DeclaredFailure("timeline.profile_attribution_exact")
			}
		}
		p.profile = &row
		return nil
	})
	if err != nil {
		return preparedTimelineImport{}, err
	}
	if err := p.prepareEvents(ctx, input, d); err != nil {
		return preparedTimelineImport{}, err
	}
	if err := p.prepareProvenance(ctx, d); err != nil {
		return preparedTimelineImport{}, err
	}
	return p, nil
}

func portableTimelineUUID(raw any) bool {
	text, ok := raw.(string)
	id, err := uuid.Parse(text)
	return ok && err == nil && id != uuid.Nil && text == id.String()
}

func portableTimelineInteger(raw any, minimum, maximum int64) bool {
	n, ok := raw.(json.Number)
	if !ok {
		return false
	}
	value, err := n.Int64()
	return err == nil && value >= minimum && value <= maximum && strconv.FormatInt(value, 10) == n.String()
}

func parsePortableTimelineProfile(raw map[string]any) (portableTimelineProfile, bool) {
	var row portableTimelineProfile
	fields := []string{"incident_id", "enabled", "local_offset_minutes", "local_label", "profile_version", "updated_at", "updated_by_user_id"}
	if len(raw) != len(fields) {
		return row, false
	}
	for _, field := range fields {
		if _, ok := raw[field]; !ok {
			return row, false
		}
	}
	if !portableTimelineUUID(raw["incident_id"]) {
		return row, false
	}
	if _, ok := raw["enabled"].(bool); !ok {
		return row, false
	}
	if !portableTimelineInteger(raw["profile_version"], 1, 1<<63-1) {
		return row, false
	}
	if raw["local_offset_minutes"] != nil && !portableTimelineInteger(raw["local_offset_minutes"], -840, 840) {
		return row, false
	}
	if raw["updated_by_user_id"] != nil && !portableTimelineUUID(raw["updated_by_user_id"]) {
		return row, false
	}
	if raw["local_label"] != nil {
		label, ok := raw["local_label"].(string)
		normalized, valid := mutationpolicy.NormalizeProfileLabel(label)
		if !ok || !valid || normalized == "" || normalized != label {
			return row, false
		}
	}
	stamp, ok := raw["updated_at"].(string)
	parsed, err := time.Parse(time.RFC3339Nano, stamp)
	_, offset := parsed.Zone()
	canonical := parsed.UTC().Format(time.RFC3339Nano)
	if !ok || err != nil || offset != 0 || parsed.Year() < 1 || parsed.Year() > 9999 || parsed.Nanosecond()%1000 != 0 || (stamp != canonical && stamp != strings.TrimSuffix(canonical, "Z")+"+00:00") {
		return row, false
	}
	encoded, err := json.Marshal(raw)
	if err != nil || json.Unmarshal(encoded, &row) != nil {
		return row, false
	}
	row.UpdatedAt = row.UpdatedAt.UTC()
	return row, true
}

func (p preparedTimelineImport) applyProfile(ctx context.Context, tx pgx.Tx, input sourceport.ImportContext, d sourceport.Descriptor) error {
	// Consume the sealed singleton through EOF before any owner mutation.
	if err := incidentportability.WalkPath(ctx, p.members, timelineBundleProfilesPath, func(map[string]any) error { return nil }); err != nil {
		return err
	}
	var occupied bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM timeline_time_conversion_profiles WHERE incident_id=$1)`, p.incidentID).Scan(&occupied); err != nil {
		return err
	}
	if occupied {
		return d.DeclaredFailure("timeline.source_identity_admitted")
	}
	if p.profile == nil {
		return nil
	}
	row := *p.profile
	var localActor *uuid.UUID
	if row.UpdatedByUserID != nil {
		localActor = &p.actorID
	}
	tag, err := tx.Exec(ctx, `INSERT INTO timeline_time_conversion_profiles(incident_id,enabled,local_offset_minutes,local_label,profile_version,updated_at,updated_by_user_id) VALUES($1,$2,$3,$4,$5,$6,$7)`, row.IncidentID, row.Enabled, row.LocalOffsetMinutes, row.LocalLabel, row.ProfileVersion, row.UpdatedAt, localActor)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return d.DeclaredFailure("timeline.source_identity_admitted")
	}
	if row.UpdatedByUserID != nil {
		return input.Attributions.RecordImportedAttribution("timeline_time_conversion_profiles", row.IncidentID.String(), "updated_by_user_id", row.UpdatedByUserID.String())
	}
	return nil
}

func (p preparedTimelineImport) validateProfile(ctx context.Context, tx pgx.Tx, input sourceport.ImportContext, d sourceport.Descriptor) error {
	var actual portableTimelineProfile
	err := tx.QueryRow(ctx, `SELECT incident_id,enabled,local_offset_minutes,local_label,profile_version,updated_at,updated_by_user_id FROM timeline_time_conversion_profiles WHERE incident_id=$1`, p.incidentID).Scan(&actual.IncidentID, &actual.Enabled, &actual.LocalOffsetMinutes, &actual.LocalLabel, &actual.ProfileVersion, &actual.UpdatedAt, &actual.UpdatedByUserID)
	if p.profile == nil {
		if err == nil {
			return d.DeclaredFailure("timeline.profile_state_exact")
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	} else {
		if errors.Is(err, pgx.ErrNoRows) {
			return d.DeclaredFailure("timeline.profile_state_exact")
		}
		if err != nil {
			return err
		}
		expected := *p.profile
		if expected.UpdatedByUserID != nil {
			expected.UpdatedByUserID = &p.actorID
		}
		actual.UpdatedAt = actual.UpdatedAt.UTC()
		if !reflect.DeepEqual(actual, expected) {
			return d.DeclaredFailure("timeline.profile_state_exact")
		}
	}
	var expected *incidentportability.ImportedAttribution
	if p.profile != nil && p.profile.UpdatedByUserID != nil {
		expected = &incidentportability.ImportedAttribution{SourceTable: "timeline_time_conversion_profiles", SourceRowID: p.incidentID.String(), SourceColumn: "updated_by_user_id", SourceActorID: p.profile.UpdatedByUserID.String(), LocalUserID: p.actorID}
	}
	err = input.Attributions.VisitImportedAttributions("timeline_time_conversion_profiles", "", func(actual incidentportability.ImportedAttribution) error {
		if expected == nil || actual != *expected {
			return d.DeclaredFailure("timeline.profile_attribution_exact")
		}
		expected = nil
		return ctx.Err()
	})
	if err != nil {
		return err
	}
	if expected != nil {
		return d.DeclaredFailure("timeline.profile_attribution_exact")
	}
	return nil
}
