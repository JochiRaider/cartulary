import type { RecordHistoryItem } from "../adapters/workbookHistoryResponse";

/** A complete semantic event for tests concerned with browsing or operations. */
export function historyDiffFixture(
  summary: string,
): RecordHistoryItem["diff_summary"] {
  return {
    schema_id: "cartulary.history_diff.v1",
    summary,
    units: [
      {
        unit_ref: "hunit_fixture_field",
        kind: "field",
        operation: "update",
        record_ids: ["20000000-0000-4000-8000-000000000001"],
        changes: [
          {
            field_key: "timeline.raw_activity_text",
            before: { state: "null" },
            after: { state: "present", value: summary },
          },
        ],
      },
    ],
  };
}
