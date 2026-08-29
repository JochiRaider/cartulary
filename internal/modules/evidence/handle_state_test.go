package evidence

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestHandleRedemptionRechecksCurrentState_Unit(t *testing.T) {
	harness := newTestStore(t, "evidence_lifecycle-handle-current-state")
	store := newTestRouteOperations(t, harness)
	actor := authstoretest.SeedLocalUserRecord(t, harness, "evidence_lifecycle-handle-current@example.test", "EvidenceLifecycle Handle Current", "EvidenceLifecycleHandleCurrent1!", false, false, true)
	incident := createTestIncident(t, harness, actor, "txn-evidence_lifecycle-handle-current-incident", "IR-P5-HANDLE-CURRENT", "Evidence handle current state")

	unsupportedRecordID := seedEvidenceAttachmentRecord(t, harness, incident.ID, actor.ID, "available")
	unsupportedBlobID := seedBlob(t, harness, incident.ID, actor.ID, "available", BlobOptions{
		ByteSize:            11,
		ObservedSize:        ptrInt64(11),
		ObservedContentType: ptrString("audio/wav"),
		ObservedSHA:         ptrString(strings.Repeat("a", 64)),
	})
	linkEvidenceBlobInStore(t, harness, unsupportedRecordID, unsupportedBlobID)
	unsupportedAccess, err := store.LoadEvidenceAccess(context.Background(), unsupportedRecordID)
	if err != nil {
		t.Fatalf("load unsupported preview access: %v", err)
	}
	if unsupportedAccess.PreviewKind != nil {
		t.Fatalf("audio evidence unexpectedly previewable: %#v", unsupportedAccess)
	}
	if reason, err := store.CheckHandleAccess(context.Background(), handleFromAccess(unsupportedAccess, "download")); err != nil || reason != "" {
		t.Fatalf("unsupported preview evidence should remain downloadable: reason=%q err=%v", reason, err)
	}

	scenarios := []struct {
		name       string
		reasonCode string
		arrange    func(recordID uuid.UUID) handleRecord
	}{
		{name: "pending blob", reasonCode: "blob_pending", arrange: func(recordID uuid.UUID) handleRecord {
			blobID := seedBlob(t, harness, incident.ID, actor.ID, "pending", BlobOptions{ByteSize: 11})
			linkEvidenceBlobInStore(t, harness, recordID, blobID)
			return currentStoreHandle(t, store, recordID, "preview")
		}},
		{name: "failed blob", reasonCode: "blob_failed", arrange: func(recordID uuid.UUID) handleRecord {
			blobID := seedBlob(t, harness, incident.ID, actor.ID, "failed", BlobOptions{ByteSize: 11})
			linkEvidenceBlobInStore(t, harness, recordID, blobID)
			return currentStoreHandle(t, store, recordID, "preview")
		}},
		{name: "quarantined", reasonCode: "evidence_quarantined", arrange: func(recordID uuid.UUID) handleRecord {
			blobID := seedBlob(t, harness, incident.ID, actor.ID, "quarantined", BlobOptions{ByteSize: 11})
			linkEvidenceBlobInStore(t, harness, recordID, blobID)
			return currentStoreHandle(t, store, recordID, "preview")
		}},
		{name: "inconsistent", reasonCode: "evidence_inconsistent", arrange: func(recordID uuid.UUID) handleRecord {
			blobID := seedBlob(t, harness, incident.ID, actor.ID, "available", BlobOptions{
				ByteSize:            11,
				ObservedSize:        ptrInt64(11),
				ObservedContentType: ptrString("text/plain"),
				ObservedSHA:         ptrString(strings.Repeat("b", 64)),
			})
			linkEvidenceBlobInStore(t, harness, recordID, blobID)
			updateEvidenceLifecycleInStore(t, harness, recordID, "received")
			return currentStoreHandle(t, store, recordID, "preview")
		}},
		{name: "row version drift", reasonCode: "evidence_inconsistent", arrange: func(recordID uuid.UUID) handleRecord {
			blobID := seedBlob(t, harness, incident.ID, actor.ID, "available", BlobOptions{
				ByteSize:            11,
				ObservedSize:        ptrInt64(11),
				ObservedContentType: ptrString("text/plain"),
				ObservedSHA:         ptrString(strings.Repeat("c", 64)),
			})
			linkEvidenceBlobInStore(t, harness, recordID, blobID)
			handle := currentStoreHandle(t, store, recordID, "preview")
			incrementRecordVersionInStore(t, harness, recordID)
			return handle
		}},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			recordID := seedEvidenceAttachmentRecord(t, harness, incident.ID, actor.ID, "available")
			handle := scenario.arrange(recordID)
			reason, err := store.CheckHandleAccess(context.Background(), handle)
			if err != nil {
				t.Fatalf("CheckHandleAccess: %v", err)
			}
			if reason != scenario.reasonCode {
				t.Fatalf("reason got %q want %q", reason, scenario.reasonCode)
			}
		})
	}
}

