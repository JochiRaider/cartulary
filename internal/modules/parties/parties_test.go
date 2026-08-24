package parties_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

type partyTestAttributionResolver struct{}

type partyFieldValue struct{ Text *string }

type partyCreateRequest struct {
	ViewSchemaID string
	ClientTxnID  string
	Values       map[string]partyFieldValue
}

type partyPatchRequest struct {
	ViewSchemaID   string
	BaseRowVersion int64
	ClientTxnID    string
	Changes        []partyPatchChange
}

type partyPatchChange struct {
	FieldKey string
	Value    *partyFieldValue
}

func (partyTestAttributionResolver) ResolveImportedSourceActorsTx(context.Context, pgx.Tx, uuid.UUID, string, string, []string) (map[string]string, error) {
	return map[string]string{}, nil
}

func TestPartyExactMatchReuseAndRawTextPreservation_Unit(t *testing.T) {
	harness := appsupport.StartStore(t, "workbook_interaction-u-9-05-parties")
	evidenceOwner := appsupport.NewEvidenceMutationOwner(harness.DB, conflicttest.NewCodec("workbook"))
	partyOwner := appsupport.NewPartyOwner(harness.DB, conflicttest.NewCodec("workbook"))
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "u905@example.test", "U905 Parties", "U905PartiesPass1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-u-9-05-incident", "IR-U905", "Workbook inspector party-storage")
	otherIncident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-u-9-05-other-incident", "IR-U905B", "Workbook inspector party-storage Other")

	createdByEmail, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, partyCreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-email",
		Values: map[string]partyFieldValue{
			"party.display_name":  textChange("IR Vendor"),
			"party.party_kind":    textChange("organization"),
			"party.primary_email": textChange("Vendor@Example.Test"),
		},
	}, "req-workbook_interaction-u-9-05-party-email", time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create party by email: %v", err)
	}
	reusedByEmail, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, partyCreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-email-reuse",
		Values: map[string]partyFieldValue{
			"party.display_name":  textChange("Phone-like label must not drive reuse"),
			"party.party_kind":    textChange("person"),
			"party.primary_email": textChange(" vendor@example.test "),
		},
	}, "req-workbook_interaction-u-9-05-party-email-reuse", time.Date(2026, 5, 18, 12, 1, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("reuse party by email: %v", err)
	}
	if reusedByEmail.RecordID != createdByEmail.RecordID || reusedByEmail.Outcome != parties.MutationReused {
		t.Fatalf("expected normalized same-incident email reuse, got created=%#v reused=%#v", createdByEmail, reusedByEmail)
	}
	requirePartyCount(t, harness, incident.ID, "lower(primary_email) = lower('vendor@example.test')", 1)
	emailCaseNoop, admissionErr := admitPartyPatchRequest(partyPatchRequest{
		ViewSchemaID: parties.ViewSchemaID, BaseRowVersion: 1,
		ClientTxnID: "txn-workbook_interaction-u-9-05-party-email-case-noop",
		Changes: []partyPatchChange{{
			FieldKey: "party.primary_email", Value: fieldValuePointer(textChange("vendor@example.test")),
		}},
	})
	if admissionErr != nil {
		t.Fatalf("admit Party email-case no-op: %v", admissionErr)
	}
	_, err = partyOwner.Patch(context.Background(), parties.PatchCommand{
		ActorUserID: actor.ID, RecordID: createdByEmail.RecordID, Admission: emailCaseNoop,
		RequestID: "req-workbook_interaction-u-9-05-party-email-case-noop",
		RouteKey:  "workbook.records.patch", ConflictRouteKey: "workbook.records.conflicts.resolve",
		Now: time.Date(2026, 5, 18, 12, 1, 30, 0, time.UTC),
	})
	var noEffect *parties.ValidationError
	if !errors.As(err, &noEffect) || noEffect.ReasonCode != "no_effective_change" {
		t.Fatalf("email equality no-op error = %v, want no_effective_change", err)
	}
	var storedEmail string
	var emailRowVersion int64
	if err := harness.DB.QueryRow(context.Background(), `
SELECT p.primary_email, r.row_version
  FROM parties p JOIN records r ON r.record_id = p.record_id
 WHERE p.record_id = $1
`, createdByEmail.RecordID).Scan(&storedEmail, &emailRowVersion); err != nil {
		t.Fatalf("load Party after email equality no-op: %v", err)
	}
	if storedEmail != "Vendor@Example.Test" || emailRowVersion != 1 {
		t.Fatalf("email equality no-op changed source state: email=%q version=%d", storedEmail, emailRowVersion)
	}
	var emailNoopReplayRows int
	if err := harness.DB.QueryRow(context.Background(), `
SELECT count(*) FROM route_idempotency
 WHERE route_key = 'workbook.records.patch'
   AND scope_key = $1
   AND client_txn_id = 'txn-workbook_interaction-u-9-05-party-email-case-noop'
`, createdByEmail.RecordID.String()).Scan(&emailNoopReplayRows); err != nil || emailNoopReplayRows != 0 {
		t.Fatalf("email equality no-op retained replay state: count=%d err=%v", emailNoopReplayRows, err)
	}

	createdByExternalRef, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, partyCreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-external-ref",
		Values: map[string]partyFieldValue{
			"party.display_name": textChange("Outside Counsel"),
			"party.party_kind":   textChange("organization"),
			"party.external_ref": textChange("EXT-42"),
		},
	}, "req-workbook_interaction-u-9-05-party-external-ref", time.Date(2026, 5, 18, 12, 2, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create party by external ref: %v", err)
	}
	reusedByExternalRef, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, partyCreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-external-ref-reuse",
		Values: map[string]partyFieldValue{
			"party.display_name": textChange("Outside Counsel Alias"),
			"party.party_kind":   textChange("organization"),
			"party.external_ref": textChange(" EXT-42 "),
		},
	}, "req-workbook_interaction-u-9-05-party-external-ref-reuse", time.Date(2026, 5, 18, 12, 3, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("reuse party by external ref: %v", err)
	}
	if reusedByExternalRef.RecordID != createdByExternalRef.RecordID || reusedByExternalRef.Outcome != parties.MutationReused {
		t.Fatalf("expected normalized same-incident external_ref reuse, got created=%#v reused=%#v", createdByExternalRef, reusedByExternalRef)
	}
	requirePartyCount(t, harness, incident.ID, "trim(external_ref) = 'EXT-42'", 1)
	caseVariantExternalRef, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, partyCreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-external-ref-case-variant",
		Values: map[string]partyFieldValue{
			"party.display_name": textChange("Outside Counsel Case Variant"),
			"party.party_kind":   textChange("organization"),
			"party.external_ref": textChange("ext-42"),
		},
	}, "req-workbook_interaction-u-9-05-party-external-ref-case-variant", time.Date(2026, 5, 18, 12, 3, 10, 0, time.UTC))
	if err != nil {
		t.Fatalf("create case-sensitive external ref variant: %v", err)
	}
	if caseVariantExternalRef.RecordID == createdByExternalRef.RecordID || caseVariantExternalRef.Outcome != parties.MutationCreated {
		t.Fatalf("external_ref case variant must create a distinct Party: %#v", caseVariantExternalRef)
	}
	conflictingPatch := partyPatchRequest{
		ViewSchemaID: parties.ViewSchemaID, BaseRowVersion: 1,
		ClientTxnID: "txn-workbook_interaction-u-9-05-party-external-ref-conflict",
		Changes: []partyPatchChange{{
			FieldKey: "party.external_ref", Value: fieldValuePointer(textChange("EXT-42")),
		}},
	}
	conflictingAdmission, admissionErr := admitPartyPatchRequest(conflictingPatch)
	if admissionErr != nil {
		t.Fatalf("admit conflicting Party patch: %v", admissionErr)
	}
	_, err = partyOwner.Patch(context.Background(), parties.PatchCommand{
		ActorUserID: actor.ID, RecordID: caseVariantExternalRef.RecordID, Admission: conflictingAdmission,
		RequestID: "req-workbook_interaction-u-9-05-party-external-ref-conflict",
		RouteKey:  "workbook.records.patch", ConflictRouteKey: "workbook.records.conflicts.resolve",
		Now: time.Date(2026, 5, 18, 12, 3, 20, 0, time.UTC),
	})
	var keyClaimed *parties.PartyMatchConflictError
	if !errors.As(err, &keyClaimed) || keyClaimed.ReasonCode != parties.PartyMatchExactKeyClaimed || strings.Join(keyClaimed.ConflictingFieldKeys, ",") != "party.external_ref" {
		t.Fatalf("expected value-free external_ref claim conflict, got %v", err)
	}
	requirePartyCount(t, harness, incident.ID, "p.record_id = '"+caseVariantExternalRef.RecordID.String()+"' AND external_ref = 'ext-42'", 1)

	clearExternalRef := partyPatchRequest{
		ViewSchemaID: parties.ViewSchemaID, BaseRowVersion: 1,
		ClientTxnID: "txn-workbook_interaction-u-9-05-party-external-ref-clear",
		Changes: []partyPatchChange{{
			FieldKey: "party.external_ref", Value: &partyFieldValue{},
		}},
	}
	clearAdmission, admissionErr := admitPartyPatchRequest(clearExternalRef)
	if admissionErr != nil {
		t.Fatalf("admit Party external-ref clear: %v", admissionErr)
	}
	if _, err := partyOwner.Patch(context.Background(), parties.PatchCommand{
		ActorUserID: actor.ID, RecordID: caseVariantExternalRef.RecordID, Admission: clearAdmission,
		RequestID: "req-workbook_interaction-u-9-05-party-external-ref-clear",
		RouteKey:  "workbook.records.patch", ConflictRouteKey: "workbook.records.conflicts.resolve",
		Now: time.Date(2026, 5, 18, 12, 3, 30, 0, time.UTC),
	}); err != nil {
		t.Fatalf("clear Party external_ref claim: %v", err)
	}
	replacementCaseVariant, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, partyCreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-external-ref-reacquire",
		Values: map[string]partyFieldValue{
			"party.display_name": textChange("Outside Counsel Replacement"),
			"party.party_kind":   textChange("organization"),
			"party.external_ref": textChange("ext-42"),
		},
	}, "req-workbook_interaction-u-9-05-party-external-ref-reacquire", time.Date(2026, 5, 18, 12, 3, 40, 0, time.UTC))
	if err != nil || replacementCaseVariant.Outcome != parties.MutationCreated || replacementCaseVariant.RecordID == caseVariantExternalRef.RecordID {
		t.Fatalf("reacquire released Party external_ref: result=%#v err=%v", replacementCaseVariant, err)
	}

	displayOnly, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, partyCreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-display-only",
		Values: map[string]partyFieldValue{
			"party.display_name":      textChange("Duplicate Display"),
			"party.party_kind":        textChange("person"),
			"party.organization_name": textChange("Duplicate Org"),
			"party.role_title":        textChange("Duplicate Role"),
		},
	}, "req-workbook_interaction-u-9-05-party-display-only", time.Date(2026, 5, 18, 12, 3, 30, 0, time.UTC))
	if err != nil {
		t.Fatalf("create display-only party: %v", err)
	}
	displayOnlyAgain, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, partyCreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-display-only-again",
		Values: map[string]partyFieldValue{
			"party.display_name":      textChange("Duplicate Display"),
			"party.party_kind":        textChange("person"),
			"party.organization_name": textChange("Duplicate Org"),
			"party.role_title":        textChange("Duplicate Role"),
		},
	}, "req-workbook_interaction-u-9-05-party-display-only-again", time.Date(2026, 5, 18, 12, 3, 45, 0, time.UTC))
	if err != nil {
		t.Fatalf("create second display-only party: %v", err)
	}
	if displayOnlyAgain.RecordID == displayOnly.RecordID {
		t.Fatalf("display name, organization, and role title must not drive implicit reuse")
	}
	requirePartyCount(t, harness, incident.ID, "display_name = 'Duplicate Display'", 2)

	phoneLike, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, partyCreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-phone-like",
		Values: map[string]partyFieldValue{
			"party.display_name":      textChange("+1 555 0100"),
			"party.party_kind":        textChange("person"),
			"party.organization_name": textChange("+1 555 0100"),
			"party.role_title":        textChange("+1 555 0100"),
		},
	}, "req-workbook_interaction-u-9-05-party-phone-like", time.Date(2026, 5, 18, 12, 3, 46, 0, time.UTC))
	if err != nil {
		t.Fatalf("create phone-like party: %v", err)
	}
	phoneLikeAgain, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, partyCreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-phone-like-again",
		Values: map[string]partyFieldValue{
			"party.display_name":      textChange("+1 555 0100"),
			"party.party_kind":        textChange("person"),
			"party.organization_name": textChange("+1 555 0100"),
			"party.role_title":        textChange("+1 555 0100"),
		},
	}, "req-workbook_interaction-u-9-05-party-phone-like-again", time.Date(2026, 5, 18, 12, 3, 47, 0, time.UTC))
	if err != nil {
		t.Fatalf("create second phone-like party: %v", err)
	}
	if phoneLikeAgain.RecordID == phoneLike.RecordID {
		t.Fatalf("phone-like text must not drive implicit reuse")
	}
	requirePartyCount(t, harness, incident.ID, "display_name = '+1 555 0100'", 2)

	var partiesBeforeCrossKey int
	if err := harness.DB.QueryRow(context.Background(), `SELECT count(*) FROM parties WHERE incident_id = $1`, incident.ID).Scan(&partiesBeforeCrossKey); err != nil {
		t.Fatalf("count Parties before cross-key conflict: %v", err)
	}
	_, err = createPartyRow(context.Background(), partyOwner, actor, incident.ID, partyCreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-cross-key",
		Values: map[string]partyFieldValue{
			"party.display_name":  textChange("Cross-Key Candidate"),
			"party.party_kind":    textChange("organization"),
			"party.primary_email": textChange("vendor@example.test"),
			"party.external_ref":  textChange("EXT-42"),
		},
	}, "req-workbook_interaction-u-9-05-party-cross-key", time.Date(2026, 5, 18, 12, 3, 59, 0, time.UTC))
	var matchConflict *parties.PartyMatchConflictError
	if !errors.As(err, &matchConflict) || matchConflict.ReasonCode != parties.PartyMatchCrossKeyExactMatch {
		t.Fatalf("expected cross-key exact-match conflict, got %v", err)
	}
	if got := strings.Join(matchConflict.ConflictingFieldKeys, ","); got != "party.external_ref,party.primary_email" {
		t.Fatalf("unexpected cross-key fields: %q", got)
	}
	var partiesAfterCrossKey int
	if err := harness.DB.QueryRow(context.Background(), `SELECT count(*) FROM parties WHERE incident_id = $1`, incident.ID).Scan(&partiesAfterCrossKey); err != nil {
		t.Fatalf("count Parties after cross-key conflict: %v", err)
	}
	if partiesAfterCrossKey != partiesBeforeCrossKey {
		t.Fatalf("cross-key conflict created a Party: before=%d after=%d", partiesBeforeCrossKey, partiesAfterCrossKey)
	}

	otherIncidentParty, err := createPartyRow(context.Background(), partyOwner, actor, otherIncident.ID, partyCreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-other-party",
		Values: map[string]partyFieldValue{
			"party.display_name":  textChange("Other Incident Vendor"),
			"party.party_kind":    textChange("organization"),
			"party.primary_email": textChange("vendor@example.test"),
		},
	}, "req-workbook_interaction-u-9-05-other-party", time.Date(2026, 5, 18, 12, 4, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create other incident party: %v", err)
	}
	if otherIncidentParty.RecordID == createdByEmail.RecordID {
		t.Fatalf("party reuse crossed incidents: %#v", otherIncidentParty)
	}

	deletedEmailParty, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, partyCreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-deleted-email-party",
		Values: map[string]partyFieldValue{
			"party.display_name":  textChange("Deleted Email Party"),
			"party.party_kind":    textChange("organization"),
			"party.primary_email": textChange("deleted-email@example.test"),
		},
	}, "req-workbook_interaction-u-9-05-deleted-email-party", time.Date(2026, 5, 18, 12, 4, 10, 0, time.UTC))
	if err != nil {
		t.Fatalf("create deleted email party: %v", err)
	}
	softDeletePartyFor(t, harness, actor, deletedEmailParty.RecordID, "txn-workbook_interaction-u-9-05-delete-email-party")
	replacementEmailParty, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, partyCreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-replacement-email-party",
		Values: map[string]partyFieldValue{
			"party.display_name":  textChange("Replacement Email Party"),
			"party.party_kind":    textChange("organization"),
			"party.primary_email": textChange(" deleted-email@example.test "),
		},
	}, "req-workbook_interaction-u-9-05-replacement-email-party", time.Date(2026, 5, 18, 12, 4, 20, 0, time.UTC))
	if err != nil {
		t.Fatalf("create replacement email party: %v", err)
	}
	if replacementEmailParty.RecordID == deletedEmailParty.RecordID || replacementEmailParty.Outcome != parties.MutationCreated {
		t.Fatalf("deleted email match must not drive reuse, got deleted=%#v replacement=%#v", deletedEmailParty, replacementEmailParty)
	}
	requirePartyCount(t, harness, incident.ID, "lower(trim(primary_email)) = lower('deleted-email@example.test')", 1)
	restoreErr := restorePartyFor(t, harness, actor, deletedEmailParty.RecordID, "txn-workbook_interaction-u-9-05-restore-email-party")
	var restoreBlocked *revisions.RecordRestoreBlockedError
	if !errors.As(restoreErr, &restoreBlocked) || restoreBlocked.Block.ReasonCode != "exact_match_key_claimed" || strings.Join(restoreBlocked.Block.ConflictingFieldKeys, ",") != "party.primary_email" {
		t.Fatalf("expected Party restore claim conflict, got %v", restoreErr)
	}
	var deletedStill bool
	if err := harness.DB.QueryRow(context.Background(), `SELECT deleted_at IS NOT NULL FROM records WHERE record_id = $1`, deletedEmailParty.RecordID).Scan(&deletedStill); err != nil || !deletedStill {
		t.Fatalf("rejected Party restore changed envelope: deleted=%t err=%v", deletedStill, err)
	}

	deletedExternalRefParty, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, partyCreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-deleted-external-ref-party",
		Values: map[string]partyFieldValue{
			"party.display_name": textChange("Deleted External Ref Party"),
			"party.party_kind":   textChange("organization"),
			"party.external_ref": textChange("DELETED-EXT-42"),
		},
	}, "req-workbook_interaction-u-9-05-deleted-external-ref-party", time.Date(2026, 5, 18, 12, 4, 30, 0, time.UTC))
	if err != nil {
		t.Fatalf("create deleted external ref party: %v", err)
	}
	softDeletePartyFor(t, harness, actor, deletedExternalRefParty.RecordID, "txn-workbook_interaction-u-9-05-delete-external-ref-party")
	replacementExternalRefParty, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, partyCreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-replacement-external-ref-party",
		Values: map[string]partyFieldValue{
			"party.display_name": textChange("Replacement External Ref Party"),
			"party.party_kind":   textChange("organization"),
			"party.external_ref": textChange(" deleted-ext-42 "),
		},
	}, "req-workbook_interaction-u-9-05-replacement-external-ref-party", time.Date(2026, 5, 18, 12, 4, 40, 0, time.UTC))
	if err != nil {
		t.Fatalf("create replacement external ref party: %v", err)
	}
	if replacementExternalRefParty.RecordID == deletedExternalRefParty.RecordID || replacementExternalRefParty.Outcome != parties.MutationCreated {
		t.Fatalf("deleted external_ref match must not drive reuse, got deleted=%#v replacement=%#v", deletedExternalRefParty, replacementExternalRefParty)
	}
	requirePartyCount(t, harness, incident.ID, "lower(trim(external_ref)) = lower('deleted-ext-42')", 1)

	evidenceResult, err := createEvidenceRow(context.Background(), evidenceOwner, actor, incident.ID, evidence.CreateRequest{
		ViewSchemaID: evidence.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-evidence-text",
		Values: map[string]evidence.FieldValue{
			"evidence.title":                evidenceTextChange("Collector source text"),
			"evidence.collector_party_text": evidenceTextChange("IR Vendor <vendor@example.test>"),
		},
	}, "req-workbook_interaction-u-9-05-evidence-text", time.Date(2026, 5, 18, 12, 5, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create evidence with party text: %v", err)
	}
	requireCellValue(t, evidenceResult.Payload["row"].(map[string]any), "evidence.collector_party_text", "IR Vendor <vendor@example.test>")
	requireCellValue(t, evidenceResult.Payload["row"].(map[string]any), "evidence.collector_party_id", nil)
	requirePartyCount(t, harness, incident.ID, "display_name = 'IR Vendor <vendor@example.test>'", 0)

	phoneTextEvidence, err := createEvidenceRow(context.Background(), evidenceOwner, actor, incident.ID, evidence.CreateRequest{
		ViewSchemaID: evidence.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-evidence-phone-text",
		Values: map[string]evidence.FieldValue{
			"evidence.title":                evidenceTextChange("Collector phone-like source text"),
			"evidence.collector_party_text": evidenceTextChange("+1 555 0100"),
		},
	}, "req-workbook_interaction-u-9-05-evidence-phone-text", time.Date(2026, 5, 18, 12, 5, 30, 0, time.UTC))
	if err != nil {
		t.Fatalf("create evidence with phone-like party text: %v", err)
	}
	requireCellValue(t, phoneTextEvidence.Payload["row"].(map[string]any), "evidence.collector_party_text", "+1 555 0100")
	requireCellValue(t, phoneTextEvidence.Payload["row"].(map[string]any), "evidence.collector_party_id", nil)
}

