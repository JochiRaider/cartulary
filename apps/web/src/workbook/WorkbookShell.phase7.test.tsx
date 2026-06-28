import {
  conflictMarkerTestId,
  rowCellTestId,
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
  changeInputValue,
  cleanupTimelineWorkbookTestGlobals,
  deferred,
  errorEnvelope,
  extractTimelineConflictResolutionBody,
  extractTimelineJSONBody,
  findWorkbookCell,
  installTimelineWorkbookTestGlobals,
  latestTimelineWebSocket,
  successEnvelope,
  type TimelineWorkbookFetchMock,
  timelineRow,
  timelineViewSchemaId,
  visibleGridRowRecordIds,
  waitForVisibleGridRowRecordIds,
} from "../testing/timelineWorkbookTestSupport";
import type { RecordHistoryItem } from "./timeline/components/TimelineHistoryPanel";
import { TimelineWorkbook } from "./timeline/components/TimelineWorkbook";
import { buildRecordRollbackTargetFromHistoryAction } from "./timeline/hooks/useTimelineHistoryActions";

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
  return successEnvelope({
    incident_id: "incident-1",
    record_id: options.recordId ?? "record-1",
    row_version: options.rowVersion ?? 4,
    deleted: options.deleted ?? false,
    items: options.items ?? [historyItem()],
  });
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

