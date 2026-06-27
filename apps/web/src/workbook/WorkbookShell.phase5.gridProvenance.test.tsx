import {
  entityInspectorTestId,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  genericEditFieldSelectTestId,
  genericEditRecordSelectTestId,
  genericEditSubmitTestId,
  genericEditValueTestId,
  gridShellTestId,
  rowCellTestId,
  surfaceTabTestId,
  workbookAddRowButtonTestId,
  workbookInlineDraftRowTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import {
  normalizeViewRowV1,
  requireViewContract,
} from "@cartulary/view-contracts";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  fullWorkbookViewRow,
  successEnvelope,
  viewRowsEnvelopeForView,
  type WorkbookViewApiRow,
  workbookCollectionValue,
} from "../testing/timelineWorkbookTestSupport";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  notesViewSchemaId,
  timelineViewSchemaId,
} from "./models/workbookSurfaceRegistry";
import { readCollectionItems } from "./timeline/models/workbookMentionChips";
import { WorkbookShell } from "./WorkbookShell";

vi.mock(
  "@cartulary/grid-adapter",
  async () => import("@cartulary/grid-adapter/test-support"),
);

const exactScenarioTitle =
  "FE-I-P5-01 Verify Hosts, Identities, and Notes grids render contract-derived columns and preserve mention/entity provenance through edit and refresh.";

const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);
const notesContract = requireViewContract(notesViewSchemaId);
const timelineContract = requireViewContract(timelineViewSchemaId);

function collectionValue(items: readonly Record<string, unknown>[] = []) {
  return workbookCollectionValue(false, items);
}

function headerFieldKeys(viewSchemaId: string): string[] {
  const grid = screen.getByTestId(gridShellTestId(viewSchemaId));
  return Array.from(
    grid.querySelectorAll('[role="columnheader"] [data-grid-field-key]'),
  ).map((node) => node.getAttribute("data-grid-field-key") ?? "");
}

