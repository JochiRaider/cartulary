package evidence

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
)

func TestQuarantineLifecycleService_Integration(t *testing.T) {
	harness := newTestStore(t, "evidence-quarantine-lifecycle")
	service := newTestBlobLifecycleService(t, harness)
	actor := authstoretest.SeedLocalUserRecord(t, harness, "evidence-quarantine@example.test", "Evidence Quarantine", "EvidenceQuarantine1!", false, false, true)
	incident := createTestIncident(t, harness, actor, "txn-evidence-quarantine-incident", "IR-EVIDENCE-QUARANTINE", "Evidence quarantine")

	recordID := seedEvidenceAttachmentRecord(t, harness, incident.ID, actor.ID, "available")
	blobID := seedBlob(t, harness, incident.ID, actor.ID, "available", BlobOptions{
		ByteSize:            4,
		ObservedSize:        ptrInt64(4),
		ObservedContentType: ptrString("text/plain"),
		ObservedSHA:         ptrString(strings.Repeat("a", 64)),
	})
	request := attachBlobRequest{ObjectBlobID: blobID, BaseRowVersion: 1, ClientTxnID: "txn-evidence-quarantine-attach"}
	if _, err := service.AttachBlob(context.Background(), actor, recordID, request, attachBlobRequestHash(request), nil, "req-evidence-quarantine-attach", time.Now().UTC()); err != nil {
		t.Fatalf("attach available blob: %v", err)
	}
	before := DurableAttachCounts(t, harness, recordID)

	if _, err := service.QuarantineBlob(context.Background(), actor.ID, blobID, "unsupported_trigger", "req-evidence-quarantine-invalid", time.Now().UTC()); !errors.Is(err, errIllegalBlobTransition) {
		t.Fatalf("unsupported quarantine trigger got %v want illegal transition", err)
	}
	result, err := service.QuarantineBlob(context.Background(), actor.ID, blobID, "content_inspection_quarantine", "req-evidence-quarantine", time.Now().UTC())
	if err != nil {
		t.Fatalf("quarantine available blob: %v", err)
	}
	if result.IncidentID != incident.ID || result.ObjectBlobID != blobID || result.ChangeSetID == uuid.Nil || result.ChangedEvidenceRecord != 1 {
		t.Fatalf("unexpected quarantine result: %#v", result)
	}
	if len(result.ChangedEvidenceRows) != 1 || result.ChangedEvidenceRows[0].RecordID != recordID {
		t.Fatalf("quarantine changes got %#v want primary record %s", result.ChangedEvidenceRows, recordID)
	}
	requireEvidenceState(t, harness, recordID, "quarantined", "quarantined", blobID)
	after := DurableAttachCounts(t, harness, recordID)
	if after.ChangeSets != before.ChangeSets+1 || after.Mutations != before.Mutations+1 || after.Revisions != before.Revisions+1 || after.BlobLinks != before.BlobLinks {
		t.Fatalf("quarantine durable history got before=%+v after=%+v", before, after)
	}

	pendingID := seedBlob(t, harness, incident.ID, actor.ID, "pending", BlobOptions{ByteSize: 1})
	if _, err := service.QuarantineBlob(context.Background(), actor.ID, pendingID, "admin_quarantine", "req-evidence-pending-quarantine", time.Now().UTC()); !errors.Is(err, errIllegalBlobTransition) {
		t.Fatalf("pending quarantine got %v want illegal transition", err)
	}
}
