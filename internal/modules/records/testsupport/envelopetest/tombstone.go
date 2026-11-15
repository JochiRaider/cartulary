package envelopetest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
)

// Tombstone describes a SQL soft delete for fixtures that bypass the
// production delete command.
type Tombstone struct {
	RecordID  uuid.UUID
	DeletedBy uuid.UUID
	// At is the deletion instant. The zero value uses the database clock.
	At time.Time
	// AtLastUpdate deletes at the record's current updated_at, leaving the
	// envelope clocks and version unchanged. It suits synthetic barrier
	// tombstones that a fixture later undoes. It excludes At.
	AtLastUpdate bool
	// RowVersion, when positive, raises row_version to at least this value,
	// as a production delete advances it.
	RowVersion int64
}

// The deletion instant is read once and locked with the row; every later
// write keeps created_at <= deleted_at <= updated_at, even between statements.
var readTombstoneInstantStatement = sqlshape.Command(`
SELECT CASE WHEN $2::boolean THEN updated_at
            ELSE COALESCE($3::timestamptz, clock_timestamp())
       END,
       created_at
  FROM records
 WHERE record_id = $1
   FOR UPDATE
`)

// Envelope clock and version columns are written only when they change, so a
// tombstone at the last update leaves every mirror-checked column untouched.
var advanceTombstoneUpdatedAtStatement = sqlshape.Command(`
UPDATE records
   SET updated_at = $2
 WHERE record_id = $1
   AND updated_at < $2
`)

var setTombstoneStatement = sqlshape.Command(`
UPDATE records
   SET deleted_at = $2,
       deleted_by_user_id = $3
 WHERE record_id = $1
`)

var raiseTombstoneVersionStatement = sqlshape.Command(`
UPDATE records
   SET row_version = $2::bigint
 WHERE record_id = $1
   AND row_version < $2::bigint
`)

// The Host and Identity source rows mirror their Records envelope; the 00011
// triggers reject any committed difference.
var envelopeMirrorStatements = []sqlshape.Statement{
	sqlshape.Command(`
UPDATE hosts AS source
   SET row_version = envelope.row_version,
       updated_at = envelope.updated_at
  FROM records AS envelope
 WHERE source.record_id = envelope.record_id
   AND envelope.record_id = $1
   AND (source.row_version, source.updated_at) IS DISTINCT FROM (envelope.row_version, envelope.updated_at)
`),
	sqlshape.Command(`
UPDATE identities AS source
   SET row_version = envelope.row_version,
       updated_at = envelope.updated_at
  FROM records AS envelope
 WHERE source.record_id = envelope.record_id
   AND envelope.record_id = $1
   AND (source.row_version, source.updated_at) IS DISTINCT FROM (envelope.row_version, envelope.updated_at)
`),
}

// SoftDeleteRecord tombstones one Records row without leaving an illegal
// envelope (Core 01 REQ-01-651: created_at <= deleted_at <= updated_at), at
// commit or between its statements. updated_at advances to the deletion
// instant only when it is earlier, and row_version only when the Tombstone
// asks. The Host/Identity source mirror follows the envelope in the same
// transaction, touching only mirror rows that differ; a caller-supplied pgx.Tx
// or *sql.Tx is used as that transaction. A missing record, or a deletion
// instant before the record's creation, fails the test.
func SoftDeleteRecord(t testing.TB, db any, tombstone Tombstone) {
	t.Helper()
	if tombstone.AtLastUpdate && !tombstone.At.IsZero() {
		t.Fatalf("soft-delete record fixture: At and AtLastUpdate are exclusive")
	}
	var requested any
	if !tombstone.At.IsZero() {
		requested = tombstone.At
	}
	err := inTransaction(db, func(run statementRunner) error {
		var at, createdAt time.Time
		if err := run.queryRow(readTombstoneInstantStatement, []any{tombstone.RecordID, tombstone.AtLastUpdate, requested}, &at, &createdAt); err != nil {
			if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("record %s is missing", tombstone.RecordID)
			}
			return err
		}
		if at.Before(createdAt) {
			return fmt.Errorf("record %s was created after the deletion instant", tombstone.RecordID)
		}
		if _, err := run.exec(advanceTombstoneUpdatedAtStatement, tombstone.RecordID, at); err != nil {
			return err
		}
		if _, err := run.exec(setTombstoneStatement, tombstone.RecordID, at, tombstone.DeletedBy); err != nil {
			return err
		}
		if _, err := run.exec(raiseTombstoneVersionStatement, tombstone.RecordID, tombstone.RowVersion); err != nil {
			return err
		}
		for _, mirror := range envelopeMirrorStatements {
			if _, err := run.exec(mirror, tombstone.RecordID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("soft-delete record fixture: %v", err)
	}
}

type statementRunner struct {
	exec     func(sqlshape.Statement, ...any) (int64, error)
	queryRow func(sqlshape.Statement, []any, ...any) error
}

func inTransaction(db any, body func(statementRunner) error) error {
	ctx := context.Background()
	switch typed := db.(type) {
	case pgx.Tx:
		return body(pgxStatementRunner(ctx, typed))
	case dbport.DB:
		tx, err := typed.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback(ctx) }()
		if err := body(pgxStatementRunner(ctx, tx)); err != nil {
			return err
		}
		return tx.Commit(ctx)
	case *sql.Tx:
		return body(sqlStatementRunner(ctx, typed))
	case *sql.DB:
		tx, err := typed.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()
		if err := body(sqlStatementRunner(ctx, tx)); err != nil {
			return err
		}
		return tx.Commit()
	default:
		return fmt.Errorf("unsupported Records test database %T", db)
	}
}

func pgxStatementRunner(ctx context.Context, tx pgx.Tx) statementRunner {
	return statementRunner{
		exec: func(statement sqlshape.Statement, args ...any) (int64, error) {
			tag, err := tx.Exec(ctx, sqlshape.CommandText(statement), args...)
			return tag.RowsAffected(), err
		},
		queryRow: func(statement sqlshape.Statement, args []any, dest ...any) error {
			return tx.QueryRow(ctx, sqlshape.CommandText(statement), args...).Scan(dest...)
		},
	}
}

func sqlStatementRunner(ctx context.Context, tx *sql.Tx) statementRunner {
	return statementRunner{
		exec: func(statement sqlshape.Statement, args ...any) (int64, error) {
			result, err := tx.ExecContext(ctx, sqlshape.CommandText(statement), args...)
			if err != nil {
				return 0, err
			}
			return result.RowsAffected()
		},
		queryRow: func(statement sqlshape.Statement, args []any, dest ...any) error {
			return tx.QueryRowContext(ctx, sqlshape.CommandText(statement), args...).Scan(dest...)
		},
	}
}
