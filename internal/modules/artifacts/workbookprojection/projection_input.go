package workbookprojection

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ProjectionEnvelope is the immutable common source state shared by every
// Artifact projection variant. Artifact type is deliberately absent: it is
// derived from the selected closed variant.
type ProjectionEnvelope struct {
	recordID          uuid.UUID
	incidentID        uuid.UUID
	rowVersion        int64
	title             *string
	body              *string
	timestampUTC      *time.Time
	updatedAt         time.Time
	createdAt         time.Time
	createdByUserID   *uuid.UUID
	timestampDay      *string
	linkedRecordCount int
}

func NewProjectionEnvelope(
	recordID uuid.UUID,
	incidentID uuid.UUID,
	rowVersion int64,
	title *string,
	body *string,
	timestampUTC *time.Time,
	updatedAt time.Time,
	createdAt time.Time,
	createdByUserID *uuid.UUID,
	timestampDay *string,
	linkedRecordCount int,
) (ProjectionEnvelope, error) {
	if recordID == uuid.Nil || incidentID == uuid.Nil || rowVersion < 1 ||
		updatedAt.IsZero() || createdAt.IsZero() || linkedRecordCount < 0 {
		return ProjectionEnvelope{}, fmt.Errorf("artifact projection envelope is incomplete")
	}
	if timestampUTC == nil && timestampDay != nil || timestampUTC != nil && (timestampDay == nil || *timestampDay == "") {
		return ProjectionEnvelope{}, fmt.Errorf("artifact projection timestamp facts are inconsistent")
	}
	return ProjectionEnvelope{
		recordID: recordID, incidentID: incidentID, rowVersion: rowVersion,
		title: cloneString(title), body: cloneString(body), timestampUTC: cloneTime(timestampUTC),
		updatedAt: updatedAt.UTC(), createdAt: createdAt.UTC(), createdByUserID: cloneUUID(createdByUserID),
		timestampDay: cloneString(timestampDay), linkedRecordCount: linkedRecordCount,
	}, nil
}

func (e ProjectionEnvelope) RecordID() uuid.UUID         { return e.recordID }
func (e ProjectionEnvelope) IncidentID() uuid.UUID       { return e.incidentID }
func (e ProjectionEnvelope) RowVersion() int64           { return e.rowVersion }
func (e ProjectionEnvelope) Title() *string              { return cloneString(e.title) }
func (e ProjectionEnvelope) Body() *string               { return cloneString(e.body) }
func (e ProjectionEnvelope) TimestampUTC() *time.Time    { return cloneTime(e.timestampUTC) }
func (e ProjectionEnvelope) UpdatedAt() time.Time        { return e.updatedAt }
func (e ProjectionEnvelope) CreatedAt() time.Time        { return e.createdAt }
func (e ProjectionEnvelope) CreatedByUserID() *uuid.UUID { return cloneUUID(e.createdByUserID) }
func (e ProjectionEnvelope) TimestampDay() *string       { return cloneString(e.timestampDay) }
func (e ProjectionEnvelope) LinkedRecordCount() int      { return e.linkedRecordCount }

type ProjectionVariant interface {
	projectionVariant()
}

type NoteVariant struct{}

func (NoteVariant) projectionVariant() {}

type CommunicationLogVariant struct {
	CommID           string
	CommType         string
	Audience         string
	ChannelOrMeeting string
	Summary          string
	NextReportAt     *time.Time
	PrivilegeTag     *string
	NextReportDay    *string
}

func (CommunicationLogVariant) projectionVariant() {}

type HandoffVariant struct {
	HandoffID           string
	OutgoingOwnerUserID uuid.UUID
	IncomingOwnerUserID uuid.UUID
	CurrentStateSummary string
	NextChecks          *string
	AcknowledgedAt      *time.Time
	AckState            string
}

func (HandoffVariant) projectionVariant() {}

type StatusReviewVariant struct {
	StatusReviewID      string
	ReviewOwnerUserID   uuid.UUID
	CurrentStateSummary string
	ActiveRisksSummary  *string
	NextReportAt        *time.Time
	NextReportDay       *string
}

func (StatusReviewVariant) projectionVariant() {}

type LessonVariant struct {
	LessonID     string
	Summary      string
	OwnerUserID  uuid.UUID
	ClosureState string
}

func (LessonVariant) projectionVariant() {}

type FindingVariant struct {
	Statement       string
	Kind            string
	State           string
	OwnerUserID     uuid.UUID
	ConfidenceScore *int
	ClosedAt        *time.Time
	UpdatedAt       time.Time
	ConfidenceBand  string
}

func (FindingVariant) projectionVariant() {}

type InvestigativeQueryVariant struct {
	QueryID         string
	Platform        string
	Purpose         string
	QueryText       string
	CreatedByUserID uuid.UUID
	CreatedAt       time.Time
	CreatedDay      string
}

func (InvestigativeQueryVariant) projectionVariant() {}

