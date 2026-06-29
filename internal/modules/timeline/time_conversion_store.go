package timeline

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func (s *Store) GetTimeConversionProfile(ctx context.Context, incidentID uuid.UUID, now time.Time) (TimeConversionProfile, error) {
	var (
		profile   TimeConversionProfile
		offset    pgtype.Int4
		label     pgtype.Text
		updatedBy pgtype.UUID
	)
	err := s.pool.QueryRow(ctx, `
SELECT incident_id, enabled, local_offset_minutes, local_label, profile_version, updated_at, updated_by_user_id
  FROM timeline_time_conversion_profiles
 WHERE incident_id = $1
`, incidentID).Scan(&profile.IncidentID, &profile.Enabled, &offset, &label, &profile.ProfileVersion, &profile.UpdatedAt, &updatedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return TimeConversionProfile{
			IncidentID:      incidentID,
			Enabled:         false,
			ProfileVersion:  1,
			UpdatedAt:       now.UTC(),
			UpdatedByUserID: nil,
		}, nil
	}
	if err != nil {
		return TimeConversionProfile{}, fmt.Errorf("get timeline time conversion profile: %w", err)
	}
	profile.LocalOffsetMinutes = optionalIntFromPG(offset)
	profile.LocalLabel = optionalTextFromPG(label)
	profile.UpdatedByUserID = optionalUUIDFromPG(updatedBy)
	profile.UpdatedAt = profile.UpdatedAt.UTC()
	return profile, nil
}

