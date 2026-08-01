import { afterEach, describe, expect, it, vi } from "vitest";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import type { WorkbookSavedViewDefinition } from "../ports/WorkbookSavedViewPort";
import { createWorkbookSavedViewAdapter } from "./createWorkbookSavedViewAdapter";

const incidentId = "00000000-0000-4000-8000-000000000001";
const otherIncidentId = "00000000-0000-4000-8000-000000000002";
const savedViewId = "10000000-0000-4000-8000-000000000001";
const userId = "20000000-0000-4000-8000-000000000001";
const now = "2026-07-31T20:00:00Z";

const definition: WorkbookSavedViewDefinition = {
  displayName: "Analyst timeline",
  layoutJson: {
    column_order: ["timeline.activity_synopsis_text"],
    column_widths: [
      { field_key: "timeline.activity_synopsis_text", width_px: 320 },
    ],
    hidden_field_keys: [],
    layout_schema_id: "cartulary.layout.v1",
  },
  queryJson: {
    filters: [],
    sort: [{ direction: "desc", field_key: "timeline.activity_sort_ts" }],
  },
  scope: "private",
  viewSchemaId: timelineViewSchemaId,
};

function savedViewResource(
  overrides: Partial<ReturnType<typeof savedViewResourceBase>> = {},
) {
  return { ...savedViewResourceBase(), ...overrides };
}

function savedViewResourceBase() {
  return {
    created_at: now,
    display_name: definition.displayName,
    incident_id: incidentId,
    layout_json: definition.layoutJson,
    owner_user_id: userId,
    query_json: definition.queryJson,
    saved_view_id: savedViewId,
    saved_view_version: 1,
    scope: "private",
    updated_at: now,
    view_schema_id: timelineViewSchemaId,
  };
}

function response(data: unknown, status = 200): Response {
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
    meta: {
      request_id: "req-saved-view",
      ...(paging === undefined ? {} : { paging }),
    },
  };
}

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

