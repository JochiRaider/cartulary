package reference_data

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// lockPublicationTx protects exactly the dependencies captured by admission.
// It never retries or rebases a proposal after waiting for another mutation.
// Lock order is repositories, pack keys, current set, then registry usage.
func lockPublicationTx(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, frozen frozenOperation) error {
	stale := func() error { return &OperationRejection{Reason: "stale_admission_state"} }
	retainedKeys, err := lockPublicationDependenciesTx(ctx, tx, operationID)
	if err != nil {
		return err
	}
	var revision int64
	var setID *string
	var configuration string
	if err := tx.QueryRow(ctx, `SELECT pack_set_id,revision,configuration_sha256 FROM reference_pack_current_set WHERE singleton FOR UPDATE`).Scan(&setID, &revision, &configuration); err != nil {
		return err
	}
	if (setID == nil) != (frozen.PackSetID == nil) || setID != nil && *setID != *frozen.PackSetID || revision != frozen.SetRevision || configuration != frozen.ConfigurationSHA256 {
		return stale()
	}
	for _, expected := range retainedKeys {
		if expected.usage == nil {
			continue
		}
		var revision int64
		err := tx.QueryRow(ctx, `SELECT revision FROM reference_pack_registry_usage WHERE pack_key=$1 FOR UPDATE`, expected.key).Scan(&revision)
		if errors.Is(err, pgx.ErrNoRows) {
			return stale()
		}
		if err != nil {
			return err
		}
		if revision != *expected.usage {
			return stale()
		}
	}
	return nil
}

type publicationKeyRevision struct {
	key      string
	revision int64
	usage    *int64
}

// All verification publications share these exact repository and key guards.
// The caller then locks and compares only its relevant configuration/set and
// registry-usage facts, preserving repository/key/set/usage lock order.
func lockPublicationDependenciesTx(ctx context.Context, tx pgx.Tx, operationID uuid.UUID) ([]publicationKeyRevision, error) {
	stale := func() error { return &OperationRejection{Reason: "stale_admission_state"} }
	repositories, err := tx.Query(ctx, `SELECT repository_id,admitted_revision,root_version FROM reference_pack_operation_repositories WHERE operation_id=$1 ORDER BY repository_id COLLATE "C"`, operationID)
	if err != nil {
		return nil, err
	}
	type repositoryRevision struct {
		id             string
		revision, root int64
	}
	retainedRepositories := []repositoryRevision{}
	for repositories.Next() {
		var r repositoryRevision
		if err := repositories.Scan(&r.id, &r.revision, &r.root); err != nil {
			repositories.Close()
			return nil, err
		}
		retainedRepositories = append(retainedRepositories, r)
	}
	err = repositories.Err()
	repositories.Close()
	if err != nil {
		return nil, err
	}
	for _, expected := range retainedRepositories {
		var revision, root int64
		if err := tx.QueryRow(ctx, `SELECT revision,root_version FROM reference_pack_repositories WHERE repository_id=$1 FOR UPDATE`, expected.id).Scan(&revision, &root); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, stale()
			}
			return nil, err
		}
		if revision != expected.revision || root != expected.root {
			return nil, stale()
		}
	}
	keys, err := tx.Query(ctx, `SELECT pack_key,admitted_revision,usage_revision FROM (
 SELECT pack_key,admitted_revision,usage_revision FROM reference_pack_operation_keys WHERE operation_id=$1
 UNION ALL SELECT pack_key,admitted_revision,NULL::bigint FROM reference_pack_operation_dependency_keys WHERE operation_id=$1
) captured ORDER BY pack_key COLLATE "C",admitted_revision`, operationID)
	if err != nil {
		return nil, err
	}
	retainedKeys := []publicationKeyRevision{}
	for keys.Next() {
		var k publicationKeyRevision
		if err := keys.Scan(&k.key, &k.revision, &k.usage); err != nil {
			keys.Close()
			return nil, err
		}
		retainedKeys = append(retainedKeys, k)
	}
	err = keys.Err()
	keys.Close()
	if err != nil {
		return nil, err
	}
	for _, expected := range retainedKeys {
		var revision int64
		if err := tx.QueryRow(ctx, `SELECT revision FROM reference_pack_key_state WHERE pack_key=$1 FOR UPDATE`, expected.key).Scan(&revision); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, stale()
			}
			return nil, err
		}
		if revision != expected.revision {
			return nil, stale()
		}
	}
	return retainedKeys, nil
}