type ForensicKeywordVariant struct {
	KeywordID     string
	Pattern       string
	Reason        string
	MatchMode     string
	CaseSensitive bool
	CreatedAt     time.Time
	CreatedDay    string
}

func (ForensicKeywordVariant) projectionVariant() {}

// ProjectionInput is an immutable discriminated union. Every valid value has
// a complete common envelope and exactly one of the eight closed variants.
type ProjectionInput struct {
	envelope ProjectionEnvelope
	variant  ProjectionVariant
}

func NewNoteProjectionInput(envelope ProjectionEnvelope, variant NoteVariant) (ProjectionInput, error) {
	return newProjectionInput(envelope, variant)
}

func NewCommunicationLogProjectionInput(envelope ProjectionEnvelope, variant CommunicationLogVariant) (ProjectionInput, error) {
	return newProjectionInput(envelope, variant)
}

func NewHandoffProjectionInput(envelope ProjectionEnvelope, variant HandoffVariant) (ProjectionInput, error) {
	return newProjectionInput(envelope, variant)
}

func NewStatusReviewProjectionInput(envelope ProjectionEnvelope, variant StatusReviewVariant) (ProjectionInput, error) {
	return newProjectionInput(envelope, variant)
}

func NewLessonProjectionInput(envelope ProjectionEnvelope, variant LessonVariant) (ProjectionInput, error) {
	return newProjectionInput(envelope, variant)
}

func NewFindingProjectionInput(envelope ProjectionEnvelope, variant FindingVariant) (ProjectionInput, error) {
	return newProjectionInput(envelope, variant)
}

func NewInvestigativeQueryProjectionInput(envelope ProjectionEnvelope, variant InvestigativeQueryVariant) (ProjectionInput, error) {
	return newProjectionInput(envelope, variant)
}

func NewForensicKeywordProjectionInput(envelope ProjectionEnvelope, variant ForensicKeywordVariant) (ProjectionInput, error) {
	return newProjectionInput(envelope, variant)
}

func newProjectionInput(envelope ProjectionEnvelope, variant ProjectionVariant) (ProjectionInput, error) {
	if envelope.recordID == uuid.Nil || envelope.incidentID == uuid.Nil || envelope.rowVersion < 1 ||
		envelope.updatedAt.IsZero() || envelope.createdAt.IsZero() || variant == nil {
		return ProjectionInput{}, fmt.Errorf("artifact projection input is incomplete")
	}
	if err := validateProjectionVariant(envelope, variant); err != nil {
		return ProjectionInput{}, err
	}
	return ProjectionInput{envelope: cloneProjectionEnvelope(envelope), variant: cloneProjectionVariant(variant)}, nil
}

func validateProjectionVariant(envelope ProjectionEnvelope, variant ProjectionVariant) error {
	switch value := variant.(type) {
	case NoteVariant:
		if blank(envelope.title) && blank(envelope.body) {
			return fmt.Errorf("artifact note projection requires title or body")
		}
	case CommunicationLogVariant:
		if envelope.timestampUTC == nil || value.CommID == "" || value.CommType == "" ||
			value.Audience == "" || value.ChannelOrMeeting == "" || value.Summary == "" ||
			!optionalPair(value.NextReportAt, value.NextReportDay) {
			return fmt.Errorf("artifact communication-log projection is incomplete")
		}
	case HandoffVariant:
		if envelope.timestampUTC == nil || value.HandoffID == "" || value.OutgoingOwnerUserID == uuid.Nil ||
			value.IncomingOwnerUserID == uuid.Nil || value.CurrentStateSummary == "" ||
			(value.AckState != "pending" && value.AckState != "acknowledged") ||
			(value.AcknowledgedAt == nil) != (value.AckState == "pending") {
			return fmt.Errorf("artifact handoff projection is incomplete")
		}
	case StatusReviewVariant:
		if envelope.timestampUTC == nil || value.StatusReviewID == "" || value.ReviewOwnerUserID == uuid.Nil ||
			value.CurrentStateSummary == "" || !optionalPair(value.NextReportAt, value.NextReportDay) {
			return fmt.Errorf("artifact status-review projection is incomplete")
		}
	case LessonVariant:
		if envelope.timestampUTC == nil || value.LessonID == "" || value.Summary == "" ||
			value.OwnerUserID == uuid.Nil || value.ClosureState == "" {
			return fmt.Errorf("artifact lesson projection is incomplete")
		}
	case FindingVariant:
		if value.Statement == "" || (value.Kind != "finding" && value.Kind != "hypothesis") ||
			(value.State != "open" && value.State != "closed") || value.OwnerUserID == uuid.Nil ||
			value.UpdatedAt.IsZero() || value.ConfidenceBand == "" ||
			(value.ConfidenceScore != nil && (*value.ConfidenceScore < 0 || *value.ConfidenceScore > 100)) {
			return fmt.Errorf("artifact finding projection is incomplete")
		}
	case InvestigativeQueryVariant:
		if value.QueryID == "" || value.Platform == "" || value.Purpose == "" || value.QueryText == "" ||
			value.CreatedByUserID == uuid.Nil || value.CreatedAt.IsZero() || value.CreatedDay == "" {
			return fmt.Errorf("artifact investigative-query projection is incomplete")
		}
	case ForensicKeywordVariant:
		if value.KeywordID == "" || value.Pattern == "" || value.Reason == "" ||
			(value.MatchMode != "literal" && value.MatchMode != "regex") || value.CreatedAt.IsZero() || value.CreatedDay == "" {
			return fmt.Errorf("artifact forensic-keyword projection is incomplete")
		}
	default:
		return fmt.Errorf("artifact projection variant is unknown")
	}
	return nil
}

