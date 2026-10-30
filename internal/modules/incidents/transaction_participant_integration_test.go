package incidents_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestIncidentParticipantSerializesWritesWithoutBlockingReferences_Integration(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "incident-participant", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "txn-participant", "incident_key": "IR-PARTICIPANT", "title": "Participant"})
	id := uuid.MustParse(incident["incident_id"].(string))
	owner, err := h.Pool.BeginTx(t.Context(), pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Rollback(context.Background())
	if found, err := incidents.NewTransactionParticipant().LockIncidentTx(t.Context(), owner, id); err != nil || !found {
		t.Fatalf("owner lock: %t %v", found, err)
	}
	// A foreign-key check takes KEY SHARE. It must coexist with the owner lock;
	// other writers and lifecycle deletion must still wait for this transaction.
	for _, mode := range []string{"KEY SHARE", "NO KEY UPDATE", "UPDATE"} {
		t.Run(mode, func(t *testing.T) {
			tx, err := h.Pool.BeginTx(t.Context(), pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(context.Background())
			var got uuid.UUID
			err = tx.QueryRow(t.Context(), `SELECT id FROM incidents WHERE id=$1 FOR `+mode+` NOWAIT`, id).Scan(&got)
			if mode == "KEY SHARE" {
				if err != nil || got != id {
					t.Fatalf("reference blocked: %v", err)
				}
				return
			}
			var conflict *pgconn.PgError
			if !errors.As(err, &conflict) || conflict.Code != "55P03" {
				t.Fatalf("concurrent writer admitted: %v", err)
			}
		})
	}
	if err := owner.Rollback(t.Context()); err != nil {
		t.Fatal(err)
	}
	next, err := h.Pool.BeginTx(t.Context(), pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer next.Rollback(context.Background())
	if found, err := incidents.NewTransactionParticipant().LockIncidentTx(t.Context(), next, id); err != nil || !found {
		t.Fatalf("settled lock not reusable: %t %v", found, err)
	}
}
