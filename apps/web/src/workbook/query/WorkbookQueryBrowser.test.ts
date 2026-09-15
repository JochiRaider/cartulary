import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it, vi } from "vitest";
import { deferred } from "../../testing/fetchMockTestSupport";
import { fullWorkbookViewRow } from "../../testing/timelineWorkbookTestSupport";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import { notesViewSchemaId } from "../models/workbookSurfaceRegistry";
import { WorkbookQueryBrowser } from "./WorkbookQueryBrowser";
import type {
  WorkbookViewQueryPort,
  WorkbookViewQueryResult,
} from "./WorkbookViewQueryPort";

const contract = requireViewContract(notesViewSchemaId);
const queryState = emptyWorkbookQueryState();
const input = () => ({
  contract,
  queryState,
  signal: new AbortController().signal,
});
const row = (number: number, version = 1) =>
  fullWorkbookViewRow(
    contract,
    `00000000-0000-4000-8000-${String(number).padStart(12, "0")}`,
    version,
    { "note.title": `Note ${number}`, "note.body": "Body" },
  );

function fixture(count = 405) {
  const cursors = new Map<string, number>();
  const query = vi.fn<WorkbookViewQueryPort["query"]>(async (request) => {
    const start =
      request.cursorToken === undefined ? 0 : cursors.get(request.cursorToken);
    if (start === undefined) throw new Error("Unknown fixture cursor");
    const end = Math.min(start + 100, count);
    const cursor = end < count ? ` opaque +/${end}= ` : null;
    if (cursor) cursors.set(cursor, end);
    return {
      kind: "accepted",
      value: {
        incidentId: "incident",
        viewSchemaId: notesViewSchemaId,
        rows: Array.from({ length: end - start }, (_, index) =>
          row(start + index + 1),
        ),
        canonicalQuery: {
          filters: request.queryState.filters,
          sort: request.queryState.sort,
        },
        paging: { limit: 100, hasMore: cursor !== null, nextCursor: cursor },
        producingRequest: {
          queryState: request.queryState,
          limit: 100,
          ...(request.cursorToken === undefined
            ? {}
            : { cursorToken: request.cursorToken }),
        },
      },
    };
  });
  const browser = new WorkbookQueryBrowser({ query }, notesViewSchemaId);
  const read = async () => {
    const result = await browser.query(input());
    if (result.kind === "accepted") browser.accept(result.value);
  };
  return { browser, query, read };
}

