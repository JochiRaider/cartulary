package savedviews_test

import (
	"context"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestSavedViewsStorageHardeningMigration52FreshSchema_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	migrationDB := harness.MigrationDatabaseThroughT(t, "saved-views-storage-hardening-fresh", 52)
	db := migrationDB.SQL()
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `SET session_replication_role = replica`); err != nil {
		t.Fatal(err)
	}

	cases := map[string]struct {
		id         string
		incidentID string
		scope      string
		owner      any
		version    int
		createdAt  string
		updatedAt  string
		constraint string
	}{
		"private owner missing": {
			id:         "52000000-0000-4000-8000-000000000001",
			incidentID: "52000000-0000-4000-8000-000000000011",
			scope:      "private", version: 1,
			createdAt: "2026-07-29T09:00:00Z", updatedAt: "2026-07-29T09:00:00Z",
			constraint: "saved_views_owner_scope_ck",
		},
		"system owner present": {
			id:         "52000000-0000-4000-8000-000000000002",
			incidentID: "52000000-0000-4000-8000-000000000012",
			scope:      "system", owner: "52000000-0000-4000-8000-000000000099", version: 1,
			createdAt: "2026-07-29T09:00:00Z", updatedAt: "2026-07-29T09:00:00Z",
			constraint: "saved_views_owner_scope_ck",
		},
		"version not positive": {
			id:         "52000000-0000-4000-8000-000000000003",
			incidentID: "52000000-0000-4000-8000-000000000013",
			scope:      "private", owner: "52000000-0000-4000-8000-000000000099", version: 0,
			createdAt: "2026-07-29T09:00:00Z", updatedAt: "2026-07-29T09:00:00Z",
			constraint: "saved_views_version_positive_ck",
		},
		"timestamps reversed": {
			id:         "52000000-0000-4000-8000-000000000004",
			incidentID: "52000000-0000-4000-8000-000000000014",
			scope:      "private", owner: "52000000-0000-4000-8000-000000000099", version: 1,
			createdAt: "2026-07-29T09:00:01Z", updatedAt: "2026-07-29T09:00:00Z",
			constraint: "saved_views_timestamp_order_ck",
		},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := db.ExecContext(ctx, `
INSERT INTO saved_views (
    saved_view_id, incident_id, view_schema_id, scope, display_name,
    query_json, layout_json, owner_user_id, saved_view_version,
    created_at, updated_at
)
VALUES (
$1, $2, 'cartulary.view.timeline.v2', $3, 'Migration constraint fixture',
'{"filters":[],"sort":[]}'::jsonb,
'{"column_order":[],"column_widths":[],"hidden_field_keys":[],"layout_schema_id":"cartulary.layout.v1"}'::jsonb,
$4, $5, $6, $7
)
`,
				testCase.id,
				testCase.incidentID,
				testCase.scope,
				testCase.owner,
				testCase.version,
				testCase.createdAt,
				testCase.updatedAt,
			)
			if err == nil || !strings.Contains(err.Error(), testCase.constraint) {
				t.Fatalf("invalid row error = %v; want constraint %s", err, testCase.constraint)
			}
		})
	}
}

func TestSavedViewsStorageHardeningMigration52PreflightReportsOnlyCounts_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	migrationDB := harness.MigrationDatabaseThroughT(t, "saved-views-storage-hardening-preflight", 51)
	db := migrationDB.SQL()
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `SET session_replication_role = replica`); err != nil {
		t.Fatal(err)
	}
	seededIDs := []string{
		"52000000-0000-4000-8000-000000000101",
		"52000000-0000-4000-8000-000000000102",
		"52000000-0000-4000-8000-000000000103",
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO saved_views (
    saved_view_id, incident_id, view_schema_id, scope, display_name,
    query_json, layout_json, owner_user_id, saved_view_version,
    created_at, updated_at
)
VALUES
(
    $1, '52000000-0000-4000-8000-000000000201',
    'cartulary.view.timeline.v2', 'system', 'Invalid owner tuple',
    '{"filters":[],"sort":[]}'::jsonb,
    '{"column_order":[],"column_widths":[],"hidden_field_keys":[],"layout_schema_id":"cartulary.layout.v1"}'::jsonb,
    '52000000-0000-4000-8000-000000000299', 1,
    '2026-07-29T09:00:00Z', '2026-07-29T09:00:00Z'
),
(
    $2, '52000000-0000-4000-8000-000000000202',
    'cartulary.view.timeline.v2', 'private', 'Invalid version',
    '{"filters":[],"sort":[]}'::jsonb,
    '{"column_order":[],"column_widths":[],"hidden_field_keys":[],"layout_schema_id":"cartulary.layout.v1"}'::jsonb,
    '52000000-0000-4000-8000-000000000299', 0,
    '2026-07-29T09:00:00Z', '2026-07-29T09:00:00Z'
),
(
    $3, '52000000-0000-4000-8000-000000000203',
    'cartulary.view.timeline.v2', 'private', 'Invalid timestamps',
    '{"filters":[],"sort":[]}'::jsonb,
    '{"column_order":[],"column_widths":[],"hidden_field_keys":[],"layout_schema_id":"cartulary.layout.v1"}'::jsonb,
    '52000000-0000-4000-8000-000000000299', 1,
    '2026-07-29T09:00:01Z', '2026-07-29T09:00:00Z'
)
`, seededIDs[0], seededIDs[1], seededIDs[2]); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `SET session_replication_role = origin`); err != nil {
		t.Fatal(err)
	}

	err := migrationDB.ApplyThrough(ctx, 52)
	if err == nil {
		t.Fatal("expected saved-view storage-hardening preflight rejection")
	}
	message := err.Error()
	for _, expected := range []string{
		"saved views storage hardening preflight failed",
		"invalid_owner_scope_count=1",
		"invalid_version_count=1",
		"invalid_timestamp_count=1",
	} {
		if !strings.Contains(message, expected) {
			t.Fatalf("preflight error missing %q: %v", expected, err)
		}
	}
	for _, seededID := range seededIDs {
		if strings.Contains(message, seededID) {
			t.Fatalf("preflight disclosed invalid row identity %s: %v", seededID, err)
		}
	}

	var strengthenedConstraints int
	if err := db.QueryRowContext(ctx, `
SELECT count(*)
  FROM pg_constraint
 WHERE conrelid = 'saved_views'::regclass
   AND conname IN (
       'saved_views_version_positive_ck',
       'saved_views_timestamp_order_ck'
   )
`).Scan(&strengthenedConstraints); err != nil {
		t.Fatal(err)
	}
	if strengthenedConstraints != 0 {
		t.Fatalf("failed preflight partially added %d constraints", strengthenedConstraints)
	}
}
