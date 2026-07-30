package storetest

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/mutationtest"
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

type rowScanner interface {
	Scan(dest ...any) error
}

type ReplayDatabase struct {
	queryRow  func(context.Context, string, ...any) rowScanner
	mutations mutationtest.Database
}

func SQLReplayDatabase(db *sql.DB) ReplayDatabase {
	return ReplayDatabase{
		queryRow: func(ctx context.Context, query string, args ...any) rowScanner {
			return db.QueryRowContext(ctx, query, args...)
		},
		mutations: mutationtest.SQLDatabase(db),
	}
}

func PostgresReplayDatabase(db postgres.DB) ReplayDatabase {
	return ReplayDatabase{
		queryRow: func(ctx context.Context, query string, args ...any) rowScanner {
			return db.QueryRow(ctx, query, args...)
		},
		mutations: mutationtest.PostgresDatabase(db),
	}
}

func CreateIncidentInStore(
	t testing.TB,
	store *incidents.Application,
	actor authn.UserRecord,
	request incidents.CreateIncidentRequest,
) incidents.CreateIncidentResult {
	t.Helper()

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

	store := incidents.NewApplication(pool)
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
	db ReplayDatabase,
	selector IncidentCreateReplaySelector,
) IncidentCreateReplaySideEffects {
	t.Helper()

	return IncidentCreateReplaySideEffects{
		BootstrapMembershipRows: queryReplayCount(
			t,
			db,
			`SELECT COUNT(*) FROM incident_memberships WHERE incident_id = $1 AND user_id = $2 AND role = 'admin'`,
			selector.IncidentID,
			selector.ActorUserID,
		),
		IncidentRows: queryReplayCount(
			t,
			db,
			`SELECT COUNT(*) FROM incidents WHERE id = $1`,
			selector.IncidentID,
		),
		IncidentWorkbookPreferenceRows: queryReplayCount(
			t,
			db,
			`SELECT COUNT(*) FROM incident_workbook_preferences WHERE incident_id = $1`,
			selector.IncidentID,
		),
		OwnerMutationRows: mutationtest.CountMutationArtifacts(
			t,
			db.mutations,
			mutationtest.MutationSelector{IncidentID: selector.IncidentID.String()},
			mutationtest.MutationOwnerIncidentResource,
			mutationtest.MutationOwnerIncidentMembership,
		),
		RouteIdempotencyRows: queryReplayCount(
			t,
			db,
			`SELECT COUNT(*) FROM route_idempotency WHERE route_key = 'incidents.create' AND actor_user_id::text = $1 AND scope_key = 'actor' AND client_txn_id = $2`,
			selector.ActorUserID.String(),
			selector.ClientTxnID,
		),
		UserWorkbookPreferenceRows: queryReplayCount(
			t,
			db,
			`SELECT COUNT(*) FROM user_workbook_preferences WHERE incident_id = $1 AND user_id = $2`,
			selector.IncidentID,
			selector.ActorUserID,
		),
	}
}

func queryReplayCount(t testing.TB, db ReplayDatabase, query string, args ...any) int {
	t.Helper()

	var count int
	if err := db.queryRow(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query replay count: %v", err)
	}
	return count
}

func QueryCount(t testing.TB, db postgres.DB, query string, args ...any) int {
	t.Helper()

	var count int
	if err := db.QueryRow(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}

func SeedMembership(
	t testing.TB,
	db any,
	incidentID uuid.UUID,
	userID uuid.UUID,
	displayName string,
	role string,
	addedByUserID uuid.UUID,
) {
	t.Helper()

	if err := execSeedDB(db, `
INSERT INTO incident_memberships (
    incident_id, user_id, role, joined_at, added_by_user_id,
    updated_at, updated_by_user_id, membership_version
)
VALUES ($1, $2, $3, now(), $4, now(), $4, 1)
ON CONFLICT (incident_id, user_id) DO UPDATE
SET role = EXCLUDED.role,
    updated_at = now(),
    updated_by_user_id = EXCLUDED.updated_by_user_id
`, incidentID, userID, role, addedByUserID); err != nil {
		t.Fatalf("seed incident membership: %v", err)
	}
	if err := execSeedDB(db, `
INSERT INTO user_workbook_preferences (incident_id, user_id, home_sheet_ref, created_at, updated_at)
VALUES ($1, $2, NULL, now(), now())
ON CONFLICT (incident_id, user_id) DO NOTHING
`, incidentID, userID); err != nil {
		t.Fatalf("seed user workbook preferences: %v", err)
	}
	_ = displayName
}

func execSeedDB(db any, query string, args ...any) error {
	switch typed := db.(type) {
	case postgres.DB:
		_, err := typed.Exec(context.Background(), query, args...)
		return err
	case *sql.DB:
		_, err := typed.ExecContext(context.Background(), query, args...)
		return err
	default:
		return fmt.Errorf("unsupported incident test database %T", db)
	}
}

func SeedIncidentMembershipSQL(t testing.TB, db *sql.DB, userID string, tag string) string {
	t.Helper()

	now := time.Now().UTC()
	incidentKey := strings.ToUpper(strings.ReplaceAll(tag, "_", "-") + "-" + uuid.NewString()[:8])
	var incidentID string
	if err := db.QueryRowContext(context.Background(), `
INSERT INTO incidents (
    incident_key, incident_key_canonical, title, status, created_by_user_id, created_at, updated_at, updated_by_user_id
)
VALUES ($1, $2, $3, 'active', $4, $5, $5, $4)
RETURNING id::text
`, incidentKey, strings.ToLower(incidentKey), "Auth socket "+tag, userID, now).Scan(&incidentID); err != nil {
		t.Fatalf("seed incident: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO incident_memberships (
    incident_id, user_id, role, joined_at, added_by_user_id, updated_at, updated_by_user_id, membership_version
)
VALUES ($1, $2, 'admin', $3, $2, $3, $2, 1)
`, incidentID, userID, now); err != nil {
		t.Fatalf("seed incident membership: %v", err)
	}
	return incidentID
}
