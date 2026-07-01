package projectionprovider

import (
	"context"
	"errors"
	"fmt"
	"time"

	sqlc "github.com/JochiRaider/cartulary/internal/gen/sql"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/timecontract"
)

type ProjectionInput struct {
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
}

type UpsertFunc func(context.Context, pgx.Tx, ProjectionInput) error

func RebuildIncidentTimelineTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, upsert UpsertFunc) error {
	if upsert == nil {
		return errors.New("timeline projection upsert callback is required")
	}
	if _, err := tx.Exec(ctx, `DELETE FROM timeline_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear timeline projection rows: %w", err)
	}

	queries := sqlc.New(tx)
	rows, err := queries.ListTimelineProjectionSourceRows(ctx, pgUUID(incidentID))
	if err != nil {
		return fmt.Errorf("list timeline projection source rows: %w", err)
	}

	for _, row := range rows {
		input, err := projectionInputFromSQL(row)
		if err != nil {
			return err
		}
		if err := upsert(ctx, tx, input); err != nil {
			return err
		}
	}
	return nil
}

func projectionInputFromSQL(row sqlc.ListTimelineProjectionSourceRowsRow) (ProjectionInput, error) {
	recordID, err := uuidFromPG(row.RecordID)
	if err != nil {
		return ProjectionInput{}, err
	}
	incidentID, err := uuidFromPG(row.IncidentID)
	if err != nil {
		return ProjectionInput{}, err
	}
	recordedAt, err := timeFromPG(row.RecordedAt)
	if err != nil {
		return ProjectionInput{}, err
	}
	editedAt, err := timeFromPG(row.EditedAt)
	if err != nil {
		return ProjectionInput{}, err
	}

	return ProjectionInput{
		RecordID:              recordID,
		IncidentID:            incidentID,
		RowVersion:            row.RowVersion,
		DateEnteredText:       optionalTextFromPG(row.DateEnteredText),
		AnalystText:           optionalTextFromPG(row.AnalystText),
		MitreStageText:        optionalTextFromPG(row.MitreStageText),
		DeviceObjectText:      optionalTextFromPG(row.DeviceObjectText),
		IPAddressText:         optionalTextFromPG(row.IpAddressText),
		ActivityUTCText:       optionalTextFromPG(row.ActivityUtcText),
		ActivityLocalText:     optionalTextFromPG(row.ActivityLocalText),
		RawActivityText:       optionalTextFromPG(row.RawActivityText),
		ActivitySynopsisText:  optionalTextFromPG(row.ActivitySynopsisText),
		DataSourceText:        optionalTextFromPG(row.DataSourceText),
		RecordedAt:            recordedAt,
		EditedAt:              editedAt,
		ActivitySortTS:        deriveActivitySortTS(optionalTextFromPG(row.ActivityUtcText), optionalTextFromPG(row.ActivityLocalText)),
		DateEnteredSortDay:    deriveDateEnteredSortDay(optionalTextFromPG(row.DateEnteredText)),
		ActivityTimePairState: row.ActivityTimePairState,
		CaptureState:          row.CaptureState,
		ReplacementRecordID:   optionalUUIDFromPG(row.ReplacementRecordID),
		EvidenceCount:         int(row.EvidenceCount),
		HasEvidence:           row.HasEvidence,
		HasUnresolvedMentions: row.HasUnresolvedMentions,
	}, nil
}

func pgUUID(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(value), Valid: true}
}

func uuidFromPG(value pgtype.UUID) (uuid.UUID, error) {
	if !value.Valid {
		return uuid.UUID{}, errors.New("missing uuid")
	}
	return uuid.FromBytes(value.Bytes[:])
}

func optionalUUIDFromPG(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	parsed := uuid.Must(uuid.FromBytes(value.Bytes[:]))
	return &parsed
}

func timeFromPG(value pgtype.Timestamptz) (time.Time, error) {
	if !value.Valid {
		return time.Time{}, errors.New("missing timestamp")
	}
	return value.Time.UTC(), nil
}

func optionalTextFromPG(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	text := value.String
	return &text
}

func deriveActivitySortTS(utcText *string, localText *string) *time.Time {
	if parsed := parseTimelineUTCText(utcText); parsed != nil {
		return parsed
	}
	if parsed := parseTimelineLocalText(localText); parsed != nil {
		return parsed
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
	if parsed := parseTimelineUTCText(text); parsed != nil {
		day := time.Date(parsed.UTC().Year(), parsed.UTC().Month(), parsed.UTC().Day(), 0, 0, 0, 0, time.UTC)
		return &day
	}
	if parsed := parseTimelineLocalText(text); parsed != nil {
		day := time.Date(parsed.UTC().Year(), parsed.UTC().Month(), parsed.UTC().Day(), 0, 0, 0, 0, time.UTC)
		return &day
	}
	return nil
}

func parseTimelineUTCText(text *string) *time.Time {
	if parsed, ok := timecontract.ParseUTC(text); ok {
		return &parsed
	}
	return nil
}

func parseTimelineLocalText(text *string) *time.Time {
	if parsed, _, ok := timecontract.ParseLocalOffset(text); ok {
		return &parsed
	}
	return nil
}
