package recoveryassembly

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type VNextSnapshotRepository struct {
	db postgres.DB
}

func NewVNextSnapshotRepository(db postgres.DB) *VNextSnapshotRepository {
	return &VNextSnapshotRepository{db: db}
}

func (repository *VNextSnapshotRepository) WithinRepeatableReadReadOnly(
	ctx context.Context,
	run func(recovery.VNextSnapshot) error,
) error {
	if repository == nil || repository.db == nil || run == nil {
		return fmt.Errorf("%w: snapshot repository dependencies are required", recovery.ErrVNextBackup)
	}
	tx, err := repository.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return fmt.Errorf("begin vNext repeatable-read snapshot: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := run(vNextSnapshot{tx: tx}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit vNext repeatable-read snapshot: %w", err)
	}
	return nil
}

type vNextSnapshot struct {
	tx pgx.Tx
}

func (snapshot vNextSnapshot) StreamCanonicalTableRows(
	ctx context.Context,
	tableName string,
	visit func(json.RawMessage) error,
) error {
	if visit == nil {
		return fmt.Errorf("%w: row visitor is required", recovery.ErrVNextBackup)
	}
	identifier := pgx.Identifier{tableName}.Sanitize()
	query := fmt.Sprintf(
		"SELECT to_jsonb(snapshot_row) FROM %s snapshot_row ORDER BY to_jsonb(snapshot_row)::text ASC",
		identifier,
	)
	rows, err := snapshot.tx.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("query vNext snapshot table %s: %w", tableName, err)
	}
	defer rows.Close()
	for rows.Next() {
		var row json.RawMessage
		if err := rows.Scan(&row); err != nil {
			return fmt.Errorf("scan vNext snapshot table %s: %w", tableName, err)
		}
		if err := visit(row); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate vNext snapshot table %s: %w", tableName, err)
	}
	return nil
}

func (snapshot vNextSnapshot) QueryRows(
	ctx context.Context,
	query string,
	args ...any,
) (recovery.VNextRows, error) {
	return snapshot.tx.Query(ctx, query, args...)
}

type VNextObjectSource struct {
	store objectstore.Store
}

func NewVNextObjectSource(store objectstore.Store) *VNextObjectSource {
	return &VNextObjectSource{store: store}
}

func (source *VNextObjectSource) OpenRecoveryObject(
	ctx context.Context,
	storageKey string,
) (io.ReadCloser, error) {
	if source == nil || source.store == nil {
		return nil, fmt.Errorf("%w: recovery object source is unavailable", recovery.ErrVNextBackup)
	}
	reader, _, err := source.store.ReadObject(ctx, storageKey, objectstore.ReadOptions{})
	if err != nil {
		return nil, err
	}
	return reader, nil
}

func (source *VNextObjectSource) StatRecoveryObject(
	ctx context.Context,
	storageKey string,
) (recovery.VNextObjectSourceInfo, error) {
	if source == nil || source.store == nil {
		return recovery.VNextObjectSourceInfo{}, fmt.Errorf(
			"%w: recovery object source is unavailable",
			recovery.ErrVNextBackup,
		)
	}
	info, err := source.store.StatObject(ctx, storageKey)
	if err != nil {
		return recovery.VNextObjectSourceInfo{}, err
	}
	return recovery.VNextObjectSourceInfo{
		PlaintextBytes: info.Size,
		ContentType:    info.ContentType,
	}, nil
}
