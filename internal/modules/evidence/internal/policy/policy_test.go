package policy

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
)

func TestEvidenceLifecyclePolicy(t *testing.T) {
	t.Parallel()
	states := []string{
		EvidenceRequested,
		EvidencePendingReceipt,
		EvidenceReceived,
		EvidenceAvailable,
		EvidenceQuarantined,
		EvidenceReleased,
	}
	legal := map[string]map[string]bool{
		EvidenceRequested: {
			EvidenceRequested: true, EvidencePendingReceipt: true, EvidenceReceived: true, EvidenceAvailable: true,
		},
		EvidencePendingReceipt: {
			EvidenceRequested: true, EvidencePendingReceipt: true, EvidenceReceived: true, EvidenceAvailable: true,
		},
		EvidenceReceived: {
			EvidencePendingReceipt: true, EvidenceReceived: true, EvidenceAvailable: true, EvidenceQuarantined: true,
		},
		EvidenceAvailable: {
			EvidenceReceived: true, EvidenceAvailable: true, EvidenceQuarantined: true, EvidenceReleased: true,
		},
		EvidenceQuarantined: {
			EvidenceReceived: true, EvidenceAvailable: true, EvidenceQuarantined: true,
		},
		EvidenceReleased: {
			EvidenceAvailable: true, EvidenceQuarantined: true, EvidenceReleased: true,
		},
	}
	for _, from := range append(states, "invalid") {
		for _, to := range append(states, "invalid") {
			if got, want := LegalEvidenceTransition(from, to), legal[from][to]; got != want {
				t.Errorf("LegalEvidenceTransition(%q, %q) = %v, want %v", from, to, got, want)
			}
		}
	}
	for _, state := range states {
		if !ValidEvidenceLifecycle(state) {
			t.Errorf("ValidEvidenceLifecycle(%q) = false", state)
		}
	}
	if ValidEvidenceLifecycle("invalid") {
		t.Fatal("invalid Evidence lifecycle accepted")
	}
}

func TestInitialLifecycleAndBlobBridgePolicy(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		state         string
		blobSupplied  bool
		blobFinalized bool
		want          InitialLifecycleDisposition
	}{
		{EvidenceRequested, false, false, InitialLifecycleAllowed},
		{EvidenceRequested, true, true, InitialLifecycleAllowed},
		{EvidencePendingReceipt, false, false, InitialLifecycleAllowed},
		{EvidencePendingReceipt, true, true, InitialLifecycleAllowed},
		{EvidenceReceived, false, false, InitialLifecycleAllowed},
		{EvidenceReceived, true, true, InitialLifecycleAllowed},
		{EvidenceAvailable, false, false, InitialLifecycleGuardViolation},
		{EvidenceAvailable, true, true, InitialLifecycleAllowed},
		{EvidenceQuarantined, false, false, InitialLifecycleAllowed},
		{EvidenceQuarantined, true, true, InitialLifecycleGuardViolation},
		{EvidenceReleased, false, false, InitialLifecycleIllegalTransition},
		{EvidenceReleased, true, true, InitialLifecycleIllegalTransition},
		{"invalid", false, false, InitialLifecycleInvalid},
	} {
		if got := InitialEvidenceLifecycleDisposition(test.state, test.blobSupplied, test.blobFinalized); got != test.want {
			t.Errorf("InitialEvidenceLifecycleDisposition(%q, %v, %v) = %v, want %v", test.state, test.blobSupplied, test.blobFinalized, got, test.want)
		}
	}
	for _, test := range []struct {
		state     string
		hasBlob   bool
		blobState string
		want      bool
	}{
		{EvidenceRequested, false, "", false},
		{EvidenceRequested, true, BlobAvailable, false},
		{EvidenceReceived, true, BlobAvailable, false},
		{EvidenceAvailable, false, "", true},
		{EvidenceAvailable, true, BlobAvailable, false},
		{EvidenceAvailable, true, BlobQuarantined, true},
		{EvidenceReleased, true, BlobAvailable, false},
		{EvidenceQuarantined, false, "", false},
		{EvidenceQuarantined, true, BlobQuarantined, false},
		{EvidenceQuarantined, true, BlobAvailable, true},
	} {
		if got := ViolatesEvidenceBlobBridge(test.state, test.hasBlob, test.blobState); got != test.want {
			t.Errorf("ViolatesEvidenceBlobBridge(%q, %v, %q) = %v, want %v", test.state, test.hasBlob, test.blobState, got, test.want)
		}
	}
}

