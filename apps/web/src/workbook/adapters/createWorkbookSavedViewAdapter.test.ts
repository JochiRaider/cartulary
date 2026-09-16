import { requireViewContract } from "@cartulary/view-contracts";
import { afterEach, describe, expect, it, vi } from "vitest";
import { buildSavedViewLayoutJson } from "../models/workbookQuery";
import { savedViewChanges } from "../models/workbookSavedViewRuntime";
import type { SavedViewResource } from "../models/workbookSavedViews";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import type { WorkbookSavedViewDefinition } from "../ports/WorkbookSavedViewPort";
import { createWorkbookSavedViewAdapter } from "./createWorkbookSavedViewAdapter";

const incidentId = "00000000-0000-4000-8000-000000000001";
const otherIncidentId = "00000000-0000-4000-8000-000000000002";
const savedViewId = "10000000-0000-4000-8000-000000000001";
const userId = "20000000-0000-4000-8000-000000000001";
const now = "2026-07-31T20:00:00Z";
const later = "2026-07-31T20:01:00Z";
const definition: WorkbookSavedViewDefinition = {
  displayName: "Analyst timeline",
  layoutJson: buildSavedViewLayoutJson(
    requireViewContract(timelineViewSchemaId),
  ),
  queryJson: {
    filters: [],
    sort: [{ direction: "desc", field_key: "timeline.activity_sort_ts" }],
  },
  scope: "private",
  viewSchemaId: timelineViewSchemaId,
};
function savedViewResource(
  overrides: Partial<SavedViewResource> = {},
): SavedViewResource {
  return {
    created_at: now,
    updated_at: now,
    display_name: definition.displayName,
    incident_id: incidentId,
    layout_json: definition.layoutJson,
    owner_user_id: userId,
    query_json: definition.queryJson,
    saved_view_id: savedViewId,
    saved_view_version: 1,
    scope: "private",
    view_schema_id: timelineViewSchemaId,
    ...overrides,
  };
}
function response(data: unknown, status = 200) {
  return new Response(JSON.stringify(data), {
    status,
    headers: { "content-type": "application/json" },
  });
}
function envelope(
  data: unknown,
  paging?: { has_more: boolean; limit: number; next_cursor: string | null },
) {
  return {
    data,
    meta: { request_id: "req-saved-view", ...(paging ? { paging } : {}) },
  };
}
const signal = () => new AbortController().signal;
const adapter = () =>
  createWorkbookSavedViewAdapter({ apiBase: undefined, incidentId });
function errorResponse(code: string, status: number, details = {}) {
  return response(
    {
      error: {
        code,
        status,
        details,
        message: code,
        request_id: "req-error",
        retryable: false,
      },
    },
    status,
  );
}
afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