func (input ProjectionInput) Envelope() ProjectionEnvelope {
	return cloneProjectionEnvelope(input.envelope)
}

func (input ProjectionInput) Variant() ProjectionVariant {
	return cloneProjectionVariant(input.variant)
}

func (input ProjectionInput) ArtifactType() string {
	switch input.variant.(type) {
	case NoteVariant:
		return "note"
	case CommunicationLogVariant:
		return "comm_log"
	case HandoffVariant:
		return "handoff"
	case StatusReviewVariant:
		return "status_review"
	case LessonVariant:
		return "lesson"
	case FindingVariant:
		return "finding"
	case InvestigativeQueryVariant:
		return "investigative_query"
	case ForensicKeywordVariant:
		return "forensic_keyword"
	default:
		return ""
	}
}

type ProjectionInputPage struct {
	inputs       []ProjectionInput
	nextRecordID *uuid.UUID
}

func NewProjectionInputPage(inputs []ProjectionInput, nextRecordID *uuid.UUID) (ProjectionInputPage, error) {
	if nextRecordID != nil && *nextRecordID == uuid.Nil {
		return ProjectionInputPage{}, fmt.Errorf("artifact projection page cursor is invalid")
	}
	for _, input := range inputs {
		if input.ArtifactType() == "" {
			return ProjectionInputPage{}, fmt.Errorf("artifact projection page contains an invalid input")
		}
	}
	return ProjectionInputPage{inputs: append([]ProjectionInput(nil), inputs...), nextRecordID: cloneUUID(nextRecordID)}, nil
}

func (page ProjectionInputPage) Inputs() []ProjectionInput {
	return append([]ProjectionInput(nil), page.inputs...)
}

func (page ProjectionInputPage) NextRecordID() *uuid.UUID {
	return cloneUUID(page.nextRecordID)
}

func cloneProjectionEnvelope(value ProjectionEnvelope) ProjectionEnvelope {
	value.title = cloneString(value.title)
	value.body = cloneString(value.body)
	value.timestampUTC = cloneTime(value.timestampUTC)
	value.createdByUserID = cloneUUID(value.createdByUserID)
	value.timestampDay = cloneString(value.timestampDay)
	return value
}

func cloneProjectionVariant(variant ProjectionVariant) ProjectionVariant {
	switch value := variant.(type) {
	case NoteVariant:
		return value
	case CommunicationLogVariant:
		value.NextReportAt = cloneTime(value.NextReportAt)
		value.PrivilegeTag = cloneString(value.PrivilegeTag)
		value.NextReportDay = cloneString(value.NextReportDay)
		return value
	case HandoffVariant:
		value.NextChecks = cloneString(value.NextChecks)
		value.AcknowledgedAt = cloneTime(value.AcknowledgedAt)
		return value
	case StatusReviewVariant:
		value.ActiveRisksSummary = cloneString(value.ActiveRisksSummary)
		value.NextReportAt = cloneTime(value.NextReportAt)
		value.NextReportDay = cloneString(value.NextReportDay)
		return value
	case LessonVariant:
		return value
	case FindingVariant:
		value.ConfidenceScore = cloneInt(value.ConfidenceScore)
		value.ClosedAt = cloneTime(value.ClosedAt)
		value.UpdatedAt = value.UpdatedAt.UTC()
		return value
	case InvestigativeQueryVariant:
		value.CreatedAt = value.CreatedAt.UTC()
		return value
	case ForensicKeywordVariant:
		value.CreatedAt = value.CreatedAt.UTC()
		return value
	default:
		return nil
	}
}

func blank(value *string) bool { return value == nil || *value == "" }

func optionalPair(timestamp *time.Time, day *string) bool {
	return timestamp == nil && day == nil || timestamp != nil && day != nil && *day != ""
}

func cloneString(value *string) *string {
	if value == nil {
		return nil
	}
	result := *value
	return &result
}

func cloneUUID(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	result := *value
	return &result
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	result := value.UTC()
	return &result
}

func cloneInt(value *int) *int {
	if value == nil {
		return nil
	}
	result := *value
	return &result
}
