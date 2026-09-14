import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it, vi } from "vitest";
import { observeAsyncOperation } from "../../../services/asyncObservation";
import { deferred } from "../../../testing/fetchMockTestSupport";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { createOrdinaryCreateTransport } from "../../adapters/createOrdinaryCreateTransport";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import { createEntityOrdinaryCreate } from "../entities/entityOrdinaryCreate";
import { ordinaryCreateContributions } from "./ordinaryCreateContributions";
import type {
  OrdinaryCreateAttempt,
  OrdinaryCreateOutcome,
  OrdinaryCreateReceipt,
} from "./ordinaryCreateOperation";
import { WorkbookOrdinaryCreateOwner } from "./WorkbookOrdinaryCreateOwner";

const incident = "10000000-0000-4000-8000-000000000001";
const actor = "99999999-9999-4999-8999-999999999999";
const view = "cartulary.view.evidence.v1";
const authority = {
  actorId: actor,
  incidentId: incident,
  role: "admin" as const,
  closed: false,
  sessionIdentity: "same-account",
};
const accepted = (version = 1): OrdinaryCreateOutcome => ({
  kind: "accepted",
  status: 201,
  receipt: {
    data: {
      view_schema_id: view,
      change_set_id: "30000000-0000-4000-8000-000000000001",
      row: fullWorkbookViewRow(
        requireViewContract(view),
        "20000000-0000-4000-8000-000000000001",
        version,
        { "evidence.title": "Original" },
      ),
    },
    meta: { request_id: "request-create" },
  },
});
function setup(options: { observe?: typeof observeAsyncOperation } = {}) {
  let id = 0;
  const effects = {
    accepted: vi.fn(),
    refresh: vi.fn(async (_receipt: OrdinaryCreateReceipt) => {}),
  };
  const send = vi.fn(
    async (
      _attempt: OrdinaryCreateAttempt,
      _signal: AbortSignal,
    ): Promise<OrdinaryCreateOutcome> => accepted(),
  );
  const verify = vi.fn(async () => {});
  const readAuthority = vi.fn(
    async (): Promise<WorkbookMutationAuthority> => authority,
  );
  const owner = new WorkbookOrdinaryCreateOwner(
    incident,
    ordinaryCreateContributions,
    {
      ids: {
        create: () =>
          `00000000-0000-4000-8000-${String(++id).padStart(12, "0")}`,
      },
      effects,
      ...options,
    },
  );
  owner.configure(readAuthority, {
    ...createOrdinaryCreateTransport(undefined),
    send,
    verify,
  });
  owner.setAuthority(authority);
  owner.update(view, "evidence.title", "  Original  ");
  return {
    owner,
    send,
    verify,
    readAuthority,
    effects,
    schema: () => {
      const schema = owner.getSnapshot().schemas[view];
      if (!schema) throw new Error("Expected visible schema");
      return schema;
    },
    entry: () => {
      const entry = owner.getSnapshot().schemas[view]?.entries.at(-1);
      if (!entry) throw new Error("Expected retained operation");
      return entry;
    },
  };
}
describe("ordinary creation submission", () => {
  it("reserves before authority reads and captures immutable bytes shared by presentations", async () => {
    const s = setup(),
      admission = deferred<typeof authority>();
    s.readAuthority.mockReturnValueOnce(admission.promise);
    const first = s.owner.submit(view);
    const second = s.owner.submit(view);
    expect(s.readAuthority).toHaveBeenCalledTimes(1);
    expect(s.entry().attempt).toMatchObject({
      operationID: "createViewRow",
      authority,
      pathParameters: { incident_id: incident, view_schema_id: view },
      draft: { revision: 1, values: { "evidence.title": "  Original  " } },
      request: { "evidence.title": "Original" },
    });
    expect(Object.isFrozen(s.entry().attempt.request)).toBe(true);
    s.owner.update(view, "evidence.title", "  Next draft  ");
    admission.resolve(authority);
    await Promise.all([first, second]);
    expect(s.send).toHaveBeenCalledTimes(1);
    expect(s.schema().draft.values["evidence.title"]).toBe("  Next draft  ");
    expect(s.schema().draft.id).not.toBe(s.entry().attempt.draft.id);
    expect(JSON.parse(s.entry().attempt.body)["evidence.title"]).toBe(
      "Original",
    );
  });
  it("replays uncertainty unchanged while newer text and another schema remain independent", async () => {
    const s = setup();
    s.send.mockResolvedValueOnce({ kind: "uncertain" });
    await s.owner.submit(view);
    const attempt = s.entry().attempt;
    s.owner.update(view, "evidence.title", "  Next unsent  ");
    await s.owner.submit(view);
    expect(s.send).toHaveBeenCalledTimes(1);
    expect(s.owner.busy("cartulary.view.hosts.v1")).toBe(false);
    s.owner.closeIncident();
    s.readAuthority.mockResolvedValueOnce({ ...authority, closed: true });
    await s.owner.replay(attempt.clientTxnId);
    expect(s.send).toHaveBeenCalledTimes(2);
    expect(s.send.mock.calls[0]?.[0]).toBe(s.send.mock.calls[1]?.[0]);
    expect(s.verify).toHaveBeenCalledTimes(1);
    expect(s.entry().receipt).not.toBeNull();
    expect(s.schema().draft.values["evidence.title"]).toBe("  Next unsent  ");
  });
  it("retains exact rejected text and uses fresh identities only after explicit correction", async () => {
    const s = setup();
    s.send.mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "validation", message: "rejected" },
    });
    await s.owner.submit(view);
    const rejected = s.entry().attempt;
    expect(s.schema().draft.values["evidence.title"]).toBe("  Original  ");
    s.owner.update(view, "evidence.title", " Corrected ");
    await s.owner.submit(view);
    expect(s.entry().attempt.clientTxnId).not.toBe(rejected.clientTxnId);
    expect(s.entry().attempt.request).toHaveProperty(
      "evidence.title",
      "Corrected",
    );
    expect(s.schema().draft.values).toEqual({});
  });
  it("makes incident-closed rejection terminal and reopening explicit", async () => {
    const s = setup();
    s.send.mockResolvedValueOnce({ kind: "uncertain" }).mockResolvedValueOnce({
      kind: "rejected",
      failure: {
        kind: "terminal",
        publicCode: "incident_closed",
        message: "closed",
      },
    });
    await s.owner.submit(view);
    const id = s.entry().attempt.clientTxnId;
    await s.owner.replay(id);
    expect(s.entry().phase).toBe("rejected");
    expect(s.owner.canAuthor()).toBe(false);
    s.owner.setAuthority(authority);
    await s.owner.replay(id);
    expect(s.send).toHaveBeenCalledTimes(2);
    await s.owner.submit(view);
    expect(s.entry().attempt.clientTxnId).not.toBe(id);
  });
  it("requires explicit fresh-request recovery for transaction conflicts", async () => {
    const s = setup();
    s.send.mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "client_txn_conflict", message: "technical identity" },
    });
    await s.owner.submit(view);
    const id = s.entry().attempt.clientTxnId;
    await s.owner.submit(view);
    expect(s.send).toHaveBeenCalledTimes(1);
    expect(s.entry().message).not.toContain("technical identity");
    await s.owner.submitFreshAfterConflict(id);
    expect(s.send).toHaveBeenCalledTimes(2);
    expect(s.entry().attempt.clientTxnId).not.toBe(id);
  });
  it("retains detached receipts and makes failed-refresh recovery read-only", async () => {
    const s = setup(),
      result = deferred<OrdinaryCreateOutcome>();
    s.send.mockReturnValueOnce(result.promise);
    s.effects.refresh.mockRejectedValueOnce(new Error("refresh failed"));
    const detach = s.owner.attach(view, Symbol("grid"));
    const dispatched = s.owner.submit(view);
    await vi.waitFor(() => expect(s.send).toHaveBeenCalledTimes(1));
    detach();
    result.resolve(accepted());
    await dispatched;
    await vi.waitFor(() => expect(s.entry().refresh).toBe("required"));
    const receipt = s.entry().receipt;
    expect(receipt).toEqual(
      (accepted() as { receipt: OrdinaryCreateReceipt }).receipt,
    );
    s.owner.update(view, "evidence.title", "next");
    s.owner.setAuthority({ ...authority, role: "viewer" });
    s.readAuthority.mockResolvedValueOnce({ ...authority, role: "viewer" });
    await s.owner.refreshAccepted(s.entry().attempt.clientTxnId);
    await s.owner.replay(s.entry().attempt.clientTxnId);
    expect(s.entry().refresh).toBe("complete");
    expect(s.send).toHaveBeenCalledTimes(1);
    expect(s.schema().draft.values["evidence.title"]).toBe("next");
    expect(s.effects.accepted).toHaveBeenCalledTimes(1);
  });
  it("accepts late acknowledgements monotonically without replacing newer row versions", async () => {
    let timeout: (() => void) | undefined;
    const s = setup({
      observe: (request) =>
        observeAsyncOperation(request, {
          now: () => 0,
          schedule: (callback) => {
            timeout = callback;
            return () => {};
          },
        }),
    });
    const result = deferred<OrdinaryCreateOutcome>();
    s.send.mockReturnValueOnce(result.promise);
    const dispatched = s.owner.submit(view);
    await vi.waitFor(() => expect(s.send).toHaveBeenCalledTimes(1));
    timeout?.();
    await dispatched;
    expect(s.entry().phase).toBe("uncertain");
    const latest = (accepted(4) as { receipt: OrdinaryCreateReceipt }).receipt
      .data.row;
    s.owner.acceptRow(latest);
    result.resolve(accepted(1));
    await vi.waitFor(() => expect(s.entry().phase).toBe("accepted"));
    expect(s.owner.latestRow(latest.record_id)?.row_version).toBe(4);
    await s.owner.replay(s.entry().attempt.clientTxnId);
    expect(s.send).toHaveBeenCalledTimes(1);
  });
  it("conceals same-account work and fences retired account callbacks", async () => {
    for (const retire of [false, true]) {
      const s = setup(),
        result = deferred<OrdinaryCreateOutcome>();
      s.send.mockReturnValueOnce(result.promise);
      const dispatched = s.owner.submit(view);
      await vi.waitFor(() => expect(s.send).toHaveBeenCalledTimes(1));
      s.owner.suspend();
      if (retire)
        s.owner.setAuthority({ ...authority, actorId: "different-account" });
      result.resolve(accepted());
      await dispatched;
      expect(s.owner.getSnapshot()).toEqual({ authority: null, schemas: {} });
      expect(s.effects.accepted).not.toHaveBeenCalled();
      s.owner.setAuthority(authority);
      if (retire) expect(s.owner.getSnapshot().schemas).toEqual({});
      else {
        expect(s.entry().receipt).not.toBeNull();
        expect(s.effects.accepted).not.toHaveBeenCalled();
        await s.owner.refreshAccepted(s.entry().attempt.clientTxnId);
        expect(s.effects.accepted).toHaveBeenCalledTimes(1);
      }
      expect(s.send).toHaveBeenCalledTimes(1);
    }
  });
  it("suspends uncertain authority and requires revalidation before explicit recovery", async () => {
    const s = setup();
    const recheck = vi.fn();
    s.owner.configure(
      s.readAuthority,
      {
        ...createOrdinaryCreateTransport(undefined),
        send: s.send,
        verify: s.verify,
      },
      recheck,
    );
    s.readAuthority.mockRejectedValueOnce(new Error("unavailable authority"));
    await s.owner.submit(view);
    expect(s.owner.getSnapshot()).toEqual({ authority: null, schemas: {} });
    expect(s.send).not.toHaveBeenCalled();
    expect(recheck).toHaveBeenCalledTimes(1);
    s.owner.setAuthority(authority);
    expect(s.schema().draft.values["evidence.title"]).toBe("  Original  ");
    expect(s.send).not.toHaveBeenCalled();
    s.send.mockResolvedValueOnce({ kind: "uncertain" });
    await s.owner.submit(view);
    const id = s.entry().attempt.clientTxnId;
    s.owner.setAuthority({ ...authority, role: "viewer" });
    await s.owner.replay(id);
    expect(s.send).toHaveBeenCalledTimes(1);
    s.owner.setAuthority(authority);
    expect(s.send).toHaveBeenCalledTimes(1);
    await s.owner.replay(id);
    expect(s.send).toHaveBeenCalledTimes(2);
  });
  it("holds Entity admission through uncertainty and releases only on terminal settlement", async () => {
    const release = vi.fn(),
      begin = vi.fn(() => release),
      versions = vi.fn();
    const owner = new WorkbookOrdinaryCreateOwner(
      incident,
      [createEntityOrdinaryCreate({ begin, acceptVersion: versions })],
      { ids: { create: () => "entity-create-id" } },
    );
    const send = vi.fn(
      async (): Promise<OrdinaryCreateOutcome> => ({ kind: "uncertain" }),
    );
    owner.configure(async () => authority, {
      ...createOrdinaryCreateTransport(undefined),
      verify: async () => {},
      send,
    });
    owner.setAuthority(authority);
    owner.update("cartulary.view.hosts.v1", "host.hostname", "host.example");
    await owner.submit("cartulary.view.hosts.v1");
    expect(begin).toHaveBeenCalledWith({
      recordIds: [],
      unknownEntityType: "host",
    });
    expect(release).not.toHaveBeenCalled();
    owner.retire();
    expect(release).toHaveBeenCalledTimes(1);
    expect(versions).not.toHaveBeenCalled();
  });
});
