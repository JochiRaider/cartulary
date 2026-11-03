package recovery_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	graphrestore "github.com/JochiRaider/cartulary/internal/modules/graphprojection/restore"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
)

func TestStructuralRestoreNeedsOnlyCoreRebuilds_Integration(t *testing.T) {
	ctx := t.Context()
	fixture := newRestoreProjectionContractFixture(t, ctx, "structural-core-only", uuid.New())
	core, err := projectionassembly.Build(fixture.Target.Postgres)
	if err != nil {
		t.Fatal(err)
	}
	fixture.Target.Projections = core.RecoveryPorts().Rebuilder
	admitted := admitCurrentRestore(t, fixture.capturedBackupContractFixture, fixture.Runner, fixture.Target)
	result, err := fixture.Runner.RestoreBackupSet(ctx, admitted.Fence.Target(), fixture.BackupSet)
	if err != nil {
		t.Fatal("structural restore required an extension participant", err)
	}
	if !result.ProjectionRebuildResult.ReadinessSatisfied() {
		t.Fatal("independent Core rebuild is incomplete")
	}
	for _, name := range graphrestore.RestoreGraphTableIDs() {
		var count int
		if err := fixture.Target.Postgres.QueryRow(ctx, "SELECT count(*) FROM "+pgx.Identifier{"public", name}.Sanitize()).Scan(&count); err != nil || count != 0 {
			t.Fatal("structural restore populated derived Graph state", name, count, err)
		}
	}
	admitted.Complete(t, result)
	control, err := recovery.ReadTargetControl(ctx, fixture.Target.Postgres)
	if err != nil || control.RequireServing(admitted.Identity.BindingDigests) != nil {
		t.Fatal("structural completion did not admit independent Core serving", err)
	}
}
