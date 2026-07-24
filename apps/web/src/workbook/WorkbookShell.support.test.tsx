import {
  autoResolutionNoticeFamilySelector,
  autoResolutionNoticeTestId,
  autoResolutionUndoButtonTestId,
  gridActionsHeaderTestId,
  gridScrollportSelector,
  gridShellTestId,
  mentionCreateEntityButtonTestId,
  mentionDismissButtonTestId,
  mentionItemTestId,
  mentionResolveExistingButtonTestId,
  mentionResolveTargetSelectTestId,
  mentionRestoreUnresolvedButtonTestId,
  relationshipChipTestId,
  relationshipItemsTestId,
  relationshipOverflowButtonTestId,
  rowCellTestId,
  rowInspectButtonTestId,
  saveStateTestId,
  timelineCollectionInputTestId,
  timelineInspectorSectionTestId,
  timelineInspectorTestId,
  timelineRowReplacementInputTestId,
  timelineRowVersionTestId,
  workbookInlineDraftRowTestId,
  workbookRowContextMenuTestId,
} from "@cartulary/ui-contracts";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import type { ComponentProps } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../testing/fetchMockTestSupport";
import {
  buildRecordChangedPayload,
  emitRecordChanged,
  extractTimelineJSONBody,
  successEnvelope,
  timelineRow,
  timelineRowsEnvelope,
} from "../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "./models/workbookSurfaceRegistry";
import { TimelineWorkbook } from "./timeline/components/TimelineWorkbook";
import {
  buildAutoResolutionNotices,
  buildInspectorMentions,
  readCollectionItems,
} from "./timeline/models/workbookMentionChips";
import { buildMentionActionPayload } from "./timeline/services/workbookCollaborationMessages";

vi.mock(
  "@cartulary/grid-adapter",
  async () => import("@cartulary/grid-adapter/test-support"),
);

