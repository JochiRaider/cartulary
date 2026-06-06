import {
  genericEditFieldSelectTestId,
  genericEditRecordSelectTestId,
  genericEditSubmitTestId,
  genericEditValueTestId,
  gridShellTestId,
  rowCellTestId,
  surfaceTabTestId,
} from "@cartulary/ui-contracts";
import {
  normalizeViewRowV1,
  requireViewContract,
  type ViewContract,
  type ViewFieldContract,
} from "@cartulary/view-contracts";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { successEnvelope } from "./timelineWorkbookTestSupport";
import { WorkbookShell } from "./WorkbookShell";
import { readCollectionItems } from "./workbookMentionChips";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  notesViewSchemaId,
  timelineViewSchemaId,
} from "./workbookSurfaceRegistry";

vi.mock(
  "@cartulary/grid-adapter",
  async () => import("@cartulary/grid-adapter/test-support"),
);

const exactScenarioTitle =
  "FE-I-P5-01 Verify Hosts, Identities, and Notes grids render contract-derived columns and preserve mention/entity provenance through edit and refresh.";

type ViewApiRow = {
  record_id: string;
  row_version: number;
  cells: Record<string, { value: unknown }>;
  view_schema_id?: string;
};

type ViewRowOverrides = Record<string, unknown>;

const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);
const notesContract = requireViewContract(notesViewSchemaId);
const timelineContract = requireViewContract(timelineViewSchemaId);

function collectionValue(items: readonly Record<string, unknown>[] = []) {
  return {
    kind: "collection_value_v1",
    ordered: false,
    items: [...items],
  };
}

function defaultCellValue(field: ViewFieldContract): unknown {
  if (field.readKind === "collection") {
    return collectionValue();
  }
  if (field.readKind === "number") {
    return 0;
  }
  if (field.readKind === "boolean") {
    return false;
  }
  return null;
}

function fullViewRow(
  contract: ViewContract,
  recordId: string,
  rowVersion: number,
  overrides: ViewRowOverrides,
): ViewApiRow {
  return {
    record_id: recordId,
    row_version: rowVersion,
    view_schema_id: contract.viewSchemaId,
    cells: Object.fromEntries(
      contract.fields.map((field) => [
        field.fieldKey,
        {
          value:
            field.fieldKey in overrides
              ? overrides[field.fieldKey]
              : defaultCellValue(field),
        },
      ]),
    ),
  };
}

function headerFieldKeys(viewSchemaId: string): string[] {
  const grid = screen.getByTestId(gridShellTestId(viewSchemaId));
  return Array.from(
    grid.querySelectorAll('[role="columnheader"] [data-grid-field-key]'),
  ).map((node) => node.getAttribute("data-grid-field-key") ?? "");
}

function queryRowsForView(
  viewSchemaId: string,
  rowsByView: Record<string, ViewApiRow[]>,
) {
  return successEnvelope({
    incident_id: "incident-1",
    view_schema_id: viewSchemaId,
    rows: rowsByView[viewSchemaId] ?? [],
  });
}

describe("FE-I-P5-01 Hosts, Identities, Notes grid provenance integration", () => {
  let fetchMock: ReturnType<typeof vi.fn>;
  let rowsByView: Record<string, ViewApiRow[]>;
  let patchRequests: Array<{ body: unknown; recordId: string }>;

  beforeEach(() => {
    window.history.replaceState({}, "", "/");
    patchRequests = [];
    rowsByView = {
      [hostsViewSchemaId]: [
        fullViewRow(hostsContract, "host-1", 3, {
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
        fullViewRow(identitiesContract, "identity-1", 4, {
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
        fullViewRow(notesContract, "note-1", 5, {
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
        fullViewRow(timelineContract, "timeline-1", 6, {
          "timeline.occurred_at": "2026-04-24T15:00:00.000Z",
          "timeline.summary": "Gateway login by analyst",
          "timeline.details": "",
          "timeline.source_text": "VPN Gateway login by Analyst Alex",
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
          "timeline.sort_ts": "2026-04-24T15:00:00.000Z",
          "timeline.capture_state": "rough",
          "timeline.replacement_record_id": null,
          "timeline.occurred_day": "2026-04-24",
          "timeline.recorded_day": "2026-04-24",
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
        return queryRowsForView(
          decodeURIComponent(queryMatch[1] ?? ""),
          rowsByView,
        );
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
    const malformedHost = {
      ...rowsByView[hostsViewSchemaId]![0]!,
      cells: { ...rowsByView[hostsViewSchemaId]![0]!.cells },
    };
    delete malformedHost.cells["host.edited_at"];
    expect(() =>
      normalizeViewRowV1(hostsContract, malformedHost, "missing host cell"),
    ).toThrow(/missing cell host\.edited_at/iu);
    expect(() =>
      normalizeViewRowV1(
        identitiesContract,
        {
          ...rowsByView[identitiesViewSchemaId]![0]!,
          cells: {
            ...rowsByView[identitiesViewSchemaId]![0]!.cells,
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
    fireEvent.click(screen.getByTestId(genericEditSubmitTestId(hostsViewSchemaId)));
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

    fireEvent.click(screen.getByTestId(surfaceTabTestId(identitiesViewSchemaId)));
    expect(headerFieldKeys(identitiesViewSchemaId)).toEqual(
      identitiesContract.defaultVisibleFields,
    );
    expect(headerFieldKeys(identitiesViewSchemaId)).not.toContain("row_version");
    expect(
      screen.getByTestId(rowCellTestId("identity-1", "identity.aliases"))
        .textContent,
    ).toContain("Analyst Alex");

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

    fireEvent.click(screen.getByTestId(surfaceTabTestId(notesViewSchemaId)));
    await screen.findByTestId(rowCellTestId("note-1", "note.title"));
    expect(headerFieldKeys(notesViewSchemaId)).toEqual(
      notesContract.defaultVisibleFields,
    );
    expect(headerFieldKeys(notesViewSchemaId)).not.toContain("row_version");
    expect(
      screen.getByTestId(rowCellTestId("note-1", "note.tags")).textContent,
    ).toContain("provenance");

    const beforeHostMentions = readCollectionItems(
      rowsByView[timelineViewSchemaId]![0]!,
      "timeline.host_refs",
    );
    const beforeIdentityMentions = readCollectionItems(
      rowsByView[timelineViewSchemaId]![0]!,
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
    fireEvent.click(screen.getByTestId(genericEditSubmitTestId(notesViewSchemaId)));
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
      rowsByView[timelineViewSchemaId]![0]!,
      "timeline.host_refs",
    );
    const afterIdentityMentions = readCollectionItems(
      rowsByView[timelineViewSchemaId]![0]!,
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
