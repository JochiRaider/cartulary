package recovery_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/recoverytest"
	"github.com/google/uuid"
)

func TestSupportObjectStoreBackup_ObjectStoreBackupManifestCanonicalAndSummaryRedacted(t *testing.T) {
	members := []recovery.VNextObjectMember{
		currentObjectMember("b", "private/objects/b.txt", []byte("second")),
		currentObjectMember("a", "private/objects/a.txt", []byte("first")),
	}
	fixture, captured := captureCurrentObjects(t, members)
	body := readCurrentArtifact(t, fixture.storage, captured.ObjectManifestProof)
	var compact bytes.Buffer
	if err := json.Compact(&compact, body); err != nil || !bytes.Equal(body, compact.Bytes()) {
		t.Fatal("current object manifest is not compact JSON")
	}
	var manifest recovery.VNextObjectStoreBackupManifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.SchemaID != recovery.ObjectStoreBackupManifestV2SchemaID || len(manifest.Objects) != 2 ||
		manifest.Objects[0].LogicalObjectID != "a" || manifest.Objects[1].LogicalObjectID != "b" {
		t.Fatalf("current logical object ordering: %#v", manifest.Objects)
	}
	for index, entry := range manifest.Objects {
		expected := members[1-index]
		if entry.StorageKey != expected.StorageKey || entry.PlaintextSHA256 != expected.PlaintextSHA256 || entry.PlaintextBytes != expected.PlaintextBytes {
			t.Fatalf("object proof changed: %#v", entry)
		}
	}
	found := false
	for _, proof := range captured.IntegrityManifest.Artifacts {
		if proof.Kind != "object_store_summary" {
			continue
		}
		found = true
		summaryBody := readCurrentArtifact(t, fixture.storage, currentStreamProof(proof))
		var summary recovery.VNextObjectStoreBackupSummary
		if err := json.Unmarshal(summaryBody, &summary); err != nil {
			t.Fatal(err)
		}
		if summary.SchemaID != recovery.ObjectStoreBackupSummaryV2SchemaID || summary.ObjectCount != 2 || summary.TotalPlaintextBytes != 11 || summary.ManifestSHA256 != manifest.ManifestSHA256 {
			t.Fatalf("current summary lost aggregate proof: %#v", summary)
		}
		for _, secret := range []string{"private/objects", "first", "second", "storage_key", "logical_object_id"} {
			if bytes.Contains(summaryBody, []byte(secret)) {
				t.Fatalf("summary contains private object detail %q", secret)
			}
		}
	}
	if !found {
		t.Fatal("capture omitted current summary")
	}
}

func TestSupportObjectStoreBackup_DuplicateArtifactKeysRejected(t *testing.T) {
	fixture, captured := captureCurrentObjects(t, nil)
	body := readCurrentArtifact(t, fixture.storage, captured.IntegrityProof)
	duplicate := append([]byte(`{"schema_id":"cartulary.backup_integrity_manifest.v4",`), body[1:]...)
	proof, err := fixture.storage.WriteArtifactStream(t.Context(), recovery.BackupArtifactStreamWriteRequest{
		LogicalRef: "malformed/integrity", EnvelopeRef: "malformed/integrity.envelope.json",
		ContentType: "application/json", Plaintext: bytes.NewReader(duplicate),
	})
	if err != nil {
		t.Fatal(err)
	}
	target := &vNextRestoreTargetFake{}
	if err := fixture.restore.Restore(t.Context(), target, proof); !errors.Is(err, recovery.ErrVNextBackup) {
		t.Fatalf("duplicate current artifact key error = %v", err)
	}
	if target.commits != 0 || target.rollbacks != 0 {
		t.Fatal("invalid artifact reached target mutation")
	}
}

func TestSupportObjectStoreBackup_CaptureCurrentArtifactsFromObjectStore(t *testing.T) {
	assertCurrentObjectRoundTrip(t, []byte("object-store backup bytes"))
}

