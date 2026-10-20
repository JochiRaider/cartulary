import { afterEach, expect, it, vi } from "vitest";
import {
  errorResponse,
  jsonResponse,
} from "../../testing/fetchMockTestSupport";
import {
  lifecycleAuthority,
  lifecycleDraft,
  lifecycleReceipt,
} from "../../testing/indicatorLifecycleTestSupport";
import { validateLifecycleDraft } from "../features/indicators/indicatorLifecycleModel";
import { createIndicatorLifecycleAdapter } from "./createIndicatorLifecycleAdapter";

const port = () =>
  createIndicatorLifecycleAdapter({
    apiBase: "https://original.test",
    incidentId: lifecycleAuthority.incidentId,
  });
const attempt = () => {
  const draft = lifecycleDraft(),
    v = validateLifecycleDraft(draft.values);
  if (!v.values) throw new Error("invalid fixture");
  return port().capture(
    lifecycleAuthority,
    1,
    draft,
    v.values,
    "secure-original",
  );
};
afterEach(() => vi.unstubAllGlobals());
it("Lifecycle transport validates complete operation receipts and treats malformed success as uncertain", async () => {
  const a = attempt(),
    valid = lifecycleReceipt();
  const invalid = [
    { ...valid, change_set_id: "" },
    { ...valid, replayed: true },
    { ...valid, affected_records: [] },
    {
      ...valid,
      affected_records: [...valid.affected_records, ...valid.affected_records],
    },
    ...[
      { incident_id: lifecycleAuthority.actorId },
      { indicator_record_id: lifecycleAuthority.actorId },
      { created_by_user_id: lifecycleAuthority.incidentId },
      { confidence: null },
      { valid_to: valid.interval.valid_from },
      { rationale: "Different" },
      { assessor: undefined },
      { support_refs: [lifecycleAuthority.actorId] },
      { created_at: "2026-09-11T14:00:00.000Z" },
      { row_version: 0 },
    ].map((patch) => ({ ...valid, interval: { ...valid.interval, ...patch } })),
  ];
  for (const data of [valid, ...invalid]) {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        jsonResponse({ data, meta: { request_id: "receipt" } }, 201),
      ),
    );
    expect((await port().send(a, new AbortController().signal)).kind).toBe(
      data === valid ? "acknowledged" : "uncertain",
    );
  }
  for (const response of [
    new Response("broken", { status: 201 }),
    new Response("broken", { status: 409 }),
    errorResponse("internal_error", 500),
  ]) {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => response),
    );
    expect((await port().send(a, new AbortController().signal)).kind).toBe(
      "uncertain",
    );
  }
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => errorResponse("row_version_conflict", 409)),
  );
  expect((await port().send(a, new AbortController().signal)).kind).toBe(
    "rejected",
  );
});
it("Lifecycle transport preserves original route bytes and support ordering on explicit replay", async () => {
  const a = attempt(),
    fetch = vi.fn(async () =>
      jsonResponse(
        {
          data: { ...lifecycleReceipt(), replayed: true },
          meta: { request_id: "replay" },
        },
        200,
      ),
    );
  vi.stubGlobal("fetch", fetch);
  for (let n = 0; n < 2; n++)
    expect((await port().send(a, new AbortController().signal)).kind).toBe(
      "acknowledged",
    );
  expect(fetch).toHaveBeenCalledTimes(2);
  for (const [url, init] of fetch.mock.calls as unknown as [
    string,
    RequestInit,
  ][]) {
    expect(url).toBe(a.path);
    expect(init.body).toBe(a.body);
    expect(init.credentials).toBe("include");
  }
});
