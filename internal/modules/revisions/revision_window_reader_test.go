package revisions_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	envelopetest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestRevisionWindowReaderOrdersRowsAndNormalizesUTC_Integration(t *testing.T) {
	harness := appsupport.StartStore(t, "revisions-window-reader")
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"revision-window@example.test",
		"Revision Window",
		"RevisionWindowPass1!",
		false,
		false,
		true,
	)
	incident := appsupport.CreateIncidentInStore(
		t,
		harness.DB,
		actor,
		"txn-revisions-window-incident",
		"IR-REVISION-WINDOW",
		"Revision window reader",
	)
	recordID := uuid.New()
	envelopetest.SeedRecordEnvelope(t, harness.DB, incident.ID, actor.ID, recordID, "host")
	changeSetID := uuid.New()
	createdAt := time.Date(2026, 8, 3, 9, 15, 0, 123456000, time.FixedZone("offset", 2*60*60))
	if _, err := harness.DB.Exec(context.Background(), `
INSERT INTO change_sets (change_set_id, incident_id, actor_user_id, source, created_at)
VALUES ($1, $2, $3, 'revision_window_test', $4)
`, changeSetID, incident.ID, actor.ID, createdAt); err != nil {
		t.Fatalf("seed change set: %v", err)
	}
	if _, err := harness.DB.Exec(context.Background(), `
INSERT INTO record_revisions (change_set_id, record_id, row_version, before_json, after_json, created_at)
VALUES
    ($1, $2, 2, '{"cells":{}}', '{"cells":{"host.name":{"value":"two"}}}', $3),
    ($1, $2, 1, NULL, '{"cells":{"host.name":{"value":"one"}}}', $3)
`, changeSetID, recordID, createdAt); err != nil {
		t.Fatalf("seed revision rows: %v", err)
	}

	tx, err := harness.DB.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin reader transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	rows, err := conflicts.NewRevisionWindowReader().LoadRevisionWindowTx(
		context.Background(),
		tx,
		recordID,
		1,
		2,
	)
	if err != nil {
		t.Fatalf("load revision window: %v", err)
	}
	if len(rows) != 2 || rows[0].RowVersion != 1 || rows[1].RowVersion != 2 {
		t.Fatalf("ordered rows = %#v", rows)
	}
	for _, row := range rows {
		if row.CreatedAt.Location() != time.UTC || !row.CreatedAt.Equal(createdAt) {
			t.Fatalf("created_at = %s (%s), want UTC instant %s", row.CreatedAt, row.CreatedAt.Location(), createdAt)
		}
	}

	if _, err := conflicts.NewRevisionWindowReader().LoadRevisionWindowTx(context.Background(), nil, recordID, 1, 2); !errors.Is(err, conflicts.ErrInvalidRevisionWindow) {
		t.Fatalf("nil transaction error = %v", err)
	}
	if _, err := conflicts.NewRevisionWindowReader().LoadRevisionWindowTx(context.Background(), tx, uuid.Nil, 1, 2); !errors.Is(err, conflicts.ErrInvalidRevisionWindow) {
		t.Fatalf("nil record error = %v", err)
	}
	if _, err := conflicts.NewRevisionWindowReader().LoadRevisionWindowTx(context.Background(), tx, recordID, 2, 1); !errors.Is(err, conflicts.ErrInvalidRevisionWindow) {
		t.Fatalf("inverted range error = %v", err)
	}
}
