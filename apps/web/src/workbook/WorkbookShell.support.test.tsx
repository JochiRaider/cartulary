import {
  autoResolutionNoticeFamilySelector,
  autoResolutionNoticeTestId,
  autoResolutionUndoButtonTestId,
  gridActionsHeaderTestId,
  gridRowTestId,
  gridRowVersionAttribute,
  gridScrollportSelector,
  gridShellTestId,
  mentionItemTestId,
  relationshipChipTestId,
  relationshipItemsTestId,
  relationshipOverflowButtonTestId,
  rowCellTestId,
  rowHistoryOpenButtonTestId,
  rowInspectButtonTestId,
  saveStateTestId,
  timelineCollectionInputTestId,
  timelineInspectorSectionTestId,
  timelineInspectorTestId,
  timelineRowSupersedeButtonTestId,
  workbookInlineDraftRowTestId,
  workbookRowContextMenuTestId,
} from "@cartulary/ui-contracts";
import { hostsViewSchemaId } from "@cartulary/view-contracts";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../testing/fetchMockTestSupport";
import { TimelineWorkbookRuntimeFixture } from "../testing/TimelineWorkbookRuntimeFixture";
import {
  buildRecordChangedPayload,
  emitRecordChanged,
  extractTimelineJSONBody,
  successEnvelope,
  timelineRow,
  timelineRowsEnvelope,
} from "../testing/timelineWorkbookTestSupport";
import { buildMentionActionPayload } from "./collaboration/workbookCollaborationMessages";
import { timelineViewSchemaId } from "./models/workbookSurfaceRegistry";
import {
  buildAutoResolutionNotices,
  buildInspectorMentions,
  readCollectionItems,
} from "./timeline/models/workbookMentionChips";

vi.mock(
  "@cartulary/grid-adapter",
  async () => import("@cartulary/grid-adapter/test-support"),
);

