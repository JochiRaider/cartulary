import { requireViewContract } from "@cartulary/view-contracts";
import { afterEach, expect, it, vi } from "vitest";
import {
  errorResponse,
  jsonResponse,
} from "../../testing/fetchMockTestSupport";
import { timelineRow } from "../../testing/timelineWorkbookTestSupport";
import {
  acceptedQueryMetadata,
  workbookQueryMeta,
} from "../../testing/workbookQueryTestSupport";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import { createWorkbookViewQueryAdapter } from "./createWorkbookViewQueryAdapter";

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

it("preserves route query cursors and projected pagination reasons without trusting unknown reasons", async () => {
  const incidentId = "00000000-0000-4000-8000-000000000001";
  const contract = requireViewContract(timelineViewSchemaId);
  const queryState = {
    ...emptyWorkbookQueryState(),
    groupBy: "timeline.capture_state",
  };
  const fetch = vi
    .fn()
    .mockResolvedValueOnce(
      jsonResponse({
        data: {
          incident_id: incidentId,
          view_schema_id: timelineViewSchemaId,
          rows: [],
        },
        meta: workbookQueryMeta(timelineViewSchemaId, queryState),
      }),
    )
    .mockResolvedValueOnce(
      errorResponse("invalid_view_query", 400, {
        reason_code: "cursor_query_mismatch",
      }),
    )
    .mockResolvedValueOnce(
      errorResponse("invalid_view_query", 400, {
        reason_code: "invented_reason",
      }),
    );
  vi.stubGlobal("fetch", fetch);
  const port = createWorkbookViewQueryAdapter({
    incidentId,
    apiBase: undefined,
  });
  const input = {
    contract,
    queryState,
    limit: 100,
    cursorToken: " cursor+/= ",
    signal: new AbortController().signal,
  };
  const result = await port.query(input);
  expect(result.kind).toBe("accepted");
  expect(JSON.parse(String(fetch.mock.calls[0]?.[1].body))).toEqual({
    limit: 100,
    cursor_token: " cursor+/= ",
    group_by: "timeline.capture_state",
  });
  if (result.kind === "accepted")
    expect(result.value.producingRequest.cursorToken).toBe(input.cursorToken);
  expect(await port.query(input)).toMatchObject({
    kind: "rejected",
    failure: {
      publicCode: "invalid_view_query",
      publicReason: "cursor_query_mismatch",
    },
  });
  const unknown = await port.query(input);
  expect(unknown.kind).toBe("rejected");
  if (unknown.kind === "rejected")
    expect(unknown.failure.publicReason).toBeUndefined();
});

it("derives and validates Workbook query transport behind the shared semantic port", async () => {
  const incidentId = "00000000-0000-4000-8000-000000000001";
  const recordId = "00000000-0000-4000-8000-000000000101";
  const fetchMock = vi.fn().mockResolvedValueOnce(
    new Response(
      JSON.stringify({
        data: {
          incident_id: incidentId,
          rows: [
            timelineRow({
              captureState: "rough",
              recordId,
              rowVersion: 3,
              summary: "Validated query row",
            }),
          ],
          view_schema_id: timelineViewSchemaId,
        },
        meta: {
          ...workbookQueryMeta(timelineViewSchemaId),
          paging: {
            limit: 100,
            has_more: true,
            next_cursor: " opaque+/cursor= ",
          },
          request_id: "req-query",
        },
      }),
      { status: 200, headers: { "content-type": "application/json" } },
    ),
  );
  vi.stubGlobal("fetch", fetchMock);
  const query = createWorkbookViewQueryAdapter({
    apiBase: "/base",
    incidentId,
  });

  await expect(
    query.query({
      contract: requireViewContract(timelineViewSchemaId),
      queryState: emptyWorkbookQueryState(),
      signal: new AbortController().signal,
    }),
  ).resolves.toMatchObject({
    kind: "accepted",
    value: {
      incidentId,
      rows: [{ record_id: recordId, row_version: 3 }],
      paging: { limit: 100, hasMore: true, nextCursor: " opaque+/cursor= " },
      canonicalQuery:
        acceptedQueryMetadata(timelineViewSchemaId).canonicalQuery,
      viewSchemaId: timelineViewSchemaId,
    },
  });
  expect(fetchMock).toHaveBeenCalledWith(
    `/base/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`,
    expect.objectContaining({ method: "POST", body: "{}" }),
  );
});

