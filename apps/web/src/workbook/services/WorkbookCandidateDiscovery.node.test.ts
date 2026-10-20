import { afterEach, describe, expect, it, vi } from "vitest";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import type {
  WorkbookCandidate,
  WorkbookCandidateReader,
} from "../ports/WorkbookCandidateReadPort";
import { WorkbookCandidateDiscovery } from "./WorkbookCandidateDiscovery";

afterEach(() => vi.useRealTimers());
const page = (number: number, next: string | null = `page-${number + 1}`) => ({
  kind: "accepted" as const,
  value: {
    candidates: Array.from({ length: 100 }, (_, index) => ({
      recordId: `${number}-${index}`,
      displayText: `Candidate ${number}-${index}`,
    })),
    hasMore: next !== null,
    nextCursor: next,
  },
});
function setup(
  read: WorkbookCandidateReader<WorkbookCandidate>,
  isCurrent = () => true,
) {
  const onAuthorityFailure = vi.fn();
  const owner = new WorkbookCandidateDiscovery({
    scope: "actor/session/incident/target/revision",
    queryState: emptyWorkbookQueryState(),
    read,
    isCurrent,
    onAuthorityFailure,
  });
  return { owner, onAuthorityFailure };
}
describe("authoring candidate discovery", () => {
  it("retains one page and ten earlier requests through long traversal and re-fetches previous checkpoints", async () => {
    const read = vi.fn(async ({ cursor }) =>
      page(cursor ? Number(cursor.slice(5)) : 1),
    );
    const { owner } = setup(read);
    await owner.start();
    for (let number = 2; number <= 16; number++) {
      await owner.next();
      expect(owner.getSnapshot().page?.candidates).toHaveLength(100);
      expect(owner.getSnapshot().previousCount).toBe(Math.min(10, number - 1));
      expect(owner.getSnapshot().page?.candidates[0]?.recordId).toBe(
        `${number}-0`,
      );
    }
    for (let number = 15; number >= 6; number--) {
      await owner.previous();
      expect(owner.getSnapshot().pageNumber).toBe(number);
      expect(read.mock.calls.at(-1)?.[0].cursor).toBe(`page-${number}`);
    }
    await owner.previous();
    expect(owner.getSnapshot().pageNumber).toBe(6);
    await owner.first();
    expect(owner.getSnapshot()).toMatchObject({
      pageNumber: 1,
      previousCount: 0,
    });
    expect(read.mock.calls.at(-1)?.[0].cursor).toBeNull();
  });
  it("preserves accepted observations during pending and failed continuation and retries the captured read exactly", async () => {
    let finish!: (
      value: Awaited<ReturnType<WorkbookCandidateReader<WorkbookCandidate>>>,
    ) => void;
    const read = vi
      .fn<WorkbookCandidateReader<WorkbookCandidate>>()
      .mockResolvedValueOnce(page(1))
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            finish = resolve;
          }),
      )
      .mockResolvedValueOnce(page(2));
    const { owner } = setup(read);
    await owner.start();
    const next = owner.next();
    await owner.next();
    expect(read).toHaveBeenCalledTimes(2);
    expect(owner.getSnapshot()).toMatchObject({
      pending: true,
      pageNumber: 1,
      previousCount: 0,
    });
    expect(owner.getSnapshot().page?.candidates[0]?.recordId).toBe("1-0");
    finish({
      kind: "rejected",
      failure: { kind: "retryable", message: "Later read failed" },
    });
    await next;
    expect(owner.getSnapshot()).toMatchObject({
      pending: false,
      pageNumber: 1,
      previousCount: 0,
      failure: { kind: "retryable" },
    });
    await owner.retry();
    expect(read.mock.calls[2]?.[0]).toMatchObject({
      cursor: "page-2",
      queryState: read.mock.calls[1]?.[0].queryState,
    });
    expect(owner.getSnapshot()).toMatchObject({
      pageNumber: 2,
      previousCount: 1,
    });
    expect(owner.getSnapshot().page?.candidates).toHaveLength(100);
  });
  it("fences obsolete success failure and authority callbacks independently of transport cancellation", async () => {
    for (const outcome of [
      page(1),
      {
        kind: "rejected" as const,
        failure: {
          kind: "authentication_required" as const,
          message: "Expired",
        },
      },
    ]) {
      let active = true;
      let finish!: (value: typeof outcome) => void;
      const { owner, onAuthorityFailure } = setup(
        () =>
          new Promise((resolve) => {
            finish = resolve;
          }),
        () => active,
      );
      const request = owner.start();
      active = false;
      finish(outcome);
      await request;
      expect(owner.getSnapshot().page).toBeNull();
      expect(onAuthorityFailure).not.toHaveBeenCalled();
      owner.dispose();
    }
  });
  it("preserves typed target incident and session outcomes and delegates authority decisions", async () => {
    for (const kind of [
      "retryable",
      "stale_target",
      "authorization_lost",
      "authentication_required",
    ] as const) {
      const failure: WorkbookOperationFailure = {
        kind,
        message: "Read failed",
      };
      const read = vi
        .fn<WorkbookCandidateReader<WorkbookCandidate>>()
        .mockResolvedValueOnce(page(1))
        .mockImplementationOnce(async (input) => {
          if (
            ["authorization_lost", "authentication_required"].includes(kind)
          ) {
            input.onAuthorityFailure?.(failure);
            input.onAuthorityFailure?.(failure);
          }
          return { kind: "rejected", failure };
        });
      const { owner, onAuthorityFailure } = setup(read);
      await owner.start();
      await owner.next();
      expect(owner.getSnapshot().failure).toBe(failure);
      expect(owner.getSnapshot().page !== null).toBe(kind === "retryable");
      expect(owner.getSnapshot().concealed).toBe(
        ["authorization_lost", "authentication_required"].includes(kind),
      );
      expect(onAuthorityFailure).toHaveBeenCalledTimes(
        ["authorization_lost", "authentication_required"].includes(kind)
          ? 1
          : 0,
      );
    }
  });
  it("retains the page after cursor rejection and restarts explicitly without a cursor", async () => {
    const read = vi
      .fn<WorkbookCandidateReader<WorkbookCandidate>>()
      .mockResolvedValueOnce(page(1))
      .mockResolvedValueOnce({
        kind: "rejected",
        failure: {
          kind: "validation",
          publicCode: "invalid_view_query",
          publicReason: "cursor_query_mismatch",
          message: "Restart required",
        },
      })
      .mockResolvedValueOnce(page(1, null));
    const { owner } = setup(read);
    await owner.start();
    await owner.next();
    expect(owner.getSnapshot().pageNumber).toBe(1);
    expect(owner.getSnapshot().failure?.publicReason).toBe(
      "cursor_query_mismatch",
    );
    await owner.first();
    expect(read.mock.calls.at(-1)?.[0].cursor).toBeNull();
    expect(owner.getSnapshot()).toMatchObject({
      failure: null,
      previousCount: 0,
    });
  });
  it("rejects oversized duplicate and repeated-cursor pages without replacing the accepted page", async () => {
    const oversized = page(2);
    const duplicate = page(2);
    for (const invalid of [
      {
        ...oversized,
        value: {
          ...oversized.value,
          candidates: [
            ...oversized.value.candidates,
            { recordId: "extra", displayText: "Extra" },
          ],
        },
      },
      {
        ...duplicate,
        value: {
          ...duplicate.value,
          candidates: [
            { recordId: "duplicate", displayText: "Duplicate" },
            { recordId: "duplicate", displayText: "Duplicate" },
          ],
        },
      },
      page(2, "page-2"),
    ]) {
      const { owner } = setup(
        vi
          .fn<WorkbookCandidateReader<WorkbookCandidate>>()
          .mockResolvedValueOnce(page(1))
          .mockResolvedValueOnce(invalid),
      );
      await owner.start();
      await owner.next();
      expect(owner.getSnapshot()).toMatchObject({
        pageNumber: 1,
        previousCount: 0,
        failure: { kind: "invalid_contract" },
      });
    }
  });
  it("bounds deadlines and cancels detached reads without late publication", async () => {
    vi.useFakeTimers();
    let signal!: AbortSignal;
    let lateAuthority: (() => void) | undefined;
    const { owner, onAuthorityFailure } = setup((input) => {
      signal = input.signal;
      lateAuthority = () =>
        input.onAuthorityFailure?.({
          kind: "authentication_required",
          message: "Expired",
        });
      return new Promise(() => {});
    });
    const pending = owner.start();
    await vi.advanceTimersByTimeAsync(30_000);
    await pending;
    expect(signal.aborted).toBe(true);
    expect(owner.getSnapshot().failure?.kind).toBe("retryable");
    lateAuthority?.();
    expect(onAuthorityFailure).not.toHaveBeenCalled();
    const retry = owner.retry();
    owner.dispose();
    await retry;
    expect(signal.aborted).toBe(true);
    expect(owner.getSnapshot()).toMatchObject({
      page: null,
      pending: false,
      previousCount: 0,
    });
    expect(vi.getTimerCount()).toBe(0);
  });
});