// Support-only mocked component coverage for Record relationships workbook helpers.
// This file is not authoritative Record relationships evidence.
// Mention mutation/creation continuity is routed through the retained-owner and
// hook tests plus the real-service mentions.resolve/lifecycle/recovery scenarios.
describe("support workbook helpers", () => {
  it("reads manual and auto-resolved collection items without dropping confidence nulls", () => {
    const items = readCollectionItems(
      {
        cells: {
          "timeline.host_refs": {
            value: {
              items: [
                {
                  item_ref: "mention-host-auto",
                  entity_type: "host",
                  item_kind: "resolved_ref",
                  display_text: "VPN Gateway",
                  raw_text: " vpn   gateway ",
                  resolved_record_id: "20000000-0000-4000-8000-000000000602",
                  mention_row_version: 11,
                  resolution_method: "auto_match",
                  auto_resolved: true,
                  provenance: "auto_match",
                  confidence: 100,
                  matched_alias_text: "VPN Gateway",
                },
                {
                  item_ref: "mention-host-manual",
                  entity_type: "host",
                  item_kind: "resolved_ref",
                  display_text: "WS-023",
                  raw_text: "WS-023",
                  resolved_record_id: "20000000-0000-4000-8000-000000000603",
                  mention_row_version: 12,
                  resolution_method: "explicit_resolve_route",
                  auto_resolved: false,
                  provenance: "manual",
                  confidence: null,
                },
              ],
            },
          },
        },
      },
      "timeline.host_refs",
    );

    expect(items).toEqual([
      {
        itemRef: "mention-host-auto",
        entityType: "host",
        itemKind: "resolved_ref" as const,
        displayText: "VPN Gateway",
        rawText: " vpn   gateway ",
        resolvedRecordId: "20000000-0000-4000-8000-000000000602",
        mentionRowVersion: 11,
        resolutionMethod: "auto_match",
        autoResolved: true,
        provenance: "auto_match",
        confidence: 100,
        matchedAliasText: "VPN Gateway",
      },
      {
        itemRef: "mention-host-manual",
        entityType: "host",
        itemKind: "resolved_ref" as const,
        displayText: "WS-023",
        rawText: "WS-023",
        resolvedRecordId: "20000000-0000-4000-8000-000000000603",
        mentionRowVersion: 12,
        resolutionMethod: "explicit_resolve_route",
        autoResolved: false,
        provenance: "manual",
        confidence: null,
        matchedAliasText: null,
      },
    ]);
  });

  it("builds explicit mention route action payloads", () => {
    expect(
      buildMentionActionPayload(
        { mentionRowVersion: 13 },
        "resolve_item",
        "timeline-client-8",
        "20000000-0000-4000-8000-000000000602",
      ),
    ).toEqual({
      base_mention_row_version: 13,
      client_txn_id: "timeline-client-8",
      action: "resolve_item",
      resolved_record_id: "20000000-0000-4000-8000-000000000602",
    });
    expect(
      buildMentionActionPayload(
        { mentionRowVersion: 14 },
        "dismiss_item",
        "timeline-client-9",
      ),
    ).toEqual({
      base_mention_row_version: 14,
      client_txn_id: "timeline-client-9",
      action: "dismiss_item",
    });
    expect(
      buildMentionActionPayload(
        { mentionRowVersion: 15 },
        "revert_to_unresolved",
        "timeline-client-10",
      ),
    ).toEqual({
      base_mention_row_version: 15,
      client_txn_id: "timeline-client-10",
      action: "revert_to_unresolved",
    });
    expect(
      buildMentionActionPayload(
        { mentionRowVersion: null },
        "revert_to_unresolved",
        "timeline-client-11",
      ),
    ).toBeNull();
  });

  it("builds inspector mentions and auto-resolution notices from row deltas", () => {
    const beforeRow = {
      recordId: "20000000-0000-4000-8000-000000000601",
      collectionValues: {
        hostRefs: [],
        identityRefs: [
          {
            itemRef: "mention-identity-manual",
            entityMentionId: "identity-mention-public",
            entityType: "identity" as const,
            itemKind: "resolved_ref" as const,
            displayText: "Alex Analyst",
            rawText: "alex.analyst@example.test",
            resolvedRecordId: "20000000-0000-4000-8000-000000000604",
            mentionRowVersion: 21,
            resolutionMethod: "explicit_resolve_route",
            autoResolved: false,
            provenance: "manual",
            confidence: null,
            matchedAliasText: null,
          },
        ],
      },
    };
    const afterRow = {
      recordId: "20000000-0000-4000-8000-000000000601",
      collectionValues: {
        hostRefs: [
          {
            itemRef: "mention-host-auto",
            entityMentionId: "host-mention-public",
            entityType: "host" as const,
            itemKind: "resolved_ref" as const,
            displayText: "VPN Gateway",
            rawText: " vpn   gateway ",
            resolvedRecordId: "20000000-0000-4000-8000-000000000602",
            mentionRowVersion: 22,
            resolutionMethod: "auto_match",
            autoResolved: true,
            provenance: "auto_match",
            confidence: 100,
            matchedAliasText: "VPN Gateway",
          },
        ],
        identityRefs: beforeRow.collectionValues.identityRefs,
      },
    };

    expect(buildAutoResolutionNotices(beforeRow, afterRow)).toEqual([
      {
        itemRef: "mention-host-auto",
        rowRecordId: "20000000-0000-4000-8000-000000000601",
        fieldKey: "timeline.host_refs",
        entityType: "host",
        rawText: " vpn   gateway ",
        resolvedRecordId: "20000000-0000-4000-8000-000000000602",
        matchedAliasText: "VPN Gateway",
      },
    ]);

    expect(
      buildInspectorMentions(afterRow, [
        {
          rowRecordId: "20000000-0000-4000-8000-000000000601",
          fieldKey: "timeline.host_refs",
          entityType: "host",
          itemRef: "mention-host-dismissed",
          entityMentionId: "dismissed-host-public",
          rawText: "WS-023",
          resolvedRecordId: "20000000-0000-4000-8000-000000000603",
          mentionRowVersion: 23,
          resolutionMethod: "explicit_resolve_route",
          autoResolved: false,
        },
      ]),
    ).toEqual([
      {
        rowRecordId: "20000000-0000-4000-8000-000000000601",
        fieldKey: "timeline.host_refs",
        entityType: "host",
        itemRef: "mention-host-auto",
        rawText: " vpn   gateway ",
        resolvedRecordId: "20000000-0000-4000-8000-000000000602",
        mentionRowVersion: 22,
        resolutionMethod: "auto_match",
        autoResolved: true,
        status: "resolved",
        chipState: "auto_resolved",
        anchor: {
          recordId: "20000000-0000-4000-8000-000000000601",
          fieldKey: "timeline.host_refs",
          itemRef: "mention-host-auto",
          entityMentionId: null,
          targetEntityRecordId: "20000000-0000-4000-8000-000000000602",
        },
        sourceKind: "entity_mention",
        isActiveRelationshipValue: true,
        priorTargetEntityRecordId: null,
        displayText: "VPN Gateway",
        provenance: "auto_match",
        confidence: 100,
        matchedAliasText: "VPN Gateway",
      },
      {
        rowRecordId: "20000000-0000-4000-8000-000000000601",
        fieldKey: "timeline.identity_refs",
        entityType: "identity",
        itemRef: "mention-identity-manual",
        rawText: "alex.analyst@example.test",
        resolvedRecordId: "20000000-0000-4000-8000-000000000604",
        mentionRowVersion: 21,
        resolutionMethod: "explicit_resolve_route",
        autoResolved: false,
        status: "resolved",
        chipState: "resolved",
        anchor: {
          recordId: "20000000-0000-4000-8000-000000000601",
          fieldKey: "timeline.identity_refs",
          itemRef: "mention-identity-manual",
          entityMentionId: null,
          targetEntityRecordId: "20000000-0000-4000-8000-000000000604",
        },
        sourceKind: "entity_mention",
        isActiveRelationshipValue: true,
        priorTargetEntityRecordId: null,
        displayText: "Alex Analyst",
        provenance: "manual",
        confidence: null,
        matchedAliasText: null,
      },
      {
        rowRecordId: "20000000-0000-4000-8000-000000000601",
        fieldKey: "timeline.host_refs",
        entityType: "host",
        itemRef: "mention-host-dismissed",
        rawText: "WS-023",
        resolvedRecordId: null,
        mentionRowVersion: 23,
        resolutionMethod: "explicit_resolve_route",
        autoResolved: false,
        status: "dismissed",
        chipState: "dismissed",
        anchor: {
          recordId: "20000000-0000-4000-8000-000000000601",
          fieldKey: "timeline.host_refs",
          itemRef: "mention-host-dismissed",
          entityMentionId: null,
          targetEntityRecordId: null,
        },
        sourceKind: "entity_mention",
        isActiveRelationshipValue: false,
        priorTargetEntityRecordId: "20000000-0000-4000-8000-000000000603",
        displayText: "WS-023",
        provenance: null,
        confidence: null,
        matchedAliasText: null,
      },
    ]);
  });
});

