import type { QueryEvidenceMetadataResponse } from "@cartulary/protocol-ts/http";
import { afterEach, expect, it, vi } from "vitest";
import { jsonResponse } from "../../testing/fetchMockTestSupport";
import type { WorkbookReadScope } from "../query/WorkbookQueryRow";
import { createWorkbookEvidenceMetadataAdapter } from "./createWorkbookEvidenceMetadataAdapter";

const incidentId = "00000000-0000-4000-8000-000000000001";
const first = "00000000-0000-4000-8000-000000000101";
const second = "00000000-0000-4000-8000-000000000102";
const scope: WorkbookReadScope = {
  actorId: "actor",
  sessionIdentity: "session",
  incidentId,
  epoch: 1,
};
const item = (
  id: string,
): QueryEvidenceMetadataResponse["data"]["items"][number] => ({
  record_id: id,
  row_version: 2,
  title: null,
  lifecycle_state: "available",
  has_blob: true,
  upload_state: "available",
  preview_eligibility: { eligible: true, reason_code: null },
});
const envelope = (items: unknown[], incident = incidentId) =>
  jsonResponse({
    data: { incident_id: incident, items },
    meta: { request_id: "metadata-test" },
  });
afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
  vi.useRealTimers();
});

it("bounds Evidence metadata requests and validates the exact ordered response binding", async () => {
  const fetch = vi
    .fn()
    .mockResolvedValue(envelope([item(first), item(second)]));
  vi.stubGlobal("fetch", fetch);
  const port = createWorkbookEvidenceMetadataAdapter({
    apiBase: undefined,
    incidentId,
    readScope: () => scope,
  });
  const input = {
    recordIds: [first, second],
    signal: new AbortController().signal,
  };
  expect(await port.query(input)).toEqual({
    kind: "accepted",
    items: [first, second].map((recordId) => ({
      recordId,
      rowVersion: 2,
      title: null,
      lifecycleState: "available",
      hasBlob: true,
      uploadState: "available",
      previewEligibility: { eligible: true, reasonCode: null },
    })),
  });
  expect(String(fetch.mock.calls[0]?.[0])).toContain(
    `/incidents/${incidentId}/evidence-metadata`,
  );
  expect(JSON.parse(String(fetch.mock.calls[0]?.[1].body))).toEqual({
    record_ids: [first, second],
  });
  for (const recordIds of [
    [],
    [first, first],
    ["INVALID"],
    ["00000000-0000-0000-0000-000000000000"],
    Array.from(
      { length: 101 },
      (_, i) => `00000000-0000-4000-8000-${String(i + 1).padStart(12, "0")}`,
    ),
  ]) {
    expect(await port.query({ ...input, recordIds })).toMatchObject({
      kind: "rejected",
      failure: { kind: "invalid_contract" },
    });
  }
  expect(fetch).toHaveBeenCalledTimes(1);
  for (const response of [
    envelope([item(first)]),
    envelope([item(second), item(first)]),
    envelope([item(first), item(first)]),
    envelope([item(first), item(second)], second),
    envelope([{ ...item(first), row_version: 1.5 }, item(second)]),
    envelope([{ ...item(first), physical_key: "must-not-leak" }, item(second)]),
    envelope([
      {
        ...item(first),
        preview_eligibility: { eligible: true, reason_code: "blob_missing" },
      },
      item(second),
    ]),
    envelope([{ ...item(first), has_blob: false }, item(second)]),
    envelope([
      { ...item(first), lifecycle_state: "quarantined" },
      item(second),
    ]),
  ]) {
    fetch.mockResolvedValueOnce(response);
    expect(await port.query(input)).toMatchObject({
      kind: "rejected",
      failure: { kind: "invalid_contract" },
    });
  }
});

it("discards Evidence metadata on abort or changed read authority without sending unscoped reads", async () => {
  let authority: WorkbookReadScope | null = null;
  const fetch = vi.fn();
  vi.stubGlobal("fetch", fetch);
  const port = createWorkbookEvidenceMetadataAdapter({
    apiBase: undefined,
    incidentId,
    readScope: () => authority,
  });
  const input = { recordIds: [first], signal: new AbortController().signal };
  expect(await port.query(input)).toEqual({ kind: "aborted" });
  authority = { ...scope, incidentId: second };
  expect(await port.query(input)).toEqual({ kind: "aborted" });
  expect(fetch).not.toHaveBeenCalled();
  for (const change of [
    () => {
      authority = null;
    },
    () => {
      authority = { ...scope, epoch: 2 };
    },
    () => {
      authority = { ...scope, sessionIdentity: "replacement" };
    },
    () => {
      authority = { ...scope, actorId: "other" };
    },
  ]) {
    authority = scope;
    let settle!: (response: Response) => void;
    fetch.mockImplementationOnce(
      () =>
        new Promise<Response>((resolve) => {
          settle = resolve;
        }),
    );
    const pending = port.query(input);
    change();
    settle(envelope([item(first)]));
    expect(await pending).toEqual({ kind: "aborted" });
  }
  authority = scope;
  const controller = new AbortController();
  controller.abort();
  const before = fetch.mock.calls.length;
  expect(await port.query({ ...input, signal: controller.signal })).toEqual({
    kind: "aborted",
  });
  expect(fetch).toHaveBeenCalledTimes(before);
  vi.useFakeTimers();
  fetch.mockImplementationOnce(() => new Promise<Response>(() => {}));
  const hung = port.query(input);
  await vi.advanceTimersByTimeAsync(30_000);
  expect(await hung).toMatchObject({
    kind: "rejected",
    failure: { kind: "retryable" },
  });
  expect(fetch.mock.calls.at(-1)?.[1].signal.aborted).toBe(true);
  expect(fetch).toHaveBeenCalledTimes(before + 1);
});
