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
  timelineInspectorTestId,
  timelineMutationSubstrateReadyTestId,
  timelineScalarEditorTestId,
  workbookInspectorFeatureActionTestId,
  workbookInspectorPanelTestId,
  workbookInspectorToggleTestId,
  workbookLayoutMetrics,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { TimelineWorkbookRuntimeFixture } from "../testing/TimelineWorkbookRuntimeFixture";
import {
  cleanupTimelineWorkbookTestGlobals,
  errorEnvelope,
  extractTimelineJSONBody,
  findWorkbookCell,
  fullWorkbookViewRow,
  installTimelineWorkbookTestGlobals,
  successEnvelope,
  type TimelineWorkbookFetchMock,
  timelineMutationEnvelope,
  timelineRecordActionCalls,
  timelineRow,
  timelineRowsEnvelope,
  waitForVisibleGridRowRecordIds,
} from "../testing/timelineWorkbookTestSupport";
import { historyDiffFixture } from "../testing/workbookHistoryTestSupport";
import { createContextualCreateTransport } from "./adapters/createContextualCreateTransport";
import { createTimelineRelatedEvidenceTransport } from "./adapters/createTimelineRelatedEvidenceTransport";
import type { RecordHistoryItem } from "./adapters/workbookHistoryResponse";
import { ContextualCreateContext } from "./features/coordination/ContextualCreateContext";
import { WorkbookContextualTaskDecisionCreateOwner } from "./features/coordination/WorkbookContextualTaskDecisionCreateOwner";
import { TimelineRelatedEvidenceContext } from "./features/evidence/TimelineRelatedEvidenceContext";
import { WorkbookTimelineRelatedEvidenceOwner } from "./features/evidence/WorkbookTimelineRelatedEvidenceOwner";
import { timelineViewSchemaId } from "./models/workbookSurfaceRegistry";
import { createBrowserSecureTransactionIdPort } from "./mutations/secureTransactionId";
import type { WorkbookMutationAuthority } from "./mutations/workbookMutationAuthority";

vi.mock(
  "@cartulary/grid-adapter",
  async () => import("@cartulary/grid-adapter/test-support"),
);

const historyItem: RecordHistoryItem = {
  actor_user_id: "40000000-0000-4000-8000-000000000401",
  committed_at: "2026-05-11T12:00:00Z",
  history_item_ref: "hitem_workbook_interaction_stable",
  operation: "field_update",
  diff_summary: historyDiffFixture("field_update timeline_record"),
  change_set_id: "30000000-0000-4000-8000-000000000001",
  reversible: true,
  available_rollback_actions: ["history_entry"],
  history_entry_ref: "href_workbook_interaction_stable",
};

function historyEnvelope(options: {
  recordId?: string;
  rowVersion?: number;
  items?: RecordHistoryItem[];
}) {
  return new Response(
    JSON.stringify({
      data: {
        incident_id: "10000000-0000-4000-8000-000000000001",
        record_id: options.recordId ?? "20000000-0000-4000-8000-000000000001",
        row_version: options.rowVersion ?? 4,
        deleted: false,
        representation_generation: "cartulary.history.1",
        items: options.items ?? [historyItem],
      },
      meta: {
        paging: { has_more: false, limit: 50, next_cursor: null },
        request_id: "req-inspector-history",
      },
    }),
    { status: 200, headers: { "content-type": "application/json" } },
  );
}

