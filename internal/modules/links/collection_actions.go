package links

import (
	"context"
	"errors"
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

func (s *Store) ApplyRecordRefCollectionTx(ctx context.Context, tx pgx.Tx, command RecordRefCollectionCommand) (bool, error) {
	if err := s.ValidateRecordRefCollectionTx(ctx, tx, RecordRefCollectionValidation{
		IncidentID:         command.IncidentID,
		FieldKey:           command.FieldKey,
		LinkType:           command.LinkType,
		ExpectedTargetType: command.ExpectedTargetType,
		AddRecordIDs:       command.AddRecordIDs,
		RemoveRecordIDs:    command.RemoveRecordIDs,
	}); err != nil {
		return false, err
	}
	changed := false
	linkType := command.LinkType.String()
	for _, recordID := range command.AddRecordIDs {
		applied, err := s.UpsertFieldReferenceTx(ctx, tx, command.IncidentID, command.SourceRecordID, recordID, command.FieldKey, linkType, command.ActorUserID, command.Now)
		if err != nil {
			return false, err
		}
		changed = changed || applied
	}
	for _, recordID := range command.RemoveRecordIDs {
		applied, err := s.TombstoneFieldReferenceTx(ctx, tx, command.IncidentID, command.SourceRecordID, recordID, command.FieldKey, linkType, command.ExpectedTargetType, command.ActorUserID, command.Now)
		if err != nil {
			if errors.Is(err, ErrFieldReferenceNotFound) {
				return false, collectionValidationError(command.FieldKey)
			}
			return false, err
		}
		changed = changed || applied
	}
	return changed, nil
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

func (s *Store) ApplyPartyRefCollectionTx(ctx context.Context, tx pgx.Tx, command PartyRefCollectionCommand) (bool, error) {
	if err := s.ValidatePartyRefCollectionTx(ctx, tx, PartyRefCollectionValidation{
		IncidentID:         command.IncidentID,
		FieldKey:           command.FieldKey,
		LinkType:           command.LinkType,
		ExpectedTargetType: command.ExpectedTargetType,
		AddPartyIDs:        command.AddPartyIDs,
		RemovePartyIDs:     command.RemovePartyIDs,
	}); err != nil {
		return false, err
	}
	changed := false
	linkType := command.LinkType.String()
	for _, partyID := range command.AddPartyIDs {
		applied, err := s.UpsertFieldReferenceTx(ctx, tx, command.IncidentID, command.SourceRecordID, partyID, command.FieldKey, linkType, command.ActorUserID, command.Now)
		if err != nil {
			return false, err
		}
		changed = changed || applied
	}
	for _, partyID := range command.RemovePartyIDs {
		applied, err := s.TombstoneFieldReferenceTx(ctx, tx, command.IncidentID, command.SourceRecordID, partyID, command.FieldKey, linkType, command.ExpectedTargetType, command.ActorUserID, command.Now)
		if err != nil {
			if errors.Is(err, ErrFieldReferenceNotFound) {
				return false, collectionValidationError(command.FieldKey)
			}
			return false, err
		}
		changed = changed || applied
	}
	return changed, nil
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

func (s *Store) ApplyTagCollectionTx(ctx context.Context, tx pgx.Tx, command TagCollectionCommand) (bool, error) {
	if err := s.ValidateTagCollectionTx(ctx, tx, TagCollectionValidation{
		FieldKey:   command.FieldKey,
		AddTags:    command.AddTags,
		RemoveTags: command.RemoveTags,
	}); err != nil {
		return false, err
	}
	changed := false
	tags := s.Tags()
	for _, tag := range command.AddTags {
		applied, err := tags.UpsertTagTx(ctx, tx, command.IncidentID, command.RecordID, tag.RawText, tag.NormalizedText, command.ActorUserID, command.Now)
		if err != nil {
			if errors.Is(err, ErrInvalidTag) {
				return false, collectionValidationError(command.FieldKey)
			}
			return false, err
		}
		changed = changed || applied
	}
	for _, tag := range command.RemoveTags {
		if tag.RecordID != command.RecordID {
			return false, collectionValidationError(command.FieldKey)
		}
		applied, err := tags.TombstoneTagTx(ctx, tx, command.IncidentID, command.RecordID, tag.RecordTagID, command.ActorUserID, command.Now)
		if err != nil {
			if errors.Is(err, ErrTagNotFound) {
				return false, collectionValidationError(command.FieldKey)
			}
			return false, err
		}
		changed = changed || applied
	}
	return changed, nil
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