describe("Workbook saved-view adapter", () => {
  it("reads a resource by ID and filters discovery before accepting correlated pages", async () => {
    const base = savedViewResource();
    const fetchMock = vi.fn().mockResolvedValueOnce(response(envelope(base)));
    vi.stubGlobal("fetch", fetchMock);
    await expect(
      adapter().getResource({ savedViewId, signal: signal() }),
    ).resolves.toEqual({ kind: "accepted", value: base });
    expect(fetchMock).toHaveBeenLastCalledWith(
      `/api/v1/incidents/${incidentId}/saved-views/${savedViewId}`,
      expect.objectContaining({ method: "GET" }),
    );
    fetchMock.mockResolvedValueOnce(
      response(
        envelope(
          savedViewResource({
            saved_view_id: "10000000-0000-4000-8000-000000000099",
          }),
        ),
      ),
    );
    await expect(
      adapter().getResource({ savedViewId, signal: signal() }),
    ).resolves.toMatchObject({
      kind: "rejected",
      failure: { kind: "invalid_contract" },
    });
    fetchMock.mockResolvedValueOnce(
      response(
        envelope(
          { saved_views: [base] },
          { has_more: false, limit: 50, next_cursor: null },
        ),
      ),
    );
    await expect(
      adapter().listPage({
        viewSchemaId: timelineViewSchemaId,
        cursorToken: null,
        limit: 50,
        signal: signal(),
      }),
    ).resolves.toMatchObject({ kind: "accepted" });
    expect(fetchMock).toHaveBeenLastCalledWith(
      `/api/v1/incidents/${incidentId}/saved-views?limit=50&view_schema_id=${timelineViewSchemaId}`,
      expect.objectContaining({ method: "GET" }),
    );
    fetchMock.mockResolvedValueOnce(
      response(
        envelope(
          { saved_views: [base] },
          { has_more: false, limit: 50, next_cursor: null },
        ),
      ),
    );
    await expect(
      adapter().listPage({
        viewSchemaId: "cartulary.view.notes.v1",
        cursorToken: null,
        limit: 50,
        signal: signal(),
      }),
    ).resolves.toMatchObject({
      kind: "rejected",
      failure: { kind: "invalid_contract" },
    });
  });

  it("accepts a normalized no-op with unchanged version and timestamp", async () => {
    const base = savedViewResource();
    const fetchMock = vi
      .fn()
      .mockImplementation(() => Promise.resolve(response(envelope(base))));
    vi.stubGlobal("fetch", fetchMock);
    for (const changes of [
      {},
      { displayName: ` ${definition.displayName} ` },
      { queryJson: { ...definition.queryJson } },
    ]) {
      await expect(
        adapter().patch({ base, changes, signal: signal() }),
      ).resolves.toEqual({ kind: "accepted", value: base });
    }
    expect(JSON.parse(fetchMock.mock.calls[0]?.[1].body)).toEqual({
      base_saved_view_version: 1,
    });
  });

  it("projects explicit paging and returns only correlated saved-view resources", async () => {
    const base = savedViewResource();
    const fetchMock = vi
      .fn()
      .mockResolvedValue(
        response(
          envelope(
            { saved_views: [base] },
            { has_more: true, limit: 2, next_cursor: "cursor-2" },
          ),
        ),
      );
    vi.stubGlobal("fetch", fetchMock);
    const port = createWorkbookSavedViewAdapter({
      apiBase: "/base",
      incidentId,
    });
    await expect(
      port.listPage({
        viewSchemaId: timelineViewSchemaId,
        cursorToken: "cursor-1",
        limit: 2,
        signal: signal(),
      }),
    ).resolves.toEqual({
      kind: "accepted",
      value: { nextCursor: "cursor-2", savedViews: [base] },
    });
    expect(fetchMock).toHaveBeenCalledWith(
      `/base/api/v1/incidents/${incidentId}/saved-views?cursor_token=cursor-1&limit=2&view_schema_id=${timelineViewSchemaId}`,
      expect.objectContaining({ method: "GET" }),
    );
  });

  it("fails closed on malformed paging, cross-incident resources, and invalid versions", async () => {
    const malformed = [
      envelope(
        { saved_views: [] },
        { has_more: true, limit: 2, next_cursor: null },
      ),
      ...[
        { incident_id: otherIncidentId },
        { saved_view_version: 0 },
        { owner_user_id: null },
        { layout_json: {} },
        { query_json: { sort: [], filters: [], group_by: null } },
        { query_json: { sort: [], filters: [], selection: "row" } },
        {
          layout_json: {
            ...definition.layoutJson,
            column_order: ["record_id"],
          },
        },
      ].map((overrides) =>
        envelope(
          { saved_views: [savedViewResource(overrides)] },
          { has_more: false, limit: 2, next_cursor: null },
        ),
      ),
    ];
    for (const payload of malformed) {
      vi.stubGlobal("fetch", vi.fn().mockResolvedValue(response(payload)));
      await expect(
        adapter().listPage({
          viewSchemaId: timelineViewSchemaId,
          cursorToken: null,
          limit: 2,
          signal: signal(),
        }),
      ).resolves.toMatchObject({
        kind: "rejected",
        failure: { kind: "invalid_contract" },
      });
    }
  });

  it("projects create, patch, and delete requests with CSRF and accepted correlations", async () => {
    vi.spyOn(document, "cookie", "get").mockReturnValue(
      "cartulary_csrf=csrf-token",
    );
    const base = savedViewResource();
    const updated = savedViewResource({
      display_name: "Updated timeline",
      saved_view_version: 2,
      updated_at: later,
    });
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(response(envelope(base), 201))
      .mockResolvedValueOnce(response(envelope(updated)))
      .mockResolvedValueOnce(
        response(envelope({ deleted: true, saved_view_id: savedViewId })),
      );
    vi.stubGlobal("fetch", fetchMock);
    const port = adapter();
    await expect(
      port.create({ definition, signal: signal() }),
    ).resolves.toMatchObject({ kind: "accepted", value: base });
    const changes = savedViewChanges(base, {
      ...definition,
      displayName: " Updated timeline ",
    });
    expect(changes).toEqual({ displayName: " Updated timeline " });
    await expect(
      port.patch({ base, changes, signal: signal() }),
    ).resolves.toEqual({ kind: "accepted", value: updated });
    await expect(
      port.delete({ savedViewId, scope: "private", signal: signal() }),
    ).resolves.toEqual({ kind: "accepted", value: undefined });
    for (const [index, method] of ["POST", "PATCH", "DELETE"].entries()) {
      const [url, init] = fetchMock.mock.calls[index] as [string, RequestInit];
      expect(url).toBe(
        `/api/v1/incidents/${incidentId}/saved-views${index === 0 ? "" : `/${savedViewId}`}`,
      );
      expect(init.method).toBe(method);
      expect(new Headers(init.headers).get("X-CSRF-Token")).toBe("csrf-token");
    }
    expect(JSON.parse(fetchMock.mock.calls[0]?.[1].body)).toEqual({
      display_name: definition.displayName,
      layout_json: definition.layoutJson,
      query_json: definition.queryJson,
      scope: "private",
      view_schema_id: timelineViewSchemaId,
    });
    expect(JSON.parse(fetchMock.mock.calls[1]?.[1].body)).toEqual({
      base_saved_view_version: 1,
      display_name: " Updated timeline ",
    });
  });

  it("rejects mismatched mutation acknowledgements and system-view mutation locally", async () => {
    const base = savedViewResource();
    for (const overrides of [
      { incident_id: otherIncidentId },
      { saved_view_id: otherIncidentId },
      { owner_user_id: otherIncidentId },
      { saved_view_version: 3 },
      { updated_at: later },
      { scope: "shared" as const },
    ]) {
      vi.stubGlobal(
        "fetch",
        vi
          .fn()
          .mockResolvedValue(response(envelope(savedViewResource(overrides)))),
      );
      await expect(
        adapter().patch({ base, changes: {}, signal: signal() }),
      ).resolves.toMatchObject({
        kind: "uncertain",
        failure: { kind: "invalid_contract" },
      });
    }
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValue(
          response(
            envelope(savedViewResource({ incident_id: otherIncidentId })),
            201,
          ),
        ),
    );
    await expect(
      adapter().create({ definition, signal: signal() }),
    ).resolves.toMatchObject({
      kind: "uncertain",
      failure: { kind: "invalid_contract" },
    });
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValue(
          response(envelope({ deleted: true, saved_view_id: otherIncidentId })),
        ),
    );
    await expect(
      adapter().delete({ savedViewId, scope: "private", signal: signal() }),
    ).resolves.toMatchObject({
      kind: "uncertain",
      failure: { kind: "invalid_contract" },
    });
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    await expect(
      adapter().patch({
        base: { ...base, scope: "system" },
        changes: {},
        signal: signal(),
      }),
    ).resolves.toMatchObject({
      kind: "rejected",
      failure: { kind: "validation" },
    });
    await expect(
      adapter().delete({ savedViewId, scope: "system", signal: signal() }),
    ).resolves.toMatchObject({
      kind: "rejected",
      failure: { kind: "validation" },
    });
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("preserves typed conflict validation access and uncertain outcomes", async () => {
    const base = savedViewResource();
    const cases = [
      [
        "saved_view_version_conflict",
        409,
        {
          saved_view_id: savedViewId,
          base_saved_view_version: 1,
          current_saved_view_version: 2,
        },
        "rejected",
        "conflict",
      ],
      [
        "saved_view_version_conflict",
        409,
        {
          saved_view_id: otherIncidentId,
          base_saved_view_version: 1,
          current_saved_view_version: 2,
        },
        "uncertain",
        "invalid_contract",
      ],
      [
        "invalid_mutation_payload",
        400,
        { field: "display_name", reason_code: "invalid_value" },
        "rejected",
        "validation",
      ],
      [
        "authentication_required",
        401,
        {},
        "rejected",
        "authentication_required",
      ],
      ["session_required", 401, {}, "rejected", "authentication_required"],
      ["csrf_verification_failed", 403, {}, "rejected", "authorization_denied"],
      ["authorization_denied", 403, {}, "rejected", "authorization_denied"],
      ["saved_view_not_found", 404, {}, "rejected", "unavailable_target"],
      ["internal_error", 500, {}, "uncertain", "terminal"],
    ] as const;
    for (const [code, status, details, kind, failureKind] of cases) {
      vi.stubGlobal(
        "fetch",
        vi.fn().mockResolvedValue(errorResponse(code, status, details)),
      );
      const result = await adapter().patch({
        base,
        changes: {},
        signal: signal(),
      });
      expect(result).toMatchObject({ kind, failure: { kind: failureKind } });
      if (failureKind === "conflict")
        expect(result).toMatchObject({
          failure: {
            publicCode: code,
            conflict: { savedViewId, baseVersion: 1, currentVersion: 2 },
          },
        });
      if (failureKind === "validation")
        expect(result).toMatchObject({
          failure: { field: "display_name", reason: "invalid_value" },
        });
    }

    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation(
        async () =>
          new Response("{", {
            status: 200,
            headers: { "Content-Type": "application/json" },
          }),
      ),
    );
    await expect(
      adapter().patch({ base, changes: {}, signal: signal() }),
    ).resolves.toMatchObject({
      kind: "uncertain",
      failure: { kind: "invalid_contract" },
    });
    await expect(
      adapter().listPage({
        viewSchemaId: timelineViewSchemaId,
        cursorToken: null,
        limit: 100,
        signal: signal(),
      }),
    ).resolves.toMatchObject({
      kind: "rejected",
      failure: { kind: "invalid_contract" },
    });
    vi.stubGlobal(
      "fetch",
      vi.fn().mockRejectedValue(new DOMException("aborted", "AbortError")),
    );
    await expect(
      adapter().create({ definition, signal: signal() }),
    ).resolves.toMatchObject({
      kind: "uncertain",
      failure: { kind: "transport" },
    });
  });

  it("preserves authoritative query normalization and rejects false no-op metadata", async () => {
    const base = savedViewResource();
    const canonical = {
      filters: [
        {
          field_key: "timeline.tags",
          op: "contains_any" as const,
          arg: { values: ["alpha", "beta"] },
        },
      ],
      sort: [],
    };
    const updated = savedViewResource({
      query_json: canonical,
      saved_view_version: 2,
      updated_at: later,
    });
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(response(envelope(updated))),
    );
    await expect(
      adapter().patch({
        base,
        changes: {
          queryJson: {
            ...canonical,
            filters: [
              {
                field_key: "timeline.tags",
                op: "contains_any",
                arg: { values: [" beta ", "alpha", "beta"] },
              },
            ],
          },
        },
        signal: signal(),
      }),
    ).resolves.toMatchObject({
      kind: "accepted",
      value: { query_json: canonical },
    });
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(response(envelope(updated))),
    );
    await expect(
      adapter().patch({ base, changes: {}, signal: signal() }),
    ).resolves.toMatchObject({
      kind: "uncertain",
      failure: { kind: "invalid_contract" },
    });
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValue(response(envelope({ ...updated, updated_at: now }))),
    );
    await expect(
      adapter().patch({
        base,
        changes: { queryJson: canonical },
        signal: signal(),
      }),
    ).resolves.toMatchObject({
      kind: "uncertain",
      failure: { kind: "invalid_contract" },
    });
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(response(envelope(updated))),
    );
    await expect(
      adapter().patch({
        base: updated,
        changes: {
          queryJson: {
            ...canonical,
            filters: [
              {
                field_key: "timeline.tags",
                op: "contains_any",
                arg: { values: ["beta", "alpha", "beta"] },
              },
            ],
          },
        },
        signal: signal(),
      }),
    ).resolves.toMatchObject({ kind: "accepted", value: updated });
  });
});
