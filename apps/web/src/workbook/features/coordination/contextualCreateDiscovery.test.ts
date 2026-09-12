import { viewSchemaRegistry } from "@cartulary/protocol-ts/view-schemas";
import {
  decisionsViewSchemaId,
  findingsViewSchemaId,
  listViewContracts,
  partiesViewSchemaId,
  requireViewContract,
  type ViewContract,
} from "@cartulary/view-contracts";
import { afterEach, describe, expect, it, vi } from "vitest";
import { jsonResponse } from "../../../testing/fetchMockTestSupport";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { createContextualCreateReader } from "../../adapters/createContextualCreateReader";
import { emptyWorkbookQueryState } from "../../models/workbookQuery";
import {
  contextualCreateDraft,
  isContextualCreateFeature,
} from "./contextualCreateModel";

afterEach(() => vi.unstubAllGlobals());
const incident = "10000000-0000-4000-8000-000000000001",
  id = "20000000-0000-4000-8000-000000000002";
const queryState = emptyWorkbookQueryState();
describe("contextual target reference discovery", () => {
  it("verifies every declared route and seed structurally against public discovery and rejects changed capabilities", async () => {
    let unavailable = false;
    vi.stubGlobal(
      "fetch",
      vi.fn(async (url: unknown) => {
        const view = String(url).split("/view-schemas/")[1];
        const contract = requireViewContract(decodeURIComponent(view ?? ""));
        const data = publicSchema(contract);
        return jsonResponse({
          data: { ...data, create_capable: !unavailable },
          meta: { request_id: "discovery" },
        });
      }),
    );
    const reader = createContextualCreateReader({
      apiBase: undefined,
      incidentId: incident,
      recheckAuthority: () => {},
    });
    let checked = 0;
    for (const source of listViewContracts()) {
      for (const feature of source.inspectorConfig.featureGroups.filter(
        (group) => isContextualCreateFeature(group.featureGroupKey),
      )) {
        const draft = contextualCreateDraft(
          1,
          {
            actorId: id,
            incidentId: incident,
            sessionIdentity: "session",
            role: "editor",
            closed: false,
          },
          {
            cells: {},
            subject: {
              kind: "live",
              recordId: id,
              rowVersion: 1,
              viewSchemaId: source.viewSchemaId,
              label: "Source",
              surfaceLabel: source.title,
            },
          },
          feature,
          { kind: "view_schema", id: source.viewSchemaId },
        );
        if (!draft) throw new Error("Expected contextual draft");
        await expect(
          reader.verify(draft, new AbortController().signal),
        ).resolves.toBeUndefined();
        unavailable = true;
        await expect(
          reader.verify(draft, new AbortController().signal),
        ).rejects.toThrow("capability changed");
        unavailable = false;
        checked++;
      }
    }
    expect(checked).toBe(23);
  });
  it("reads pages of 100 for every implemented record surface, including Parties, Decisions and optional collections", async () => {
    const fetch = vi.fn(async (_url: unknown, init: RequestInit) => {
      const view = String(_url).split("/views/")[1]?.split("/")[0];
      if (!view) throw new Error("Expected view query");
      const contract = requireViewContract(decodeURIComponent(view));
      const body = JSON.parse(init.body as string);
      return jsonResponse({
        data: {
          incident_id: incident,
          view_schema_id: contract.viewSchemaId,
          rows: [fullWorkbookViewRow(contract, id, 1, {})],
        },
        meta: {
          request_id: "query",
          query: { sort: body.sort ?? [], filters: body.filters ?? [] },
          paging: { limit: 100, has_more: true, next_cursor: "next" },
        },
      });
    });
    vi.stubGlobal("fetch", fetch);
    const recheckAuthority = vi.fn(),
      reader = createContextualCreateReader({
        apiBase: undefined,
        incidentId: incident,
        recheckAuthority,
      });
    for (const contract of listViewContracts()) {
      const result = await reader.page({
        viewSchemaId: contract.viewSchemaId,
        queryState,
        cursor: null,
        signal: new AbortController().signal,
      });
      expect(result).toMatchObject({
        kind: "accepted",
        value: {
          hasMore: true,
          nextCursor: "next",
          candidates: [{ recordId: id, viewSchemaId: contract.viewSchemaId }],
        },
      });
    }
    expect(
      fetch.mock.calls.every(
        ([, init]) => JSON.parse(init.body as string).limit === 100,
      ),
    ).toBe(true);
    expect(
      fetch.mock.calls
        .map(([url]) => String(url))
        .some((url) => url.includes(encodeURIComponent(partiesViewSchemaId))),
    ).toBe(true);
    expect(
      fetch.mock.calls
        .map(([url]) => String(url))
        .some((url) => url.includes(encodeURIComponent(decisionsViewSchemaId))),
    ).toBe(true);
    expect(
      fetch.mock.calls
        .map(([url]) => String(url))
        .some((url) => url.includes(encodeURIComponent(findingsViewSchemaId))),
    ).toBe(true);
    expect(recheckAuthority).not.toHaveBeenCalled();
  });
  it("paginates membership inventory and rejects mismatched incident and cursor responses", async () => {
    const membership = {
      incident_id: incident,
      user_id: id,
      display_name: "Incident member",
      role: "editor",
      added_by_user_id: id,
      joined_at: "2026-09-12T20:00:00Z",
      membership_version: 1,
      updated_at: "2026-09-12T20:00:00Z",
      updated_by_user_id: null,
    };
    const fetch = vi
      .fn()
      .mockResolvedValueOnce(
        jsonResponse({
          data: { memberships: [membership] },
          meta: {
            request_id: "members",
            paging: { limit: 100, has_more: true, next_cursor: "opaque" },
          },
        }),
      )
      .mockResolvedValueOnce(
        jsonResponse({
          data: { memberships: [] },
          meta: {
            request_id: "members2",
            paging: { limit: 100, has_more: false, next_cursor: null },
          },
        }),
      )
      .mockResolvedValueOnce(
        jsonResponse({
          data: { memberships: [{ ...membership, incident_id: id }] },
          meta: {
            request_id: "members3",
            paging: { limit: 100, has_more: false, next_cursor: null },
          },
        }),
      );
    vi.stubGlobal("fetch", fetch);
    const reader = createContextualCreateReader({
      apiBase: undefined,
      incidentId: incident,
      recheckAuthority: () => {},
    });
    const input = {
      viewSchemaId: "incident_members",
      queryState,
      cursor: null,
      signal: new AbortController().signal,
    };
    expect(await reader.page(input)).toMatchObject({
      kind: "accepted",
      value: {
        candidates: [{ recordId: id, displayText: "Incident member" }],
        hasMore: true,
        nextCursor: "opaque",
      },
    });
    expect(await reader.page({ ...input, cursor: "opaque" })).toMatchObject({
      kind: "accepted",
      value: { candidates: [], hasMore: false },
    });
    expect(String(fetch.mock.calls[1]?.[0])).toContain("cursor_token=opaque");
    expect(String(fetch.mock.calls[1]?.[0])).toContain("limit=100");
    expect(await reader.page(input)).toMatchObject({ kind: "rejected" });
  });
});