describe("support TimelineWorkbookRuntimeFixture", () => {
  let fetchMock: ReturnType<typeof vi.fn>;
  let webSocketInstance: {
    onmessage: ((event: MessageEvent) => void) | null;
    close: () => void;
  } | null;

  beforeEach(() => {
    fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    webSocketInstance = null;
    Object.defineProperty(window, "focus", {
      configurable: true,
      value: vi.fn(),
    });
    vi.stubGlobal(
      "WebSocket",
      class {
        onmessage: ((event: MessageEvent) => void) | null = null;

        constructor() {
          webSocketInstance = this;
        }

        close() {}
      } as unknown as typeof WebSocket,
    );
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("opens Timeline row actions from the committed row context menu", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000601",
            rowVersion: 2,
            summary: "Alpha",
            captureState: "rough",
          }),
        ],
      }),
    );

    render(
      <TimelineWorkbookRuntimeFixture
        incidentId="10000000-0000-4000-8000-000000000001"
        currentIncidentRole="admin"
      />,
    );

    const summaryCell = await screen.findByTestId(
      rowCellTestId(
        "20000000-0000-4000-8000-000000000601",
        "timeline.activity_synopsis_text",
      ),
    );
    expect(
      screen.queryByTestId(gridActionsHeaderTestId(timelineViewSchemaId)),
    ).toBeNull();

    fireEvent.contextMenu(summaryCell, { clientX: 32, clientY: 48 });
    const contextMenu = await screen.findByTestId(
      workbookRowContextMenuTestId(
        timelineViewSchemaId,
        "20000000-0000-4000-8000-000000000601",
      ),
    );
    expect(contextMenu.getAttribute("role")).toBe("dialog");
    const inspectButton = screen.getByTestId(
      rowInspectButtonTestId("20000000-0000-4000-8000-000000000601"),
    );
    expect(document.activeElement).toBe(inspectButton);
    fireEvent.keyDown(inspectButton, { key: "ArrowDown" });
    expect(document.activeElement).toBe(
      screen.getByTestId(
        rowHistoryOpenButtonTestId("20000000-0000-4000-8000-000000000601"),
      ),
    );
    const supersedeShortcut = screen.getByTestId(
      timelineRowSupersedeButtonTestId("20000000-0000-4000-8000-000000000601"),
    );
    supersedeShortcut.focus();
    fireEvent.scroll(window);
    expect(
      screen.getByTestId(
        workbookRowContextMenuTestId(
          timelineViewSchemaId,
          "20000000-0000-4000-8000-000000000601",
        ),
      ),
    ).toBeTruthy();

    supersedeShortcut.blur();
    fireEvent.scroll(window);
    await waitFor(() => {
      expect(
        screen.queryByTestId(
          workbookRowContextMenuTestId(
            timelineViewSchemaId,
            "20000000-0000-4000-8000-000000000601",
          ),
        ),
      ).toBeNull();
    });
    fireEvent.contextMenu(summaryCell, { clientX: 32, clientY: 48 });
    const reopenedContextMenu = await screen.findByTestId(
      workbookRowContextMenuTestId(
        timelineViewSchemaId,
        "20000000-0000-4000-8000-000000000601",
      ),
    );

    fireEvent.keyDown(reopenedContextMenu, { key: "Escape" });
    await waitFor(() => {
      expect(
        screen.queryByTestId(
          workbookRowContextMenuTestId(
            timelineViewSchemaId,
            "20000000-0000-4000-8000-000000000601",
          ),
        ),
      ).toBeNull();
    });
    expect(document.activeElement).toBe(
      screen.getByLabelText("Timeline row interaction layer"),
    );
  });

  it("opens Timeline row actions from the keyboard and ignores draft rows", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000601",
            rowVersion: 2,
            summary: "Alpha",
            captureState: "rough",
          }),
        ],
      }),
    );

    render(
      <TimelineWorkbookRuntimeFixture
        incidentId="10000000-0000-4000-8000-000000000001"
        currentIncidentRole="admin"
      />,
    );

    const draftRow = await screen.findByTestId(
      workbookInlineDraftRowTestId(timelineViewSchemaId),
    );
    fireEvent.contextMenu(draftRow, { clientX: 12, clientY: 24 });
    expect(
      screen.queryByTestId(
        workbookRowContextMenuTestId(
          timelineViewSchemaId,
          "20000000-0000-4000-8000-000000000601",
        ),
      ),
    ).toBeNull();

    const summaryCell = await screen.findByTestId(
      rowCellTestId(
        "20000000-0000-4000-8000-000000000601",
        "timeline.activity_synopsis_text",
      ),
    );
    summaryCell.focus();
    fireEvent.keyDown(summaryCell, { key: "F10", shiftKey: true });
    await screen.findByTestId(
      workbookRowContextMenuTestId(
        timelineViewSchemaId,
        "20000000-0000-4000-8000-000000000601",
      ),
    );
    expect(document.activeElement).toBe(
      screen.getByTestId(
        rowInspectButtonTestId("20000000-0000-4000-8000-000000000601"),
      ),
    );
  });

  it("renders auto-resolved chips distinctly from manual resolved chips", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000601",
            rowVersion: 2,
            summary: "Alpha",
            captureState: "reviewed",
            hostRefs: [
              resolvedItem({
                itemRef: "mention-host-auto",
                entityMentionId: "host-mention-public",
                entityType: "host",
                rawText: " vpn   gateway ",
                displayText: "VPN Gateway",
                resolvedRecordId: "20000000-0000-4000-8000-000000000602",
                resolutionMethod: "auto_match",
                autoResolved: true,
                provenance: "auto_match",
                confidence: 100,
                matchedAliasText: "VPN Gateway",
              }),
            ],
            identityRefs: [
              resolvedItem({
                itemRef: "mention-identity-manual",
                entityMentionId: "identity-mention-public",
                entityType: "identity",
                rawText: "alex.analyst@example.test",
                displayText: "Alex Analyst",
                resolvedRecordId: "20000000-0000-4000-8000-000000000604",
                resolutionMethod: "explicit_resolve_route",
                autoResolved: false,
                provenance: "manual",
                confidence: null,
              }),
            ],
          }),
        ],
      }),
    );

    render(
      <TimelineWorkbookRuntimeFixture
        incidentId="10000000-0000-4000-8000-000000000001"
        currentIncidentRole="admin"
      />,
    );

    await openTimelineInspectorFromContext(
      "20000000-0000-4000-8000-000000000601",
    );
    const autoChips = await screen.findAllByTestId(
      relationshipChipTestId("mention-host-auto"),
    );
    const manualChips = screen.getAllByTestId(
      relationshipChipTestId("mention-identity-manual"),
    );
    const autoChip = autoChips[0];
    const manualChip = manualChips[0];
    if (!autoChip || !manualChip) {
      throw new Error("Expected relationship chips to render in the inspector");
    }
    expect(autoChip.textContent).toContain("auto");
    expect(manualChip.textContent).not.toContain("auto");
    fireEvent.click(
      screen.getByTestId(mentionItemTestId("mention-identity-manual")),
    );
    expect(screen.getByText("Manual")).toBeTruthy();
  });

  it("reveals a clipped inspect action after an auto-resolution collection patch", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000601",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "reviewed",
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000603",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000601",
          rowVersion: 2,
          summary: "Alpha",
          captureState: "reviewed",
          hostRefs: [
            resolvedItem({
              itemRef: "mention-host-auto",
              entityMentionId: "host-mention-public",
              entityType: "host",
              rawText: " vpn   gateway ",
              displayText: "Gateway node",
              resolvedRecordId: "20000000-0000-4000-8000-000000000602",
              resolutionMethod: "auto_match",
              autoResolved: true,
              provenance: "auto_match",
              confidence: 100,
              matchedAliasText: "VPN Gateway",
            }),
          ],
        }),
      }),
    );

    render(
      <TimelineWorkbookRuntimeFixture
        incidentId="10000000-0000-4000-8000-000000000001"
        currentIncidentRole="admin"
      />,
    );

    await openTimelineInspectorFromContext(
      "20000000-0000-4000-8000-000000000601",
    );
    const relationshipInput = (await screen.findByTestId(
      timelineCollectionInputTestId(
        "20000000-0000-4000-8000-000000000601",
        "timeline.host_refs",
      ),
    )) as HTMLInputElement;
    installTimelineInspectGeometry("20000000-0000-4000-8000-000000000601", {
      containerHeight: 300,
      containerLeft: 40,
      containerTop: 100,
      containerWidth: 400,
      contentLeft: 368,
      contentTop: 310,
      targetHeight: 40,
      targetWidth: 80,
    });

    const preservedScroll = setTimelineGridScroll(240, 18);

    expect(
      isTimelineFocusTargetFullyVisibleWithinGrid(
        "20000000-0000-4000-8000-000000000601",
      ),
    ).toBe(false);

    fireEvent.change(relationshipInput, {
      target: { value: " vpn   gateway " },
    });
    fireEvent.keyDown(relationshipInput, { key: "Enter" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    await expectTimelineFocusAndScroll(
      "20000000-0000-4000-8000-000000000601",
      preservedScroll,
      {
        expectedLeft: 48,
        requireVisibleWithinGrid: true,
      },
    );
    expect(
      await screen.findByTestId(
        autoResolutionNoticeTestId("mention-host-auto"),
      ),
    ).toBeTruthy();
  });

  it("sends auto-resolution Undo with the current post-resolution row version", async () => {
    const mentionItemRef =
      "entity_mention:11111111-1111-4111-8111-000000000404";
    let committed = false;
    const autoRow = timelineRow({
      recordId: "20000000-0000-4000-8000-000000000601",
      rowVersion: 2,
      summary: "Alpha",
      captureState: "reviewed",
      hostRefs: [
        resolvedItem({
          itemRef: mentionItemRef,
          entityMentionId: "11111111-1111-4111-8111-000000000404",
          entityType: "host",
          rawText: " vpn   gateway ",
          displayText: "Gateway node",
          resolvedRecordId: "20000000-0000-4000-8000-000000000602",
          resolutionMethod: "auto_match",
          autoResolved: true,
          provenance: "auto_match",
          confidence: 100,
        }),
      ],
    });
    const restored = timelineRow({
      recordId: autoRow.record_id,
      rowVersion: 3,
      summary: "Alpha",
      captureState: "reviewed",
      hostRefs: [
        unresolvedItem({
          itemRef: mentionItemRef,
          entityMentionId: "11111111-1111-4111-8111-000000000404",
          entityType: "host",
          rawText: " vpn   gateway ",
          mentionRowVersion: 2,
        }),
      ],
    });
    fetchMock.mockImplementation(async (url, init) => {
      if (String(url).includes("/entity-mentions/")) {
        committed = true;
        return mentionActionEnvelope({
          actionStatus: "unresolved",
          mentionId: "11111111-1111-4111-8111-000000000404",
          rawText: " vpn   gateway ",
          resolvedRecordId: null,
          sourceRowVersion: 3,
          mentionRowVersion: 2,
          resolutionMethod: null,
        });
      }
      if (init?.method === "PATCH")
        return successEnvelope({
          view_schema_id: timelineViewSchemaId,
          change_set_id: "30000000-0000-4000-8000-000000000603",
          row: autoRow,
        });
      if (String(url).includes(hostsViewSchemaId))
        return successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          view_schema_id: hostsViewSchemaId,
          rows: [],
        });
      return timelineRowsEnvelope([
        committed
          ? restored
          : timelineRow({
              recordId: autoRow.record_id,
              rowVersion: 1,
              summary: "Alpha",
              captureState: "reviewed",
            }),
      ]);
    });

    render(
      <TimelineWorkbookRuntimeFixture
        incidentId="10000000-0000-4000-8000-000000000001"
        currentIncidentRole="admin"
      />,
    );

    await openTimelineInspectorFromContext(
      "20000000-0000-4000-8000-000000000601",
    );
    const relationshipInput = (await screen.findByTestId(
      timelineCollectionInputTestId(
        "20000000-0000-4000-8000-000000000601",
        "timeline.host_refs",
      ),
    )) as HTMLInputElement;
    fireEvent.change(relationshipInput, {
      target: { value: " vpn   gateway " },
    });
    fireEvent.keyDown(relationshipInput, { key: "Enter" });

    const notice = await screen.findByTestId(
      autoResolutionNoticeTestId(mentionItemRef),
    );
    const undoButton = within(notice).getByTestId(
      autoResolutionUndoButtonTestId(mentionItemRef),
    );
    const preservedScroll = setTimelineGridScroll(240, 140);
    undoButton.focus();
    expect(document.activeElement).toBe(undoButton);
    fireEvent.click(undoButton);

    await expectTimelineFocusAndScroll(
      "20000000-0000-4000-8000-000000000601",
      preservedScroll,
    );
    fireEvent.click(screen.getByRole("button", { name: /^Recovery \(/ }));
    fireEvent.click(
      await screen.findByRole("button", { name: /Mention resolution ·/ }),
    );
    await waitFor(() => {
      expect(
        screen.getByLabelText("Retained mention operations").textContent,
      ).toContain("Mention action completed.");
    });
    expect(
      screen.queryByTestId(autoResolutionNoticeTestId(mentionItemRef)),
    ).toBeNull();
    const request = fetchMock.mock.calls.find((call) =>
      String(call[0]).includes("/entity-mentions/"),
    );
    expect(String(request?.[0])).toContain(
      "/api/v1/entity-mentions/11111111-1111-4111-8111-000000000404/resolve",
    );
    expect(JSON.parse(request?.[1].body)).toMatchObject({
      base_mention_row_version: 1,
      action: "revert_to_unresolved",
    });
  });

  it("keeps the workbook mounted after committing a relationship-cell edit", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000601",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "reviewed",
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000605",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000601",
          rowVersion: 2,
          summary: "Alpha",
          captureState: "reviewed",
          hostRefs: [
            unresolvedItem({
              itemRef: "mention-host-new",
              entityType: "host",
              rawText: "WS-023",
            }),
          ],
        }),
      }),
    );

    render(
      <TimelineWorkbookRuntimeFixture
        incidentId="10000000-0000-4000-8000-000000000001"
        currentIncidentRole="admin"
      />,
    );

    await openTimelineInspectorFromContext(
      "20000000-0000-4000-8000-000000000601",
    );
    const relationshipInput = (await screen.findByTestId(
      timelineCollectionInputTestId(
        "20000000-0000-4000-8000-000000000601",
        "timeline.host_refs",
      ),
    )) as HTMLInputElement;
    fireEvent.change(relationshipInput, { target: { value: "WS-023" } });
    fireEvent.keyDown(relationshipInput, { key: "Enter" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    await waitFor(() => {
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
    });

    expect(
      screen.getByTestId(gridShellTestId(timelineViewSchemaId)),
    ).toBeTruthy();
    expect(
      screen.getByTestId(
        timelineCollectionInputTestId(
          "20000000-0000-4000-8000-000000000601",
          "timeline.host_refs",
        ),
      ),
    ).toBeTruthy();
    expect(
      screen.getByTestId(
        relationshipItemsTestId(
          "20000000-0000-4000-8000-000000000601",
          "timeline.host_refs",
        ),
      ).textContent,
    ).toContain("WS-023");
    fireEvent.contextMenu(
      screen.getByTestId(
        rowCellTestId(
          "20000000-0000-4000-8000-000000000601",
          "timeline.activity_synopsis_text",
        ),
      ),
      { clientX: 32, clientY: 48 },
    );
    expect(
      screen.getByTestId(
        rowInspectButtonTestId("20000000-0000-4000-8000-000000000601"),
      ),
    ).toBeTruthy();
  });

  it("commits sequential hostRefs edits with the latest returned row versions", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000601",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "reviewed",
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000606",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000601",
          rowVersion: 2,
          summary: "Alpha",
          captureState: "reviewed",
          hostRefs: [
            unresolvedItem({
              itemRef: "mention-20000000-0000-4000-8000-000000000602",
              entityType: "host",
              rawText: "WS-023",
            }),
          ],
        }),
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000607",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000601",
          rowVersion: 3,
          summary: "Alpha",
          captureState: "reviewed",
          hostRefs: [
            unresolvedItem({
              itemRef: "mention-20000000-0000-4000-8000-000000000602",
              entityType: "host",
              rawText: "WS-023",
            }),
            unresolvedItem({
              itemRef: "mention-20000000-0000-4000-8000-000000000603",
              entityType: "host",
              rawText: "WS-024",
            }),
          ],
        }),
      }),
    );

    render(
      <TimelineWorkbookRuntimeFixture
        incidentId="10000000-0000-4000-8000-000000000001"
        currentIncidentRole="admin"
      />,
    );

    await openTimelineInspectorFromContext(
      "20000000-0000-4000-8000-000000000601",
    );
    const firstInput = (await screen.findByTestId(
      timelineCollectionInputTestId(
        "20000000-0000-4000-8000-000000000601",
        "timeline.host_refs",
      ),
    )) as HTMLInputElement;
    fireEvent.change(firstInput, { target: { value: "WS-023" } });
    fireEvent.keyDown(firstInput, { key: "Enter" });
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
      expect(
        screen
          .getByTestId(
            gridRowTestId(
              timelineViewSchemaId,
              "20000000-0000-4000-8000-000000000601",
            ),
          )
          .getAttribute(gridRowVersionAttribute),
      ).toBe("2");
    });

    const secondInput = screen.getByTestId(
      timelineCollectionInputTestId(
        "20000000-0000-4000-8000-000000000601",
        "timeline.host_refs",
      ),
    ) as HTMLInputElement;
    fireEvent.change(secondInput, { target: { value: "WS-024" } });
    fireEvent.keyDown(secondInput, { key: "Enter" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
      expect(
        screen
          .getByTestId(
            gridRowTestId(
              timelineViewSchemaId,
              "20000000-0000-4000-8000-000000000601",
            ),
          )
          .getAttribute(gridRowVersionAttribute),
      ).toBe("3");
    });

    expectHostRefAddRequest(fetchMock, 1, 1, "WS-023");
    expectHostRefAddRequest(fetchMock, 2, 2, "WS-024");
    expect(
      screen.getByTestId(
        relationshipItemsTestId(
          "20000000-0000-4000-8000-000000000601",
          "timeline.host_refs",
        ),
      ).textContent,
    ).toContain("WS-023");
    expect(
      screen.getByTestId(
        relationshipItemsTestId(
          "20000000-0000-4000-8000-000000000601",
          "timeline.host_refs",
        ),
      ).textContent,
    ).toContain("WS-024");
    expect(
      document.querySelector(autoResolutionNoticeFamilySelector()),
    ).toBeNull();
  });

  it("treats Enter followed by blur as one hostRefs collection commit", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000601",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "reviewed",
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000606",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000601",
          rowVersion: 2,
          summary: "Alpha",
          captureState: "reviewed",
          hostRefs: [
            unresolvedItem({
              itemRef: "mention-20000000-0000-4000-8000-000000000602",
              entityType: "host",
              rawText: "WS-023",
            }),
          ],
        }),
      }),
    );

    render(
      <TimelineWorkbookRuntimeFixture
        incidentId="10000000-0000-4000-8000-000000000001"
        currentIncidentRole="admin"
      />,
    );

    await openTimelineInspectorFromContext(
      "20000000-0000-4000-8000-000000000601",
    );
    const relationshipInput = (await screen.findByTestId(
      timelineCollectionInputTestId(
        "20000000-0000-4000-8000-000000000601",
        "timeline.host_refs",
      ),
    )) as HTMLInputElement;
    fireEvent.change(relationshipInput, { target: { value: "WS-023" } });
    fireEvent.keyDown(relationshipInput, { key: "Enter" });
    fireEvent.blur(relationshipInput);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
    });
    await new Promise((resolve) => window.setTimeout(resolve, 50));

    expect(fetchMock).toHaveBeenCalledTimes(2);
    expectHostRefAddRequest(fetchMock, 1, 1, "WS-023");
  });

  it("dispatches a queued second hostRefs edit with the first response row version", async () => {
    const firstPatch = deferred<Response>();
    const secondPatch = deferred<Response>();
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000601",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "reviewed",
          }),
        ],
      }),
    );
    fetchMock.mockReturnValueOnce(firstPatch.promise);
    fetchMock.mockReturnValueOnce(secondPatch.promise);

    render(
      <TimelineWorkbookRuntimeFixture
        incidentId="10000000-0000-4000-8000-000000000001"
        currentIncidentRole="admin"
      />,
    );

    await openTimelineInspectorFromContext(
      "20000000-0000-4000-8000-000000000601",
    );
    const relationshipInput = (await screen.findByTestId(
      timelineCollectionInputTestId(
        "20000000-0000-4000-8000-000000000601",
        "timeline.host_refs",
      ),
    )) as HTMLInputElement;
    fireEvent.change(relationshipInput, { target: { value: "WS-023" } });
    fireEvent.keyDown(relationshipInput, { key: "Enter" });
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Syncing");
    });

    fireEvent.change(relationshipInput, { target: { value: "WS-024" } });
    fireEvent.keyDown(relationshipInput, { key: "Enter" });
    await new Promise((resolve) => window.setTimeout(resolve, 0));
    expect(fetchMock).toHaveBeenCalledTimes(2);

    firstPatch.resolve(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000606",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000601",
          rowVersion: 2,
          summary: "Alpha",
          captureState: "reviewed",
          hostRefs: [
            unresolvedItem({
              itemRef: "mention-20000000-0000-4000-8000-000000000602",
              entityType: "host",
              rawText: "WS-023",
            }),
          ],
        }),
      }),
    );

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });

    expectHostRefAddRequest(fetchMock, 1, 1, "WS-023");
    expectHostRefAddRequest(fetchMock, 2, 2, "WS-024");

    secondPatch.resolve(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000607",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000601",
          rowVersion: 3,
          summary: "Alpha",
          captureState: "reviewed",
          hostRefs: [
            unresolvedItem({
              itemRef: "mention-20000000-0000-4000-8000-000000000602",
              entityType: "host",
              rawText: "WS-023",
            }),
            unresolvedItem({
              itemRef: "mention-20000000-0000-4000-8000-000000000603",
              entityType: "host",
              rawText: "WS-024",
            }),
          ],
        }),
      }),
    );

    await waitFor(() => {
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
      expect(
        screen.getByTestId(
          relationshipItemsTestId(
            "20000000-0000-4000-8000-000000000601",
            "timeline.host_refs",
          ),
        ).textContent,
      ).toContain("WS-024");
    });
  });

  it("opens the Relationships inspector from compact relationship overflow", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000601",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "reviewed",
            hostRefs: [
              unresolvedItem({
                itemRef: "mention-20000000-0000-4000-8000-000000000602",
                entityType: "host",
                rawText: "WS-023",
              }),
              unresolvedItem({
                itemRef: "mention-20000000-0000-4000-8000-000000000603",
                entityType: "host",
                rawText: "WS-024",
              }),
            ],
          }),
        ],
      }),
    );

    render(
      <TimelineWorkbookRuntimeFixture
        incidentId="10000000-0000-4000-8000-000000000001"
        currentIncidentRole="admin"
      />,
    );

    fireEvent.click(await screen.findByRole("button", { name: "Columns" }));
    fireEvent.click(screen.getByRole("menuitemcheckbox", { name: "Hosts" }));
    fireEvent.keyDown(screen.getByRole("menu", { name: "Column controls" }), {
      key: "Escape",
    });
    const overflowButton = await screen.findByTestId(
      relationshipOverflowButtonTestId(
        "20000000-0000-4000-8000-000000000601",
        "timeline.host_refs",
      ),
    );
    expect(overflowButton.textContent).toBe("+1");

    fireEvent.click(overflowButton);

    await waitFor(() => {
      expect(screen.getByTestId(timelineInspectorTestId())).toBeTruthy();
      expect(
        screen.getByTestId(timelineInspectorSectionTestId("relationships"))
          .textContent,
      ).toContain("WS-024");
    });
    await waitFor(() => {
      expect(document.activeElement).toBe(
        screen.getByTestId(
          mentionItemTestId("mention-20000000-0000-4000-8000-000000000603"),
        ),
      );
    });
    fireEvent.click(screen.getByRole("button", { name: "Close inspector" }));
    await waitFor(() => expect(document.activeElement).toBe(overflowButton));
  });

  it("suppresses self-originated websocket invalidations and reloads for external ones", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000601",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "reviewed",
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000608",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000601",
          rowVersion: 2,
          summary: "Alpha",
          captureState: "reviewed",
          hostRefs: [
            unresolvedItem({
              itemRef: "mention-host-new",
              entityType: "host",
              rawText: "WS-023",
            }),
          ],
        }),
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000601",
            rowVersion: 3,
            summary: "Alpha",
            captureState: "reviewed",
            hostRefs: [
              unresolvedItem({
                itemRef: "mention-host-new",
                entityType: "host",
                rawText: "WS-023",
              }),
            ],
          }),
        ],
      }),
    );

    render(
      <TimelineWorkbookRuntimeFixture
        incidentId="10000000-0000-4000-8000-000000000001"
        currentIncidentRole="admin"
      />,
    );

    await openTimelineInspectorFromContext(
      "20000000-0000-4000-8000-000000000601",
    );
    const relationshipInput = (await screen.findByTestId(
      timelineCollectionInputTestId(
        "20000000-0000-4000-8000-000000000601",
        "timeline.host_refs",
      ),
    )) as HTMLInputElement;
    fireEvent.change(relationshipInput, { target: { value: "WS-023" } });
    fireEvent.blur(relationshipInput);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    const submittedClientTxnId = String(
      extractTimelineJSONBody(fetchMock, 1).client_txn_id,
    );

    emitRecordChanged(
      webSocketInstance,
      buildRecordChangedPayload({
        recordId: "20000000-0000-4000-8000-000000000601",
        rowVersion: 2,
        clientTxnId: submittedClientTxnId,
      }),
    );

    await waitFor(
      () => {
        expect(fetchMock).toHaveBeenCalledTimes(2);
      },
      { timeout: 200 },
    );

    emitRecordChanged(
      webSocketInstance,
      buildRecordChangedPayload({
        recordId: "20000000-0000-4000-8000-000000000601",
        rowVersion: 3,
        clientTxnId: "someone-else",
      }),
    );

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
  });
});