describe("Workbook saved-view adapter", () => {
  it("projects explicit paging and returns only correlated saved-view resources", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        response(
          envelope(
            { saved_views: [savedViewResource()] },
            { has_more: true, limit: 2, next_cursor: "cursor-2" },
          ),
        ),
      );
    vi.stubGlobal("fetch", fetchMock);
    const savedViews = createWorkbookSavedViewAdapter({
      apiBase: "/base",
      incidentId,
    });

    await expect(
      savedViews.listPage({
        cursorToken: "cursor-1",
        limit: 2,
        signal: new AbortController().signal,
      }),
    ).resolves.toMatchObject({
      kind: "accepted",
      value: {
        nextCursor: "cursor-2",
        savedViews: [{ saved_view_id: savedViewId, saved_view_version: 1 }],
      },
    });
    expect(fetchMock).toHaveBeenCalledWith(
      `/base/api/v1/incidents/${incidentId}/saved-views?cursor_token=cursor-1&limit=2`,
      expect.objectContaining({ method: "GET" }),
    );
  });

  it("fails closed on malformed paging, cross-incident resources, and invalid versions", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        response(
          envelope(
            { saved_views: [] },
            { has_more: true, limit: 2, next_cursor: null },
          ),
        ),
      )
      .mockResolvedValueOnce(
        response(
          envelope(
            {
              saved_views: [
                savedViewResource({ incident_id: otherIncidentId }),
              ],
            },
            { has_more: false, limit: 2, next_cursor: null },
          ),
        ),
      )
      .mockResolvedValueOnce(
        response(
          envelope(
            { saved_views: [savedViewResource({ saved_view_version: 0 })] },
            { has_more: false, limit: 2, next_cursor: null },
          ),
        ),
      );
    vi.stubGlobal("fetch", fetchMock);
    const savedViews = createWorkbookSavedViewAdapter({
      apiBase: undefined,
      incidentId,
    });

    for (let attempt = 0; attempt < 3; attempt += 1) {
      await expect(
        savedViews.listPage({
          cursorToken: null,
          limit: 2,
          signal: new AbortController().signal,
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
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(response(envelope(savedViewResource()), 201))
      .mockResolvedValueOnce(
        response(
          envelope(
            savedViewResource({
              display_name: "Updated timeline",
              saved_view_version: 2,
            }),
          ),
        ),
      )
      .mockResolvedValueOnce(
        response(envelope({ deleted: true, saved_view_id: savedViewId })),
      );
    vi.stubGlobal("fetch", fetchMock);
    const savedViews = createWorkbookSavedViewAdapter({
      apiBase: undefined,
      incidentId,
    });

    await expect(
      savedViews.create({
        definition,
        signal: new AbortController().signal,
      }),
    ).resolves.toMatchObject({
      kind: "accepted",
      value: { saved_view_id: savedViewId },
    });
    await expect(
      savedViews.patch({
        baseVersion: 1,
        definition: { ...definition, displayName: "Updated timeline" },
        savedViewId,
        scope: "private",
        signal: new AbortController().signal,
        viewSchemaId: timelineViewSchemaId,
      }),
    ).resolves.toMatchObject({
      kind: "accepted",
      value: { saved_view_id: savedViewId, saved_view_version: 2 },
    });
    await expect(
      savedViews.delete({
        savedViewId,
        scope: "private",
        signal: new AbortController().signal,
      }),
    ).resolves.toEqual({ kind: "accepted", value: undefined });

    const expectedRequests = [
      ["POST", `/api/v1/incidents/${incidentId}/saved-views`],
      ["PATCH", `/api/v1/incidents/${incidentId}/saved-views/${savedViewId}`],
      ["DELETE", `/api/v1/incidents/${incidentId}/saved-views/${savedViewId}`],
    ] as const;
    for (const [index, [method, path]] of expectedRequests.entries()) {
      const [url, init] = fetchMock.mock.calls[index] as [string, RequestInit];
      expect(url).toBe(path);
      expect(init.method).toBe(method);
      expect(new Headers(init.headers).get("X-CSRF-Token")).toBe("csrf-token");
    }
    expect(JSON.parse(String(fetchMock.mock.calls[0]?.[1]?.body))).toEqual({
      display_name: definition.displayName,
      layout_json: definition.layoutJson,
      query_json: definition.queryJson,
      scope: definition.scope,
      view_schema_id: definition.viewSchemaId,
    });
    expect(
      JSON.parse(String(fetchMock.mock.calls[1]?.[1]?.body)),
    ).toMatchObject({
      base_saved_view_version: 1,
      display_name: "Updated timeline",
      scope: "private",
    });
  });

  it("rejects mismatched mutation acknowledgements and system-view mutation locally", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        response(
          envelope(savedViewResource({ incident_id: otherIncidentId })),
          201,
        ),
      )
      .mockResolvedValueOnce(
        response(
          envelope(
            savedViewResource({
              saved_view_id: "10000000-0000-4000-8000-000000000009",
              saved_view_version: 2,
            }),
          ),
        ),
      )
      .mockResolvedValueOnce(
        response(
          envelope({
            deleted: true,
            saved_view_id: "10000000-0000-4000-8000-000000000009",
          }),
        ),
      );
    vi.stubGlobal("fetch", fetchMock);
    const savedViews = createWorkbookSavedViewAdapter({
      apiBase: undefined,
      incidentId,
    });

    await expect(
      savedViews.create({ definition, signal: new AbortController().signal }),
    ).resolves.toMatchObject({
      kind: "rejected",
      failure: { kind: "invalid_contract" },
    });
    await expect(
      savedViews.patch({
        baseVersion: 1,
        definition,
        savedViewId,
        scope: "private",
        signal: new AbortController().signal,
        viewSchemaId: timelineViewSchemaId,
      }),
    ).resolves.toMatchObject({
      kind: "rejected",
      failure: { kind: "invalid_contract" },
    });
    await expect(
      savedViews.delete({
        savedViewId,
        scope: "private",
        signal: new AbortController().signal,
      }),
    ).resolves.toMatchObject({
      kind: "rejected",
      failure: { kind: "invalid_contract" },
    });
    await expect(
      savedViews.patch({
        baseVersion: 1,
        definition,
        savedViewId,
        scope: "system",
        signal: new AbortController().signal,
        viewSchemaId: timelineViewSchemaId,
      }),
    ).resolves.toMatchObject({
      kind: "rejected",
      failure: { kind: "validation" },
    });
    await expect(
      savedViews.delete({
        savedViewId,
        scope: "system",
        signal: new AbortController().signal,
      }),
    ).resolves.toMatchObject({
      kind: "rejected",
      failure: { kind: "validation" },
    });
    expect(fetchMock).toHaveBeenCalledTimes(3);
  });
});
