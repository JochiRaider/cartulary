package deleterestore

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
)

type HostSource struct{}

var _ deleterestorecontract.DeleteRestoreSource = HostSource{}

func NewHostSource() HostSource {
	return HostSource{}
}

func (HostSource) SnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return deleterestorecontract.ScanSnapshot(tx.QueryRow(ctx, `
SELECT jsonb_build_object('record', to_jsonb(r), 'source', to_jsonb(h))
  FROM records r
  JOIN hosts h
    ON h.record_id = r.record_id
 WHERE r.record_id = $1
`, recordID))
}

func (HostSource) UpdateSourceDeleteStateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, _ uuid.UUID, _ time.Time, _ bool) error {
	return (HostSource{}).SyncEnvelopeMirrorTx(ctx, tx, recordID)
}

func (HostSource) SyncEnvelopeMirrorTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	tag, err := tx.Exec(ctx, `
UPDATE hosts h
   SET row_version = r.row_version,
       updated_at = r.updated_at,
       updated_by_user_id = r.updated_by_user_id
  FROM records r
 WHERE r.record_id = $1
   AND h.record_id = r.record_id
`, recordID)
	if err != nil {
		return fmt.Errorf("synchronize host envelope mirrors: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("synchronize host envelope mirrors: updated %d rows, want 1", tag.RowsAffected())
	}
	return nil
}

func (HostSource) ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error) {
	return "cartulary.view.hosts.v1", nil
}

func (HostSource) PrepareStateTransitionTx(ctx context.Context, tx pgx.Tx, request deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	return prepareEntityStateTransitionTx(ctx, tx, request, "host")
}

type IdentitySource struct{}

var _ deleterestorecontract.DeleteRestoreSource = IdentitySource{}

func NewIdentitySource() IdentitySource {
	return IdentitySource{}
}

func (IdentitySource) SnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return deleterestorecontract.ScanSnapshot(tx.QueryRow(ctx, `
SELECT jsonb_build_object('record', to_jsonb(r), 'source', to_jsonb(i))
  FROM records r
  JOIN identities i
    ON i.record_id = r.record_id
 WHERE r.record_id = $1
`, recordID))
}

func (IdentitySource) UpdateSourceDeleteStateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, _ uuid.UUID, _ time.Time, _ bool) error {
	return (IdentitySource{}).SyncEnvelopeMirrorTx(ctx, tx, recordID)
}

func (IdentitySource) SyncEnvelopeMirrorTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	tag, err := tx.Exec(ctx, `
UPDATE identities i
   SET row_version = r.row_version,
       updated_at = r.updated_at,
       updated_by_user_id = r.updated_by_user_id
  FROM records r
 WHERE r.record_id = $1
   AND i.record_id = r.record_id
`, recordID)
	if err != nil {
		return fmt.Errorf("synchronize identity envelope mirrors: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("synchronize identity envelope mirrors: updated %d rows, want 1", tag.RowsAffected())
	}
	return nil
}

func (IdentitySource) ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error) {
	return "cartulary.view.identities.v1", nil
}

func (IdentitySource) PrepareStateTransitionTx(ctx context.Context, tx pgx.Tx, request deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	return prepareEntityStateTransitionTx(ctx, tx, request, "identity")
}

func prepareEntityStateTransitionTx(ctx context.Context, tx pgx.Tx, request deleterestorecontract.StateTransitionRequest, entityType string) (deleterestorecontract.StateTransitionPreparation, error) {
	releasing := false
	switch request.Kind {
	case deleterestorecontract.StateTransitionDelete:
		releasing = true
	case deleterestorecontract.StateTransitionRestore:
	default:
		return deleterestorecontract.StateTransitionPreparation{}, fmt.Errorf("unsupported Entities state transition %q", request.Kind)
	}
	conflict, err := hostidentity.PrepareActiveIdentifierStateTransitionTx(
		ctx,
		tx,
		request.IncidentID,
		entityType,
		request.RecordID,
		releasing,
	)
	if err != nil {
		return deleterestorecontract.StateTransitionPreparation{}, err
	}
	if conflict == nil {
		return deleterestorecontract.StateTransitionPreparation{}, nil
	}
	return deleterestorecontract.StateTransitionPreparation{Blocked: &deleterestorecontract.StateTransitionBlock{
		ReasonCode: "active_entity_identifier_conflict",
		ActiveIdentifierConflict: &deleterestorecontract.ActiveIdentifierConflict{
			EntityType:       entityType,
			IdentifierClass:  conflict.IdentifierClass,
			NormalizedValue:  conflict.NormalizedValue,
			BlockingRecordID: conflict.BlockingRecordID,
		},
	}}, nil
}
