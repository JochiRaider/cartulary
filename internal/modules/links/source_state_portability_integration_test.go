package links_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	linktest "github.com/JochiRaider/cartulary/internal/modules/links/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

func TestLinksSourceStateStrictRoundTripAndAtomicFailure_Integration(t *testing.T) {
	ctx := context.Background()
	port, err := links.NewIncidentBundleSourcePort()
	if err != nil {
		t.Fatalf("construct Links source port: %v", err)
	}
	harness := appsupport.StartStore(t, "links-source-state-round-trip")
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "links-source-state@example.test", DisplayName: "Links Source State", DeploymentAdmin: true})
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

	exported, err := portabilitysupport.Export(port, ctx, sourceport.ExportContext{Query: harness.DB, IncidentID: incident.ID, PortableAttributions: incidentbundles.PortableSourceAttributions()})
	if err != nil {
		t.Fatalf("export source state: %v", err)
	}
	bundle := make(portabilitysupport.MapBundle, len(exported))
	for _, file := range exported {
		bundle[file.Path] = append([]byte(nil), file.Payload...)
	}
	if _, err := harness.DB.Exec(ctx, `DELETE FROM record_tags WHERE incident_id = $1`, incident.ID); err != nil {
		t.Fatalf("remove exported tag rows: %v", err)
	}
	if _, err := harness.DB.Exec(ctx, `DELETE FROM record_links WHERE incident_id = $1`, incident.ID); err != nil {
		t.Fatalf("remove exported link rows: %v", err)
	}

	for _, test := range []struct {
		name, want                                          string
		catalogMissing, duplicate, foreign, missingEndpoint bool
	}{
		{name: "catalog precedes uniqueness", want: "links_tags.tag_catalog_exact", catalogMissing: true, duplicate: true},
		{name: "uniqueness precedes source scope", want: "links_tags.link_unique", duplicate: true, foreign: true},
		{name: "source scope", want: "links_tags.endpoints_same_incident", foreign: true},
		{name: "invalid endpoint is classified before writes", want: "links_tags.endpoints_same_incident", missingEndpoint: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			invalid := clonePortableBundle(bundle)
			rows, err := incidentportability.DecodeNDJSON(invalid["data/record_links.ndjson"])
			if err != nil || len(rows) != 1 {
				t.Fatalf("decode exported link: %v", err)
			}
			if test.foreign {
				rows[0]["incident_id"] = "42000000-0000-0000-0000-000000000098"
			}
			if test.missingEndpoint {
				rows[0]["dst_record_id"] = "42000000-0000-0000-0000-000000000099"
			}
			if test.duplicate {
				duplicate := map[string]any{}
				for key, value := range rows[0] {
					duplicate[key] = value
				}
				duplicate["record_link_id"] = "42000000-0000-0000-0000-000000000005"
				rows = append(rows, duplicate)
			}
			invalid["data/record_links.ndjson"] = nil
			for _, row := range rows {
				raw, err := incidentportability.CanonicalJSONString(row)
				if err != nil {
					t.Fatal(err)
				}
				invalid["data/record_links.ndjson"] = append(invalid["data/record_links.ndjson"], raw...)
			}
			if test.catalogMissing {
				invalid["data/tags.ndjson"] = nil
			}
			recorder := &linksPortableAttributionRecorder{}
			importContext := sourceport.ImportContext{IncidentID: incident.ID, ActorUserID: actor.ID, BundleVersion: 6, OperationID: test.name, Attributions: recorder}
			prepared, err := port.PrepareImport(ctx, invalid, importContext)
			if err != nil {
				t.Fatalf("prepare row-local valid input: %v", err)
			}
			tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			err = port.ApplyImportTx(ctx, tx, prepared, importContext)
			var failure *sourceport.Failure
			if !errors.As(err, &failure) || failure.InvariantID() != test.want {
				t.Fatalf("error = %T %[1]v, want %s", err, test.want)
			}
			var count int
			if err := tx.QueryRow(ctx, `SELECT (SELECT count(*) FROM record_links WHERE incident_id = $1) + (SELECT count(*) FROM record_tags WHERE incident_id = $1)`, incident.ID).Scan(&count); err != nil {
				t.Fatal(err)
			}
			if count != 0 || len(recorder.values) != 0 {
				t.Fatalf("failed import left count=%d attributions=%#v", count, recorder.values)
			}
		})
	}

	recorder := &linksPortableAttributionRecorder{}
	importContext := sourceport.ImportContext{
		IncidentID: incident.ID, ActorUserID: actor.ID, BundleVersion: 6,
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
	reexported, err := portabilitysupport.Export(port, ctx, sourceport.ExportContext{Query: tx, IncidentID: incident.ID, PortableAttributions: incidentbundles.PortableSourceAttributions()})
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

func clonePortableBundle(source portabilitysupport.MapBundle) portabilitysupport.MapBundle {
	result := make(portabilitysupport.MapBundle, len(source))
	for path, payload := range source {
		result[path] = append([]byte(nil), payload...)
	}
	return result
}

func (r *linksPortableAttributionRecorder) VisitImportedAttributions(table, rowID string, visit func(incidentportability.ImportedAttribution) error) error {
	for _, row := range r.ImportedAttributions() {
		if row.SourceTable == table && (rowID == "" || row.SourceRowID == rowID) {
			if err := visit(row); err != nil {
				return err
			}
		}
	}
	return nil
}

func TestLinksBundleValidationCommitIsolation_Integration(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "links-bundle-validation-commit")
	first, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = first.Rollback(ctx) }()
	second, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = second.Rollback(ctx) }()
	for _, tx := range []pgx.Tx{first, second} {
		if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout='2s'`); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO links_bundle_validation_catalog(tag_name,normalized_tag_name) VALUES ('Portable','portable')`); err != nil {
			t.Fatalf("independent transaction keys collided: %v", err)
		}
	}
	err = first.Commit(ctx)
	var databaseError *pgconn.PgError
	if !errors.As(err, &databaseError) || databaseError.Code != "P0001" {
		t.Fatalf("unreleased work commit = %v", err)
	}
	if _, err := second.Exec(ctx, `DELETE FROM links_bundle_validation_catalog WHERE transaction_id=pg_current_xact_id()`); err != nil {
		t.Fatal(err)
	}
	if err := second.Commit(ctx); err != nil {
		t.Fatalf("released work commit: %v", err)
	}
	var count int
	if err := harness.DB.QueryRow(ctx, `SELECT count(*) FROM links_bundle_validation_catalog`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("committed scratch rows=%d", count)
	}
}