describe("Phase 7 workbook history support coverage", () => {
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
      rowCellTestId(recordId, "timeline.activity_synopsis_text"),
    );
    fireEvent.contextMenu(summaryCell, { clientX: 32, clientY: 48 });
    fireEvent.click(
      await screen.findByTestId(rowHistoryOpenButtonTestId(recordId)),
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
          incident_id: "incident-1",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "record-1",
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

    render(<TimelineWorkbook incidentId="incident-1" />);
    await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-1",
      "timeline.activity_synopsis_text",
    );
    await openTimelineHistoryFromContext("record-1");

    await screen.findByTestId(historyItemTestId(historyRecord));
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/records/record-1/history",
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
      change_set_id: "change-set-record-1",
      history_entry_ref: "href_record_1",
      history_item_ref: "hitem_record_1",
      operation: "field_update_record_1",
    });
    const record2History = historyItem({
      change_set_id: "change-set-record-2",
      history_entry_ref: "href_record_2",
      history_item_ref: "hitem_record_2",
      operation: "field_update_record_2",
    });
    const record2HistoryResponse = deferred<Response>();
    fetchMock
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "incident-1",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "record-1",
              rowVersion: 4,
              summary: "History row one",
              captureState: "rough",
            }),
            timelineRow({
              recordId: "record-2",
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
          recordId: "record-1",
          rowVersion: 4,
        }),
      )
      .mockImplementationOnce(() => record2HistoryResponse.promise);

    render(<TimelineWorkbook incidentId="incident-1" />);
    await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-1",
      "timeline.activity_synopsis_text",
    );
    const record2Summary = await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-2",
      "timeline.activity_synopsis_text",
    );
    await openTimelineHistoryFromContext("record-1");
    await screen.findByTestId(historyItemTestId(record1History));

    fireEvent.focus(record2Summary);

    await screen.findByTestId(rowHistoryLoadingTestId());
    expect(screen.getByTestId(rowHistoryPanelTestId()).textContent).toContain(
      "Record record-2",
    );
    expect(screen.queryByTestId(historyItemTestId(record1History))).toBeNull();
    expect(
      screen.getByTestId(rowHistoryPanelTestId()).textContent,
    ).not.toContain("field_update_record_1");

    record2HistoryResponse.resolve(
      historyEnvelope({
        items: [record2History],
        recordId: "record-2",
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
      change_set_id: "change-set-record-1",
      history_entry_ref: "href_record_1",
      history_item_ref: "hitem_record_1",
    });
    const record2History = historyItem({
      change_set_id: "change-set-record-2",
      history_entry_ref: "href_record_2",
      history_item_ref: "hitem_record_2",
    });
    const record2HistoryResponse = deferred<Response>();
    fetchMock
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "incident-1",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "record-1",
              rowVersion: 4,
              summary: "Pending rollback row one",
              captureState: "rough",
            }),
            timelineRow({
              recordId: "record-2",
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
          recordId: "record-1",
          rowVersion: 4,
        }),
      )
      .mockImplementationOnce(() => record2HistoryResponse.promise);

    render(<TimelineWorkbook incidentId="incident-1" />);
    await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-1",
      "timeline.activity_synopsis_text",
    );
    const record2Summary = await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-2",
      "timeline.activity_synopsis_text",
    );
    await openTimelineHistoryFromContext("record-1");
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
        recordId: "record-2",
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
        historyItem({ change_set_id: "change-set-a" }),
        historyItem({ change_set_id: "change-set-b" }),
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
          incident_id: "incident-1",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "record-1",
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

    render(<TimelineWorkbook incidentId="incident-1" />);
    await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-1",
      "timeline.activity_synopsis_text",
    );
    await openTimelineHistoryFromContext("record-1");

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
          incident_id: "incident-1",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "record-1",
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
          incident_id: "incident-1",
          record_id: "record-1",
          row_version: 5,
          target: {
            kind: "history_entry",
            history_entry_ref: "href_server_selector",
          },
          target_change_set_id: changeSetId,
          rollback_change_set_id: "33333333-3333-4333-8333-333333333333",
          affected_record_ids: ["record-1"],
        }),
      )
      .mockResolvedValueOnce(historyEnvelope({ rowVersion: 5 }))
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "incident-1",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "record-1",
              rowVersion: 5,
              summary: "Rollback previous",
              captureState: "rough",
            }),
          ],
        }),
      );

    render(<TimelineWorkbook incidentId="incident-1" />);
    await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-1",
      "timeline.activity_synopsis_text",
    );
    const rollbackItem = historyItem();
    await openTimelineHistoryFromContext("record-1");
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
          String(url).endsWith("/api/v1/records/record-1/rollback"),
        ),
      ).toBe(true);
    });
    const rollbackCallIndex = fetchMock.mock.calls.findIndex(([url]) =>
      String(url).endsWith("/api/v1/records/record-1/rollback"),
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
          incident_id: "incident-1",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "record-1",
              rowVersion: 5,
              summary: "Delete me",
              captureState: "rough",
            }),
            timelineRow({
              recordId: "record-2",
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
          incident_id: "incident-1",
          record_id: "record-1",
          row_version: 6,
          deleted: true,
          deleted_at: "2026-05-11T12:05:00Z",
          deleted_by_user_id: actorUserId,
          change_set_id: "44444444-4444-4444-8444-444444444444",
        }),
      )
      .mockResolvedValueOnce(historyEnvelope({ deleted: true, rowVersion: 6 }))
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "incident-1",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "record-2",
              rowVersion: 2,
              summary: "Keep me visible",
              captureState: "rough",
            }),
          ],
        }),
      )
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "incident-1",
          record_id: "record-1",
          row_version: 7,
          deleted: false,
          deleted_at: null,
          deleted_by_user_id: null,
          change_set_id: "55555555-5555-4555-8555-555555555555",
        }),
      )
      .mockResolvedValueOnce(historyEnvelope({ deleted: false, rowVersion: 7 }))
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "incident-1",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "record-1",
              rowVersion: 7,
              summary: "Delete me",
              captureState: "rough",
            }),
            timelineRow({
              recordId: "record-2",
              rowVersion: 2,
              summary: "Keep me visible",
              captureState: "rough",
            }),
          ],
        }),
      );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);
    await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-1",
      "timeline.activity_synopsis_text",
    );
    await openTimelineHistoryFromContext("record-1");
    await screen.findByTestId(rowHistoryDeleteButtonTestId());
    fireEvent.click(screen.getByTestId(rowHistoryDeleteButtonTestId()));
    fireEvent.click(
      screen.getByTestId(
        rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
      ),
    );

    await screen.findByTestId(rowHistoryRestoreButtonTestId());
    await waitFor(() => {
      expect(visibleGridRowRecordIds(container)).toEqual(["record-2"]);
    });
    expect(screen.getByTestId(rowHistoryPanelTestId()).textContent).toContain(
      "Record record-1",
    );
    expect(
      screen.getByTestId(rowHistoryPanelTestId()).textContent,
    ).not.toContain("Record record-2");
    const deleteCallIndex = fetchMock.mock.calls.findIndex(
      ([url, init]) =>
        String(url).endsWith("/api/v1/records/record-1") &&
        init?.method === "DELETE",
    );
    expect(extractTimelineJSONBody(fetchMock, deleteCallIndex)).toMatchObject({
      base_row_version: 5,
    });

    fireEvent.click(screen.getByTestId(rowHistoryRestoreButtonTestId()));
    fireEvent.click(
      await screen.findByTestId(
        rowHistoryDestructiveConfirmButtonTestId({ operation: "restore" }),
      ),
    );
    await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-1",
      "timeline.activity_synopsis_text",
    );
    const restoreCallIndex = fetchMock.mock.calls.findIndex(([url]) =>
      String(url).endsWith("/api/v1/records/record-1/restore"),
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
          incident_id: "incident-1",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "record-1",
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
          incident_id: "incident-1",
          view_schema_id: timelineViewSchemaId,
          rows: [],
        }),
      )
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "incident-1",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "record-1",
              rowVersion: 3,
              summary: "Socket row restored",
              captureState: "rough",
            }),
          ],
        }),
      );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);
    await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-1",
      "timeline.activity_synopsis_text",
    );
    await openTimelineHistoryFromContext("record-1");
    await screen.findByTestId(historyItemTestId(historyRecord));

    latestTimelineWebSocket()?.emit({
      type: "record_changed",
      stream_seq: 1,
      payload: {
        record_id: "record-1",
        row_version: 2,
        change_set_id: "delete-change-set",
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
        record_id: "record-1",
        row_version: 3,
        change_set_id: "restore-change-set",
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
        "record-1",
      ]);
    });
    expect(
      await findWorkbookCell(
        container,
        timelineViewSchemaId,
        "record-1",
        "timeline.activity_synopsis_text",
      ),
    ).toBeTruthy();
  });

  it("FE-U-P7-02 keeps resolver state anchored to stable row and field identities across refresh and reorder", async () => {
    fetchMock
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "incident-1",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "record-1",
              rowVersion: 1,
              summary: "Record one base",
              captureState: "rough",
            }),
            timelineRow({
              recordId: "record-2",
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
          record_id: "record-1",
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
          incident_id: "incident-1",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "record-2",
              rowVersion: 3,
              summary: "Record two moved first",
              captureState: "rough",
            }),
            timelineRow({
              recordId: "record-1",
              rowVersion: 2,
              summary: "Record one refreshed",
              captureState: "rough",
            }),
          ],
        }),
      );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);
    await waitForVisibleGridRowRecordIds(container, ["record-1", "record-2"]);
    const input = (await findWorkbookCell(
      container,
      timelineViewSchemaId,
      "record-1",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    fireEvent.focus(input);
    await changeInputValue(input, "Record one local draft");
    fireEvent.blur(input);

    const resolver = await screen.findByTestId("conflict-resolver");
    expect(resolver.getAttribute("data-conflict-record-id")).toBe("record-1");
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
        record_id: "record-2",
        row_version: 3,
        change_set_id: "change-set-reorder",
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

    await waitForVisibleGridRowRecordIds(container, ["record-2", "record-1"]);
    expect(
      screen
        .getByTestId("conflict-resolver")
        .getAttribute("data-conflict-record-id"),
    ).toBe("record-1");
    expect(
      screen
        .getByTestId("conflict-resolver")
        .getAttribute("data-conflict-base-row-version"),
    ).toBe("1");
    expect(screen.getByTestId("conflict-local-value")).toHaveProperty(
      "value",
      "Record one local draft",
    );
    expect(
      screen.getByTestId(
        conflictMarkerTestId("record-1", "timeline.activity_synopsis_text"),
      ),
    ).toBeTruthy();

    fireEvent.click(screen.getByTestId("conflict-close"));
    await waitFor(() => {
      expect(screen.queryByTestId("conflict-resolver")).toBeNull();
      expect(document.activeElement).toBe(
        screen.getByTestId(
          rowCellTestId("record-1", "timeline.activity_synopsis_text"),
        ),
      );
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe(
        "Conflict",
      );
    });
  });

  it("FE-U-P7-02 keeps resolver keyboard focus non-destructive and cancellation preserves conflict draft", async () => {
    fetchMock
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "incident-1",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "record-1",
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
          record_id: "record-1",
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

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);
    await waitForVisibleGridRowRecordIds(container, ["record-1"]);
    const input = (await findWorkbookCell(
      container,
      timelineViewSchemaId,
      "record-1",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    fireEvent.focus(input);
    await changeInputValue(input, "Keyboard local draft");
    fireEvent.blur(input);

    const resolver = await screen.findByTestId("conflict-resolver");
    const summary = screen.getByTestId("conflict-resolver-summary");
    await waitFor(() => {
      expect(document.activeElement).toBe(summary);
    });

    fireEvent.keyDown(summary, { key: "Enter" });
    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(screen.getByTestId("conflict-resolver")).toBe(resolver);
    expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Conflict");

    fireEvent.keyDown(resolver, { key: "Escape" });
    await waitFor(() => {
      expect(screen.queryByTestId("conflict-resolver")).toBeNull();
      expect(document.activeElement).toBe(
        screen.getByTestId(
          rowCellTestId("record-1", "timeline.activity_synopsis_text"),
        ),
      );
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe(
        "Conflict",
      );
    });
    expect(
      screen.getByTestId(
        conflictMarkerTestId("record-1", "timeline.activity_synopsis_text"),
      ),
    ).toBeTruthy();
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("FE-U-P7-02 preserves resolver draft and focus when a stale conflict token refreshes the same anchor", async () => {
    fetchMock
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "incident-1",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "record-1",
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
          record_id: "record-1",
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
          record_id: "record-1",
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

    render(<TimelineWorkbook incidentId="incident-1" />);
    const input = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-1",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    fireEvent.focus(input);
    await changeInputValue(input, "Original local");
    fireEvent.blur(input);

    await screen.findByTestId("conflict-resolver");
    fireEvent.change(screen.getByTestId("conflict-merged-value"), {
      target: { value: "Analyst merged draft" },
    });
    fireEvent.click(screen.getByTestId("conflict-use-merged"));

    await waitFor(() => {
      expect(
        screen
          .getByTestId("conflict-resolver")
          .getAttribute("data-conflict-base-row-version"),
      ).toBe("5");
      expect(
        screen
          .getByTestId("conflict-resolver")
          .getAttribute("data-conflict-current-row-version"),
      ).toBe("6");
      expect(screen.getByTestId("conflict-server-value")).toHaveProperty(
        "value",
        "Fresh server",
      );
      expect(screen.getByTestId("conflict-local-value")).toHaveProperty(
        "value",
        "Analyst merged draft",
      );
      expect(screen.getByTestId("conflict-merged-value")).toHaveProperty(
        "value",
        "Analyst merged draft",
      );
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe(
        "Conflict",
      );
    });
    expect(extractTimelineConflictResolutionBody(fetchMock, 0)).toMatchObject({
      conflict_token: "conflict-token-stale-original",
      resolution_kind: "merged_value",
      resolved_value: "Analyst merged draft",
    });

    fireEvent.click(screen.getByTestId("conflict-close"));
    await waitFor(() => {
      expect(screen.queryByTestId("conflict-resolver")).toBeNull();
      expect(document.activeElement).toBe(
        screen.getByTestId(
          rowCellTestId("record-1", "timeline.activity_synopsis_text"),
        ),
      );
    });
  });
});
