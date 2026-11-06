package extensionassembly

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	"github.com/google/uuid"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/stagedobjects"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestStagedObjects_Integration_AllocationPublicationAndCleanup(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 7, 24, 18, 0, 0, 0, time.UTC)
	pool, physical, objectBytes, objectStore := newStagedObjectsIntegrationFixture(t)
	repository, err := NewStagedObjectRepository(physical)
	if err != nil {
		t.Fatal(err)
	}
	var sequence int
	service, err := stagedobjects.NewService(stagedobjects.ServiceOptions{
		Repository: repository,
		Bytes:      objectBytes,
		Now:        func() time.Time { return now },
		NewID: func() (string, error) {
			sequence++
			return "staging-" + string(rune('0'+sequence)), nil
		},
		FatalSink: func(err error) { t.Fatalf("unexpected allocation fatal: %v", err) },
	})
	if err != nil {
		t.Fatal(err)
	}

	scope, err := stagedobjects.NewScope("operation-1", "network_flow_activity", service)
	if err != nil {
		t.Fatal(err)
	}
	stagingID, err := scope.Allocate(ctx, "operation-1", []byte("published payload"))
	if err != nil {
		t.Fatal(err)
	}
	stored, err := physical.StagedObject(ctx, stagingID)
	if err != nil ||
		stored.State != extensionstore.StagedReady ||
		stored.StorageIdentity == "" ||
		stored.ExpectedByteSize == nil || *stored.ExpectedByteSize != int64(len("published payload")) {
		t.Fatalf("ready staged object = %#v/%v", stored, err)
	}
	if _, err := objectStore.StatObject(ctx, stored.StorageIdentity); err != nil {
		t.Fatalf("ready bytes missing: %v", err)
	}

	transfer, err := scope.Transfer("operation-1")
	if err != nil {
		t.Fatal(err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	capability, err := NewStagedPublicationCapability("operation-1", tx, func() time.Time { return now.Add(time.Minute) })
	if err != nil {
		t.Fatal(err)
	}
	if err := stagedobjects.PublishTransferred(ctx, transfer, []stagedobjects.Publication{{
		StagingID:    stagingID,
		ResourceKind: "incident_bundle",
		ResourceID:   "bundle-1",
		ByteSize:     *stored.ExpectedByteSize,
		SHA256:       *stored.ExpectedSHA256,
	}}, capability); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	published, err := physical.RedeemableStagedObject(ctx, stagingID, now.Add(2*time.Minute))
	if err != nil || published.State != extensionstore.StagedPublished {
		t.Fatalf("published staged object = %#v/%v", published, err)
	}

	cleanupScope, err := stagedobjects.NewScope("operation-2", "network_flow_activity", service)
	if err != nil {
		t.Fatal(err)
	}
	cleanupID, err := cleanupScope.Allocate(ctx, "operation-2", []byte("abandoned payload"))
	if err != nil {
		t.Fatal(err)
	}
	cleanupRecord, err := physical.StagedObject(ctx, cleanupID)
	if err != nil {
		t.Fatal(err)
	}
	if err := cleanupScope.Abandon(ctx); err != nil {
		t.Fatal(err)
	}
	janitor, err := stagedobjects.NewJanitor(stagedobjects.JanitorOptions{
		Repository:       repository,
		Bytes:            objectBytes,
		Now:              func() time.Time { return now.Add(time.Minute) },
		BatchLimit:       1,
		OperationTimeout: time.Second,
		FatalSink:        func(err error) { t.Fatalf("unexpected cleanup fatal: %v", err) },
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := janitor.Sweep(ctx); err != nil {
		t.Fatal(err)
	}
	deleted, err := physical.StagedObject(ctx, cleanupID)
	if err != nil || deleted.DeleteState != extensionstore.DeleteDeleted {
		t.Fatalf("cleaned staged object = %#v/%v", deleted, err)
	}
	if _, err := objectStore.StatObject(ctx, cleanupRecord.StorageIdentity); err == nil {
		t.Fatal("physical bytes remain after successful cleanup")
	}
}

func TestStagedObjects_Integration_PublicationRollbackLeavesReadyAndInaccessible(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 7, 24, 19, 0, 0, 0, time.UTC)
	pool, physical, objectBytes, _ := newStagedObjectsIntegrationFixture(t)
	repository, err := NewStagedObjectRepository(physical)
	if err != nil {
		t.Fatal(err)
	}
	service, err := stagedobjects.NewService(stagedobjects.ServiceOptions{
		Repository: repository,
		Bytes:      objectBytes,
		Now:        func() time.Time { return now },
		NewID:      func() (string, error) { return "staging-rollback", nil },
		FatalSink:  func(err error) { t.Fatalf("unexpected fatal: %v", err) },
	})
	if err != nil {
		t.Fatal(err)
	}
	scope, err := stagedobjects.NewScope("operation-rollback", "network_flow_activity", service)
	if err != nil {
		t.Fatal(err)
	}
	stagingID, err := scope.Allocate(ctx, "operation-rollback", []byte("payload"))
	if err != nil {
		t.Fatal(err)
	}
	record, err := physical.StagedObject(ctx, stagingID)
	if err != nil {
		t.Fatal(err)
	}
	transfer, err := scope.Transfer("operation-rollback")
	if err != nil {
		t.Fatal(err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	capability, err := NewStagedPublicationCapability("operation-rollback", tx, func() time.Time { return now.Add(time.Minute) })
	if err != nil {
		t.Fatal(err)
	}
	if err := stagedobjects.PublishTransferred(ctx, transfer, []stagedobjects.Publication{{
		StagingID:    stagingID,
		ResourceKind: "incident_bundle",
		ResourceID:   "bundle-rollback",
		ByteSize:     *record.ExpectedByteSize,
		SHA256:       *record.ExpectedSHA256,
	}}, capability); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	rolledBack, err := physical.StagedObject(ctx, stagingID)
	if err != nil || rolledBack.State != extensionstore.StagedReady {
		t.Fatalf("rolled-back publication = %#v/%v", rolledBack, err)
	}
	if _, err := physical.RedeemableStagedObject(ctx, stagingID, now.Add(2*time.Minute)); !errors.Is(err, extensionstore.ErrNotFound) {
		t.Fatalf("rolled-back staged object became redeemable: %v", err)
	}
	var references int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM extension_staged_object_references WHERE staging_id = $1`, stagingID).Scan(&references); err != nil {
		t.Fatal(err)
	}
	if references != 0 {
		t.Fatalf("rolled-back publication retained %d references", references)
	}
}

func TestStagedObjects_Integration_ReferenceContradictionIsFatalBeforeDelete(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 7, 24, 20, 0, 0, 0, time.UTC)
	pool, physical, objectBytes, _ := newStagedObjectsIntegrationFixture(t)
	repository, err := NewStagedObjectRepository(physical)
	if err != nil {
		t.Fatal(err)
	}
	record := extensionstore.NewStagedObject(
		"staging-contradiction",
		"operation-contradiction",
		"network_flow_activity",
		".cartulary/staged/network_flow_activity/staging-contradiction",
		now.Add(-25*time.Hour),
	)
	if err := physical.AllocateStagedObject(ctx, record); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO extension_staged_object_references (
    staging_id, owner_resource_kind, owner_resource_id,
    expected_byte_size, expected_sha256, committed_at
) VALUES ($1, 'incident_bundle', 'bundle-contradiction', $2, $3, $4)
`, record.StagingID, 7, "239f59ed55e737c77147cf55ad0c1b030b6d7ee748a7426952f9b852d5a935e5", now); err != nil {
		t.Fatal(err)
	}
	var fatalCalls int
	janitor, err := stagedobjects.NewJanitor(stagedobjects.JanitorOptions{
		Repository:       repository,
		Bytes:            objectBytes,
		Now:              func() time.Time { return now },
		BatchLimit:       1,
		OperationTimeout: time.Second,
		FatalSink:        func(error) { fatalCalls++ },
	})
	if err != nil {
		t.Fatal(err)
	}
	err = janitor.Sweep(ctx)
	if !stagedobjects.IsFatalIntegrity(err) || fatalCalls != 1 {
		t.Fatalf("reference contradiction = %v fatal_calls=%d", err, fatalCalls)
	}
	current, err := physical.StagedObject(ctx, record.StagingID)
	if err != nil || current.DeleteState != extensionstore.DeleteNotApplicable {
		t.Fatalf("contradictory object was mutated = %#v/%v", current, err)
	}
}

func newStagedObjectsIntegrationFixture(t testing.TB) (*pgxpool.Pool, *extensionstore.Store, *StagedObjectBytes, objectstore.Store) {
	t.Helper()
	harness := pgtest.Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "staged-objects-owner")
	pool, err := pgxpool.New(context.Background(), testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	physical, err := extensionstore.New(pool, nil)
	if err != nil {
		t.Fatal(err)
	}
	objectStore, err := objectstore.NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = objectStore.Close() })
	objectBytes, err := NewStagedObjectBytes(objectStore)
	if err != nil {
		t.Fatal(err)
	}
	return pool, physical, objectBytes, objectStore
}

func TestStagedObjects_Integration_ReservedOwnerAndCompetingAttempts(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	pool, physical, bytes, objects := newStagedObjectsIntegrationFixture(t)
	repository, err := NewStagedObjectRepository(physical)
	if err != nil {
		t.Fatal(err)
	}
	service, err := stagedobjects.NewService(stagedobjects.ServiceOptions{Repository: repository, Bytes: bytes, Now: func() time.Time { return now }, ResourceIdentity: evidence.IncidentBundleResourceIdentity, FatalSink: func(err error) { t.Errorf("fatal: %v", err) }})
	if err != nil {
		t.Fatal(err)
	}
	incident, blob := uuid.New(), uuid.New()
	resource := stagedobjects.ResourceBinding{Kind: "object_blob", ID: incident.String() + ":" + blob.String()}
	type attempt struct {
		resources *stagedobjects.PreparedResources
		key       string
		id        string
	}
	allocate := func(operation, content string) attempt {
		t.Helper()
		scope, err := stagedobjects.NewScope(operation, "evidence", service)
		if err != nil {
			t.Fatal(err)
		}
		sealed, key, err := scope.AllocateResourceStream(ctx, operation, resource, strings.NewReader(content), int64(len(content)))
		if err != nil {
			t.Fatal(err)
		}
		parts, err := blobref.ParseObjectBlobStorageKey(key)
		if err != nil || parts.IncidentID != incident || parts.ObjectBlobID != blob {
			t.Fatalf("owner key: %s %v", key, err)
		}
		stored, err := physical.StagedObject(ctx, sealed.StagingID)
		if err != nil || stored.ReservedResourceID == nil || *stored.ReservedResourceID != resource.ID {
			t.Fatalf("missing reservation: %#v %v", stored, err)
		}
		prepared, err := scope.PrepareResources([]stagedobjects.Publication{{StagingID: sealed.StagingID, ResourceKind: resource.Kind, ResourceID: resource.ID, ByteSize: sealed.ByteSize, SHA256: sealed.SHA256}})
		if err != nil {
			t.Fatal(err)
		}
		return attempt{prepared, key, sealed.StagingID}
	}
	winner, loser := allocate("winner", "retained evidence"), allocate("loser", "competing bytes")
	if winner.key == loser.key {
		t.Fatal("competing attempts shared a physical key")
	}
	plan := winner.resources.Plan()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	capability, err := NewStagedPublicationCapability("winner", tx, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	altered := append([]stagedobjects.Publication(nil), plan.Publications...)
	altered[0].ResourceID = incident.String() + ":" + uuid.NewString()
	if err := stagedobjects.PublishTransferred(ctx, plan.Transfer, altered, capability); err == nil {
		t.Fatal("changed reserved owner published")
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	capability, err = NewStagedPublicationCapability("winner", tx, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	if err := stagedobjects.PublishTransferred(ctx, plan.Transfer, plan.Publications, capability); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := loser.resources.Abandon(ctx); err != nil {
		t.Fatal(err)
	}
	janitor, err := stagedobjects.NewJanitor(stagedobjects.JanitorOptions{Repository: repository, Bytes: bytes, Health: stagedobjects.NewHealth(), Now: func() time.Time { return now.Add(time.Minute) }, BatchLimit: 10, OperationTimeout: time.Minute, FatalSink: func(err error) { t.Errorf("janitor fatal: %v", err) }})
	if err != nil {
		t.Fatal(err)
	}
	if err := janitor.Sweep(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := objects.StatObject(ctx, winner.key); err != nil {
		t.Fatalf("committed winner deleted: %v", err)
	}
	if _, err := objects.StatObject(ctx, loser.key); err == nil {
		t.Fatal("abandoned loser remains")
	}
	published, err := physical.RedeemableStagedObject(ctx, winner.id, now.Add(48*time.Hour))
	if err != nil || published.State != extensionstore.StagedPublished {
		t.Fatalf("published owner lost: %#v %v", published, err)
	}
}
