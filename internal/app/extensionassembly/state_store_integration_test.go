package extensionassembly

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestStateStore_Integration_LogicalPortAndScope(t *testing.T) {
	harness := pgtest.Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "extension-state-store-adapter")
	pool, err := pgxpool.New(context.Background(), testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	physical, err := extensionstore.New(pool, networkflow.ExtensionStateFamilyCounters())
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewStateStore(physical)
	if err != nil {
		t.Fatal(err)
	}
	familyIDs := extensionstore.SortedFamilyIDs(networkflow.ExtensionStateFamilyCounters())

	if err := store.WithProfileSession(context.Background(), "network_flow_activity", time.Second, func(session extensions.ProfileSession) error {
		snapshot, err := session.Snapshot(
			context.Background(),
			"network_flow_activity",
			"network_flow_activity.state_v1",
			familyIDs,
		)
		if err != nil {
			return err
		}
		if snapshot.Metadata == nil ||
			snapshot.Metadata.StateVersion != 1 ||
			snapshot.Metadata.MigrationLineageID != "network_flow_activity.state_v1" ||
			len(snapshot.Ledger) != 0 {
			t.Fatalf("logical snapshot = %#v", snapshot)
		}
		for _, familyID := range familyIDs {
			if snapshot.FamilyCounts[familyID] != 0 {
				t.Fatalf("family %q count = %d, want 0", familyID, snapshot.FamilyCounts[familyID])
			}
		}

		transaction, err := session.Begin(context.Background(), "network_flow_activity", familyIDs)
		if err != nil {
			return err
		}
		if transaction.CapabilityID() == "" ||
			!transaction.IsStateWriteCapability() ||
			!equalStateFamilies(transaction.StateFamilyIDs(), familyIDs) {
			t.Fatalf("logical transaction capability = %q/%t/%v", transaction.CapabilityID(), transaction.IsStateWriteCapability(), transaction.StateFamilyIDs())
		}
		if _, err := transaction.FamilyCounts(context.Background(), familyIDs[:len(familyIDs)-1]); err == nil {
			t.Fatal("transaction accepted a family request outside its exact capability scope")
		}
		outcome, err := transaction.Rollback(context.Background())
		if err != nil || outcome != extensions.CommitAbsent {
			t.Fatalf("rollback = %q/%v", outcome, err)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := NewStateStore(nil); err == nil {
		t.Fatal("nil physical store was accepted")
	}
}

func TestStateStore_Integration_MapsProfileLockTimeout(t *testing.T) {
	harness := pgtest.Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "extension-state-store-lock-timeout")
	pool, err := pgxpool.New(context.Background(), testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	physical, err := extensionstore.New(pool, networkflow.ExtensionStateFamilyCounters())
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewStateStore(physical)
	if err != nil {
		t.Fatal(err)
	}

	acquired := make(chan struct{})
	release := make(chan struct{})
	holderDone := make(chan error, 1)
	go func() {
		holderDone <- physical.WithProfileLock(context.Background(), "network_flow_activity", time.Second, func(*extensionstore.Session) error {
			close(acquired)
			<-release
			return nil
		})
	}()
	<-acquired

	err = store.WithProfileSession(context.Background(), "network_flow_activity", time.Millisecond, func(extensions.ProfileSession) error {
		t.Fatal("operation ran without acquiring the profile lock")
		return nil
	})
	if !errors.Is(err, extensions.ErrStateMigrationLockTimeout) {
		t.Fatalf("mapped lock error = %v", err)
	}
	close(release)
	if err := <-holderDone; err != nil {
		t.Fatal(err)
	}
}
