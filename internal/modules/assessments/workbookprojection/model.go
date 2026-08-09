package workbookprojection

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ProjectionInput struct {
	RecordID            uuid.UUID
	IncidentID          uuid.UUID
	RowVersion          int64
	SubjectRef          uuid.UUID
	SubjectType         string
	AssessmentState     string
	ConfidenceScore     *int
	ConfidenceBand      string
	Rationale           string
	Assessor            uuid.UUID
	AssessedAt          time.Time
	SupportingLinkCount int
}

func (input ProjectionInput) Validate() error {
	switch {
	case input.RecordID == uuid.Nil:
		return errors.New("assessment projection input record_id is required")
	case input.IncidentID == uuid.Nil:
		return errors.New("assessment projection input incident_id is required")
	case input.RowVersion <= 0:
		return errors.New("assessment projection input row_version must be positive")
	case input.SubjectRef == uuid.Nil:
		return errors.New("assessment projection input subject_ref is required")
	case input.SubjectType != "host" && input.SubjectType != "identity":
		return fmt.Errorf("assessment projection input subject_type %q is invalid", input.SubjectType)
	case input.Assessor == uuid.Nil:
		return errors.New("assessment projection input assessor is required")
	case input.AssessedAt.IsZero():
		return errors.New("assessment projection input assessed_at is required")
	case input.SupportingLinkCount < 0:
		return errors.New("assessment projection input supporting_link_count must be non-negative")
	}
	switch input.AssessmentState {
	case "unknown", "suspected", "confirmed", "disproven", "cleared":
	default:
		return fmt.Errorf(
			"assessment projection input assessment_state %q is invalid",
			input.AssessmentState,
		)
	}
	if input.ConfidenceBand != confidenceBand(input.ConfidenceScore) {
		return errors.New("assessment projection input confidence_band is not canonical")
	}
	return nil
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
		return errors.New("assessment projection mutation record_id is required")
	}
	switch mutation.Kind {
	case ProjectionMutationUpsert:
		if mutation.Input.RecordID != mutation.RecordID {
			return errors.New("assessment projection upsert record_id does not match input")
		}
		return mutation.Input.Validate()
	case ProjectionMutationDelete:
		if !mutation.Input.isZero() {
			return errors.New("assessment projection delete must not carry an input")
		}
		return nil
	default:
		return fmt.Errorf("unsupported assessment projection mutation kind %q", mutation.Kind)
	}
}

func (input ProjectionInput) isZero() bool {
	return input.RecordID == uuid.Nil &&
		input.IncidentID == uuid.Nil &&
		input.RowVersion == 0 &&
		input.SubjectRef == uuid.Nil &&
		input.SubjectType == "" &&
		input.AssessmentState == "" &&
		input.ConfidenceScore == nil &&
		input.ConfidenceBand == "" &&
		input.Rationale == "" &&
		input.Assessor == uuid.Nil &&
		input.AssessedAt.IsZero() &&
		input.SupportingLinkCount == 0
}

func confidenceBand(score *int) string {
	switch {
	case score == nil:
		return "unset"
	case *score >= 0 && *score <= 39:
		return "low"
	case *score >= 40 && *score <= 69:
		return "medium"
	case *score >= 70 && *score <= 100:
		return "high"
	default:
		return ""
	}
}

type ProjectionInputPage struct {
	Inputs       []ProjectionInput
	NextRecordID *uuid.UUID
}

type Envelope struct {
	RecordID   uuid.UUID
	IncidentID uuid.UUID
	RecordType string
	RowVersion int64
	DeletedAt  *time.Time
}

type SupportFacts struct {
	ActiveTargetCount int
}

type EnvelopeReader interface {
	LoadAssessmentProjectionEnvelopeTx(context.Context, pgx.Tx, uuid.UUID) (Envelope, bool, error)
}

type SupportFactReader interface {
	LoadAssessmentProjectionSupportFactsTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (SupportFacts, error)
}