func textChange(value string) partyFieldValue {
	return partyFieldValue{Text: &value}
}

func fieldValuePointer(value partyFieldValue) *partyFieldValue { return &value }

func evidenceTextChange(value string) evidence.FieldValue {
	return evidence.FieldValue{Text: &value}
}

func createEvidenceRow(
	ctx context.Context,
	owner evidence.MutationContribution,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	request evidence.CreateRequest,
	requestID string,
	now time.Time,
) (evidence.MutationResult, error) {
	return owner.Create(ctx, evidence.CreateCommand{
		Actor: actor, IncidentID: incidentID, Request: request,
		RequestHash: evidence.CreateRequestHash(request), RequestID: requestID,
		RouteKey: "workbook.rows.create", Now: now,
	})
}

func createPartyRow(
	ctx context.Context,
	owner *parties.MutationFacade,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	request partyCreateRequest,
	requestID string,
	now time.Time,
) (parties.MutationResult, error) {
	admission, err := admitPartyCreateRequest(request)
	if err != nil {
		return parties.MutationResult{}, err
	}
	return owner.Create(ctx, parties.CreateCommand{
		ActorUserID: actor.ID, IncidentID: incidentID, Admission: admission, RequestID: requestID,
		RouteKey: "workbook.rows.create", Now: now,
	})
}

