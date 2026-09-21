package appsupport

import (
	"bytes"
	"context"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Re-import into a rolled-back transaction so the test never rewrites retained
// history or receipts. Links owns canonical source-state portability.
func RequireLinksPortableRoundTrip(t testing.TB, db postgres.DB, incidentID, actorID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	port, err := links.NewIncidentBundleSourcePort()
	if err != nil {
		t.Fatal(err)
	}
	exported, err := port.Export(ctx, sourceport.ExportContext{Query: db, IncidentID: incidentID})
	if err != nil {
		t.Fatal(err)
	}
	bundle := sourceport.MapBundle{}
	for _, file := range exported {
		bundle[file.Path] = file.Payload
	}
	imported := sourceport.ImportContext{IncidentID: incidentID, ActorUserID: actorID, BundleVersion: 3, OperationID: "note-associations-round-trip", Attributions: &associationAttributions{}}
	prepared, err := port.PrepareImport(ctx, bundle, imported)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	for _, table := range []string{"record_links", "record_tags"} {
		if _, err := tx.Exec(ctx, "DELETE FROM "+table+" WHERE incident_id=$1", incidentID); err != nil {
			t.Fatal(err)
		}
	}
	if err := port.ApplyImportTx(ctx, tx, prepared, imported); err != nil {
		t.Fatal(err)
	}
	if err := port.ValidateImportTx(ctx, tx, prepared, imported); err != nil {
		t.Fatal(err)
	}
	again, err := port.Export(ctx, sourceport.ExportContext{Query: tx, IncidentID: incidentID})
	if err != nil {
		t.Fatal(err)
	}
	if len(again) != len(exported) {
		t.Fatal("portable file set changed")
	}
	for i := range again {
		if again[i].Path != exported[i].Path || !bytes.Equal(again[i].Payload, exported[i].Payload) {
			t.Fatalf("association identity or provenance changed in %s", again[i].Path)
		}
	}
}

type associationAttributions struct {
	values []incidentportability.ImportedAttribution
}

func (a *associationAttributions) RecordImportedAttribution(table, row, column, actor string) error {
	a.values = append(a.values, incidentportability.ImportedAttribution{SourceTable: table, SourceRowID: row, SourceColumn: column, SourceActorID: actor})
	return nil
}
func (a *associationAttributions) ImportedAttributions() []incidentportability.ImportedAttribution {
	return a.values
}
