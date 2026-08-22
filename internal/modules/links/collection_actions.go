package links

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type RecordRefCollectionValidation struct {
	IncidentID         uuid.UUID
	FieldKey           string
	LinkType           LinkType
	ExpectedTargetType string
	AddRecordIDs       []uuid.UUID
	RemoveRecordIDs    []uuid.UUID
}

type RecordRefCollectionCommand struct {
	IncidentID         uuid.UUID
	SourceRecordID     uuid.UUID
	ActorUserID        uuid.UUID
	FieldKey           string
	LinkType           LinkType
	ExpectedTargetType string
	AddRecordIDs       []uuid.UUID
	RemoveRecordIDs    []uuid.UUID
	Now                time.Time
}

type PartyRefCollectionValidation struct {
	IncidentID         uuid.UUID
	FieldKey           string
	LinkType           LinkType
	ExpectedTargetType string
	AddPartyIDs        []uuid.UUID
	RemovePartyIDs     []uuid.UUID
}

type PartyRefCollectionCommand struct {
	IncidentID         uuid.UUID
	SourceRecordID     uuid.UUID
	ActorUserID        uuid.UUID
	FieldKey           string
	LinkType           LinkType
	ExpectedTargetType string
	AddPartyIDs        []uuid.UUID
	RemovePartyIDs     []uuid.UUID
	Now                time.Time
}

type TagCollectionValidation struct {
	FieldKey   string
	AddTags    []TagCollectionAdd
	RemoveTags []RecordTagRef
}

type TagCollectionCommand struct {
	IncidentID  uuid.UUID
	RecordID    uuid.UUID
	ActorUserID uuid.UUID
	FieldKey    string
	AddTags     []TagCollectionAdd
	RemoveTags  []RecordTagRef
	Now         time.Time
}

type TagCollectionAdd struct {
	RawText        string
	NormalizedText string
}

type RecordTagRef struct {
	RecordID    uuid.UUID
	RecordTagID uuid.UUID
}

type CollectionValidationError struct {
	Field      string
	ReasonCode string
}

func (e *CollectionValidationError) Error() string {
	return "links: invalid collection action"
}

func (s *Store) ValidateRecordRefCollectionTx(ctx context.Context, tx pgx.Tx, command RecordRefCollectionValidation) error {
	if err := validateLinkCollectionPolicy(command.FieldKey, command.LinkType); err != nil {
		return err
	}
	for _, recordID := range command.AddRecordIDs {
		if err := validateCollectionTargetRecordTx(ctx, tx, command.IncidentID, recordID, command.ExpectedTargetType, command.FieldKey); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ValidatePartyRefCollectionTx(ctx context.Context, tx pgx.Tx, command PartyRefCollectionValidation) error {
	if err := validateLinkCollectionPolicy(command.FieldKey, command.LinkType); err != nil {
		return err
	}
	for _, partyID := range command.AddPartyIDs {
		if err := validateCollectionTargetRecordTx(ctx, tx, command.IncidentID, partyID, command.ExpectedTargetType, command.FieldKey); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ValidateTagCollectionTx(ctx context.Context, tx pgx.Tx, command TagCollectionValidation) error {
	_ = ctx
	_ = tx
	if err := validateTagCollectionPolicy(command.FieldKey); err != nil {
		return err
	}
	for _, tag := range command.AddTags {
		if tag.RawText == "" || tag.NormalizedText == "" {
			return collectionValidationError(command.FieldKey)
		}
	}
	for _, tag := range command.RemoveTags {
		if tag.RecordID == uuid.Nil || tag.RecordTagID == uuid.Nil {
			return collectionValidationError(command.FieldKey)
		}
	}
	return nil
}

func validateLinkCollectionPolicy(fieldKey string, linkType LinkType) error {
	if fieldKey == "" {
		return collectionValidationError("field_key")
	}
	if linkType.String() == "" {
		return collectionValidationError(fieldKey)
	}
	return nil
}

func validateTagCollectionPolicy(fieldKey string) error {
	if fieldKey == "" {
		return collectionValidationError("field_key")
	}
	return nil
}

func validateCollectionTargetRecordTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, expectedType string, field string) error {
	var exists bool
	if expectedType == "" {
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records
     WHERE incident_id = $1
       AND record_id = $2
       AND deleted_at IS NULL
)
`, incidentID, recordID).Scan(&exists); err != nil {
			return fmt.Errorf("validate collection target: %w", err)
		}
	} else if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records
     WHERE incident_id = $1
       AND record_id = $2
       AND record_type = $3
       AND deleted_at IS NULL
)
`, incidentID, recordID, expectedType).Scan(&exists); err != nil {
		return fmt.Errorf("validate collection target: %w", err)
	}
	if !exists {
		return collectionValidationError(field)
	}
	return nil
}

func collectionValidationError(field string) *CollectionValidationError {
	return &CollectionValidationError{Field: field, ReasonCode: "invalid_value"}
}
