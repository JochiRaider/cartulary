package recovery_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

func TestSupportPhaseF_MigrationRunCanonicalStateGuardsAndRedaction(t *testing.T) {
	runID := uuid.MustParse("00000000-0000-0000-0000-000000130001")
	now := time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC)
	run, err := recovery.NewObjectStoreMigrationRun(runID, now, "operator@example.test", "http://source.internal:9000", "http://target.internal:8333", "private-bucket", "private-bucket")
	if err != nil {
		t.Fatalf("new migration run: %v", err)
	}
	if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventCopyStarted, now, nil); err == nil {
		t.Fatalf("copy_started from planned unexpectedly succeeded")
	}
	if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventPreflightPassed, now.Add(time.Second), nil); err != nil {
		t.Fatalf("preflight event: %v", err)
	}
	if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventBlockingFailure, now.Add(2*time.Second), map[string]string{"reason": "unit_test"}); err != nil {
		t.Fatalf("blocking failure event: %v", err)
	}
	if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventRollbackRequested, now.Add(3*time.Second), nil); err == nil {
		t.Fatalf("terminal run accepted another transition")
	}
	body, err := recovery.EncodeObjectStoreMigrationRun(run)
	if err != nil {
		t.Fatalf("encode migration run: %v", err)
	}
	if !bytes.HasSuffix(body, []byte("\n")) || bytes.Contains(body, []byte("private-bucket")) || bytes.Contains(body, []byte("source.internal")) {
		t.Fatalf("migration run body is not canonical redacted JSON: %s", body)
	}
	decoded, err := recovery.DecodeObjectStoreMigrationRun(body)
	if err != nil {
		t.Fatalf("decode migration run: %v", err)
	}
	if decoded.TerminalResult == nil || *decoded.TerminalResult != recovery.ObjectStoreMigrationStateFailed {
		t.Fatalf("terminal result not retained: %#v", decoded.TerminalResult)
	}
}

func TestSupportPhaseF_WriteQuiescenceRejectsOperatorAssertionOnly(t *testing.T) {
	valid := recovery.ObjectStoreMigrationWriteQuiescenceProof{
		SchemaID:                recovery.ObjectStoreMigrationProofSchemaID,
		ProofKind:               "process_stopped",
		CheckedAt:               time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC),
		ProcessState:            "absent",
		HTTPListenerClosed:      true,
		WebSocketListenerClosed: true,
	}
	if err := recovery.ValidateObjectStoreMigrationWriteQuiescenceProof(valid); err != nil {
		t.Fatalf("valid process_stopped proof rejected: %v", err)
	}
	valid.ProofKind = "operator_assertion_only"
	if err := recovery.ValidateObjectStoreMigrationWriteQuiescenceProof(valid); err == nil {
		t.Fatalf("operator_assertion_only proof unexpectedly accepted")
	}
}

