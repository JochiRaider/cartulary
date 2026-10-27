package entities_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mentioneffects"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/versionid"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

type mentionUnusedLinks struct{ mentions.LinkOperationsPort }
type mentionEffectWitness struct {
	t       *testing.T
	tx      pgx.Tx
	state   mentioneffects.ActionState
	command mentioneffects.ActionCommand
	calls   []string
	fail    error
}

func (w *mentionEffectWitness) PrepareMentionActionTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (mentioneffects.ActionState, error) {
	w.calls = append(w.calls, "prepare")
	w.tx = tx
	var version int64
	if err := tx.QueryRow(ctx, "SELECT row_version FROM records WHERE record_id=$1", id).Scan(&version); err != nil {
		return mentioneffects.ActionState{}, err
	}
	w.state = mentioneffects.ActionState{SourceRecordID: id, RowVersion: version, BeforeVersionID: versionid.Format(id, version)}
	return w.state, nil
}
func (w *mentionEffectWitness) ApplyMentionActionEffectsTx(ctx context.Context, tx pgx.Tx, state mentioneffects.ActionState, command mentioneffects.ActionCommand) (mentioneffects.ActionResult, error) {
	w.calls = append(w.calls, "apply")
	if tx != w.tx || !reflect.DeepEqual(state, w.state) || command != w.command {
		w.t.Fatalf("effect transaction/state/command mismatch: %v %#v %#v", tx == w.tx, state, command)
	}
	var status string
	if err := tx.QueryRow(ctx, "SELECT resolution_status FROM entity_mentions WHERE source_record_id=$1", state.SourceRecordID).Scan(&status); err != nil {
		return mentioneffects.ActionResult{}, err
	}
	if status != "dismissed" {
		w.t.Fatalf("effect cannot see source mutation: %s", status)
	}
	// The injected participant writes through the supplied transaction. A later
	// error must undo both this write and the mention mutation already performed.
	if _, err := tx.Exec(ctx, "UPDATE records SET row_version=row_version+1, updated_at=$2, updated_by_user_id=$3 WHERE record_id=$1", state.SourceRecordID, command.EffectiveAt, command.ActorUserID); err != nil {
		return mentioneffects.ActionResult{}, err
	}
	if _, err := tx.Exec(ctx, "UPDATE timeline_events SET row_version=row_version+1, edited_at=$2, updated_by_user_id=$3 WHERE record_id=$1", state.SourceRecordID, command.EffectiveAt, command.ActorUserID); err != nil {
		return mentioneffects.ActionResult{}, err
	}
	return mentioneffects.ActionResult{SourceRecordID: state.SourceRecordID, RowVersion: state.RowVersion + 1, BeforeVersionID: state.BeforeVersionID, AfterVersionID: versionid.Format(state.SourceRecordID, state.RowVersion+1)}, w.fail
}

func TestMentionCommandEffectsTransaction_Integration(t *testing.T) {
	for _, failure := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "effect failure"}[failure], func(t *testing.T) {
			harness := appsupport.StartStore(t, "mention-effect")
			actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "mention-effect@example.test", DisplayName: "Mention effect"})
			incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "mention-effect-incident", "IR-EFFECT", "Mention effects")
			recordID, mentionID := uuid.New(), uuid.New()
			timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, recordID)
			entitytest.SeedMention(t, harness.DB, actor.ID, mentionID, recordID, timelinetest.FieldHostRefs, "host", "raw preserved", "unresolved", nil, nil)
			now := time.Now().UTC().Truncate(time.Microsecond)
			witness := &mentionEffectWitness{t: t, command: mentioneffects.ActionCommand{IncidentID: incident.ID, ActorUserID: actor.ID, EffectiveAt: now}}
			sentinel := errors.New("injected effect failure")
			if failure {
				witness.fail = sentinel
			}
			composition := revisionsupport.MustComposition(t)
			projection, err := projectionassembly.Build(harness.DB)
			if err != nil {
				t.Fatal(err)
			}
			store, err := mentions.NewStore(mentions.StoreDependencies{Postgres: harness.DB, Revisions: composition.Runtime.Appender(), Links: mentionUnusedLinks{}, Projections: projection.EntityMutationRows(), Timeline: witness, Collaboration: composition.RecordChanges})
			if err != nil {
				t.Fatal(err)
			}
			counts := func() []int {
				var result []int
				for _, query := range []string{
					"SELECT count(*) FROM change_sets", "SELECT count(*) FROM change_set_mutations", "SELECT count(*) FROM record_revisions", "SELECT count(*) FROM route_idempotency",
				} {
					result = append(result, appsupport.QueryCount(t, harness.DB, query))
				}
				result = append(result, collaborationsupport.CountIntents(t, harness.DB, collaborationsupport.IntentSelector{IncidentID: incident.ID.String()}))
				return result
			}
			before := counts()
			result, err := store.ApplyMentionAction(context.Background(), actor, mentionID, mentions.MentionActionRequest{Action: "dismiss_item", BaseMentionRowVersion: 1, ClientTxnID: "mention-effect-command"}, []byte("effect"), "effect-request", now)
			if failure {
				if !errors.Is(err, sentinel) {
					t.Fatalf("failure = %v", err)
				}
				if after := counts(); !reflect.DeepEqual(before, after) {
					t.Fatalf("failure published effects: %v -> %v", before, after)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if result.SourceRecordRowVersion != 2 || result.ChangeSetID == uuid.Nil {
					t.Fatalf("result = %#v", result)
				}
				after := counts()
				for i := range after {
					if after[i] <= before[i] {
						t.Fatalf("missing durable effect %d: %v -> %v", i, before, after)
					}
				}
			}
			if !reflect.DeepEqual(witness.calls, []string{"prepare", "apply"}) {
				t.Fatalf("effect calls = %v", witness.calls)
			}
			mention := entitytest.LookupMention(t, harness.DB, mentionID)
			expectedStatus := "dismissed"
			expectedVersion := int64(2)
			if failure {
				expectedStatus = "unresolved"
				expectedVersion = 1
			}
			if mention.ResolutionStatus != expectedStatus || mention.RowVersion != expectedVersion || mention.RawText != "raw preserved" {
				t.Fatalf("mention = %#v", mention)
			}
			for _, table := range []string{"records", "timeline_events"} {
				var version int64
				if err := harness.DB.QueryRow(context.Background(), "SELECT row_version FROM "+table+" WHERE record_id=$1", recordID).Scan(&version); err != nil {
					t.Fatal(err)
				}
				if version != expectedVersion {
					t.Fatalf("%s version=%d want %d", table, version, expectedVersion)
				}
			}
		})
	}
}
