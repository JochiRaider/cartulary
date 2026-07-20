// Package extensionstore owns the PostgreSQL adapter for generic extension
// coordination records. It does not own profile state or profile behavior.
package extensionstore

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrProfileLockTimeout = errors.New("extension migration lock timeout")
	ErrInvalidTransition  = errors.New("invalid extension coordination transition")
	ErrNotFound           = errors.New("extension coordination record not found")
	ErrIntegrity          = errors.New("extension coordination integrity failure")
)

type Querier interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

// FamilyCounter is supplied by a named state owner. The generated contract
// identifies the logical family; this adapter only performs the owner's physical
// count implementation inside the shared lock/transaction boundary.
type FamilyCounter struct {
	FamilyID string
	Count    func(context.Context, Querier) (int64, error)
}

type FamilyCountReader interface {
	FamilyCounts(context.Context, []string) (map[string]int64, error)
}

type Store struct {
	pool     *pgxpool.Pool
	counters map[string]FamilyCounter
}

func New(pool *pgxpool.Pool, counters []FamilyCounter) (*Store, error) {
	if pool == nil {
		return nil, errors.New("extension store requires PostgreSQL")
	}
	indexed := make(map[string]FamilyCounter, len(counters))
	for _, counter := range counters {
		if counter.FamilyID == "" || counter.Count == nil {
			return nil, errors.New("extension family counter is incomplete")
		}
		if _, duplicate := indexed[counter.FamilyID]; duplicate {
			return nil, fmt.Errorf("duplicate extension family counter %s", counter.FamilyID)
		}
		indexed[counter.FamilyID] = counter
	}
	return &Store{pool: pool, counters: indexed}, nil
}

type Session struct {
	conn     *pgxpool.Conn
	counters map[string]FamilyCounter
}

func (s *Store) WithProfileLock(ctx context.Context, profileID string, timeout time.Duration, operation func(*Session) error) error {
	if s == nil || s.pool == nil || operation == nil || profileID == "" {
		return errors.New("invalid extension profile lock request")
	}
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire extension profile lock session: %w", err)
	}
	defer conn.Release()
	lockCtx := ctx
	cancel := func() {}
	if timeout > 0 {
		lockCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()
	if _, err := conn.Exec(lockCtx, `SELECT pg_advisory_lock(hashtextextended($1, 0))`, "cartulary.extension.migration:"+profileID); err != nil {
		if errors.Is(lockCtx.Err(), context.DeadlineExceeded) {
			return ErrProfileLockTimeout
		}
		return fmt.Errorf("acquire extension profile lock: %w", err)
	}
	defer func() {
		unlockCtx, unlockCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer unlockCancel()
		_, _ = conn.Exec(unlockCtx, `SELECT pg_advisory_unlock(hashtextextended($1, 0))`, "cartulary.extension.migration:"+profileID)
	}()
	return operation(&Session{conn: conn, counters: s.counters})
}

func (s *Session) Begin(ctx context.Context) (*Tx, error) {
	if s == nil || s.conn == nil {
		return nil, errors.New("extension profile session unavailable")
	}
	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx, counters: s.counters}, nil
}

type CommitOutcome string

const (
	CommitProven  CommitOutcome = "committed"
	CommitAbsent  CommitOutcome = "absent"
	CommitUnknown CommitOutcome = "indeterminate"
)

type Tx struct {
	tx       pgx.Tx
	counters map[string]FamilyCounter
	closed   bool
}

func (t *Tx) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	if t == nil || t.tx == nil || t.closed {
		return nil, errors.New("extension state transaction unavailable")
	}
	return t.tx.Query(ctx, query, args...)
}

func (t *Tx) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	return t.tx.QueryRow(ctx, query, args...)
}

func (t *Tx) Exec(ctx context.Context, query string, args ...any) error {
	if t == nil || t.tx == nil || t.closed {
		return errors.New("extension state transaction unavailable")
	}
	_, err := t.tx.Exec(ctx, query, args...)
	return err
}

func (t *Tx) FamilyCounts(ctx context.Context, familyIDs []string) (map[string]int64, error) {
	if t == nil || t.tx == nil || t.closed {
		return nil, errors.New("extension state transaction unavailable")
	}
	result := make(map[string]int64, len(familyIDs))
	previous := ""
	for _, familyID := range familyIDs {
		if familyID == "" || (previous != "" && previous >= familyID) {
			return nil, errors.New("extension family IDs must be sorted and unique")
		}
		previous = familyID
		counter, ok := t.counters[familyID]
		if !ok {
			return nil, fmt.Errorf("extension family counter unavailable: %s", familyID)
		}
		count, err := counter.Count(ctx, t.tx)
		if err != nil {
			return nil, fmt.Errorf("count extension family %s: %w", familyID, err)
		}
		if count < 0 {
			return nil, fmt.Errorf("negative extension family count %s", familyID)
		}
		result[familyID] = count
	}
	return result, nil
}

