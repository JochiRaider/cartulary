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
)

type ContextualNoteCreateRequest struct {
	ClientTxnID string
	Values      map[string]FieldValue
	Collections map[string]CollectionActionPayload
}

type ContextualNoteCreateCommand struct {
	ActorUserID    uuid.UUID
	SourceRecordID uuid.UUID
	Request        ContextualNoteCreateRequest
	RequestHash    []byte
	RequestID      string
	OperationID    OperationID
	Now            time.Time
}

func (f *MutationFacade) SourceIncident(
	ctx context.Context,
	sourceRecordID uuid.UUID,
) (uuid.UUID, error) {
	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return uuid.UUID{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	incidentID, err := f.contextIncidentTx(ctx, tx, sourceRecordID)
	if err != nil {
		return uuid.UUID{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.UUID{}, err
	}
	return incidentID, nil
}

func (f *MutationFacade) CreateContextualNote(
	ctx context.Context,
	command ContextualNoteCreateCommand,
) (MutationResult, error) {
	if f == nil {
		return MutationResult{}, fmt.Errorf("artifacts: mutation facade is not configured")
	}
	request := CreateRequest{
		ViewSchemaID: NotesViewSchemaID,
		ClientTxnID:  command.Request.ClientTxnID,
		Values:       command.Request.Values,
		Collections:  command.Request.Collections,
	}
	return f.create(ctx, CreateCommand{
		ActorUserID: command.ActorUserID,
		Request:     request,
		RequestHash: command.RequestHash,
		RequestID:   command.RequestID,
		OperationID: command.OperationID,
		Now:         command.Now,
	}, &command.SourceRecordID)
}

func (f *MutationFacade) contextIncidentTx(
	ctx context.Context,
	tx pgx.Tx,
	sourceRecordID uuid.UUID,
) (uuid.UUID, error) {
	envelope, err := f.recordEnvelopes.LoadEnvelopeTx(ctx, tx, sourceRecordID, false)
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
