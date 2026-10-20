package links_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/links/partyassignment"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestPartyAssignmentSnapshotSource(t *testing.T) {
	harness := appsupport.StartStore(t, "party-assignment-source")
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "assignment-source@example.test", DisplayName: "Assignment source", DeploymentAdmin: true})
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "assignment-source-incident", "IR-ASSIGNMENT", "Assignment snapshot isolation")
	testPartyAssignmentFacts(t, harness.DB, incident.ID, actor.ID)
}

func testPartyAssignmentFacts(t *testing.T, db postgres.DB, incident, actor uuid.UUID) {
	t.Run("state version contract", testPartyAssignmentStateContract)
	ctx := context.Background()
	now := time.Date(2026, 10, 15, 12, 0, 0, 123456000, time.UTC)
	owner := appsupport.NewPartyOwner(db, conflicttest.NewCodec("assignment-source"))
	admission, apiErr := parties.AdmitCreateJSON(strings.NewReader(`{"client_txn_id":"assignment-party","party.display_name":"Assignment party","party.party_kind":"organization"}`))
	if apiErr != nil {
		t.Fatal(apiErr)
	}
	party, err := owner.Create(ctx, parties.CreateCommand{ActorUserID: actor, IncidentID: incident, Admission: admission, RequestID: "assignment-party", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	host, identity := uuid.New(), uuid.New()
	entitytest.SeedHostRecord(t, db, incident, actor, host, "Assignment host", "assignment-host", "", "")
	entitytest.SeedIdentityRecord(t, db, incident, actor, identity, "Assignment identity", "assignment@example.test", "", "")
	store := links.NewStore()
	for name, apply := range map[string]func(pgx.Tx) (links.CollectionMutationResult, error){
		"wrong family": func(tx pgx.Tx) (links.CollectionMutationResult, error) {
			return store.ApplyPartyRefCollectionWithMutationValuesTx(ctx, tx, links.PartyRefCollectionCommand{IncidentID: incident, SourceRecordID: identity, ActorUserID: actor, FieldKey: "host.subject_party_ids", LinkType: links.LinkTypeReferencesRecord, ExpectedTargetType: "party", AddPartyIDs: []uuid.UUID{party.RecordID}, Now: now})
		},
		"generic target bypass": func(tx pgx.Tx) (links.CollectionMutationResult, error) {
			return store.ApplyRecordRefCollectionWithMutationValuesTx(ctx, tx, links.RecordRefCollectionCommand{IncidentID: incident, SourceRecordID: host, ActorUserID: actor, FieldKey: "host.subject_party_ids", LinkType: links.LinkTypeReferencesRecord, AddRecordIDs: []uuid.UUID{identity}, Now: now})
		},
		"wrong link kind": func(tx pgx.Tx) (links.CollectionMutationResult, error) {
			return store.ApplyPartyRefCollectionWithMutationValuesTx(ctx, tx, links.PartyRefCollectionCommand{IncidentID: incident, SourceRecordID: host, ActorUserID: actor, FieldKey: "host.subject_party_ids", LinkType: links.LinkTypeSupportedBy, ExpectedTargetType: "party", AddPartyIDs: []uuid.UUID{party.RecordID}, Now: now})
		},
		"scalar replacement": func(tx pgx.Tx) (links.CollectionMutationResult, error) {
			return store.SyncFieldReferenceWithMutationValuesTx(ctx, tx, links.SyncFieldReferenceCommand{IncidentID: incident, SrcRecordID: host, ActorUserID: actor, FieldKey: "host.subject_party_ids", LinkType: links.LinkTypeReferencesRecord, TargetID: &party.RecordID, Now: now})
		},
	} {
		t.Run(name, func(t *testing.T) {
			tx, err := db.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			result, err := apply(tx)
			var invalid *links.CollectionValidationError
			if !errors.As(err, &invalid) || len(result.Mutations()) != 0 {
				t.Fatal("invalid assignment mutated source", err)
			}
			var count int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM record_links WHERE incident_id=$1`, incident).Scan(&count); err != nil || count != 0 {
				t.Fatal("rejected assignment left partial state", count, err)
			}
		})
	}
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	for family, id := range map[string]uuid.UUID{"host": host, "identity": identity} {
		for _, suffix := range []string{"subject_party_ids", "responsible_party_ids", "observer_party_ids"} {
			field := family + "." + suffix
			if !links.IsPartyAssignmentField(field) {
				t.Fatal("owned assignment field was not recognized", field)
			}
			_, err := store.ApplyPartyRefCollectionWithMutationValuesTx(ctx, tx, links.PartyRefCollectionCommand{
				IncidentID: incident, SourceRecordID: id, ActorUserID: actor, FieldKey: field,
				LinkType: links.LinkTypeReferencesRecord, ExpectedTargetType: "party", AddPartyIDs: []uuid.UUID{party.RecordID}, Now: now,
			})
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"comm_log.audience_party_ids", "host.party_ids", "host.subject_party_ids ", "party.subject_party_ids", "host.subject_party_ids.extra", ""} {
		if links.IsPartyAssignmentField(field) {
			t.Fatal("unowned collection became an assignment", field)
		}
	}
	reader := links.FactReader{}
	frozen, err := db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = frozen.Rollback(ctx) }()
	var isolation string
	if err := frozen.QueryRow(ctx, "SHOW transaction_isolation").Scan(&isolation); err != nil || isolation != "repeatable read" {
		t.Fatal("assignment snapshot requires a real repeatable-read transaction", isolation, err)
	}
	facts, err := reader.LoadPartyAssignmentsTx(ctx, frozen, incident)
	if err != nil || len(facts) != 6 {
		t.Fatal("assignment inventory did not close", len(facts), err)
	}
	for i, fact := range facts {
		if fact.Validate() != nil || fact.CreatedAt != "2026-10-15T12:00:00.123456Z" || fact.CreatedByUserID != actor || fact.PartyID != party.RecordID || fact.Provenance != "manual" || fact.SourceRef != nil || (i > 0 && facts[i-1].AssignmentID.String() >= fact.AssignmentID.String()) {
			t.Fatal("source provenance or deterministic ordering was lost")
		}
	}
	readCurrent := func() ([]partyassignment.Fact, error) {
		current, err := db.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
		if err != nil {
			return nil, err
		}
		defer func() { _ = current.Rollback(ctx) }()
		return reader.LoadPartyAssignmentsTx(ctx, current, incident)
	}
	first := facts[0]
	if _, err := db.Exec(ctx, `UPDATE record_links SET provenance='import' WHERE record_link_id=$1`, first.AssignmentID); err != nil {
		t.Fatal(err)
	}
	unchanged, err := reader.LoadPartyAssignmentsTx(ctx, frozen, incident)
	if err != nil || len(unchanged) != 6 || unchanged[0].RowVersion != first.RowVersion || unchanged[0].Provenance != "manual" {
		t.Fatal("source projection escaped the caller's snapshot", err)
	}
	changed, err := readCurrent()
	if err != nil || len(changed) != 6 || changed[0].RowVersion == first.RowVersion || changed[0].Provenance != "import" {
		t.Fatal("changed assignment state retained its old version", err)
	}
	for _, test := range []struct {
		sql   string
		count int
	}{
		{`UPDATE record_links SET provenance='auto_match' WHERE record_link_id=$1`, -1},
		{`UPDATE record_links SET provenance='manual',link_type='supported_by' WHERE record_link_id=$1`, -1},
		{`UPDATE record_links SET link_type='references_record',field_key='party.subject_party_ids' WHERE record_link_id=$1`, 5},
	} {
		if _, err := db.Exec(ctx, test.sql, first.AssignmentID); err != nil {
			t.Fatal(err)
		}
		got, err := readCurrent()
		if test.count >= 0 {
			// An unowned key is a generic link, never an inferred assignment.
			if err != nil || len(got) != test.count {
				t.Fatal("generic link was inferred as assignment", err)
			}
		} else if err == nil || got != nil {
			t.Fatal("malformed assignment produced a partial projection")
		}
	}
	if _, err := db.Exec(ctx, `UPDATE records SET deleted_at=$2,deleted_by_user_id=$3 WHERE record_id=$1`, party.RecordID, now.Add(time.Hour), actor); err != nil {
		t.Fatal(err)
	}
	got, err := readCurrent()
	if err != nil || len(got) != 0 {
		t.Fatal("deleted Party contributed an active assignment", err)
	}
}

func testPartyAssignmentStateContract(t *testing.T) {
	state := partyassignment.State{
		AssignmentID:      uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		IncidentID:        uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		SubjectRecordID:   uuid.MustParse("00000000-0000-0000-0000-000000000003"),
		SubjectRecordType: "host", PartyID: uuid.MustParse("00000000-0000-0000-0000-000000000004"),
		AssignmentKind: "subject_party", AssignmentState: "active", Provenance: "manual",
		CreatedByUserID: uuid.MustParse("00000000-0000-0000-0000-000000000005"), CreatedAt: "2026-10-15T12:00:00.123456Z",
	}
	fact, err := partyassignment.NewFact(state)
	// Independent fixed byte/domain vector, not a digest calculated through
	// the implementation under test.
	// Keep the prefix separate so secret scanners do not mistake this digest
	// for an API token.
	const want = "rpa_" + "4d10d665a93a16143f54a5524cf0df3d9e1497077af365c3b76de125be0ebb7e"
	if err != nil || fact.RowVersion != want || fact.Validate() != nil {
		t.Fatal("invalid assignment version vector", err)
	}
	fact.Provenance = "import"
	if fact.Validate() == nil {
		t.Fatal("modified source retained an admitted version")
	}
	for name, change := range map[string]func(*partyassignment.State){
		"missing identity":    func(s *partyassignment.State) { s.AssignmentID = uuid.Nil },
		"self assignment":     func(s *partyassignment.State) { s.PartyID = s.SubjectRecordID },
		"wrong source type":   func(s *partyassignment.State) { s.SubjectRecordType = "timeline_event" },
		"unknown kind":        func(s *partyassignment.State) { s.AssignmentKind = "owner_party" },
		"inactive":            func(s *partyassignment.State) { s.AssignmentState = "deleted" },
		"inferred provenance": func(s *partyassignment.State) { s.Provenance = "auto_match" },
		"invented source":     func(s *partyassignment.State) { s.SourceRef = &struct{}{} },
		"superseded":          func(s *partyassignment.State) { s.SupersededByAssignmentID = &s.AssignmentID },
		"deleted":             func(s *partyassignment.State) { s.DeletedAt = &s.CreatedAt },
		"imprecise time":      func(s *partyassignment.State) { s.CreatedAt = "2026-10-15T12:00:00Z" },
		"nanosecond time":     func(s *partyassignment.State) { s.CreatedAt = "2026-10-15T12:00:00.123456789Z" },
		"invalid date":        func(s *partyassignment.State) { s.CreatedAt = "2026-02-30T12:00:00.123456Z" },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := state
			change(&candidate)
			if got, err := partyassignment.NewFact(candidate); err == nil || got.RowVersion != "" {
				t.Fatal("invalid source state received a version")
			}
		})
	}
}
