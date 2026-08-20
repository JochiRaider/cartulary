package mentioneffects

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/rowpresenter"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/versionid"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

var ErrSourceRecordNotFound = errors.New("timeline mention effects: source record not found")

type RecordPort interface {
	sourcerepository.EnvelopeReader
	AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error)
	LoadRowVersionTx(context.Context, pgx.Tx, uuid.UUID) (int64, error)
}

type CollectionPort interface {
	LoadTimelineCollectionFactsTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (workbookprojection.CollectionFacts, error)
}

type Provider struct {
	records          RecordPort
	collections      CollectionPort
	projections      workbookprojection.Writer
	source           *sourcerepository.Repository
	projectionSource *workbookprojection.Source
}

func NewProvider(records RecordPort, collections CollectionPort, projections workbookprojection.Writer) *Provider {
	return &Provider{
		records:          records,
		collections:      collections,
		projections:      projections,
		source:           sourcerepository.New(records),
		projectionSource: workbookprojection.NewSource(records, collections),
	}
}

type ActionState struct {
	SourceRecordID  uuid.UUID
	RowVersion      int64
	BeforeVersionID string
	BeforeRow       map[string]any
}

type ActionCommand struct {
	IncidentID  uuid.UUID
	ActorUserID uuid.UUID
	EffectiveAt time.Time
}

type ActionResult struct {
	SourceRecordID  uuid.UUID
	RowVersion      int64
	BeforeVersionID string
	AfterVersionID  string
	BeforeRow       map[string]any
	AfterRow        map[string]any
}

func (p *Provider) PrepareMentionActionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (ActionState, error) {
	record, row, err := p.loadRecordAndRowTx(ctx, tx, recordID)
	if err != nil {
		return ActionState{}, err
	}
	return ActionState{
		SourceRecordID:  record.RecordID,
		RowVersion:      record.RowVersion,
		BeforeVersionID: versionid.Format(record.RecordID, record.RowVersion),
		BeforeRow:       row,
	}, nil
}

func (p *Provider) ApplyMentionActionEffectsTx(ctx context.Context, tx pgx.Tx, state ActionState, command ActionCommand) (ActionResult, error) {
	if p == nil || p.records == nil || p.collections == nil || p.projections == nil || p.source == nil {
		return ActionResult{}, errors.New("timeline mention effect dependencies are required")
	}
	rowVersion, err := p.records.AdvanceVersionTx(ctx, tx, state.SourceRecordID, command.ActorUserID, command.EffectiveAt)
	if err != nil {
		return ActionResult{}, err
	}
	if _, err := tx.Exec(ctx, `
UPDATE timeline_events
   SET row_version = $2,
       edited_at = $3,
       updated_by_user_id = $4
 WHERE record_id = $1
`, state.SourceRecordID, rowVersion, command.EffectiveAt.UTC(), command.ActorUserID); err != nil {
		return ActionResult{}, fmt.Errorf("persist mention source record: %w", err)
	}
	record, afterRow, err := p.loadRecordAndRowTx(ctx, tx, state.SourceRecordID)
	if err != nil {
		return ActionResult{}, err
	}
	derived := workbookprojection.Derive(record, nil)
	facts, err := p.collections.LoadTimelineCollectionFactsTx(ctx, tx, derived.IncidentID, derived.RecordID)
	if err != nil {
		return ActionResult{}, err
	}
	workbookprojection.ApplyCollectionFacts(&derived, facts)
	if err := p.projections.ApplyTimelineMutationTx(ctx, tx, workbookprojection.ProjectionMutation{
		Kind:     workbookprojection.ProjectionMutationUpsert,
		RecordID: state.SourceRecordID,
		Input:    derived.ProjectionInput(),
	}); err != nil {
		return ActionResult{}, err
	}
	return ActionResult{
		SourceRecordID:  state.SourceRecordID,
		RowVersion:      rowVersion,
		BeforeVersionID: state.BeforeVersionID,
		AfterVersionID:  versionid.Format(state.SourceRecordID, rowVersion),
		BeforeRow:       state.BeforeRow,
		AfterRow:        afterRow,
	}, nil
}