func admitPartyCreateRequest(request partyCreateRequest) (parties.CreateAdmission, error) {
	payload := map[string]any{"client_txn_id": request.ClientTxnID}
	for fieldKey, value := range request.Values {
		payload[fieldKey] = value.Text
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return parties.CreateAdmission{}, fmt.Errorf("marshal Party create request: %w", err)
	}
	admission, apiErr := parties.AdmitCreateJSON(strings.NewReader(string(encoded)))
	if apiErr != nil {
		return parties.CreateAdmission{}, fmt.Errorf("admit Party create request: %s", apiErr.ReasonCode)
	}
	return admission, nil
}

func admitPartyPatchRequest(request partyPatchRequest) (parties.PatchAdmission, error) {
	changes := make([]map[string]any, 0, len(request.Changes))
	for _, change := range request.Changes {
		var value any
		if change.Value != nil {
			value = change.Value.Text
		}
		changes = append(changes, map[string]any{"field_key": change.FieldKey, "value": value})
	}
	encoded, err := json.Marshal(map[string]any{
		"view_schema_id":   request.ViewSchemaID,
		"base_row_version": request.BaseRowVersion,
		"client_txn_id":    request.ClientTxnID,
		"changes":          changes,
	})
	if err != nil {
		return parties.PatchAdmission{}, fmt.Errorf("marshal Party patch request: %w", err)
	}
	admission, apiErr := parties.AdmitPatchJSON(strings.NewReader(string(encoded)))
	if apiErr != nil {
		return parties.PatchAdmission{}, fmt.Errorf("admit Party patch request: %s", apiErr.ReasonCode)
	}
	return admission, nil
}

