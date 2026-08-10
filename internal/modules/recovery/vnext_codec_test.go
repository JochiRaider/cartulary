package recovery_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
)

func TestVNextCaptureRestoreCodecsRemainParallelAndCatalogDriven_Unit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	stateCatalog, err := recoveryassembly.CurrentRecoveryStateCatalog()
	if err != nil {
		t.Fatalf("build state catalog: %v", err)
	}
	body := bytes.Repeat([]byte("owner-object\n"), 1024)
	digest := sha256.Sum256(body)
	providers := make([]recovery.VNextObjectInventoryProvider, 0)
	for index, family := range stateCatalog.Document().ObjectFamilies {
		family := family
		inventory := func(context.Context, recovery.VNextSnapshot) ([]recovery.VNextObjectMember, error) {
			return []recovery.VNextObjectMember{}, nil
		}
		if index == 0 {
			inventory = func(context.Context, recovery.VNextSnapshot) ([]recovery.VNextObjectMember, error) {
				return []recovery.VNextObjectMember{{
					LogicalObjectID: "fixture-object",
					StorageKey:      "owners/fixture-object",
					ContentType:     "application/octet-stream",
					PlaintextBytes:  int64(len(body)),
					PlaintextSHA256: hex.EncodeToString(digest[:]),
					Open: func(context.Context) (io.ReadCloser, error) {
						return io.NopCloser(bytes.NewReader(body)), nil
					},
				}}, nil
			}
		}
		providers = append(providers, recovery.NewVNextObjectInventoryProvider(
			family.OwnerID,
			family.ObjectFamilyID,
			*family.InventoryAlgorithmID,
			inventory,
		))
	}
	if _, err := recovery.NewVNextObjectInventoryCatalog(
		stateCatalog,
		providers[:len(providers)-1]...,
	); !errors.Is(err, recovery.ErrVNextBackup) {
		t.Fatalf("incomplete object inventory error = %v, want %v", err, recovery.ErrVNextBackup)
	}
	inventories, err := recovery.NewVNextObjectInventoryCatalog(stateCatalog, providers...)
	if err != nil {
		t.Fatalf("build object inventories: %v", err)
	}

	rawStorage, err := recoveryassembly.NewFilesystemStorage(t.TempDir())
	if err != nil {
		t.Fatalf("create storage: %v", err)
	}
	t.Cleanup(func() { _ = rawStorage.Close() })
	key, err := recovery.ParseRecoveryEncryptionKey(vNextRecoveryMasterKey)
	if err != nil {
		t.Fatalf("parse key: %v", err)
	}
	byteStorage, err := recovery.NewEncryptedBackupStorage(rawStorage, key)
	if err != nil {
		t.Fatalf("create encrypted storage: %v", err)
	}
	streamingStorage, err := recovery.RequireStreamingBackupStorage(byteStorage)
	if err != nil {
		t.Fatalf("require streaming storage: %v", err)
	}
	snapshots := &vNextSnapshotRepositoryFake{
		rows: map[string][]json.RawMessage{
			stateCatalog.RequiredTableNames()[0]: {
				json.RawMessage(`{"z":"last","a":"first"}`),
			},
		},
	}
	capture, err := recovery.NewVNextCaptureService(
		snapshots,
		streamingStorage,
		stateCatalog,
		inventories,
	)
	if err != nil {
		t.Fatalf("create vNext capture: %v", err)
	}
	at := time.Date(2026, 7, 29, 20, 0, 0, 0, time.UTC)
	backupID := uuid.MustParse("10000000-0000-4000-8000-000000000008")
	captured, err := capture.Capture(ctx, recovery.VNextCaptureParams{
		BackupSetID: backupID, ConsistencyPointAt: at,
		CreatedAt: at.Add(time.Second), RetainedUntil: at.Add(30 * 24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("capture vNext backup: %v", err)
	}
	if snapshots.transactions != 1 {
		t.Fatalf("snapshot transactions = %d, want 1", snapshots.transactions)
	}
	wantArtifactProofs := len(stateCatalog.RequiredTableNames()) + 3
	if got := len(captured.IntegrityManifest.Artifacts); got != wantArtifactProofs {
		t.Fatalf("integrity artifact proofs = %d, want %d", got, wantArtifactProofs)
	}
	if captured.IntegrityManifest.SchemaID != recovery.BackupIntegrityManifestV3SchemaID ||
		recovery.BackupIntegrityManifestSchemaID != "cartulary.backup_integrity_manifest.v2" ||
		recovery.PostgresSnapshotArtifactSchemaID != "cartulary.postgres_snapshot_artifact.v1" {
		t.Fatalf("vNext capture changed historical writer schema identities")
	}

	algorithmIDs := recovery.RequiredVNextRestoreAlgorithmIDs(stateCatalog)
	algorithms, err := recovery.NewVNextRestoreAlgorithmCatalog(stateCatalog, algorithmIDs...)
	if err != nil {
		t.Fatalf("build restore algorithms: %v", err)
	}
	restore, err := recovery.NewVNextRestoreService(streamingStorage, stateCatalog, algorithms)
	if err != nil {
		t.Fatalf("create vNext restore: %v", err)
	}
	target := &vNextRestoreTargetFake{}
	if err := restore.Restore(ctx, target, captured.IntegrityProof); err != nil {
		t.Fatalf("restore vNext backup: %v", err)
	}
	if target.commits != 1 || target.rollbacks != 0 {
		t.Fatalf("atomic restore commits/rollbacks = %d/%d, want 1/0", target.commits, target.rollbacks)
	}
	if got := len(target.tables); got != 83 {
		t.Fatalf("restored tables = %d, want 83", got)
	}
	firstTable := stateCatalog.RequiredTableNames()[0]
	if got := string(target.rows[firstTable][0]); got != `{"a":"first","z":"last"}` {
		t.Fatalf("canonical restored row = %s", got)
	}
	if got := target.objects["owners/fixture-object"]; !bytes.Equal(got, body) {
		t.Fatalf("restored object differs")
	}
	sort.Strings(target.algorithms)
	if got, want := target.algorithms, algorithmIDs; !equalStrings(got, want) {
		t.Fatalf("restore algorithms = %v, want %v", got, want)
	}
}

const vNextRecoveryMasterKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="

type vNextSnapshotRepositoryFake struct {
	rows         map[string][]json.RawMessage
	transactions int
}

func (repository *vNextSnapshotRepositoryFake) WithinRepeatableReadReadOnly(
	_ context.Context,
	run func(recovery.VNextSnapshot) error,
) error {
	repository.transactions++
	return run(vNextSnapshotFake{rows: repository.rows})
}

type vNextSnapshotFake struct {
	rows map[string][]json.RawMessage
}

func (snapshot vNextSnapshotFake) StreamCanonicalTableRows(
	_ context.Context,
	tableName string,
	visit func(json.RawMessage) error,
) error {
	for _, row := range snapshot.rows[tableName] {
		if err := visit(row); err != nil {
			return err
		}
	}
	return nil
}

func (vNextSnapshotFake) QueryRows(
	context.Context,
	string,
	...any,
) (recovery.VNextRows, error) {
	return nil, errors.New("unexpected owner query")
}

type vNextRestoreTargetFake struct {
	tables     []string
	rows       map[string][]json.RawMessage
	objects    map[string][]byte
	algorithms []string
	commits    int
	rollbacks  int
}

func (target *vNextRestoreTargetFake) WithAtomicRestore(
	ctx context.Context,
	run func(recovery.VNextRestoreMutation) error,
) error {
	mutation := &vNextRestoreMutationFake{
		rows:    make(map[string][]json.RawMessage),
		objects: make(map[string][]byte),
	}
	if err := run(mutation); err != nil {
		target.rollbacks++
		return err
	}
	target.tables = mutation.tables
	target.rows = mutation.rows
	target.objects = mutation.objects
	target.algorithms = mutation.algorithms
	target.commits++
	return ctx.Err()
}

type vNextRestoreMutationFake struct {
	tables     []string
	rows       map[string][]json.RawMessage
	objects    map[string][]byte
	algorithms []string
}

func (mutation *vNextRestoreMutationFake) PreparePostgresTables(
	_ context.Context,
	tableNames []string,
) error {
	mutation.tables = append([]string(nil), tableNames...)
	return nil
}

func (mutation *vNextRestoreMutationFake) InsertPostgresRow(
	_ context.Context,
	tableName string,
	row json.RawMessage,
) error {
	mutation.rows[tableName] = append(mutation.rows[tableName], append(json.RawMessage(nil), row...))
	return nil
}

func (*vNextRestoreMutationFake) FinishPostgresTable(context.Context, string) error {
	return nil
}

func (mutation *vNextRestoreMutationFake) RestoreObject(
	_ context.Context,
	object recovery.VNextObjectManifestEntry,
	reader io.Reader,
) error {
	body, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	mutation.objects[object.StorageKey] = body
	return nil
}

func (mutation *vNextRestoreMutationFake) RunCatalogAlgorithm(
	_ context.Context,
	algorithmID string,
) error {
	mutation.algorithms = append(mutation.algorithms, algorithmID)
	return nil
}

func equalStrings(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
