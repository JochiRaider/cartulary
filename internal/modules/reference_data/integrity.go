package reference_data

import (
	"context"
	"errors"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// IntegrityOptions makes consumer-detected loss an application operation even
// when operator administration is disabled. No Job or trusted clock is needed.
type IntegrityOptions struct {
	Observer  OperationObserver
	Finalizer ActionFinalizer
	Limits    Limits
}

type integrityService struct {
	*referenceDependencies
	actionFinalizer ActionFinalizer
}

func newIntegrityService(pool postgres.DB, storage ArtifactStorage, now func() time.Time, options IntegrityOptions) (*integrityService, error) {
	if pool == nil || storage == nil || now == nil || options.Finalizer == nil || options.Limits.ReferencePacks.MaxVerificationSeconds < 60 || options.Limits.ReferencePacks.MaxVerificationSeconds > 86400 {
		return nil, errors.New("reference pack: incomplete integrity dependencies")
	}
	return &integrityService{referenceDependencies: &referenceDependencies{observer: options.Observer, pool: pool, storage: storage, now: now, limits: options.Limits}, actionFinalizer: options.Finalizer}, nil
}

// InvalidateUnavailablePack rechecks physical content outside publication
// locks. A stale observation cannot condemn a version restored concurrently.
// The first committed detection owns the attestation; later reads are no-ops.
func (c *integrityService) InvalidateUnavailablePack(ctx context.Context, key, version string) (resultErr error) {
	ctx, end := observeReferenceOperation(ctx, c.observer, "reference_pack.invalidate")
	defer func() { end(referenceOutcome(resultErr)) }()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(c.limits.ReferencePacks.MaxVerificationSeconds)*time.Second)
	defer cancel()
	if !validConsumerKey(key) || version == "" || len(version) > 128 {
		return consumerError("invalid_pack_request")
	}
	var revision int64
	var health string
	var failure *string
	var removed bool
	err := c.pool.QueryRow(ctx, `SELECT k.revision,c.health,c.last_failure_code,c.removed FROM reference_pack_candidates c JOIN reference_pack_key_state k USING(pack_key) WHERE c.pack_key=$1 AND c.pack_version=$2 AND c.current_envelope_id IS NOT NULL`, key, version).Scan(&revision, &health, &failure, &removed)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if removed || health == "missing" || health == "failed" && (failure == nil || *failure != "metadata_expired") {
		return nil
	}
	err = c.checkRetainedMembers(ctx, frozenMember{Key: key, Version: version})
	if err == nil {
		return nil
	}
	var rejected *ContentRejection
	if !errors.As(err, &rejected) {
		return err
	}
	at := c.now().UTC()
	operation := uuid.New()
	tx, err := c.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	frozen, err := admitOperationTx(ctx, tx, operationAdmission{ID: operation, Kind: "integrity", ActorKind: "system", At: at, Keys: []string{key}, Version: version, TimeoutSeconds: c.limits.ReferencePacks.MaxVerificationSeconds})
	if err != nil {
		return err
	}
	var current int64
	if err := tx.QueryRow(ctx, `SELECT revision FROM reference_pack_key_state WHERE pack_key=$1`, key).Scan(&current); err != nil {
		return err
	}
	if revision != current {
		return &OperationRejection{Reason: "stale_admission_state"}
	}
	health = "failed"
	var missing *string
	if rejected.Code == "payload_missing" {
		health = "missing"
		reason := "storage_loss"
		missing = &reason
	}
	if _, err := tx.Exec(ctx, `UPDATE reference_pack_candidates SET health=$3,last_failure_code=$4,missing_reason=$5 WHERE pack_key=$1 AND pack_version=$2`, key, version, health, rejected.Code, missing); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE reference_pack_key_state SET revision=revision+1 WHERE pack_key=$1`, key); err != nil {
		return err
	}
	if err := publishFallbackTx(ctx, tx, executionAttempt{OperationID: operation, Frozen: frozen}, at); err != nil {
		return err
	}
	var next *string
	if err := tx.QueryRow(ctx, `SELECT pack_set_id FROM reference_pack_current_set WHERE singleton`).Scan(&next); err != nil {
		return err
	}
	event, err := appendPackAttestationTx(ctx, tx, attestationInput{Operation: operation, Key: key, Version: version, Kind: "payload_invalidation", Result: "failed", At: at, PreviousSet: frozen.previousSetID(), ResultingSet: next})
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE reference_pack_operations SET terminal_at=$2,final_outcome=$3 WHERE operation_id=$1`, operation, at, event); err != nil {
		return err
	}
	return c.actionFinalizer.FinalizeReferencePackAction(ctx, tx, func(proofCtx context.Context) (bool, error) {
		var committed bool
		err := c.pool.QueryRow(proofCtx, `SELECT terminal_at IS NOT NULL AND final_outcome=$2 FROM reference_pack_operations WHERE operation_id=$1`, operation, event).Scan(&committed)
		return committed, err
	})
}

// A consumer participating in a source transaction must release that complete
// transaction before the independent integrity mutation can acquire its guards.
// Definitive content loss aborts the source operation; it cannot commit after
// the consumer returns failure. No source-owner changes are committed here.
func (c *integrityService) invalidateAfterRollback(tx pgx.Tx) func(context.Context, string, string) error {
	return func(ctx context.Context, key, version string) error {
		if err := tx.Rollback(context.WithoutCancel(ctx)); err != nil {
			return err
		}
		return c.InvalidateUnavailablePack(ctx, key, version)
	}
}
