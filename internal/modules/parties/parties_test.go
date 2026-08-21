package parties_test

import (
	"context"
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

	createdByEmail, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, parties.CreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-email",
		Values: map[string]parties.FieldValue{
			"party.display_name":  textChange("IR Vendor"),
			"party.party_kind":    textChange("organization"),
			"party.primary_email": textChange("Vendor@Example.Test"),
		},
	}, "req-workbook_interaction-u-9-05-party-email", time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create party by email: %v", err)
	}
	reusedByEmail, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, parties.CreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-email-reuse",
		Values: map[string]parties.FieldValue{
			"party.display_name":  textChange("Phone-like label must not drive reuse"),
			"party.party_kind":    textChange("person"),
			"party.primary_email": textChange(" vendor@example.test "),
		},
	}, "req-workbook_interaction-u-9-05-party-email-reuse", time.Date(2026, 5, 18, 12, 1, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("reuse party by email: %v", err)
	}
	if reusedByEmail.RecordID != createdByEmail.RecordID || reusedByEmail.StatusCode != 200 {
		t.Fatalf("expected normalized same-incident email reuse, got created=%#v reused=%#v", createdByEmail, reusedByEmail)
	}
	requirePartyCount(t, harness, incident.ID, "lower(primary_email) = lower('vendor@example.test')", 1)

	createdByExternalRef, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, parties.CreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-external-ref",
		Values: map[string]parties.FieldValue{
			"party.display_name": textChange("Outside Counsel"),
			"party.party_kind":   textChange("organization"),
			"party.external_ref": textChange("EXT-42"),
		},
	}, "req-workbook_interaction-u-9-05-party-external-ref", time.Date(2026, 5, 18, 12, 2, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create party by external ref: %v", err)
	}
	reusedByExternalRef, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, parties.CreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-external-ref-reuse",
		Values: map[string]parties.FieldValue{
			"party.display_name": textChange("Outside Counsel Alias"),
			"party.party_kind":   textChange("organization"),
			"party.external_ref": textChange(" ext-42 "),
		},
	}, "req-workbook_interaction-u-9-05-party-external-ref-reuse", time.Date(2026, 5, 18, 12, 3, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("reuse party by external ref: %v", err)
	}
	if reusedByExternalRef.RecordID != createdByExternalRef.RecordID || reusedByExternalRef.StatusCode != 200 {
		t.Fatalf("expected normalized same-incident external_ref reuse, got created=%#v reused=%#v", createdByExternalRef, reusedByExternalRef)
	}
	requirePartyCount(t, harness, incident.ID, "lower(external_ref) = lower('ext-42')", 1)

	displayOnly, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, parties.CreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-display-only",
		Values: map[string]parties.FieldValue{
			"party.display_name":      textChange("Duplicate Display"),
			"party.party_kind":        textChange("person"),
			"party.organization_name": textChange("Duplicate Org"),
			"party.role_title":        textChange("Duplicate Role"),
		},
	}, "req-workbook_interaction-u-9-05-party-display-only", time.Date(2026, 5, 18, 12, 3, 30, 0, time.UTC))
	if err != nil {
		t.Fatalf("create display-only party: %v", err)
	}
	displayOnlyAgain, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, parties.CreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-display-only-again",
		Values: map[string]parties.FieldValue{
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

	phoneLike, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, parties.CreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-phone-like",
		Values: map[string]parties.FieldValue{
			"party.display_name":      textChange("+1 555 0100"),
			"party.party_kind":        textChange("person"),
			"party.organization_name": textChange("+1 555 0100"),
			"party.role_title":        textChange("+1 555 0100"),
		},
	}, "req-workbook_interaction-u-9-05-party-phone-like", time.Date(2026, 5, 18, 12, 3, 46, 0, time.UTC))
	if err != nil {
		t.Fatalf("create phone-like party: %v", err)
	}
	phoneLikeAgain, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, parties.CreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-phone-like-again",
		Values: map[string]parties.FieldValue{
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

	ambiguousAnchor, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, parties.CreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-ambiguous-anchor",
		Values: map[string]parties.FieldValue{
			"party.display_name":  textChange("Ambiguous Anchor"),
			"party.party_kind":    textChange("organization"),
			"party.primary_email": textChange("ambiguous@example.test"),
		},
	}, "req-workbook_interaction-u-9-05-party-ambiguous-anchor", time.Date(2026, 5, 18, 12, 3, 50, 0, time.UTC))
	if err != nil {
		t.Fatalf("create ambiguous anchor party: %v", err)
	}
	ambiguousDuplicate, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, parties.CreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-ambiguous-duplicate",
		Values: map[string]parties.FieldValue{
			"party.display_name": textChange("Ambiguous Duplicate"),
			"party.party_kind":   textChange("organization"),
		},
	}, "req-workbook_interaction-u-9-05-party-ambiguous-duplicate", time.Date(2026, 5, 18, 12, 3, 55, 0, time.UTC))
	if err != nil {
		t.Fatalf("create ambiguous duplicate party: %v", err)
	}
	if _, err := harness.DB.Exec(context.Background(), `UPDATE parties SET primary_email = 'ambiguous@example.test' WHERE record_id = $1`, ambiguousDuplicate.RecordID); err != nil {
		t.Fatalf("seed ambiguous party email: %v", err)
	}
	ambiguousCreate, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, parties.CreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-party-ambiguous-create",
		Values: map[string]parties.FieldValue{
			"party.display_name":  textChange("Ambiguous Create"),
			"party.party_kind":    textChange("organization"),
			"party.primary_email": textChange(" ambiguous@example.test "),
		},
	}, "req-workbook_interaction-u-9-05-party-ambiguous-create", time.Date(2026, 5, 18, 12, 3, 59, 0, time.UTC))
	if err != nil {
		t.Fatalf("create party with ambiguous email: %v", err)
	}
	if ambiguousCreate.RecordID == ambiguousAnchor.RecordID || ambiguousCreate.RecordID == ambiguousDuplicate.RecordID {
		t.Fatalf("ambiguous same-incident email match must not select a party implicitly")
	}
	requirePartyCount(t, harness, incident.ID, "lower(trim(primary_email)) = lower('ambiguous@example.test')", 3)

	otherIncidentParty, err := createPartyRow(context.Background(), partyOwner, actor, otherIncident.ID, parties.CreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-other-party",
		Values: map[string]parties.FieldValue{
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

	deletedEmailParty, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, parties.CreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-deleted-email-party",
		Values: map[string]parties.FieldValue{
			"party.display_name":  textChange("Deleted Email Party"),
			"party.party_kind":    textChange("organization"),
			"party.primary_email": textChange("deleted-email@example.test"),
		},
	}, "req-workbook_interaction-u-9-05-deleted-email-party", time.Date(2026, 5, 18, 12, 4, 10, 0, time.UTC))
	if err != nil {
		t.Fatalf("create deleted email party: %v", err)
	}
	softDeletePartyFor(t, harness, actor, deletedEmailParty.RecordID, "txn-workbook_interaction-u-9-05-delete-email-party")
	replacementEmailParty, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, parties.CreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-replacement-email-party",
		Values: map[string]parties.FieldValue{
			"party.display_name":  textChange("Replacement Email Party"),
			"party.party_kind":    textChange("organization"),
			"party.primary_email": textChange(" deleted-email@example.test "),
		},
	}, "req-workbook_interaction-u-9-05-replacement-email-party", time.Date(2026, 5, 18, 12, 4, 20, 0, time.UTC))
	if err != nil {
		t.Fatalf("create replacement email party: %v", err)
	}
	if replacementEmailParty.RecordID == deletedEmailParty.RecordID || replacementEmailParty.StatusCode != 201 {
		t.Fatalf("deleted email match must not drive reuse, got deleted=%#v replacement=%#v", deletedEmailParty, replacementEmailParty)
	}
	requirePartyCount(t, harness, incident.ID, "lower(trim(primary_email)) = lower('deleted-email@example.test')", 1)

	deletedExternalRefParty, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, parties.CreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-deleted-external-ref-party",
		Values: map[string]parties.FieldValue{
			"party.display_name": textChange("Deleted External Ref Party"),
			"party.party_kind":   textChange("organization"),
			"party.external_ref": textChange("DELETED-EXT-42"),
		},
	}, "req-workbook_interaction-u-9-05-deleted-external-ref-party", time.Date(2026, 5, 18, 12, 4, 30, 0, time.UTC))
	if err != nil {
		t.Fatalf("create deleted external ref party: %v", err)
	}
	softDeletePartyFor(t, harness, actor, deletedExternalRefParty.RecordID, "txn-workbook_interaction-u-9-05-delete-external-ref-party")
	replacementExternalRefParty, err := createPartyRow(context.Background(), partyOwner, actor, incident.ID, parties.CreateRequest{
		ViewSchemaID: parties.ViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-05-replacement-external-ref-party",
		Values: map[string]parties.FieldValue{
			"party.display_name": textChange("Replacement External Ref Party"),
			"party.party_kind":   textChange("organization"),
			"party.external_ref": textChange(" deleted-ext-42 "),
		},
	}, "req-workbook_interaction-u-9-05-replacement-external-ref-party", time.Date(2026, 5, 18, 12, 4, 40, 0, time.UTC))
	if err != nil {
		t.Fatalf("create replacement external ref party: %v", err)
	}
	if replacementExternalRefParty.RecordID == deletedExternalRefParty.RecordID || replacementExternalRefParty.StatusCode != 201 {
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

func textChange(value string) parties.FieldValue {
	return parties.FieldValue{Text: &value}
}

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
	request parties.CreateRequest,
	requestID string,
	now time.Time,
) (parties.MutationResult, error) {
	return owner.Create(ctx, parties.CreateCommand{
		Actor: actor, IncidentID: incidentID, Request: request,
		RequestHash: parties.CreateRequestHash(request), RequestID: requestID,
		RouteKey: "workbook.rows.create", Now: now,
	})
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