describe("Workbook query browsing", () => {
  it("coalesces live invalidations and stages only the current three-page window", async () => {
    const { browser, query, read } = fixture();
    await read();
    await browser.activate("more", read);
    await browser.activate("more", read);
    const original = browser.getSnapshot().accepted;
    const implementation = query.getMockImplementation();
    if (!implementation) throw new Error("Missing fixture");
    const pending = deferred<WorkbookViewQueryResult>();
    query.mockImplementationOnce(() => pending.promise);
    const first = browser.reconcile(read);
    await Promise.resolve();
    const queued = browser.reconcile(read);
    for (let count = 0; count < 20; count++) browser.reconcile(read);
    expect(query).toHaveBeenCalledTimes(4);
    expect(browser.getSnapshot().accepted).toBe(original);
    const pendingRead = query.mock.calls[3]?.[0];
    if (!pendingRead) throw new Error("Missing pending read");
    pending.resolve(await implementation(pendingRead));
    await Promise.all([first, queued]);
    expect(query).toHaveBeenCalledTimes(9);
    expect(browser.getSnapshot().pageCount).toBe(3);
    expect(
      query.mock.calls.slice(3).map(([request]) => request.cursorToken),
    ).toEqual([
      undefined,
      " opaque +/100= ",
      " opaque +/200= ",
      undefined,
      " opaque +/100= ",
      " opaque +/200= ",
    ]);
    expect(browser.getSnapshot().accepted?.rows).toHaveLength(300);
  });

  it("admits overlap once and bounds freshness recovery without replacing newer committed evidence", async () => {
    const { browser, query, read } = fixture();
    await read();
    const implementation = query.getMockImplementation();
    if (!implementation) throw new Error("Missing fixture");
    query.mockImplementationOnce(async (request) => {
      const result = await implementation(request);
      if (result.kind !== "accepted") return result;
      return {
        kind: "accepted",
        value: {
          ...result.value,
          rows: [row(100, 2), ...result.value.rows.slice(1)],
        },
      };
    });
    await browser.activate("more", read);
    expect(browser.getSnapshot().accepted?.rows).toHaveLength(199);
    expect(
      browser
        .getSnapshot()
        .accepted?.rows.filter((item) => item.record_id === row(100).record_id),
    ).toEqual([row(100, 2)]);
    browser.observeRows([row(100, 5)]);
    await browser.activate("restart", read);
    expect(query).toHaveBeenCalledTimes(5);
    expect(browser.getSnapshot().failure?.kind).toBe("stale_target");
    expect(
      browser
        .getSnapshot()
        .accepted?.rows.find((item) => item.record_id === row(100).record_id)
        ?.row_version,
    ).toBe(5);
    expect(browser.recoveryAttemptsUsed()).toBe(2);
  });

  it("rejects a producing-request mismatch and follows authoritative continuation on short and empty pages", async () => {
    const { browser, query, read } = fixture();
    const implementation = query.getMockImplementation();
    if (!implementation) throw new Error("Missing fixture");
    query.mockImplementationOnce(async (request) => {
      const result = await implementation(request);
      if (result.kind !== "accepted") return result;
      return {
        kind: "accepted",
        value: {
          ...result.value,
          producingRequest: {
            ...result.value.producingRequest,
            cursorToken: "wrong",
          },
        },
      };
    });
    await read();
    expect(browser.getSnapshot().accepted).toBeNull();
    expect(browser.getSnapshot().failure?.kind).toBe("invalid_contract");
    query.mockImplementationOnce(async (request) => {
      const result = await implementation(request);
      return result.kind === "accepted"
        ? { kind: "accepted", value: { ...result.value, rows: [] } }
        : result;
    });
    await read();
    expect(browser.getSnapshot().accepted?.rows).toHaveLength(0);
    expect(browser.getSnapshot().accepted?.paging.hasMore).toBe(true);
    await browser.activate("more", read);
    expect(query.mock.calls.at(-1)?.[0].cursorToken).toBe(" opaque +/100= ");
    expect(browser.getSnapshot().accepted?.rows).toHaveLength(100);
  });
  it("admits complete windows atomically and traverses beyond 300 rows with bounded return history", async () => {
    const { browser, query, read } = fixture(2505);
    const staged = await browser.query(input());
    expect(browser.getSnapshot().accepted).toBeNull();
    if (staged.kind !== "accepted") throw new Error("Expected page");
    expect(browser.accept(staged.value)).toBe(true);
    for (let index = 0; index < 25; index++)
      await browser.activate("more", read);
    expect(query).toHaveBeenCalledTimes(26);
    expect(browser.getSnapshot()).toMatchObject({
      pageCount: 3,
      earlierEvicted: true,
      hasEarlier: true,
      accepted: { paging: { hasMore: false } },
    });
    expect(browser.getSnapshot().accepted?.rows).toHaveLength(205);
    expect(browser.getSnapshot().accepted?.rows.at(-1)?.record_id).toBe(
      row(2505).record_id,
    );
    for (let index = 0; index < 20; index++)
      await browser.activate("earlier", read);
    expect(browser.getSnapshot().hasEarlier).toBe(false);
    expect(query).toHaveBeenCalledTimes(46);
    await browser.activate("earlier", read);
    expect(query).toHaveBeenCalledTimes(46);
    await browser.activate("restart", read);
    expect(query.mock.calls.at(-1)?.[0].cursorToken).toBeUndefined();
    expect(browser.getSnapshot().earlierEvicted).toBe(false);
  });

  it("guards duplicate activation and retries the captured continuation without losing accepted rows", async () => {
    const { browser, query, read } = fixture();
    await read();
    const accepted = browser.getSnapshot().accepted;
    const pending = deferred<WorkbookViewQueryResult>();
    query.mockImplementationOnce(() => pending.promise);
    const first = browser.activate("more", read);
    const duplicate = browser.activate("more", read);
    expect(query).toHaveBeenCalledTimes(2);
    pending.resolve({
      kind: "rejected",
      failure: { kind: "retryable", message: "Offline" },
    });
    await Promise.all([first, duplicate]);
    expect(browser.getSnapshot().accepted).toBe(accepted);
    await browser.activate("retry", read);
    expect(query.mock.calls[2]?.[0].cursorToken).toBe(" opaque +/100= ");
    expect(query.mock.calls[2]?.[0].queryState.sort).toEqual([]);
    expect(browser.getSnapshot().accepted?.rows).toHaveLength(200);
  });

  it("discards invalid cursor chains once and retains failed restart recovery without the old cursor", async () => {
    for (const publicReason of [
      "invalid_cursor_token",
      "cursor_query_mismatch",
      "cursor_snapshot_unavailable",
      "invalid_limit",
    ] as const) {
      const { browser, query, read } = fixture();
      await read();
      query.mockResolvedValueOnce({
        kind: "rejected",
        failure: {
          kind: "validation",
          publicCode: "invalid_view_query",
          publicReason,
          message: "Invalid cursor",
        },
      });
      query.mockResolvedValueOnce({
        kind: "rejected",
        failure: { kind: "retryable", message: "Offline" },
      });
      await browser.activate("more", read);
      expect(query).toHaveBeenCalledTimes(3);
      expect(query.mock.calls[2]?.[0].cursorToken).toBeUndefined();
      expect(browser.getSnapshot().canLoadMore).toBe(false);
      await browser.activate("more", read);
      expect(query).toHaveBeenCalledTimes(3);
      await browser.activate("retry", read);
      expect(query.mock.calls[3]?.[0].cursorToken).toBeUndefined();
      expect(browser.getSnapshot().accepted?.rows).toHaveLength(100);
    }
  });

  it("fences superseded results and errors and clears protected cursors on authority invalidation", async () => {
    const { browser, query, read } = fixture();
    await read();
    const pending = deferred<WorkbookViewQueryResult>();
    query.mockImplementationOnce(() => pending.promise);
    const old = browser.query(input());
    const current = browser.query(input());
    const accepted = await current;
    if (accepted.kind !== "accepted") throw new Error("Expected page");
    browser.accept(accepted.value);
    pending.resolve({
      kind: "rejected",
      failure: { kind: "authorization_lost", message: "Old denial" },
    });
    expect(await old).toEqual({ kind: "aborted" });
    expect(browser.getSnapshot().failure).toBeNull();
    browser.invalidate();
    expect(browser.getSnapshot()).toMatchObject({
      accepted: null,
      hasEarlier: false,
      pending: null,
    });
    expect(browser.accept(accepted.value)).toBe(false);
  });

  it("re-fetches a detached position and keeps canonical presentation separate from unapplied intent", async () => {
    const { browser, query, read } = fixture();
    await read();
    for (let index = 0; index < 3; index++)
      await browser.activate("more", read);
    browser.detach(row(250).record_id);
    expect(browser.getSnapshot().accepted).toBeNull();
    await read();
    expect(query.mock.calls.at(-1)?.[0].cursorToken).toBe(" opaque +/200= ");
    const replacement = {
      ...queryState,
      filters: [
        {
          fieldKey: "note.full_text",
          op: "full_text" as const,
          arg: { query: "Beta ALPHA" },
        },
      ],
    };
    query.mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "retryable", message: "Offline" },
    });
    await browser.query({ ...input(), queryState: replacement });
    expect(browser.presentationQuery(replacement)).toEqual(queryState);
    expect(browser.hasUnapplied(replacement)).toBe(true);
  });

  it("retains an accepted empty result after refresh failure and trusts short-page continuation", async () => {
    const { browser, query, read } = fixture(0);
    await read();
    const accepted = browser.getSnapshot().accepted;
    expect(accepted?.rows).toEqual([]);
    query.mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "retryable", message: "Offline" },
    });
    await browser.activate("restart", read);
    expect(browser.getSnapshot().accepted).toBe(accepted);
    const short = fixture(101);
    await short.read();
    await short.browser.activate("more", short.read);
    expect(short.browser.getSnapshot().accepted?.rows).toHaveLength(101);
    expect(short.browser.getSnapshot().accepted?.paging.hasMore).toBe(false);
  });
});
