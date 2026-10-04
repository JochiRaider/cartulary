package reference_data_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/jackc/pgx/v5"
)

var errCollectionCommitAck = errors.New("fixture collection acknowledgement lost")

type collectionCommitFault struct {
	postgres.DB
	begins int
}

func (d *collectionCommitFault) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	tx, err := d.DB.BeginTx(ctx, opts)
	d.begins++
	if err != nil || d.begins != 2 {
		return tx, err
	}
	return collectionCommitFaultTx{tx}, nil
}

type collectionCommitFaultTx struct{ pgx.Tx }

func (t collectionCommitFaultTx) Commit(ctx context.Context) error {
	if err := t.Tx.Commit(ctx); err != nil {
		return err
	}
	return errCollectionCommitAck
}

type backupSnapshotOrder struct {
	postgres.DB
	levels []pgx.TxIsoLevel
}

func (d *backupSnapshotOrder) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	d.levels = append(d.levels, opts.IsoLevel)
	return d.DB.BeginTx(ctx, opts)
}

func testPublicationCollectionAndBackupGuards(t *testing.T, db postgres.DB, storage *referenceassembly.RootStorage) {
	t.Helper()
	ctx := context.Background()
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte("orphan")))
	ref, lease, err := storage.PublishStream(ctx, digest, 6, strings.NewReader("orphan"))
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	blocked, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	err = reference_data.CollectUnreferencedObjects(blocked, db, storage)
	cancel()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("unregistered live publication collected", err)
	}
	if err := lease.Close(); err != nil {
		t.Fatal(err)
	}
	// A proven database retirement whose acknowledgement was lost still cannot
	// authorize unlinking in that attempt. The next sweep reads its durable state.
	fault := &collectionCommitFault{DB: db}
	if err := reference_data.CollectUnreferencedObjects(ctx, fault, storage); !errors.Is(err, errCollectionCommitAck) {
		t.Fatal("lost collection commit", err)
	}
	reader, _, err := storage.OpenPublished(ctx, ref)
	if err != nil {
		t.Fatal("uncertain retirement deleted bytes", err)
	}
	_ = reader.Close()
	snapshotDB := &backupSnapshotOrder{DB: db}
	if err := recoveryassembly.NewVNextSnapshotRepository(snapshotDB).WithinRepeatableReadReadOnly(ctx, func(snapshot recovery.VNextSnapshot) error {
		if len(snapshotDB.levels) != 2 || snapshotDB.levels[0] != pgx.ReadCommitted || snapshotDB.levels[1] != pgx.RepeatableRead {
			t.Fatal("backup snapshot preceded retention guard", snapshotDB.levels)
		}
		rows, err := snapshot.QueryRows(ctx, `SELECT count(*) FROM reference_pack_objects`)
		if err != nil {
			return err
		}
		rows.Close()
		blocked, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
		defer cancel()
		if err := reference_data.CollectUnreferencedObjects(blocked, db, storage); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal("capture released guard before object streaming", err)
		}
		reader, _, err := storage.OpenPublished(ctx, ref)
		if err != nil {
			return err
		}
		return reader.Close()
	}); err != nil {
		t.Fatal(err)
	}
	if err := reference_data.CollectUnreferencedObjects(ctx, db, storage); err != nil {
		t.Fatal(err)
	}
	if _, _, err := storage.OpenPublished(ctx, ref); !errors.Is(err, reference_data.ErrArtifactUnavailable) {
		t.Fatal("orphan not collected", err)
	}
	if err := reference_data.ValidateRequiredState(ctx, db, storage, reference_data.DefaultLimits()); err != nil {
		t.Fatal("collection damaged retained Base", err)
	}
}
