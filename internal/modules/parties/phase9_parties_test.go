package parties_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app"
	recordstoretest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type partyTestAttributionResolver struct{}

func (partyTestAttributionResolver) ResolveImportedSourceActorsTx(context.Context, pgx.Tx, uuid.UUID, string, string, []string) (map[string]string, error) {
	return map[string]string{}, nil
}

func TestPhase9_PartyExactMatchReuseAndRawTextPreservation_U_9_05(t *testing.T) {
	harness := recordstoretest.StartStore(t, "phase9-u-9-05-parties")
	store := workbook.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "u905@example.test", "U905 Parties", "U905PartiesPass1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-u-9-05-incident", "IR-U905", "Phase 9 U-9-05")
	otherIncident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-u-9-05-other-incident", "IR-U905B", "Phase 9 U-9-05 Other")

	createdByEmail, err := store.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.PartiesViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-05-party-email",
		Values: map[string]workbook.ValueChange{
			"party.display_name":  textChange("IR Vendor"),
			"party.party_kind":    textChange("organization"),
			"party.primary_email": textChange("Vendor@Example.Test"),
		},
	}, []byte("txn-phase9-u-9-05-party-email"), "req-phase9-u-9-05-party-email", time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create party by email: %v", err)
	}
	reusedByEmail, err := store.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.PartiesViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-05-party-email-reuse",
		Values: map[string]workbook.ValueChange{
			"party.display_name":  textChange("Phone-like label must not drive reuse"),
			"party.party_kind":    textChange("person"),
			"party.primary_email": textChange(" vendor@example.test "),
		},
	}, []byte("txn-phase9-u-9-05-party-email-reuse"), "req-phase9-u-9-05-party-email-reuse", time.Date(2026, 5, 18, 12, 1, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("reuse party by email: %v", err)
	}
	if reusedByEmail.RecordID != createdByEmail.RecordID || reusedByEmail.StatusCode != 200 {
		t.Fatalf("expected normalized same-incident email reuse, got created=%#v reused=%#v", createdByEmail, reusedByEmail)
	}
	requirePartyCount(t, harness, incident.ID, "lower(primary_email) = lower('vendor@example.test')", 1)

	createdByExternalRef, err := store.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.PartiesViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-05-party-external-ref",
		Values: map[string]workbook.ValueChange{
			"party.display_name": textChange("Outside Counsel"),
			"party.party_kind":   textChange("organization"),
			"party.external_ref": textChange("EXT-42"),
		},
	}, []byte("txn-phase9-u-9-05-party-external-ref"), "req-phase9-u-9-05-party-external-ref", time.Date(2026, 5, 18, 12, 2, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create party by external ref: %v", err)
	}
	reusedByExternalRef, err := store.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.PartiesViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-05-party-external-ref-reuse",
		Values: map[string]workbook.ValueChange{
			"party.display_name": textChange("Outside Counsel Alias"),
			"party.party_kind":   textChange("organization"),
			"party.external_ref": textChange(" ext-42 "),
		},
	}, []byte("txn-phase9-u-9-05-party-external-ref-reuse"), "req-phase9-u-9-05-party-external-ref-reuse", time.Date(2026, 5, 18, 12, 3, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("reuse party by external ref: %v", err)
	}
	if reusedByExternalRef.RecordID != createdByExternalRef.RecordID || reusedByExternalRef.StatusCode != 200 {
		t.Fatalf("expected normalized same-incident external_ref reuse, got created=%#v reused=%#v", createdByExternalRef, reusedByExternalRef)
	}
	requirePartyCount(t, harness, incident.ID, "lower(external_ref) = lower('ext-42')", 1)

	displayOnly, err := store.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.PartiesViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-05-party-display-only",
		Values: map[string]workbook.ValueChange{
			"party.display_name":      textChange("Duplicate Display"),
			"party.party_kind":        textChange("person"),
			"party.organization_name": textChange("Duplicate Org"),
			"party.role_title":        textChange("Duplicate Role"),
		},
	}, []byte("txn-phase9-u-9-05-party-display-only"), "req-phase9-u-9-05-party-display-only", time.Date(2026, 5, 18, 12, 3, 30, 0, time.UTC))
	if err != nil {
		t.Fatalf("create display-only party: %v", err)
	}
	displayOnlyAgain, err := store.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.PartiesViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-05-party-display-only-again",
		Values: map[string]workbook.ValueChange{
			"party.display_name":      textChange("Duplicate Display"),
			"party.party_kind":        textChange("person"),
			"party.organization_name": textChange("Duplicate Org"),
			"party.role_title":        textChange("Duplicate Role"),
		},
	}, []byte("txn-phase9-u-9-05-party-display-only-again"), "req-phase9-u-9-05-party-display-only-again", time.Date(2026, 5, 18, 12, 3, 45, 0, time.UTC))
	if err != nil {
		t.Fatalf("create second display-only party: %v", err)
	}
	if displayOnlyAgain.RecordID == displayOnly.RecordID {
		t.Fatalf("display name, organization, and role title must not drive implicit reuse")
	}
	requirePartyCount(t, harness, incident.ID, "display_name = 'Duplicate Display'", 2)

	phoneLike, err := store.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.PartiesViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-05-party-phone-like",
		Values: map[string]workbook.ValueChange{
			"party.display_name":      textChange("+1 555 0100"),
			"party.party_kind":        textChange("person"),
			"party.organization_name": textChange("+1 555 0100"),
			"party.role_title":        textChange("+1 555 0100"),
		},
	}, []byte("txn-phase9-u-9-05-party-phone-like"), "req-phase9-u-9-05-party-phone-like", time.Date(2026, 5, 18, 12, 3, 46, 0, time.UTC))
	if err != nil {
		t.Fatalf("create phone-like party: %v", err)
	}
	phoneLikeAgain, err := store.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.PartiesViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-05-party-phone-like-again",
		Values: map[string]workbook.ValueChange{
			"party.display_name":      textChange("+1 555 0100"),
			"party.party_kind":        textChange("person"),
			"party.organization_name": textChange("+1 555 0100"),
			"party.role_title":        textChange("+1 555 0100"),
		},
	}, []byte("txn-phase9-u-9-05-party-phone-like-again"), "req-phase9-u-9-05-party-phone-like-again", time.Date(2026, 5, 18, 12, 3, 47, 0, time.UTC))
	if err != nil {
		t.Fatalf("create second phone-like party: %v", err)
	}
	if phoneLikeAgain.RecordID == phoneLike.RecordID {
		t.Fatalf("phone-like text must not drive implicit reuse")
	}
	requirePartyCount(t, harness, incident.ID, "display_name = '+1 555 0100'", 2)

	ambiguousAnchor, err := store.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.PartiesViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-05-party-ambiguous-anchor",
		Values: map[string]workbook.ValueChange{
			"party.display_name":  textChange("Ambiguous Anchor"),
			"party.party_kind":    textChange("organization"),
			"party.primary_email": textChange("ambiguous@example.test"),
		},
	}, []byte("txn-phase9-u-9-05-party-ambiguous-anchor"), "req-phase9-u-9-05-party-ambiguous-anchor", time.Date(2026, 5, 18, 12, 3, 50, 0, time.UTC))
	if err != nil {
		t.Fatalf("create ambiguous anchor party: %v", err)
	}
	ambiguousDuplicate, err := store.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.PartiesViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-05-party-ambiguous-duplicate",
		Values: map[string]workbook.ValueChange{
			"party.display_name": textChange("Ambiguous Duplicate"),
			"party.party_kind":   textChange("organization"),
		},
	}, []byte("txn-phase9-u-9-05-party-ambiguous-duplicate"), "req-phase9-u-9-05-party-ambiguous-duplicate", time.Date(2026, 5, 18, 12, 3, 55, 0, time.UTC))
	if err != nil {
		t.Fatalf("create ambiguous duplicate party: %v", err)
	}
	if _, err := harness.DB.Exec(context.Background(), `UPDATE parties SET primary_email = 'ambiguous@example.test' WHERE record_id = $1`, ambiguousDuplicate.RecordID); err != nil {
		t.Fatalf("seed ambiguous party email: %v", err)
	}
	ambiguousCreate, err := store.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.PartiesViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-05-party-ambiguous-create",
		Values: map[string]workbook.ValueChange{
			"party.display_name":  textChange("Ambiguous Create"),
			"party.party_kind":    textChange("organization"),
			"party.primary_email": textChange(" ambiguous@example.test "),
		},
	}, []byte("txn-phase9-u-9-05-party-ambiguous-create"), "req-phase9-u-9-05-party-ambiguous-create", time.Date(2026, 5, 18, 12, 3, 59, 0, time.UTC))
	if err != nil {
		t.Fatalf("create party with ambiguous email: %v", err)
	}
	if ambiguousCreate.RecordID == ambiguousAnchor.RecordID || ambiguousCreate.RecordID == ambiguousDuplicate.RecordID {
		t.Fatalf("ambiguous same-incident email match must not select a party implicitly")
	}
	requirePartyCount(t, harness, incident.ID, "lower(trim(primary_email)) = lower('ambiguous@example.test')", 3)

	otherIncidentParty, err := store.CreateWorkbookRow(context.Background(), actor, otherIncident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.PartiesViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-05-other-party",
		Values: map[string]workbook.ValueChange{
			"party.display_name":  textChange("Other Incident Vendor"),
			"party.party_kind":    textChange("organization"),
			"party.primary_email": textChange("vendor@example.test"),
		},
	}, []byte("txn-phase9-u-9-05-other-party"), "req-phase9-u-9-05-other-party", time.Date(2026, 5, 18, 12, 4, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create other incident party: %v", err)
	}
	if otherIncidentParty.RecordID == createdByEmail.RecordID {
		t.Fatalf("party reuse crossed incidents: %#v", otherIncidentParty)
	}

	deletedEmailParty, err := store.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.PartiesViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-05-deleted-email-party",
		Values: map[string]workbook.ValueChange{
			"party.display_name":  textChange("Deleted Email Party"),
			"party.party_kind":    textChange("organization"),
			"party.primary_email": textChange("deleted-email@example.test"),
		},
	}, []byte("txn-phase9-u-9-05-deleted-email-party"), "req-phase9-u-9-05-deleted-email-party", time.Date(2026, 5, 18, 12, 4, 10, 0, time.UTC))
	if err != nil {
		t.Fatalf("create deleted email party: %v", err)
	}
	softDeletePartyForU905(t, harness, actor, deletedEmailParty.RecordID, "txn-phase9-u-9-05-delete-email-party")
	replacementEmailParty, err := store.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.PartiesViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-05-replacement-email-party",
		Values: map[string]workbook.ValueChange{
			"party.display_name":  textChange("Replacement Email Party"),
			"party.party_kind":    textChange("organization"),
			"party.primary_email": textChange(" deleted-email@example.test "),
		},
	}, []byte("txn-phase9-u-9-05-replacement-email-party"), "req-phase9-u-9-05-replacement-email-party", time.Date(2026, 5, 18, 12, 4, 20, 0, time.UTC))
	if err != nil {
		t.Fatalf("create replacement email party: %v", err)
	}
	if replacementEmailParty.RecordID == deletedEmailParty.RecordID || replacementEmailParty.StatusCode != 201 {
		t.Fatalf("deleted email match must not drive reuse, got deleted=%#v replacement=%#v", deletedEmailParty, replacementEmailParty)
	}
	requirePartyCount(t, harness, incident.ID, "lower(trim(primary_email)) = lower('deleted-email@example.test')", 1)

	deletedExternalRefParty, err := store.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.PartiesViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-05-deleted-external-ref-party",
		Values: map[string]workbook.ValueChange{
			"party.display_name": textChange("Deleted External Ref Party"),
			"party.party_kind":   textChange("organization"),
			"party.external_ref": textChange("DELETED-EXT-42"),
		},
	}, []byte("txn-phase9-u-9-05-deleted-external-ref-party"), "req-phase9-u-9-05-deleted-external-ref-party", time.Date(2026, 5, 18, 12, 4, 30, 0, time.UTC))
	if err != nil {
		t.Fatalf("create deleted external ref party: %v", err)
	}
	softDeletePartyForU905(t, harness, actor, deletedExternalRefParty.RecordID, "txn-phase9-u-9-05-delete-external-ref-party")
	replacementExternalRefParty, err := store.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.PartiesViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-05-replacement-external-ref-party",
		Values: map[string]workbook.ValueChange{
			"party.display_name": textChange("Replacement External Ref Party"),
			"party.party_kind":   textChange("organization"),
			"party.external_ref": textChange(" deleted-ext-42 "),
		},
	}, []byte("txn-phase9-u-9-05-replacement-external-ref-party"), "req-phase9-u-9-05-replacement-external-ref-party", time.Date(2026, 5, 18, 12, 4, 40, 0, time.UTC))
	if err != nil {
		t.Fatalf("create replacement external ref party: %v", err)
	}
	if replacementExternalRefParty.RecordID == deletedExternalRefParty.RecordID || replacementExternalRefParty.StatusCode != 201 {
		t.Fatalf("deleted external_ref match must not drive reuse, got deleted=%#v replacement=%#v", deletedExternalRefParty, replacementExternalRefParty)
	}
	requirePartyCount(t, harness, incident.ID, "lower(trim(external_ref)) = lower('deleted-ext-42')", 1)

	evidence, err := store.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.EvidenceViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-05-evidence-text",
		Values: map[string]workbook.ValueChange{
			"evidence.title":                textChange("Collector source text"),
			"evidence.collector_party_text": textChange("IR Vendor <vendor@example.test>"),
		},
	}, []byte("txn-phase9-u-9-05-evidence-text"), "req-phase9-u-9-05-evidence-text", time.Date(2026, 5, 18, 12, 5, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create evidence with party text: %v", err)
	}
	requireCellValue(t, evidence.Payload["row"].(map[string]any), "evidence.collector_party_text", "IR Vendor <vendor@example.test>")
	requireCellValue(t, evidence.Payload["row"].(map[string]any), "evidence.collector_party_id", nil)
	requirePartyCount(t, harness, incident.ID, "display_name = 'IR Vendor <vendor@example.test>'", 0)

	phoneTextEvidence, err := store.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.EvidenceViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-05-evidence-phone-text",
		Values: map[string]workbook.ValueChange{
			"evidence.title":                textChange("Collector phone-like source text"),
			"evidence.collector_party_text": textChange("+1 555 0100"),
		},
	}, []byte("txn-phase9-u-9-05-evidence-phone-text"), "req-phase9-u-9-05-evidence-phone-text", time.Date(2026, 5, 18, 12, 5, 30, 0, time.UTC))
	if err != nil {
		t.Fatalf("create evidence with phone-like party text: %v", err)
	}
	requireCellValue(t, phoneTextEvidence.Payload["row"].(map[string]any), "evidence.collector_party_text", "+1 555 0100")
	requireCellValue(t, phoneTextEvidence.Payload["row"].(map[string]any), "evidence.collector_party_id", nil)
}