func (s *Store) PutTimeConversionProfile(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request TimeConversionProfilePutRequest, now time.Time) (TimeConversionProfile, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return TimeConversionProfile{}, fmt.Errorf("begin timeline time conversion profile transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	current, err := getTimeConversionProfileTx(ctx, tx, incidentID, now)
	if err != nil {
		return TimeConversionProfile{}, err
	}
	if current.ProfileVersion != request.BaseProfileVersion {
		return TimeConversionProfile{}, &RowVersionConflictError{
			BaseRowVersion:    request.BaseProfileVersion,
			CurrentRowVersion: current.ProfileVersion,
		}
	}
	nextVersion := current.ProfileVersion + 1
	var (
		profile   TimeConversionProfile
		offset    pgtype.Int4
		label     pgtype.Text
		updatedBy pgtype.UUID
	)
	err = tx.QueryRow(ctx, `
INSERT INTO timeline_time_conversion_profiles (
    incident_id,
    enabled,
    local_offset_minutes,
    local_label,
    profile_version,
    updated_at,
    updated_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (incident_id) DO UPDATE
SET enabled = EXCLUDED.enabled,
    local_offset_minutes = EXCLUDED.local_offset_minutes,
    local_label = EXCLUDED.local_label,
    profile_version = EXCLUDED.profile_version,
    updated_at = EXCLUDED.updated_at,
    updated_by_user_id = EXCLUDED.updated_by_user_id
RETURNING incident_id, enabled, local_offset_minutes, local_label, profile_version, updated_at, updated_by_user_id
`, incidentID, request.Enabled, request.LocalOffsetMinutes, request.LocalLabel, nextVersion, now.UTC(), actor.ID).Scan(&profile.IncidentID, &profile.Enabled, &offset, &label, &profile.ProfileVersion, &profile.UpdatedAt, &updatedBy)
	if err != nil {
		return TimeConversionProfile{}, fmt.Errorf("put timeline time conversion profile: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return TimeConversionProfile{}, fmt.Errorf("commit timeline time conversion profile: %w", err)
	}
	profile.LocalOffsetMinutes = optionalIntFromPG(offset)
	profile.LocalLabel = optionalTextFromPG(label)
	profile.UpdatedByUserID = optionalUUIDFromPG(updatedBy)
	profile.UpdatedAt = profile.UpdatedAt.UTC()
	return profile, nil
}

func getTimeConversionProfileTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, now time.Time) (TimeConversionProfile, error) {
	var (
		profile   TimeConversionProfile
		offset    pgtype.Int4
		label     pgtype.Text
		updatedBy pgtype.UUID
	)
	err := tx.QueryRow(ctx, `
SELECT incident_id, enabled, local_offset_minutes, local_label, profile_version, updated_at, updated_by_user_id
  FROM timeline_time_conversion_profiles
 WHERE incident_id = $1
 FOR UPDATE
`, incidentID).Scan(&profile.IncidentID, &profile.Enabled, &offset, &label, &profile.ProfileVersion, &profile.UpdatedAt, &updatedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return TimeConversionProfile{
			IncidentID:     incidentID,
			Enabled:        false,
			ProfileVersion: 1,
			UpdatedAt:      now.UTC(),
		}, nil
	}
	if err != nil {
		return TimeConversionProfile{}, fmt.Errorf("get timeline time conversion profile: %w", err)
	}
	profile.LocalOffsetMinutes = optionalIntFromPG(offset)
	profile.LocalLabel = optionalTextFromPG(label)
	profile.UpdatedByUserID = optionalUUIDFromPG(updatedBy)
	profile.UpdatedAt = profile.UpdatedAt.UTC()
	return profile, nil
}

func applyTimelineTimeConversion(record *sourceRecord, profile TimeConversionProfile) {
	if !profile.Enabled || profile.LocalOffsetMinutes == nil {
		record.ActivityTimePairState = "disabled"
		return
	}
	offsetMinutes := *profile.LocalOffsetMinutes
	utcEmpty := timelineVisibleTextEmpty(record.ActivityUTCText)
	localEmpty := timelineVisibleTextEmpty(record.ActivityLocalText)
	if utcEmpty && localEmpty {
		record.ActivityTimePairState = "empty"
		return
	}

	utcParsed, utcOK := parseTimelineUTCTextExact(record.ActivityUTCText)
	localParsed, localOffset, localOK := parseTimelineLocalTextExact(record.ActivityLocalText)

	if !utcEmpty && (localEmpty || record.ActivityLocalGenerated) {
		if !utcOK {
			record.ActivityTimePairState = "conversion_unavailable"
			return
		}
		generated := formatTimelineLocalText(utcParsed, offsetMinutes)
		record.ActivityLocalText = &generated
		record.ActivityLocalGenerated = true
		record.ActivityUTCGenerated = false
		record.ActivityTimePairState = "paired_generated"
		return
	}
	if !localEmpty && (utcEmpty || record.ActivityUTCGenerated) {
		if !localOK {
			record.ActivityTimePairState = "conversion_unavailable"
			return
		}
		generated := formatTimelineUTCText(localParsed)
		record.ActivityUTCText = &generated
		record.ActivityUTCGenerated = true
		record.ActivityLocalGenerated = false
		record.ActivityTimePairState = "paired_generated"
		return
	}
	if !utcOK || !localOK {
		record.ActivityTimePairState = "conversion_unavailable"
		return
	}
	if utcParsed.Equal(localParsed) && localOffset == offsetMinutes*60 {
		record.ActivityTimePairState = "paired_user_preserved"
		return
	}
	record.ActivityTimePairState = "paired_mismatch"
}

func timelineVisibleTextEmpty(text *string) bool {
	return text == nil || *text == ""
}

func parseTimelineUTCTextExact(text *string) (time.Time, bool) {
	if text == nil || *text == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse("2006-01-02T15:04:05Z", *text)
	if err != nil {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func parseTimelineLocalTextExact(text *string) (time.Time, int, bool) {
	if text == nil || *text == "" {
		return time.Time{}, 0, false
	}
	parsed, err := time.Parse("2006-01-02T15:04:05-07:00", *text)
	if err != nil {
		return time.Time{}, 0, false
	}
	_, offsetSeconds := parsed.Zone()
	return parsed.UTC(), offsetSeconds, true
}

func formatTimelineUTCText(value time.Time) string {
	return value.UTC().Format("2006-01-02T15:04:05Z")
}

func formatTimelineLocalText(value time.Time, offsetMinutes int) string {
	location := time.FixedZone("", offsetMinutes*60)
	return value.UTC().In(location).Format("2006-01-02T15:04:05-07:00")
}
