import { getReferenceFieldContract } from "@cartulary/view-contracts";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  emptyWorkbookQueryState,
  type WorkbookQueryState,
} from "../models/workbookQuery";
import type { WorkbookPortResult } from "../ports/WorkbookPortResult";
import {
  type WorkbookReference,
  type WorkbookReferencePage,
  type WorkbookReferenceRequest,
  workbookReferenceKey,
} from "../ports/WorkbookReferenceReadPort";
import type { WorkbookViewQueryResult } from "../query/WorkbookViewQueryPort";
import { WorkbookReferenceSelection } from "./WorkbookReferenceSelection";
import { createWorkbookReferenceReader } from "./workbookReferenceReader";

const view = "cartulary.view.evidence.v1";
const query = emptyWorkbookQueryState();
const request: WorkbookReferenceRequest = {
  identityKind: "record",
  viewSchemaId: view,
  queryState: query,
};
function candidate(id: string, source = view): WorkbookReference {
  return {
    identity: { kind: "record", id },
    displayText: `Label ${id}`,
    viewSchemaId: source,
    presentation: "observed",
  };
}
function page(
  input: WorkbookReferenceRequest,
  ids: readonly string[],
  nextCursor: string | null = null,
): WorkbookPortResult<WorkbookReferencePage> {
  return {
    kind: "accepted",
    value: {
      candidates: ids.map((id) => candidate(id, input.viewSchemaId)),
      canonicalQuery: { filters: [], sort: [] },
      paging: { limit: 100, hasMore: nextCursor !== null, nextCursor },
      producingRequest: input,
    },
  };
}
function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((r) => {
    resolve = r;
  });
  return { promise, resolve };
}
const field = getReferenceFieldContract(
  "cartulary.view.task_requests.v1",
  "task.linked_record_ids",
);
if (!field) throw new Error("Missing reference field");