function resolvedItem({
  entityMentionId = null,
  itemRef,
  entityType,
  rawText,
  displayText,
  resolvedRecordId,
  mentionRowVersion = 1,
  resolutionMethod,
  autoResolved,
  provenance,
  confidence,
  matchedAliasText,
}: {
  itemRef: string;
  entityMentionId?: string | null;
  entityType: "host" | "identity";
  rawText: string;
  displayText: string;
  resolvedRecordId: string;
  mentionRowVersion?: number;
  resolutionMethod: string;
  autoResolved: boolean;
  provenance: string;
  confidence: number | null;
  matchedAliasText?: string;
}) {
  return {
    item_ref: itemRef,
    entity_mention_id: entityMentionId,
    entity_type: entityType,
    item_kind: "resolved_ref",
    display_text: displayText,
    raw_text: rawText,
    resolved_record_id: resolvedRecordId,
    mention_row_version: mentionRowVersion,
    resolution_method: resolutionMethod,
    auto_resolved: autoResolved,
    provenance,
    confidence,
    matched_alias_text: matchedAliasText,
  };
}

function mentionActionEnvelope({
  actionStatus,
  entityType = "host",
  mentionId,
  rawText,
  resolvedRecordId,
  sourceFieldKey = "timeline.host_refs",
  sourceRowVersion,
  mentionRowVersion,
  resolutionMethod,
}: {
  actionStatus: "dismissed" | "resolved" | "unresolved";
  entityType?: "host" | "identity";
  mentionId: string;
  rawText: string;
  resolvedRecordId: string | null;
  sourceFieldKey?: "timeline.host_refs" | "timeline.identity_refs";
  sourceRowVersion: number;
  mentionRowVersion: number;
  resolutionMethod: string | null;
}) {
  return successEnvelope({
    incident_id: "10000000-0000-4000-8000-000000000001",
    entity_mention: {
      entity_mention_id: mentionId,
      source_record_id: "20000000-0000-4000-8000-000000000601",
      source_field_key: sourceFieldKey,
      entity_type: entityType,
      raw_text: rawText,
      resolution_status: actionStatus,
      resolved_record_id: resolvedRecordId,
      row_version: mentionRowVersion,
      resolution_method: actionStatus === "resolved" ? resolutionMethod : null,
      normalized_text: rawText.trim(),
      resolved_at: actionStatus === "resolved" ? "2026-09-12T04:00:00Z" : null,
      resolved_by_user_id: null,
    },
    source_record: {
      record_id: "20000000-0000-4000-8000-000000000601",
      row_version: sourceRowVersion,
    },
    change_set_id:
      actionStatus === "dismissed"
        ? "30000000-0000-4000-8000-000000000609"
        : actionStatus === "resolved"
          ? "30000000-0000-4000-8000-000000000610"
          : "30000000-0000-4000-8000-000000000611",
  });
}

