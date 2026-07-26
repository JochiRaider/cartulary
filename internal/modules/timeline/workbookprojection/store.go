package workbookprojection

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
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

type ProjectionMutationKind string

const (
	ProjectionMutationUpsert ProjectionMutationKind = "upsert"
	ProjectionMutationDelete ProjectionMutationKind = "delete"
)

type ProjectionMutation struct {
	Kind     ProjectionMutationKind
	RecordID uuid.UUID
	Input    ProjectionInput
}

func (mutation ProjectionMutation) Validate() error {
	if mutation.RecordID == uuid.Nil {
		return errors.New("timeline projection mutation record_id is required")
	}
	switch mutation.Kind {
	case ProjectionMutationUpsert:
		if mutation.Input.RecordID != mutation.RecordID {
			return errors.New("timeline projection upsert record_id does not match input")
		}
		return validateProjectionInput(mutation.Input)
	case ProjectionMutationDelete:
		if mutation.Input != (ProjectionInput{}) {
			return errors.New("timeline projection delete must not carry an input")
		}
		return nil
	default:
		return fmt.Errorf("unsupported timeline projection mutation kind %q", mutation.Kind)
	}
}

type CollectionHydrator interface {
	HydrateTimelineCollectionsTx(context.Context, pgx.Tx, *DerivedRecord) error
}

type Source struct {
	repository  *sourcerepository.Repository
	collections CollectionHydrator
}

func NewSource(envelopes sourcerepository.EnvelopeReader, collections CollectionHydrator) *Source {
	return &Source{
		repository:  sourcerepository.New(envelopes),
		collections: collections,
	}
}

func (s *Source) BuildProjectionMutationTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (ProjectionMutation, error) {
	if recordID == uuid.Nil {
		return ProjectionMutation{}, errors.New("timeline projection mutation record_id is required")
	}
	if s == nil || s.repository == nil || s.collections == nil {
		return ProjectionMutation{}, errors.New("timeline projection source dependencies are required")
	}
	snapshot, err := s.repository.LoadUnlockedTx(ctx, tx, recordID)
	if errors.Is(err, sourcerepository.ErrNotFound) {
		return ProjectionMutation{Kind: ProjectionMutationDelete, RecordID: recordID}, nil
	}
	if err != nil {
		return ProjectionMutation{}, err
	}
	derived := Derive(snapshot, nil)
	if err := s.collections.HydrateTimelineCollectionsTx(ctx, tx, &derived); err != nil {
		return ProjectionMutation{}, err
	}
	input := derived.ProjectionInput()
	if err := validateProjectionInput(input); err != nil {
		return ProjectionMutation{}, err
	}
	return ProjectionMutation{Kind: ProjectionMutationUpsert, RecordID: recordID, Input: input}, nil
}

type ProjectionInputPage struct {
	Inputs       []ProjectionInput
	NextRecordID *uuid.UUID
}

func (s *Source) ListProjectionInputsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, afterRecordID *uuid.UUID, limit int) (ProjectionInputPage, error) {
	if incidentID == uuid.Nil {
		return ProjectionInputPage{}, errors.New("timeline projection enumeration incident_id is required")
	}
	if limit <= 0 || limit > 1000 {
		return ProjectionInputPage{}, fmt.Errorf("timeline projection enumeration limit %d is outside 1..1000", limit)
	}
	if s == nil || s.repository == nil || s.collections == nil {
		return ProjectionInputPage{}, errors.New("timeline projection source dependencies are required")
	}
	sourcePage, err := s.repository.ListIncidentPageTx(ctx, tx, incidentID, afterRecordID, limit)
	if err != nil {
		return ProjectionInputPage{}, err
	}

	inputs := make([]ProjectionInput, 0, len(sourcePage.Snapshots))
	for _, snapshot := range sourcePage.Snapshots {
		derived := Derive(snapshot, nil)
		if err := s.collections.HydrateTimelineCollectionsTx(ctx, tx, &derived); err != nil {
			return ProjectionInputPage{}, err
		}
		input := derived.ProjectionInput()
		if err := validateProjectionInput(input); err != nil {
			return ProjectionInputPage{}, err
		}
		inputs = append(inputs, input)
	}
	return ProjectionInputPage{
		Inputs:       inputs,
		NextRecordID: sourcePage.NextRecordID,
	}, nil
}

func validateProjectionInput(input ProjectionInput) error {
	switch {
	case input.RecordID == uuid.Nil:
		return errors.New("timeline projection input record_id is required")
	case input.IncidentID == uuid.Nil:
		return errors.New("timeline projection input incident_id is required")
	case input.RowVersion <= 0:
		return errors.New("timeline projection input row_version must be positive")
	case input.RecordedAt.IsZero():
		return errors.New("timeline projection input recorded_at is required")
	case input.EditedAt.IsZero():
		return errors.New("timeline projection input edited_at is required")
	}
	switch input.CaptureState {
	case "rough", "enriched", "reviewed", "superseded":
	default:
		return fmt.Errorf("timeline projection input capture_state %q is invalid", input.CaptureState)
	}
	switch input.ActivityTimePairState {
	case "disabled", "empty", "paired_generated", "paired_user_preserved", "paired_mismatch", "conversion_unavailable":
	default:
		return fmt.Errorf("timeline projection input activity_time_pair_state %q is invalid", input.ActivityTimePairState)
	}
	return nil
}