describe("browser.inspector-history inspector and row-local action coverage", () => {
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
          recordId: "20000000-0000-4000-8000-000000000001",
          rowVersion: 1,
          summary: "Grid-first default",
          captureState: "rough",
          editedAt: "2026-06-17T14:18:51.837049343Z",
        }),
      ]),
    );

    const { container, rerender } = render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    await waitForVisibleGridRowRecordIds(container, [
      "20000000-0000-4000-8000-000000000001",
    ]);

    expect(screen.queryByTestId(timelineInspectorTestId())).toBeNull();
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
    const workbookShell = screen.getByRole("region", {
      name: "Timeline row interaction layer frame",
    });
    expect(workbookShell).toBeInstanceOf(HTMLElement);
    expect((workbookShell as HTMLElement).style.gridTemplateRows).toBe(
      "var(--ct-layout-viewBarHeight) auto minmax(0, 1fr) var(--ct-layout-statusStripHeight)",
    );
    expect(
      screen.getByTestId(workbookShellSlotTestId("view-bar")).style.gridRow,
    ).toBe("1");
    expect((workArea as HTMLElement).style.gridRow).toBe("2");
    expect(
      screen.getByTestId(workbookShellSlotTestId("status-strip")).style.gridRow,
    ).toBe("4");
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
      screen.queryByTestId(
        rowCellTestId(
          "20000000-0000-4000-8000-000000000001",
          "timeline.edited_at",
        ),
      ),
    ).toBeNull();
    expect(
      screen.getByTestId(
        rowCellTestId(
          "20000000-0000-4000-8000-000000000001",
          "timeline.activity_utc_text",
        ),
      ),
    ).toBeTruthy();

    fireEvent.click(
      screen.getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId)),
    );

    await waitFor(() => {
      expect(screen.getByTestId(timelineInspectorTestId())).toBeTruthy();
    });
    const inspectorSlot = screen.getByTestId(timelineInspectorTestId())
      .parentElement as HTMLElement;
    expect(inspectorSlot.style.position).toBe("relative");
    expect(inspectorSlot.style.zIndex).toBe("8");
    expect(inspectorSlot.style.inlineSize).toBe("420px");
    expect(inspectorSlot.style.overflow).toBe("hidden");
    expect((workArea as HTMLElement).style.gridTemplateColumns).toBe(
      "minmax(0, 1fr) 420px",
    );
    const resizeSeparator = screen.getByRole("separator", {
      name: "Resize inspector",
    });
    const inspectorEffectiveMaximum = workbookLayoutMetrics(
      window.visualViewport?.width ?? window.innerWidth,
    ).inspectorEffectiveMaxWidthCssPx;
    expect(resizeSeparator.getAttribute("aria-valuemax")).toBe(
      String(inspectorEffectiveMaximum),
    );
    expect(resizeSeparator.getAttribute("aria-valuenow")).toBe("420");
    fireEvent.keyDown(resizeSeparator, { key: "ArrowLeft" });
    expect(resizeSeparator.getAttribute("aria-valuenow")).toBe("436");
    fireEvent.keyDown(resizeSeparator, { key: "Home" });
    expect(resizeSeparator.getAttribute("aria-valuenow")).toBe("360");
    fireEvent.keyDown(resizeSeparator, { key: "End" });
    expect(resizeSeparator.getAttribute("aria-valuenow")).toBe(
      String(inspectorEffectiveMaximum),
    );

    rerender(
      <TimelineWorkbookRuntimeFixture
        chromeMode="narrow_desktop"
        incidentId="10000000-0000-4000-8000-000000000001"
      />,
    );
    expect(inspectorSlot.style.position).toBe("absolute");
    expect(primaryGridSlot.hasAttribute("inert")).toBe(true);
    expect(
      screen
        .getByTestId(timelineMutationSubstrateReadyTestId())
        .getAttribute("data-inspector-layout"),
    ).toBe("right_overlay");
    expect(
      screen.queryByRole("separator", { name: "Resize inspector" }),
    ).toBeNull();

    rerender(
      <TimelineWorkbookRuntimeFixture
        chromeMode="compact_desktop"
        incidentId="10000000-0000-4000-8000-000000000001"
      />,
    );
    expect(inspectorSlot.style.inlineSize).toBe("100%");
    expect(
      screen
        .getByTestId(timelineMutationSubstrateReadyTestId())
        .getAttribute("data-inspector-layout"),
    ).toBe("full_overlay");

    const inspectorToggle = screen.getByTestId(
      workbookInspectorToggleTestId(timelineViewSchemaId),
    );
    inspectorToggle.focus();
    fireEvent.keyDown(screen.getByTestId(timelineInspectorTestId()), {
      key: "Escape",
    });
    await waitFor(() => {
      expect(screen.queryByTestId(timelineInspectorTestId())).toBeNull();
      expect(document.activeElement).toBe(inspectorToggle);
    });
    expect((workArea as HTMLElement).style.gridTemplateRows).toBe(
      "minmax(0, 1fr)",
    );
    expect(gridShell.style.inlineSize).toBe("100%");
    expect(gridShell.style.blockSize).toBe("100%");
    expect((scrollport as HTMLElement).style.width).toBe("100%");
    expect(["0", "0px"]).toContain((scrollport as HTMLElement).style.minWidth);
  });

  it("closes the Timeline inspector when the active sheet identity changes", async () => {
    fetchMock.mockResolvedValueOnce(
      timelineRowsEnvelope([
        timelineRow({
          recordId: "20000000-0000-4000-8000-000000000001",
          rowVersion: 1,
          summary: "Reset inspector",
          captureState: "rough",
        }),
      ]),
    );

    const { container, rerender } = render(
      <TimelineWorkbookRuntimeFixture
        currentIncidentRole="editor"
        incidentId="10000000-0000-4000-8000-000000000001"
        inspectorResetKey="cartulary.view.timeline.v2:base"
      />,
    );
    await waitForVisibleGridRowRecordIds(container, [
      "20000000-0000-4000-8000-000000000001",
    ]);
    fireEvent.click(
      screen.getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId)),
    );
    expect(await screen.findByTestId(timelineInspectorTestId())).toBeTruthy();

    rerender(
      <TimelineWorkbookRuntimeFixture
        incidentId="10000000-0000-4000-8000-000000000001"
        inspectorResetKey="cartulary.view.timeline.v2:saved-view"
      />,
    );

    await waitFor(() => {
      expect(screen.queryByTestId(timelineInspectorTestId())).toBeNull();
    });
    expect(
      screen.getByTestId(
        rowCellTestId(
          "20000000-0000-4000-8000-000000000001",
          "timeline.activity_synopsis_text",
        ),
      ),
    ).toBeTruthy();
  });

  it("renders the configured no-row state when no saved row is selected", async () => {
    fetchMock.mockResolvedValueOnce(timelineRowsEnvelope([]));

    const { container } = render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    await waitForVisibleGridRowRecordIds(container, []);
    fireEvent.click(
      screen.getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId)),
    );

    expect(
      (await screen.findByTestId(timelineInspectorTestId())).getAttribute(
        "data-inspector-state",
      ),
    ).toBe("no_row_selected");
    expect(
      screen.getByText("Select a saved row to inspect its details."),
    ).toBeTruthy();
    expect(screen.queryByText("no_row_selected")).toBeNull();
    expect(
      screen.queryByTestId(timelineInspectorSectionTestId("evidence")),
    ).toBeNull();
    expect(
      screen.queryByTestId(
        timelineScalarEditorTestId({
          fieldKey: "timeline.raw_activity_text",
          recordId: "20000000-0000-4000-8000-000000000001",
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

  it("renders complete Timeline collections and local inputs in the inspector", async () => {
    fetchMock.mockResolvedValueOnce(
      timelineRowsEnvelope([
        timelineRow({
          recordId: "20000000-0000-4000-8000-000000000001",
          rowVersion: 1,
          summary: "Compact collections",
          captureState: "rough",
          hostRefs: [
            {
              item_kind: "unresolved_mention",
              item_ref: "host-ref-1",
              raw_text: "wide-host-token-1",
              entity_type: "host",
              resolution_status: "unresolved",
            },
            {
              item_kind: "unresolved_mention",
              item_ref: "host-ref-2",
              raw_text: "wide-host-token-2",
              entity_type: "host",
              resolution_status: "unresolved",
            },
          ],
          identityRefs: [
            {
              item_kind: "unresolved_mention",
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

    const { container } = render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    await waitForVisibleGridRowRecordIds(container, [
      "20000000-0000-4000-8000-000000000001",
    ]);
    fireEvent.contextMenu(
      screen.getByTestId(
        rowCellTestId(
          "20000000-0000-4000-8000-000000000001",
          "timeline.activity_synopsis_text",
        ),
      ),
      { clientX: 32, clientY: 48 },
    );
    fireEvent.click(
      screen.getByTestId(
        rowInspectButtonTestId("20000000-0000-4000-8000-000000000001"),
      ),
    );

    const hostItems = screen.getByTestId(
      relationshipItemsTestId(
        "20000000-0000-4000-8000-000000000001",
        "timeline.host_refs",
      ),
    );
    expect(hostItems.textContent).toContain("wide-host-token-1");
    expect(hostItems.textContent).toContain("wide-host-token-2");
    expect(hostItems.querySelectorAll("button")).toHaveLength(2);

    const identityItems = screen.getByTestId(
      relationshipItemsTestId(
        "20000000-0000-4000-8000-000000000001",
        "timeline.identity_refs",
      ),
    );
    expect(identityItems.textContent).toContain("wide-identity-token");

    const tagItems = screen.getByTestId(
      relationshipItemsTestId(
        "20000000-0000-4000-8000-000000000001",
        "timeline.tags",
      ),
    );
    expect(tagItems.textContent).toContain("tag-one");
    expect(tagItems.textContent).toContain("tag-two");
    expect(tagItems.textContent).toContain("tag-three");

    for (const [fieldKey, placeholder] of [
      ["timeline.host_refs", "Add hosts token"],
      ["timeline.identity_refs", "Add identities token"],
      ["timeline.tags", "Add tags token"],
    ] as const) {
      const input = screen.getByTestId(
        timelineCollectionInputTestId(
          "20000000-0000-4000-8000-000000000001",
          fieldKey,
        ),
      ) as HTMLInputElement;
      expect(input.getAttribute("placeholder")).toBe(placeholder);
      expect(input.tabIndex).toBe(0);
    }
  });

  it("Verify inspector selection, tab state, details, relationships, evidence, and history anchors are record_id based and survive row refresh.", async () => {
    const stableRelationship = {
      item_kind: "unresolved_mention",
      item_ref: "rel_ref_workbook_interaction_stable",
      raw_text: "Workbook inspector visible host label",
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
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000002",
            rowVersion: 3,
            summary: "Workbook inspector selected row",
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
            recordId: "20000000-0000-4000-8000-000000000003",
            rowVersion: 1,
            summary: "Inserted before selected row",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000002",
            rowVersion: 4,
            summary: "Workbook inspector selected row refreshed",
            details: "Selected row details refreshed",
            captureState: "rough",
            evidenceCount: 2,
            hasEvidence: true,
            hostRefs: [renamedRelationship],
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "rough",
          }),
        ]),
      )
      .mockResolvedValueOnce(
        timelineRowsEnvelope([
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000003",
            rowVersion: 1,
            summary: "Inserted before selected row",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "rough",
          }),
        ]),
      );

    const { container, rerender } = render(
      <TimelineWorkbookRuntimeFixture
        currentIncidentRole="editor"
        incidentId="10000000-0000-4000-8000-000000000001"
        reloadToken={0}
      />,
    );
    await waitForVisibleGridRowRecordIds(container, [
      "20000000-0000-4000-8000-000000000001",
      "20000000-0000-4000-8000-000000000002",
    ]);
    fireEvent.contextMenu(
      screen.getByTestId(
        rowCellTestId(
          "20000000-0000-4000-8000-000000000002",
          "timeline.activity_synopsis_text",
        ),
      ),
      { clientX: 32, clientY: 48 },
    );
    fireEvent.click(
      screen.getByTestId(
        rowInspectButtonTestId("20000000-0000-4000-8000-000000000002"),
      ),
    );

    expect(screen.getByTestId(timelineInspectorTestId()).textContent).toContain(
      "Workbook inspector selected row",
    );
    fireEvent.click(screen.getByRole("button", { name: "Edit RAW Activity" }));
    expect(
      (
        screen.getByTestId(
          timelineScalarEditorTestId({
            fieldKey: "timeline.raw_activity_text",
            recordId: "20000000-0000-4000-8000-000000000002",
            surface: "inspector",
          }),
        ) as HTMLTextAreaElement
      ).value,
    ).toBe("Selected row details");
    for (const section of ["relationships", "evidence", "history"] as const) {
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
        workbookInspectorFeatureActionTestId(
          timelineViewSchemaId,
          "create_related.task_request",
        ),
      ),
    ).toBeTruthy();
    const markReviewed = screen.getByTestId(
      workbookInspectorFeatureActionTestId(
        timelineViewSchemaId,
        "timeline.mark_reviewed",
      ),
    ) as HTMLButtonElement;
    expect(markReviewed.disabled).toBe(true);
    const permissionDescription = markReviewed.getAttribute("aria-describedby");
    expect(permissionDescription).toBeTruthy();
    expect(
      permissionDescription
        ?.split(/\s+/u)
        .map((id) => document.getElementById(id)?.textContent),
    ).toContain("Requires the reviewer incident role.");
    await userEvent.setup().click(markReviewed);
    expect(timelineRecordActionCalls(fetchMock, "mark-reviewed")).toHaveLength(
      0,
    );
    expect(
      screen
        .getByTestId(
          workbookInspectorFeatureActionTestId(
            timelineViewSchemaId,
            "indicator.observations.manage",
          ),
        )
        .getAttribute("data-route-owner"),
    ).toBe("indicator_observations_route");
    expect(
      screen.getByTestId(
        relationshipItemsTestId(
          "20000000-0000-4000-8000-000000000002",
          "timeline.host_refs",
        ),
      ),
    ).toBeTruthy();
    expect(
      screen.getAllByTestId(
        relationshipChipTestId("rel_ref_workbook_interaction_stable"),
      ).length,
    ).toBeGreaterThan(0);

    rerender(
      <TimelineWorkbookRuntimeFixture
        currentIncidentRole="editor"
        incidentId="10000000-0000-4000-8000-000000000001"
        reloadToken={1}
      />,
    );
    await waitForVisibleGridRowRecordIds(container, [
      "20000000-0000-4000-8000-000000000003",
      "20000000-0000-4000-8000-000000000002",
      "20000000-0000-4000-8000-000000000001",
    ]);
    expect(screen.getByTestId(timelineInspectorTestId()).textContent).toContain(
      "Workbook inspector selected row",
    );
    await waitFor(() => {
      expect(
        (
          screen.getByTestId(
            timelineScalarEditorTestId({
              fieldKey: "timeline.raw_activity_text",
              recordId: "20000000-0000-4000-8000-000000000002",
              surface: "inspector",
            }),
          ) as HTMLTextAreaElement
        ).value,
      ).toBe("Selected row details refreshed");
    });
    expect(
      screen.getAllByTestId(
        relationshipChipTestId("rel_ref_workbook_interaction_stable"),
      ).length,
    ).toBeGreaterThan(0);

    const selectedCell = await findWorkbookCell(
      container,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000002",
      "timeline.activity_synopsis_text",
    );
    selectedCell.focus();
    rerender(
      <TimelineWorkbookRuntimeFixture
        incidentId="10000000-0000-4000-8000-000000000001"
        reloadToken={2}
      />,
    );
    await waitForVisibleGridRowRecordIds(container, [
      "20000000-0000-4000-8000-000000000003",
      "20000000-0000-4000-8000-000000000001",
    ]);
    await waitFor(() => {
      expect(
        screen.getByTestId(timelineInspectorTestId()).textContent,
      ).toContain("Workbook inspector selected row");
      expect(
        screen.getByTestId(
          rowCellTestId(
            "20000000-0000-4000-8000-000000000003",
            "timeline.activity_synopsis_text",
          ),
        ),
      ).toBeTruthy();
    });
  });

  it("creates a related Task Request from the Timeline inspector using emitted seed bindings", async () => {
    const taskRequestsViewSchemaId = "cartulary.view.task_requests.v1";
    const authority: WorkbookMutationAuthority = {
      actorId: "40000000-0000-4000-8000-000000000401",
      incidentId: "10000000-0000-4000-8000-000000000001",
      sessionIdentity: "fixture-session",
      role: "editor",
      closed: false,
    };
    const contextualCreate = new WorkbookContextualTaskDecisionCreateOwner(
      authority.incidentId,
      createBrowserSecureTransactionIdPort(),
      {
        coordinate: async () => ({
          kind: "settled" as const,
          minimumRowVersion: 0,
        }),
        accepted: () => {},
        observed: () => {},
        refresh: async () => {},
      },
    );
    contextualCreate.setAuthority(authority);
    contextualCreate.configure(
      {
        availableViews: async () => ({
          kind: "accepted" as const,
          value: [timelineViewSchemaId],
        }),
        verify: async () => {},
        page: async () => ({
          kind: "accepted",
          value: {
            candidates: [
              {
                recordId: "20000000-0000-4000-8000-000000000001",
                displayText: "Create task source",
                viewSchemaId: timelineViewSchemaId,
                row: timelineRow({
                  recordId: "20000000-0000-4000-8000-000000000001",
                  rowVersion: 5,
                  summary: "Create task source",
                  captureState: "rough",
                }),
              },
            ],
            hasMore: false,
            nextCursor: null,
          },
        }),
      },
      async () => authority,
      createContextualCreateTransport(undefined),
    );
    fetchMock
      .mockResolvedValueOnce(
        timelineRowsEnvelope([
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 5,
            summary: "Create task source",
            captureState: "rough",
          }),
        ]),
      )
      .mockResolvedValueOnce(
        successEnvelope({
          view_schema_id: taskRequestsViewSchemaId,
          change_set_id: "30000000-0000-4000-8000-000000000001",
          row: fullWorkbookViewRow(
            requireViewContract(taskRequestsViewSchemaId),
            "20000000-0000-4000-8000-000000000401",
            1,
            {
              "task.title": "Follow up on source row",
              "task.task_kind": "follow_up",
            },
          ),
        }),
      );

    const { container } = render(
      <ContextualCreateContext.Provider
        value={{
          owner: contextualCreate,
          sheetRef: { kind: "view_schema", id: timelineViewSchemaId },
        }}
      >
        <TimelineWorkbookRuntimeFixture
          currentIncidentRole="editor"
          currentUserId={authority.actorId}
          incidentId={authority.incidentId}
        />
      </ContextualCreateContext.Provider>,
    );
    await waitForVisibleGridRowRecordIds(container, [
      "20000000-0000-4000-8000-000000000001",
    ]);
    fireEvent.contextMenu(
      screen.getByTestId(
        rowCellTestId(
          "20000000-0000-4000-8000-000000000001",
          "timeline.activity_synopsis_text",
        ),
      ),
      { clientX: 32, clientY: 48 },
    );
    fireEvent.click(
      screen.getByTestId(
        rowInspectButtonTestId("20000000-0000-4000-8000-000000000001"),
      ),
    );
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
            `/api/v1/incidents/10000000-0000-4000-8000-000000000001/views/${taskRequestsViewSchemaId}/rows`,
          ),
        ),
      ).toBe(true);
    });
    const createCallIndex = fetchMock.mock.calls.findIndex(([url]) =>
      String(url).endsWith(
        `/api/v1/incidents/10000000-0000-4000-8000-000000000001/views/${taskRequestsViewSchemaId}/rows`,
      ),
    );
    expect(extractTimelineJSONBody(fetchMock, createCallIndex)).toMatchObject({
      "task.title": "Follow up on source row",
      "task.task_kind": "follow_up",
      "task.linked_record_ids": {
        kind: "collection_actions_v1",
        actions: [
          {
            op: "add_record_ref",
            linked_record_id: "20000000-0000-4000-8000-000000000001",
          },
        ],
      },
    });
    await waitFor(() =>
      expect(
        contextualCreate.getSnapshot().entries[0]?.receipt?.data,
      ).toMatchObject({
        view_schema_id: taskRequestsViewSchemaId,
        row: {
          record_id: "20000000-0000-4000-8000-000000000401",
          row_version: 1,
        },
      }),
    );
    expect(screen.getByTestId(timelineInspectorTestId())).toBeTruthy();
  });

  it("creates related Evidence from the Timeline inspector and links it back through the Timeline patch route", async () => {
    const evidenceViewSchemaId = "cartulary.view.evidence.v1";
    const authority: WorkbookMutationAuthority = {
      actorId: "40000000-0000-4000-8000-000000000401",
      incidentId: "10000000-0000-4000-8000-000000000001",
      role: "editor",
      closed: false,
      sessionIdentity: "session",
    };
    const sourceRow = timelineRow({
      recordId: "20000000-0000-4000-8000-000000000001",
      rowVersion: 5,
      summary: "Create evidence source",
      captureState: "rough",
    });
    const targetRow = fullWorkbookViewRow(
      requireViewContract(evidenceViewSchemaId),
      "20000000-0000-4000-8000-000000000402",
      1,
      {},
    );
    const owner = new WorkbookTimelineRelatedEvidenceOwner(
      authority.incidentId,
      createBrowserSecureTransactionIdPort(),
      {
        coordinate: async () => ({
          kind: "settled" as const,
          minimumRowVersion: 0,
        }),
        accepted: () => {},
        refresh: async () => {},
        conflict: () => {},
      },
    );
    owner.setAuthority(authority);
    owner.configure(
      {
        availableViews: async () => ({
          kind: "accepted" as const,
          value: [timelineViewSchemaId, evidenceViewSchemaId],
        }),
        verify: async () => {},
        page: async ({ viewSchemaId }) => {
          const row =
            viewSchemaId === timelineViewSchemaId ? sourceRow : targetRow;
          return {
            kind: "accepted",
            value: {
              candidates: [
                {
                  recordId: row.record_id,
                  viewSchemaId,
                  displayText: "Record",
                  row,
                },
              ],
              hasMore: false,
              nextCursor: null,
            },
          };
        },
      },
      async () => authority,
      createTimelineRelatedEvidenceTransport(undefined),
    );
    fetchMock
      .mockResolvedValueOnce(
        timelineRowsEnvelope([
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 5,
            summary: "Create evidence source",
            captureState: "rough",
          }),
        ]),
      )
      .mockResolvedValueOnce(
        successEnvelope({
          view_schema_id: evidenceViewSchemaId,
          change_set_id: "30000000-0000-4000-8000-000000000001",
          row: targetRow,
        }),
      )
      .mockResolvedValueOnce(
        timelineMutationEnvelope(
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 6,
            summary: "Create evidence source",
            captureState: "rough",
            evidenceCount: 1,
            hasEvidence: true,
          }),
          "30000000-0000-4000-8000-000000000402",
        ),
      )
      .mockResolvedValueOnce(
        timelineRowsEnvelope([
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 6,
            summary: "Create evidence source",
            captureState: "rough",
            evidenceCount: 1,
            hasEvidence: true,
          }),
        ]),
      );

    const { container } = render(
      <TimelineRelatedEvidenceContext.Provider
        value={{
          owner,
          sheetRef: { kind: "view_schema", id: timelineViewSchemaId },
        }}
      >
        <TimelineWorkbookRuntimeFixture
          currentIncidentRole="editor"
          currentUserId={authority.actorId}
          incidentId={authority.incidentId}
        />
      </TimelineRelatedEvidenceContext.Provider>,
    );
    await waitForVisibleGridRowRecordIds(container, [
      "20000000-0000-4000-8000-000000000001",
    ]);
    fireEvent.contextMenu(
      screen.getByTestId(
        rowCellTestId(
          "20000000-0000-4000-8000-000000000001",
          "timeline.activity_synopsis_text",
        ),
      ),
      { clientX: 32, clientY: 48 },
    );
    fireEvent.click(
      screen.getByTestId(
        rowInspectButtonTestId("20000000-0000-4000-8000-000000000001"),
      ),
    );
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

    await waitFor(() =>
      expect(owner.getSnapshot().checkpoints[0]?.links[0]?.phase).toBe(
        "accepted",
      ),
    );
    expect(
      owner.getSnapshot().checkpoints[0]?.create.receipt?.data.row.record_id,
    ).toBe("20000000-0000-4000-8000-000000000402");
    expect(screen.getByTestId(timelineInspectorTestId())).toBeTruthy();
    const createCallIndex = fetchMock.mock.calls.findIndex(([url]) =>
      String(url).endsWith(
        `/api/v1/incidents/10000000-0000-4000-8000-000000000001/views/${evidenceViewSchemaId}/rows`,
      ),
    );
    expect(extractTimelineJSONBody(fetchMock, createCallIndex)).toMatchObject({
      "evidence.title": "Source row evidence",
    });
    const patchCallIndex = fetchMock.mock.calls.findIndex(
      ([url, init]) =>
        String(url).endsWith(
          "/api/v1/records/20000000-0000-4000-8000-000000000001",
        ) && init?.method === "PATCH",
    );
    expect(extractTimelineJSONBody(fetchMock, patchCallIndex)).toMatchObject({
      view_schema_id: timelineViewSchemaId,
      base_row_version: 5,
      changes: [
        {
          field_key: "timeline.attached_evidence_ids",
          action_payload: {
            kind: "collection_actions_v1",
            actions: [
              {
                op: "add_record_ref",
                linked_record_id: "20000000-0000-4000-8000-000000000402",
              },
            ],
          },
        },
      ],
    });
  });

  it("Verify history and rollback preview/action use public route contracts, preserve retained history, and render public error envelopes.", async () => {
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
              recordId: "20000000-0000-4000-8000-000000000001",
              rowVersion: 4,
              summary: `Workbook inspector rollback ${errorCode}`,
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
        <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
      );
      await waitForVisibleGridRowRecordIds(container, [
        "20000000-0000-4000-8000-000000000001",
      ]);
      fireEvent.contextMenu(
        screen.getByTestId(
          rowCellTestId(
            "20000000-0000-4000-8000-000000000001",
            "timeline.activity_synopsis_text",
          ),
        ),
        { clientX: 32, clientY: 48 },
      );
      fireEvent.click(
        screen.getByTestId(
          rowHistoryOpenButtonTestId("20000000-0000-4000-8000-000000000001"),
        ),
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
      fireEvent.click(
        await screen.findByTestId(rowHistoryActionTestId(actionAnchor)),
      );
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
        String(url).endsWith(
          "/api/v1/records/20000000-0000-4000-8000-000000000001/rollback",
        ),
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
