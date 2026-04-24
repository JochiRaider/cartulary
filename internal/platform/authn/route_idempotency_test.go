package authn_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestRouteIdempotencyActorScopedUniqueness(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareDatabaseT(t, "route-idempotency-actor-scope")

	pool, err := pgxpool.New(context.Background(), testDB.DSN)
	if err != nil {
		t.Fatalf("open pgx pool: %v", err)
	}
	t.Cleanup(pool.Close)

	ctx := context.Background()
	actorA := uuid.MustParse("10000000-0000-0000-0000-000000010001")
	actorB := uuid.MustParse("10000000-0000-0000-0000-000000010002")
	for _, actorID := range []uuid.UUID{actorA, actorB} {
		if _, err := pool.Exec(ctx, `
INSERT INTO users (id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, $3, 'hash', false, true, false)
`, actorID, actorID.String()+"@example.test", actorID.String()); err != nil {
			t.Fatalf("insert actor user %s: %v", actorID, err)
		}
	}

	insert := func(t testing.TB, actorID uuid.UUID, payload map[string]any) error {
		t.Helper()

		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer func() {
			_ = tx.Rollback(ctx)
		}()

		key := authn.RouteIdempotencyKey{
			RouteKey:    "test.route",
			ActorUserID: actorID,
			ScopeKey:    "shared-scope",
			ClientTxnID: "txn-shared",
		}
		if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, []byte("request"), 200, payload); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}

	if err := insert(t, actorA, map[string]any{"actor": "a"}); err != nil {
		t.Fatalf("insert actor A idempotency: %v", err)
	}
	if err := insert(t, actorB, map[string]any{"actor": "b"}); err != nil {
		t.Fatalf("cross-actor duplicate key should be allowed: %v", err)
	}
	if err := insert(t, actorA, map[string]any{"actor": "a-again"}); !authn.IsUniqueViolation(err) {
		t.Fatalf("same actor duplicate key should violate uniqueness, got %v", err)
	}

	store := authn.NewStore(pool)
	for _, tc := range []struct {
		actorID uuid.UUID
	}{
		{actorID: actorA},
		{actorID: actorB},
	} {
		record, err := store.GetRouteIdempotency(ctx, authn.RouteIdempotencyKey{
			RouteKey:    "test.route",
			ActorUserID: tc.actorID,
			ScopeKey:    "shared-scope",
			ClientTxnID: "txn-shared",
		})
		if err != nil {
			t.Fatalf("lookup actor %s route idempotency: %v", tc.actorID, err)
		}
		if record.ActorUserID != tc.actorID {
			t.Fatalf("lookup returned actor %s, want %s", record.ActorUserID, tc.actorID)
		}
	}
}
