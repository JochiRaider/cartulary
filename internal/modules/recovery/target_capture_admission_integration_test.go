package recovery_test

import (
	"errors"
	"io/fs"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestCaptureTargetAdmissionUsesSnapshotAndRejectsBeforeArtifacts_Integration(t *testing.T) {
	ctx := t.Context()
	fixture := newCapturedBackupContractFixture(t, ctx, pgtest.Start(t), "capture-control", uuid.New())
	controlDB, err := pgtest.OpenPurposeDatabase(ctx, fixture.source.DSN, postgres.PurposeRecovery)
	if err != nil {
		t.Fatal(err)
	}
	defer controlDB.Close()
	repository := recoveryassembly.NewVNextSnapshotRepository(fixture.sourcePool)
	err = repository.WithinRepeatableReadReadOnly(ctx, func(snapshot recovery.VNextSnapshot, control recovery.TargetControl) error {
		if control.State != recovery.TargetInitialized {
			return errors.New("unexpected initial control state")
		}
		if _, err := controlDB.ExecContext(ctx, `UPDATE public.recovery_target_state SET state='unproven',revision=revision+1`); err != nil {
			return err
		}
		rows, err := snapshot.QueryRows(ctx, `SELECT state FROM public.recovery_target_state WHERE singleton`)
		if err != nil {
			return err
		}
		defer rows.Close()
		if !rows.Next() {
			return errors.New("snapshot lost its control record")
		}
		var phase string
		if err := rows.Scan(&phase); err != nil {
			return err
		}
		if phase != string(control.State) {
			return errors.New("control and inventory reads used different snapshots")
		}
		return rows.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
	var current string
	if err := controlDB.QueryRowContext(ctx, `SELECT state FROM public.recovery_target_state WHERE singleton`).Scan(&current); err != nil || current != "unproven" {
		t.Fatalf("independent control transition did not commit: %q %v", current, err)
	}
	paths := func() []string {
		t.Helper()
		var entries []string
		if err := filepath.WalkDir(fixture.backupRoot, func(path string, _ fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			entries = append(entries, path)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		return entries
	}
	before := paths()
	_, err = fixture.Capture.Capture(ctx, recovery.VNextCaptureParams{
		BackupSetID: uuid.New(), ConsistencyPointAt: fixture.AsOf, CreatedAt: fixture.AsOf,
		RetainedUntil: fixture.AsOf.Add(31 * 24 * time.Hour),
	})
	if !errors.Is(err, recovery.ErrTargetIncomplete) {
		t.Fatalf("unproven capture was not rejected: %v", err)
	}
	if !slices.Equal(before, paths()) {
		t.Fatal("rejected capture created backup artifacts")
	}
}