func TestSupportPhaseF_MigrationValidationCanonicalDigestDuplicateKeysAndResult(t *testing.T) {
	ctx := context.Background()
	source := newMigrationFilesystemStore(t)
	target := newMigrationFilesystemStore(t)
	body := []byte("validation object")
	putMigrationObject(t, source, "objects/a.bin", body)
	putMigrationObject(t, target, "objects/a.bin", body)
	runID := uuid.MustParse("00000000-0000-0000-0000-000000130002")
	objects := []recovery.ObjectStoreMigrationBlob{{
		ObjectBlobID:       uuid.MustParse("00000000-0000-0000-0000-000000130003"),
		IncidentID:         uuid.MustParse("00000000-0000-0000-0000-000000130004"),
		StorageKey:         "objects/a.bin",
		EvidenceStorageRef: "object://00000000-0000-0000-0000-000000130003",
		ByteSize:           int64(len(body)),
	}}
	artifact, artifactBody, err := recovery.ValidateObjectStoreMigration(ctx, recovery.ObjectStoreMigrationValidationParams{
		RunID:         runID,
		StartedAt:     time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC),
		CompletedAt:   time.Date(2026, 6, 4, 10, 0, 1, 0, time.UTC),
		SourceBackend: recovery.ObjectStoreBackendMinIOS3,
		TargetBackend: recovery.ObjectStoreBackendSeaweedFSS3,
		SourceBucket:  "private-bucket",
		TargetBucket:  "private-bucket",
		SourceStore:   source,
		TargetStore:   target,
		Objects:       objects,
	})
	if err != nil {
		t.Fatalf("validate migration: %v", err)
	}
	if artifact.Result != "pass" {
		t.Fatalf("validation did not pass: %#v", artifact)
	}
	if !bytes.HasSuffix(artifactBody, []byte("\n")) || bytes.Contains(artifactBody, []byte(": ")) || bytes.Contains(artifactBody, []byte("\n  ")) {
		t.Fatalf("validation is not compact canonical JSON: %s", artifactBody)
	}
	decoded, err := recovery.DecodeObjectStoreMigrationValidation(artifactBody)
	if err != nil {
		t.Fatalf("decode validation: %v", err)
	}
	if decoded.ArtifactSHA256 != artifact.ArtifactSHA256 {
		t.Fatalf("artifact digest changed: got %s want %s", decoded.ArtifactSHA256, artifact.ArtifactSHA256)
	}
	tampered := decoded
	tampered.ArtifactSHA256 = strings.Repeat("0", 64)
	if err := recovery.ValidateObjectStoreMigrationValidation(tampered); err == nil {
		t.Fatalf("validation accepted artifact_sha256 that was not computed with digest omission")
	}
	duplicate := []byte(`{"schema_id":"cartulary.object_store_migration_validation.v1","schema_id":"cartulary.object_store_migration_validation.v1"}` + "\n")
	if _, err := recovery.DecodeObjectStoreMigrationValidation(duplicate); err == nil || !strings.Contains(err.Error(), "unique") {
		t.Fatalf("duplicate validation keys error got %v", err)
	}

	withBlocking := decoded
	withBlocking.BlockingDiagnostics = append(withBlocking.BlockingDiagnostics, recovery.ObjectStoreMigrationDiagnostic{
		DiagnosticID: "blocking-extra",
		Severity:     "blocking",
		ReasonCode:   "hash_mismatch",
		Message:      "blocking diagnostic",
		Refs:         []recovery.RedactionRef{},
	})
	if got := recovery.ComputeObjectStoreMigrationValidationResult(withBlocking); got != "fail" {
		t.Fatalf("blocking diagnostic result got %s want fail", got)
	}
	withSampleFailure := decoded
	withSampleFailure.PreviewSampleChecks = append(withSampleFailure.PreviewSampleChecks, recovery.ObjectStoreMigrationPreviewSampleCheck{
		ObjectBlobID: objects[0].ObjectBlobID.String(),
		IncidentID:   objects[0].IncidentID.String(),
		RouteClass:   "download",
		Status:       "fail",
		ReasonCode:   "download_handle_failed",
	})
	if got := recovery.ComputeObjectStoreMigrationValidationResult(withSampleFailure); got != "fail" {
		t.Fatalf("preview sample failure result got %s want fail", got)
	}
}

func TestSupportPhaseF_CopyLedgerStatusesAndZeroByteObjects(t *testing.T) {
	ctx := context.Background()
	runID := uuid.MustParse("00000000-0000-0000-0000-000000130005")
	cases := []struct {
		name         string
		sourceBody   []byte
		targetBody   []byte
		createSource bool
		createTarget bool
		wantStatus   recovery.ObjectStoreMigrationCopyStatus
		wantResult   string
	}{
		{name: "copied", sourceBody: []byte("copy me"), createSource: true, wantStatus: recovery.ObjectStoreMigrationCopyCopied, wantResult: "pass"},
		{name: "already_copied", sourceBody: []byte("same"), targetBody: []byte("same"), createSource: true, createTarget: true, wantStatus: recovery.ObjectStoreMigrationCopyAlreadyCopied, wantResult: "pass"},
		{name: "missing_source", createSource: false, wantStatus: recovery.ObjectStoreMigrationCopyMissingSource, wantResult: "fail"},
		{name: "target_mismatch", sourceBody: []byte("source"), targetBody: []byte("different"), createSource: true, createTarget: true, wantStatus: recovery.ObjectStoreMigrationCopyTargetMismatch, wantResult: "fail"},
		{name: "zero_byte", sourceBody: []byte{}, createSource: true, wantStatus: recovery.ObjectStoreMigrationCopyCopied, wantResult: "pass"},
	}
	for index, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := newMigrationFilesystemStore(t)
			target := newMigrationFilesystemStore(t)
			key := "objects/" + tc.name + ".bin"
			if tc.createSource {
				putMigrationObject(t, source, key, tc.sourceBody)
			}
			if tc.createTarget {
				putMigrationObject(t, target, key, tc.targetBody)
			}
			objectBlobID := uuid.MustParse("00000000-0000-0000-0000-000000130010")
			objectBlobID = uuid.MustParse(objectBlobID.String()[:35] + string(rune('0'+index)))
			ledger, ledgerBody, err := recovery.CopyObjectStoreMigrationObjects(ctx, recovery.ObjectStoreMigrationCopyParams{
				RunID:         runID,
				SourceBackend: recovery.ObjectStoreBackendMinIOS3,
				TargetBackend: recovery.ObjectStoreBackendSeaweedFSS3,
				SourceBucket:  "private-bucket",
				TargetBucket:  "private-bucket",
				SourceStore:   source,
				TargetStore:   target,
				Objects: []recovery.ObjectStoreMigrationBlob{{
					ObjectBlobID: objectBlobID,
					IncidentID:   uuid.MustParse("00000000-0000-0000-0000-000000130020"),
					StorageKey:   key,
					ByteSize:     int64(len(tc.sourceBody)),
				}},
			})
			if err != nil {
				t.Fatalf("copy objects: %v", err)
			}
			if ledger.Result != tc.wantResult || len(ledger.Items) != 1 || ledger.Items[0].Status != tc.wantStatus {
				t.Fatalf("ledger got result=%s status=%s want result=%s status=%s body=%s", ledger.Result, ledger.Items[0].Status, tc.wantResult, tc.wantStatus, ledgerBody)
			}
			if bytes.Contains(ledgerBody, []byte("private-bucket")) || bytes.Contains(ledgerBody, []byte(key)) {
				t.Fatalf("copy ledger leaked raw bucket/key: %s", ledgerBody)
			}
			if _, err := recovery.DecodeObjectStoreMigrationCopyLedger(ledgerBody); err != nil {
				t.Fatalf("decode copy ledger: %v", err)
			}
			if tc.wantStatus == recovery.ObjectStoreMigrationCopyCopied {
				reader, _, err := target.ReadObject(ctx, key, objectstore.ReadOptions{})
				if err != nil {
					t.Fatalf("read copied target: %v", err)
				}
				copied, readErr := ioReadAllAndClose(reader)
				if readErr != nil {
					t.Fatalf("read copied target body: %v", readErr)
				}
				if !bytes.Equal(copied, tc.sourceBody) {
					t.Fatalf("target bytes changed got %q want %q", copied, tc.sourceBody)
				}
			}
		})
	}
}

