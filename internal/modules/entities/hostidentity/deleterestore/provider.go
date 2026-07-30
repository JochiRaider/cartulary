package deleterestore

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

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

func (HostSource) UpdateSourceDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) error {
	return nil
}

func (HostSource) ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error) {
	return "cartulary.view.hosts.v1", nil
}

func (HostSource) ValidateDeletePreconditionsTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (string, bool, error) {
	return "", false, nil
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

func (IdentitySource) UpdateSourceDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) error {
	return nil
}

func (IdentitySource) ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error) {
	return "cartulary.view.identities.v1", nil
}

func (IdentitySource) ValidateDeletePreconditionsTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (string, bool, error) {
	return "", false, nil
}
