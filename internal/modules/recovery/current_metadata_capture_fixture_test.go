package recovery_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/testutil/recoverytest"
)

// Store-local tests use the shipped current encoder and real encrypted storage.
// This empty snapshot port does not establish database snapshot isolation; the
// service-backed capture tests exercise that separate boundary.
type currentMetadataCaptureFixture struct {
	Service *recovery.VNextCaptureService
	Store   *recovery.Store
	Storage recovery.StreamingBackupStorage
}

func newCurrentMetadataCapture(t *testing.T, store *recovery.Store) currentMetadataCaptureFixture {
	t.Helper()
	storage, err := recovery.RequireStreamingBackupStorage(newEncryptedBackupStorage(t, t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	state := currentStateCatalog(t)
	var providers []recovery.VNextObjectInventoryProvider
	for _, family := range state.Document().ObjectFamilies {
		providers = append(providers, recovery.NewVNextObjectInventoryProvider(
			family.OwnerID, family.ObjectFamilyID, *family.InventoryAlgorithmID,
			func(context.Context, recovery.VNextSnapshot) ([]recovery.VNextObjectMember, error) { return nil, nil },
		))
	}
	inventories, err := recovery.NewVNextObjectInventoryCatalog(state, providers...)
	if err != nil {
		t.Fatal(err)
	}
	service, err := recovery.NewVNextCaptureService(emptyMetadataSnapshotRepository{}, storage, state, inventories, testExtensionBackupCatalog(t), recoverytest.ProofScratch(t))
	if err != nil {
		t.Fatal(err)
	}
	return currentMetadataCaptureFixture{Service: service, Store: store, Storage: storage}
}

func (fixture currentMetadataCaptureFixture) CaptureAndPublish(ctx context.Context, params recovery.VNextCaptureParams) (recovery.BackupSet, error) {
	captured, err := fixture.Service.Capture(ctx, params)
	if err != nil {
		return recovery.BackupSet{}, err
	}
	return fixture.Store.PublishVNextCapturedBackup(ctx, captured)
}

type emptyMetadataSnapshotRepository struct{}

func (emptyMetadataSnapshotRepository) WithinRepeatableReadReadOnly(ctx context.Context, run func(recovery.VNextSnapshot, recovery.TargetControl) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return run(emptyMetadataSnapshot{}, recovery.TargetControl{SchemaID: recovery.TargetStateSchemaID, State: recovery.TargetInitialized, Revision: 1})
}

type emptyMetadataSnapshot struct{}

func (emptyMetadataSnapshot) StreamCanonicalTableRows(ctx context.Context, _ string, _ func(json.RawMessage) error) error {
	return ctx.Err()
}

func (emptyMetadataSnapshot) QueryRows(ctx context.Context, _ string, _ ...any) (recovery.VNextRows, error) {
	return emptyMetadataRows{}, ctx.Err()
}

type emptyMetadataRows struct{}

func (emptyMetadataRows) Next() bool        { return false }
func (emptyMetadataRows) Scan(...any) error { return errors.New("empty snapshot has no current row") }
func (emptyMetadataRows) Err() error        { return nil }
func (emptyMetadataRows) Close()            {}
