import { afterEach, describe, expect, it, vi } from "vitest";
import {
  mergeIncidentId,
  mergeReceipt,
  mergeReview,
} from "../../testing/entityMergeTestSupport";
import { createWorkbookEntityMergeAdapter } from "./createWorkbookEntityMergeAdapter";

const response = (data: unknown, status = 200) =>
  new Response(JSON.stringify({ data, meta: { request_id: "merge-test" } }), {
    status,
    headers: { "content-type": "application/json" },
  });
const error = (
  code: string,
  details: Record<string, unknown> = {},
  status = 409,
) =>
  new Response(
    JSON.stringify({
      error: {
        code,
        status,
        request_id: "merge-test",
        message: "Rejected",
        retryable: false,
        details,
      },
    }),
    { status, headers: { "content-type": "application/json" } },
  );
afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("Entity merge captured transport", () => {
  it("replays identical path and bytes with current CSRF and retains every receipt field for both types", async () => {
    for (const type of ["host", "identity"] as const) {
      const receipt = mergeReceipt(type);
      const fetch = vi.fn(async () => response(receipt));
      vi.stubGlobal("fetch", fetch);
      const cookie = vi
        .spyOn(document, "cookie", "get")
        .mockReturnValue("cartulary_csrf=first");
      const port = createWorkbookEntityMergeAdapter({
        apiBase: "/proxy",
        incidentId: mergeIncidentId,
      });
      const attempt = port.capture(mergeReview(type), "fixed-transaction");
      expect(await port.send(attempt, new AbortController().signal)).toEqual({
        kind: "acknowledged",
        receipt,
      });
      cookie.mockReturnValue("cartulary_csrf=recovered");
      expect(await port.send(attempt, new AbortController().signal)).toEqual({
        kind: "acknowledged",
        receipt,
      });
      const calls = fetch.mock.calls as unknown as [string, RequestInit][];
      expect(calls.map(([path]) => path)).toEqual([attempt.path, attempt.path]);
      expect(calls.map(([, init]) => init.body)).toEqual([
        attempt.body,
        attempt.body,
      ]);
      expect((calls[1]?.[1].headers as Headers).get("X-CSRF-Token")).toBe(
        "recovered",
      );
      expect(attempt.body).toBe(
        JSON.stringify({
          loser_record_id: mergeReview(type).loser.recordId,
          survivor_base_row_version: 7,
          loser_base_row_version: 2,
          client_txn_id: "fixed-transaction",
          reason: "Merge duplicate entity",
        }),
      );
      cookie.mockRestore();
    }
  });
  it("treats thrown transport, malformed success and inconsistent identities, versions or summary as uncertain", async () => {
    for (const type of ["host", "identity"] as const) {
      const valid = mergeReceipt(type);
      const port = createWorkbookEntityMergeAdapter({
        apiBase: undefined,
        incidentId: mergeIncidentId,
      });
      const attempt = port.capture(mergeReview(type), "fixed");
      const fetch = vi.fn();
      vi.stubGlobal("fetch", fetch);
      const summaries = [
        {
          ...valid.merge_summary,
          record_type: type === "host" ? "identity" : "host",
        },
        { ...valid.merge_summary, repointed_link_count: -1 },
        {
          ...valid.merge_summary,
          repointed_tag_count: Number.MAX_SAFE_INTEGER + 1,
        },
        {
          ...valid.merge_summary,
          exact_match_classes: [
            ...valid.merge_summary.exact_match_classes,
          ].reverse(),
        },
        { ...valid.merge_summary, exact_match_classes: [] },
        {
          ...valid.merge_summary,
          exact_match_classes: valid.merge_summary.exact_match_classes.map(
            (entry) => ({ ...entry, promoted_count: 2 }),
          ),
        },
      ];
      for (const payload of [
        null,
        {},
        { ...valid, incident_id: valid.change_set_id },
        { ...valid, record_type: type === "host" ? "identity" : "host" },
        { ...valid, survivor_record_id: valid.loser_record_id },
        { ...valid, loser_record_id: valid.survivor_record_id },
        { ...valid, merged_into_record_id: valid.loser_record_id },
        { ...valid, survivor_row_version: 7 },
        { ...valid, loser_row_version: 2 },
        { ...valid, change_set_id: "invalid" },
        ...summaries.map((merge_summary) => ({ ...valid, merge_summary })),
      ]) {
        fetch.mockResolvedValueOnce(response(payload));
        expect(await port.send(attempt, new AbortController().signal)).toEqual({
          kind: "uncertain",
        });
      }
      fetch.mockRejectedValueOnce(new TypeError("transport"));
      expect(await port.send(attempt, new AbortController().signal)).toEqual({
        kind: "uncertain",
      });
      fetch.mockResolvedValueOnce(new Response("not JSON", { status: 200 }));
      expect(await port.send(attempt, new AbortController().signal)).toEqual({
        kind: "uncertain",
      });
      fetch.mockResolvedValueOnce(error("internal_error", {}, 500));
      expect(await port.send(attempt, new AbortController().signal)).toEqual({
        kind: "uncertain",
      });
    }
  });
  it("preserves permitted long collision details and classifies definitive stale, transaction and lock rejection", async () => {
    const port = createWorkbookEntityMergeAdapter({
      apiBase: undefined,
      incidentId: mergeIncidentId,
    });
    const attempt = port.capture(mergeReview(), "fixed");
    const normalized = "a".repeat(600);
    const fetch = vi.fn().mockResolvedValueOnce(
      error("merge_precondition_failed", {
        reason_code: "carry_forward_identifier_collision",
        identifier_class: "hostname",
        normalized_value: normalized,
        blocking_record_id: mergeReceipt().change_set_id,
        unapproved_extra: "not projected",
      }),
    );
    vi.stubGlobal("fetch", fetch);
    const outcome = await port.send(attempt, new AbortController().signal);
    expect(outcome).toMatchObject({
      kind: "rejected",
      failure: {
        kind: "validation",
        fields: expect.arrayContaining([
          { field: "normalized_value", message: normalized },
        ]),
      },
    });
    if (outcome.kind === "rejected" && outcome.failure.kind === "validation")
      expect(
        outcome.failure.fields?.some(
          (field) => field.field === "unapproved_extra",
        ),
      ).toBe(false);
    for (const [code, kind] of [
      ["row_version_conflict", "stale_target"],
      ["client_txn_conflict", "client_txn_conflict"],
      ["record_locked", "terminal"],
    ] as const) {
      fetch.mockResolvedValueOnce(error(code));
      expect(
        await port.send(attempt, new AbortController().signal),
      ).toMatchObject({ kind: "rejected", failure: { kind } });
    }
  });
});
