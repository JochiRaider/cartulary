import { afterEach, expect, it, vi } from "vitest";
import { deferred } from "../../../testing/fetchMockTestSupport";
import { timelineCaptureReview } from "../../../testing/timelineCaptureActionTestSupport";
import {
  errorEnvelope,
  successEnvelope,
  timelineRow,
  timelineRowsEnvelope,
} from "../../../testing/timelineWorkbookTestSupport";
import { createTimelineCandidateReader } from "./createTimelineCandidateReader";
import { createTimelineRecordActionAdapter } from "./createTimelineRecordActionAdapter";

afterEach(() => vi.unstubAllGlobals());
it("Timeline candidates use independent bounded query pages and reject inconsistent identity cursors and malformed rows", async () => {
  const review = timelineCaptureReview();
  const reader = createTimelineCandidateReader({
    apiBase: "/base",
    incidentId: review.authority.incidentId,
  });
  const row = timelineRow({
    recordId: review.target.recordId,
    rowVersion: 4,
    captureState: "enriched",
    summary: "Visible candidate",
  });
  const envelope = await timelineRowsEnvelope([row]).json();
  const fetchMock = vi.fn(
    async (_input: RequestInfo | URL, _init?: RequestInit) =>
      new Response(JSON.stringify(envelope), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
  );
  vi.stubGlobal("fetch", fetchMock);
  await expect(
    reader.page("earlier-page", new AbortController().signal),
  ).resolves.toMatchObject({
    kind: "accepted",
    value: {
      rows: [{ recordId: row.record_id }],
      hasMore: false,
      nextCursor: null,
    },
  });
  expect(fetchMock.mock.calls[0]?.[0]).toBe(
    `/base/api/v1/incidents/${review.authority.incidentId}/views/cartulary.view.timeline.v2/query`,
  );
  expect(JSON.parse(String(fetchMock.mock.calls[0]?.[1]?.body))).toEqual({
    limit: 100,
    cursor_token: "earlier-page",
  });
  for (const invalid of [
    {
      ...envelope,
      data: {
        ...envelope.data,
        incident_id: "10000000-0000-4000-8000-000000000002",
      },
    },
    {
      ...envelope,
      data: { ...envelope.data, rows: [{ ...row, row_version: 0 }] },
    },
    {
      ...envelope,
      data: { ...envelope.data, rows: Array.from({ length: 101 }, () => row) },
    },
    {
      ...envelope,
      meta: {
        ...envelope.meta,
        paging: { limit: 100, has_more: true, next_cursor: null },
      },
    },
    {
      ...envelope,
      meta: {
        ...envelope.meta,
        paging: { limit: 100, has_more: true, next_cursor: "earlier-page" },
      },
    },
  ]) {
    fetchMock.mockImplementation(
      async () =>
        new Response(JSON.stringify(invalid), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
    );
    await expect(
      reader.page("earlier-page", new AbortController().signal),
    ).resolves.toMatchObject({ kind: "rejected" });
  }
  fetchMock.mockImplementation(async () =>
    errorEnvelope("authorization_denied", 403),
  );
  await expect(
    reader.page(null, new AbortController().signal),
  ).resolves.toMatchObject({
    kind: "rejected",
    failure: { publicCode: "authorization_denied" },
  });
  const cancelled = new AbortController();
  cancelled.abort();
  await expect(reader.page(null, cancelled.signal)).resolves.toEqual({
    kind: "aborted",
  });
});
it("Timeline transport validates operation specific receipt identity version reason and replacement", async () => {
  const review = timelineCaptureReview(),
    adapter = createTimelineRecordActionAdapter({ apiBase: "/base" });
  const data = {
    record_id: review.target.recordId,
    incident_id: review.authority.incidentId,
    row_version: 5,
    capture_state: "reviewed",
    change_set_id: "30000000-0000-4000-8000-000000000001",
    reason: null,
    replacement_record_id: null,
  };
  for (const change of [
    { record_id: "20000000-0000-4000-8000-000000000002" },
    { incident_id: "10000000-0000-4000-8000-000000000002" },
    { row_version: 4 },
    { row_version: 1.5 },
    { capture_state: "superseded" },
    { reason: "Substituted reason" },
    { replacement_record_id: review.target.recordId },
    { change_set_id: "" },
  ]) {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => successEnvelope({ ...data, ...change })),
    );
    await expect(
      adapter.send(
        adapter.capture(review, "receipt-test"),
        new AbortController().signal,
      ),
    ).resolves.toEqual({ kind: "uncertain" });
  }
  vi.stubGlobal(
    "fetch",
    vi.fn(async () =>
      successEnvelope({
        target_record_id: review.target.recordId,
        target_row_version: 5,
        superseding_record_id: "20000000-0000-4000-8000-000000000002",
        superseding_row_version: 2,
        incident_id: review.authority.incidentId,
        change_set_id: data.change_set_id,
        reason: "Decision variant",
      }),
    ),
  );
  await expect(
    adapter.send(
      adapter.capture(review, "receipt-test"),
      new AbortController().signal,
    ),
  ).resolves.toEqual({ kind: "uncertain" });
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => successEnvelope(data)),
  );
  await expect(
    adapter.send(
      adapter.capture(review, "receipt-test"),
      new AbortController().signal,
    ),
  ).resolves.toEqual({
    kind: "acknowledged",
    receipt: { operation: "mark-reviewed", data },
  });
  const dispatchedScope = {
    actorId: review.authority.actorId,
    sessionIdentity: review.authority.sessionIdentity,
    incidentId: review.authority.incidentId,
    epoch: 1,
  };
  let scope = dispatchedScope;
  const response = deferred<Response>();
  vi.stubGlobal(
    "fetch",
    vi.fn(() => response.promise),
  );
  const scopedAdapter = createTimelineRecordActionAdapter({
    apiBase: "/base",
    readScope: () => scope,
  });
  const pending = scopedAdapter.send(
    scopedAdapter.capture(review, "delayed-receipt"),
    new AbortController().signal,
  );
  scope = { ...scope, sessionIdentity: "replacement-session", epoch: 2 };
  response.resolve(successEnvelope(data));
  await expect(pending).resolves.toMatchObject({
    kind: "acknowledged",
    receipt: {
      observation: {
        recordId: data.record_id,
        rowVersion: data.row_version,
        scope: dispatchedScope,
      },
    },
  });
});
it("Timeline transport distinguishes definitive rejection from exceptions ambiguous status and malformed success", async () => {
  const adapter = createTimelineRecordActionAdapter({ apiBase: undefined }),
    attempt = adapter.capture(timelineCaptureReview(), "uncertain-test");
  for (const response of [
    async () => {
      throw new TypeError("Connection lost");
    },
    async () => new Response("not json", { status: 200 }),
    async () => successEnvelope({ unrelated: true }),
    async () => errorEnvelope("internal_error", 500),
  ]) {
    vi.stubGlobal("fetch", vi.fn(response));
    await expect(
      adapter.send(attempt, new AbortController().signal),
    ).resolves.toEqual({ kind: "uncertain" });
  }
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => errorEnvelope("incident_closed", 409)),
  );
  await expect(
    adapter.send(attempt, new AbortController().signal),
  ).resolves.toMatchObject({
    kind: "rejected",
    failure: { publicCode: "incident_closed" },
  });
});
