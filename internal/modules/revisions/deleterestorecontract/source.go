package deleterestorecontract

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
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
	// SnapshotsTx returns the snapshot of each requested record the source
	// owns, keyed by record, with one read however many records are requested.
	// A record it does not own is absent.
	SnapshotsTx(context.Context, pgx.Tx, []uuid.UUID) (map[uuid.UUID]map[string]any, error)
	UpdateSourceDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) error
	ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error)
	PrepareStateTransitionTx(context.Context, pgx.Tx, StateTransitionRequest) (StateTransitionPreparation, error)
	ValidateStateTransitionTx(context.Context, pgx.Tx, StateTransitionRequest) (StateTransitionPreparation, error)
}

// EnvelopeMirrorSource is implemented only by sources that persist selected
// record-envelope fields as authoritative mirrors.
type EnvelopeMirrorSource interface {
	SyncEnvelopeMirrorTx(context.Context, pgx.Tx, uuid.UUID) error
}

// CollectSnapshots reads the rows of a source owner's fixed snapshot query,
// each a record_id and its record/source snapshot object, keyed by record.
func CollectSnapshots(rows pgx.Rows, err error) (map[uuid.UUID]map[string]any, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	snapshots := make(map[uuid.UUID]map[string]any)
	for rows.Next() {
		var recordID uuid.UUID
		var raw []byte
		if err := rows.Scan(&recordID, &raw); err != nil {
			return nil, err
		}
		snapshot, err := DecodeSnapshot(raw)
		if err != nil {
			return nil, err
		}
		snapshots[recordID] = snapshot
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return snapshots, nil
}

// SnapshotOf returns one record's snapshot from an owner's batch read, or
// pgx.ErrNoRows when the owner has no such record.
func SnapshotOf(snapshots map[uuid.UUID]map[string]any, recordID uuid.UUID, err error) (map[string]any, error) {
	if err != nil {
		return nil, err
	}
	snapshot, found := snapshots[recordID]
	if !found {
		return nil, pgx.ErrNoRows
	}
	return snapshot, nil
}

// DecodeSnapshot decodes the stable record/source snapshot object a source
// owner's fixed query returns, keeping exact integers.
func DecodeSnapshot(raw []byte) (map[string]any, error) {
	var snapshot map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&snapshot); err != nil {
		return nil, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		if err == nil {
			return nil, errors.New("snapshot has more than one JSON value")
		}
		return nil, err
	}
	return snapshot, nil
}
