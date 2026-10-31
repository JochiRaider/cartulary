package recoveryassembly

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	extensionsrecovery "github.com/JochiRaider/cartulary/internal/modules/extensions/recoverycontribution"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestExtensionRecoveryInventorySnapshot_Integration(t *testing.T) {
	runtimeDB, recoveryDB := extensionInventoryDatabase(t)
	ctx := t.Context()
	// Insert out of order, using only lifecycle states admitted by the real DDL.
	// No Coordinator, claim set, or profile callbacks participate in enumeration.
	for _, row := range []struct{ id, state, deletion string }{
		{"z_published", "published", "not_applicable"},
		{"b_allocated", "allocated", "not_applicable"},
		{"d_deleted", "abandoned", "deleted"},
		{"c_pending", "abandoned", "pending"},
		{"a_ready", "ready", "not_applicable"},
	} {
		insertExtensionInventoryRow(t, runtimeDB, row.id, row.state, row.deletion)
	}
	source := &extensionInventoryBytes{bodies: map[string][]byte{
		"staged/a_ready":     extensionInventoryBody("a_ready"),
		"staged/z_published": extensionInventoryBody("z_published"),
		"staged/m_new":       extensionInventoryBody("m_new"),
	}}
	provider := extensionsrecovery.VNextRecoveryObjectInventory(source)
	repository := NewVNextSnapshotRepository(recoveryDB)
	err := repository.WithinRepeatableReadReadOnly(ctx, func(snapshot recovery.VNextSnapshot) error {
		for query, want := range map[string]string{"SHOW transaction_isolation": "repeatable read", "SHOW transaction_read_only": "on"} {
			rows, err := snapshot.QueryRows(ctx, query)
			if err != nil {
				return err
			}
			var got string
			if !rows.Next() {
				rows.Close()
				t.Fatal("missing transaction setting")
			}
			err = rows.Scan(&got)
			rows.Close()
			if err != nil {
				return err
			}
			if got != want {
				t.Fatalf("%s = %q, want %q", query, got, want)
			}
		}
		first, err := provider.Inventory(ctx, snapshot)
		if err != nil {
			return err
		}
		if source.opens != 0 {
			t.Fatal("enumeration opened object bytes")
		}
		assertExtensionInventoryMembers(t, first, []string{"a_ready", "z_published"})

		// A second connection commits while the original snapshot remains open.
		// Synchronous commit gives an explicit ordering boundary without sleeps.
		writer, err := runtimeDB.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return err
		}
		defer func() { _ = writer.Rollback(ctx) }()
		if _, err := writer.Exec(ctx, `UPDATE public.extension_staged_objects
SET state='abandoned', ready_at=NULL, abandoned_at=staged_at + interval '2 hours',
    delete_state='pending', next_delete_attempt_at=staged_at + interval '2 hours'
WHERE staging_id='a_ready'`); err != nil {
			return err
		}
		insertExtensionInventoryRow(t, writer, "m_new", "ready", "not_applicable")
		if err := writer.Commit(ctx); err != nil {
			return err
		}
		second, err := provider.Inventory(ctx, snapshot)
		if err != nil {
			return err
		}
		assertExtensionInventoryMembers(t, second, []string{"a_ready", "z_published"})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.WithinRepeatableReadReadOnly(ctx, func(snapshot recovery.VNextSnapshot) error {
		members, err := provider.Inventory(ctx, snapshot)
		if err != nil {
			return err
		}
		assertExtensionInventoryMembers(t, members, []string{"m_new", "z_published"})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestExtensionRecoveryCaptureRejectsInvalidBytes_Integration(t *testing.T) {
	runtimeDB, recoveryDB := extensionInventoryDatabase(t)
	insertExtensionInventoryRow(t, runtimeDB, "capture", "published", "not_applicable")
	state, err := CurrentRecoveryStateCatalog()
	if err != nil {
		t.Fatal(err)
	}
	if state.DigestSHA256() != "1675a4c47a153ecafb2514ec80855b201941e69dc5816402e7c0bb1f4ca1b555" {
		t.Fatal("Recovery compatibility identity changed")
	}
	wantBody := extensionInventoryBody("capture")
	unavailable := errors.New("staged object unavailable")
	for _, name := range []string{"valid", "unavailable", "size_mismatch", "digest_mismatch"} {
		t.Run(name, func(t *testing.T) {
			body := append([]byte(nil), wantBody...)
			source := &extensionInventoryBytes{bodies: map[string][]byte{}}
			switch name {
			case "unavailable":
				source.openErr = unavailable
			case "size_mismatch":
				body = append(body, '!')
			case "digest_mismatch":
				body[0] ^= 1
			}
			source.bodies["staged/capture"] = body
			inventories, err := CurrentVNextObjectInventoryCatalog(source, source, source)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := NewFilesystemStorage(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = raw.Close() })
			key, err := recovery.ParseRecoveryEncryptionKey(recoveryEvidenceTestKey)
			if err != nil {
				t.Fatal(err)
			}
			encrypted, err := recovery.NewEncryptedBackupStorage(raw, key)
			if err != nil {
				t.Fatal(err)
			}
			storage, err := recovery.RequireStreamingBackupStorage(encrypted)
			if err != nil {
				t.Fatal(err)
			}
			observedStorage := &extensionCaptureStorage{StreamingBackupStorage: storage}
			capture, err := recovery.NewVNextCaptureService(NewVNextSnapshotRepository(recoveryDB), observedStorage, state, inventories)
			if err != nil {
				t.Fatal(err)
			}
			at := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
			captured, err := capture.Capture(t.Context(), recovery.VNextCaptureParams{
				BackupSetID: uuid.New(), ConsistencyPointAt: at, CreatedAt: at, RetainedUntil: at.Add(24 * time.Hour),
			})
			if source.opens != 1 {
				t.Fatalf("capture did not reach the real staged inventory: opens=%d, error=%v", source.opens, err)
			}
			if name != "valid" {
				wantErr := recovery.ErrVNextBackup
				if name == "unavailable" {
					wantErr = unavailable
				}
				if !errors.Is(err, wantErr) || !reflect.DeepEqual(captured, recovery.VNextCapturedBackup{}) {
					t.Fatalf("invalid bytes yielded backup evidence or wrong failure: %v / %#v", err, captured)
				}
				for _, ref := range observedStorage.written {
					if strings.HasSuffix(ref, "/integrity_manifest.json") || strings.HasSuffix(ref, "/object_store_manifest.json") {
						t.Fatalf("invalid bytes published a completion manifest: %s", ref)
					}
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if captured.IntegrityManifest.RecoveryStateCatalogSHA256 != state.DigestSHA256() || captured.IntegrityProof.PlaintextSHA256 == "" {
				t.Fatal("successful backup lacks catalog-bound integrity evidence")
			}
			var output bytes.Buffer
			if err := storage.ReadArtifactStream(t.Context(), captured.ObjectManifestProof, &output); err != nil {
				t.Fatal(err)
			}
			var manifest recovery.VNextObjectStoreBackupManifest
			if err := json.Unmarshal(output.Bytes(), &manifest); err != nil {
				t.Fatal(err)
			}
			if len(manifest.Objects) != 1 {
				t.Fatalf("captured objects = %d, want 1", len(manifest.Objects))
			}
			member := manifest.Objects[0]
			if member.OwnerID != "module.extensions" || member.ObjectFamilyID != "extensions.staged_objects" || member.LogicalObjectID != "capture" || member.StorageKey != "staged/capture" || member.PlaintextBytes != int64(len(wantBody)) || member.PlaintextSHA256 != extensionInventoryDigest(wantBody) {
				t.Fatalf("capture lost source-owner object facts: %#v", member)
			}
		})
	}
}

func extensionInventoryDatabase(t *testing.T) (postgres.AdmittedPool, postgres.AdmittedPool) {
	t.Helper()
	db := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "extension-recovery-inventory")
	open := func(purpose postgres.Purpose, role string) postgres.AdmittedPool {
		t.Helper()
		dsn, err := db.DSNForPurpose(purpose)
		if err != nil {
			t.Fatal(err)
		}
		pool, err := postgres.Setup(t.Context(), postgres.Settings{BindingKind: "managed_service", DSN: dsn, Purpose: purpose, ExpectedRole: role})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(pool.Close)
		return pool
	}
	return open(postgres.PurposeRuntime, "cartulary_runtime"), open(postgres.PurposeRecovery, "cartulary_recovery")
}

type extensionInventoryWriter interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func insertExtensionInventoryRow(t *testing.T, db extensionInventoryWriter, id, state, deletion string) {
	t.Helper()
	at := time.Date(2026, 10, 31, 1, 0, 0, 0, time.UTC)
	var ready, published, abandoned, nextDelete any
	switch state {
	case "ready":
		ready = at.Add(time.Minute)
	case "published":
		ready, published = at.Add(time.Minute), at.Add(2*time.Minute)
	case "abandoned":
		abandoned = at.Add(time.Minute)
		if deletion == "pending" {
			nextDelete = at.Add(2 * time.Minute)
		}
	}
	body := extensionInventoryBody(id)
	if _, err := db.Exec(t.Context(), `INSERT INTO public.extension_staged_objects
(staging_id, owner_profile_id, storage_identity, expected_byte_size, expected_sha256,
 staged_at, staging_expires_at, ready_at, published_at, abandoned_at, state,
 delete_state, delete_attempt_count, next_delete_attempt_at, last_delete_error_code)
VALUES ($1,'import',$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,0,$12,NULL)`,
		id, "staged/"+id, int64(len(body)), extensionInventoryDigest(body), at, at.Add(24*time.Hour), ready, published, abandoned, state, deletion, nextDelete,
	); err != nil {
		t.Fatalf("insert lifecycle fixture %s: %v", id, err)
	}
}

func assertExtensionInventoryMembers(t *testing.T, members []recovery.VNextObjectMember, ids []string) {
	t.Helper()
	if len(members) != len(ids) {
		t.Fatalf("inventory count = %d, want %d", len(members), len(ids))
	}
	for i, id := range ids {
		member, want := members[i], extensionInventoryBody(id)
		if member.LogicalObjectID != id || member.StorageKey != "staged/"+id || member.PlaintextBytes != int64(len(want)) || member.PlaintextSHA256 != extensionInventoryDigest(want) || member.ContentType != "application/octet-stream" {
			t.Fatalf("ordered inventory member %d = %#v, want %s", i, member, id)
		}
		reader, err := member.Open(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		body, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil || closeErr != nil || !bytes.Equal(body, want) {
			t.Fatalf("staged bytes %s differ: %v / %v", id, readErr, closeErr)
		}
	}
}

func extensionInventoryBody(id string) []byte { return []byte("immutable staged bytes: " + id) }
func extensionInventoryDigest(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

type extensionInventoryBytes struct {
	bodies  map[string][]byte
	openErr error
	opens   int
}

func (s *extensionInventoryBytes) OpenRecoveryObject(_ context.Context, key string) (io.ReadCloser, error) {
	s.opens++
	if s.openErr != nil {
		return nil, s.openErr
	}
	body, ok := s.bodies[key]
	if !ok {
		return nil, errors.New("unexpected inventory object")
	}
	return io.NopCloser(bytes.NewReader(body)), nil
}
func (s *extensionInventoryBytes) StatRecoveryObject(_ context.Context, key string) (recovery.VNextObjectSourceInfo, error) {
	body, ok := s.bodies[key]
	if !ok {
		return recovery.VNextObjectSourceInfo{}, errors.New("unexpected inventory object")
	}
	return recovery.VNextObjectSourceInfo{PlaintextBytes: int64(len(body)), ContentType: "application/octet-stream"}, nil
}

type extensionCaptureStorage struct {
	recovery.StreamingBackupStorage
	written []string
}

func (s *extensionCaptureStorage) WriteArtifactStream(ctx context.Context, request recovery.BackupArtifactStreamWriteRequest) (recovery.BackupArtifactStreamProof, error) {
	proof, err := s.StreamingBackupStorage.WriteArtifactStream(ctx, request)
	if err == nil {
		s.written = append(s.written, request.LogicalRef)
	}
	return proof, err
}
