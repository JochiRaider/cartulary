package phase2test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type IncidentCreateReplaySelector struct {
	ActorUserID uuid.UUID
	ClientTxnID string
	IncidentID  uuid.UUID
}

type IncidentCreateReplaySideEffects struct {
	BootstrapMembershipRows        int
	IncidentRows                   int
	IncidentWorkbookPreferenceRows int
	OwnerMutationRows              int
	RouteIdempotencyRows           int
	UserWorkbookPreferenceRows     int
}

func CreateIncidentInStore(
	t testing.TB,
	pool postgres.DB,
	actor authn.UserRecord,
	request incidents.CreateIncidentRequest,
) incidents.CreateIncidentResult {
	t.Helper()

	store := incidents.NewStore(pool)
	result, err := store.CreateIncident(
		context.Background(),
		actor,
		request,
		incidents.IncidentCreateRequestHash(request),
		"req-"+request.ClientTxnID,
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("create incident in store: %v", err)
	}
	return result
}

func CreateMembershipInStore(
	t testing.TB,
	pool postgres.DB,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	targetUser authn.UserRecord,
	request incidents.MembershipCreateRequest,
) incidents.MembershipCreateResult {
	t.Helper()

	store := incidents.NewStore(pool)
	result, err := store.CreateMembership(
		context.Background(),
		actor,
		incidentID,
		targetUser,
		request,
		[]byte(request.ClientTxnID),
		"req-"+request.ClientTxnID,
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("create membership in store: %v", err)
	}
	return result
}

func LookupUserByEmail(t testing.TB, pool postgres.DB, email string) authn.UserRecord {
	t.Helper()

	store := authn.NewStore(pool)
	record, err := store.GetUserByNormalizedEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("lookup user by email %q: %v", email, err)
	}
	return record
}

func LookupUserByID(t testing.TB, pool postgres.DB, userID uuid.UUID) authn.UserRecord {
	t.Helper()

	store := authn.NewStore(pool)
	record, err := store.GetUserByID(context.Background(), userID)
	if err != nil {
		t.Fatalf("lookup user by id %s: %v", userID, err)
	}
	return record
}

func SnapshotIncidentCreateReplaySideEffects(
	t testing.TB,
	db *sql.DB,
	selector IncidentCreateReplaySelector,
) IncidentCreateReplaySideEffects {
	t.Helper()

	return IncidentCreateReplaySideEffects{
		BootstrapMembershipRows: QueryCount(
			t,
			db,
			`SELECT COUNT(*) FROM incident_memberships WHERE incident_id = $1 AND user_id = $2 AND role = 'admin'`,
			selector.IncidentID,
			selector.ActorUserID,
		),
		IncidentRows: QueryCount(
			t,
			db,
			`SELECT COUNT(*) FROM incidents WHERE id = $1`,
			selector.IncidentID,
		),
		IncidentWorkbookPreferenceRows: QueryCount(
			t,
			db,
			`SELECT COUNT(*) FROM incident_workbook_preferences WHERE incident_id = $1`,
			selector.IncidentID,
		),
		OwnerMutationRows: CountMutationArtifacts(
			t,
			db,
			MutationSelector{IncidentID: selector.IncidentID.String()},
			MutationOwnerIncidentResource,
			MutationOwnerIncidentMembership,
		),
		RouteIdempotencyRows: QueryCount(
			t,
			db,
			`SELECT COUNT(*) FROM route_idempotency WHERE route_key = 'incidents.create' AND actor_user_id::text = $1 AND scope_key = 'actor' AND client_txn_id = $2`,
			selector.ActorUserID.String(),
			selector.ClientTxnID,
		),
		UserWorkbookPreferenceRows: QueryCount(
			t,
			db,
			`SELECT COUNT(*) FROM user_workbook_preferences WHERE incident_id = $1 AND user_id = $2`,
			selector.IncidentID,
			selector.ActorUserID,
		),
	}
}