function unresolvedItem({
  itemRef,
  entityMentionId = null,
  entityType,
  rawText,
  mentionRowVersion = 1,
}: {
  itemRef: string;
  entityMentionId?: string | null;
  entityType: "host" | "identity";
  rawText: string;
  mentionRowVersion?: number;
}) {
  return {
    item_ref: itemRef,
    entity_mention_id: entityMentionId,
    entity_type: entityType,
    item_kind: "unresolved_mention",
    display_text: rawText,
    raw_text: rawText,
    mention_row_version: mentionRowVersion,
  };
}

function expectHostRefAddRequest(
  fetchMock: ReturnType<typeof vi.fn>,
  index: number,
  baseRowVersion: number,
  rawText: string,
) {
  expect(extractTimelineJSONBody(fetchMock, index)).toMatchObject({
    view_schema_id: timelineViewSchemaId,
    base_row_version: baseRowVersion,
    changes: [
      {
        field_key: "timeline.host_refs",
        action_payload: {
          kind: "collection_actions_v1",
          actions: [
            {
              op: "add_token",
              raw_text: rawText,
            },
          ],
        },
      },
    ],
  });
}

function setTimelineGridScroll(top: number, left: number) {
  const grid = timelineGridScrollport();
  grid.scrollTop = top;
  grid.scrollLeft = left;
  return { top, left };
}

