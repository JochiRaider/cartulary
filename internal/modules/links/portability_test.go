package links_test

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	linktest "github.com/JochiRaider/cartulary/internal/modules/links/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestLinksIncidentBundleRejectsUnknownLinkMembersBeforeMutation(t *testing.T) {
	descriptor := links.NewIncidentBundleSourcePort().Descriptor()
	if !slices.Contains(descriptor.InvariantIDs, "links_tags.link_tuple_legal") {
		t.Fatalf("Links source descriptor omits link-tuple invariant: %#v", descriptor.InvariantIDs)
	}
	harness := appsupport.StartStore(t, "links-bundle-unknown-member-characterization")
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "links-bundle-shape@example.test", "Links Bundle Shape", "LinksBundleShapePass1!", false, true, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-links-bundle-shape-incident", "IR-LINK-BUNDLE", "Links bundle shape")
	src := uuid.New()
	dst := uuid.New()
	for _, recordID := range []uuid.UUID{src, dst} {
		timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, recordID)
	}
	now := time.Date(2026, time.August, 21, 18, 42, 0, 0, time.UTC)
	tx, err := harness.DB.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin unknown-member import: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	for _, member := range []string{"note", "description", "comment", "future_undeclared_member"} {
		t.Run(member, func(t *testing.T) {
			linkID := uuid.New()
			row := map[string]any{
				"record_link_id": linkID.String(), "incident_id": incident.ID.String(),
				"src_record_id": src.String(), "dst_record_id": dst.String(),
				"link_type": "references_record", "field_key": "fixture.links.bundle",
				"provenance": "manual", "confidence": nil,
				"owner_user_id": actor.ID.String(), "created_by_user_id": actor.ID.String(),
				"decided_at": now.Format(time.RFC3339Nano), "created_at": now.Format(time.RFC3339Nano),
				"deleted_at": nil, "deleted_by_user_id": nil,
				member: "unsupported value",
			}
			payload, err := json.Marshal(row)
			if err != nil {
				t.Fatalf("encode unknown-member link row: %v", err)
			}
			attributions := &countingLinkAttributions{}
			err = linktest.ImportIncidentBundleFilesTx(context.Background(), tx, map[string][]byte{
				"data/record_links.ndjson": append(payload, '\n'),
				"data/record_tags.ndjson":  {},
			}, actor.ID, attributions)
			var unexpectedColumns *incidentportability.UnexpectedColumnsError
			if !errors.As(err, &unexpectedColumns) {
				t.Fatalf("unknown member error = %T %[1]v, want UnexpectedColumnsError", err)
			}
			if attributions.calls != 0 {
				t.Fatalf("unknown member recorded %d attributions before rejection", attributions.calls)
			}
			var count int
			if err := tx.QueryRow(context.Background(), `SELECT count(*) FROM record_links WHERE record_link_id = $1`, linkID).Scan(&count); err != nil {
				t.Fatalf("query rejected link row: %v", err)
			}
			if count != 0 {
				t.Fatalf("unknown-member import inserted %d rows, want 0", count)
			}
		})
	}
	facts, err := (links.FactReader{}).LoadIncidentTx(context.Background(), tx, incident.ID)
	if err != nil {
		t.Fatalf("load empty active fact set: %v", err)
	}
	if facts.RecordLinks == nil || facts.RecordTags == nil || len(facts.RecordLinks) != 0 || len(facts.RecordTags) != 0 {
		t.Fatalf("empty active facts lost non-nil collection posture: %#v", facts)
	}
	recordFacts, err := (links.FactReader{}).LoadRecordTx(context.Background(), tx, incident.ID, src)
	if err != nil {
		t.Fatalf("load empty record fact set: %v", err)
	}
	if recordFacts.RecordLinks == nil || recordFacts.RecordTags == nil || len(recordFacts.RecordLinks) != 0 || len(recordFacts.RecordTags) != 0 {
		t.Fatalf("empty record facts lost non-nil collection posture: %#v", recordFacts)
	}
	if err := tx.Rollback(context.Background()); err != nil {
		t.Fatalf("close active fact reader transaction: %v", err)
	}
	reader := links.FactReader{}
	closedIncidentFacts, err := reader.LoadIncidentTx(context.Background(), tx, incident.ID)
	var readFailure *links.FactReadError
	if !errors.As(err, &readFailure) || closedIncidentFacts.RecordLinks != nil || closedIncidentFacts.RecordTags != nil {
		t.Fatalf("incident fact failure = (%#v, %T %[2]v), want FactReadError without partial facts", closedIncidentFacts, err)
	}
	closedRecordFacts, err := reader.LoadRecordTx(context.Background(), tx, incident.ID, src)
	readFailure = nil
	if !errors.As(err, &readFailure) || closedRecordFacts.RecordLinks != nil || closedRecordFacts.RecordTags != nil {
		t.Fatalf("record fact failure = (%#v, %T %[2]v), want FactReadError without partial facts", closedRecordFacts, err)
	}
	closedChanges, err := reader.LoadCollectionChangesTx(context.Background(), tx, incident.ID, src, now)
	readFailure = nil
	if !errors.As(err, &readFailure) || closedChanges.LinkFieldKeys != nil || closedChanges.TagsChanged {
		t.Fatalf("collection-change failure = (%#v, %T %[2]v), want FactReadError without partial facts", closedChanges, err)
	}
}

type countingLinkAttributions struct {
	calls int
}

func (recorder *countingLinkAttributions) RecordImportedAttribution(string, string, string, string) error {
	recorder.calls++
	return nil
}

func (*countingLinkAttributions) ImportedAttributions() []incidentportability.ImportedAttribution {
	return []incidentportability.ImportedAttribution{}
}