describe("Workbook reference selection", () => {
  afterEach(() => vi.useRealTimers());
  it("reaches later pages with one accepted page and ten checkpoints while retaining ordered choices", async () => {
    const read = vi.fn(async (input: WorkbookReferenceRequest) => {
      const index = Number(input.cursorToken ?? 0);
      return page(
        input,
        Array.from({ length: 100 }, (_, i) => `${index * 100 + i}`),
        `${index + 1}`,
      );
    });
    const picker = new WorkbookReferenceSelection({
      field,
      reader: { page: read },
      selected: [candidate("off-page")],
      onAuthorityFailure: vi.fn(),
    });
    await picker.replace(view);
    picker.selectPage([workbookReferenceKey(candidate("3"))]);
    for (let i = 0; i < 15; i += 1) await picker.next();
    picker.selectPage([workbookReferenceKey(candidate("1504"))]);
    expect(picker.getSnapshot().page?.candidates).toHaveLength(100);
    expect(picker.getSnapshot().previousCount).toBe(10);
    expect(
      picker.getSnapshot().selected.map((item) => item.identity.id),
    ).toEqual(["off-page", "3", "1504"]);
    expect(read).toHaveBeenCalledTimes(16);
    await picker.previous();
    expect(read.mock.lastCall?.[0].cursorToken).toBe("14");
    await picker.next();
    expect(read.mock.lastCall?.[0].cursorToken).toBe("15");
    await picker.first();
    expect(picker.getSnapshot().previousCount).toBe(0);
    expect(picker.getSnapshot().selected).toHaveLength(3);
    expect(read).toHaveBeenCalledTimes(19);
    picker.dispose();
    expect(picker.getSnapshot().page).toBeNull();
  });
  it("preserves the accepted page and typed continuation failure until an explicit read retry succeeds", async () => {
    const failure = {
      kind: "retryable",
      message: "Try again",
      publicCode: "service_unavailable",
    } as const;
    const read = vi
      .fn()
      .mockResolvedValueOnce(page(request, ["one"], "opaque"))
      .mockResolvedValueOnce({ kind: "rejected", failure })
      .mockResolvedValueOnce(page(request, ["two"]));
    const picker = new WorkbookReferenceSelection({
      field,
      reader: { page: read },
      selected: [candidate("retained")],
      onAuthorityFailure: vi.fn(),
    });
    await picker.replace(view);
    await picker.next();
    expect(picker.getSnapshot()).toMatchObject({
      failure: { phase: "continuation", detail: failure },
      pageNumber: 1,
      previousCount: 0,
    });
    expect(picker.getSnapshot().page?.candidates[0]?.identity.id).toBe("one");
    await picker.retry();
    expect(read.mock.calls[2]?.[0]).toEqual(read.mock.calls[1]?.[0]);
    expect(picker.getSnapshot().pageNumber).toBe(2);
    expect(picker.getSnapshot().selected[0]?.identity.id).toBe("retained");
    picker.dispose();
  });
  it("fences replaced queries and failed sources without clearing off-page selections", async () => {
    const late = deferred<WorkbookPortResult<WorkbookReferencePage>>();
    const read = vi
      .fn()
      .mockReturnValueOnce(late.promise)
      .mockResolvedValueOnce({
        kind: "rejected",
        failure: { kind: "retryable", message: "Failed source" },
      })
      .mockResolvedValueOnce(page(request, []));
    const picker = new WorkbookReferenceSelection({
      field,
      reader: { page: read },
      selected: [candidate("retained")],
      onAuthorityFailure: vi.fn(),
    });
    const first = picker.replace(view);
    await picker.replace("cartulary.view.indicators.v1");
    expect(read.mock.calls[0]?.[1].aborted).toBe(true);
    late.resolve(page(request, ["obsolete"]));
    await first;
    expect(picker.getSnapshot()).toMatchObject({
      page: null,
      failure: { phase: "initial" },
      source: "cartulary.view.indicators.v1",
    });
    expect(picker.getSnapshot().selected[0]?.identity.id).toBe("retained");
    await picker.retry();
    expect(picker.getSnapshot()).toMatchObject({
      failure: null,
      page: { candidates: [], paging: { hasMore: false } },
    });
    picker.dispose();
  });
  it("fences rapid applied query replacements while retaining selections and the latest canonical page", async () => {
    const old = deferred<WorkbookPortResult<WorkbookReferencePage>>();
    const read = vi.fn(
      (input: WorkbookReferenceRequest, _signal: AbortSignal) =>
        input.queryState.filters.length
          ? Promise.resolve(page(input, ["new-query"]))
          : old.promise,
    );
    const picker = new WorkbookReferenceSelection({
      field,
      reader: { page: read },
      selected: [candidate("off-page")],
      onAuthorityFailure: vi.fn(),
    });
    const first = picker.replace(view);
    const nextQuery: WorkbookQueryState = {
      ...query,
      filters: [
        { fieldKey: "evidence.title", op: "prefix", arg: { value: "applied" } },
      ],
    };
    await picker.replace(view, nextQuery);
    expect(read.mock.calls[0]?.[1].aborted).toBe(true);
    old.resolve(page(request, ["obsolete"], "obsolete-cursor"));
    await first;
    expect(
      picker.getSnapshot().page?.candidates.map((item) => item.identity.id),
    ).toEqual(["new-query"]);
    expect(picker.getSnapshot().page?.producingRequest.queryState).toEqual(
      nextQuery,
    );
    expect(
      picker.getSnapshot().selected.map((item) => item.identity.id),
    ).toEqual(["off-page"]);
    expect(picker.getSnapshot().previousCount).toBe(0);
    expect(read).toHaveBeenCalledTimes(2);
    picker.dispose();
  });
  it("keeps pending action limits and source exclusion separate from collection size and page validity", async () => {
    const read = vi.fn(async (input: WorkbookReferenceRequest) =>
      page(
        input,
        Array.from({ length: 100 }, (_, i) => String(i)),
      ),
    );
    const picker = new WorkbookReferenceSelection({
      field,
      reader: { page: read },
      sourceRecordId: "0",
      selected: [],
      onAuthorityFailure: vi.fn(),
    });
    await picker.replace(view);
    expect(picker.getSnapshot().page?.candidates).toHaveLength(99);
    picker.selectPage(Array.from({ length: 65 }, (_, i) => `record:${i + 1}`));
    expect(picker.getSnapshot().selected).toHaveLength(0);
    expect(picker.getSnapshot().selectionError).toContain("64");
    picker.selectPage(["record:1"]);
    await picker.replace("cartulary.view.assessments.v1");
    expect(picker.getSnapshot().selected[0]?.identity.id).toBe("1");
    picker.dispose();
  });
  it("conceals protected candidate presentation on authority failure and requests the existing lifecycle owner", async () => {
    const failure = {
      kind: "authentication_required",
      message: "Recover session",
    } as const;
    const onAuthorityFailure = vi.fn();
    const read = vi
      .fn()
      .mockResolvedValueOnce(page(request, ["visible"], "next"))
      .mockResolvedValueOnce({ kind: "rejected", failure });
    const picker = new WorkbookReferenceSelection({
      field,
      reader: { page: read },
      selected: [candidate("retained")],
      onAuthorityFailure,
    });
    await picker.replace(view);
    await picker.next();
    expect(onAuthorityFailure).toHaveBeenCalledWith(failure);
    expect(picker.getSnapshot()).toMatchObject({
      concealed: true,
      page: null,
      selected: [],
    });
    await picker.retry();
    expect(read).toHaveBeenCalledTimes(2);
    picker.dispose();
  });
  it("shares only in-flight reads and cancels one consumer independently with typed query metadata intact", async () => {
    const pending = deferred<WorkbookViewQueryResult>();
    const viewQuery = { query: vi.fn(() => pending.promise) };
    const reader = createWorkbookReferenceReader({
      authorityScope: "account:incident:1",
      viewQuery,
      readMembers: vi.fn(),
    });
    const first = new AbortController();
    const second = new AbortController();
    const a = reader.page(request, first.signal);
    const b = reader.page(request, second.signal);
    first.abort();
    expect(viewQuery.query).toHaveBeenCalledTimes(1);
    expect(await a).toEqual({ kind: "aborted" });
    pending.resolve({
      kind: "accepted",
      value: {
        incidentId: "incident",
        viewSchemaId: view,
        rows: [],
        canonicalQuery: { filters: [], sort: [] },
        paging: { limit: 100, hasMore: true, nextCursor: "opaque" },
        producingRequest: { queryState: query, limit: 100 },
      },
    });
    expect(await b).toMatchObject({
      kind: "accepted",
      value: {
        paging: { nextCursor: "opaque" },
        canonicalQuery: { filters: [], sort: [] },
      },
    });
    await reader.page(request, second.signal);
    expect(viewQuery.query).toHaveBeenCalledTimes(2);
    reader.dispose();
  });
  it("keys scope source query and continuation and fences disposed or abandoned reads", async () => {
    const pending = deferred<WorkbookViewQueryResult>();
    const viewQuery = { query: vi.fn((_input: unknown) => pending.promise) };
    const reader = createWorkbookReferenceReader({
      authorityScope: "one",
      viewQuery,
      readMembers: vi.fn(),
    });
    const signal = new AbortController().signal;
    const reads = [
      reader.page(request, signal),
      reader.page({ ...request, cursorToken: "next" }, signal),
      reader.page(
        { ...request, viewSchemaId: "cartulary.view.notes.v1" },
        signal,
      ),
      reader.page(
        {
          ...request,
          queryState: {
            ...query,
            filters: [
              { fieldKey: "evidence.title", op: "prefix", arg: { value: "A" } },
            ],
          },
        },
        signal,
      ),
    ];
    expect(viewQuery.query).toHaveBeenCalledTimes(4);
    reader.dispose();
    expect(await Promise.all(reads)).toEqual(
      Array.from({ length: 4 }, () => ({ kind: "aborted" })),
    );
    const cancelled = new AbortController();
    cancelled.abort();
    await reader.page(request, cancelled.signal);
    expect(viewQuery.query).toHaveBeenCalledTimes(4);
  });
  it("bounds stalled reads at thirty seconds and retains independent typed member paging", async () => {
    vi.useFakeTimers();
    const viewQuery = {
      query: vi.fn(() => new Promise<WorkbookViewQueryResult>(() => {})),
    };
    const readMembers = vi.fn(
      async (_cursorToken: string | undefined, _signal: AbortSignal) => ({
        kind: "accepted" as const,
        value: {
          members: [{ userId: "user-id", displayName: "Member" }],
          paging: { limit: 100, hasMore: true, nextCursor: "member-next" },
        },
      }),
    );
    const reader = createWorkbookReferenceReader({
      authorityScope: "one",
      viewQuery,
      readMembers,
    });
    const read = reader.page(request, new AbortController().signal);
    await vi.advanceTimersByTimeAsync(30_000);
    expect(await read).toMatchObject({
      kind: "rejected",
      failure: { kind: "retryable" },
    });
    const member = await reader.page(
      {
        identityKind: "incident_member",
        viewSchemaId: "incident_members",
        queryState: query,
        cursorToken: "member-cursor",
      },
      new AbortController().signal,
    );
    expect(member).toMatchObject({
      kind: "accepted",
      value: {
        canonicalQuery: null,
        candidates: [{ identity: { kind: "incident_member", id: "user-id" } }],
        paging: { nextCursor: "member-next" },
      },
    });
    expect(readMembers.mock.calls[0]?.[0]).toBe("member-cursor");
    expect(viewQuery.query).toHaveBeenCalledTimes(1);
    reader.dispose();
  });
});