async function openTimelineInspectorFromContext(recordId: string) {
  const summaryCell = await screen.findByTestId(
    rowCellTestId(recordId, "timeline.activity_synopsis_text"),
  );
  fireEvent.contextMenu(summaryCell, { clientX: 32, clientY: 48 });
  fireEvent.click(await screen.findByTestId(rowInspectButtonTestId(recordId)));
}

async function expectTimelineFocusAndScroll(
  recordId: string,
  _preservedScroll: { top: number; left: number },
  options: {
    expectedLeft?: number | null;
    expectedTop?: number | null;
    requireVisibleWithinGrid?: boolean;
  } = {},
) {
  await waitFor(() => {
    expect(document.activeElement).toBe(
      screen
        .getByTestId(rowCellTestId(recordId, "timeline.activity_synopsis_text"))
        .closest('[role="gridcell"]'),
    );
    const grid = timelineGridScrollport();
    if (typeof options.expectedTop === "number") {
      expect(grid.scrollTop).toBe(options.expectedTop);
    }
    if (typeof options.expectedLeft === "number") {
      expect(grid.scrollLeft).toBe(options.expectedLeft);
    }
    if (options.requireVisibleWithinGrid) {
      expect(isTimelineFocusTargetFullyVisibleWithinGrid(recordId)).toBe(true);
    }
  });
}

