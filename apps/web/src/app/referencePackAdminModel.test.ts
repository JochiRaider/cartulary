import { describe, expect, it } from "vitest";
import type { listReferencePacks } from "../services/referencePacks";
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
  verificationResult: "",
};
describe("Reference Pack ownership", () => {
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
        code: "transport",
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
