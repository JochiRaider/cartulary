package timeline

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

var (
	ErrRecordNotFound         = errors.New("timeline: record not found")
	ErrRowVersionConflict     = errors.New("timeline: row version conflict")
	ErrIllegalTransition      = errors.New("timeline: illegal transition")
	ErrNoEffectiveChange      = errors.New("timeline: no effective change")
	ErrIncidentClosed         = errors.New("timeline: incident closed")
	ErrRecordDeleted          = errors.New("timeline: record deleted use restore")
	ErrResolvedRecordNotFound = errors.New("timeline: resolved record not found")
)

type RowVersionConflictError struct {
	RecordID          uuid.UUID
	BaseRowVersion    int64
	CurrentRowVersion int64
}

func (e *RowVersionConflictError) Error() string {
	return ErrRowVersionConflict.Error()
}

func (e *RowVersionConflictError) Unwrap() error {
	return ErrRowVersionConflict
}

func (e *RowVersionConflictError) Details() map[string]any {
	if e == nil {
		return map[string]any{}
	}
	return map[string]any{
		"record_id":           e.RecordID.String(),
		"base_row_version":    e.BaseRowVersion,
		"current_row_version": e.CurrentRowVersion,
	}
}

type store struct {
	pool              postgres.DB
	idempotencyStore  IdempotencyPort
	incidentAccess    IncidentPort
	recordStore       RecordPort
	revisionsStore    RevisionPort
	projectionStore   workbookprojection.Writer
	entityProjections EntityProjectionPort
	linkStore         LinkPort
	mentionStore      MentionPort
	entityStore       EntityPort
	evidenceStore     EvidencePort
	collectionReader  CollectionFactPort
	collaboration     CollaborationPort
	sourceRepository  *sourcerepository.Repository
	conflictTokens    conflicttokens.ConflictTokenCodec
	conflictSnapshots conflicttokens.RevisionSnapshotProjector
}

type attachedEvidenceMutation struct {
	RecordLinkID uuid.UUID
	Operation    string
	BeforeValue  map[string]any
	AfterValue   map[string]any
}

type recordTagMutation struct {
	RecordTagID uuid.UUID
	RecordID    uuid.UUID
	Operation   string
	BeforeValue map[string]any
	AfterValue  map[string]any
}

type TimeConversionProfile struct {
	IncidentID         uuid.UUID
	Enabled            bool
	LocalOffsetMinutes *int
	LocalLabel         *string
	ProfileVersion     int64
	UpdatedAt          time.Time
	UpdatedByUserID    *uuid.UUID
}

func newStore(pool postgres.DB, collaborators Collaborators, conflictTokens conflicttokens.ConflictTokenCodec) *store {
	return &store{
		pool:              pool,
		idempotencyStore:  collaborators.Core.Idempotency,
		incidentAccess:    collaborators.Core.Incidents,
		recordStore:       collaborators.Core.Records,
		revisionsStore:    collaborators.Core.Revisions,
		projectionStore:   collaborators.Commit.Projection,
		entityProjections: collaborators.Commit.EntityProjection,
		linkStore:         collaborators.Collections.Links,
		mentionStore:      collaborators.Collections.Mentions,
		entityStore:       collaborators.Collections.Entities,
		evidenceStore:     collaborators.Collections.Evidence,
		collectionReader:  collaborators.Collections.Facts,
		collaboration:     collaborators.Commit.Collaboration,
		sourceRepository:  sourcerepository.New(collaborators.Core.Records),
		conflictTokens:    conflictTokens,
		conflictSnapshots: newTimelineConflictSnapshotProjector(),
	}
}

func (s *store) getRecordIncident(ctx context.Context, recordID uuid.UUID) (uuid.UUID, error) {
	incidentID, err := s.recordStore.ResolveIncident(ctx, recordID)
	if err != nil {
		return uuid.UUID{}, ErrRecordNotFound
	}
	return incidentID, nil
}

func hasMaterialChange(current sourcerepository.Snapshot, next sourcerepository.Snapshot) bool {
	return !stringPointersEqual(current.DateEnteredText, next.DateEnteredText) ||
		!stringPointersEqual(current.AnalystText, next.AnalystText) ||
		!stringPointersEqual(current.MitreStageText, next.MitreStageText) ||
		!stringPointersEqual(current.DeviceObjectText, next.DeviceObjectText) ||
		!stringPointersEqual(current.IPAddressText, next.IPAddressText) ||
		!stringPointersEqual(current.ActivityUTCText, next.ActivityUTCText) ||
		!stringPointersEqual(current.ActivityLocalText, next.ActivityLocalText) ||
		!stringPointersEqual(current.RawActivityText, next.RawActivityText) ||
		!stringPointersEqual(current.ActivitySynopsisText, next.ActivitySynopsisText) ||
		!stringPointersEqual(current.DataSourceText, next.DataSourceText) ||
		current.ActivityUTCGenerated != next.ActivityUTCGenerated ||
		current.ActivityLocalGenerated != next.ActivityLocalGenerated ||
		current.ActivityTimePairState != next.ActivityTimePairState
}

func isRecordLinkConflict(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505"
}

func extractUUIDFromPayload(payload map[string]any, path ...string) (uuid.UUID, error) {
	current := any(payload)
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return uuid.UUID{}, fmt.Errorf("decode payload path %q", strings.Join(path, "."))
		}
		current = object[segment]
	}
	text, ok := current.(string)
	if !ok || text == "" {
		return uuid.UUID{}, fmt.Errorf("decode payload uuid path %q", strings.Join(path, "."))
	}
	parsed, err := uuid.Parse(text)
	if err != nil {
		return uuid.UUID{}, err
	}
	return parsed, nil
}

func stringPointersEqual(left *string, right *string) bool {
	switch {
	case left == nil && right == nil:
		return true
	case left == nil || right == nil:
		return false
	default:
		return *left == *right
	}
}

func optionalUUIDFromPG(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	parsed := uuid.Must(uuid.FromBytes(value.Bytes[:]))
	return &parsed
}

func optionalTextFromPG(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	text := value.String
	return &text
}

func optionalIntFromPG(value pgtype.Int4) *int {
	if !value.Valid {
		return nil
	}
	parsed := int(value.Int32)
	return &parsed
}
