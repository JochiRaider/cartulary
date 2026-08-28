package entities_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func newEntityTestTimelineBundle(t testing.TB, pool postgres.DB) *timelineassembly.Bundle {
	t.Helper()
	bundle, _ := newEntityTestTimelineComposition(t, pool)
	return bundle
}

func newEntityTestTimelineComposition(t testing.TB, pool postgres.DB) (*timelineassembly.Bundle, *projectionassembly.Runtime) {
	t.Helper()
	revisionComposition := revisionsupport.MustComposition(t)
	projections, err := projectionassembly.Build(pool)
	if err != nil {
		t.Fatalf("compose Projections: %v", err)
	}
	conflictTokens := conflicttest.NewCodec("timeline")
	evidenceOwner := appsupport.NewEvidenceOwnerRuntimeForTimeline(
		pool,
		conflictTokens,
		revisionComposition.Runtime.Appender(),
		revisionComposition.Publications,
		projections,
	)
	bundle, err := timelineassembly.NewBundle(timelineassembly.Dependencies{
		Postgres:            pool,
		ConflictTokens:      conflictTokens,
		ConflictFields:      revisionComposition.Runtime.ConflictFieldResolver(),
		Revisions:           revisionComposition.Runtime.Appender(),
		Collaboration:       revisionComposition.Publications,
		EvidenceAttachments: evidenceOwner.TimelineAttachmentContribution(),
		TimelineProjection:  projections.TimelinePorts().Writer,
		EntityProjection:    projections.EntityPorts().Writer,
		AssessmentRows:      projections.AssessmentPorts().Rows,
	})
	if err != nil {
		t.Fatalf("compose Timeline: %v", err)
	}
	return bundle, projections
}

func newEntityTestStore(t testing.TB, pool postgres.DB) *hostidentity.Store {
	t.Helper()
	revisionComposition := revisionsupport.MustComposition(t)
	projection, err := projectionassembly.Build(pool)
	if err != nil {
		t.Fatalf("compose Projections: %v", err)
	}
	store, err := hostidentity.NewStore(hostidentity.StoreDependencies{
		Postgres:             pool,
		Revisions:            revisionComposition.Runtime.Appender(),
		ProjectionWriter:     projection.EntityPorts().Writer,
		ProjectionReader:     projection.EntityPorts().Reader,
		KeepSavedIdempotency: workbookassembly.NewConflictIdempotencyPort(pool),
		Collaboration:        revisionComposition.Publications,
	})
	if err != nil {
		t.Fatalf("compose Host/Identity store: %v", err)
	}
	return store
}

func mustDefaultQueryMeta(t testing.TB, viewSchemaID string) viewschema.QueryMeta {
	t.Helper()
	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok {
		t.Fatalf("view schema %q not registered", viewSchemaID)
	}
	return schema.DefaultQueryMeta()
}

func requireTimelineMutationAfterSnapshotMatchesRecord(t testing.TB, db postgres.DB, changeSetID uuid.UUID, queryRow map[string]any) {
	t.Helper()

	var rawAfterValue []byte
	if err := db.QueryRow(context.Background(), `
SELECT after_value
  FROM change_set_mutations
 WHERE change_set_id = $1
   AND target_kind = 'timeline_record'
   AND operation_kind = 'patch'
 ORDER BY sequence_no DESC
 LIMIT 1
`, changeSetID).Scan(&rawAfterValue); err != nil {
		t.Fatalf("query timeline mutation after row: %v", err)
	}
	var mutationRow map[string]any
	if err := json.Unmarshal(rawAfterValue, &mutationRow); err != nil {
		t.Fatalf("decode timeline mutation after row: %v", err)
	}
	normalizedQueryRow := normalizeJSONMap(t, queryRow)
	if mutationRow["snapshot_schema_id"] != "cartulary.revisions.snapshot.timeline_event.v1" {
		t.Fatalf("timeline lifecycle mutation has unexpected snapshot schema: %#v", mutationRow)
	}
	record, recordOK := mutationRow["record"].(map[string]any)
	source, sourceOK := mutationRow["source"].(map[string]any)
	if !recordOK || !sourceOK || len(source) == 0 || record["record_id"] != normalizedQueryRow["record_id"] || record["row_version"] != normalizedQueryRow["row_version"] {
		t.Fatalf("timeline lifecycle snapshot does not match the queried record identity/version:\nsnapshot=%#v\nquery=%#v", mutationRow, normalizedQueryRow)
	}
}