func linkEvidenceBlobInStore(t testing.TB, db postgres.DB, recordID uuid.UUID, blobID uuid.UUID) {
	t.Helper()
	if _, err := db.Exec(context.Background(), `
UPDATE evidence
   SET object_blob_id = $2,
       lifecycle_state = 'available',
       upload_state = (SELECT upload_state FROM object_blobs WHERE object_blob_id = $2),
       updated_at = now()
 WHERE record_id = $1
`, recordID, blobID); err != nil {
		t.Fatalf("link evidence blob in store: %v", err)
	}
}

func currentStoreHandle(t testing.TB, store *routeOperations, recordID uuid.UUID, kind string) handleRecord {
	t.Helper()
	access, err := store.LoadEvidenceAccess(context.Background(), recordID)
	if err != nil {
		t.Fatalf("load current handle access: %v", err)
	}
	return handleFromAccess(access, kind)
}

func handleFromAccess(access evidenceAccessRecord, kind string) handleRecord {
	disposition := "attachment"
	previewKind := access.PreviewKind
	if kind == "preview" {
		disposition = "inline"
	} else {
		previewKind = nil
	}
	var objectBlobID uuid.UUID
	if access.ObjectBlobID != nil {
		objectBlobID = *access.ObjectBlobID
	}
	var storageKey string
	if access.StorageKey != nil {
		storageKey = *access.StorageKey
	}
	return handleRecord{
		Token:                  "test-handle-" + uuid.NewString(),
		IncidentID:             access.IncidentID,
		RecordID:               access.RecordID,
		RecordRowVersion:       access.RecordRowVersion,
		ObjectBlobID:           objectBlobID,
		StorageKey:             storageKey,
		SessionID:              uuid.New(),
		HandleKind:             kind,
		MediaClass:             access.MediaClass,
		PreviewKind:            previewKind,
		Disposition:            disposition,
		Filename:               access.FilenameSource,
		ContentType:            access.ContentType,
		SizeBytes:              access.SizeBytes,
		SHA256:                 access.SHA256,
		EvidenceLifecycleState: access.EvidenceLifecycleState,
		UploadState:            access.UploadState,
	}
}

func updateEvidenceLifecycleInStore(t testing.TB, db postgres.DB, recordID uuid.UUID, lifecycleState string) {
	t.Helper()
	if _, err := db.Exec(context.Background(), `
UPDATE evidence
   SET lifecycle_state = $2,
       updated_at = now()
 WHERE record_id = $1
`, recordID, lifecycleState); err != nil {
		t.Fatalf("update evidence lifecycle in store: %v", err)
	}
}

func incrementRecordVersionInStore(t testing.TB, db postgres.DB, recordID uuid.UUID) {
	t.Helper()
	if _, err := db.Exec(context.Background(), `
UPDATE records
   SET row_version = row_version + 1,
       updated_at = now()
 WHERE record_id = $1
`, recordID); err != nil {
		t.Fatalf("increment record row version in store: %v", err)
	}
}
