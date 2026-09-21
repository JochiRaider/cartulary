import type { RecordHistoryItem } from "../workbook/adapters/workbookHistoryResponse";
import {
  acceptHistoryPage,
  beginHistoryRead,
  initialHistoryBrowsing,
} from "../workbook/history/workbookHistoryBrowsing";
import {
  initialWorkbookRecordHistoryState,
  workbookRecordHistoryReducer,
} from "../workbook/inspector/workbookRecordHistoryModel";

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

/** Explicit accepted first page for presentation tests; never a production fallback. */
export function historyPresentationFixture(
  subject: import("../workbook/ports/WorkbookRecordSubject").WorkbookRecordSubject,
  page: import("../workbook/history/workbookHistoryPage").HistoryPage,
) {
  const requested = beginHistoryRead(
    initialHistoryBrowsing(
      {
        actorId: "reviewer",
        incidentId: page.incident_id,
        sessionIdentity: "test",
        epoch: 0,
      },
      subject.recordId,
      subject.viewSchemaId,
    ),
    "initial",
  );
  if (!requested.pending) throw new Error("History fixture request missing");
  return workbookRecordHistoryReducer(
    initialWorkbookRecordHistoryState(subject),
    {
      type: "browsing_changed",
      browsing: acceptHistoryPage(requested, requested.pending, page, {
        rowVersion: subject.rowVersion,
        deleted: subject.kind === "deleted",
      }),
    },
  );
}