type StateMetadata struct {
	ProfileID          string
	MigrationLineageID string
	StateVersion       int
	LastMigrationID    *string
	MetadataVersion    int
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (t *Tx) StateMetadata(ctx context.Context, profileID string) (*StateMetadata, error) {
	var metadata StateMetadata
	err := t.tx.QueryRow(ctx, `
SELECT profile_id, migration_lineage_id, state_version, last_migration_id,
       metadata_version, created_at, updated_at
  FROM extension_state_metadata
 WHERE profile_id = $1
 FOR UPDATE
`, profileID).Scan(
		&metadata.ProfileID,
		&metadata.MigrationLineageID,
		&metadata.StateVersion,
		&metadata.LastMigrationID,
		&metadata.MetadataVersion,
		&metadata.CreatedAt,
		&metadata.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

func (t *Tx) InsertStateMetadata(ctx context.Context, metadata StateMetadata) error {
	if metadata.ProfileID == "" || metadata.MigrationLineageID == "" || metadata.StateVersion < 1 || metadata.MetadataVersion != 1 || metadata.LastMigrationID != nil {
		return ErrInvalidTransition
	}
	_, err := t.tx.Exec(ctx, `
INSERT INTO extension_state_metadata (
    profile_id, migration_lineage_id, state_version, last_migration_id,
    metadata_version, created_at, updated_at
) VALUES ($1, $2, $3, NULL, 1, $4, $4)
`, metadata.ProfileID, metadata.MigrationLineageID, metadata.StateVersion, metadata.CreatedAt.UTC())
	return err
}

func (t *Tx) UpdateStateMetadata(ctx context.Context, before StateMetadata, stateVersion int, migrationID string, now time.Time) error {
	if before.MetadataVersion < 1 || before.MetadataVersion == 2147483647 || stateVersion != before.StateVersion+1 || migrationID == "" {
		return ErrInvalidTransition
	}
	tag, err := t.tx.Exec(ctx, `
UPDATE extension_state_metadata
   SET state_version = $2,
       last_migration_id = $3,
       metadata_version = metadata_version + 1,
       updated_at = $4
 WHERE profile_id = $1
   AND migration_lineage_id = $5
   AND state_version = $6
   AND metadata_version = $7
`, before.ProfileID, stateVersion, migrationID, now.UTC(), before.MigrationLineageID, before.StateVersion, before.MetadataVersion)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrIntegrity
	}
	return nil
}

type MigrationLedgerEntry struct {
	ProfileID                 string
	MigrationLineageID        string
	MigrationID               string
	FromStateVersion          int
	ToStateVersion            int
	MigrationDefinitionSHA256 string
	CommittedAt               time.Time
	ResultingStateVersion     int
}

func (t *Tx) MigrationLedger(ctx context.Context, profileID, lineageID string) ([]MigrationLedgerEntry, error) {
	rows, err := t.tx.Query(ctx, `
SELECT profile_id, migration_lineage_id, migration_id, from_state_version,
       to_state_version, migration_definition_sha256, committed_at,
       resulting_state_version
  FROM extension_migration_ledger
 WHERE profile_id = $1 AND migration_lineage_id = $2
 ORDER BY from_state_version, to_state_version, migration_id
`, profileID, lineageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []MigrationLedgerEntry
	for rows.Next() {
		var entry MigrationLedgerEntry
		if err := rows.Scan(
			&entry.ProfileID,
			&entry.MigrationLineageID,
			&entry.MigrationID,
			&entry.FromStateVersion,
			&entry.ToStateVersion,
			&entry.MigrationDefinitionSHA256,
			&entry.CommittedAt,
			&entry.ResultingStateVersion,
		); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (t *Tx) InsertMigrationLedger(ctx context.Context, entry MigrationLedgerEntry) error {
	if entry.ToStateVersion != entry.FromStateVersion+1 || entry.ResultingStateVersion != entry.ToStateVersion {
		return ErrInvalidTransition
	}
	_, err := t.tx.Exec(ctx, `
INSERT INTO extension_migration_ledger (
    profile_id, migration_lineage_id, migration_id, from_state_version,
    to_state_version, migration_definition_sha256, committed_at,
    resulting_state_version
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, entry.ProfileID, entry.MigrationLineageID, entry.MigrationID, entry.FromStateVersion,
		entry.ToStateVersion, entry.MigrationDefinitionSHA256, entry.CommittedAt.UTC(),
		entry.ResultingStateVersion)
	return err
}

func (t *Tx) Commit(ctx context.Context) (CommitOutcome, error) {
	if t == nil || t.tx == nil || t.closed {
		return CommitAbsent, ErrInvalidTransition
	}
	t.closed = true
	if err := t.tx.Commit(ctx); err != nil {
		return CommitUnknown, err
	}
	return CommitProven, nil
}

func (t *Tx) Rollback(ctx context.Context) (CommitOutcome, error) {
	if t == nil || t.tx == nil || t.closed {
		return CommitAbsent, nil
	}
	t.closed = true
	err := t.tx.Rollback(ctx)
	if err == nil || errors.Is(err, pgx.ErrTxClosed) {
		return CommitAbsent, nil
	}
	return CommitUnknown, err
}

func SortedFamilyIDs(counters []FamilyCounter) []string {
	ids := make([]string, 0, len(counters))
	for _, counter := range counters {
		ids = append(ids, counter.FamilyID)
	}
	sort.Strings(ids)
	return ids
}
