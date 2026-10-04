package reference_data

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SetBinding is retained with a consumer artifact. Provenance is the immutable
// first-success anchor of the set, never the current administrative envelope.
type SetBinding struct {
	SetID      string           `json:"pack_set_id"`
	SHA256     string           `json:"pack_set_sha256"`
	Provenance []PackProvenance `json:"provenance"`
}

// Retention participates in the caller's admission transaction. Owners never
// inspect Reference Data tables, storage references or index generations.
type Retention interface {
	CaptureCurrentTx(context.Context, pgx.Tx, string, string, uuid.UUID) (SetBinding, error)
	CaptureRetainedTx(context.Context, pgx.Tx, string, string, string, uuid.UUID) (SetBinding, error)
	ExportReferencesTx(context.Context, pgx.Tx, []string) ([]byte, error)
	ValidateBinding(context.Context, SetBinding) error
}

type retention struct {
	repository *canonicalRepository
	integrity  *Coordinator
}

func NewRetention(pool *pgxpool.Pool, storage ArtifactStorage, now func() time.Time, integrity IntegrityOptions) (Retention, error) {
	if pool == nil || storage == nil {
		return nil, errors.New("reference pack: incomplete retention dependencies")
	}
	coordinator, err := newIntegrityCoordinator(pool, storage, now, integrity)
	if err != nil {
		return nil, err
	}
	return &retention{repository: &canonicalRepository{pool: pool, storage: storage, invalidate: coordinator.InvalidateUnavailablePack}, integrity: coordinator}, nil
}
func (r *retention) CaptureCurrentTx(ctx context.Context, tx pgx.Tx, ownerKind, ownerID string, operationID uuid.UUID) (SetBinding, error) {
	set, err := PinCurrentTx(ctx, tx, ownerKind, ownerID, operationID)
	if err != nil {
		return SetBinding{}, err
	}
	return r.bindingTx(ctx, tx, set)
}

// CaptureRetainedTx pins an exact historical set. It does not resolve the
// current set or change deployment-wide activation. Removal and pin creation
// serialize on the same ordered pack-key guards.
func (r *retention) CaptureRetainedTx(ctx context.Context, tx pgx.Tx, setID, ownerKind, ownerID string, operationID uuid.UUID) (SetBinding, error) {
	if tx == nil || !setIDPattern.MatchString(setID) {
		return SetBinding{}, consumerError("invalid_pack_request")
	}
	repository := canonicalRepository{pool: tx}
	set, err := repository.RetainedSet(ctx, setID)
	if err != nil {
		return SetBinding{}, err
	}
	for _, member := range set.Members {
		var revision int64
		if err := tx.QueryRow(ctx, `SELECT revision FROM reference_pack_key_state WHERE pack_key=$1 FOR UPDATE`, member.Key).Scan(&revision); err != nil {
			return SetBinding{}, err
		}
	}
	// Check availability before adding the new pin, so that the new pin cannot
	// itself authorize a previously unpinned invalid version.
	binding, err := r.bindingTx(ctx, tx, set)
	if err != nil {
		return SetBinding{}, err
	}
	if err := insertPinTx(ctx, tx, set.ID, ownerKind, ownerID, operationID); err != nil {
		return SetBinding{}, err
	}
	return binding, nil
}

func (r *retention) bindingTx(ctx context.Context, tx pgx.Tx, set PackSet) (SetBinding, error) {
	var data []byte
	if err := tx.QueryRow(ctx, `SELECT canonical_provenance FROM reference_pack_sets WHERE pack_set_id=$1`, set.ID).Scan(&data); err != nil {
		return SetBinding{}, err
	}
	binding := SetBinding{SetID: set.ID, SHA256: set.SHA256}
	var err error
	if binding.Provenance, err = decodeRetainedProvenance(data, set); err != nil {
		return SetBinding{}, err
	}
	validator := retention{repository: &canonicalRepository{pool: tx, storage: r.repository.storage, invalidate: r.integrity.invalidateAfterRollback(tx)}}
	if err := validator.ValidateBinding(ctx, binding); err != nil {
		return SetBinding{}, err
	}
	return binding, nil
}
func (r *retention) ValidateBinding(ctx context.Context, binding SetBinding) error {
	invalid := consumerError("pack_unavailable")
	if !setIDPattern.MatchString(binding.SetID) || binding.Provenance == nil {
		return invalid
	}
	set, err := r.repository.RetainedSet(ctx, binding.SetID)
	if err != nil {
		return err
	}
	if set.SHA256 != binding.SHA256 || len(set.Members) != len(binding.Provenance) {
		return invalid
	}
	for i, m := range set.Members {
		p, err := r.repository.Provenance(ctx, set.ID, m.Key)
		if err != nil {
			return err
		}
		expected, err := canonicaljson.Marshal(p)
		if err != nil {
			return err
		}
		actual, err := canonicaljson.Marshal(binding.Provenance[i])
		if err != nil {
			return err
		}
		if !bytes.Equal(expected, actual) {
			return invalid
		}
	}
	return nil
}
