package revisions_test

import (
	"context"
	"errors"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"testing"
	"time"
)

func exerciseAppendAfterState(t *testing.T) {
	h := appsupport.StartStore(t, "append-after-state")
	actor := authstoretest.SeedActor(t, h.DB, authstoretest.ActorAttributes{Email: "append-after@example.test", DisplayName: "Append after"})
	incident := appsupport.CreateIncidentInStore(t, h.DB, actor, "append-after-incident", "IR-AFTER", "Append after")
	record := uuid.New()
	entitytest.SeedHostRecord(t, h.DB, incident.ID, actor.ID, record, "Before", "append-after", "", "")
	appender := revisionsupport.MustComposition(t).Runtime.Appender()
	for _, kind := range []string{"mutation", "live", "historical"} {
		t.Run(kind, func(t *testing.T) {
			tx, err := h.DB.BeginTx(t.Context(), pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(context.Background())
			before, err := appender.CaptureRecordSnapshotTx(t.Context(), tx, record)
			if err != nil {
				t.Fatal(err)
			}
			changed, err := appender.AppendChangeSetTx(t.Context(), tx, revisions.AppendChangeSetParams{IncidentID: incident.ID, ActorUserID: actor.ID, Source: "append.after.test", CreatedAt: time.Now().UTC()})
			if err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "mutation":
				err = appender.AppendRecordMutationTx(t.Context(), tx, revisions.AppendRecordMutationParams{ChangeSetID: changed, SequenceNo: 1, TargetKind: "host", RecordID: record, OperationKind: "patch", BeforeSnapshot: &before})
			case "live":
				err = appender.AppendLiveRevisionTx(t.Context(), tx, revisions.LiveRevisionInput{ChangeSetID: changed, RecordID: record, RowVersion: 2, BeforeSnapshot: &before})
			case "historical":
				err = appender.AppendHistoricalRevisionTx(t.Context(), tx, revisions.HistoricalRevisionInput{ChangeSetID: changed, RecordID: record, RowVersion: 2, BeforeSnapshot: &before})
			}
			if !errors.Is(err, revisions.ErrInvalidRecordSnapshot) {
				t.Errorf("missing after-state error=%v; want invalid snapshot", err)
			}
			var mutations, revisionsCount int
			if err := tx.QueryRow(t.Context(), `SELECT (SELECT count(*) FROM change_set_mutations WHERE change_set_id=$1),(SELECT count(*) FROM record_revisions WHERE change_set_id=$1)`, changed).Scan(&mutations, &revisionsCount); err != nil || mutations != 0 || revisionsCount != 0 {
				t.Errorf("invalid append persisted history: mutations=%d revisions=%d error=%v", mutations, revisionsCount, err)
			}
		})
	}
}
