package revisions

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
)

type commandTransactionWitness struct {
	postgres.DB
	tx     *commandObservedTx
	begins int
}
type commandObservedTx struct {
	pgx.Tx
	commits, rollbacks int
}

func (w *commandTransactionWitness) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := w.DB.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	w.begins++
	w.tx = &commandObservedTx{Tx: tx}
	return w.tx, nil
}
func (tx *commandObservedTx) Commit(ctx context.Context) error {
	tx.commits++
	return tx.Tx.Commit(ctx)
}
func (tx *commandObservedTx) Rollback(ctx context.Context) error {
	tx.rollbacks++
	return tx.Tx.Rollback(ctx)
}

type commandEnvelopeWitness struct{ commandServiceTestEnvelopes }

func (w commandEnvelopeWitness) LoadEnvelopeTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, lock bool) (RecordEnvelope, error) {
	query := "SELECT record_id,incident_id,record_type,row_version,created_by_user_id,created_at,updated_by_user_id,updated_at,deleted_at,deleted_by_user_id FROM records WHERE record_id=$1"
	if lock {
		query += " FOR UPDATE"
	}
	var e RecordEnvelope
	err := tx.QueryRow(ctx, query, id).Scan(&e.RecordID, &e.IncidentID, &e.RecordType, &e.RowVersion, &e.CreatedByUserID, &e.CreatedAt, &e.UpdatedByUserID, &e.UpdatedAt, &e.DeletedAt, &e.DeletedByUserID)
	return e, err
}
func (w commandEnvelopeWitness) LockDestructiveRecordsNowaitTx(ctx context.Context, tx pgx.Tx, ids []uuid.UUID) error {
	for _, id := range ids {
		var found uuid.UUID
		if err := tx.QueryRow(ctx, "SELECT record_id FROM records WHERE record_id=$1 FOR UPDATE NOWAIT", id).Scan(&found); err != nil {
			return err
		}
	}
	return nil
}

type denyingCommandAuthorizer struct {
	t       *testing.T
	runner  *commandTransactionWitness
	fixture historyRepositoryFixture
	kind    CommandKind
	calls   int
}

func (w *denyingCommandAuthorizer) AuthorizeCommandTx(ctx context.Context, tx pgx.Tx, incident uuid.UUID, actor ActorID, kind CommandKind) error {
	w.calls++
	if tx != w.runner.tx || incident != w.fixture.record.IncidentID || actor.UUID() != w.fixture.actorID || kind != w.kind {
		w.t.Fatalf("wrong authorization boundary: tx=%v incident=%s actor=%s command=%s", tx == w.runner.tx, incident, actor.UUID(), kind)
	}
	// A write made through this participant is a rollback witness, not an
	// authorization implementation. It must not escape a denied transaction.
	if _, err := tx.Exec(ctx, "UPDATE records SET row_version=row_version+100 WHERE record_id=$1", w.fixture.record.RecordID); err != nil {
		return err
	}
	return ErrCommandRoleDenied
}

type authorizationTargetProvider struct {
	stubNonRowProvider
	recordID uuid.UUID
}

func (p authorizationTargetProvider) DescribeTx(context.Context, pgx.Tx, rollbackcontract.DescribeRequest) (rollbackcontract.TargetDescriptor, error) {
	return rollbackcontract.TargetDescriptor{AffectedRecordIDs: []uuid.UUID{p.recordID}}, nil
}
func TestDestructiveCommandAuthorizationTransaction_Integration(t *testing.T) {
	for _, kind := range []CommandKind{CommandSoftDelete, CommandRestore, CommandRollback} {
		t.Run(string(kind), func(t *testing.T) {
			fixture := newHistoryRepositoryFixture(t)
			ctx := context.Background()
			if kind == CommandRestore {
				if _, err := fixture.database.Exec(ctx, "UPDATE records SET deleted_at=now(),deleted_by_user_id=$2 WHERE record_id=$1", fixture.record.RecordID, fixture.actorID); err != nil {
					t.Fatal(err)
				}
			}
			runner := &commandTransactionWitness{DB: fixture.database}
			authorization := &denyingCommandAuthorizer{t: t, runner: runner, fixture: fixture, kind: kind}
			dependencies := validCommandServiceDependencies(t)
			dependencies.Transactions = runner
			dependencies.Authorization = authorization
			dependencies.RecordEnvelopes = commandEnvelopeWitness{}
			contributions := validProviderContributions()
			for i := range contributions {
				for j := range contributions[i].NonRowTargets {
					if contributions[i].NonRowTargets[j].TargetKind == "record_tag" {
						contributions[i].NonRowTargets[j].RollbackProvider = authorizationTargetProvider{recordID: fixture.record.RecordID}
					}
				}
			}
			dependencies.TargetSemantics = validTargetSemanticsCatalog(t, contributions)
			service, err := NewCommandService(dependencies)
			if err != nil {
				t.Fatal(err)
			}
			state := func() map[string]string {
				result := map[string]string{}
				for _, table := range []string{"records", "hosts", "record_tags", "host_grid_projection", "change_sets", "change_set_mutations", "record_revisions", "route_idempotency"} {
					var value string
					if err := fixture.database.QueryRow(ctx, "SELECT COALESCE(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text),'[]'::jsonb)::text FROM "+table+" t").Scan(&value); err != nil {
						t.Fatal(err)
					}
					result[table] = value
				}
				result["intents"] = fmt.Sprint(collaborationsupport.CountIntents(t, fixture.database, collaborationsupport.IntentSelector{IncidentID: fixture.record.IncidentID.String()}))
				return result
			}
			before := state()
			command := DeleteRestoreCommand{Actor: NewActorID(fixture.actorID), RecordID: fixture.record.RecordID, Request: DeleteRestoreRequest{BaseRowVersion: 1, ClientTxnID: "deny-command"}}
			switch kind {
			case CommandSoftDelete:
				_, err = service.SoftDeleteRecord(ctx, command)
			case CommandRestore:
				_, err = service.RestoreRecord(ctx, command)
			case CommandRollback:
				_, err = service.RollbackRecord(ctx, RollbackCommand{Actor: command.Actor, RecordID: command.RecordID, Request: RollbackRequest{BaseRowVersion: 1, ClientTxnID: "deny-command", Target: RollbackTarget{Kind: "change_set", ChangeSetID: fixture.changeSetID.String()}}})
			}
			if !errors.Is(err, ErrCommandRoleDenied) {
				t.Fatalf("command denial=%v", err)
			}
			if runner.begins != 1 || authorization.calls != 1 || runner.tx.commits != 0 || runner.tx.rollbacks != 1 {
				t.Fatalf("transaction lifecycle: begins=%d auth=%d tx=%#v", runner.begins, authorization.calls, runner.tx)
			}
			if after := state(); !reflect.DeepEqual(before, after) {
				t.Fatalf("denial changed durable state: before=%v after=%v", before, after)
			}
		})
	}
}
