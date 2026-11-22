package portability

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

// RequireTagExportImportRoundTrip exercises a tag-only incident through the Links bundle port.
func RequireTagExportImportRoundTrip(t *testing.T, db *pgxpool.Pool, incident, actor uuid.UUID) {
	t.Helper()
	ctx := t.Context()
	port, err := links.NewIncidentBundleSourcePort()
	if err != nil {
		t.Fatal(err)
	}
	files, err := portabilitysupport.Export(port, ctx, sourceport.ExportContext{
		Query: db, IncidentID: incident, PortableAttributions: incidentbundles.PortableSourceAttributions(),
	})
	if err != nil {
		t.Fatal(err)
	}
	bundle := portabilitysupport.MapBundle{}
	for _, file := range files {
		bundle[file.Path] = file.Payload
	}
	actors, err := portabilitysupport.ReferencedActors(bundle)
	if err != nil {
		t.Fatal(err)
	}
	input := sourceport.ImportContext{
		IncidentID: incident, ActorUserID: actor, BundleVersion: 8, OperationID: "repaired-timeline-tags",
		Actors: actors, Attributions: portabilitysupport.AttributionCapability{}, Preparation: portabilitysupport.NewPreparationStore(),
	}
	prepared, err := port.PrepareImport(ctx, bundle, input)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM public.record_tags WHERE incident_id=$1`, incident); err != nil {
		t.Fatal(err)
	}
	if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
		t.Fatal(err)
	}
	if err := port.ValidateImportTx(ctx, tx, prepared, input); err != nil {
		t.Fatal(err)
	}
}
