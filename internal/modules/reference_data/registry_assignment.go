package reference_data

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var mandatoryRegistryKeys = []string{"type_registry.evidence", "type_registry.host", "type_registry.indicator"}

// RegistryAssignments participates in a source owner's existing transaction.
// Acquire before locking source records. A multi-owner mutation acquires all
// mandatory registry guards together, in publication order, so participants
// cannot invert the order by assigning different kinds of record in turn.
type RegistryAssignments interface {
	BeginTx(context.Context, pgx.Tx) (RegistryAssignment, error)
}

// RegistryAssignment exposes the same exact-set consumer operations over the
// guarded transaction. RecordUsage is required after adding or changing a type
// assignment; unchanged identities and unrelated edits do not advance usage.
// Neither the session nor its consumer may escape the calling transaction.
// Definitive authoritative content loss rolls back the source transaction
// before publishing integrity invalidation, and returns pack_unavailable.
type RegistryAssignment interface {
	Consumer
	RecordUsage(context.Context, string) error
}

type registryAssignments struct {
	storage   ArtifactStorage
	codec     *pagination.Codec
	now       func() time.Time
	integrity *Coordinator
}

type registryAssignment struct {
	Consumer
	tx       pgx.Tx
	recorded map[string]bool
}

func NewRegistryAssignments(pool *pgxpool.Pool, storage ArtifactStorage, codec *pagination.Codec, now func() time.Time, integrity IntegrityOptions) (RegistryAssignments, error) {
	if storage == nil || codec == nil || now == nil {
		return nil, errors.New("reference pack: incomplete registry assignment dependencies")
	}
	coordinator, err := newIntegrityCoordinator(pool, storage, now, integrity)
	if err != nil {
		return nil, err
	}
	return &registryAssignments{storage: storage, codec: codec, now: now, integrity: coordinator}, nil
}

func (r *registryAssignments) BeginTx(ctx context.Context, tx pgx.Tx) (RegistryAssignment, error) {
	if tx == nil {
		return nil, consumerError("pack_unavailable")
	}
	for _, key := range mandatoryRegistryKeys {
		var revision int64
		if err := tx.QueryRow(ctx, `SELECT revision FROM reference_pack_key_state WHERE pack_key=$1 FOR UPDATE`, key).Scan(&revision); err != nil {
			return nil, err
		}
	}
	var setID *string
	if err := tx.QueryRow(ctx, `SELECT pack_set_id FROM reference_pack_current_set WHERE singleton FOR UPDATE`).Scan(&setID); err != nil {
		return nil, err
	}
	if setID == nil {
		return nil, consumerError("pack_unavailable")
	}
	for _, key := range mandatoryRegistryKeys {
		var revision int64
		if err := tx.QueryRow(ctx, `SELECT revision FROM reference_pack_registry_usage WHERE pack_key=$1 FOR UPDATE`, key).Scan(&revision); err != nil {
			return nil, err
		}
	}
	return &registryAssignment{
		Consumer: observeConsumer(newPackConsumer(&canonicalRepository{pool: tx, storage: r.storage, invalidate: r.integrity.invalidateAfterRollback(tx)}, r.codec, r.now), r.integrity.observer),
		tx:       tx, recorded: map[string]bool{},
	}, nil
}

func (r *registryAssignment) RecordUsage(ctx context.Context, key string) error {
	if !slices.Contains(mandatoryRegistryKeys, key) {
		return consumerError("invalid_pack_request")
	}
	if r.recorded[key] {
		return nil
	}
	result, err := r.tx.Exec(ctx, `UPDATE reference_pack_registry_usage SET revision=revision+1 WHERE pack_key=$1`, key)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return consumerError("pack_unavailable")
	}
	r.recorded[key] = true
	return nil
}