func (p *Provider) LoadTimelineInvalidationsTx(ctx context.Context, tx pgx.Tx, fieldKeysByRecord map[uuid.UUID][]string) ([]TimelineInvalidation, error) {
	if len(fieldKeysByRecord) == 0 {
		return []TimelineInvalidation{}, nil
	}
	recordIDs := make([]uuid.UUID, 0, len(fieldKeysByRecord))
	for recordID := range fieldKeysByRecord {
		recordIDs = append(recordIDs, recordID)
	}
	slices.SortFunc(recordIDs, func(left uuid.UUID, right uuid.UUID) int {
		return strings.Compare(left.String(), right.String())
	})
	result := make([]TimelineInvalidation, 0, len(recordIDs))
	for _, recordID := range recordIDs {
		rowVersion, err := p.records.LoadRowVersionTx(ctx, tx, recordID)
		if err != nil {
			return nil, fmt.Errorf("load record invalidation row_version: %w", err)
		}
		fieldKeys := append([]string(nil), fieldKeysByRecord[recordID]...)
		slices.Sort(fieldKeys)
		fieldKeys = slices.Compact(fieldKeys)
		result = append(result, TimelineInvalidation{
			RecordID:         recordID,
			RowVersion:       rowVersion,
			ChangedFieldKeys: fieldKeys,
		})
	}
	return result, nil
}

func (p *Provider) LoadRelationshipInvalidationsTx(ctx context.Context, tx pgx.Tx, linkTypesByRecord map[uuid.UUID][]string) ([]TimelineInvalidation, error) {
	fieldKeysByRecord := make(map[uuid.UUID][]string, len(linkTypesByRecord))
	for recordID, linkTypes := range linkTypesByRecord {
		for _, linkType := range linkTypes {
			switch linkType {
			case "observed_on_host":
				fieldKeysByRecord[recordID] = append(fieldKeysByRecord[recordID], "timeline.host_refs")
			case "observed_as_identity":
				fieldKeysByRecord[recordID] = append(fieldKeysByRecord[recordID], "timeline.identity_refs")
			}
		}
	}
	return p.LoadTimelineInvalidationsTx(ctx, tx, fieldKeysByRecord)
}

func (p *Provider) RefreshTimelineProjectionRowsTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) error {
	if p == nil || p.projectionSource == nil || p.projections == nil {
		return errors.New("timeline mention effect dependencies are required")
	}
	recordIDs = append([]uuid.UUID(nil), recordIDs...)
	slices.SortFunc(recordIDs, func(left uuid.UUID, right uuid.UUID) int {
		return strings.Compare(left.String(), right.String())
	})
	recordIDs = slices.Compact(recordIDs)
	for _, recordID := range recordIDs {
		mutation, err := p.projectionSource.BuildProjectionMutationTx(ctx, tx, recordID)
		if err != nil {
			return fmt.Errorf("build exact timeline projection mutation for %s: %w", recordID, err)
		}
		if err := p.projections.ApplyTimelineMutationTx(ctx, tx, mutation); err != nil {
			return fmt.Errorf("apply exact timeline projection mutation for %s: %w", recordID, err)
		}
	}
	return nil
}

type TimelineInvalidation struct {
	RecordID         uuid.UUID
	RowVersion       int64
	ChangedFieldKeys []string
}

func (p *Provider) loadRecordAndRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (sourcerepository.Snapshot, map[string]any, error) {
	if p == nil || p.source == nil || p.collections == nil {
		return sourcerepository.Snapshot{}, nil, errors.New("timeline mention effect dependencies are required")
	}
	record, err := p.source.LoadTx(ctx, tx, recordID)
	if errors.Is(err, sourcerepository.ErrNotFound) {
		return sourcerepository.Snapshot{}, nil, ErrSourceRecordNotFound
	}
	if err != nil {
		return sourcerepository.Snapshot{}, nil, err
	}
	derived := workbookprojection.Derive(record, nil)
	facts, err := p.collections.LoadTimelineCollectionFactsTx(ctx, tx, derived.IncidentID, derived.RecordID)
	if err != nil {
		return sourcerepository.Snapshot{}, nil, err
	}
	workbookprojection.ApplyCollectionFacts(&derived, facts)
	return record, rowpresenter.BuildRow(derived.PresenterRecord()), nil
}
