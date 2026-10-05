package reference_data

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ReconcileTrustBootstrap installs previously unseen repositories only. A
// changed deployment file cannot replace retained trust or lower counters.
func ReconcileTrustBootstrap(ctx context.Context, pool postgres.DB, bootstrap TrustBootstrap, at time.Time) error {
	if pool == nil || len(bootstrap.repositories) == 0 || at.IsZero() {
		return errors.New("reference pack: admitted trust bootstrap required")
	}
	ids := make([]string, 0, len(bootstrap.repositories))
	for id := range bootstrap.repositories {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, id := range ids {
		root := bootstrap.repositories[id].Root
		version, err := packformat.RootVersion(root)
		if err != nil {
			return err
		}
		insert, err := tx.Exec(ctx, `INSERT INTO reference_pack_repositories(repository_id,revision,root_version) VALUES($1,1,$2) ON CONFLICT DO NOTHING`, id, version)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `SELECT 1 FROM reference_pack_repositories WHERE repository_id=$1 FOR UPDATE`, id); err != nil {
			return err
		}
		if insert.RowsAffected() == 1 {
			if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_roots(repository_id,root_version,canonical_bytes,sha256,transition_evidence) VALUES($1,$2,$3,$4,'{}'::jsonb)`, id, version, root, packformat.Digest(root)); err != nil {
				return err
			}
			// The root row, operation and audit are one mutation. A restart that
			// sees exact retained bytes does not fabricate another root-import event.
			operation := uuid.New()
			frozen, err := canonicaljson.Marshal(map[string]any{"schema_id": "cartulary.reference_pack_root_bootstrap.v1", "repository_id": id, "root_version": version, "root_sha256": packformat.Digest(root)})
			if err != nil {
				return err
			}
			outcome, err := canonicaljson.Marshal(map[string]any{"schema_id": "cartulary.reference_pack_root_bootstrap_result.v1", "result": "succeeded"})
			if err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_operations(operation_id,kind,actor_kind,admitted_at,terminal_at,frozen_input,final_outcome) VALUES($1,'reconcile','system',$2,$2,$3,$4)`, operation, at.UTC(), frozen, outcome); err != nil {
				return err
			}
			if err := appendPackAuditTx(ctx, tx, operation, "root_import", "succeeded", at.UTC(), "", "", nil); err != nil {
				return err
			}
		} else {
			var retained []byte
			if err := tx.QueryRow(ctx, `SELECT canonical_bytes FROM reference_pack_roots WHERE repository_id=$1 AND root_version=$2`, id, version).Scan(&retained); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return errors.New("reference pack: bootstrap root is not retained trust history")
				}
				return err
			}
			if !bytes.Equal(retained, root) {
				return errors.New("reference pack: bootstrap root conflicts with retained trust history")
			}
		}
	}
	return tx.Commit(ctx)
}
