import {
  relationshipChipTestId,
  relationshipItemsTestId,
  rowCellTestId,
  rowHistoryActionTestId,
  rowHistoryMessageTestId,
  rowHistoryOpenButtonTestId,
  rowHistoryPanelTestId,
  rowHistoryRollbackCancelButtonTestId,
  rowHistoryRollbackConfirmButtonTestId,
  rowHistoryRollbackPreviewTestId,
  rowInspectButtonTestId,
  timelineInspectorSectionTestId,
  timelineScalarEditorTestId,
} from "@cartulary/ui-contracts";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  cleanupTimelineWorkbookTestGlobals,
  errorEnvelope,
  extractTimelineJSONBody,
  findWorkbookCell,
  installTimelineWorkbookTestGlobals,
  successEnvelope,
  type TimelineWorkbookFetchMock,
  timelineRow,
  timelineRowsEnvelope,
  timelineViewSchemaId,
  waitForVisibleGridRowRecordIds,
} from "./timelineWorkbookTestSupport";
import { type RecordHistoryItem, TimelineWorkbook } from "./WorkbookShell";

vi.mock(
  "@cartulary/grid-adapter",
  async () => import("@cartulary/grid-adapter/test-support"),
);

const historyItem: RecordHistoryItem = {
  actor_user_id: "user-history",
  committed_at: "2026-05-11T12:00:00Z",
  history_item_ref: "hitem_phase9_stable",
  operation: "field_update",
  diff_summary: {
    summary: "field_update timeline_record",
    units: [{ history_unit_kind: "mutation" }],
  },
  change_set_id: "22222222-2222-4222-8222-222222222222",
  reversible: true,
  available_rollback_actions: ["history_entry"],
  history_entry_ref: "href_phase9_stable",
};

function historyEnvelope(options: {
  recordId?: string;
  rowVersion?: number;
  items?: RecordHistoryItem[];
}) {
  return successEnvelope({
    incident_id: "incident-1",
    record_id: options.recordId ?? "record-1",
    row_version: options.rowVersion ?? 4,
    deleted: false,
    items: options.items ?? [historyItem],
  });
}