function installTimelineInspectGeometry(
  recordId: string,
  options: {
    containerHeight: number;
    containerLeft: number;
    containerTop: number;
    containerWidth: number;
    contentLeft: number;
    contentTop: number | (() => number);
    targetHeight: number;
    targetWidth: number;
  },
) {
  const focusTargetTestId = rowCellTestId(
    recordId,
    "timeline.activity_synopsis_text",
  );
  const original = HTMLElement.prototype.getBoundingClientRect;

  vi.spyOn(HTMLElement.prototype, "getBoundingClientRect").mockImplementation(
    function mockRect(this: HTMLElement) {
      const testId = this.getAttribute("data-testid");
      const isFocusGridCell =
        this.getAttribute("role") === "gridcell" &&
        Array.from(this.querySelectorAll<HTMLElement>("[data-testid]")).some(
          (element) =>
            element.getAttribute("data-testid") === focusTargetTestId,
        );
      if (this.matches(gridScrollportSelector())) {
        return rectFromBox({
          height: options.containerHeight,
          left: options.containerLeft,
          top: options.containerTop,
          width: options.containerWidth,
        });
      }
      if (testId === focusTargetTestId || isFocusGridCell) {
        const grid = timelineGridScrollport();
        const contentTop =
          typeof options.contentTop === "function"
            ? options.contentTop()
            : options.contentTop;
        return rectFromBox({
          height: options.targetHeight,
          left: options.containerLeft + options.contentLeft - grid.scrollLeft,
          top: options.containerTop + contentTop - grid.scrollTop,
          width: options.targetWidth,
        });
      }
      return original.call(this);
    },
  );
}

