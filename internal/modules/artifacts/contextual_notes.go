package artifacts

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicttokens"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type ContextualNoteCreateRequest struct {
	ClientTxnID string
	Values      map[string]FieldValue
	Collections map[string]WorkbookCollectionActionPayload
}

type ContextualNoteCreateCommand struct {
	Actor          authn.UserRecord
	SourceRecordID uuid.UUID
	Request        ContextualNoteCreateRequest
	RequestHash    []byte
	RequestID      string
	RouteKey       string
	Now            time.Time
}

type ContextualNoteMutationResult = WorkbookMutationResult

type ContextualNoteFacade struct {
	owner *WorkbookFacade
}

func NewContextualNoteFacade(pool postgres.DB, appender *revisions.Appender) *ContextualNoteFacade {
	return &ContextualNoteFacade{
		owner: NewWorkbookFacade(pool, conflicttokens.ConflictTokenCodec{}, appender),
	}
}

func (f *ContextualNoteFacade) SourceIncident(
	ctx context.Context,
	sourceRecordID uuid.UUID,
) (uuid.UUID, error) {
	tx, err := f.owner.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return uuid.UUID{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	incidentID, err := f.owner.contextIncidentTx(ctx, tx, sourceRecordID)
	if err != nil {
		return uuid.UUID{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.UUID{}, err
	}
	return incidentID, nil
}

func (f *ContextualNoteFacade) Create(
	ctx context.Context,
	command ContextualNoteCreateCommand,
) (ContextualNoteMutationResult, error) {
	if f == nil || f.owner == nil {
		return ContextualNoteMutationResult{}, fmt.Errorf("artifacts: contextual note facade is not configured")
	}
	request := WorkbookCreateRequest{
		ViewSchemaID: NotesViewSchemaID,
		ClientTxnID:  command.Request.ClientTxnID,
		Values:       command.Request.Values,
		Collections:  command.Request.Collections,
	}
	return f.owner.create(ctx, WorkbookCreateCommand{
		Actor:       command.Actor,
		Request:     request,
		RequestHash: command.RequestHash,
		RequestID:   command.RequestID,
		RouteKey:    command.RouteKey,
		Now:         command.Now,
	}, &command.SourceRecordID)
}

func (f *WorkbookFacade) contextIncidentTx(
	ctx context.Context,
	tx pgx.Tx,
	sourceRecordID uuid.UUID,
) (uuid.UUID, error) {
	envelope, err := f.source.records.LoadEnvelopeTx(ctx, tx, sourceRecordID, false)
	if errors.Is(err, records.ErrEnvelopeNotFound) {
		return uuid.UUID{}, pgx.ErrNoRows
	}
	if err != nil {
		return uuid.UUID{}, err
	}
	if envelope.DeletedAt != nil || !slices.Contains(
		[]string{"timeline_event", "host", "identity", "evidence"},
		envelope.RecordType,
	) {
		return uuid.UUID{}, pgx.ErrNoRows
	}
	return envelope.IncidentID, nil
}
