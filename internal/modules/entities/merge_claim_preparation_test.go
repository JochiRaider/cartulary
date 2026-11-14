package entities_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func exerciseMergeClaimPreparationDrift(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "merge-claim-preparation", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actor := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	for _, kind := range []string{"host", "identity"} {
		for _, side := range []string{"survivor", "loser"} {
			t.Run(kind+" "+side, func(t *testing.T) {
				incidentText := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "claim-drift-" + kind + side, "incident_key": "IR-DRIFT-" + kind + side, "title": "Claim preparation drift"})["incident_id"].(string)
				incident := uuid.MustParse(incidentText)
				survivor, loser := uuid.New(), uuid.New()
				oldValue := func(id uuid.UUID) string {
					if kind == "host" {
						return "old-" + id.String()
					}
					return "old-" + id.String() + "@example.test"
				}
				for _, id := range []uuid.UUID{survivor, loser} {
					if kind == "host" {
						entitytest.SeedHostRecord(t, h.DB, incident, actor, id, "Claim participant", oldValue(id), "", "")
					} else {
						entitytest.SeedIdentityRecord(t, h.DB, incident, actor, id, "Claim participant", oldValue(id), id.String()+"@example.test", id.String())
					}
				}
				changed, survivorVersion, loserVersion := survivor, 2, 1
				if side == "loser" {
					changed, survivorVersion, loserVersion = loser, 1, 2
				}
				ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
				defer cancel()
				holder, err := h.Pool.Begin(ctx)
				if err != nil {
					t.Fatal(err)
				}
				defer holder.Rollback(context.Background())
				identifierType := "hostname"
				if kind == "identity" {
					identifierType = "upn"
				}
				// Hold the old canonical claim while committing an independently staged
				// source update. These are admission fixtures, not portable histories.
				key := incident.String() + "\x1f" + kind + "\x1f" + identifierType + "\x1f" + oldValue(changed)
				if _, err := holder.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, key); err != nil {
					t.Fatal(err)
				}
				if _, err := holder.Exec(ctx, `UPDATE records SET row_version=2 WHERE record_id=$1`, changed); err != nil {
					t.Fatal(err)
				}
				newValue := "new-" + changed.String()
				query := `UPDATE hosts SET hostname=$2,row_version=2 WHERE record_id=$1`
				if kind == "identity" {
					newValue += "@example.test"
					query = `UPDATE identities SET upn=$2,row_version=2 WHERE record_id=$1`
				}
				if _, err := holder.Exec(ctx, query, changed, newValue); err != nil {
					t.Fatal(err)
				}
				snapshot := func(q interface {
					QueryRow(context.Context, string, ...any) pgx.Row
				}) [32]byte {
					var content string
					if err := q.QueryRow(ctx, `SELECT jsonb_build_object(
 'records',(SELECT jsonb_agg(to_jsonb(x) ORDER BY record_id) FROM records x WHERE incident_id=$1),
 'hosts',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY record_id),'[]') FROM hosts x WHERE incident_id=$1),
 'identities',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY record_id),'[]') FROM identities x WHERE incident_id=$1),
 'claims',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY to_jsonb(x)::text),'[]') FROM entity_active_identifier_claims x WHERE incident_id=$1),
 'aliases',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY entity_alias_id),'[]') FROM entity_aliases x WHERE incident_id=$1),
 'identifiers',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY entity_preserved_identifier_id),'[]') FROM entity_preserved_identifiers x WHERE incident_id=$1),
 'changes',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY change_set_id),'[]') FROM change_sets x WHERE incident_id=$1),
 'mutations',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY x.change_set_id,sequence_no),'[]') FROM change_set_mutations x JOIN change_sets c USING(change_set_id) WHERE c.incident_id=$1),
 'revisions',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY to_jsonb(x)::text),'[]') FROM record_revisions x JOIN records r USING(record_id) WHERE r.incident_id=$1),
 'receipts',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY id),'[]') FROM route_idempotency x WHERE split_part(scope_key,':',1)=$1::text OR split_part(scope_key,':',1) IN (SELECT record_id::text FROM records WHERE incident_id=$1)),
 'hosts_projection',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY record_id),'[]') FROM host_grid_projection x WHERE incident_id=$1),
 'identities_projection',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY record_id),'[]') FROM identity_grid_projection x WHERE incident_id=$1))::text`, incident).Scan(&content); err != nil {
						t.Fatal(err)
					}
					return sha256.Sum256([]byte(content))
				}
				// Capture the exact expected committed fixture state inside its transaction,
				// before Merge can finish; a post-request snapshot is not its own oracle.
				wantState := snapshot(holder)
				wantPublication := collaborationsupport.SnapshotIntentContent(t, h.DB, incident)
				body, err := json.Marshal(map[string]any{"loser_record_id": loser.String(), "survivor_base_row_version": survivorVersion, "loser_base_row_version": loserVersion, "client_txn_id": "claim-drift-" + changed.String()})
				if err != nil {
					t.Fatal(err)
				}
				client := *http.DefaultClient
				client.Timeout = 10 * time.Second
				send := func() (*http.Response, error) {
					req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.Server.HTTP.URL+"/api/v1/records/"+survivor.String()+"/merge", bytes.NewReader(body))
					if err != nil {
						return nil, err
					}
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set(authn.CSRFHeaderName, login.CSRF.Value)
					req.AddCookie(login.Session)
					req.AddCookie(login.CSRF)
					return client.Do(req)
				}
				type response struct {
					value *http.Response
					err   error
				}
				ready := make(chan response, 1)
				go func() { value, err := send(); ready <- response{value, err} }()
				for {
					var waiting string
					if err := h.Pool.QueryRow(ctx, `SELECT coalesce((SELECT locktype FROM pg_locks WHERE NOT granted AND $1::integer=ANY(pg_blocking_pids(pid)) ORDER BY locktype LIMIT 1),'')`, int32(holder.Conn().PgConn().PID())).Scan(&waiting); err != nil {
						t.Fatal(err)
					}
					if waiting != "" {
						if waiting != "advisory" {
							t.Fatalf("Merge waited on %s before claim preparation", waiting)
						}
						break
					}
					select {
					case early := <-ready:
						if early.value != nil {
							early.value.Body.Close()
						}
						t.Fatalf("Merge completed before claim protection: %v", early.err)
					case <-ctx.Done():
						t.Fatal(ctx.Err())
					case <-time.After(10 * time.Millisecond):
					}
				}
				if err := holder.Commit(ctx); err != nil {
					t.Fatal(err)
				}
				var first response
				select {
				case first = <-ready:
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				}
				if first.err != nil {
					t.Fatal(first.err)
				}
				failure := httptestx.RequireErrorEnvelope(t, first.value, http.StatusConflict, "merge_precondition_failed")["error"].(map[string]any)
				if !reflect.DeepEqual(failure["details"], map[string]any{"reason_code": "protected_set_changed"}) {
					t.Fatalf("claim drift=%#v", failure)
				}
				if snapshot(h.Pool) != wantState || collaborationsupport.SnapshotIntentContent(t, h.DB, incident) != wantPublication {
					t.Fatal("claim drift changed state or consumed transaction identity")
				}
				accepted, err := send()
				if err != nil {
					t.Fatal(err)
				}
				data := httptestx.RequireSuccessData(t, accepted, http.StatusOK)
				var versions bool
				if err := h.Pool.QueryRow(ctx, `SELECT (SELECT row_version=$2 FROM records WHERE record_id=$1) AND (SELECT row_version=$4 FROM records WHERE record_id=$3)`, survivor, survivorVersion+1, loser, loserVersion+1).Scan(&versions); err != nil {
					t.Fatal(err)
				}
				if !versions {
					t.Fatal("fresh preparation did not advance each participant once")
				}
				after, publication := snapshot(h.Pool), collaborationsupport.SnapshotIntentContent(t, h.DB, incident)
				replayed, err := send()
				if err != nil {
					t.Fatal(err)
				}
				if replay := httptestx.RequireSuccessData(t, replayed, http.StatusOK); !reflect.DeepEqual(data, replay) || snapshot(h.Pool) != after || collaborationsupport.SnapshotIntentContent(t, h.DB, incident) != publication {
					t.Fatal("prepared Merge replay changed payload or effects")
				}
			})
		}
	}
}