function isTimelineFocusTargetFullyVisibleWithinGrid(recordId: string) {
  const tolerancePx = 1;
  const grid = timelineGridScrollport();
  const focusTarget = screen.getByTestId(
    rowCellTestId(recordId, "timeline.activity_synopsis_text"),
  );
  const gridRect = grid.getBoundingClientRect();
  const inspectRect = focusTarget.getBoundingClientRect();

  return (
    inspectRect.top >= gridRect.top - tolerancePx &&
    inspectRect.left >= gridRect.left - tolerancePx &&
    inspectRect.bottom <= gridRect.bottom + tolerancePx &&
    inspectRect.right <= gridRect.right + tolerancePx
  );
}

function timelineGridScrollport() {
  const grid = screen
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .querySelector(gridScrollportSelector());
  if (!(grid instanceof HTMLDivElement)) {
    throw new Error("Expected timeline grid scrollport to exist");
  }
  return grid;
}

function rectFromBox(options: {
  height: number;
  left: number;
  top: number;
  width: number;
}) {
  return {
    bottom: options.top + options.height,
    height: options.height,
    left: options.left,
    right: options.left + options.width,
    top: options.top,
    width: options.width,
    x: options.left,
    y: options.top,
    toJSON: () => ({}),
  } as DOMRect;
}
