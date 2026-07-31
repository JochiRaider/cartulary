import { requireViewContract } from "@cartulary/view-contracts";
import { afterEach, expect, it, vi } from "vitest";
import { timelineRow } from "../../../testing/timelineWorkbookTestSupport";
import { emptyWorkbookQueryState } from "../../models/workbookQuery";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { createTimelineViewQueryAdapter } from "./createTimelineViewQueryAdapter";

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

it("derives and validates Timeline query transport behind the semantic query port", async () => {
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
          query: { filters: [], sort: [] },
          request_id: "req-query",
        },
      }),
      { status: 200, headers: { "content-type": "application/json" } },
    ),
  );
  vi.stubGlobal("fetch", fetchMock);
  const query = createTimelineViewQueryAdapter({
    apiBase: "/base",
    incidentId,
    timelineContract: requireViewContract(timelineViewSchemaId),
  });

  await expect(
    query.query({
      queryState: emptyWorkbookQueryState(),
      signal: new AbortController().signal,
    }),
  ).resolves.toMatchObject({
    kind: "accepted",
    value: {
      incidentId,
      rows: [{ recordId, rowVersion: 3 }],
      viewSchemaId: timelineViewSchemaId,
    },
  });
  expect(fetchMock).toHaveBeenCalledWith(
    `/base/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`,
    expect.objectContaining({ method: "POST", body: "{}" }),
  );
});

it("fails closed on malformed Timeline query success and contains aborts", async () => {
  const incidentId = "00000000-0000-4000-8000-000000000001";
  const fetchMock = vi
    .fn()
    .mockResolvedValueOnce(
      new Response(JSON.stringify({ data: { rows: [] } }), {
        status: 200,
        headers: { "content-type": "application/json" },
      }),
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
  const query = createTimelineViewQueryAdapter({
    apiBase: undefined,
    incidentId,
    timelineContract: requireViewContract(timelineViewSchemaId),
  });

  await expect(
    query.query({
      queryState: emptyWorkbookQueryState(),
      signal: new AbortController().signal,
    }),
  ).resolves.toEqual({
    kind: "rejected",
    failure: {
      kind: "invalid_contract",
      message: "Timeline projection load failed.",
    },
  });

  const controller = new AbortController();
  const pending = query.query({
    queryState: emptyWorkbookQueryState(),
    signal: controller.signal,
  });
  controller.abort();
  await expect(pending).resolves.toEqual({ kind: "aborted" });
});
