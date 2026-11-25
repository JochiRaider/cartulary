package httptestx

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/harnessruntime/lockwaits"
	"github.com/JochiRaider/cartulary/internal/platform/harnessruntime/runevidence"
)

// A lock-wait snapshot (TH-HARNESS-REQ-864) is defined by package lockwaits;
// this file records it under the run root for HTTP polling tests.
const lockWaitDirectory = "lock-waits"

var lockWaitSequence atomic.Int64

// SnapshotLockWaits reads the blocked sessions of db's current database and
// the sessions blocking them, blocked sessions first, at most 64.
func SnapshotLockWaits(ctx context.Context, db *sql.DB) (lockwaits.Snapshot, error) {
	if db == nil {
		return lockwaits.Snapshot{}, errors.New("lock-wait snapshot requires a database")
	}
	rows, err := db.QueryContext(ctx, lockwaits.Query, lockwaits.SessionLimit)
	if err != nil {
		return lockwaits.Snapshot{}, err
	}
	defer rows.Close()
	return lockwaits.Collect(rows, time.Now())
}

// RecordLockWaits snapshots db's lock waits and writes the snapshot under the
// run root, resolved by the rule shared with contract evidence and fixture
// artifacts, as lock-waits/<pid>-<n>.json. It returns the run-root-relative
// path. It never creates the run root and never follows a link out of it.
func RecordLockWaits(ctx context.Context, db *sql.DB) (string, lockwaits.Snapshot, error) {
	return recordLockWaitsTo(ctx, db, os.LookupEnv)
}

func recordLockWaitsTo(ctx context.Context, db *sql.DB, lookup func(string) (string, bool)) (string, lockwaits.Snapshot, error) {
	snapshot, err := SnapshotLockWaits(ctx, db)
	if err != nil {
		return "", lockwaits.Snapshot{}, err
	}
	runRoot, err := runevidence.RunRoot(lookup)
	if err != nil {
		return "", snapshot, err
	}
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return "", snapshot, err
	}
	name := fmt.Sprintf("%s/%d-%d-%d.json", lockWaitDirectory, os.Getpid(), time.Now().UnixNano(), lockWaitSequence.Add(1))
	if err := writeRunRootFile(runRoot, name, append(data, '\n')); err != nil {
		return "", snapshot, err
	}
	return name, snapshot, nil
}

func writeRunRootFile(runRoot, name string, data []byte) error {
	info, err := os.Lstat(runRoot)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || !ownedByCaller(info) {
		return errors.New("unsafe lock-wait run root")
	}
	root, err := os.OpenRoot(runRoot)
	if err != nil {
		return err
	}
	defer root.Close()
	if err := root.Mkdir(lockWaitDirectory, 0o700); err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}
	directory, err := root.Lstat(lockWaitDirectory)
	if err != nil {
		return err
	}
	if !directory.IsDir() || directory.Mode()&os.ModeSymlink != 0 || directory.Mode().Perm() != 0o700 || !ownedByCaller(directory) {
		return errors.New("unsafe lock-wait directory")
	}
	file, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL|syscall.O_NOFOLLOW, 0o600)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	return errors.Join(err, file.Close())
}

func ownedByCaller(info fs.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && int(stat.Uid) == os.Getuid()
}

// describeLockWaits records a snapshot after a poll's deadline and returns a
// one-line summary for the failure message.
func describeLockWaits(db *sql.DB) string {
	ctx, cancel := context.WithTimeout(context.Background(), lockwaits.QueryTimeout)
	defer cancel()
	path, snapshot, err := RecordLockWaits(ctx, db)
	switch {
	case err != nil && snapshot.Sessions == nil:
		return fmt.Sprintf("lock-wait snapshot unavailable: %v", err)
	case err != nil:
		return fmt.Sprintf("lock waits: %d sessions, %d blocked (not recorded: %v)", len(snapshot.Sessions), snapshot.Blocked(), err)
	default:
		return fmt.Sprintf("lock waits: %d sessions, %d blocked, recorded in %s", len(snapshot.Sessions), snapshot.Blocked(), path)
	}
}
