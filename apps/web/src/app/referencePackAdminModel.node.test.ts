import { describe, expect, it, vi } from "vitest";
import { listReferencePacks } from "../services/referencePacks";
import { deferred } from "../testing/fetchMockTestSupport";
import { referencePackFixture } from "../testing/referencePackTestSupport";
import {
  ReferencePackAdminController,
  referencePackTransportPorts,
} from "./referencePackAdminController";
import {
  catalogAccepted,
  catalogFailed,
  catalogStarted,
  initialReferencePackState,
  normalizeReferencePackQuery,
  referencePackEligible,
  referencePackProblemText,
  selectionChanged,
} from "./referencePackAdminModel";

const pack = (key: string, version = "1") =>
  referencePackFixture({ pack_key: key, pack_version: version });
const paging = { limit: 100, has_more: false, next_cursor: null } as const;
const authority = { lifetime: "session-1", actorId: "operator" };
const query = {
  active: "",
  packVersionState: "",
  search: "",
};
describe("Reference Pack ownership", () => {
  it("uses server pin and selected-key work projections for action eligibility", () => {
    const pinned = referencePackFixture({ reproducibility_pinned: true });
    expect(referencePackEligible(pinned, "remove")).toBe(false);
    expect(referencePackEligible(pinned, "activate")).toBe(true);
    for (const action of [
      "activate",
      "disable",
      "reverify",
      "remove",
    ] as const) {
      expect(
        referencePackEligible(
          referencePackFixture({ pending_work: true }),
          action,
        ),
      ).toBe(false);
    }
    expect(
      referencePackEligible(
        referencePackFixture({
          health: "failed",
          administratively_disabled: true,
        }),
        "activate",
      ),
    ).toBe(false);
  });
  it("gates catalog success and failure by the newest admitted generation", () => {
    let state = catalogStarted(initialReferencePackState(), 1, false);
    state = catalogAccepted(state, 1, [pack("accepted")], paging);
    state = catalogStarted(state, 2, false);
    state = catalogStarted(state, 3, false);
    expect(catalogAccepted(state, 2, [pack("obsolete")], paging)).toBe(state);
    expect(
      catalogFailed(state, 2, {
        kind: "transport",
        status: 0,
        code: "transport_unavailable",
      }),
    ).toBe(state);
    expect(state.catalog.rows[0]?.pack_key).toBe("accepted");
    expect(catalogAccepted(state, 3, [], paging).catalog.rows).toEqual([]);
  });
  it("coalesces exact repeated versions without changing server order or key selection", () => {
    let state = catalogStarted(initialReferencePackState(), 1, false);
    state = catalogAccepted(
      state,
      1,
      [pack("a", "10"), pack("a", "2")],
      paging,
    );
    state = selectionChanged(state, "a", true);
    state = selectionChanged(state, "a", true);
    state = catalogStarted(state, 2, true);
    state = catalogAccepted(state, 2, [pack("a", "2"), pack("b")], paging);
    expect(
      state.catalog.rows.map((p) => `${p.pack_key}@${p.pack_version}`),
    ).toEqual(["a@10", "a@2", "b@1"]);
    expect(state.selectedKeys).toEqual(["a"]);
    state = catalogAccepted(catalogStarted(state, 3, false), 3, [], paging);
    expect(state.selectedKeys).toEqual(["a"]);
  });
  it("normalizes Unicode input while retaining server search semantics", () => {
    expect(
      normalizeReferencePackQuery({
        ...query,
        search: "\u0085e\u0301 Host\u2003",
      }).search,
    ).toBe("é Host");
  });
  it("presents a declared service_unavailable catalog read as temporary unavailability and retries it", async () => {
    const fetchMock = vi.fn(() =>
      Promise.resolve(
        new Response(
          JSON.stringify({
            error: {
              code: "service_unavailable",
              details: { reason_code: "extension_publication_pending" },
              message: "service unavailable",
              request_id: "request-unavailable",
              retryable: true,
              status: 503,
            },
          }),
          { status: 503, headers: { "Content-Type": "application/json" } },
        ),
      ),
    );
    vi.stubGlobal("fetch", fetchMock);
    let unavailable: Awaited<ReturnType<typeof listReferencePacks>>;
    try {
      unavailable = await listReferencePacks({
        query,
        signal: new AbortController().signal,
      });
    } finally {
      vi.unstubAllGlobals();
    }
    expect(unavailable).toEqual({
      kind: "failed",
      problem: {
        kind: "rejected",
        status: 503,
        code: "service_unavailable",
        reason: "extension_publication_pending",
      },
    });
    if (unavailable.kind !== "failed") throw new Error("Expected failure");
    expect(referencePackProblemText(unavailable.problem)).toBe(
      "Cartulary is temporarily unavailable. Try again shortly.",
    );
    const list = vi
      .fn<typeof listReferencePacks>()
      .mockResolvedValueOnce(unavailable)
      .mockResolvedValue({
        kind: "read",
        value: {
          data: { pack_versions: [pack("available")] },
          meta: { request_id: "list", paging },
        },
      });
    const controller = new ReferencePackAdminController({
      ...referencePackTransportPorts,
      list,
      confirmAccess: async () => ({ kind: "authorized" }),
      isCurrent: () => true,
      authorizationFailed: () => {},
      transactionId: () => "txn",
    });
    controller.setAuthority(authority);
    controller.setActive(true);
    await vi.waitFor(() =>
      expect(controller.getSnapshot().catalog.problem).toMatchObject({
        code: "service_unavailable",
      }),
    );
    await controller.reload();
    expect(list).toHaveBeenCalledTimes(2);
    expect(controller.getSnapshot().catalog.problem).toBeNull();
    expect(
      controller.getSnapshot().catalog.rows.map((row) => row.pack_key),
    ).toEqual(["available"]);
  });
  it("keeps query generations independent from transport cancellation and invalidation", async () => {
    const pending = deferred<Awaited<ReturnType<typeof listReferencePacks>>>();
    const calls: string[] = [];
    const controller = new ReferencePackAdminController({
      ...referencePackTransportPorts,
      list: async (options) => {
        calls.push(options.query.search);
        if (options.query.search === "first") return pending.promise;
        return {
          kind: "read",
          value: {
            data: { pack_versions: [] },
            meta: { request_id: "list", paging },
          },
        };
      },
      confirmAccess: async () => ({ kind: "authorized" }),
      isCurrent: () => true,
      authorizationFailed: () => {},
      transactionId: () => "txn",
    });
    controller.setAuthority(authority);
    expect(calls).toEqual([]);
    controller.setActive(true);
    await new Promise((resolve) => setTimeout(resolve, 0));
    controller.setQuery({ ...query, search: "first" });
    await Promise.resolve();
    controller.invalidateCatalog();
    controller.setQuery({ ...query, search: "newest" });
    await new Promise((resolve) => setTimeout(resolve, 0));
    pending.resolve({
      kind: "read",
      value: {
        data: { pack_versions: [pack("stale")] },
        meta: { request_id: "old", paging },
      },
    });
    await new Promise((resolve) => setTimeout(resolve, 0));
    expect(controller.getSnapshot().catalog.rows).toEqual([]);
    expect(controller.getSnapshot().catalog.acceptedQuery?.search).toBe(
      "newest",
    );
    expect(calls).toEqual(["", "first", "newest"]);
    controller.setSelected("retained", true);
    controller.setActive(false);
    expect(controller.getSnapshot().selectedKeys).toEqual(["retained"]);
    controller.retire();
    expect(controller.getSnapshot().selectedKeys).toEqual([]);
    controller.dispose();
  });
});