describe("FE-P9 inspector and row-local action coverage", () => {
  let fetchMock: TimelineWorkbookFetchMock;

  beforeEach(() => {
    fetchMock = installTimelineWorkbookTestGlobals();
  });

  afterEach(() => {
    cleanup();
    cleanupTimelineWorkbookTestGlobals();
  });

  it("FE-U-P9-01 Verify inspector selection, tab state, details, relationships, evidence, and history anchors are record_id based and survive row refresh.", async () => {
    const stableRelationship = {
      item_kind: "unresolved_ref",
      item_ref: "rel_ref_phase9_stable",
      raw_text: "Phase 9 visible host label",
      entity_type: "host",
      resolution_status: "unresolved",
    };
    const renamedRelationship = {
      ...stableRelationship,
      raw_text: "Renamed host label after refresh",
    };
    fetchMock
      .mockResolvedValueOnce(
        timelineRowsEnvelope([
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "record-2",
            rowVersion: 3,
            summary: "Phase 9 selected row",
            details: "Selected row details",
            captureState: "rough",
            evidenceCount: 2,
            hasEvidence: true,
            hostRefs: [stableRelationship],
          }),
        ]),
      )
      .mockResolvedValueOnce(
        timelineRowsEnvelope([
          timelineRow({
            recordId: "record-3",
            rowVersion: 1,
            summary: "Inserted before selected row",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "record-2",
            rowVersion: 4,
            summary: "Phase 9 selected row refreshed",
            details: "Selected row details refreshed",
            captureState: "rough",
            evidenceCount: 2,
            hasEvidence: true,
            hostRefs: [renamedRelationship],
          }),
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "rough",
          }),
        ]),
      )
      .mockResolvedValueOnce(
        timelineRowsEnvelope([
          timelineRow({
            recordId: "record-3",
            rowVersion: 1,
            summary: "Inserted before selected row",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "rough",
          }),
        ]),
      );

    const { container, rerender } = render(
      <TimelineWorkbook incidentId="incident-1" reloadToken={0} />,
    );
    await waitForVisibleGridRowRecordIds(container, ["record-1", "record-2"]);
    fireEvent.click(screen.getByTestId(rowInspectButtonTestId("record-2")));

    expect(screen.getByTestId("timeline-inspector").textContent).toContain(
      "Timeline row record-2",
    );
    expect(
      (
        screen.getByTestId(
          timelineScalarEditorTestId({
            fieldKey: "timeline.details",
            recordId: "record-2",
            surface: "inspector",
          }),
        ) as HTMLTextAreaElement
      ).value,
    ).toBe("Selected row details");
    for (const section of [
      "details",
      "relationships",
      "evidence",
      "history",
    ] as const) {
      expect(
        screen.getByTestId(timelineInspectorSectionTestId(section)),
      ).toBeTruthy();
    }
    expect(
      screen.getByTestId(
        relationshipItemsTestId("record-2", "timeline.host_refs"),
      ),
    ).toBeTruthy();
    expect(
      screen.getAllByTestId(relationshipChipTestId("rel_ref_phase9_stable"))
        .length,
    ).toBeGreaterThan(0);

    rerender(<TimelineWorkbook incidentId="incident-1" reloadToken={1} />);
    await waitForVisibleGridRowRecordIds(container, [
      "record-3",
      "record-2",
      "record-1",
    ]);
    expect(screen.getByTestId("timeline-inspector").textContent).toContain(
      "Timeline row record-2",
    );
    await waitFor(() => {
      expect(
        (
          screen.getByTestId(
            timelineScalarEditorTestId({
              fieldKey: "timeline.details",
              recordId: "record-2",
              surface: "inspector",
            }),
          ) as HTMLTextAreaElement
        ).value,
      ).toBe("Selected row details refreshed");
    });
    expect(
      screen.getAllByTestId(relationshipChipTestId("rel_ref_phase9_stable"))
        .length,
    ).toBeGreaterThan(0);

    const selectedCell = await findWorkbookCell(
      container,
      timelineViewSchemaId,
      "record-2",
      "timeline.summary",
    );
    selectedCell.focus();
    rerender(<TimelineWorkbook incidentId="incident-1" reloadToken={2} />);
    await waitForVisibleGridRowRecordIds(container, ["record-3", "record-1"]);
    await waitFor(() => {
      expect(
        screen.getByTestId("timeline-inspector").textContent,
      ).not.toContain("Timeline row record-2");
      expect(document.activeElement).toBe(
        screen.getByTestId(rowCellTestId("record-3", "timeline.summary")),
      );
    });
  });

  it("FE-I-P9-01 Verify history and rollback preview/action use public route contracts, preserve retained history, and render public error envelopes.", async () => {
    for (const errorCode of [
      "row_conflict",
      "stale_row",
      "authorization_denied",
      "invalid_rollback_target",
      "history_missing",
      "rollback_unavailable",
    ]) {
      cleanup();
      cleanupTimelineWorkbookTestGlobals();
      fetchMock = installTimelineWorkbookTestGlobals();
      fetchMock
        .mockResolvedValueOnce(
          timelineRowsEnvelope([
            timelineRow({
              recordId: "record-1",
              rowVersion: 4,
              summary: `Phase 9 rollback ${errorCode}`,
              captureState: "rough",
            }),
          ]),
        )
        .mockResolvedValueOnce(
          errorCode === "history_missing"
            ? errorEnvelope(errorCode, 404)
            : historyEnvelope({ rowVersion: 4 }),
        );

      const { container } = render(
        <TimelineWorkbook incidentId="incident-1" />,
      );
      await waitForVisibleGridRowRecordIds(container, ["record-1"]);
      fireEvent.click(
        screen.getByTestId(rowHistoryOpenButtonTestId("record-1")),
      );

      if (errorCode === "history_missing") {
        expect(
          (await screen.findByTestId(rowHistoryMessageTestId())).textContent,
        ).toContain("history_missing");
        continue;
      }

      const actionAnchor = {
        action: "history_entry" as const,
        historyItemRef: historyItem.history_item_ref,
      };
      await screen.findByTestId(rowHistoryPanelTestId());
      fireEvent.click(screen.getByTestId(rowHistoryActionTestId(actionAnchor)));
      expect(
        screen.getByTestId(rowHistoryRollbackPreviewTestId(actionAnchor))
          .textContent,
      ).toContain(historyItem.history_item_ref);
      fireEvent.click(
        screen.getByTestId(rowHistoryRollbackCancelButtonTestId(actionAnchor)),
      );
      expect(
        screen.queryByTestId(rowHistoryRollbackPreviewTestId(actionAnchor)),
      ).toBeNull();

      fetchMock.mockResolvedValueOnce(errorEnvelope(errorCode, 409));
      fireEvent.click(screen.getByTestId(rowHistoryActionTestId(actionAnchor)));
      fireEvent.click(
        screen.getByTestId(rowHistoryRollbackConfirmButtonTestId(actionAnchor)),
      );
      await waitFor(() => {
        expect(
          screen.getByTestId(rowHistoryMessageTestId()).textContent,
        ).toContain(errorCode);
      });
      const rollbackCallIndex = fetchMock.mock.calls.findIndex(([url]) =>
        String(url).endsWith("/api/v1/records/record-1/rollback"),
      );
      expect(rollbackCallIndex).toBeGreaterThanOrEqual(0);
      const body = extractTimelineJSONBody(fetchMock, rollbackCallIndex);
      expect(body).toMatchObject({
        base_row_version: 4,
        target: {
          kind: "history_entry",
          history_entry_ref: historyItem.history_entry_ref,
        },
      });
      expect(String(body.client_txn_id)).toMatch(/^timeline-client-/u);
    }
  });
});