func normalizeJSONMap(t testing.TB, value map[string]any) map[string]any {
	t.Helper()

	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("encode row for normalization: %v", err)
	}
	var normalized map[string]any
	if err := json.Unmarshal(raw, &normalized); err != nil {
		t.Fatalf("decode row for normalization: %v", err)
	}
	return normalized
}

// entity-storage / REQ-02-034, REQ-02-038, REQ-02-054..REQ-02-055 / AC-020, AC-021, AC-186.

func pgxTxOptions() pgx.TxOptions {
	return pgx.TxOptions{}
}

func requireEntityRow(t testing.TB, rows []map[string]any, recordID uuid.UUID) map[string]any {
	t.Helper()
	for _, row := range rows {
		if row["record_id"] == recordID.String() {
			return row
		}
	}
	t.Fatalf("expected row record_id=%s in %#v", recordID, rows)
	return nil
}

func collectionItemsFromEntityRow(t testing.TB, row map[string]any, fieldKey string) []map[string]any {
	t.Helper()
	cells := row["cells"].(map[string]any)
	cell := cells[fieldKey].(map[string]any)
	value := cell["value"].(map[string]any)
	switch rawItems := value["items"].(type) {
	case []map[string]any:
		items := make([]map[string]any, 0, len(rawItems))
		items = append(items, rawItems...)
		return items
	case []any:
		items := make([]map[string]any, 0, len(rawItems))
		for _, rawItem := range rawItems {
			items = append(items, rawItem.(map[string]any))
		}
		return items
	default:
		t.Fatalf("unexpected collection item payload for %s: %#v", fieldKey, value["items"])
	}
	return nil
}

func requireReusableIdentifierItem(t testing.TB, row map[string]any, fieldKey string, identifierClass string, rawValue string) map[string]any {
	t.Helper()
	normalized, ok := fieldnorm.NormalizeIdentifier(identifierClass, rawValue)
	if !ok {
		t.Fatalf("test raw value %q does not normalize for %s", rawValue, identifierClass)
	}
	for _, item := range collectionItemsFromEntityRow(t, row, fieldKey) {
		if item["identifier_class"] == identifierClass && item["normalized_value"] == normalized {
			if item["item_kind"] != "reusable_identifier" {
				t.Fatalf("expected reusable identifier item kind, got %#v", item)
			}
			if item["raw_value"] != rawValue {
				t.Fatalf("expected raw_value=%q, got %#v", rawValue, item)
			}
			if item["item_ref"] == "" {
				t.Fatalf("expected reusable identifier item_ref, got %#v", item)
			}
			return item
		}
	}
	t.Fatalf("expected reusable identifier class=%s normalized=%s in %s, got %#v", identifierClass, normalized, fieldKey, collectionItemsFromEntityRow(t, row, fieldKey))
	return nil
}

func requireNoReusableIdentifierItem(t testing.TB, row map[string]any, fieldKey string, identifierClass string, rawValue string) {
	t.Helper()
	normalized, ok := fieldnorm.NormalizeIdentifier(identifierClass, rawValue)
	if !ok {
		t.Fatalf("test raw value %q does not normalize for %s", rawValue, identifierClass)
	}
	for _, item := range collectionItemsFromEntityRow(t, row, fieldKey) {
		if item["identifier_class"] == identifierClass && item["normalized_value"] == normalized {
			t.Fatalf("did not expect reusable identifier class=%s normalized=%s in %s, got %#v", identifierClass, normalized, fieldKey, item)
		}
	}
}
