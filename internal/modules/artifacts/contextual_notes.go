package artifacts

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records"
)

type ContextualNoteCreateCommand struct {
	ActorUserID    uuid.UUID
	SourceRecordID uuid.UUID
	Admission      ContextualNoteAdmission
	RequestID      string
	Now            time.Time
}

func (f *MutationFacade) CreateContextualNote(
	ctx context.Context,
	command ContextualNoteCreateCommand,
) (MutationResult, error) {
	if f == nil {
		return MutationResult{}, fmt.Errorf("artifacts: mutation facade is not configured")
	}
	if !command.Admission.valid() {
		return MutationResult{}, &ValidationError{Field: "payload", ReasonCode: "invalid_value"}
	}
	request := command.Admission.requestValue()
	admission := CreateAdmission{request: request, admitted: true}
	copy(admission.hash[:], command.Admission.requestHash(command.SourceRecordID))
	return f.create(ctx, CreateCommand{
		ActorUserID: command.ActorUserID,
		Admission:   admission,
		RequestID:   command.RequestID,
		Now:         command.Now,
	}, &command.SourceRecordID)
}

// contextIncidentTx classifies the linked-note source record through Records
// under FOR UPDATE (REQ-01-705): a missing record reports
// records.ErrTargetNotFound and a soft-deleted one records.ErrTargetDeleted,
// before its type and the incident lifecycle. The caller's exact-replay lookup
// precedes this classification.
func (f *MutationFacade) contextIncidentTx(
	ctx context.Context,
	tx pgx.Tx,
	sourceRecordID uuid.UUID,
) (uuid.UUID, error) {
	envelope, err := f.recordEnvelopes.ClassifyTargetTx(ctx, tx, sourceRecordID, records.TargetLockUpdate)
	if err != nil {
		return uuid.UUID{}, err
	}
	if !slices.Contains(
		[]string{"timeline_event", "host", "identity", "evidence"},
		envelope.RecordType,
	) {
		return uuid.UUID{}, pgx.ErrNoRows
	}
	return envelope.IncidentID, nil
}
