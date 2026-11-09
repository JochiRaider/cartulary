package reference_data

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/contractadmission"
	"github.com/jackc/pgx/v5"
)

// AdmitRestoredState checks authoritative bindings and retained logical bytes
// inside Recovery's still-uncommitted transaction. It does not rebuild indexes,
// reinterpret historical freshness, publish availability or repair any bytes.
func AdmitRestoredState(ctx context.Context, tx pgx.Tx, storage VerificationStorage) error {
	if tx == nil || storage == nil {
		return errors.New("reference pack: restore admission dependencies required")
	}
	if err := contractadmission.Validate(ctx, tx); err != nil {
		return err
	}
	afterKey, afterVersion := "", ""
	for {
		var key, version string
		err := tx.QueryRow(ctx, `SELECT v.pack_key,v.pack_version FROM reference_pack_versions v JOIN reference_pack_candidates c USING(pack_key,pack_version) WHERE NOT c.removed AND (v.pack_key COLLATE "C",v.pack_version COLLATE "C")>($1,$2) ORDER BY v.pack_key COLLATE "C",v.pack_version COLLATE "C" LIMIT 1`, afterKey, afterVersion).Scan(&key, &version)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := checkRetainedVersion(ctx, tx, storage, key, version); err != nil {
			return err
		}
		afterKey, afterVersion = key, version
	}
}
