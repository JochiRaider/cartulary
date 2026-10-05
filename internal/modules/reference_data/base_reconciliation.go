package reference_data

import (
	"context"
	"errors"
	"slices"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packstate"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/jackc/pgx/v5"
)

type reconciledVersion struct {
	state packstate.Version
	// Only definitive content rejection sets code. Read failures abort startup
	// before the publication transaction and cannot condemn retained content.
	code string
}

type reconciliationCapture struct {
	exists   bool
	revision int64
	keys     map[string]int64
	versions []reconciledVersion
}

// Capture a coherent selection without holding mutation locks while inspecting
// bytes. Startup has the same stale-observation rule as ordinary verification.
func inspectReconciliation(ctx context.Context, pool postgres.DB, storage ArtifactStorage) (reconciliationCapture, error) {
	capture := reconciliationCapture{keys: map[string]int64{}}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return capture, err
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	var setID *string
	err = tx.QueryRow(ctx, `SELECT pack_set_id,revision FROM reference_pack_current_set WHERE singleton`).Scan(&setID, &capture.revision)
	if errors.Is(err, pgx.ErrNoRows) {
		return capture, nil
	}
	if err != nil {
		return capture, err
	}
	capture.exists = true
	keys := slices.Clone(mandatoryRegistryKeys)
	if setID != nil {
		var data []byte
		if err := tx.QueryRow(ctx, `SELECT canonical_set FROM reference_pack_sets WHERE pack_set_id=$1`, *setID).Scan(&data); err != nil {
			return capture, err
		}
		set, err := decodeRetainedSet(data)
		if err != nil {
			return capture, err
		}
		for _, member := range set.Members {
			state, err := loadStateVersionTx(ctx, tx, member.Key, member.Version)
			if err != nil {
				return capture, err
			}
			if PackSetMember(state.Member) != member {
				return capture, errHistoricalIntegrity
			}
			capture.versions = append(capture.versions, reconciledVersion{state: state})
			keys = append(keys, member.Key)
		}
	}
	slices.Sort(keys)
	for _, key := range slices.Compact(keys) {
		var revision int64
		if err := tx.QueryRow(ctx, `SELECT revision FROM reference_pack_key_state WHERE pack_key=$1`, key).Scan(&revision); err != nil {
			return capture, err
		}
		capture.keys[key] = revision
	}
	if err := tx.Commit(ctx); err != nil {
		return capture, err
	}
	for i := range capture.versions {
		version := &capture.versions[i]
		state := &version.state
		if state.Health != packstate.Available || state.Disabled || state.Removed {
			continue
		}
		err := checkRetainedVersion(ctx, pool, storage, state.Member.Key, state.Member.Version)
		if err == nil {
			err = packformat.ValidateRuntimeCompatibility(state.Manifest)
		}
		if err == nil {
			err = packformat.ValidateContent(ctx, state.Manifest, retainedContentSource{pool, storage, state.Member.Key, state.Member.Version}, nil)
		}
		if err == nil {
			continue
		}
		var rejected *ContentRejection
		var failure *packformat.Failure
		switch {
		case errors.As(err, &rejected):
			version.code = rejected.Code
		case errors.As(err, &failure):
			version.code = failure.Code
		default:
			return capture, err
		}
		state.Health, state.MissingReason = packstate.Failed, nil
		if version.code == "payload_missing" {
			reason := "storage_loss"
			state.Health, state.MissingReason = packstate.Missing, &reason
		}
	}
	return capture, nil
}

// All captured keys and the current-set singleton are already locked. Compare
// before publishing any health, Base version, set, audit or attestation effect.
func (capture reconciliationCapture) guard(ctx context.Context, tx pgx.Tx, inserted bool, revision int64) error {
	if capture.exists == inserted || capture.exists && revision != capture.revision {
		return &OperationRejection{Reason: "stale_admission_state"}
	}
	for key, expected := range capture.keys {
		var current int64
		if err := tx.QueryRow(ctx, `SELECT revision FROM reference_pack_key_state WHERE pack_key=$1`, key).Scan(&current); err != nil {
			return err
		}
		if current != expected {
			return &OperationRejection{Reason: "stale_admission_state"}
		}
	}
	return nil
}