func assertCurrentObjectRoundTrip(t *testing.T, payload []byte) {
	t.Helper()
	source, err := objectstore.NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = source.Close() })
	member := currentObjectMember("proof", "objects/proof.txt", payload)
	if err := source.PutObject(t.Context(), member.StorageKey, bytes.NewReader(payload), int64(len(payload)), member.ContentType); err != nil {
		t.Fatal(err)
	}
	member.Open = func(ctx context.Context) (io.ReadCloser, error) {
		return recoveryassembly.NewVNextObjectSource(source).OpenRecoveryObject(ctx, member.StorageKey)
	}
	fixture, captured := captureCurrentObjects(t, []recovery.VNextObjectMember{member})
	target := &vNextRestoreTargetFake{}
	if err := fixture.restore.Restore(t.Context(), target, captured.IntegrityProof); err != nil {
		t.Fatal(err)
	}
	if target.commits != 1 || target.rollbacks != 0 || !bytes.Equal(target.objects[member.StorageKey], payload) {
		t.Fatal("current encrypted capture/restore changed object bytes or atomic outcome")
	}
}

func currentObjectMember(id, key string, body []byte) recovery.VNextObjectMember {
	return recovery.VNextObjectMember{LogicalObjectID: id, StorageKey: key, ContentType: "text/plain",
		PlaintextBytes: int64(len(body)), PlaintextSHA256: recovery.SHA256String(string(body)),
		Open: func(context.Context) (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(body)), nil },
	}
}

type currentObjectCodecFixture struct {
	storage recovery.StreamingBackupStorage
	restore *recovery.VNextRestoreService
}

func captureCurrentObjects(t *testing.T, members []recovery.VNextObjectMember) (currentObjectCodecFixture, recovery.VNextCapturedBackup) {
	t.Helper()
	state := currentStateCatalog(t)
	storage, err := recovery.RequireStreamingBackupStorage(newEncryptedBackupStorage(t, t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	var providers []recovery.VNextObjectInventoryProvider
	for i, family := range state.Document().ObjectFamilies {
		var selected []recovery.VNextObjectMember
		if i == 0 {
			selected = members
		}
		providers = append(providers, recovery.NewVNextObjectInventoryProvider(family.OwnerID, family.ObjectFamilyID, *family.InventoryAlgorithmID,
			func(context.Context, recovery.VNextSnapshot) ([]recovery.VNextObjectMember, error) {
				return append([]recovery.VNextObjectMember(nil), selected...), nil
			}))
	}
	inventories, err := recovery.NewVNextObjectInventoryCatalog(state, providers...)
	if err != nil {
		t.Fatal(err)
	}
	extensions := emptyExtensionBackupCatalog(t)
	capture, err := recovery.NewVNextCaptureService(&vNextSnapshotRepositoryFake{}, storage, state, inventories, extensions, recoverytest.ProofScratch(t))
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 7, 24, 15, 0, 0, 0, time.UTC)
	captured, err := capture.Capture(t.Context(), recovery.VNextCaptureParams{BackupSetID: uuid.New(), ConsistencyPointAt: at, CreatedAt: at, RetainedUntil: at.Add(31 * 24 * time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	algorithms, err := recovery.NewVNextRestoreAlgorithmCatalog(state, recovery.RequiredVNextRestoreAlgorithmIDs(state)...)
	if err != nil {
		t.Fatal(err)
	}
	restore, err := recovery.NewVNextRestoreService(storage, state, algorithms, extensions)
	if err != nil {
		t.Fatal(err)
	}
	return currentObjectCodecFixture{storage, restore}, captured
}

func readCurrentArtifact(t testing.TB, storage recovery.StreamingBackupStorage, proof recovery.BackupArtifactStreamProof) []byte {
	t.Helper()
	var body bytes.Buffer
	if err := storage.ReadArtifactStream(t.Context(), proof, &body); err != nil {
		t.Fatal(err)
	}
	return body.Bytes()
}

func currentStreamProof(proof recovery.VNextArtifactProof) recovery.BackupArtifactStreamProof {
	return recovery.BackupArtifactStreamProof{LogicalRef: proof.LogicalRef, EnvelopeRef: proof.EnvelopeRef, ContentType: proof.ContentType,
		PlaintextSHA256: proof.PlaintextSHA256, PlaintextBytes: proof.PlaintextBytes,
		EnvelopeSHA256: proof.EnvelopeSHA256}
}
