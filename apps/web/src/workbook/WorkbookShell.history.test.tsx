import {
  rowHistoryActionTestId,
  rowHistoryDeleteButtonTestId,
  rowHistoryDestructiveConfirmButtonTestId,
  rowHistoryItemTestId,
  rowHistoryLoadingTestId,
  rowHistoryMessageTestId,
  rowHistoryOpenButtonTestId,
  rowHistoryPanelTestId,
  rowHistoryRestoreButtonTestId,
  rowHistoryRollbackConfirmButtonTestId,
  rowHistoryRollbackPreviewTestId,
  saveStateTestId,
  timelineScalarEditorTestId,
  workbookConflictControlTestId,
  workbookConflictLocalValueTestId,
  workbookConflictResolverTestId,
  workbookConflictSavedValueTestId,
  workbookConflictSummaryTestId,
} from "@cartulary/ui-contracts";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../testing/fetchMockTestSupport";
import { TimelineWorkbookRuntimeFixture } from "../testing/TimelineWorkbookRuntimeFixture";
import {
  changeInputValue,
  cleanupTimelineWorkbookTestGlobals,
  errorEnvelope,
  extractTimelineConflictResolutionBody,
  extractTimelineJSONBody,
  findWorkbookCell,
  installTimelineWorkbookTestGlobals,
  latestTimelineWebSocket,
  successEnvelope,
  type TimelineWorkbookFetchMock,
  timelineRow,
  visibleGridRowRecordIds,
  waitForVisibleGridRowRecordIds,
  workbookAsyncTimeoutMs,
} from "../testing/timelineWorkbookTestSupport";
import {
  buildRecordRollbackTargetFromHistoryAction,
  type RecordHistoryItem,
} from "./inspector/workbookRecordHistoryModel";
import { timelineViewSchemaId } from "./models/workbookSurfaceRegistry";

vi.mock(
  "@cartulary/grid-adapter",
  async () => import("@cartulary/grid-adapter/test-support"),
);

const actorUserId = "11111111-1111-4111-8111-111111111111";
const changeSetId = "22222222-2222-4222-8222-222222222222";

function historyItem(
  overrides: Partial<RecordHistoryItem> = {},
): RecordHistoryItem {
  return {
    actor_user_id: actorUserId,
    committed_at: "2026-05-11T12:00:00Z",
    history_item_ref: "hitem_server_selector",
    operation: "field_update",
    diff_summary: {
      summary: "field_update timeline_record",
      units: [{ history_unit_kind: "mutation" }],
    },
    change_set_id: changeSetId,
    reversible: true,
    available_rollback_actions: ["history_entry", "change_set"],
    history_entry_ref: "href_server_selector",
    ...overrides,
  };
}

function historyEnvelope(options: {
  deleted?: boolean;
  items?: RecordHistoryItem[];
  recordId?: string;
  rowVersion?: number;
}) {
  return new Response(
    JSON.stringify({
      data: {
        incident_id: "10000000-0000-4000-8000-000000000001",
        record_id: options.recordId ?? "20000000-0000-4000-8000-000000000001",
        row_version: options.rowVersion ?? 4,
        deleted: options.deleted ?? false,
        items: options.items ?? [historyItem()],
      },
      meta: {
        paging: { has_more: false, limit: 50, next_cursor: null },
        request_id: "req-history",
      },
    }),
    { status: 200, headers: { "content-type": "application/json" } },
  );
}

function historyItemTestId(item: RecordHistoryItem) {
  return rowHistoryItemTestId({
    historyItemRef: item.history_item_ref,
  });
}

function historyActionTestId(
  item: RecordHistoryItem,
  action: "change_set" | "history_entry" | "row_restore",
) {
  return rowHistoryActionTestId({
    action,
    historyItemRef: item.history_item_ref,
  });
}

