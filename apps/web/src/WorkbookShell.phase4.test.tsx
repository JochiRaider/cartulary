import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import type { ComponentProps } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { TimelineWorkbook } from "./WorkbookShell";
import {
  buildAutoResolutionNotices,
  buildInspectorMentions,
  buildMentionPatchPayload,
  readCollectionItems,
} from "./workbookShellPhase4";

// Support-only mocked component coverage for Phase 4 workbook helpers.
// This file is not authoritative route or browser evidence.
const timelineViewSchemaId = "cartulary.view.timeline.v1";
type TimelineWorkbookProps = ComponentProps<typeof TimelineWorkbook>;
type EntityIndex = NonNullable<TimelineWorkbookProps["entityIndex"]>;
type EntityRowFixture = EntityIndex[string];

describe("Phase 4 workbook helpers", () => {
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
        resolutionMethod: "explicit_resolve_route",
        autoResolved: false,
        provenance: "manual",
        confidence: null,
        matchedAliasText: null,
      },
    ]);
  });

  it("builds resolve, dismiss, and revert mention payloads from the same internal contract", () => {
    expect(
      buildMentionPatchPayload(
        { rowVersion: 7 },
        { itemRef: "mention-host-auto", fieldKey: "timeline.host_refs" },
        "resolve_item",
        "timeline-client-7",
        "host-1",
      ),
    ).toEqual({
      view_schema_id: timelineViewSchemaId,
      base_row_version: 7,
      client_txn_id: "timeline-client-7",
      changes: [
        {
          field_key: "timeline.host_refs",
          action_payload: {
            kind: "collection_actions_v1",
            actions: [
              {
                op: "resolve_item",
                item_ref: "mention-host-auto",
                resolved_record_id: "host-1",
              },
            ],
          },
        },
      ],
    });
    expect(
      buildMentionPatchPayload(
        { rowVersion: 7 },
        { itemRef: "mention-host-auto", fieldKey: "timeline.host_refs" },
        "dismiss_item",
        "timeline-client-8",
      ),
    ).toEqual({
      view_schema_id: timelineViewSchemaId,
      base_row_version: 7,
      client_txn_id: "timeline-client-8",
      changes: [
        {
          field_key: "timeline.host_refs",
          action_payload: {
            kind: "collection_actions_v1",
            actions: [
              {
                op: "dismiss_item",
                item_ref: "mention-host-auto",
              },
            ],
          },
        },
      ],
    });
    expect(
      buildMentionPatchPayload(
        { rowVersion: 7 },
        { itemRef: "mention-host-auto", fieldKey: "timeline.host_refs" },
        "revert_to_unresolved",
        "timeline-client-9",
      ),
    ).toEqual({
      view_schema_id: timelineViewSchemaId,
      base_row_version: 7,
      client_txn_id: "timeline-client-9",
      changes: [
        {
          field_key: "timeline.host_refs",
          action_payload: {
            kind: "collection_actions_v1",
            actions: [
              {
                op: "revert_to_unresolved",
                item_ref: "mention-host-auto",
              },
            ],
          },
        },
      ],
    });
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
        resolutionMethod: "auto_match",
        autoResolved: true,
        status: "resolved",
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
        resolutionMethod: "explicit_resolve_route",
        autoResolved: false,
        status: "resolved",
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
        resolvedRecordId: "host-2",
        resolutionMethod: "explicit_resolve_route",
        autoResolved: false,
        status: "dismissed",
        displayText: "WS-023",
        provenance: null,
        confidence: null,
        matchedAliasText: null,
      },
    ]);
  });
});