func textChange(value string) workbook.ValueChange {
	return workbook.ValueChange{Kind: "text", Text: &value}
}

func requireCellValue(t testing.TB, row map[string]any, fieldKey string, want any) {
	t.Helper()
	cells := row["cells"].(map[string]any)
	got := cells[fieldKey].(map[string]any)["value"]
	if got != want {
		t.Fatalf("unexpected %s value: got %#v want %#v", fieldKey, got, want)
	}
}

func requirePartyCount(t testing.TB, harness *recordstoretest.StoreHarness, incidentID uuid.UUID, predicate string, want int) {
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

func softDeletePartyForU905(t testing.TB, harness *recordstoretest.StoreHarness, actor authn.UserRecord, recordID uuid.UUID, clientTxnID string) {
	t.Helper()
	store, err := app.NewRevisionsCommandService(harness.DB, partyTestAttributionResolver{})
	if err != nil {
		t.Fatalf("compose revisions command service: %v", err)
	}
	request := revisions.DeleteRestoreRequest{
		BaseRowVersion: 1,
		ClientTxnID:    clientTxnID,
	}
	if _, err := store.SoftDeleteRecord(context.Background(), actor, recordID, request, revisions.DeleteRestoreRequestHash(request), "req-"+clientTxnID, time.Date(2026, 5, 18, 12, 4, 0, 0, time.UTC)); err != nil {
		t.Fatalf("soft-delete party %s: %v", recordID, err)
	}
}
