package networkflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

// The authenticated route fixture supplies a real committed receipt and owners.
// Only invocation ordering and the failed coordinator read are controlled here.
func AssertIndicatorLinkReplayAdmission(t *testing.T, store *store, incidentID, actorID uuid.UUID, body, expected map[string]any, status int, configure func(*Module) crossownertransaction.Options) {
	t.Helper()
	options := configure(&Module{limits: store.limits})
	ctx := t.Context()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	var members map[string]json.RawMessage
	if err := json.Unmarshal(raw, &members); err != nil {
		t.Fatal(err)
	}
	request, failure := decodeIndicatorLinkObject(members, store.limits)
	if failure != nil {
		t.Fatal(failure)
	}
	hash, err := indicatorLinkRequestHash(request)
	if err != nil {
		t.Fatal(err)
	}
	key := indicatorLinkIdempotencyKey(actorID, incidentID, request.ClientTxnID)
	reader := authn.NewStore(store.pool)
	original, err := reader.GetRouteIdempotency(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	var membership string
	if err := store.pool.QueryRow(ctx, "SELECT row_to_json(m)::text FROM incident_memberships m WHERE incident_id=$1 AND user_id=$2", incidentID, actorID).Scan(&membership); err != nil {
		t.Fatal(err)
	}
	target, err := indicatorIDFromLinkPayload(expected)
	if err != nil {
		t.Fatal(err)
	}
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := store.pool.Exec(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	snapshot := func() []string {
		t.Helper()
		var values [5]string
		queries := []string{
			"SELECT COALESCE(jsonb_agg(to_jsonb(r) ORDER BY r.route_key,r.actor_user_id,r.client_txn_id)::text,'[]') FROM route_idempotency r WHERE scope_key=$1",
			"SELECT COALESCE(jsonb_agg(to_jsonb(r) ORDER BY to_jsonb(r)::text)::text,'[]') FROM network_flow_indicator_bindings r WHERE incident_id=$1",
			"SELECT COALESCE(jsonb_agg(to_jsonb(r) ORDER BY to_jsonb(r)::text)::text,'[]') FROM indicators r WHERE incident_id=$1",
			"SELECT COALESCE(jsonb_agg(to_jsonb(r) ORDER BY to_jsonb(r)::text)::text,'[]') FROM indicator_observations r WHERE incident_id=$1",
			"SELECT COALESCE(jsonb_agg(to_jsonb(r) ORDER BY to_jsonb(r)::text)::text,'[]') FROM deployment_admin_audit_events r WHERE incident_id=$1",
		}
		for i, q := range queries {
			var arg any = incidentID
			if i == 0 {
				arg = incidentID.String()
			}
			if err := store.pool.QueryRow(ctx, q, arg).Scan(&values[i]); err != nil {
				t.Fatal(err)
			}
		}
		return values[:]
	}
	for _, postTransaction := range []bool{false, true} {
		for _, tc := range []struct {
			name, change, code  string
			status              int
			mismatch, malformed bool
		}{
			{"revoked membership", "membership", "incident_not_found", 404, false, false},
			{"downgraded role", "role", "authorization_denied", 403, false, false},
			{"closed incident", "closure", "incident_closed", 409, false, false},
			{"denial before hash mismatch", "role", "authorization_denied", 403, true, false},
			{"revocation before hash mismatch", "membership", "incident_not_found", 404, true, false},
			{"denial before corrupt receipt", "role", "authorization_denied", 403, false, true},
			{"authorized exact replay", "", "", status, false, false},
			{"authorized hash mismatch", "", "client_txn_conflict", 409, true, false},
			{"corrupt receipt", "", "internal_error", 500, false, true},
			{"hidden target", "target", "network_flow_indicator_link_forbidden", 403, false, false},
			{"deleted source table", "source", "", status, false, false},
		} {
			t.Run(fmt.Sprintf("replay %t %s", postTransaction, tc.name), func(t *testing.T) {
				// Each controlled mutation is committed before the post-lookup check.
				changed := false
				change := func() {
					changed = true
					switch tc.change {
					case "membership":
						exec("DELETE FROM incident_memberships WHERE incident_id=$1 AND user_id=$2", incidentID, actorID)
					case "role":
						exec("UPDATE incident_memberships SET role='reviewer' WHERE incident_id=$1 AND user_id=$2", incidentID, actorID)
					case "closure":
						exec("UPDATE incidents SET status='closed',closed_at=now() WHERE id=$1", incidentID)
					case "target":
						exec("UPDATE records SET deleted_at=now(),deleted_by_user_id=$2 WHERE record_id=$1", target, actorID)
					case "source":
						exec("UPDATE network_flow_tables SET table_status='soft_deleted',deleted_at=now() WHERE incident_id=$1 AND network_flow_table_id=$2", incidentID, request.Selector.TableID)
					}
				}
				defer func() {
					if changed {
						switch tc.change {
						case "membership":
							exec("INSERT INTO incident_memberships SELECT * FROM json_populate_record(NULL::incident_memberships,$1::json)", membership)
						case "role":
							exec("UPDATE incident_memberships SET role='admin' WHERE incident_id=$1 AND user_id=$2", incidentID, actorID)
						case "closure":
							exec("UPDATE incidents SET status='active',closed_at=NULL WHERE id=$1", incidentID)
						case "target":
							exec("UPDATE records SET deleted_at=NULL,deleted_by_user_id=NULL WHERE record_id=$1", target)
						case "source":
							exec("UPDATE network_flow_tables SET table_status='active',deleted_at=NULL WHERE incident_id=$1 AND network_flow_table_id=$2", incidentID, request.Selector.TableID)
						}
					}
					if tc.malformed {
						exec("UPDATE route_idempotency SET response_json=$3 WHERE actor_user_id=$1 AND client_txn_id=$2", actorID, request.ClientTxnID, original.ResponseJSON)
					}
				}()
				if tc.malformed {
					exec("UPDATE route_idempotency SET response_json='{}' WHERE actor_user_id=$1 AND client_txn_id=$2", actorID, request.ClientTxnID)
				}
				before := snapshot()
				checker := &linkReplayAdmissionProbe{incidentAdmissionChecker: admission.NewChecker(store.pool)}
				db := &linkReplayReceiptDB{DB: store.pool, hideFirst: postTransaction}
				backend := &linkReplayBackend{db: store.pool, err: errors.New("owner read failed before mutable checks")}
				if postTransaction {
					backend.change = change
				} else {
					checker.afterFirst = change
				}
				opts := options
				opts.Backend = backend
				coordinator, err := crossownertransaction.New(opts)
				if err != nil {
					t.Fatal(err)
				}
				app := indicatorLinkApplication{store: store, incidentAccess: checker, receipts: indicatorLinkReceiptAdapter{reader: authn.NewStore(db)}, transactions: coordinator, now: func() time.Time { return time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC) }}
				submitted := bytes.Clone(hash)
				if tc.mismatch {
					submitted[0] ^= 1
				}
				outcome, failure := app.execute(ctx, incidentID, authn.UserRecord{ID: actorID}, request, submitted, "replay-admission")
				if tc.code == "" {
					if failure != nil {
						t.Fatalf("replay failed: %v", failure)
					}
					payload, gotStatus := indicatorLinkOutcomeResponse(outcome)
					gotJSON, gotErr := json.Marshal(payload)
					wantJSON, wantErr := json.Marshal(expected)
					if gotErr != nil || wantErr != nil || gotStatus != tc.status || !bytes.Equal(gotJSON, wantJSON) {
						t.Fatal("authorized receipt changed")
					}
				} else {
					if failure == nil {
						t.Fatalf("returned receipt despite %s", tc.name)
					}
					api := semanticHTTPError(failure)
					if api.Status != tc.status || api.Code != tc.code || !reflect.DeepEqual(outcome, indicatorLinkOutcome{}) {
						t.Fatalf("replay verdict=%+v outcome=%+v", api, outcome)
					}
				}
				wantBegins := 0
				if postTransaction {
					wantBegins = 1
				}
				if backend.begins != wantBegins || backend.rollbacks != wantBegins || backend.commits != 0 || backend.writes != 0 {
					t.Fatalf("unexpected transaction effects: %+v", backend)
				}
				if !reflect.DeepEqual(before, snapshot()) {
					t.Fatal("replay changed durable receipts, bindings, indicators or audit")
				}
			})
		}
	}
	for _, state := range []string{"40P01", "40001"} {
		for _, rollbackTimeout := range []bool{false, true} {
			t.Run(fmt.Sprintf("admission conflict %s rollback timeout=%t", state, rollbackTimeout), func(t *testing.T) {
				before := snapshot()
				backend := &linkReplayBackend{db: store.pool, err: fmt.Errorf("owner read: %w", &pgconn.PgError{Code: state})}
				opts := options
				opts.Backend = backend
				var clock atomic.Int64
				if rollbackTimeout {
					opts.Clock = clock.Load
					opts.CancellationGrace = 100 * time.Millisecond
					backend.beforeRollback = func() { clock.Store(int64(opts.Timeout)) }
				}
				coordinator, err := crossownertransaction.New(opts)
				if err != nil {
					t.Fatal(err)
				}
				// Both lookups miss, so the application must translate the coordinator verdict.
				db := &linkReplayReceiptDB{DB: store.pool, hideAll: true}
				app := indicatorLinkApplication{store: store, incidentAccess: admission.NewChecker(store.pool), receipts: indicatorLinkReceiptAdapter{reader: authn.NewStore(db)}, transactions: coordinator, now: time.Now}
				outcome, failure := app.execute(ctx, incidentID, authn.UserRecord{ID: actorID}, request, hash, "conflict")
				if failure == nil {
					t.Fatal("conflict succeeded")
				}
				api := semanticHTTPError(failure)
				reason := "deadlock_detected"
				if state == "40001" {
					reason = "serialization_failure"
				}
				wantStatus, wantCode := 409, "transaction_conflict"
				if rollbackTimeout {
					wantStatus, wantCode, reason = 503, "service_unavailable", "extension_transaction_timeout"
				}
				if ctx.Err() != nil {
					t.Fatal("outer request context was interrupted")
				}
				if api.Status != wantStatus || api.Code != wantCode || !api.Retryable || failure.reason != reason || api.Details["reason_code"] != reason || !reflect.DeepEqual(outcome, indicatorLinkOutcome{}) {
					t.Fatalf("conflict verdict=%+v failure=%+v", api, failure)
				}
				if backend.begins != 1 || backend.rollbacks != 1 || backend.commits != 0 || backend.writes != 0 || !reflect.DeepEqual(before, snapshot()) {
					t.Fatal("conflict retained effects or retried")
				}
			})
		}
	}
}

type linkReplayAdmissionProbe struct {
	incidentAdmissionChecker
	calls      int
	afterFirst func()
}

func (p *linkReplayAdmissionProbe) Check(ctx context.Context, incident, actor uuid.UUID, requirement admission.Requirement) (admission.Grant, error) {
	grant, err := p.incidentAdmissionChecker.Check(ctx, incident, actor, requirement)
	p.calls++
	if p.calls == 1 && err == nil && p.afterFirst != nil {
		p.afterFirst()
	}
	return grant, err
}

type linkReplayReceiptDB struct {
	dbport.DB
	calls              int
	hideFirst, hideAll bool
}

func (db *linkReplayReceiptDB) QueryRow(ctx context.Context, q string, args ...any) pgx.Row {
	db.calls++
	if db.hideAll || db.hideFirst && db.calls == 1 {
		return missingLinkReceiptRow{}
	}
	return db.DB.QueryRow(ctx, q, args...)
}

type missingLinkReceiptRow struct{}

func (missingLinkReceiptRow) Scan(...any) error { return pgx.ErrNoRows }

type linkReplayBackend struct {
	db                                 dbport.DB
	err                                error
	change                             func()
	beforeRollback                     func()
	begins, commits, rollbacks, writes int
}

func (b *linkReplayBackend) Begin(ctx context.Context, _ []crossownertransaction.Descriptor) (crossownertransaction.Transaction, error) {
	b.begins++
	tx, err := b.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	return &linkReplayTransaction{Tx: tx, backend: b}, nil
}

type linkReplayTransaction struct {
	pgx.Tx
	backend *linkReplayBackend
}

func (*linkReplayTransaction) AcquireSerializationLock(context.Context, crossownertransaction.OrderedSerializationKey) error {
	return nil
}
func (t *linkReplayTransaction) ReadCapability(string) (crossownertransaction.ReadCapability, error) {
	return t, nil
}
func (t *linkReplayTransaction) WriteCapability(string) (crossownertransaction.WriteCapability, error) {
	t.backend.writes++
	return nil, crossownertransaction.ErrUnavailable
}
func (t *linkReplayTransaction) PreconditionCapability() (crossownertransaction.PreconditionCapability, error) {
	return t, nil
}
func (*linkReplayTransaction) FinalizationCapability() (crossownertransaction.FinalizationCapability, error) {
	return nil, crossownertransaction.ErrUnavailable
}
func (*linkReplayTransaction) ParticipantScope() string  { return IndicatorLinkParticipantID }
func (*linkReplayTransaction) PreconditionScope() string { return "shared.preconditions" }
func (t *linkReplayTransaction) IndicatorLinkAdmission() (crossownertransaction.ReadCapability, error) {
	return t, nil
}
func (t *linkReplayTransaction) CheckIndicatorLinkAdmission(context.Context, indicatorLinkMutation) error {
	if t.backend.change != nil {
		t.backend.change()
	}
	return t.backend.err
}
func (t *linkReplayTransaction) Commit(ctx context.Context) (crossownertransaction.CommitOutcome, error) {
	t.backend.commits++
	return crossownertransaction.CommitUnknown, t.Tx.Commit(ctx)
}
func (t *linkReplayTransaction) Rollback(ctx context.Context) (crossownertransaction.CommitOutcome, error) {
	t.backend.rollbacks++
	if t.backend.beforeRollback != nil {
		t.backend.beforeRollback()
	}
	err := t.Tx.Rollback(ctx)
	if err != nil {
		return crossownertransaction.CommitUnknown, err
	}
	return crossownertransaction.CommitAbsent, nil
}