describe("workbook history support coverage", () => {
  let fetchMock: TimelineWorkbookFetchMock;

  beforeEach(() => {
    fetchMock = installTimelineWorkbookTestGlobals();
  });

  afterEach(() => {
    cleanup();
    cleanupTimelineWorkbookTestGlobals();
  });

  async function openTimelineHistoryFromContext(recordId: string) {
    const summaryCell = await screen.findByTestId(
      timelineScalarEditorTestId({
        fieldKey: "timeline.activity_synopsis_text",
        recordId,
        surface: "grid",
      }),
    );
    fireEvent.contextMenu(summaryCell, { clientX: 32, clientY: 48 });
    fireEvent.click(
      await screen.findByTestId(rowHistoryOpenButtonTestId(recordId)),
    );
  }

  function findHistoryDestructiveConfirmButton(
    operation: "delete" | "restore",
  ) {
    return screen.findByTestId(
      rowHistoryDestructiveConfirmButtonTestId({ operation }),
      undefined,
      { timeout: workbookAsyncTimeoutMs },
    );
  }

  it("opens row-centric history from a selected workbook row and renders server metadata only", async () => {
    const historyRecord = historyItem({
      available_rollback_actions: ["change_set", "row_restore"],
      history_entry_ref: "href_hidden_because_not_advertised",
      revision_no: 3,
    });
    fetchMock
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "20000000-0000-4000-8000-000000000001",
              rowVersion: 4,
              summary: "History base",
              captureState: "rough",
            }),
          ],
        }),
      )
      .mockResolvedValueOnce(
        historyEnvelope({
          items: [historyRecord],
        }),
      );

    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    );
    await openTimelineHistoryFromContext(
      "20000000-0000-4000-8000-000000000001",
    );

    await screen.findByTestId(historyItemTestId(historyRecord));
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/records/20000000-0000-4000-8000-000000000001/history",
      expect.objectContaining({ headers: expect.any(Object) }),
    );
    expect(screen.getByTestId(rowHistoryPanelTestId()).textContent).toContain(
      actorUserId,
    );
    expect(screen.getByTestId(rowHistoryPanelTestId()).textContent).toContain(
      "2026-05-11T12:00:00.000Z",
    );
    expect(screen.getByTestId(rowHistoryPanelTestId()).textContent).toContain(
      "field_update timeline_record",
    );
    expect(
      screen.queryByTestId(
        historyActionTestId(
          historyItem({
            available_rollback_actions: ["change_set", "row_restore"],
            history_entry_ref: "href_hidden_because_not_advertised",
            revision_no: 3,
          }),
          "history_entry",
        ),
      ),
    ).toBeNull();
    const advertisedItem = historyItem({
      available_rollback_actions: ["change_set", "row_restore"],
      history_entry_ref: "href_hidden_because_not_advertised",
      revision_no: 3,
    });
    expect(
      screen.getByTestId(historyActionTestId(advertisedItem, "change_set")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(historyActionTestId(advertisedItem, "row_restore")),
    ).toBeTruthy();
    const renamedHistoryRecord = {
      ...advertisedItem,
      operation: "Localized rollback label",
    };
    expect(historyItemTestId(renamedHistoryRecord)).toBe(
      historyItemTestId(advertisedItem),
    );
    expect(historyActionTestId(renamedHistoryRecord, "change_set")).toBe(
      historyActionTestId(advertisedItem, "change_set"),
    );
  });

  it("retargets open row history to the newly selected inspector row", async () => {
    const record1History = historyItem({
      change_set_id: "30000000-0000-4000-8000-000000000001",
      history_entry_ref: "href_record_1",
      history_item_ref: "hitem_record_1",
      operation: "field_update_record_1",
    });
    const record2History = historyItem({
      change_set_id: "30000000-0000-4000-8000-000000000001",
      history_entry_ref: "href_record_2",
      history_item_ref: "hitem_record_2",
      operation: "field_update_record_2",
    });
    const record2HistoryResponse = deferred<Response>();
    fetchMock
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "20000000-0000-4000-8000-000000000001",
              rowVersion: 4,
              summary: "History row one",
              captureState: "rough",
            }),
            timelineRow({
              recordId: "20000000-0000-4000-8000-000000000002",
              rowVersion: 7,
              summary: "History row two",
              captureState: "rough",
            }),
          ],
        }),
      )
      .mockResolvedValueOnce(
        historyEnvelope({
          items: [record1History],
          recordId: "20000000-0000-4000-8000-000000000001",
          rowVersion: 4,
        }),
      )
      .mockImplementationOnce(() => record2HistoryResponse.promise);

    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    );
    await openTimelineHistoryFromContext(
      "20000000-0000-4000-8000-000000000001",
    );
    await screen.findByTestId(historyItemTestId(record1History));

    const record2Summary = await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000002",
      "timeline.activity_synopsis_text",
    );
    fireEvent.focus(record2Summary);

    await screen.findByTestId(rowHistoryLoadingTestId());
    expect(screen.getByTestId(rowHistoryPanelTestId()).textContent).toContain(
      "20000000-0000-4000-8000-000000000002",
    );
    expect(screen.queryByTestId(historyItemTestId(record1History))).toBeNull();
    expect(
      screen.getByTestId(rowHistoryPanelTestId()).textContent,
    ).not.toContain("field_update_record_1");

    record2HistoryResponse.resolve(
      historyEnvelope({
        items: [record2History],
        recordId: "20000000-0000-4000-8000-000000000002",
        rowVersion: 7,
      }),
    );
    await screen.findByTestId(historyItemTestId(record2History));
    expect(screen.queryByTestId(historyItemTestId(record1History))).toBeNull();
    expect(screen.getByTestId(rowHistoryPanelTestId()).textContent).toContain(
      "field_update_record_2",
    );
  });

  it("clears stale rollback previews when open history retargets", async () => {
    const record1History = historyItem({
      change_set_id: "30000000-0000-4000-8000-000000000001",
      history_entry_ref: "href_record_1",
      history_item_ref: "hitem_record_1",
    });
    const record2History = historyItem({
      change_set_id: "30000000-0000-4000-8000-000000000001",
      history_entry_ref: "href_record_2",
      history_item_ref: "hitem_record_2",
    });
    const record2HistoryResponse = deferred<Response>();
    fetchMock
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "20000000-0000-4000-8000-000000000001",
              rowVersion: 4,
              summary: "Pending rollback row one",
              captureState: "rough",
            }),
            timelineRow({
              recordId: "20000000-0000-4000-8000-000000000002",
              rowVersion: 7,
              summary: "Pending rollback row two",
              captureState: "rough",
            }),
          ],
        }),
      )
      .mockResolvedValueOnce(
        historyEnvelope({
          items: [record1History],
          recordId: "20000000-0000-4000-8000-000000000001",
          rowVersion: 4,
        }),
      )
      .mockImplementationOnce(() => record2HistoryResponse.promise);

    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    );
    await openTimelineHistoryFromContext(
      "20000000-0000-4000-8000-000000000001",
    );
    await screen.findByTestId(
      historyActionTestId(record1History, "history_entry"),
    );
    fireEvent.click(
      screen.getByTestId(historyActionTestId(record1History, "history_entry")),
    );
    expect(
      screen.getByTestId(
        rowHistoryRollbackPreviewTestId({
          action: "history_entry",
          historyItemRef: record1History.history_item_ref,
        }),
      ),
    ).toBeTruthy();

    const record2Summary = await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000002",
      "timeline.activity_synopsis_text",
    );
    fireEvent.focus(record2Summary);

    await screen.findByTestId(rowHistoryLoadingTestId());
    expect(
      screen.queryByTestId(
        rowHistoryRollbackPreviewTestId({
          action: "history_entry",
          historyItemRef: record1History.history_item_ref,
        }),
      ),
    ).toBeNull();

    record2HistoryResponse.resolve(
      historyEnvelope({
        items: [record2History],
        recordId: "20000000-0000-4000-8000-000000000002",
        rowVersion: 7,
      }),
    );
    await screen.findByTestId(historyItemTestId(record2History));
  });

  it.each([
    [
      "missing history_item_ref",
      [
        {
          ...historyItem(),
          history_item_ref: undefined,
        } as unknown as RecordHistoryItem,
      ],
    ],
    [
      "non-string history_item_ref",
      [
        {
          ...historyItem(),
          history_item_ref: 42,
        } as unknown as RecordHistoryItem,
      ],
    ],
    [
      "duplicate history_item_ref",
      [
        historyItem({ change_set_id: "30000000-0000-4000-8000-000000000001" }),
        historyItem({ change_set_id: "30000000-0000-4000-8000-000000000001" }),
      ],
    ],
    [
      "invalid revision_no",
      [
        historyItem({
          available_rollback_actions: ["row_restore"],
          revision_no: 0,
        }),
      ],
    ],
    [
      "non-canonical action order",
      [
        historyItem({
          available_rollback_actions: ["change_set", "history_entry"],
        }),
      ],
    ],
    [
      "missing history_entry rollback selector",
      [
        {
          ...historyItem({
            available_rollback_actions: ["history_entry"],
          }),
          history_entry_ref: undefined,
        } as unknown as RecordHistoryItem,
      ],
    ],
  ])("fails closed for %s", async (_caseName, items) => {
    fetchMock
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "20000000-0000-4000-8000-000000000001",
              rowVersion: 4,
              summary: "History base",
              captureState: "rough",
            }),
          ],
        }),
      )
      .mockResolvedValueOnce(
        historyEnvelope({
          items: items as RecordHistoryItem[],
        }),
      );

    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    );
    await openTimelineHistoryFromContext(
      "20000000-0000-4000-8000-000000000001",
    );

    expect(
      (await screen.findByTestId(rowHistoryMessageTestId())).textContent,
    ).toBe("Invalid row history response.");
    expect(screen.queryByTestId(rowHistoryDeleteButtonTestId())).toBeNull();
  });

  it("builds rollback targets only from advertised server actions and selectors", () => {
    const item = historyItem({
      available_rollback_actions: ["change_set", "row_restore"],
      history_entry_ref: "href_present_but_not_legal",
      revision_no: 7,
    });

    expect(
      buildRecordRollbackTargetFromHistoryAction(item, "history_entry"),
    ).toBeNull();
    expect(
      buildRecordRollbackTargetFromHistoryAction(item, "change_set"),
    ).toEqual({
      kind: "change_set",
      change_set_id: changeSetId,
    });
    expect(
      buildRecordRollbackTargetFromHistoryAction(item, "row_restore"),
    ).toEqual({
      kind: "row_restore",
      restore_to_revision_no: 7,
    });
  });

  it("submits single-entry rollback with only the server-provided history selector", async () => {
    fetchMock
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "20000000-0000-4000-8000-000000000001",
              rowVersion: 4,
              summary: "Rollback current",
              captureState: "rough",
            }),
          ],
        }),
      )
      .mockResolvedValueOnce(historyEnvelope({ rowVersion: 4 }))
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          record_id: "20000000-0000-4000-8000-000000000001",
          row_version: 5,
          target: {
            kind: "history_entry",
            history_entry_ref: "href_server_selector",
          },
          target_change_set_id: changeSetId,
          rollback_change_set_id: "30000000-0000-4000-8000-000000000001",
          affected_record_ids: ["20000000-0000-4000-8000-000000000001"],
        }),
      )
      .mockResolvedValueOnce(historyEnvelope({ rowVersion: 5 }))
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "20000000-0000-4000-8000-000000000001",
              rowVersion: 5,
              summary: "Rollback previous",
              captureState: "rough",
            }),
          ],
        }),
      );

    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    );
    const rollbackItem = historyItem();
    await openTimelineHistoryFromContext(
      "20000000-0000-4000-8000-000000000001",
    );
    await screen.findByTestId(
      historyActionTestId(rollbackItem, "history_entry"),
    );
    fireEvent.click(
      screen.getByTestId(historyActionTestId(rollbackItem, "history_entry")),
    );
    fireEvent.click(
      screen.getByTestId(
        rowHistoryRollbackConfirmButtonTestId({
          action: "history_entry",
          historyItemRef: rollbackItem.history_item_ref,
        }),
      ),
    );

    await waitFor(() => {
      expect(
        fetchMock.mock.calls.some(([url]) =>
          String(url).endsWith(
            "/api/v1/records/20000000-0000-4000-8000-000000000001/rollback",
          ),
        ),
      ).toBe(true);
    });
    const rollbackCallIndex = fetchMock.mock.calls.findIndex(([url]) =>
      String(url).endsWith(
        "/api/v1/records/20000000-0000-4000-8000-000000000001/rollback",
      ),
    );
    const body = extractTimelineJSONBody(fetchMock, rollbackCallIndex);
    expect(body).toMatchObject({
      base_row_version: 4,
      target: {
        kind: "history_entry",
        history_entry_ref: "href_server_selector",
      },
    });
    expect(Object.keys(body.target as Record<string, unknown>).sort()).toEqual([
      "history_entry_ref",
      "kind",
    ]);
  });

  it("uses active and tombstone row versions for delete and restore controls", async () => {
    fetchMock
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "20000000-0000-4000-8000-000000000001",
              rowVersion: 5,
              summary: "Delete me",
              captureState: "rough",
            }),
            timelineRow({
              recordId: "20000000-0000-4000-8000-000000000002",
              rowVersion: 2,
              summary: "Keep me visible",
              captureState: "rough",
            }),
          ],
        }),
      )
      .mockResolvedValueOnce(historyEnvelope({ rowVersion: 5 }))
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          record_id: "20000000-0000-4000-8000-000000000001",
          row_version: 6,
          deleted: true,
          deleted_at: "2026-05-11T12:05:00Z",
          deleted_by_user_id: actorUserId,
          change_set_id: "30000000-0000-4000-8000-000000000001",
        }),
      )
      .mockResolvedValueOnce(historyEnvelope({ deleted: true, rowVersion: 6 }))
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "20000000-0000-4000-8000-000000000002",
              rowVersion: 2,
              summary: "Keep me visible",
              captureState: "rough",
            }),
          ],
        }),
      )
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          record_id: "20000000-0000-4000-8000-000000000001",
          row_version: 7,
          deleted: false,
          deleted_at: null,
          deleted_by_user_id: null,
          change_set_id: "30000000-0000-4000-8000-000000000001",
        }),
      )
      .mockResolvedValueOnce(historyEnvelope({ deleted: false, rowVersion: 7 }))
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "20000000-0000-4000-8000-000000000001",
              rowVersion: 7,
              summary: "Delete me",
              captureState: "rough",
            }),
            timelineRow({
              recordId: "20000000-0000-4000-8000-000000000002",
              rowVersion: 2,
              summary: "Keep me visible",
              captureState: "rough",
            }),
          ],
        }),
      );

    const { container } = render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    );
    await openTimelineHistoryFromContext(
      "20000000-0000-4000-8000-000000000001",
    );
    await screen.findByTestId(rowHistoryDeleteButtonTestId());
    fireEvent.click(screen.getByTestId(rowHistoryDeleteButtonTestId()));
    fireEvent.click(await findHistoryDestructiveConfirmButton("delete"));

    await screen.findByTestId(rowHistoryRestoreButtonTestId());
    await waitFor(() => {
      expect(visibleGridRowRecordIds(container)).toEqual([
        "20000000-0000-4000-8000-000000000002",
      ]);
    });
    expect(screen.getByTestId(rowHistoryPanelTestId()).textContent).toContain(
      "20000000-0000-4000-8000-000000000001",
    );
    expect(
      screen.getByTestId(rowHistoryPanelTestId()).textContent,
    ).not.toContain("20000000-0000-4000-8000-000000000002");
    const deleteCallIndex = fetchMock.mock.calls.findIndex(
      ([url, init]) =>
        String(url).endsWith(
          "/api/v1/records/20000000-0000-4000-8000-000000000001",
        ) && init?.method === "DELETE",
    );
    expect(extractTimelineJSONBody(fetchMock, deleteCallIndex)).toMatchObject({
      base_row_version: 5,
    });

    fireEvent.click(screen.getByTestId(rowHistoryRestoreButtonTestId()));
    fireEvent.click(await findHistoryDestructiveConfirmButton("restore"));
    await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    );
    const restoreCallIndex = fetchMock.mock.calls.findIndex(([url]) =>
      String(url).endsWith(
        "/api/v1/records/20000000-0000-4000-8000-000000000001/restore",
      ),
    );
    expect(extractTimelineJSONBody(fetchMock, restoreCallIndex)).toMatchObject({
      base_row_version: 6,
    });
  });

  it("keeps workbook continuity through ordinary remove and invalidate socket updates", async () => {
    const historyRecord = historyItem();
    fetchMock
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "20000000-0000-4000-8000-000000000001",
              rowVersion: 1,
              summary: "Socket row",
              captureState: "rough",
            }),
          ],
        }),
      )
      .mockResolvedValueOnce(
        historyEnvelope({ items: [historyRecord], rowVersion: 1 }),
      )
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          view_schema_id: timelineViewSchemaId,
          rows: [],
        }),
      )
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "20000000-0000-4000-8000-000000000001",
              rowVersion: 3,
              summary: "Socket row restored",
              captureState: "rough",
            }),
          ],
        }),
      );

    const { container } = render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    );
    await openTimelineHistoryFromContext(
      "20000000-0000-4000-8000-000000000001",
    );
    await screen.findByTestId(historyItemTestId(historyRecord));

    latestTimelineWebSocket()?.emit({
      type: "record_changed",
      stream_seq: 1,
      payload: {
        record_id: "20000000-0000-4000-8000-000000000001",
        row_version: 2,
        change_set_id: "30000000-0000-4000-8000-000000000001",
        client_txn_id: "remote-delete",
        actor_user_id: actorUserId,
        changed_field_keys: [],
        affected_views: [
          {
            view_schema_id: timelineViewSchemaId,
            change_kind: "remove",
          },
        ],
      },
    });
    await waitFor(() => {
      expect(visibleGridRowRecordIds(container)).toEqual([]);
    });

    latestTimelineWebSocket()?.emit({
      type: "record_changed",
      stream_seq: 2,
      payload: {
        record_id: "20000000-0000-4000-8000-000000000001",
        row_version: 3,
        change_set_id: "30000000-0000-4000-8000-000000000001",
        client_txn_id: "remote-restore",
        actor_user_id: actorUserId,
        changed_field_keys: [],
        affected_views: [
          {
            view_schema_id: timelineViewSchemaId,
            change_kind: "invalidate",
          },
        ],
      },
    });

    await waitFor(() => {
      expect(visibleGridRowRecordIds(container, timelineViewSchemaId)).toEqual([
        "20000000-0000-4000-8000-000000000001",
      ]);
    });
    expect(
      await findWorkbookCell(
        container,
        timelineViewSchemaId,
        "20000000-0000-4000-8000-000000000001",
        "timeline.activity_synopsis_text",
      ),
    ).toBeTruthy();
  });

  it("keeps resolver state anchored to stable row and field identities across refresh and reorder", async () => {
    fetchMock
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "20000000-0000-4000-8000-000000000001",
              rowVersion: 1,
              summary: "Record one base",
              captureState: "rough",
            }),
            timelineRow({
              recordId: "20000000-0000-4000-8000-000000000002",
              rowVersion: 1,
              summary: "Record two base",
              captureState: "rough",
            }),
          ],
        }),
      )
      .mockResolvedValueOnce(
        errorEnvelope("same_field_conflict", 409, {
          conflict_token: "conflict-token-anchor-ui",
          record_id: "20000000-0000-4000-8000-000000000001",
          field_key: "timeline.activity_synopsis_text",
          conflict_resolution_class: "text_compare_merge",
          base_row_version: 1,
          current_row_version: 2,
          base_value: "Record one base",
          server_value: "Record one server conflict",
          client_value: "Record one local draft",
          suggested_merged_value: "Record one merged suggestion",
        }),
      )
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "20000000-0000-4000-8000-000000000002",
              rowVersion: 3,
              summary: "Record two moved first",
              captureState: "rough",
            }),
            timelineRow({
              recordId: "20000000-0000-4000-8000-000000000001",
              rowVersion: 2,
              summary: "Record one refreshed",
              captureState: "rough",
            }),
          ],
        }),
      );

    const { container } = render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    await waitForVisibleGridRowRecordIds(container, [
      "20000000-0000-4000-8000-000000000001",
      "20000000-0000-4000-8000-000000000002",
    ]);
    const input = (await findWorkbookCell(
      container,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    fireEvent.focus(input);
    await changeInputValue(input, "Record one local draft");
    fireEvent.blur(input);

    const resolver = await screen.findByTestId(
      workbookConflictResolverTestId(),
    );
    expect(resolver.getAttribute("data-conflict-record-id")).toBe(
      "20000000-0000-4000-8000-000000000001",
    );
    expect(resolver.getAttribute("data-conflict-field-key")).toBe(
      "timeline.activity_synopsis_text",
    );
    expect(resolver.getAttribute("data-conflict-base-row-version")).toBe("1");
    expect(resolver.getAttribute("data-conflict-current-row-version")).toBe(
      "2",
    );
    expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Conflict");

    latestTimelineWebSocket()?.emit({
      type: "record_changed",
      stream_seq: 1,
      payload: {
        record_id: "20000000-0000-4000-8000-000000000002",
        row_version: 3,
        change_set_id: "30000000-0000-4000-8000-000000000001",
        client_txn_id: "remote-reorder",
        actor_user_id: actorUserId,
        changed_field_keys: ["timeline.activity_synopsis_text"],
        affected_views: [
          {
            view_schema_id: timelineViewSchemaId,
            change_kind: "invalidate",
          },
        ],
      },
    });

    await waitForVisibleGridRowRecordIds(container, [
      "20000000-0000-4000-8000-000000000002",
      "20000000-0000-4000-8000-000000000001",
    ]);
    expect(
      screen
        .getByTestId(workbookConflictResolverTestId())
        .getAttribute("data-conflict-record-id"),
    ).toBe("20000000-0000-4000-8000-000000000001");
    expect(
      screen
        .getByTestId(workbookConflictResolverTestId())
        .getAttribute("data-conflict-base-row-version"),
    ).toBe("1");
    expect(
      screen.getByTestId(workbookConflictLocalValueTestId()),
    ).toHaveProperty("value", "Record one local draft");
    fireEvent.click(screen.getByTestId(workbookConflictControlTestId("close")));
    await waitFor(() => {
      expect(screen.queryByTestId(workbookConflictResolverTestId())).toBeNull();
      expect(document.activeElement).toBe(
        screen.getByTestId(
          timelineScalarEditorTestId({
            fieldKey: "timeline.activity_synopsis_text",
            recordId: "20000000-0000-4000-8000-000000000001",
            surface: "grid",
          }),
        ),
      );
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe(
        "Conflict",
      );
    });
  });

  it("keeps resolver keyboard focus non-destructive and cancellation preserves conflict draft", async () => {
    fetchMock
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "20000000-0000-4000-8000-000000000001",
              rowVersion: 1,
              summary: "Keyboard conflict base",
              captureState: "rough",
            }),
          ],
        }),
      )
      .mockResolvedValueOnce(
        errorEnvelope("same_field_conflict", 409, {
          conflict_token: "conflict-token-keyboard",
          record_id: "20000000-0000-4000-8000-000000000001",
          field_key: "timeline.activity_synopsis_text",
          conflict_resolution_class: "text_compare_merge",
          base_row_version: 1,
          current_row_version: 2,
          base_value: "Keyboard conflict base",
          server_value: "Keyboard server value",
          client_value: "Keyboard local draft",
          suggested_merged_value: "Keyboard merged suggestion",
        }),
      );

    const { container } = render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    await waitForVisibleGridRowRecordIds(container, [
      "20000000-0000-4000-8000-000000000001",
    ]);
    const input = (await findWorkbookCell(
      container,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    fireEvent.focus(input);
    await changeInputValue(input, "Keyboard local draft");
    fireEvent.blur(input);

    const resolver = await screen.findByTestId(
      workbookConflictResolverTestId(),
    );
    const summary = screen.getByTestId(workbookConflictSummaryTestId());
    await waitFor(() => {
      expect(document.activeElement).toBe(summary);
    });

    fireEvent.keyDown(summary, { key: "Enter" });
    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(screen.getByTestId(workbookConflictResolverTestId())).toBe(resolver);
    expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Conflict");

    fireEvent.keyDown(resolver, { key: "Escape" });
    await waitFor(() => {
      expect(screen.queryByTestId(workbookConflictResolverTestId())).toBeNull();
      expect(document.activeElement).toBe(
        screen.getByTestId(
          timelineScalarEditorTestId({
            fieldKey: "timeline.activity_synopsis_text",
            recordId: "20000000-0000-4000-8000-000000000001",
            surface: "grid",
          }),
        ),
      );
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe(
        "Conflict",
      );
    });
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("preserves resolver draft and focus when a stale conflict token refreshes the same anchor", async () => {
    fetchMock
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "20000000-0000-4000-8000-000000000001",
              rowVersion: 4,
              summary: "Stale token base",
              captureState: "rough",
            }),
          ],
        }),
      )
      .mockResolvedValueOnce(
        errorEnvelope("same_field_conflict", 409, {
          conflict_token: "conflict-token-stale-original",
          record_id: "20000000-0000-4000-8000-000000000001",
          field_key: "timeline.activity_synopsis_text",
          conflict_resolution_class: "text_compare_merge",
          base_row_version: 4,
          current_row_version: 5,
          base_value: "Stale token base",
          server_value: "Original server",
          client_value: "Original local",
          suggested_merged_value: "Original suggested",
        }),
      )
      .mockResolvedValueOnce(
        errorEnvelope("same_field_conflict", 409, {
          conflict_token: "conflict-token-stale-refresh",
          record_id: "20000000-0000-4000-8000-000000000001",
          field_key: "timeline.activity_synopsis_text",
          conflict_resolution_class: "text_compare_merge",
          base_row_version: 5,
          current_row_version: 6,
          base_value: "Original server",
          server_value: "Fresh server",
          client_value: "Analyst merged draft",
          suggested_merged_value: "Fresh server suggested",
        }),
      );

    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    const input = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    fireEvent.focus(input);
    await changeInputValue(input, "Original local");
    fireEvent.blur(input);

    await screen.findByTestId(workbookConflictResolverTestId());
    fireEvent.change(
      screen.getByTestId(workbookConflictControlTestId("merged-value")),
      {
        target: { value: "Analyst merged draft" },
      },
    );
    fireEvent.click(
      screen.getByTestId(workbookConflictControlTestId("use-merged")),
    );

    await waitFor(() => {
      expect(
        screen
          .getByTestId(workbookConflictResolverTestId())
          .getAttribute("data-conflict-base-row-version"),
      ).toBe("5");
      expect(
        screen
          .getByTestId(workbookConflictResolverTestId())
          .getAttribute("data-conflict-current-row-version"),
      ).toBe("6");
      expect(
        screen.getByTestId(workbookConflictSavedValueTestId()),
      ).toHaveProperty("value", "Fresh server");
      expect(
        screen.getByTestId(workbookConflictLocalValueTestId()),
      ).toHaveProperty("value", "Analyst merged draft");
      expect(
        screen.getByTestId(workbookConflictControlTestId("merged-value")),
      ).toHaveProperty("value", "Analyst merged draft");
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe(
        "Conflict",
      );
    });
    expect(extractTimelineConflictResolutionBody(fetchMock, 0)).toMatchObject({
      conflict_token: "conflict-token-stale-original",
      resolution_kind: "merged_value",
      resolved_value: "Analyst merged draft",
    });

    fireEvent.click(screen.getByTestId(workbookConflictControlTestId("close")));
    await waitFor(() => {
      expect(screen.queryByTestId(workbookConflictResolverTestId())).toBeNull();
      expect(document.activeElement).toBe(
        screen.getByTestId(
          timelineScalarEditorTestId({
            fieldKey: "timeline.activity_synopsis_text",
            recordId: "20000000-0000-4000-8000-000000000001",
            surface: "grid",
          }),
        ),
      );
    });
  });
});