func TestSupportPhaseF_MigrationBlobReferencePreflight(t *testing.T) {
	objectBlobID := uuid.MustParse("00000000-0000-0000-0000-000000130030")
	incidentID := uuid.MustParse("00000000-0000-0000-0000-000000130031")
	validBlob := func() recovery.ObjectStoreMigrationBlob {
		return recovery.ObjectStoreMigrationBlob{
			ObjectBlobID:       objectBlobID,
			IncidentID:         incidentID,
			StorageKey:         "incidents/00000000-0000-0000-0000-000000130031/object-blobs/00000000-0000-0000-0000-000000130030",
			EvidenceStorageRef: "object://00000000-0000-0000-0000-000000130030",
			ByteSize:           1,
		}
	}
	valid := []recovery.ObjectStoreMigrationBlob{validBlob()}
	if err := recovery.ValidateObjectStoreMigrationBlobReferences(valid); err != nil {
		t.Fatalf("valid canonical migration refs rejected: %v", err)
	}

	mismatchedKey := []recovery.ObjectStoreMigrationBlob{validBlob()}
	mismatchedKey[0].StorageKey = "incidents/00000000-0000-0000-0000-000000130031/object-blobs/00000000-0000-0000-0000-000000130099"
	if err := recovery.ValidateObjectStoreMigrationBlobReferences(mismatchedKey); err == nil {
		t.Fatalf("mismatched canonical storage_key unexpectedly accepted")
	}

	mismatchedRef := []recovery.ObjectStoreMigrationBlob{validBlob()}
	mismatchedRef[0].EvidenceStorageRef = "object://00000000-0000-0000-0000-000000130099"
	if err := recovery.ValidateObjectStoreMigrationBlobReferences(mismatchedRef); err == nil {
		t.Fatalf("mismatched server-managed storage_ref unexpectedly accepted")
	}

	externalRef := []recovery.ObjectStoreMigrationBlob{validBlob()}
	externalRef[0].EvidenceStorageRef = "ticket://collect-legacy"
	if err := recovery.ValidateObjectStoreMigrationBlobReferences(externalRef); err != nil {
		t.Fatalf("external evidence storage_ref should remain migratable: %v", err)
	}
}

func newMigrationFilesystemStore(t testing.TB) *objectstore.FilesystemStore {
	t.Helper()
	store, err := objectstore.NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatalf("create filesystem object store: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})
	return store
}

func putMigrationObject(t testing.TB, store objectstore.Store, key string, body []byte) {
	t.Helper()
	if err := store.PutObject(context.Background(), key, bytes.NewReader(body), int64(len(body)), "application/octet-stream"); err != nil {
		t.Fatalf("put object %s: %v", key, err)
	}
}

func ioReadAllAndClose(reader io.ReadCloser) ([]byte, error) {
	body, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	if readErr != nil {
		return nil, readErr
	}
	return body, closeErr
}
