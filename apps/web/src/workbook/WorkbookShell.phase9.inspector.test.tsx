import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridScrollportSelector,
  gridShellTestId,
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
  timelineCollectionInputTestId,
  timelineDraftEvidenceAttachSectionTestId,
  timelineDraftEvidenceFileInputTestId,
  timelineInspectorSectionTestId,
  timelineScalarEditorTestId,
  workbookInspectorFeatureActionTestId,
  workbookInspectorFeatureGroupTestId,
  workbookInspectorPanelTestId,
  workbookInspectorToggleTestId,
  workbookShellSlotTestId,
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
  timelineMutationEnvelope,
  timelineRow,
  timelineRowsEnvelope,
  timelineViewSchemaId,
  waitForVisibleGridRowRecordIds,
} from "../testing/timelineWorkbookTestSupport";
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

  it("keeps the Timeline inspector unmounted until explicit activation", async () => {
    fetchMock.mockResolvedValueOnce(
      timelineRowsEnvelope([
        timelineRow({
          recordId: "record-1",
          rowVersion: 1,
          summary: "Grid-first default",
          captureState: "rough",
          editedAt: "2026-06-17T14:18:51.837049343Z",
        }),
      ]),
    );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);
    await waitForVisibleGridRowRecordIds(container, ["record-1"]);

    expect(screen.queryByTestId("timeline-inspector")).toBeNull();
    const primaryGridSlot = screen.getByTestId(
      workbookShellSlotTestId("primary-grid"),
    );
    const workArea = primaryGridSlot.parentElement;
    expect(workArea).toBeInstanceOf(HTMLElement);
    expect((workArea as HTMLElement).style.position).toBe("relative");
    expect((workArea as HTMLElement).style.display).toBe("grid");
    expect((workArea as HTMLElement).style.gridTemplateRows).toBe(
      "minmax(0, 1fr)",
    );
    expect((workArea as HTMLElement).style.inlineSize).toBe("100%");
    expect((workArea as HTMLElement).style.blockSize).toBe("100%");
    expect((workArea as HTMLElement).style.overflow).toBe("hidden");
    expect(primaryGridSlot.style.inlineSize).toBe("100%");
    expect(primaryGridSlot.style.blockSize).toBe("100%");
    const workbookShell = workArea?.parentElement;
    expect(workbookShell).toBeInstanceOf(HTMLElement);
    expect((workbookShell as HTMLElement).style.gridTemplateRows).toBe(
      "var(--ct-layout-viewBarHeight) minmax(0, 1fr) var(--ct-layout-statusStripHeight)",
    );
    expect(
      screen.getByTestId(workbookShellSlotTestId("view-bar")).style.gridRow,
    ).toBe("1");
    expect((workArea as HTMLElement).style.gridRow).toBe("2");
    expect(
      screen.getByTestId(workbookShellSlotTestId("status-strip")).style.gridRow,
    ).toBe("3");
    const gridShell = screen.getByTestId(gridShellTestId(timelineViewSchemaId));
    expect(gridShell.style.inlineSize).toBe("100%");
    expect(gridShell.style.blockSize).toBe("100%");
    const scrollport = gridShell.querySelector(
      gridScrollportSelector(),
    ) as HTMLElement | null;
    expect(scrollport).toBeInstanceOf(HTMLElement);
    expect((scrollport as HTMLElement).style.width).toBe("100%");
    expect(["0", "0px"]).toContain((scrollport as HTMLElement).style.minWidth);
    expect(
      screen.queryByTestId(rowCellTestId("record-1", "timeline.edited_at")),
    ).toBeNull();
    expect(
      screen.getByTestId(
        rowCellTestId("record-1", "timeline.activity_utc_text"),
      ),
    ).toBeTruthy();

    fireEvent.click(
      screen.getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId)),
    );

    await waitFor(() => {
      expect(screen.getByTestId("timeline-inspector")).toBeTruthy();
    });
    const inspectorSlot = screen.getByTestId("timeline-inspector")
      .parentElement as HTMLElement;
    expect(inspectorSlot.style.position).toBe("absolute");
    expect(["0", "0px"]).toContain(
      inspectorSlot.style.getPropertyValue("inset-block"),
    );
    expect(["0", "0px"]).toContain(
      inspectorSlot.style.getPropertyValue("inset-inline-end"),
    );
    expect(inspectorSlot.style.zIndex).toBe("8");
    expect(inspectorSlot.style.inlineSize).toBe(
      "min(var(--ct-layout-inspectorDefaultWidth), calc(100% - var(--ct-spacing-xl)))",
    );
    expect(inspectorSlot.style.overflow).toBe("hidden");
    expect((workArea as HTMLElement).style.gridTemplateRows).toBe(
      "minmax(0, 1fr)",
    );
    expect(gridShell.style.inlineSize).toBe("100%");
    expect(gridShell.style.blockSize).toBe("100%");
    expect((scrollport as HTMLElement).style.width).toBe("100%");
    expect(["0", "0px"]).toContain((scrollport as HTMLElement).style.minWidth);
  });

  it("FE-U-P9-02 closes the Timeline inspector when the active sheet identity changes", async () => {
    fetchMock.mockResolvedValueOnce(
      timelineRowsEnvelope([
        timelineRow({
          recordId: "record-1",
          rowVersion: 1,
          summary: "Reset inspector",
          captureState: "rough",
        }),
      ]),
    );

    const { container, rerender } = render(
      <TimelineWorkbook
        incidentId="incident-1"
        inspectorResetKey="cartulary.view.timeline.v2:base"
      />,
    );
    await waitForVisibleGridRowRecordIds(container, ["record-1"]);
    fireEvent.click(
      screen.getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId)),
    );
    expect(await screen.findByTestId("timeline-inspector")).toBeTruthy();

    rerender(
      <TimelineWorkbook
        incidentId="incident-1"
        inspectorResetKey="cartulary.view.timeline.v2:saved-view"
      />,
    );

    await waitFor(() => {
      expect(screen.queryByTestId("timeline-inspector")).toBeNull();
    });
    expect(
      screen.getByTestId(
        rowCellTestId("record-1", "timeline.activity_synopsis_text"),
      ),
    ).toBeTruthy();
  });

  it("FE-U-P9-02 renders the configured no-row state when no saved row is selected", async () => {
    fetchMock.mockResolvedValueOnce(timelineRowsEnvelope([]));

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);
    await waitForVisibleGridRowRecordIds(container, []);
    fireEvent.click(
      screen.getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId)),
    );

    expect(
      (await screen.findByTestId("timeline-inspector")).textContent,
    ).toContain("no_row_selected");
    expect(
      screen.queryByTestId(timelineInspectorSectionTestId("evidence")),
    ).toBeNull();
    expect(
      screen.queryByTestId(
        timelineScalarEditorTestId({
          fieldKey: "timeline.raw_activity_text",
          recordId: "record-1",
          surface: "inspector",
        }),
      ),
    ).toBeNull();
    expect(
      screen.getByTestId(timelineDraftEvidenceAttachSectionTestId()),
    ).toBeTruthy();
    const draftEvidenceInput = screen.getByTestId(
      timelineDraftEvidenceFileInputTestId(),
    );
    expect(draftEvidenceInput).toBeInstanceOf(HTMLInputElement);
    expect((draftEvidenceInput as HTMLInputElement).type).toBe("file");
  });

  it("renders Timeline collection cells compactly until inline edit activation", async () => {
    fetchMock.mockResolvedValueOnce(
      timelineRowsEnvelope([
        timelineRow({
          recordId: "record-1",
          rowVersion: 1,
          summary: "Compact collections",
          captureState: "rough",
          hostRefs: [
            {
              item_kind: "unresolved_ref",
              item_ref: "host-ref-1",
              raw_text: "wide-host-token-1",
              entity_type: "host",
              resolution_status: "unresolved",
            },
            {
              item_kind: "unresolved_ref",
              item_ref: "host-ref-2",
              raw_text: "wide-host-token-2",
              entity_type: "host",
              resolution_status: "unresolved",
            },
          ],
          identityRefs: [
            {
              item_kind: "unresolved_ref",
              item_ref: "identity-ref-1",
              raw_text: "wide-identity-token",
              entity_type: "identity",
              resolution_status: "unresolved",
            },
          ],
          tags: ["tag-one", "tag-two", "tag-three"],
        }),
      ]),
    );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);
    await waitForVisibleGridRowRecordIds(container, ["record-1"]);
    fireEvent.contextMenu(
      screen.getByTestId(
        rowCellTestId("record-1", "timeline.activity_synopsis_text"),
      ),
      { clientX: 32, clientY: 48 },
    );
    fireEvent.click(screen.getByTestId(rowInspectButtonTestId("record-1")));

    const hostItems = screen.getByTestId(
      relationshipItemsTestId("record-1", "timeline.host_refs"),
    );
    expect(hostItems.textContent).toContain("wide-host-token-1");
    expect(hostItems.textContent).toContain("+1");
    expect(hostItems.style.flexWrap).toBe("nowrap");
    expect(hostItems.style.overflow).toBe("hidden");

    const identityItems = screen.getByTestId(
      relationshipItemsTestId("record-1", "timeline.identity_refs"),
    );
    expect(identityItems.textContent).toContain("wide-identity-token");

    const tagItems = screen.getByTestId(
      relationshipItemsTestId("record-1", "timeline.tags"),
    );
    expect(tagItems.textContent).toContain("tag-one");
    expect(tagItems.textContent).toContain("+2");

    for (const fieldKey of [
      "timeline.host_refs",
      "timeline.identity_refs",
      "timeline.tags",
    ] as const) {
      expect(
        screen
          .getByTestId(timelineCollectionInputTestId("record-1", fieldKey))
          .getAttribute("placeholder"),
      ).toBeNull();
    }

    for (const [fieldKey, placeholder] of [
      ["timeline.host_refs", "Add hosts token"],
      ["timeline.identity_refs", "Add identities token"],
      ["timeline.tags", "Add tags token"],
    ] as const) {
      const input = screen.getByTestId(
        timelineCollectionInputTestId("record-1", fieldKey),
      ) as HTMLInputElement;
      fireEvent.click(input.parentElement as HTMLElement);
      await waitFor(() => {
        expect(input.getAttribute("placeholder")).toBe(placeholder);
      });
    }
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
    fireEvent.contextMenu(
      screen.getByTestId(
        rowCellTestId("record-2", "timeline.activity_synopsis_text"),
      ),
      { clientX: 32, clientY: 48 },
    );
    fireEvent.click(screen.getByTestId(rowInspectButtonTestId("record-2")));

    expect(screen.getByTestId("timeline-inspector").textContent).toContain(
      "Phase 9 selected row",
    );
    expect(
      (
        screen.getByTestId(
          timelineScalarEditorTestId({
            fieldKey: "timeline.raw_activity_text",
            recordId: "record-2",
            surface: "inspector",
          }),
        ) as HTMLTextAreaElement
      ).value,
    ).toBe("Selected row details");
    for (const section of [
      "operational-text",
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
        workbookInspectorPanelTestId(timelineViewSchemaId, "workflow"),
      ),
    ).toBeTruthy();
    expect(
      screen.getByTestId(
        workbookInspectorFeatureGroupTestId(
          timelineViewSchemaId,
          "create_related.task_request",
        ),
      ),
    ).toBeTruthy();
    expect(
      screen
        .getByTestId(
          workbookInspectorFeatureActionTestId(
            timelineViewSchemaId,
            "timeline.mark_reviewed",
          ),
        )
        .getAttribute("data-route-owner"),
    ).toBe("record_mark_reviewed_route");
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
      "Phase 9 selected row",
    );
    await waitFor(() => {
      expect(
        (
          screen.getByTestId(
            timelineScalarEditorTestId({
              fieldKey: "timeline.raw_activity_text",
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
      "timeline.activity_synopsis_text",
    );
    selectedCell.focus();
    rerender(<TimelineWorkbook incidentId="incident-1" reloadToken={2} />);
    await waitForVisibleGridRowRecordIds(container, ["record-3", "record-1"]);
    await waitFor(() => {
      expect(
        screen.getByTestId("timeline-inspector").textContent,
      ).not.toContain("Phase 9 selected row");
      expect(document.activeElement).toBe(
        screen.getByTestId(
          rowCellTestId("record-3", "timeline.activity_synopsis_text"),
        ),
      );
    });
  });

  it("creates a related Task Request from the Timeline inspector using emitted seed bindings", async () => {
    const taskRequestsViewSchemaId = "cartulary.view.task_requests.v1";
    fetchMock
      .mockResolvedValueOnce(
        timelineRowsEnvelope([
          timelineRow({
            recordId: "record-1",
            rowVersion: 5,
            summary: "Create task source",
            captureState: "rough",
          }),
        ]),
      )
      .mockResolvedValueOnce(
        successEnvelope({
          view_schema_id: taskRequestsViewSchemaId,
          change_set_id: "task-change-set",
          row: {
            record_id: "task-1",
            row_version: 1,
            cells: {},
          },
        }),
      );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);
    await waitForVisibleGridRowRecordIds(container, ["record-1"]);
    fireEvent.contextMenu(
      screen.getByTestId(
        rowCellTestId("record-1", "timeline.activity_synopsis_text"),
      ),
      { clientX: 32, clientY: 48 },
    );
    fireEvent.click(screen.getByTestId(rowInspectButtonTestId("record-1")));
    fireEvent.click(
      screen.getByTestId(
        workbookInspectorFeatureActionTestId(
          timelineViewSchemaId,
          "create_related.task_request",
        ),
      ),
    );

    fireEvent.change(
      screen.getByTestId(genericCreateFieldTestId("task.title")),
      {
        target: { value: "Follow up on source row" },
      },
    );
    fireEvent.change(
      screen.getByTestId(genericCreateFieldTestId("task.task_kind")),
      {
        target: { value: "follow_up" },
      },
    );
    fireEvent.click(
      screen.getByTestId(genericCreateSubmitTestId(taskRequestsViewSchemaId)),
    );

    await waitFor(() => {
      expect(
        fetchMock.mock.calls.some(([url]) =>
          String(url).endsWith(
            `/api/v1/incidents/incident-1/views/${taskRequestsViewSchemaId}/rows`,
          ),
        ),
      ).toBe(true);
    });
    const createCallIndex = fetchMock.mock.calls.findIndex(([url]) =>
      String(url).endsWith(
        `/api/v1/incidents/incident-1/views/${taskRequestsViewSchemaId}/rows`,
      ),
    );
    expect(extractTimelineJSONBody(fetchMock, createCallIndex)).toMatchObject({
      "task.title": "Follow up on source row",
      "task.task_kind": "follow_up",
      "task.linked_record_ids": {
        kind: "collection_actions_v1",
        actions: [{ op: "add_record_ref", linked_record_id: "record-1" }],
      },
    });
    expect(screen.getByTestId("timeline-inspector").textContent).toContain(
      "Created related cartulary.view.task_requests.v1 row task-1.",
    );
  });

  it("creates related Evidence from the Timeline inspector and links it back through the Timeline patch route", async () => {
    const evidenceViewSchemaId = "cartulary.view.evidence.v1";
    fetchMock
      .mockResolvedValueOnce(
        timelineRowsEnvelope([
          timelineRow({
            recordId: "record-1",
            rowVersion: 5,
            summary: "Create evidence source",
            captureState: "rough",
          }),
        ]),
      )
      .mockResolvedValueOnce(
        successEnvelope({
          view_schema_id: evidenceViewSchemaId,
          change_set_id: "evidence-change-set",
          row: {
            record_id: "evidence-1",
            row_version: 1,
            cells: {},
          },
        }),
      )
      .mockResolvedValueOnce(
        timelineMutationEnvelope(
          timelineRow({
            recordId: "record-1",
            rowVersion: 6,
            summary: "Create evidence source",
            captureState: "rough",
            evidenceCount: 1,
            hasEvidence: true,
          }),
          "timeline-link-change-set",
        ),
      )
      .mockResolvedValueOnce(
        timelineRowsEnvelope([
          timelineRow({
            recordId: "record-1",
            rowVersion: 6,
            summary: "Create evidence source",
            captureState: "rough",
            evidenceCount: 1,
            hasEvidence: true,
          }),
        ]),
      );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);
    await waitForVisibleGridRowRecordIds(container, ["record-1"]);
    fireEvent.contextMenu(
      screen.getByTestId(
        rowCellTestId("record-1", "timeline.activity_synopsis_text"),
      ),
      { clientX: 32, clientY: 48 },
    );
    fireEvent.click(screen.getByTestId(rowInspectButtonTestId("record-1")));
    fireEvent.click(
      screen.getByTestId(
        workbookInspectorFeatureActionTestId(
          timelineViewSchemaId,
          "create_related.evidence",
        ),
      ),
    );
    fireEvent.change(
      screen.getByTestId(genericCreateFieldTestId("evidence.title")),
      {
        target: { value: "Source row evidence" },
      },
    );
    fireEvent.click(
      screen.getByTestId(genericCreateSubmitTestId(evidenceViewSchemaId)),
    );

    await waitFor(() => {
      expect(screen.getByTestId("timeline-inspector").textContent).toContain(
        "Created and linked evidence evidence-1.",
      );
    });
    const createCallIndex = fetchMock.mock.calls.findIndex(([url]) =>
      String(url).endsWith(
        `/api/v1/incidents/incident-1/views/${evidenceViewSchemaId}/rows`,
      ),
    );
    expect(extractTimelineJSONBody(fetchMock, createCallIndex)).toMatchObject({
      "evidence.title": "Source row evidence",
    });
    const patchCallIndex = fetchMock.mock.calls.findIndex(
      ([url, init]) =>
        String(url).endsWith("/api/v1/records/record-1") &&
        init?.method === "PATCH",
    );
    expect(extractTimelineJSONBody(fetchMock, patchCallIndex)).toMatchObject({
      view_schema_id: timelineViewSchemaId,
      base_row_version: 5,
      changes: [
        {
          field_key: "timeline.attached_evidence_ids",
          action_payload: {
            kind: "collection_actions_v1",
            actions: [{ op: "add_record_ref", linked_record_id: "evidence-1" }],
          },
        },
      ],
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
      fireEvent.contextMenu(
        screen.getByTestId(
          rowCellTestId("record-1", "timeline.activity_synopsis_text"),
        ),
        { clientX: 32, clientY: 48 },
      );
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
