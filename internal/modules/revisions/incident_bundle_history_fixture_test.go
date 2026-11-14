package revisions_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
	"github.com/google/uuid"
)

// Explicit auxiliary creation facts keep focused non-row fixtures complete.
// Callers supply the full source; this helper never discovers or repairs rows.
func appendPortableCreationFixture(t testing.TB, bundle portabilitysupport.MapBundle, h revisionsPortabilityHarness, recordID uuid.UUID, recordType, schema string, source map[string]any) {
	t.Helper()
	record := clonePortableRow(t, h.snapshot["record"].(map[string]any))
	record["record_id"], record["record_type"], record["row_version"] = recordID.String(), recordType, 1
	snapshot := map[string]any{"snapshot_schema_id": schema, "record": record, "source": source}
	history := decodePortableRows(t, bundle[revisionsTestHistoryPath])
	history = append(history, map[string]any{"revision_id": len(history) + 1, "change_set_id": h.changeSet.String(), "record_id": recordID.String(), "row_version": 1, "before_json": nil, "after_json": snapshot, "created_at": h.createdAt.Format(time.RFC3339Nano)})
	bundle[revisionsTestHistoryPath] = encodePortableRows(t, history)
	targetKind := recordType
	if recordType == "timeline_event" {
		targetKind = "timeline_record"
	}
	mutations := decodePortableRows(t, bundle[revisionsTestMutationsPath])
	mutations = append(mutations, map[string]any{"change_set_id": h.changeSet.String(), "sequence_no": len(mutations) + 1, "target_kind": targetKind, "target_id": recordID.String(), "operation_kind": "create", "before_version_id": nil, "after_version_id": fmt.Sprintf("%s:%s:1", targetKind, recordID), "before_value": nil, "after_value": snapshot})
	bundle[revisionsTestMutationsPath] = encodePortableRows(t, mutations)
}

func portableIndicatorFixture(h revisionsPortabilityHarness, recordID uuid.UUID, value, dedupe string) map[string]any {
	return map[string]any{"record_id": recordID.String(), "incident_id": h.incidentID.String(), "indicator_type": "domain_name", "value_kind": "atomic", "display_value": value, "normalized_value": value, "dedupe_key": dedupe, "defanged_value": nil, "hash_algorithm": nil, "hash_value": nil, "stix_pattern": nil}
}
