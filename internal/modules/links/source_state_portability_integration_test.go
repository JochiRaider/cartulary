package links_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	linktest "github.com/JochiRaider/cartulary/internal/modules/links/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestLinksSourceStateStrictRoundTripAndAtomicFailure_Integration(t *testing.T) {
	ctx := context.Background()
	port, err := links.NewIncidentBundleSourcePort()
	if err != nil {
		t.Fatalf("construct Links source port: %v", err)
	}
	harness := appsupport.StartStore(t, "links-source-state-round-trip")
	actor := authstoretest.SeedLocalUserRecord(
		t, harness.DB, "links-source-state@example.test", "Links Source State", "LinksSourceStatePass1!", false, true, true,
	)
	incident := appsupport.CreateIncidentInStore(
		t, harness.DB, actor, "txn-links-source-state-incident", "IR-LINK-SOURCE-STATE", "Links source-state round trip",
	)
	sourceID := uuid.MustParse("42000000-0000-0000-0000-000000000001")
	targetID := uuid.MustParse("42000000-0000-0000-0000-000000000002")
	for _, recordID := range []uuid.UUID{sourceID, targetID} {
		timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, recordID)
	}
	linkID := uuid.MustParse("42000000-0000-0000-0000-000000000003")
	tagID := uuid.MustParse("42000000-0000-0000-0000-000000000004")
	linktest.SeedRecordLink(
		t, harness.DB, incident.ID, actor.ID, linkID, sourceID, targetID, "references_record", "manual", nil,
	)
	linktest.SeedRecordTag(t, harness.DB, incident.ID, actor.ID, tagID, sourceID, "urgent")

	exported, err := port.Export(ctx, sourceport.ExportContext{Query: harness.DB, IncidentID: incident.ID})
	if err != nil {
		t.Fatalf("export source state: %v", err)
	}
	bundle := make(sourceport.MapBundle, len(exported))
	for _, file := range exported {
		bundle[file.Path] = append([]byte(nil), file.Payload...)
	}
	if _, err := harness.DB.Exec(ctx, `DELETE FROM record_tags WHERE incident_id = $1`, incident.ID); err != nil {
		t.Fatalf("remove exported tag rows: %v", err)
	}
	if _, err := harness.DB.Exec(ctx, `DELETE FROM record_links WHERE incident_id = $1`, incident.ID); err != nil {
		t.Fatalf("remove exported link rows: %v", err)
	}

	t.Run("invalid endpoint is classified before writes", func(t *testing.T) {
		invalid := clonePortableBundle(bundle)
		rows, err := incidentportability.DecodeNDJSON(invalid["data/record_links.ndjson"])
		if err != nil || len(rows) != 1 {
			t.Fatalf("decode exported link row: rows=%d err=%v", len(rows), err)
		}
		rows[0]["dst_record_id"] = uuid.MustParse("42000000-0000-0000-0000-000000000099").String()
		invalid["data/record_links.ndjson"], err = incidentportability.CanonicalJSONString(rows[0])
		if err != nil {
			t.Fatalf("encode invalid endpoint row: %v", err)
		}
		recorder := &linksPortableAttributionRecorder{}
		importContext := sourceport.ImportContext{
			IncidentID: incident.ID, ActorUserID: actor.ID, BundleVersion: 3,
			OperationID: "links-invalid-endpoint", Attributions: recorder,
		}
		prepared, err := port.PrepareImport(ctx, invalid, importContext)
		if err != nil {
			t.Fatalf("prepare row-local-valid endpoint failure: %v", err)
		}
		tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin invalid endpoint import: %v", err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		err = port.ApplyImportTx(ctx, tx, prepared, importContext)
		var failure *sourceport.Failure
		if !errors.As(err, &failure) || failure.InvariantID() != "links_tags.endpoints_same_incident" {
			t.Fatalf("invalid endpoint error = %T %[1]v", err)
		}
		var count int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM record_links WHERE incident_id = $1`, incident.ID).Scan(&count); err != nil {
			t.Fatalf("count rejected link rows: %v", err)
		}
		if count != 0 || len(recorder.values) != 0 {
			t.Fatalf("failed import left count=%d attributions=%#v", count, recorder.values)
		}
	})

	recorder := &linksPortableAttributionRecorder{}
	importContext := sourceport.ImportContext{
		IncidentID: incident.ID, ActorUserID: actor.ID, BundleVersion: 3,
		OperationID: "links-valid-round-trip", Attributions: recorder,
	}
	prepared, err := port.PrepareImport(ctx, bundle, importContext)
	if err != nil {
		t.Fatalf("prepare owner-exported v3 bytes: %v", err)
	}
	tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin valid import: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := port.ApplyImportTx(ctx, tx, prepared, importContext); err != nil {
		t.Fatalf("apply owner-exported v3 bytes: %v", err)
	}
	if err := port.ValidateImportTx(ctx, tx, prepared, importContext); err != nil {
		t.Fatalf("validate owner-exported v3 bytes: %v", err)
	}
	reexported, err := port.Export(ctx, sourceport.ExportContext{Query: tx, IncidentID: incident.ID})
	if err != nil {
		t.Fatalf("re-export imported source state: %v", err)
	}
	if len(reexported) != len(exported) {
		t.Fatalf("re-export file count = %d, want %d", len(reexported), len(exported))
	}
	for index := range exported {
		if reexported[index].Path != exported[index].Path || !bytes.Equal(reexported[index].Payload, exported[index].Payload) {
			t.Fatalf("re-exported file %d differs: got %q %q, want %q %q", index, reexported[index].Path, reexported[index].Payload, exported[index].Path, exported[index].Payload)
		}
	}
}

type linksPortableAttributionRecorder struct {
	values []incidentportability.ImportedAttribution
}

func (recorder *linksPortableAttributionRecorder) RecordImportedAttribution(
	table string,
	rowID string,
	column string,
	actorID string,
) error {
	recorder.values = append(recorder.values, incidentportability.ImportedAttribution{
		SourceTable: table, SourceRowID: rowID, SourceColumn: column, SourceActorID: actorID,
	})
	return nil
}

func (recorder *linksPortableAttributionRecorder) ImportedAttributions() []incidentportability.ImportedAttribution {
	return append([]incidentportability.ImportedAttribution(nil), recorder.values...)
}

func clonePortableBundle(source sourceport.MapBundle) sourceport.MapBundle {
	result := make(sourceport.MapBundle, len(source))
	for path, payload := range source {
		result[path] = append([]byte(nil), payload...)
	}
	return result
}