describe("Phase 4 TimelineWorkbook", () => {
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
    vi.unstubAllGlobals();
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

    const autoChip = await screen.findByTestId("chip-mention-host-auto");
    const manualChip = screen.getByTestId("chip-mention-identity-manual");
    expect(autoChip.textContent).toContain("Auto");
    expect(manualChip.textContent).not.toContain("Auto");
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
                itemRef: "host-unresolved",
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
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-resolve",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "Alpha",
          captureState: "reviewed",
          hostRefs: [
            resolvedItem({
              itemRef: "host-unresolved",
              entityType: "host",
              rawText: "WS-023?",
              displayText: "WS-023",
              resolvedRecordId: "host-1",
              resolutionMethod: "explicit_resolve_route",
              autoResolved: false,
              provenance: "manual",
              confidence: null,
            }),
          ],
        }),
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

    fireEvent.click(await screen.findByTestId("row-record-1-inspect"));
    fireEvent.click(screen.getByTestId("mention-host-unresolved"));
    const preservedScroll = setTimelineGridScroll(240, 140);
    fireEvent.change(screen.getByTestId("inspector-resolve-target"), {
      target: { value: "host-1" },
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Resolve to existing" }),
    );

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    await expectTimelineContinuity("record-1", preservedScroll);
    expect(
      screen
        .getByTestId("row-record-1-hostRefs-items")
        .querySelector('[aria-label="Resolved WS-023"]'),
    ).toBeTruthy();
  });

  it("preserves continuity through create-from-mention entity refresh rerenders", async () => {
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
                itemRef: "identity-create",
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
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-create",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "Alpha",
          captureState: "reviewed",
          identityRefs: [
            resolvedItem({
              itemRef: "identity-create",
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
      }),
    );

    let rerenderTimelineWorkbook:
      | ReturnType<typeof render>["rerender"]
      | undefined;
    const onRefreshEntities = vi.fn(async () => {
      rerenderTimelineWorkbook?.(
        <TimelineWorkbook
          incidentId="incident-1"
          currentIncidentRole="admin"
          identityEntities={[createdIdentity]}
          entityIndex={buildEntityIndex(createdIdentity)}
          onRefreshEntities={onRefreshEntities}
        />,
      );
    });

    const renderResult = render(
      <TimelineWorkbook
        incidentId="incident-1"
        currentIncidentRole="admin"
        onRefreshEntities={onRefreshEntities}
      />,
    );
    rerenderTimelineWorkbook = renderResult.rerender;

    fireEvent.click(await screen.findByTestId("row-record-1-inspect"));
    fireEvent.click(screen.getByTestId("mention-identity-create"));
    const preservedScroll = setTimelineGridScroll(280, 175);
    fireEvent.click(screen.getByText("Create identity"));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    await waitFor(() => {
      expect(onRefreshEntities).toHaveBeenCalledTimes(1);
    });
    await waitFor(() => {
      expect(screen.getByTestId("timeline-inspector").textContent).toContain(
        "VPN User",
      );
    });
    await expectTimelineContinuity("record-1", preservedScroll);
  });

  it("renders a dismissed mention restore action after the dismiss flow completes", async () => {
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
              resolvedItem({
                itemRef: "mention-host-resolved",
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
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-dismiss",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "Alpha",
          captureState: "reviewed",
          hostRefs: [],
        }),
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-restore",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 3,
          summary: "Alpha",
          captureState: "reviewed",
          hostRefs: [
            unresolvedItem({
              itemRef: "mention-host-resolved",
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

    fireEvent.click(await screen.findByTestId("row-record-1-inspect"));
    await screen.findByText("Dismiss");
    const dismissScroll = setTimelineGridScroll(320, 180);
    fireEvent.click(screen.getByText("Dismiss"));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    await expectTimelineContinuity("record-1", dismissScroll);
    expect(screen.getByText("Restore to unresolved")).toBeTruthy();
    expect(screen.getByText("Dismissed")).toBeTruthy();

    const restoreScroll = setTimelineGridScroll(360, 90);
    fireEvent.click(screen.getByText("Restore to unresolved"));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
    await expectTimelineContinuity("record-1", restoreScroll);
    expect(
      screen
        .getByTestId("row-record-1-hostRefs-items")
        .querySelector('[aria-label="Unresolved WS-023"]'),
    ).toBeTruthy();
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

    const relationshipInput = (await screen.findByTestId(
      "row-record-1-hostRefs-input",
    )) as HTMLInputElement;
    fireEvent.change(relationshipInput, { target: { value: "WS-023" } });
    fireEvent.blur(relationshipInput);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });

    webSocketInstance?.onmessage?.(
      new MessageEvent("message", {
        data: JSON.stringify(
          recordChangedMessage({
            recordId: "record-1",
            rowVersion: 2,
            clientTxnId: "timeline-client-1",
          }),
        ),
      }),
    );

    await waitFor(
      () => {
        expect(fetchMock).toHaveBeenCalledTimes(2);
      },
      { timeout: 200 },
    );

    webSocketInstance?.onmessage?.(
      new MessageEvent("message", {
        data: JSON.stringify(
          recordChangedMessage({
            recordId: "record-1",
            rowVersion: 3,
            clientTxnId: "someone-else",
          }),
        ),
      }),
    );

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
  });
});