describe("FE-I-P5-01 Hosts, Identities, Notes grid provenance integration", () => {
  let fetchMock: ReturnType<typeof vi.fn>;
  let rowsByView: Record<string, WorkbookViewApiRow[]>;
  let patchRequests: Array<{ body: unknown; recordId: string }>;
  let createRequests: Array<{
    body: Record<string, unknown>;
    viewSchemaId: string;
  }>;
  let createdRecordCounts: Record<string, number>;

  function requireFirstViewRow(viewSchemaId: string): WorkbookViewApiRow {
    const row = rowsByView[viewSchemaId]?.[0];
    if (!row) {
      throw new Error(`test fixture missing first row for ${viewSchemaId}`);
    }
    return row;
  }

  beforeEach(() => {
    window.history.replaceState({}, "", "/");
    patchRequests = [];
    createRequests = [];
    createdRecordCounts = {};
    rowsByView = {
      [hostsViewSchemaId]: [
        fullWorkbookViewRow(hostsContract, "host-1", 3, {
          "host.display_name": "Gateway Host",
          "host.hostname": "gw-01",
          "host.aliases": collectionValue([
            {
              item_ref: "entity_alias:host-alias-1",
              item_kind: "alias",
              display_text: "VPN Gateway",
              alias_text: "VPN Gateway",
            },
          ]),
          "host.host_state": "canonical",
          "host.linked_event_count": 1,
          "host.evidence_count": 0,
          "host.location": "Datacenter A",
          "host.os_platform": "Linux",
          "host.business_owner": "SOC",
          "host.criticality": "high",
          "host.containment_status": "monitoring",
          "host.edited_at": "2026-04-24T15:00:00.000Z",
        }),
      ],
      [identitiesViewSchemaId]: [
        fullWorkbookViewRow(identitiesContract, "identity-1", 4, {
          "identity.display_name": "Alex Analyst",
          "identity.upn": "alex.analyst@example.test",
          "identity.email": "alex.analyst@example.test",
          "identity.sam_account_name": "aalyst",
          "identity.aliases": collectionValue([
            {
              item_ref: "entity_alias:identity-alias-1",
              item_kind: "alias",
              display_text: "Analyst Alex",
              alias_text: "Analyst Alex",
            },
          ]),
          "identity.identity_state": "canonical",
          "identity.linked_event_count": 1,
          "identity.evidence_count": 0,
          "identity.privilege_level": "standard",
          "identity.mfa_state": "enabled",
          "identity.reset_status": "not_requested",
          "identity.edited_at": "2026-04-24T15:00:00.000Z",
        }),
      ],
      [notesViewSchemaId]: [
        fullWorkbookViewRow(notesContract, "note-1", 5, {
          "note.title": "Investigation note",
          "note.body": "Initial note body",
          "note.tags": collectionValue([
            {
              item_ref: "record_tag:note-1:tag-1",
              item_kind: "tag",
              display_text: "provenance",
              tag_name: "provenance",
            },
          ]),
          "note.linked_record_count": 1,
          "note.updated_at": "2026-04-24T15:05:00.000Z",
          "note.created_by_user_id": "user-1",
        }),
      ],
      [timelineViewSchemaId]: [
        fullWorkbookViewRow(timelineContract, "timeline-1", 6, {
          "timeline.activity_utc_text": "2026-04-24T15:00:00.000Z",
          "timeline.activity_synopsis_text": "Gateway login by analyst",
          "timeline.raw_activity_text": "VPN Gateway login by Analyst Alex",
          "timeline.host_refs": {
            kind: "collection_value_v1",
            ordered: true,
            items: [
              {
                item_ref: "entity_mention:host-mention-1",
                entity_type: "host",
                item_kind: "resolved_ref",
                display_text: "Gateway Host",
                raw_text: " vpn   gateway ",
                resolved_record_id: "host-1",
                resolution_method: "auto_match",
                auto_resolved: true,
                provenance: "auto_match",
                confidence: 100,
                matched_alias_text: "VPN Gateway",
              },
              {
                item_ref: "entity_mention:host-unresolved-1",
                entity_type: "host",
                item_kind: "unresolved_mention",
                display_text: "Unmatched Host",
                raw_text: "Unmatched Host",
                resolved_record_id: null,
                resolution_method: null,
                auto_resolved: false,
                provenance: null,
                confidence: null,
                matched_alias_text: null,
              },
            ],
          },
          "timeline.identity_refs": {
            kind: "collection_value_v1",
            ordered: true,
            items: [
              {
                item_ref: "entity_mention:identity-mention-1",
                entity_type: "identity",
                item_kind: "resolved_ref",
                display_text: "Alex Analyst",
                raw_text: " Analyst Alex ",
                resolved_record_id: "identity-1",
                resolution_method: "explicit_resolve_route",
                auto_resolved: false,
                provenance: "manual",
                confidence: 87,
                matched_alias_text: "Analyst Alex",
              },
              {
                item_ref: "entity_mention:identity-unresolved-1",
                entity_type: "identity",
                item_kind: "unresolved_mention",
                display_text: "Unmatched Identity",
                raw_text: "Unmatched Identity",
                resolved_record_id: null,
                resolution_method: null,
                auto_resolved: false,
                provenance: null,
                confidence: null,
                matched_alias_text: null,
              },
            ],
          },
          "timeline.evidence_count": 0,
          "timeline.tags": collectionValue(),
          "timeline.attached_evidence_ids": collectionValue(),
          "timeline.edited_at": "2026-04-24T15:05:00.000Z",
          "timeline.recorded_at": "2026-04-24T15:00:00.000Z",
          "timeline.activity_sort_ts": "2026-04-24T15:00:00.000Z",
          "timeline.capture_state": "rough",
          "timeline.replacement_record_id": null,
          "timeline.date_entered_sort_day": "2026-04-24",
          "timeline.has_evidence": false,
          "timeline.has_unresolved_mentions": false,
        }),
      ],
    };

    fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();
      if (url.endsWith("/api/v1/auth/session")) {
        return successEnvelope({
          user_id: "user-1",
          memberships: [{ incident_id: "incident-1", role: "admin" }],
        });
      }
      if (url.endsWith("/api/v1/incidents/incident-1")) {
        return successEnvelope({
          incident_id: "incident-1",
          incident_key: "IR-1",
          title: "Incident 1",
          incident_version: 1,
        });
      }
      if (url.endsWith("/api/v1/incidents/incident-1/memberships")) {
        return successEnvelope({ memberships: [] });
      }
      if (url.includes("/api/v1/incidents/incident-1/workbook-preferences/")) {
        return successEnvelope({
          default_sheet_ref: null,
          home_sheet_ref: null,
        });
      }
      if (url.includes("/api/v1/incidents/incident-1/workbook-startup")) {
        return successEnvelope({
          incident_id: "incident-1",
          selected_sheet_ref: { kind: "view_schema", id: timelineViewSchemaId },
          selected_view_schema_id: timelineViewSchemaId,
          selected_saved_view: null,
          source: "timeline",
          cleared_pointers: [],
          home_sheet_ref: null,
          default_sheet_ref: null,
        });
      }
      if (
        method === "GET" &&
        url.includes("/api/v1/incidents/incident-1/saved-views")
      ) {
        return successEnvelope({ saved_views: [] });
      }
      const queryMatch = url.match(
        /\/api\/v1\/incidents\/incident-1\/views\/([^/]+)\/query(?:\?.*)?$/,
      );
      if (queryMatch) {
        return viewRowsEnvelopeForView(
          decodeURIComponent(queryMatch[1] ?? ""),
          rowsByView,
        );
      }
      const createMatch = url.match(
        /\/api\/v1\/incidents\/incident-1\/views\/([^/]+)\/rows$/,
      );
      if (method === "POST" && createMatch) {
        const viewSchemaId = decodeURIComponent(createMatch[1] ?? "");
        const body = JSON.parse(String(init?.body ?? "{}")) as Record<
          string,
          unknown
        >;
        createRequests.push({ body, viewSchemaId });
        const contract =
          viewSchemaId === hostsViewSchemaId
            ? hostsContract
            : viewSchemaId === identitiesViewSchemaId
              ? identitiesContract
              : null;
        if (contract) {
          createdRecordCounts[viewSchemaId] =
            (createdRecordCounts[viewSchemaId] ?? 0) + 1;
          const recordPrefix =
            viewSchemaId === hostsViewSchemaId ? "host" : "identity";
          const recordId = `${recordPrefix}-created-${createdRecordCounts[viewSchemaId]}`;
          const valueOverrides = Object.fromEntries(
            Object.entries(body).filter(([key]) => key !== "client_txn_id"),
          );
          const row = fullWorkbookViewRow(
            contract,
            recordId,
            1,
            valueOverrides,
          );
          rowsByView[viewSchemaId] = [...(rowsByView[viewSchemaId] ?? []), row];
          return successEnvelope({
            view_schema_id: viewSchemaId,
            change_set_id: `change-${recordId}`,
            row,
          });
        }
      }
      const patchMatch = url.match(/\/api\/v1\/records\/([^/?]+)$/);
      if (method === "PATCH" && patchMatch) {
        const recordId = patchMatch[1] ?? "";
        const body = JSON.parse(String(init?.body ?? "{}")) as {
          changes?: Array<{ field_key?: string; value?: unknown }>;
          view_schema_id?: string;
        };
        patchRequests.push({ body, recordId });
        const rows = rowsByView[body.view_schema_id ?? ""] ?? [];
        const row = rows.find((candidate) => candidate.record_id === recordId);
        const change = body.changes?.[0];
        if (row && change?.field_key) {
          row.row_version += 1;
          row.cells[change.field_key] = { value: change.value };
          return successEnvelope({
            view_schema_id: body.view_schema_id,
            change_set_id: "change-entity-edit",
            row,
          });
        }
      }
      return successEnvelope({});
    });
    vi.stubGlobal("fetch", fetchMock);
    vi.stubGlobal(
      "WebSocket",
      class {
        onmessage: ((event: MessageEvent) => void) | null = null;

        close() {}
      } as unknown as typeof WebSocket,
    );
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it(exactScenarioTitle, async () => {
    const hostFixtureRow = requireFirstViewRow(hostsViewSchemaId);
    const identityFixtureRow = requireFirstViewRow(identitiesViewSchemaId);
    const malformedHost = {
      ...hostFixtureRow,
      cells: { ...hostFixtureRow.cells },
    };
    delete malformedHost.cells["host.edited_at"];
    expect(() =>
      normalizeViewRowV1(hostsContract, malformedHost, "missing host cell"),
    ).toThrow(/missing cell host\.edited_at/iu);
    expect(() =>
      normalizeViewRowV1(
        identitiesContract,
        {
          ...identityFixtureRow,
          cells: {
            ...identityFixtureRow.cells,
            row_version: { value: 4 },
          },
        },
        "technical identity cell",
      ),
    ).toThrow(/technical cell row_version/iu);

    render(<WorkbookShell incidentId="incident-1" />);

    fireEvent.click(screen.getByTestId(surfaceTabTestId(hostsViewSchemaId)));
    await screen.findByTestId(rowCellTestId("host-1", "host.display_name"));
    expect(headerFieldKeys(hostsViewSchemaId)).toEqual(
      hostsContract.defaultVisibleFields,
    );
    expect(headerFieldKeys(hostsViewSchemaId)).not.toContain("row_version");
    expect(
      screen.getByTestId(rowCellTestId("host-1", "host.aliases")).textContent,
    ).toContain("VPN Gateway");

    fireEvent.click(
      screen.getByTestId(workbookInspectorToggleTestId(hostsViewSchemaId)),
    );
    const hostGridOverlayShell = screen.getByTestId(
      gridShellTestId(hostsViewSchemaId),
    ).parentElement;
    expect(hostGridOverlayShell).toBeInstanceOf(HTMLElement);
    expect((hostGridOverlayShell as HTMLElement).style.blockSize).toBe("100%");
    const hostWorkArea = hostGridOverlayShell?.parentElement;
    expect(hostWorkArea).toBeInstanceOf(HTMLElement);
    expect((hostWorkArea as HTMLElement).style.position).toBe("relative");
    expect((hostWorkArea as HTMLElement).style.gridTemplateRows).toBe(
      "minmax(0, 1fr)",
    );
    const hostInspector = screen.getByTestId(entityInspectorTestId("host"));
    const hostInspectorSlot = hostInspector.parentElement as HTMLElement;
    expect(hostInspectorSlot.style.position).toBe("absolute");
    expect(["0", "0px"]).toContain(
      hostInspectorSlot.style.getPropertyValue("inset-block"),
    );
    expect(hostInspector.style.blockSize).toBe("100%");
    expect(hostInspector.style.overflow).toBe("auto");
    fireEvent.change(
      screen.getByTestId(genericEditRecordSelectTestId(hostsViewSchemaId)),
      { target: { value: "host-1" } },
    );
    fireEvent.change(
      screen.getByTestId(genericEditFieldSelectTestId(hostsViewSchemaId)),
      { target: { value: "host.display_name" } },
    );
    fireEvent.change(
      screen.getByTestId(genericEditValueTestId(hostsViewSchemaId)),
      { target: { value: "Gateway Host Edited" } },
    );
    fireEvent.click(
      screen.getByTestId(genericEditSubmitTestId(hostsViewSchemaId)),
    );
    await waitFor(() => {
      expect(
        screen.getByTestId(rowCellTestId("host-1", "host.display_name"))
          .textContent,
      ).toBe("Gateway Host Edited");
    });
    expect(patchRequests[0]).toMatchObject({
      recordId: "host-1",
      body: {
        view_schema_id: hostsViewSchemaId,
        base_row_version: 3,
        client_txn_id: expect.stringMatching(
          /^entity-patch-cartulary\.view\.hosts\.v1-\d+$/u,
        ),
        changes: [
          { field_key: "host.display_name", value: "Gateway Host Edited" },
        ],
      },
    });

    const hostAddRowButton = screen.getByTestId(
      workbookAddRowButtonTestId(hostsViewSchemaId),
    );
    fireEvent.click(hostAddRowButton);
    expect(
      await screen.findByTestId(
        workbookInlineDraftRowTestId(hostsViewSchemaId),
      ),
    ).toBeInstanceOf(HTMLElement);
    const hostCreateField = screen.getByTestId(
      genericCreateFieldTestId("host.display_name"),
    );
    await waitFor(() => {
      expect(document.activeElement).toBe(hostCreateField);
    });
    fireEvent.change(hostCreateField, {
      target: { value: "Created Host" },
    });
    fireEvent.click(
      screen.getByTestId(genericCreateSubmitTestId(hostsViewSchemaId)),
    );
    await waitFor(() => {
      expect(
        screen.getByTestId(rowCellTestId("host-created-1", "host.display_name"))
          .textContent,
      ).toBe("Created Host");
    });
    expect(createRequests[0]).toMatchObject({
      viewSchemaId: hostsViewSchemaId,
      body: {
        client_txn_id: expect.stringMatching(
          /^entity-create-cartulary\.view\.hosts\.v1-\d+$/u,
        ),
        "host.display_name": "Created Host",
      },
    });

    fireEvent.click(
      screen.getByTestId(surfaceTabTestId(identitiesViewSchemaId)),
    );
    expect(headerFieldKeys(identitiesViewSchemaId)).toEqual(
      identitiesContract.defaultVisibleFields,
    );
    expect(headerFieldKeys(identitiesViewSchemaId)).not.toContain(
      "row_version",
    );
    expect(
      screen.getByTestId(rowCellTestId("identity-1", "identity.aliases"))
        .textContent,
    ).toContain("Analyst Alex");

    fireEvent.click(
      screen.getByTestId(workbookInspectorToggleTestId(identitiesViewSchemaId)),
    );
    fireEvent.change(
      screen.getByTestId(genericEditRecordSelectTestId(identitiesViewSchemaId)),
      { target: { value: "identity-1" } },
    );
    fireEvent.change(
      screen.getByTestId(genericEditFieldSelectTestId(identitiesViewSchemaId)),
      { target: { value: "identity.display_name" } },
    );
    fireEvent.change(
      screen.getByTestId(genericEditValueTestId(identitiesViewSchemaId)),
      { target: { value: "Alex Analyst Edited" } },
    );
    fireEvent.click(
      screen.getByTestId(genericEditSubmitTestId(identitiesViewSchemaId)),
    );
    await waitFor(() => {
      expect(
        screen.getByTestId(rowCellTestId("identity-1", "identity.display_name"))
          .textContent,
      ).toBe("Alex Analyst Edited");
    });
    expect(patchRequests[1]).toMatchObject({
      recordId: "identity-1",
      body: {
        view_schema_id: identitiesViewSchemaId,
        base_row_version: 4,
        client_txn_id: expect.stringMatching(
          /^entity-patch-cartulary\.view\.identities\.v1-\d+$/u,
        ),
        changes: [
          {
            field_key: "identity.display_name",
            value: "Alex Analyst Edited",
          },
        ],
      },
    });

    const identityAddRowButton = screen.getByTestId(
      workbookAddRowButtonTestId(identitiesViewSchemaId),
    );
    fireEvent.click(identityAddRowButton);
    expect(
      await screen.findByTestId(
        workbookInlineDraftRowTestId(identitiesViewSchemaId),
      ),
    ).toBeInstanceOf(HTMLElement);
    const identityCreateField = screen.getByTestId(
      genericCreateFieldTestId("identity.display_name"),
    );
    await waitFor(() => {
      expect(document.activeElement).toBe(identityCreateField);
    });
    fireEvent.change(identityCreateField, {
      target: { value: "Created Identity" },
    });
    fireEvent.click(
      screen.getByTestId(genericCreateSubmitTestId(identitiesViewSchemaId)),
    );
    await waitFor(() => {
      expect(
        screen.getByTestId(
          rowCellTestId("identity-created-1", "identity.display_name"),
        ).textContent,
      ).toBe("Created Identity");
    });
    expect(createRequests[1]).toMatchObject({
      viewSchemaId: identitiesViewSchemaId,
      body: {
        client_txn_id: expect.stringMatching(
          /^entity-create-cartulary\.view\.identities\.v1-\d+$/u,
        ),
        "identity.display_name": "Created Identity",
      },
    });

    fireEvent.click(screen.getByTestId(surfaceTabTestId(notesViewSchemaId)));
    await screen.findByTestId(rowCellTestId("note-1", "note.title"));
    expect(headerFieldKeys(notesViewSchemaId)).toEqual(
      notesContract.defaultVisibleFields,
    );
    expect(headerFieldKeys(notesViewSchemaId)).not.toContain("row_version");
    expect(
      screen.getByTestId(rowCellTestId("note-1", "note.tags")).textContent,
    ).toContain("provenance");

    const timelineFixtureRow = requireFirstViewRow(timelineViewSchemaId);
    const beforeHostMentions = readCollectionItems(
      timelineFixtureRow,
      "timeline.host_refs",
    );
    const beforeIdentityMentions = readCollectionItems(
      timelineFixtureRow,
      "timeline.identity_refs",
    );
    const beforeHostMention = beforeHostMentions.find(
      (item) => item.rawText === " vpn   gateway ",
    );
    const beforeUnresolvedHostMention = beforeHostMentions.find(
      (item) => item.rawText === "Unmatched Host",
    );
    const beforeUnresolvedIdentityMention = beforeIdentityMentions.find(
      (item) => item.rawText === "Unmatched Identity",
    );
    expect(beforeHostMention).toBeDefined();
    expect(beforeUnresolvedHostMention).toMatchObject({
      itemKind: "unresolved_mention",
      resolvedRecordId: null,
    });
    expect(beforeUnresolvedIdentityMention).toMatchObject({
      itemKind: "unresolved_mention",
      resolvedRecordId: null,
    });
    fireEvent.click(
      screen.getByTestId(workbookInspectorToggleTestId(notesViewSchemaId)),
    );
    fireEvent.change(
      screen.getByTestId(genericEditRecordSelectTestId(notesViewSchemaId)),
      { target: { value: "note-1" } },
    );
    fireEvent.change(
      screen.getByTestId(genericEditFieldSelectTestId(notesViewSchemaId)),
      { target: { value: "note.body" } },
    );
    fireEvent.change(
      screen.getByTestId(genericEditValueTestId(notesViewSchemaId)),
      { target: { value: "Edited note body" } },
    );
    fireEvent.click(
      screen.getByTestId(genericEditSubmitTestId(notesViewSchemaId)),
    );
    await waitFor(() => {
      expect(
        screen.getByTestId(rowCellTestId("note-1", "note.body")).textContent,
      ).toBe("Edited note body");
    });
    expect(patchRequests[2]).toMatchObject({
      recordId: "note-1",
      body: {
        view_schema_id: notesViewSchemaId,
        base_row_version: 5,
        client_txn_id: expect.stringMatching(
          /^generic-patch-cartulary\.view\.notes\.v1-\d+$/u,
        ),
        changes: [{ field_key: "note.body", value: "Edited note body" }],
      },
    });

    const afterHostMentions = readCollectionItems(
      timelineFixtureRow,
      "timeline.host_refs",
    );
    const afterIdentityMentions = readCollectionItems(
      timelineFixtureRow,
      "timeline.identity_refs",
    );
    const afterHostMention = afterHostMentions.find(
      (item) => item.rawText === " vpn   gateway ",
    );
    const afterUnresolvedHostMention = afterHostMentions.find(
      (item) => item.rawText === "Unmatched Host",
    );
    const afterUnresolvedIdentityMention = afterIdentityMentions.find(
      (item) => item.rawText === "Unmatched Identity",
    );
    expect(afterHostMention).toEqual(beforeHostMention);
    expect(afterUnresolvedHostMention).toEqual(beforeUnresolvedHostMention);
    expect(afterUnresolvedIdentityMention).toEqual(
      beforeUnresolvedIdentityMention,
    );
  });
});