// Project protocol fixtures from typed contracts; object key order is deliberately
// reversed to exercise structural comparison rather than serialization identity.
function publicSchema(contract: ViewContract) {
  return {
    view_schema_id: contract.viewSchemaId,
    surface_kind: contract.surfaceKind,
    title: contract.title,
    source_record_types:
      viewSchemaRegistry.view_schemas.find(
        (view) => view.view_schema_id === contract.viewSchemaId,
      )?.source_record_types ?? [],
    technical_fields: contract.technicalFields,
    required_reference_pack_keys: contract.requiredReferencePackKeys,
    default_sort: snakeKeys(contract.defaultSort),
    sort_fields: contract.sortFields,
    sort_null_order: contract.sortNullOrder,
    filter_fields: contract.filterFields,
    synthetic_filter_predicates: [],
    grouping_fields: contract.groupingFields,
    create_capable: contract.createCapable,
    create_inputs: snakeKeys(contract.createInputs),
    inline_create: {
      minimum_create_field_sets: contract.minimumCreateFieldSets,
      permits_zero_field_create: contract.permitsZeroFieldCreate,
    },
    inspector_config: snakeKeys(contract.inspectorConfig),
    fields: contract.fields.map(({ writeAction: _writeAction, ...field }) =>
      snakeKeys(field),
    ),
  };
}
function snakeKeys(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(snakeKeys);
  if (value && typeof value === "object")
    return Object.fromEntries(
      Object.entries(value)
        .reverse()
        .map(([key, child]) => [
          key.replace(/[A-Z]/g, (letter) => `_${letter.toLowerCase()}`),
          snakeKeys(child),
        ]),
    );
  return value;
}
