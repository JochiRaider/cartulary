package artifacts_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestArtifactWorkbookFacadePersistsArtifactBackedNotes_Unit(t *testing.T) {
	harness := appsupport.StartStore(t, "artifacts-workbook-facade-note")
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"artifact-owner@example.test",
		"Artifact Owner",
		"ArtifactOwnerPass1!",
		false,
		false,
		true,
	)
	incident := appsupport.CreateIncidentInStore(
		t,
		harness.DB,
		actor,
		"txn-artifacts-note-incident",
		"IR-ARTIFACTS-NOTE",
		"Artifact owner note",
	)
	facade := artifacts.NewWorkbookFacade(
		harness.DB,
		conflicttest.NewCodec("artifacts"),
		revisionsupport.MustAppender(t),
	)
	title := "Artifact-owned note"
	body := "Persisted by the artifact source facade"
	result, err := facade.Create(context.Background(), artifacts.WorkbookCreateCommand{
		Actor:      actor,
		IncidentID: incident.ID,
		Request: artifacts.WorkbookCreateRequest{
			ViewSchemaID: artifacts.NotesViewSchemaID,
			ClientTxnID:  "txn-artifacts-note-create",
			Values: map[string]artifacts.FieldValue{
				"note.title": {Text: &title},
				"note.body":  {Text: &body},
			},
			Collections: map[string]artifacts.WorkbookCollectionActionPayload{
				"note.tags": {
					Actions: []artifacts.WorkbookCollectionAction{{
						Op:             "add_tag",
						RawText:        "owner-local",
						NormalizedText: "owner-local",
					}},
				},
			},
		},
		RequestHash: []byte("artifacts-note-create"),
		RequestID:   "req-artifacts-note-create",
		RouteKey:    "workbook.rows.create",
		Now:         time.Date(2026, 7, 27, 1, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("create note through artifact owner facade: %v", err)
	}
	if result.RecordID == uuid.Nil {
		t.Fatal("artifact owner facade returned a nil record id")
	}

	requireCount(t, harness, `
SELECT count(*)
  FROM records r
  JOIN artifacts a
    ON a.incident_id = r.incident_id
   AND a.record_id = r.record_id
 WHERE r.record_id = $1
   AND r.record_type = 'artifact'
   AND a.artifact_type = 'note'
   AND a.title = $2
   AND a.body = $3
`, result.RecordID, title, body, 1)
	requireCount(
		t,
		harness,
		`SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'notes'`,
		0,
	)
	requireCount(t, harness, `
SELECT count(*)
  FROM record_tags
 WHERE incident_id = $1
   AND record_id = $2
   AND normalized_tag_name = 'owner-local'
   AND deleted_at IS NULL
`, incident.ID, result.RecordID, 1)
	requireCount(
		t,
		harness,
		`SELECT count(*) FROM record_revisions WHERE record_id = $1 AND row_version = 1`,
		result.RecordID,
		1,
	)
}

func requireCount(
	t testing.TB,
	harness *appsupport.StoreHarness,
	query string,
	argsAndWant ...any,
) {
	t.Helper()
	if len(argsAndWant) == 0 {
		t.Fatal("count assertion requires an expected value")
	}
	want := argsAndWant[len(argsAndWant)-1].(int)
	args := argsAndWant[:len(argsAndWant)-1]
	var got int
	if err := harness.DB.QueryRow(context.Background(), query, args...).Scan(&got); err != nil {
		t.Fatalf("query count: %v", err)
	}
	if got != want {
		t.Fatalf("count = %d, want %d", got, want)
	}
}
