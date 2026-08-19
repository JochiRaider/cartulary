package sourceboundary_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/sourceboundary"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestSourceBoundaryV1CanonicalBytesAndSelection_Integration(t *testing.T) {
	ctx := context.Background()
	revisionsResolver := sourceboundary.NewResolver()
	testDB := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "reporting-source-boundary-characterization")
	conn, err := pgx.Connect(ctx, testDB.DSN)
	if err != nil {
		t.Fatalf("connect source-boundary database: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close(context.Background()) })
	writerConn, err := pgx.Connect(ctx, testDB.DSN)
	if err != nil {
		t.Fatalf("connect source-boundary writer: %v", err)
	}
	t.Cleanup(func() { _ = writerConn.Close(context.Background()) })

	actorID := uuid.MustParse("77777777-7777-4777-8777-777777777777")
	if _, err := conn.Exec(ctx, `
INSERT INTO users (id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, 'source-boundary@example.test', 'Source Boundary', 'test-only', false, true, false)
`, actorID); err != nil {
		t.Fatalf("seed actor: %v", err)
	}

	type visibleChangeSet struct {
		id        uuid.UUID
		createdAt time.Time
	}
	testCases := []struct {
		name          string
		incidentID    uuid.UUID
		version       int64
		changeSets    []visibleChangeSet
		canonicalJSON string
		token         string
	}{
		{
			name:          "zero change sets",
			incidentID:    uuid.MustParse("11111111-1111-4111-8111-111111111111"),
			version:       1,
			canonicalJSON: `{"incident_id":"11111111-1111-4111-8111-111111111111","incident_version":1,"latest_change_set_id":null,"latest_change_set_created_at":null}`,
			token:         "cartulary.source_boundary.v1:7118b27f199de626a091552abc20fb026c5f25f04ca303e0edc790f86ee82478",
		},
		{
			name:       "one change set uses UTC RFC3339Nano",
			incidentID: uuid.MustParse("11111111-1111-4111-8111-111111111112"),
			version:    7,
			changeSets: []visibleChangeSet{
				{
					id:        uuid.MustParse("22222222-2222-4222-8222-222222222222"),
					createdAt: time.Date(2026, 8, 18, 16, 1, 2, 345678900, time.FixedZone("EDT", -4*60*60)),
				},
			},
			canonicalJSON: `{"incident_id":"11111111-1111-4111-8111-111111111112","incident_version":7,"latest_change_set_id":"22222222-2222-4222-8222-222222222222","latest_change_set_created_at":"2026-08-18T20:01:02.345678Z"}`,
			token:         "cartulary.source_boundary.v1:71316dc602a4af569f4e887d5939c01344a2a5ad5f8d47249d5d7368a4555499",
		},
		{
			name:       "many change sets select latest timestamp",
			incidentID: uuid.MustParse("11111111-1111-4111-8111-111111111113"),
			version:    9,
			changeSets: []visibleChangeSet{
				{id: uuid.MustParse("33333333-3333-4333-8333-333333333333"), createdAt: time.Date(2026, 8, 18, 20, 1, 2, 0, time.UTC)},
				{id: uuid.MustParse("44444444-4444-4444-8444-444444444444"), createdAt: time.Date(2026, 8, 18, 20, 1, 3, 0, time.UTC)},
			},
			canonicalJSON: `{"incident_id":"11111111-1111-4111-8111-111111111113","incident_version":9,"latest_change_set_id":"44444444-4444-4444-8444-444444444444","latest_change_set_created_at":"2026-08-18T20:01:03Z"}`,
			token:         "cartulary.source_boundary.v1:c354bb52fa314a05e08e5219fa161a86e379c53f6046ec4df7e45df0c17c9c69",
		},
		{
			name:       "timestamp tie selects descending change set UUID",
			incidentID: uuid.MustParse("11111111-1111-4111-8111-111111111114"),
			version:    10,
			changeSets: []visibleChangeSet{
				{id: uuid.MustParse("55555555-5555-4555-8555-555555555555"), createdAt: time.Date(2026, 8, 18, 20, 1, 4, 0, time.UTC)},
				{id: uuid.MustParse("66666666-6666-4666-8666-666666666666"), createdAt: time.Date(2026, 8, 18, 20, 1, 4, 0, time.UTC)},
			},
			canonicalJSON: `{"incident_id":"11111111-1111-4111-8111-111111111114","incident_version":10,"latest_change_set_id":"66666666-6666-4666-8666-666666666666","latest_change_set_created_at":"2026-08-18T20:01:04Z"}`,
			token:         "cartulary.source_boundary.v1:2f9e885772349eef223bb79e1886ee8dcf67308b835f1f1d714303c2635457dc",
		},
	}

	for index, testCase := range testCases {
		if _, err := conn.Exec(ctx, `
INSERT INTO incidents (
    id, incident_key, incident_key_canonical, title, status,
    created_by_user_id, updated_by_user_id, incident_version
) VALUES ($1, $2, $2, $3, 'active', $4, $4, $5)
`, testCase.incidentID, "IR-SOURCE-BOUNDARY-"+string(rune('A'+index)), testCase.name, actorID, testCase.version); err != nil {
			t.Fatalf("seed %s incident: %v", testCase.name, err)
		}
		for _, changeSet := range testCase.changeSets {
			if _, err := conn.Exec(ctx, `
INSERT INTO change_sets (change_set_id, incident_id, actor_user_id, source, created_at)
VALUES ($1, $2, $3, 'source_boundary_characterization', $4)
`, changeSet.id, testCase.incidentID, actorID, changeSet.createdAt); err != nil {
				t.Fatalf("seed %s change set: %v", testCase.name, err)
			}
		}
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
			if err != nil {
				t.Fatalf("begin repeatable-read transaction: %v", err)
			}
			defer func() { _ = tx.Rollback(context.Background()) }()

			boundary, err := revisionsResolver.ResolveCurrentTx(ctx, tx, sourceboundary.ResolveInput{
				IncidentID:      testCase.incidentID,
				IncidentVersion: testCase.version,
			})
			if err != nil {
				t.Fatalf("resolve Revisions-owned source boundary: %v", err)
			}
			if got := string(boundary.CanonicalJSON); got != testCase.canonicalJSON {
				t.Fatalf("canonical JSON = %q, want %q", got, testCase.canonicalJSON)
			}
			if boundary.Token != testCase.token {
				t.Fatalf("token = %q, want %q", boundary.Token, testCase.token)
			}
		})
	}

	t.Run("repeatable-read visibility is stable", func(t *testing.T) {
		testCase := testCases[2]
		tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
		if err != nil {
			t.Fatalf("begin repeatable-read transaction: %v", err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()
		before, err := revisionsResolver.ResolveCurrentTx(ctx, tx, sourceboundary.ResolveInput{
			IncidentID:      testCase.incidentID,
			IncidentVersion: testCase.version,
		})
		if err != nil {
			t.Fatalf("resolve initial boundary: %v", err)
		}

		laterID := uuid.MustParse("99999999-9999-4999-8999-999999999999")
		if _, err := writerConn.Exec(ctx, `
INSERT INTO change_sets (change_set_id, incident_id, actor_user_id, source, created_at)
VALUES ($1, $2, $3, 'source_boundary_concurrent_writer', $4)
`, laterID, testCase.incidentID, actorID, time.Date(2026, 8, 18, 20, 1, 5, 0, time.UTC)); err != nil {
			t.Fatalf("insert concurrent change set: %v", err)
		}

		after, err := revisionsResolver.ResolveCurrentTx(ctx, tx, sourceboundary.ResolveInput{
			IncidentID:      testCase.incidentID,
			IncidentVersion: testCase.version,
		})
		if err != nil {
			t.Fatalf("resolve repeated boundary: %v", err)
		}
		if string(after.CanonicalJSON) != string(before.CanonicalJSON) || after.Token != before.Token {
			t.Fatalf("repeatable-read boundary changed: before=%q/%q after=%q/%q", before.CanonicalJSON, before.Token, after.CanonicalJSON, after.Token)
		}
	})
}