it("fails closed on malformed or cross-context query success and contains aborts", async () => {
  const incidentId = "00000000-0000-4000-8000-000000000001";
  const fetchMock = vi
    .fn()
    .mockResolvedValueOnce(
      new Response(JSON.stringify({ data: { rows: [] } }), {
        status: 200,
        headers: { "content-type": "application/json" },
      }),
    )
    .mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          data: {
            incident_id: "00000000-0000-4000-8000-000000000099",
            rows: [],
            view_schema_id: timelineViewSchemaId,
          },
          meta: {
            request_id: "req-cross-context",
          },
        }),
        { status: 200, headers: { "content-type": "application/json" } },
      ),
    )
    .mockImplementationOnce(
      (_input: RequestInfo | URL, init?: RequestInit) =>
        new Promise<Response>((_resolve, reject) => {
          init?.signal?.addEventListener("abort", () => {
            reject(new DOMException("Aborted", "AbortError"));
          });
        }),
    );
  vi.stubGlobal("fetch", fetchMock);
  const query = createWorkbookViewQueryAdapter({
    apiBase: undefined,
    incidentId,
  });

  await expect(
    query.query({
      contract: requireViewContract(timelineViewSchemaId),
      queryState: emptyWorkbookQueryState(),
      signal: new AbortController().signal,
    }),
  ).resolves.toEqual({
    kind: "rejected",
    failure: {
      kind: "invalid_contract",
      message: "Workbook view load failed.",
    },
  });

  await expect(
    query.query({
      contract: requireViewContract(timelineViewSchemaId),
      queryState: emptyWorkbookQueryState(),
      signal: new AbortController().signal,
    }),
  ).resolves.toEqual({
    kind: "rejected",
    failure: {
      kind: "invalid_contract",
      message: "Workbook view load failed.",
    },
  });

  const controller = new AbortController();
  const pending = query.query({
    contract: requireViewContract(timelineViewSchemaId),
    queryState: emptyWorkbookQueryState(),
    signal: controller.signal,
  });
  controller.abort();
  await expect(pending).resolves.toEqual({ kind: "aborted" });
});

it("binds accepted query rows to captured authority and rejects late old-session responses", async () => {
  const incidentId = "00000000-0000-4000-8000-000000000001";
  const recordId = "00000000-0000-4000-8000-000000000101";
  let scope = {
    actorId: "actor",
    sessionIdentity: "old-session",
    incidentId,
    epoch: 0,
  };
  const response = () =>
    new Response(
      JSON.stringify({
        data: {
          incident_id: incidentId,
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId,
              rowVersion: 3,
              captureState: "rough",
              summary: "Scoped row",
            }),
          ],
        },
        meta: {
          ...workbookQueryMeta(timelineViewSchemaId),
          request_id: "scope-query",
        },
      }),
      { status: 200, headers: { "content-type": "application/json" } },
    );
  let finish: ((value: Response) => void) | undefined;
  const fetchMock = vi
    .fn()
    .mockImplementationOnce(
      () =>
        new Promise<Response>((resolve) => {
          finish = resolve;
        }),
    )
    .mockImplementation(() => Promise.resolve(response()));
  vi.stubGlobal("fetch", fetchMock);
  const port = createWorkbookViewQueryAdapter({
    apiBase: undefined,
    incidentId,
    readScope: () => scope,
  });
  const input = {
    contract: requireViewContract(timelineViewSchemaId),
    queryState: emptyWorkbookQueryState(),
    signal: new AbortController().signal,
  };
  const unavailable = createWorkbookViewQueryAdapter({
    apiBase: undefined,
    incidentId,
    readScope: () => null,
  });
  expect(await unavailable.query(input)).toEqual({ kind: "aborted" });
  expect(fetchMock).not.toHaveBeenCalled();
  const pending = port.query(input);
  await vi.waitFor(() => expect(finish).toBeDefined());
  scope = { ...scope, sessionIdentity: "new-session", epoch: 2 };
  if (!finish) throw new Error("Query did not dispatch");
  finish(response());
  expect(await pending).toEqual({ kind: "aborted" });
  const fresh = await port.query(input);
  expect(fresh).toMatchObject({
    kind: "accepted",
    value: { rows: [{ observation: { recordId, rowVersion: 3, scope } }] },
  });
});
