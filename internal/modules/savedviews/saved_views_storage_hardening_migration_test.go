package savedviews_test

import (
	"context"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestSavedViewsStorageHeadSchemaContract_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.OpenIsolatedDatabaseT(t, "saved-views-storage-head", postgres.PurposeRecovery)
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