// Support-only mocked component coverage for Record relationships workbook helpers.
// This file is not authoritative Record relationships evidence.
type TimelineWorkbookProps = ComponentProps<typeof TimelineWorkbook>;
type EntityIndex = NonNullable<TimelineWorkbookProps["entityIndex"]>;
type EntityRowFixture = EntityIndex[string];

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
                  resolved_record_id: "host-1",
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
                  resolved_record_id: "host-2",
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
        itemKind: "resolved_ref",
        displayText: "VPN Gateway",
        rawText: " vpn   gateway ",
        resolvedRecordId: "host-1",
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
        itemKind: "resolved_ref",
        displayText: "WS-023",
        rawText: "WS-023",
        resolvedRecordId: "host-2",
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
        "host-1",
      ),
    ).toEqual({
      base_mention_row_version: 13,
      client_txn_id: "timeline-client-8",
      action: "resolve_item",
      resolved_record_id: "host-1",
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
      recordId: "record-1",
      collectionValues: {
        hostRefs: [],
        identityRefs: [
          {
            itemRef: "mention-identity-manual",
            entityType: "identity" as const,
            itemKind: "resolved_ref",
            displayText: "Alex Analyst",
            rawText: "alex.analyst@example.test",
            resolvedRecordId: "identity-1",
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
      recordId: "record-1",
      collectionValues: {
        hostRefs: [
          {
            itemRef: "mention-host-auto",
            entityType: "host" as const,
            itemKind: "resolved_ref",
            displayText: "VPN Gateway",
            rawText: " vpn   gateway ",
            resolvedRecordId: "host-1",
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
        rowRecordId: "record-1",
        fieldKey: "timeline.host_refs",
        entityType: "host",
        rawText: " vpn   gateway ",
        resolvedRecordId: "host-1",
        matchedAliasText: "VPN Gateway",
      },
    ]);

    expect(
      buildInspectorMentions(afterRow, [
        {
          rowRecordId: "record-1",
          fieldKey: "timeline.host_refs",
          entityType: "host",
          itemRef: "mention-host-dismissed",
          rawText: "WS-023",
          resolvedRecordId: "host-2",
          mentionRowVersion: 23,
          resolutionMethod: "explicit_resolve_route",
          autoResolved: false,
        },
      ]),
    ).toEqual([
      {
        rowRecordId: "record-1",
        fieldKey: "timeline.host_refs",
        entityType: "host",
        itemRef: "mention-host-auto",
        rawText: " vpn   gateway ",
        resolvedRecordId: "host-1",
        mentionRowVersion: 22,
        resolutionMethod: "auto_match",
        autoResolved: true,
        status: "resolved",
        chipState: "auto_resolved",
        anchor: {
          recordId: "record-1",
          fieldKey: "timeline.host_refs",
          itemRef: "mention-host-auto",
          entityMentionId: null,
          targetEntityRecordId: "host-1",
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
        rowRecordId: "record-1",
        fieldKey: "timeline.identity_refs",
        entityType: "identity",
        itemRef: "mention-identity-manual",
        rawText: "alex.analyst@example.test",
        resolvedRecordId: "identity-1",
        mentionRowVersion: 21,
        resolutionMethod: "explicit_resolve_route",
        autoResolved: false,
        status: "resolved",
        chipState: "resolved",
        anchor: {
          recordId: "record-1",
          fieldKey: "timeline.identity_refs",
          itemRef: "mention-identity-manual",
          entityMentionId: null,
          targetEntityRecordId: "identity-1",
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
        rowRecordId: "record-1",
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
          recordId: "record-1",
          fieldKey: "timeline.host_refs",
          itemRef: "mention-host-dismissed",
          entityMentionId: null,
          targetEntityRecordId: null,
        },
        sourceKind: "entity_mention",
        isActiveRelationshipValue: false,
        priorTargetEntityRecordId: "host-2",
        displayText: "WS-023",
        provenance: null,
        confidence: null,
        matchedAliasText: null,
      },
    ]);
  });
});

describe("support TimelineWorkbook", () => {
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
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 2,
            summary: "Alpha",
            captureState: "rough",
          }),
        ],
      }),
    );

    render(
      <TimelineWorkbook incidentId="incident-1" currentIncidentRole="admin" />,
    );

    const summaryCell = await screen.findByTestId(
      rowCellTestId("record-1", "timeline.activity_synopsis_text"),
    );
    expect(
      screen.queryByTestId(gridActionsHeaderTestId(timelineViewSchemaId)),
    ).toBeNull();

    fireEvent.contextMenu(summaryCell, { clientX: 32, clientY: 48 });
    const contextMenu = await screen.findByTestId(
      workbookRowContextMenuTestId(timelineViewSchemaId, "record-1"),
    );
    expect(contextMenu.getAttribute("role")).toBe("dialog");
    expect(screen.getByTestId(rowInspectButtonTestId("record-1"))).toBeTruthy();
    const replacementInput = screen.getByTestId(
      timelineRowReplacementInputTestId("record-1"),
    );
    replacementInput.focus();
    fireEvent.scroll(window);
    expect(
      screen.getByTestId(
        workbookRowContextMenuTestId(timelineViewSchemaId, "record-1"),
      ),
    ).toBeTruthy();

    replacementInput.blur();
    fireEvent.scroll(window);
    await waitFor(() => {
      expect(
        screen.queryByTestId(
          workbookRowContextMenuTestId(timelineViewSchemaId, "record-1"),
        ),
      ).toBeNull();
    });

    fireEvent.contextMenu(summaryCell, { clientX: 32, clientY: 48 });
    const reopenedContextMenu = await screen.findByTestId(
      workbookRowContextMenuTestId(timelineViewSchemaId, "record-1"),
    );

    fireEvent.keyDown(reopenedContextMenu, { key: "Escape" });
    await waitFor(() => {
      expect(
        screen.queryByTestId(
          workbookRowContextMenuTestId(timelineViewSchemaId, "record-1"),
        ),
      ).toBeNull();
    });
  });

  it("opens Timeline row actions from the keyboard and ignores draft rows", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 2,
            summary: "Alpha",
            captureState: "rough",
          }),
        ],
      }),
    );

    render(
      <TimelineWorkbook incidentId="incident-1" currentIncidentRole="admin" />,
    );

    const draftRow = await screen.findByTestId(
      workbookInlineDraftRowTestId(timelineViewSchemaId),
    );
    fireEvent.contextMenu(draftRow, { clientX: 12, clientY: 24 });
    expect(
      screen.queryByTestId(
        workbookRowContextMenuTestId(timelineViewSchemaId, "record-1"),
      ),
    ).toBeNull();

    const summaryCell = await screen.findByTestId(
      rowCellTestId("record-1", "timeline.activity_synopsis_text"),
    );
    summaryCell.focus();
    fireEvent.keyDown(summaryCell, { key: "F10", shiftKey: true });
    expect(
      await screen.findByTestId(
        workbookRowContextMenuTestId(timelineViewSchemaId, "record-1"),
      ),
    ).toBeTruthy();
  });

  it("renders auto-resolved chips distinctly from manual resolved chips", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 2,
            summary: "Alpha",
            captureState: "reviewed",
            hostRefs: [
              resolvedItem({
                itemRef: "mention-host-auto",
                entityType: "host",
                rawText: " vpn   gateway ",
                displayText: "VPN Gateway",
                resolvedRecordId: "host-1",
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
                entityType: "identity",
                rawText: "alex.analyst@example.test",
                displayText: "Alex Analyst",
                resolvedRecordId: "identity-1",
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
      <TimelineWorkbook incidentId="incident-1" currentIncidentRole="admin" />,
    );

    await openTimelineInspectorFromContext("record-1");
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
    expect(autoChip.textContent).toContain("Auto");
    expect(manualChip.textContent).not.toContain("Auto");
    expect(manualChip.textContent).toContain("Manual");
  });

  it("preserves continuity when resolving a mention to an existing entity", async () => {
    const existingHost = buildEntityRow({
      entityType: "host",
      recordId: "host-1",
      rowVersion: 1,
      label: "WS-023",
      secondaryText: "ws-023.corp.example.test",
      state: "stub",
      identifiers: [
        {
          key: "host.hostname",
          label: "Hostname",
          value: "ws-023.corp.example.test",
        },
      ],
    });

    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "reviewed",
            hostRefs: [
              unresolvedItem({
                itemRef: "entity_mention:11111111-1111-4111-8111-000000000401",
                entityType: "host",
                rawText: "WS-023?",
              }),
            ],
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        entity_mention: {
          entity_mention_id: "11111111-1111-4111-8111-000000000401",
          source_record_id: "record-1",
          source_field_key: "timeline.host_refs",
          entity_type: "host",
          raw_text: "WS-023?",
          resolution_status: "resolved",
          resolved_record_id: "host-1",
          row_version: 2,
          resolution_method: "explicit_resolve_route",
        },
        source_record: {
          record_id: "record-1",
          row_version: 2,
        },
        change_set_id: "change-set-resolve",
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 2,
            summary: "Alpha",
            captureState: "reviewed",
            hostRefs: [
              resolvedItem({
                itemRef: "entity_mention:11111111-1111-4111-8111-000000000401",
                entityType: "host",
                rawText: "WS-023?",
                displayText: "WS-023",
                resolvedRecordId: "host-1",
                resolutionMethod: "explicit_resolve_route",
                autoResolved: false,
                provenance: "manual",
                confidence: null,
                mentionRowVersion: 2,
              }),
            ],
          }),
        ],
      }),
    );

    render(
      <TimelineWorkbook
        incidentId="incident-1"
        currentIncidentRole="admin"
        hostEntities={[existingHost]}
        entityIndex={buildEntityIndex(existingHost)}
      />,
    );

    await openTimelineInspectorFromContext("record-1");
    fireEvent.click(
      screen.getByTestId(
        mentionItemTestId(
          "entity_mention:11111111-1111-4111-8111-000000000401",
        ),
      ),
    );
    const preservedScroll = setTimelineGridScroll(240, 140);
    fireEvent.change(screen.getByTestId(mentionResolveTargetSelectTestId()), {
      target: { value: "host-1" },
    });
    fireEvent.click(screen.getByTestId(mentionResolveExistingButtonTestId()));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
    await expectTimelineFocusAndScroll("record-1", preservedScroll);
    expect(String(fetchMock.mock.calls[1]?.[0])).toContain(
      "/api/v1/entity-mentions/11111111-1111-4111-8111-000000000401/resolve",
    );
    expect(extractTimelineJSONBody(fetchMock, 1)).toMatchObject({
      base_mention_row_version: 1,
      action: "resolve_item",
      resolved_record_id: "host-1",
    });
    await waitFor(() => {
      expect(
        screen
          .getByTestId(
            relationshipItemsTestId("record-1", "timeline.host_refs"),
          )
          .querySelector('[aria-label^="Resolved WS-023"]'),
      ).toBeTruthy();
    });
  });

  it("preserves continuity through create-from-mention entity refresh rerenders", async () => {
    let inspectContentTop = 560;
    let maxScrollTop = 400;
    const createdIdentity = buildEntityRow({
      entityType: "identity",
      recordId: "identity-1",
      rowVersion: 1,
      label: "VPN User",
      secondaryText: "vpn.user@example.test",
      state: "stub",
      identifiers: [
        {
          key: "identity.upn",
          label: "UPN",
          value: "vpn.user@example.test",
        },
      ],
    });

    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "reviewed",
            identityRefs: [
              unresolvedItem({
                itemRef: "entity_mention:identity-create",
                entityType: "identity",
                rawText: "vpn.user@example.test",
              }),
            ],
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: "cartulary.view.identities.v1",
        change_set_id: "change-set-identity-create",
        row: {
          record_id: "identity-1",
          row_version: 1,
          cells: {},
        },
      }),
    );
    fetchMock.mockResolvedValueOnce(
      mentionActionEnvelope({
        actionStatus: "resolved",
        entityType: "identity",
        mentionId: "identity-create",
        rawText: "vpn.user@example.test",
        resolvedRecordId: "identity-1",
        sourceFieldKey: "timeline.identity_refs",
        sourceRowVersion: 2,
        mentionRowVersion: 2,
        resolutionMethod: "explicit_resolve_route",
      }),
    );
    fetchMock.mockResolvedValueOnce(
      timelineRowsEnvelope([
        timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "Alpha",
          captureState: "reviewed",
          identityRefs: [
            resolvedItem({
              itemRef: "entity_mention:identity-create",
              entityType: "identity",
              rawText: "vpn.user@example.test",
              displayText: "vpn.user@example.test",
              resolvedRecordId: "identity-1",
              resolutionMethod: "explicit_resolve_route",
              autoResolved: false,
              provenance: "manual",
              confidence: null,
            }),
          ],
        }),
      ]),
    );

    let rerenderTimelineWorkbook:
      | ReturnType<typeof render>["rerender"]
      | undefined;
    const refreshGate = deferred<void>();
    const onRefreshEntities = vi.fn(async () => {
      const renderRefreshedWorkbook = (identity: EntityRowFixture) => {
        rerenderTimelineWorkbook?.(
          <TimelineWorkbook
            incidentId="incident-1"
            currentIncidentRole="admin"
            identityEntities={[identity]}
            entityIndex={buildEntityIndex(identity)}
            onRefreshEntities={onRefreshEntities}
          />,
        );
      };
      await refreshGate.promise;
      inspectContentTop = 80;
      maxScrollTop = 120;
      renderRefreshedWorkbook(createdIdentity);
      await waitForPostRenderFrame();
      renderRefreshedWorkbook({
        ...createdIdentity,
        aliasTexts: [...createdIdentity.aliasTexts],
        identifiers: [...createdIdentity.identifiers],
      });
    });

    const renderResult = render(
      <TimelineWorkbook
        incidentId="incident-1"
        currentIncidentRole="admin"
        onRefreshEntities={onRefreshEntities}
      />,
    );
    rerenderTimelineWorkbook = renderResult.rerender;

    await screen.findByTestId(
      rowCellTestId("record-1", "timeline.activity_synopsis_text"),
    );
    installTimelineGridScrollClamp(() => maxScrollTop);
    installTimelineInspectGeometry("record-1", {
      containerHeight: 300,
      containerLeft: 40,
      containerTop: 100,
      containerWidth: 400,
      contentLeft: 220,
      contentTop: () => inspectContentTop,
      targetHeight: 40,
      targetWidth: 80,
    });

    await openTimelineInspectorFromContext("record-1");
    fireEvent.click(
      screen.getByTestId(mentionItemTestId("entity_mention:identity-create")),
    );
    const preservedScroll = setTimelineGridScroll(400, 175);
    expect(isTimelineFocusTargetFullyVisibleWithinGrid("record-1")).toBe(true);
    fireEvent.click(
      screen.getByTestId(mentionCreateEntityButtonTestId("identity")),
    );

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(4);
    });
    await waitFor(() => {
      expect(onRefreshEntities).toHaveBeenCalledTimes(1);
    });
    refreshGate.resolve();
    await waitFor(() => {
      expect(
        screen.getByTestId(timelineInspectorTestId()).textContent,
      ).toContain("VPN User");
    });
    await waitForPostRenderFrame();
    await expectTimelineFocusAndScroll("record-1", preservedScroll, {
      expectedTop: null,
      requireVisibleWithinGrid: true,
    });
  });

  it("reveals a clipped inspect action after an auto-resolution collection patch", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
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
        change_set_id: "change-set-auto-resolve",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "Alpha",
          captureState: "reviewed",
          hostRefs: [
            resolvedItem({
              itemRef: "mention-host-auto",
              entityType: "host",
              rawText: " vpn   gateway ",
              displayText: "Gateway node",
              resolvedRecordId: "host-1",
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
      <TimelineWorkbook incidentId="incident-1" currentIncidentRole="admin" />,
    );

    await openTimelineInspectorFromContext("record-1");
    const relationshipInput = (await screen.findByTestId(
      timelineCollectionInputTestId("record-1", "timeline.host_refs"),
    )) as HTMLInputElement;
    installTimelineInspectGeometry("record-1", {
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

    expect(isTimelineFocusTargetFullyVisibleWithinGrid("record-1")).toBe(false);

    fireEvent.change(relationshipInput, {
      target: { value: " vpn   gateway " },
    });
    fireEvent.keyDown(relationshipInput, { key: "Enter" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    await expectTimelineFocusAndScroll("record-1", preservedScroll, {
      expectedLeft: 48,
      requireVisibleWithinGrid: true,
    });
    expect(
      await screen.findByTestId(
        autoResolutionNoticeTestId("mention-host-auto"),
      ),
    ).toBeTruthy();
  });

  it("sends auto-resolution Undo with the current post-resolution row version", async () => {
    const mentionItemRef =
      "entity_mention:11111111-1111-4111-8111-000000000404";
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
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
        change_set_id: "change-set-auto-resolve",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "Alpha",
          captureState: "reviewed",
          hostRefs: [
            resolvedItem({
              itemRef: mentionItemRef,
              entityType: "host",
              rawText: " vpn   gateway ",
              displayText: "Gateway node",
              resolvedRecordId: "host-1",
              resolutionMethod: "auto_match",
              autoResolved: true,
              provenance: "auto_match",
              confidence: 100,
            }),
          ],
        }),
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        entity_mention: {
          entity_mention_id: "11111111-1111-4111-8111-000000000404",
          source_record_id: "record-1",
          source_field_key: "timeline.host_refs",
          entity_type: "host",
          raw_text: " vpn   gateway ",
          resolution_status: "unresolved",
          resolved_record_id: null,
          row_version: 2,
          resolution_method: null,
        },
        source_record: {
          record_id: "record-1",
          row_version: 3,
        },
        change_set_id: "change-set-auto-undo",
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 3,
            summary: "Alpha",
            captureState: "reviewed",
            hostRefs: [
              unresolvedItem({
                itemRef: mentionItemRef,
                entityType: "host",
                rawText: " vpn   gateway ",
                mentionRowVersion: 2,
              }),
            ],
          }),
        ],
      }),
    );

    render(
      <TimelineWorkbook incidentId="incident-1" currentIncidentRole="admin" />,
    );

    await openTimelineInspectorFromContext("record-1");
    const relationshipInput = (await screen.findByTestId(
      timelineCollectionInputTestId("record-1", "timeline.host_refs"),
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

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(4);
    });
    await expectTimelineFocusAndScroll("record-1", preservedScroll);
    expect(
      screen.queryByTestId(autoResolutionNoticeTestId(mentionItemRef)),
    ).toBeNull();
    expect(String(fetchMock.mock.calls[2]?.[0])).toContain(
      "/api/v1/entity-mentions/11111111-1111-4111-8111-000000000404/resolve",
    );
    expect(extractTimelineJSONBody(fetchMock, 2)).toMatchObject({
      base_mention_row_version: 1,
      action: "revert_to_unresolved",
    });
  });

  it("keeps the workbook mounted after committing a relationship-cell edit", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
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
        change_set_id: "change-set-relationship-commit",
        row: timelineRow({
          recordId: "record-1",
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
      <TimelineWorkbook incidentId="incident-1" currentIncidentRole="admin" />,
    );

    await openTimelineInspectorFromContext("record-1");
    const relationshipInput = (await screen.findByTestId(
      timelineCollectionInputTestId("record-1", "timeline.host_refs"),
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
        timelineCollectionInputTestId("record-1", "timeline.host_refs"),
      ),
    ).toBeTruthy();
    expect(
      screen.getByTestId(
        relationshipItemsTestId("record-1", "timeline.host_refs"),
      ).textContent,
    ).toContain("WS-023");
    fireEvent.contextMenu(
      screen.getByTestId(
        rowCellTestId("record-1", "timeline.activity_synopsis_text"),
      ),
      { clientX: 32, clientY: 48 },
    );
    expect(screen.getByTestId(rowInspectButtonTestId("record-1"))).toBeTruthy();
  });

  it("commits sequential hostRefs edits with the latest returned row versions", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
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
        change_set_id: "change-set-host-1",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "Alpha",
          captureState: "reviewed",
          hostRefs: [
            unresolvedItem({
              itemRef: "mention-host-1",
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
        change_set_id: "change-set-host-2",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 3,
          summary: "Alpha",
          captureState: "reviewed",
          hostRefs: [
            unresolvedItem({
              itemRef: "mention-host-1",
              entityType: "host",
              rawText: "WS-023",
            }),
            unresolvedItem({
              itemRef: "mention-host-2",
              entityType: "host",
              rawText: "WS-024",
            }),
          ],
        }),
      }),
    );

    render(
      <TimelineWorkbook incidentId="incident-1" currentIncidentRole="admin" />,
    );

    await openTimelineInspectorFromContext("record-1");
    const firstInput = (await screen.findByTestId(
      timelineCollectionInputTestId("record-1", "timeline.host_refs"),
    )) as HTMLInputElement;
    fireEvent.change(firstInput, { target: { value: "WS-023" } });
    fireEvent.keyDown(firstInput, { key: "Enter" });
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
      expect(
        screen.getByTestId(timelineRowVersionTestId("record-1")).textContent,
      ).toBe("2");
    });

    const secondInput = screen.getByTestId(
      timelineCollectionInputTestId("record-1", "timeline.host_refs"),
    ) as HTMLInputElement;
    fireEvent.change(secondInput, { target: { value: "WS-024" } });
    fireEvent.keyDown(secondInput, { key: "Enter" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
      expect(
        screen.getByTestId(timelineRowVersionTestId("record-1")).textContent,
      ).toBe("3");
    });

    expectHostRefAddRequest(fetchMock, 1, 1, "WS-023");
    expectHostRefAddRequest(fetchMock, 2, 2, "WS-024");
    expect(
      screen.getByTestId(
        relationshipItemsTestId("record-1", "timeline.host_refs"),
      ).textContent,
    ).toContain("WS-023");
    expect(
      screen.getByTestId(
        relationshipItemsTestId("record-1", "timeline.host_refs"),
      ).textContent,
    ).toContain("WS-024");
    expect(
      document.querySelector(autoResolutionNoticeFamilySelector()),
    ).toBeNull();
  });

  it("treats Enter followed by blur as one hostRefs collection commit", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
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
        change_set_id: "change-set-host-1",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "Alpha",
          captureState: "reviewed",
          hostRefs: [
            unresolvedItem({
              itemRef: "mention-host-1",
              entityType: "host",
              rawText: "WS-023",
            }),
          ],
        }),
      }),
    );

    render(
      <TimelineWorkbook incidentId="incident-1" currentIncidentRole="admin" />,
    );

    await openTimelineInspectorFromContext("record-1");
    const relationshipInput = (await screen.findByTestId(
      timelineCollectionInputTestId("record-1", "timeline.host_refs"),
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
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
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
      <TimelineWorkbook incidentId="incident-1" currentIncidentRole="admin" />,
    );

    await openTimelineInspectorFromContext("record-1");
    const relationshipInput = (await screen.findByTestId(
      timelineCollectionInputTestId("record-1", "timeline.host_refs"),
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
        change_set_id: "change-set-host-1",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "Alpha",
          captureState: "reviewed",
          hostRefs: [
            unresolvedItem({
              itemRef: "mention-host-1",
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
        change_set_id: "change-set-host-2",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 3,
          summary: "Alpha",
          captureState: "reviewed",
          hostRefs: [
            unresolvedItem({
              itemRef: "mention-host-1",
              entityType: "host",
              rawText: "WS-023",
            }),
            unresolvedItem({
              itemRef: "mention-host-2",
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
          relationshipItemsTestId("record-1", "timeline.host_refs"),
        ).textContent,
      ).toContain("WS-024");
    });
  });

  it("opens the Relationships inspector from compact relationship overflow", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "reviewed",
            hostRefs: [
              unresolvedItem({
                itemRef: "mention-host-1",
                entityType: "host",
                rawText: "WS-023",
              }),
              unresolvedItem({
                itemRef: "mention-host-2",
                entityType: "host",
                rawText: "WS-024",
              }),
            ],
          }),
        ],
      }),
    );

    render(
      <TimelineWorkbook incidentId="incident-1" currentIncidentRole="admin" />,
    );

    await openTimelineInspectorFromContext("record-1");
    const overflowButton = await screen.findByTestId(
      relationshipOverflowButtonTestId("record-1", "timeline.host_refs"),
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
        screen.getByTestId(mentionItemTestId("mention-host-2")),
      );
    });
  });

  it("renders a dismissed mention restore action after the dismiss flow completes", async () => {
    const mentionId = "11111111-1111-4111-8111-000000000402";
    const mentionItemRef = `entity_mention:${mentionId}`;
    mockDismissRestoreMentionResponses(fetchMock, {
      mentionId,
      mentionItemRef,
    });

    render(
      <TimelineWorkbook incidentId="incident-1" currentIncidentRole="admin" />,
    );

    await openTimelineInspectorFromContext("record-1");
    await screen.findByTestId(mentionDismissButtonTestId());
    const dismissScroll = setTimelineGridScroll(320, 180);
    fireEvent.click(screen.getByTestId(mentionDismissButtonTestId()));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
    await expectTimelineFocusAndScroll("record-1", dismissScroll);
    expect(
      screen.getByTestId(mentionRestoreUnresolvedButtonTestId()),
    ).toBeTruthy();
    expect(screen.getAllByText("Dismissed").length).toBeGreaterThanOrEqual(2);

    const restoreScroll = setTimelineGridScroll(360, 90);
    fireEvent.click(screen.getByTestId(mentionRestoreUnresolvedButtonTestId()));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(5);
    });
    await expectTimelineFocusAndScroll("record-1", restoreScroll);
    await waitFor(() => {
      expect(
        screen
          .getByTestId(
            relationshipItemsTestId("record-1", "timeline.host_refs"),
          )
          .querySelector('[aria-label="Unresolved WS-023"]'),
      ).toBeTruthy();
    });
  });

  it("reveals a vertically clipped inspect action through dismiss and restore continuity", async () => {
    const mentionId = "11111111-1111-4111-8111-000000000403";
    const mentionItemRef = `entity_mention:${mentionId}`;
    mockDismissRestoreMentionResponses(fetchMock, {
      mentionId,
      mentionItemRef,
    });

    render(
      <TimelineWorkbook incidentId="incident-1" currentIncidentRole="admin" />,
    );

    await screen.findByTestId(
      rowCellTestId("record-1", "timeline.activity_synopsis_text"),
    );
    installTimelineInspectGeometry("record-1", {
      containerHeight: 300,
      containerLeft: 40,
      containerTop: 100,
      containerWidth: 400,
      contentLeft: 85,
      contentTop: 610,
      targetHeight: 40,
      targetWidth: 80,
    });

    await openTimelineInspectorFromContext("record-1");
    await screen.findByTestId(mentionDismissButtonTestId());

    const dismissScroll = setTimelineGridScroll(320, 18);
    expect(isTimelineFocusTargetFullyVisibleWithinGrid("record-1")).toBe(false);
    fireEvent.click(screen.getByTestId(mentionDismissButtonTestId()));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
    await expectTimelineFocusAndScroll("record-1", dismissScroll, {
      expectedTop: 350,
      requireVisibleWithinGrid: true,
    });
    expect(
      screen.getByTestId(mentionRestoreUnresolvedButtonTestId()),
    ).toBeTruthy();

    const restoreScroll = setTimelineGridScroll(340, 18);
    expect(isTimelineFocusTargetFullyVisibleWithinGrid("record-1")).toBe(false);
    fireEvent.click(screen.getByTestId(mentionRestoreUnresolvedButtonTestId()));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(5);
    });
    await expectTimelineFocusAndScroll("record-1", restoreScroll, {
      expectedTop: 350,
      requireVisibleWithinGrid: true,
    });
    await waitFor(() => {
      expect(
        screen
          .getByTestId(
            relationshipItemsTestId("record-1", "timeline.host_refs"),
          )
          .querySelector('[aria-label="Unresolved WS-023"]'),
      ).toBeTruthy();
    });
  });

  it("suppresses self-originated websocket invalidations and reloads for external ones", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
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
        change_set_id: "change-set-patch",
        row: timelineRow({
          recordId: "record-1",
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
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
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
      <TimelineWorkbook incidentId="incident-1" currentIncidentRole="admin" />,
    );

    await openTimelineInspectorFromContext("record-1");
    const relationshipInput = (await screen.findByTestId(
      timelineCollectionInputTestId("record-1", "timeline.host_refs"),
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
        recordId: "record-1",
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
        recordId: "record-1",
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
    incident_id: "incident-1",
    entity_mention: {
      entity_mention_id: mentionId,
      source_record_id: "record-1",
      source_field_key: sourceFieldKey,
      entity_type: entityType,
      raw_text: rawText,
      resolution_status: actionStatus,
      resolved_record_id: resolvedRecordId,
      row_version: mentionRowVersion,
      resolution_method: resolutionMethod,
    },
    source_record: {
      record_id: "record-1",
      row_version: sourceRowVersion,
    },
    change_set_id: `change-set-${actionStatus}`,
  });
}

function mockDismissRestoreMentionResponses(
  fetchMock: ReturnType<typeof vi.fn>,
  options: {
    mentionId: string;
    mentionItemRef: string;
  },
) {
  fetchMock.mockResolvedValueOnce(
    timelineRowsEnvelope([
      timelineRow({
        recordId: "record-1",
        rowVersion: 1,
        summary: "Alpha",
        captureState: "reviewed",
        hostRefs: [
          resolvedItem({
            itemRef: options.mentionItemRef,
            entityType: "host",
            rawText: "WS-023",
            displayText: "WS-023",
            resolvedRecordId: "host-1",
            resolutionMethod: "explicit_resolve_route",
            autoResolved: false,
            provenance: "manual",
            confidence: null,
          }),
        ],
      }),
    ]),
  );
  fetchMock.mockResolvedValueOnce(
    mentionActionEnvelope({
      actionStatus: "dismissed",
      mentionId: options.mentionId,
      rawText: "WS-023",
      resolvedRecordId: null,
      sourceRowVersion: 2,
      mentionRowVersion: 2,
      resolutionMethod: "explicit_resolve_route",
    }),
  );
  fetchMock.mockResolvedValueOnce(
    timelineRowsEnvelope([
      timelineRow({
        recordId: "record-1",
        rowVersion: 2,
        summary: "Alpha",
        captureState: "reviewed",
        hostRefs: [],
      }),
    ]),
  );
  fetchMock.mockResolvedValueOnce(
    mentionActionEnvelope({
      actionStatus: "unresolved",
      mentionId: options.mentionId,
      rawText: "WS-023",
      resolvedRecordId: null,
      sourceRowVersion: 3,
      mentionRowVersion: 3,
      resolutionMethod: null,
    }),
  );
  fetchMock.mockResolvedValueOnce(
    timelineRowsEnvelope([
      timelineRow({
        recordId: "record-1",
        rowVersion: 3,
        summary: "Alpha",
        captureState: "reviewed",
        hostRefs: [
          unresolvedItem({
            itemRef: options.mentionItemRef,
            entityType: "host",
            rawText: "WS-023",
            mentionRowVersion: 3,
          }),
        ],
      }),
    ]),
  );
}

function unresolvedItem({
  itemRef,
  entityType,
  rawText,
  mentionRowVersion = 1,
}: {
  itemRef: string;
  entityType: "host" | "identity";
  rawText: string;
  mentionRowVersion?: number;
}) {
  return {
    item_ref: itemRef,
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

async function waitForPostRenderFrame() {
  await new Promise<void>((resolve) => {
    window.requestAnimationFrame(() => resolve());
  });
  await new Promise<void>((resolve) => {
    window.requestAnimationFrame(() => resolve());
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

function installTimelineGridScrollClamp(maxTop: () => number) {
  const grid = timelineGridScrollport();
  let scrollTop = grid.scrollTop;
  Object.defineProperty(grid, "scrollTop", {
    configurable: true,
    get: () => scrollTop,
    set: (value: number) => {
      const numericValue =
        typeof value === "number" && Number.isFinite(value) ? value : 0;
      scrollTop = Math.max(0, Math.min(numericValue, maxTop()));
    },
  });
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

function buildEntityIndex(...rows: EntityRowFixture[]): EntityIndex {
  const index: EntityIndex = {};
  for (const row of rows) {
    index[row.recordId] = row;
  }
  return index;
}

function buildEntityRow({
  entityType,
  recordId,
  rowVersion,
  label,
  secondaryText,
  state,
  identifiers,
}: {
  entityType: "host" | "identity";
  recordId: string;
  rowVersion: number;
  label: string;
  secondaryText: string;
  state: string;
  identifiers: Array<{
    key: string;
    label: string;
    value: string;
  }>;
}): EntityRowFixture {
  return {
    entityType,
    recordId,
    rowVersion,
    label,
    secondaryText,
    state,
    aliasTexts: [],
    linkedEventCount: 0,
    rawRow: timelineRow({
      recordId,
      rowVersion,
      summary: label,
      captureState: state,
    }),
    identifiers,
  };
}