func TestAssociationFailureAndIncidentPolicy(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 11, 20, 0, 0, 0, time.FixedZone("test", -4*60*60))
	for _, test := range []struct {
		state   string
		expires time.Time
		want    AssociationBlobDisposition
	}{
		{BlobAvailable, now, AssociationBlobAvailable},
		{BlobPending, now.Add(time.Second), AssociationBlobNeedsFinalization},
		{BlobPending, now, AssociationBlobExpired},
		{BlobPending, now.Add(-time.Second), AssociationBlobExpired},
		{BlobFailed, now, AssociationBlobFailed},
		{BlobQuarantined, now, AssociationBlobQuarantined},
		{"invalid", now, AssociationBlobInconsistent},
	} {
		if got := ClassifyBlobForAssociation(test.state, test.expires, now); got != test.want {
			t.Errorf("ClassifyBlobForAssociation(%q) = %v, want %v", test.state, got, test.want)
		}
	}
	schedule := ScheduleFailure(now)
	if !schedule.FailedAt.Equal(now.UTC()) || !schedule.CleanupDueAt.Equal(now.UTC().Add(45*time.Minute)) {
		t.Fatalf("ScheduleFailure() = %+v", schedule)
	}
	if FinalizeAttemptIsTerminal(AllowedNonTerminalFinalizeFailures) || !FinalizeAttemptIsTerminal(TerminalFinalizeAttempt) {
		t.Fatal("finalize attempt threshold does not preserve three retries then terminal fourth failure")
	}
	if !IncidentMutationBlocked("closed") || IncidentMutationBlocked("active") {
		t.Fatal("incident mutation policy mismatch")
	}
	for _, test := range []struct {
		from    string
		to      string
		trigger string
		want    bool
	}{
		{BlobPending, BlobAvailable, "", true},
		{BlobPending, BlobFailed, "", true},
		{BlobPending, BlobQuarantined, QuarantineTriggerAdmin, false},
		{BlobAvailable, BlobQuarantined, QuarantineTriggerAdmin, true},
		{BlobAvailable, BlobQuarantined, QuarantineTriggerContentInspection, true},
		{BlobAvailable, BlobQuarantined, "unknown", false},
		{BlobQuarantined, BlobAvailable, QuarantineClearAdmin, true},
		{BlobQuarantined, BlobAvailable, QuarantineClearContentInspection, true},
		{BlobFailed, BlobPending, "", false},
	} {
		if got := LegalBlobTransition(test.from, test.to, test.trigger); got != test.want {
			t.Errorf("LegalBlobTransition(%q, %q, %q) = %v, want %v", test.from, test.to, test.trigger, got, test.want)
		}
	}
}

func TestReferencePolicy(t *testing.T) {
	t.Parallel()
	incidentID := uuid.New()
	objectBlobID := uuid.New()
	key, err := blobref.ObjectBlobStorageKey(incidentID, objectBlobID)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidatePersistedObjectBlobStorageKey(key, incidentID, objectBlobID); err != nil {
		t.Fatalf("valid storage key rejected: %v", err)
	}
	for _, test := range []struct {
		key        string
		incidentID uuid.UUID
		blobID     uuid.UUID
		wantReason string
	}{
		{"malformed", incidentID, objectBlobID, ObjectBlobStorageKeyMalformedReason},
		{key, uuid.New(), objectBlobID, ObjectBlobStorageKeyIdentityMismatchReason},
		{key, incidentID, uuid.New(), ObjectBlobStorageKeyIdentityMismatchReason},
	} {
		err := ValidatePersistedObjectBlobStorageKey(test.key, test.incidentID, test.blobID)
		if got, ok := PersistedObjectBlobStorageKeyErrorReason(err); !ok || got != test.wantReason {
			t.Errorf("storage key error = (%q, %v), want %q", got, ok, test.wantReason)
		}
	}
	ref, err := blobref.ObjectBlobStorageRef(objectBlobID)
	if err != nil {
		t.Fatal(err)
	}
	if !IsServerManagedStorageRef(" "+ref+" ") || !IsServerManagedStorageRef("object://not-a-uuid") || IsServerManagedStorageRef("ticket://external") {
		t.Fatal("server-managed storage-ref classification mismatch")
	}
	if !ServerManagedStorageRefMatchesAssociation(ref, &objectBlobID) ||
		ServerManagedStorageRefMatchesAssociation(ref, nil) ||
		ServerManagedStorageRefMatchesAssociation(ref, ptrUUID(uuid.New())) ||
		ServerManagedStorageRefMatchesAssociation("object://not-a-uuid", nil) ||
		!ServerManagedStorageRefMatchesAssociation("ticket://external", nil) {
		t.Fatal("server-managed storage-ref association policy mismatch")
	}
}

func ptrUUID(value uuid.UUID) *uuid.UUID { return &value }
