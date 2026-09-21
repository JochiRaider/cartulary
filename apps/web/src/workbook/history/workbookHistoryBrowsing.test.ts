import { describe, expect, it } from "vitest";
import { historyDiffFixture } from "../../testing/workbookHistoryTestSupport";
import type { RecordHistoryItem } from "../adapters/workbookHistoryResponse";
import {
  workbookRecordHistoryLoadedData,
  workbookRecordHistoryOperationId,
  workbookRecordHistoryReducer,
} from "../inspector/workbookRecordHistoryModel";
import {
  acceptHistoryPage,
  beginHistoryRead,
  initialHistoryBrowsing,
  rejectHistoryRead,
} from "./workbookHistoryBrowsing";
import type { HistoryPage } from "./workbookHistoryPage";

const scope = {
  epoch: 1,
  actorId: "actor",
  incidentId: "incident",
  sessionIdentity: "session",
};
const current = { rowVersion: 4, deleted: false };
const item = (ref: string): RecordHistoryItem => ({
  actor_user_id: "actor",
  committed_at: "2026-09-10T00:00:00Z",
  history_item_ref: ref,
  history_entry_ref: `entry-${ref}`,
  change_set_id: `change-${ref}`,
  operation: "patch",
  reversible: true,
  available_rollback_actions: ["history_entry"],
  diff_summary: historyDiffFixture(ref),
});
const page = (
  refs: string[],
  next: string | null,
  rowVersion = 4,
): HistoryPage => ({
  incident_id: "incident",
  record_id: "record",
  row_version: rowVersion,
  deleted: false,
  representation_generation: "cartulary.history.1",
  items: refs.map(item),
  paging:
    next === null
      ? { limit: 100, has_more: false, next_cursor: null }
      : { limit: 100, has_more: true, next_cursor: next },
});
const initial = () => initialHistoryBrowsing(scope, "record", "view");
function loaded(refs = ["a"], cursor: string | null = "next") {
  const state = beginHistoryRead(initial(), "initial");
  return acceptHistoryPage(
    state,
    required(state.pending),
    page(refs, cursor),
    current,
  );
}
describe("History browsing state", () => {
  it("restarts across representation generations before comparing committed items", () => {
    for (const refs of [["a"], []]) {
      const state = beginHistoryRead(loaded(), "continuation");
      const changed = acceptHistoryPage(
        state,
        required(state.pending),
        {
          ...page(refs, null),
          representation_generation: "cartulary.history.2",
        },
        current,
      );
      expect(changed.accepted).toBeNull();
      expect(changed.failure?.restart).toBe(true);
      const fresh = beginHistoryRead(changed, "refresh");
      expect(fresh.pending?.request).toEqual({});
      const accepted = acceptHistoryPage(
        fresh,
        required(fresh.pending),
        {
          ...page(["a"], null),
          representation_generation: "cartulary.history.2",
        },
        current,
      );
      expect(accepted.accepted?.data.representation_generation).toBe(
        "cartulary.history.2",
      );
      expect(accepted.accepted?.data.items).toHaveLength(1);
    }
  });
  it("bounds review payloads provenance and continuation while preserving ordinary browsing", () => {
    let review = initialHistoryBrowsing(scope, "record", "view", 3);
    let ordinary = initial();
    for (let index = 0; index < 12; index++) {
      for (const [state, update] of [
        [
          review,
          (next: typeof review) => {
            review = next;
          },
        ],
        [
          ordinary,
          (next: typeof review) => {
            ordinary = next;
          },
        ],
      ] as const) {
        const requested = beginHistoryRead(
          state,
          index === 0 ? "initial" : "continuation",
        );
        update(
          acceptHistoryPage(
            requested,
            required(requested.pending),
            page([`entry-${index}`], `cursor-${index}`),
            current,
          ),
        );
      }
      expect(review.accepted?.pages.length).toBeLessThanOrEqual(3);
      expect(review.accepted?.data.items.length).toBeLessThanOrEqual(3);
      expect(review.accepted?.provenance.size).toBeLessThanOrEqual(3);
      expect(review.cursors.length).toBeLessThanOrEqual(3);
    }
    expect(
      review.accepted?.data.items.map((entry) => entry.history_item_ref),
    ).toEqual(["entry-9", "entry-10", "entry-11"]);
    expect(ordinary.accepted?.data.items).toHaveLength(12);
    const restart = beginHistoryRead(review, "refresh");
    expect(restart.pending?.request).toEqual({});
    const refreshed = acceptHistoryPage(
      restart,
      required(restart.pending),
      page(["newest"], "fresh"),
      current,
    );
    expect(refreshed.accepted?.data.items).toHaveLength(1);
  });
  it("uses server continuation for short and empty pages and preserves server order", () => {
    let state = loaded();
    expect(state.accepted?.data.paging.has_more).toBe(true);
    state = beginHistoryRead(state, "continuation");
    state = acceptHistoryPage(
      state,
      required(state.pending),
      page([], "next-2"),
      current,
    );
    expect(state.accepted?.data.items.map((i) => i.history_item_ref)).toEqual([
      "a",
    ]);
    expect(state.accepted?.data.paging.has_more).toBe(true);
    state = beginHistoryRead(state, "continuation");
    state = acceptHistoryPage(
      state,
      required(state.pending),
      page(["b", "c"], null),
      current,
    );
    expect(state.accepted?.data.items.map((i) => i.history_item_ref)).toEqual([
      "a",
      "b",
      "c",
    ]);
    expect(state.accepted?.data.paging.has_more).toBe(false);
    expect(beginHistoryRead(state, "continuation")).toBe(state);
  });
  it("deduplicates overlap in place and accepts refreshed eligibility", () => {
    let state = beginHistoryRead(loaded(["a", "b"]), "continuation");
    const next: HistoryPage = {
      ...page(["b", "c"], null, 5),
      representation_generation: "cartulary.history.1",
      items: [
        { ...item("b"), reversible: false, available_rollback_actions: [] },
        item("c"),
      ],
    };
    state = acceptHistoryPage(state, required(state.pending), next, current);
    expect(state.accepted?.data.items.map((i) => i.history_item_ref)).toEqual([
      "a",
      "b",
      "c",
    ]);
    expect(state.accepted?.data.items[1]?.available_rollback_actions).toEqual(
      [],
    );
    expect(state.accepted?.provenance.get("b")?.request.cursorToken).toBe(
      "next",
    );
  });
  it("rejects duplicate identities changed content malformed paging and stalled cursors", () => {
    for (const next of [
      page(["b", "b"], null),
      {
        ...page(["b"], null),
        paging: { limit: 100, has_more: true, next_cursor: null },
      },
      page(["b"], "next"),
      page(["a"], "other"),
      page(["a"], null),
      {
        ...page(["a", "b"], null),
        representation_generation: "cartulary.history.1",
        items: [{ ...item("a"), operation: "rewritten" }, item("b")],
      },
      {
        ...page(["b"], null),
        paging: { limit: 99, has_more: false, next_cursor: null },
      },
    ]) {
      const state = beginHistoryRead(loaded(), "continuation");
      const result = acceptHistoryPage(
        state,
        required(state.pending),
        next as HistoryPage,
        current,
      );
      expect(result.failure?.restart).toBe(true);
      expect(result.accepted).toBe(state.accepted);
      expect(result.accepted?.data.paging.has_more).toBe(true);
    }
  });
  it("retains accepted pages on failure and retries the same page exactly once", () => {
    let state = beginHistoryRead(loaded(), "continuation");
    const pending = required(state.pending);
    expect(beginHistoryRead(state, "continuation")).toBe(state);
    state = rejectHistoryRead(state, pending, {
      kind: "retryable",
      message: "Read failed",
    });
    expect(state.accepted?.data.items).toHaveLength(1);
    expect(beginHistoryRead(state, "continuation")).toBe(state);
    state = beginHistoryRead(state, "continuation", true);
    expect(state.pending?.request).toEqual(pending.request);
    expect(
      acceptHistoryPage(state, pending, page(["late"], null), current),
    ).toBe(state);
    state = acceptHistoryPage(
      state,
      required(state.pending),
      page(["b"], null),
      current,
    );
    expect(state.accepted?.data.items).toHaveLength(2);
  });
  it("keeps failed-refresh entries but retires continuation until a replacement chain succeeds", () => {
    let state = beginHistoryRead(loaded(), "refresh");
    const request = required(state.pending);
    state = rejectHistoryRead(state, request, {
      kind: "retryable",
      message: "Refresh failed",
    });
    expect(state.accepted?.data.items).toHaveLength(1);
    expect(beginHistoryRead(state, "continuation")).toBe(state);
    state = beginHistoryRead(state, "refresh", true);
    expect(state.pending?.request).toEqual({});
    state = acceptHistoryPage(
      state,
      required(state.pending),
      page(["new"], "fresh"),
      current,
    );
    expect(state.accepted?.data.items.map((i) => i.history_item_ref)).toEqual([
      "new",
    ]);
    expect(state.accepted?.pages).toHaveLength(1);
  });
  it("requires explicit fresh-chain recovery for invalid cursors", () => {
    let state = beginHistoryRead(loaded(), "continuation");
    state = rejectHistoryRead(state, required(state.pending), {
      kind: "validation",
      publicCode: "invalid_pagination_request",
      message: "Invalid cursor",
    });
    expect(state.failure?.restart).toBe(true);
    expect(beginHistoryRead(state, "continuation", true)).toBe(state);
    state = beginHistoryRead(state, "refresh");
    expect(state.pending?.request).toEqual({});
    expect(state.accepted?.data.items).toHaveLength(1);
  });
  it("conceals protected entries on access or session loss", () => {
    for (const kind of [
      "authentication_required",
      "authorization_lost",
    ] as const) {
      const state = beginHistoryRead(loaded(), "continuation");
      const next = rejectHistoryRead(state, required(state.pending), {
        kind,
        message: "Access unavailable",
      });
      expect(next.accepted).toBeNull();
      expect(next.chainValid).toBe(false);
    }
  });
  it("retains historical items without regressing current record metadata", () => {
    let state = beginHistoryRead(loaded(), "continuation");
    state = acceptHistoryPage(
      state,
      required(state.pending),
      page(["b"], null),
      {
        rowVersion: 8,
        deleted: true,
      },
    );
    expect(state.accepted?.data).toMatchObject({
      row_version: 8,
      deleted: true,
    });
    expect(state.accepted?.data.items).toHaveLength(2);
    const browsing = loaded();
    const operationId = workbookRecordHistoryOperationId(1);
    const acknowledged = workbookRecordHistoryReducer(
      {
        phase: "submitting",
        browsing,
        subject: {
          kind: "live",
          recordId: "record",
          viewSchemaId: "view",
          rowVersion: 4,
          label: "Row",
          surfaceLabel: "Generic",
        },
        submission: {
          operationId,
          pendingAction: {
            kind: "destructive",
            operation: "delete",
            recordId: "record",
            rowVersion: 4,
          },
        },
      },
      {
        type: "operation_accepted",
        operationId,
        recordId: "record",
        rowVersion: 5,
      },
    );
    expect(workbookRecordHistoryLoadedData(acknowledged)).toMatchObject({
      row_version: 5,
      deleted: true,
    });
    expect(workbookRecordHistoryLoadedData(acknowledged)?.items).toEqual(
      required(browsing.accepted).data.items,
    );
  });
  it("fences obsolete refresh and retargeted responses", () => {
    const state = beginHistoryRead(loaded(), "continuation");
    const replacement = beginHistoryRead(state, "refresh");
    expect(
      acceptHistoryPage(
        replacement,
        required(state.pending),
        page(["late"], null),
        current,
      ),
    ).toBe(replacement);
    const retargeted = beginHistoryRead(
      initialHistoryBrowsing(scope, "other", "view"),
      "initial",
    );
    expect(
      acceptHistoryPage(
        retargeted,
        required(state.pending),
        page(["late"], null),
        current,
      ),
    ).toBe(retargeted);
  });
  it("distinguishes a genuinely empty terminal first page", () => {
    const state = loaded([], null);
    expect(state.accepted?.data.items).toEqual([]);
    expect(state.accepted?.data.paging).toMatchObject({
      has_more: false,
      next_cursor: null,
    });
    expect(state.failure).toBeNull();
  });
});

function required<T>(value: T | undefined | null): T {
  if (value === undefined || value === null)
    throw new Error("Missing history fixture value");
  return value;
}