function timelineRow({
  recordId,
  rowVersion,
  occurredAt = "",
  summary = "",
  details = "",
  sourceText = "",
  captureState,
  hostRefs = [],
  identityRefs = [],
}: {
  recordId: string;
  rowVersion: number;
  occurredAt?: string;
  summary?: string;
  details?: string;
  sourceText?: string;
  captureState: string;
  hostRefs?: Array<Record<string, unknown>>;
  identityRefs?: Array<Record<string, unknown>>;
}) {
  return {
    record_id: recordId,
    row_version: rowVersion,
    cells: {
      "timeline.occurred_at": { value: occurredAt },
      "timeline.summary": { value: summary },
      "timeline.details": { value: details },
      "timeline.source_text": { value: sourceText },
      "timeline.capture_state": { value: captureState },
      "timeline.host_refs": { value: { items: hostRefs } },
      "timeline.identity_refs": { value: { items: identityRefs } },
    },
  };
}

function resolvedItem({
  itemRef,
  entityType,
  rawText,
  displayText,
  resolvedRecordId,
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
    resolution_method: resolutionMethod,
    auto_resolved: autoResolved,
    provenance,
    confidence,
    matched_alias_text: matchedAliasText,
  };
}

function unresolvedItem({
  itemRef,
  entityType,
  rawText,
}: {
  itemRef: string;
  entityType: "host" | "identity";
  rawText: string;
}) {
  return {
    item_ref: itemRef,
    entity_type: entityType,
    item_kind: "unresolved_mention",
    display_text: rawText,
    raw_text: rawText,
  };
}

function recordChangedMessage({
  recordId,
  rowVersion,
  clientTxnId,
}: {
  recordId: string;
  rowVersion: number;
  clientTxnId: string;
}) {
  return {
    type: "record_changed",
    payload: {
      record_id: recordId,
      row_version: rowVersion,
      change_set_id: `change-set-${rowVersion}`,
      client_txn_id: clientTxnId,
      actor_user_id: "user-1",
      changed_field_keys: ["timeline.host_refs"],
      affected_views: [
        {
          view_schema_id: timelineViewSchemaId,
          change_kind: "invalidate",
        },
      ],
    },
  };
}

function successEnvelope(payload: unknown) {
  return Promise.resolve({
    ok: true,
    status: 200,
    json: async () => ({ data: payload }),
  } as Response);
}

function setTimelineGridScroll(top: number, left: number) {
  const grid = screen.getByTestId("timeline-grid-shell") as HTMLDivElement;
  grid.scrollTop = top;
  grid.scrollLeft = left;
  return { top, left };
}

async function expectTimelineContinuity(
  recordId: string,
  preservedScroll: { top: number; left: number },
) {
  await waitFor(() => {
    expect(document.activeElement).toBe(
      screen.getByTestId(`row-${recordId}-inspect`),
    );
    const grid = screen.getByTestId("timeline-grid-shell") as HTMLDivElement;
    expect(grid.scrollTop).toBe(preservedScroll.top);
    expect(grid.scrollLeft).toBe(preservedScroll.left);
  });
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