func requireCellValue(t testing.TB, row map[string]any, fieldKey string, want any) {
	t.Helper()
	cells := row["cells"].(map[string]any)
	got := cells[fieldKey].(map[string]any)["value"]
	if got != want {
		t.Fatalf("unexpected %s value: got %#v want %#v", fieldKey, got, want)
	}
}

func requirePartyCount(t testing.TB, harness *appsupport.StoreHarness, incidentID uuid.UUID, predicate string, want int) {
	t.Helper()
	var got int
	query := "SELECT count(*) FROM parties p JOIN records r ON r.incident_id = p.incident_id AND r.record_id = p.record_id WHERE p.incident_id = $1 AND r.deleted_at IS NULL AND " + predicate
	if err := harness.DB.QueryRow(context.Background(), query, incidentID).Scan(&got); err != nil {
		t.Fatalf("count parties: %v", err)
	}
	if got != want {
		t.Fatalf("unexpected party count for %s: got %d want %d", predicate, got, want)
	}
}

func softDeletePartyFor(t testing.TB, harness *appsupport.StoreHarness, actor authn.UserRecord, recordID uuid.UUID, clientTxnID string) {
	t.Helper()
	revisionComposition := revisionsupport.MustComposition(t)
	revisionRuntime := revisionComposition.Runtime
	projections, err := projectionassembly.Build(harness.DB)
	if err != nil {
		t.Fatalf("compose Projections: %v", err)
	}
	store, err := revisionRuntime.NewCommandService(
		harness.DB,
		partyTestAttributionResolver{},
		projections.RevisionRebuilder(),
		projections.RevisionLiveRecords(),
		func() time.Time { return time.Date(2026, 5, 18, 12, 4, 0, 0, time.UTC) },
	)
	if err != nil {
		t.Fatalf("compose revisions command service: %v", err)
	}
	request := revisions.DeleteRestoreRequest{
		BaseRowVersion: 1,
		ClientTxnID:    clientTxnID,
	}
	if _, err := store.SoftDeleteRecord(context.Background(), revisions.DeleteRestoreCommand{
		Actor:       revisions.NewActorID(actor.ID),
		RecordID:    recordID,
		Request:     request,
		RequestHash: revisions.DeleteRestoreRequestHash(request),
		RequestID:   "req-" + clientTxnID,
	}); err != nil {
		t.Fatalf("soft-delete party %s: %v", recordID, err)
	}
}

func restorePartyFor(t testing.TB, harness *appsupport.StoreHarness, actor authn.UserRecord, recordID uuid.UUID, clientTxnID string) error {
	t.Helper()
	revisionComposition := revisionsupport.MustComposition(t)
	projections, err := projectionassembly.Build(harness.DB)
	if err != nil {
		t.Fatalf("compose Projections: %v", err)
	}
	store, err := revisionComposition.Runtime.NewCommandService(
		harness.DB,
		partyTestAttributionResolver{},
		projections.RevisionRebuilder(),
		projections.RevisionLiveRecords(),
		func() time.Time { return time.Date(2026, 5, 18, 12, 4, 25, 0, time.UTC) },
	)
	if err != nil {
		t.Fatalf("compose revisions command service: %v", err)
	}
	request := revisions.DeleteRestoreRequest{BaseRowVersion: 2, ClientTxnID: clientTxnID}
	_, err = store.RestoreRecord(context.Background(), revisions.DeleteRestoreCommand{
		Actor: revisions.NewActorID(actor.ID), RecordID: recordID, Request: request,
		RequestHash: revisions.DeleteRestoreRequestHash(request), RequestID: "req-" + clientTxnID,
	})
	return err
}
