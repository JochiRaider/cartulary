import { requireViewContract } from "@cartulary/view-contracts";
import { afterEach, expect, it, vi } from "vitest";
import {
  decisionAuthority,
  decisionReceipt,
  decisionReview,
  decisionTargetId,
} from "../../testing/decisionSupersessionTestSupport";
import {
  errorResponse,
  jsonResponse,
} from "../../testing/fetchMockTestSupport";
import { fullWorkbookViewRow } from "../../testing/timelineWorkbookTestSupport";
import { decisionViewId } from "../features/coordination/decisionSupersessionModel";
import { createWorkbookDecisionSupersessionAdapter } from "./createWorkbookDecisionSupersessionAdapter";

const port = () =>
  createWorkbookDecisionSupersessionAdapter({
    apiBase: undefined,
    incidentId: decisionAuthority.incidentId,
  });
const signal = () => new AbortController().signal;
afterEach(() => vi.unstubAllGlobals());
it("Decision transport validates complete receipts and distinguishes malformed success from rejection", async () => {
  for (const status of ["proposed", "approved", "executed"]) {
    const p = port(),
      attempt = p.capture(decisionReview(status), "captured");
    const valid = decisionReceipt(status);
    for (const data of [
      valid,
      { ...valid, reason: "different" },
      { ...valid, superseding_row_version: 0 },
      { ...valid, target_row_version: 4 },
      { ...valid, target_status: "approved" },
      { ...valid, change_set_id: "" },
      { ...valid, target_record_id: valid.superseding_record_id },
      { ...valid, reason: undefined },
      {
        capture_state: "superseded",
        change_set_id: valid.change_set_id,
        incident_id: decisionAuthority.incidentId,
        reason: valid.reason,
        record_id: valid.target_record_id,
        replacement_record_id: valid.superseding_record_id,
        row_version: 5,
      },
    ]) {
      vi.stubGlobal(
        "fetch",
        vi.fn(async () =>
          jsonResponse({ data, meta: { request_id: "receipt" } }),
        ),
      );
      expect((await p.send(attempt, signal())).kind).toBe(
        data === valid ? "acknowledged" : "uncertain",
      );
    }
  }
  const p = port(),
    attempt = p.capture(decisionReview(), "captured");
  for (const code of [
    "row_version_conflict",
    "client_txn_conflict",
    "incident_closed",
  ]) {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => errorResponse(code, 409)),
    );
    expect((await p.send(attempt, signal())).kind).toBe("rejected");
  }
  for (const response of [
    errorResponse("internal_error", 500),
    new Response("invalid json", { status: 200 }),
    new Response("invalid error", { status: 409 }),
  ]) {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => response),
    );
    expect((await p.send(attempt, signal())).kind).toBe("uncertain");
  }
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => {
      throw new Error("connection lost");
    }),
  );
  expect((await p.send(attempt, signal())).kind).toBe("uncertain");
});
it("Decision transport replays captured bytes with current credentials and no replacement precondition", async () => {
  const p = port(),
    attempt = p.capture(decisionReview(), "one-crypto-id");
  const fetch = vi.fn(async () =>
    jsonResponse({ data: decisionReceipt(), meta: { request_id: "receipt" } }),
  );
  vi.stubGlobal("fetch", fetch);
  await p.send(attempt, signal());
  await p.send(attempt, signal());
  expect(fetch.mock.calls).toHaveLength(2);
  for (const [url, init] of fetch.mock.calls as unknown as [
    string,
    RequestInit,
  ][]) {
    expect(url).toBe(attempt.path);
    expect(init.body).toBe(attempt.body);
    expect(init.credentials).toBe("include");
    expect(JSON.parse(String(init.body))).toEqual({
      base_row_version: 4,
      client_txn_id: "one-crypto-id",
      replacement_record_id: attempt.review.replacement.recordId,
      reason: "Later evidence",
    });
  }
});
it("Decision candidates preserve opaque paging and reject missing metadata cross-context and access failures", async () => {
  const p = port();
  const { view_schema_id: _view, ...row } = fullWorkbookViewRow(
    requireViewContract(decisionViewId),
    decisionTargetId,
    4,
    { "decision.summary": "Visible Decision", "decision.status": "proposed" },
  );
  const envelope = {
    data: {
      incident_id: decisionAuthority.incidentId,
      view_schema_id: decisionViewId,
      rows: [row],
    },
    meta: {
      request_id: "candidates",
      query: { filters: [], sort: [] },
      paging: { has_more: true, next_cursor: "opaque:next", limit: 100 },
    },
  };
  const fetch = vi.fn(async () => jsonResponse(envelope));
  vi.stubGlobal("fetch", fetch);
  expect(await p.page(null, signal())).toMatchObject({
    kind: "accepted",
    value: {
      hasMore: true,
      nextCursor: "opaque:next",
      rows: [{ record_id: decisionTargetId, row_version: 4 }],
    },
  });
  expect(await p.page("opaque:previous", signal())).toMatchObject({
    kind: "accepted",
  });
  expect(
    (fetch.mock.calls as unknown as [string, RequestInit][]).map(([, init]) =>
      JSON.parse(String(init.body)),
    ),
  ).toEqual([{ limit: 100 }, { limit: 100, cursor_token: "opaque:previous" }]);
  expect(await p.page("opaque:next", signal())).toMatchObject({
    kind: "rejected",
  });
  for (const invalid of [
    { ...envelope, meta: { request_id: "missing" } },
    {
      ...envelope,
      meta: {
        ...envelope.meta,
        paging: { ...envelope.meta.paging, has_more: false },
      },
    },
    {
      ...envelope,
      data: {
        ...envelope.data,
        incident_id: "00000000-0000-4000-8000-000000000099",
      },
    },
  ]) {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => jsonResponse(invalid)),
    );
    expect(await p.page(null, signal())).toMatchObject({ kind: "rejected" });
  }
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => errorResponse("forbidden", 403)),
  );
  expect(await p.page(null, signal())).toMatchObject({ kind: "rejected" });
});
