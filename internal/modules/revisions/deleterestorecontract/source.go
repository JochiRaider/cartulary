package deleterestorecontract

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type StateTransitionKind string

const (
	StateTransitionDelete  StateTransitionKind = "delete"
	StateTransitionRestore StateTransitionKind = "restore"
)

type StateTransitionRequest struct {
	Kind       StateTransitionKind
	IncidentID uuid.UUID
	RecordID   uuid.UUID
}

type ActiveIdentifierConflict struct {
	EntityType       string
	IdentifierClass  string
	NormalizedValue  string
	BlockingRecordID uuid.UUID
}

type StateTransitionBlock struct {
	ReasonCode               string
	ActiveIdentifierConflict *ActiveIdentifierConflict
	ConflictingFieldKeys     []string
}

type StateTransitionPreparation struct {
	Blocked *StateTransitionBlock
}

// DeleteRestoreSource is the consumer-owned port through which Revisions asks
// a record-type owner to expose its source-specific delete/restore behavior.
// Implementations must use fixed SQL owned by the source module.
type DeleteRestoreSource interface {
	SnapshotTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error)
	UpdateSourceDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) error
	ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error)
	PrepareStateTransitionTx(context.Context, pgx.Tx, StateTransitionRequest) (StateTransitionPreparation, error)
}

// EnvelopeMirrorSource is implemented only by sources that persist selected
// record-envelope fields as authoritative mirrors.
type EnvelopeMirrorSource interface {
	SyncEnvelopeMirrorTx(context.Context, pgx.Tx, uuid.UUID) error
}

// ScanSnapshot decodes the stable record/source snapshot shape returned by a
// source owner's fixed query.
func ScanSnapshot(row pgx.Row) (map[string]any, error) {
	var raw []byte
	if err := row.Scan(&raw); err != nil {
		return nil, err
	}
	var snapshot map[string]any
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}
