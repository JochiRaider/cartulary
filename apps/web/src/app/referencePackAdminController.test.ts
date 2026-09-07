import { waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  type ReferencePackJobResource,
  type ReferencePackMutation,
  type ReferencePackRead,
  referencePackTransportProblem,
} from "../services/referencePacks";
import { deferred } from "../testing/fetchMockTestSupport";
import {
  referencePackFixture,
  referencePackJobFixture,
  referencePackListFixture,
  referencePackTestActor,
  referencePackTestJobId,
} from "../testing/referencePackTestSupport";
import {
  ReferencePackAdminController,
  type ReferencePackAdminPorts,
  referencePackTiming,
} from "./referencePackAdminController";
import {
  acceptsReferencePackJob,
  emptyReferencePackQuery,
  referencePackBusy,
} from "./referencePackAdminModel";

const controllers: ReferencePackAdminController[] = [];
const authority = { lifetime: "lifetime-1", actorId: referencePackTestActor };
const failedRead = {
  kind: "failed",
  problem: referencePackTransportProblem,
} as const;
const secondId = "33333333-3333-4333-8333-333333333333";
function job(
  status: ReferencePackJobResource["status"],
  id = referencePackTestJobId,
) {
  return referencePackJobFixture(status, {
    job_id: id,
    status_route: `/api/v1/jobs/${id}`,
  });
}
function setup(overrides: Partial<ReferencePackAdminPorts> = {}) {
  let transaction = 0;
  const ports = {
    list: vi.fn<ReferencePackAdminPorts["list"]>().mockResolvedValue({
      kind: "read",
      value: referencePackListFixture([referencePackFixture()]),
    }),
    version: vi
      .fn<ReferencePackAdminPorts["version"]>()
      .mockResolvedValue({ kind: "read", value: referencePackFixture() }),
    submit: vi
      .fn<ReferencePackAdminPorts["submit"]>()
      .mockResolvedValue({ kind: "accepted", job: job("queued") }),
    job: vi
      .fn<ReferencePackAdminPorts["job"]>()
      .mockResolvedValue({ kind: "read", value: job("running") }),
    cancel: vi.fn<ReferencePackAdminPorts["cancel"]>().mockResolvedValue({
      kind: "acknowledged",
      job: job("cancel_requested"),
    }),
    confirmAccess: vi
      .fn<ReferencePackAdminPorts["confirmAccess"]>()
      .mockResolvedValue({ kind: "authorized" }),
    authorizationFailed: vi.fn(),
    isCurrent: () => true,
    transactionId: () => `txn-${++transaction}`,
  };
  const controller = new ReferencePackAdminController({
    ...ports,
    ...overrides,
  });
  controllers.push(controller);
  controller.setAuthority(authority);
  controller.setActive(true);
  return { controller, ports };
}
async function ready(controller: ReferencePackAdminController) {
  await waitFor(() => expect(controller.getSnapshot().reconciling).toBe(false));
}
async function observed(
  controller: ReferencePackAdminController,
  status = "running",
  id = referencePackTestJobId,
) {
  await waitFor(() => {
    expect(controller.getSnapshot().jobs[id]?.snapshot.status).toBe(status);
    expect(controller.getSnapshot().jobs[id]?.observation).toBe("idle");
  });
}
afterEach(() => {
  for (const controller of controllers.splice(0)) controller.dispose();
  vi.useRealTimers();
});

describe("Reference Pack execution", () => {
  it("reconciles a terminal replay receipt after another known job finishes reading", async () => {
    const { controller, ports } = setup();
    await ready(controller);
    await controller.run({ kind: "refresh_all" });
    await observed(controller);
    ports.job.mockResolvedValue({ kind: "read", value: job("succeeded") });
    await controller.retryObservation(referencePackTestJobId);
    ports.submit.mockResolvedValueOnce({
      kind: "uncertain",
      problem: referencePackTransportProblem,
    });
    await controller.run({
      kind: "refresh_selected",
      packKeys: ["type_registry.host"],
    });
    const oldRead = deferred<ReferencePackRead<ReferencePackJobResource>>();
    ports.job.mockReturnValueOnce(oldRead.promise);
    const reading = controller.retryObservation(referencePackTestJobId);
    await waitFor(() =>
      expect(
        controller.getSnapshot().jobs[referencePackTestJobId]?.observation,
      ).toBe("reading"),
    );
    vi.useFakeTimers();
    ports.submit.mockResolvedValueOnce({
      kind: "accepted",
      job: job("succeeded", secondId),
    });
    ports.job.mockResolvedValue({
      kind: "read",
      value: job("succeeded", secondId),
    });
    await controller.retryAttempt();
    expect(controller.getSnapshot().operation?.phase).toBe("accepted");
    oldRead.resolve({ kind: "read", value: job("succeeded") });
    await reading;
    await vi.advanceTimersByTimeAsync(referencePackTiming.poll);
    expect(controller.getSnapshot().operation?.phase).toBe("committed");
    expect(
      controller.getSnapshot().jobs[referencePackTestJobId]?.snapshot.status,
    ).toBe("succeeded");
    expect(ports.job.mock.calls.at(-1)?.[0]).toBe(secondId);
  });
  it("retires captured upload bytes and ignores a late admission after capability loss", async () => {
    const admission = deferred<ReferencePackMutation>();
    const { controller, ports } = setup();
    await ready(controller);
    controller.setFile(new File(["protected bytes"], "private.zip"));
    controller.setSelected("private-key", true);
    ports.submit.mockReturnValueOnce(admission.promise);
    const pending = controller.run({ kind: "import", filename: "private.zip" });
    await waitFor(() => expect(ports.submit).toHaveBeenCalledTimes(1));
    controller.setAuthority(null);
    expect(controller.getSnapshot().file).toBeNull();
    expect(controller.getSnapshot().operation).toBeNull();
    expect(controller.getSnapshot().catalog.rows).toEqual([]);
    expect(controller.getSnapshot().selectedKeys).toEqual([]);
    admission.resolve({ kind: "accepted", job: job("queued") });
    await pending;
    expect(controller.getSnapshot().jobs).toEqual({});
    expect(ports.job).not.toHaveBeenCalled();
    controller.dispose();
    const readsBeforeDisposal = ports.confirmAccess.mock.calls.length;
    controller.setAuthority(authority);
    await controller.retryAccess();
    expect(controller.getSnapshot().authority).toBeNull();
    expect(ports.confirmAccess).toHaveBeenCalledTimes(readsBeforeDisposal);
  });
  it("captures selected scope once and admits no conflicting operation", async () => {
    const admission = deferred<ReferencePackMutation>();
    const { controller, ports } = setup();
    await ready(controller);
    ports.submit.mockReturnValueOnce(admission.promise);
    controller.setSelected("outside", true);
    const keys = ["outside", "outside", "type_registry.host"];
    const pending = controller.run({
      kind: "refresh_selected",
      packKeys: keys,
    });
    keys.push("changed");
    await waitFor(() => expect(ports.submit).toHaveBeenCalledTimes(1));
    await controller.run({ kind: "refresh_all" });
    expect(ports.submit).toHaveBeenCalledTimes(1);
    expect(ports.submit.mock.calls[0]?.[0].command).toEqual({
      kind: "refresh_selected",
      packKeys: ["outside", "type_registry.host"],
    });
    admission.resolve({ kind: "accepted", job: job("queued") });
    await pending;
    await observed(controller);
    await controller.run({ kind: "refresh_all" });
    expect(ports.submit).toHaveBeenCalledTimes(1);
    ports.job.mockResolvedValue({ kind: "read", value: job("succeeded") });
    await controller.retryObservation(referencePackTestJobId);
    expect(referencePackBusy(controller.getSnapshot())).toBe(false);
    expect(controller.getSnapshot().selectedKeys).toEqual(["outside"]);
  });

  it("replays an uncertain upload with the original bytes target and transaction", async () => {
    const { controller, ports } = setup();
    await ready(controller);
    const file = new File([new Uint8Array([0, 1, 255, 42])], "bundle.zip", {
      type: "application/zip",
    });
    controller.setFile(file);
    ports.submit.mockResolvedValueOnce({
      kind: "uncertain",
      problem: referencePackTransportProblem,
    });
    await controller.run({ kind: "import", filename: file.name });
    const captured = ports.submit.mock.calls[0]?.[0];
    controller.setFile(new File(["different"], "replacement.zip"));
    expect(controller.getSnapshot().file).toBe(file);
    controller.setActive(false);
    expect(controller.getSnapshot().operation?.attempt).toBe(captured);
    controller.setActive(true);
    await ready(controller);
    await controller.retryAttempt();
    expect(ports.submit.mock.calls[1]?.[0]).toBe(captured);
    expect(ports.submit.mock.calls[1]?.[2]).toBe(true);
    expect(captured?.file).toBe(file);
    expect(captured?.request.client_txn_id).toBe("txn-1");
    expect(controller.getSnapshot().file).toBeNull();
  });

  it("keeps an inline commit committed when catalog reconciliation fails", async () => {
    const { controller, ports } = setup();
    await ready(controller);
    ports.submit.mockResolvedValue({
      kind: "committed",
      pack: referencePackFixture({ active: true }),
    });
    ports.list.mockResolvedValue(failedRead);
    await controller.run({
      kind: "activate",
      target: { pack_key: "type_registry.host", pack_version: "1" },
    });
    await waitFor(() =>
      expect(controller.getSnapshot().catalog.problem).not.toBeNull(),
    );
    expect(controller.getSnapshot().operation?.phase).toBe("committed");
    await controller.reload();
    expect(ports.submit).toHaveBeenCalledTimes(1);
    expect(controller.getSnapshot().operation?.phase).toBe("committed");
  });

  it("reviews exact version eligibility without substituting a different mutation", async () => {
    const { controller, ports } = setup();
    await ready(controller);
    ports.version.mockResolvedValue({
      kind: "read",
      value: referencePackFixture({ pack_version_state: "disabled" }),
    });
    const target = { pack_key: "type_registry.host", pack_version: "1" };
    await controller.run({ kind: "activate", target });
    expect(ports.version.mock.calls[0]?.[0]).toEqual(target);
    expect(ports.submit).not.toHaveBeenCalled();
    expect(controller.getSnapshot().operation?.problem?.kind).toBe(
      "state_conflict",
    );
    await controller.reload();
    expect(ports.submit).not.toHaveBeenCalled();
  });

  it("recovers observation failure and keeps multiple known jobs isolated from late reads", async () => {
    const oldRead = deferred<ReferencePackRead<ReferencePackJobResource>>();
    const { controller, ports } = setup();
    await ready(controller);
    ports.job.mockResolvedValueOnce(failedRead);
    await controller.run({ kind: "refresh_all" });
    await waitFor(() =>
      expect(
        controller.getSnapshot().jobs[referencePackTestJobId]?.observation,
      ).toBe("failed"),
    );
    expect(controller.getSnapshot().operation?.phase).toBe("accepted");
    ports.job.mockReturnValueOnce(oldRead.promise);
    const pending = controller.retryObservation(referencePackTestJobId);
    await waitFor(() => expect(ports.job).toHaveBeenCalledTimes(2));
    controller.setActive(false);
    ports.job.mockResolvedValue({ kind: "read", value: job("succeeded") });
    controller.setActive(true);
    await ready(controller);
    await observed(controller, "succeeded");
    ports.submit.mockResolvedValue({
      kind: "accepted",
      job: job("queued", secondId),
    });
    ports.job.mockResolvedValue({
      kind: "read",
      value: job("running", secondId),
    });
    await controller.run({
      kind: "refresh_selected",
      packKeys: ["type_registry.host"],
    });
    await observed(controller, "running", secondId);
    oldRead.resolve({ kind: "read", value: job("running") });
    await pending;
    expect(Object.keys(controller.getSnapshot().jobs)).toEqual([
      referencePackTestJobId,
      secondId,
    ]);
    expect(
      controller.getSnapshot().jobs[referencePackTestJobId]?.snapshot.status,
    ).toBe("succeeded");
    expect(controller.getSnapshot().operation?.jobId).toBe(secondId);
    const pausedRead = deferred<ReferencePackRead<ReferencePackJobResource>>();
    ports.job.mockReturnValueOnce(pausedRead.promise);
    const observingPaused = controller.retryObservation(secondId);
    await waitFor(() =>
      expect(controller.getSnapshot().jobs[secondId]?.observation).toBe(
        "reading",
      ),
    );
    controller.setActive(false);
    ports.job.mockImplementation(async (id) =>
      id === referencePackTestJobId
        ? {
            kind: "failed",
            problem: {
              kind: "unavailable",
              status: 404,
              code: "job_not_found",
            },
          }
        : { kind: "read", value: job("running", secondId) },
    );
    vi.useFakeTimers();
    controller.setActive(true);
    await vi.advanceTimersByTimeAsync(referencePackTiming.poll + 100);
    expect(
      controller.getSnapshot().jobs[referencePackTestJobId]?.observation,
    ).toBe("unavailable");
    expect(controller.getSnapshot().jobs[secondId]?.observation).toBe("idle");
    vi.useRealTimers();
    pausedRead.resolve({ kind: "read", value: job("running", secondId) });
    await observingPaused;
    const dismissedRead =
      deferred<ReferencePackRead<ReferencePackJobResource>>();
    ports.job.mockReturnValueOnce(dismissedRead.promise);
    const observingDismissed = controller.retryObservation(
      referencePackTestJobId,
    );
    await waitFor(() =>
      expect(
        controller.getSnapshot().jobs[referencePackTestJobId]?.observation,
      ).toBe("reading"),
    );
    controller.dismissJob(referencePackTestJobId);
    const announcement = controller.getSnapshot().announcement.serial;
    dismissedRead.resolve(failedRead);
    await observingDismissed;
    expect(controller.getSnapshot().announcement.serial).toBe(announcement);
    expect(controller.getSnapshot().jobs[secondId]?.snapshot.status).toBe(
      "running",
    );
  });

  it("coalesces completion behind the newest query and preserves failed query recovery", async () => {
    const queryRead =
      deferred<Awaited<ReturnType<ReferencePackAdminPorts["list"]>>>();
    const { controller, ports } = setup();
    await ready(controller);
    await controller.run({ kind: "refresh_all" });
    await observed(controller);
    ports.list.mockReturnValueOnce(queryRead.promise);
    controller.setQuery({ ...emptyReferencePackQuery, search: "newest" });
    await waitFor(() => expect(ports.list).toHaveBeenCalledTimes(2));
    ports.job.mockResolvedValue({ kind: "read", value: job("succeeded") });
    await controller.retryObservation(referencePackTestJobId);
    expect(ports.list).toHaveBeenCalledTimes(2);
    queryRead.resolve(failedRead);
    await waitFor(() =>
      expect(controller.getSnapshot().catalog.problem).not.toBeNull(),
    );
    expect(ports.list).toHaveBeenCalledTimes(2);
    expect(controller.getSnapshot().operation?.phase).toBe("committed");
    ports.list.mockResolvedValue({
      kind: "read",
      value: referencePackListFixture(),
    });
    await controller.reload();
    expect(ports.list.mock.calls[2]?.[0].query.search).toBe("newest");
    expect(controller.getSnapshot().catalog.problem).toBeNull();
  });

  it("accepts a hidden admission without automatic reads and clears every retired continuation", async () => {
    const admission = deferred<ReferencePackMutation>();
    const { controller, ports } = setup();
    await ready(controller);
    ports.submit.mockReturnValueOnce(admission.promise);
    const pending = controller.run({ kind: "refresh_all" });
    await waitFor(() => expect(ports.submit).toHaveBeenCalledTimes(1));
    controller.setActive(false);
    admission.resolve({ kind: "accepted", job: job("queued") });
    await pending;
    expect(controller.getSnapshot().operation?.phase).toBe("accepted");
    expect(ports.job).not.toHaveBeenCalled();
    expect(ports.list).toHaveBeenCalledTimes(1);
    controller.setActive(true);
    await ready(controller);
    await observed(controller);
    const late = deferred<ReferencePackRead<ReferencePackJobResource>>();
    ports.job.mockReturnValueOnce(late.promise);
    const reading = controller.retryObservation(referencePackTestJobId);
    await waitFor(() => expect(ports.job).toHaveBeenCalledTimes(2));
    controller.setSelected("protected", true);
    controller.setAuthority({ lifetime: "lifetime-2", actorId: "replacement" });
    late.resolve({ kind: "read", value: job("succeeded") });
    await reading;
    expect(controller.getSnapshot().jobs).toEqual({});
    expect(controller.getSnapshot().operation).toBeNull();
    expect(controller.getSnapshot().selectedKeys).toEqual([]);
    expect(controller.getSnapshot().file).toBeNull();
  });

  it("bounds ignored-abort admission and observation without treating read failure as job failure", async () => {
    const { controller, ports } = setup();
    await ready(controller);
    vi.useFakeTimers();
    const admission = deferred<ReferencePackMutation>();
    ports.submit.mockReturnValueOnce(admission.promise);
    const pending = controller.run({ kind: "refresh_all" });
    await vi.advanceTimersByTimeAsync(referencePackTiming.admission);
    await pending;
    expect(controller.getSnapshot().operation?.phase).toBe("uncertain");
    const read = deferred<ReferencePackRead<ReferencePackJobResource>>();
    ports.job.mockReturnValueOnce(read.promise);
    await controller.retryAttempt();
    await vi.advanceTimersByTimeAsync(referencePackTiming.read);
    expect(
      controller.getSnapshot().jobs[referencePackTestJobId]?.observation,
    ).toBe("failed");
    expect(controller.getSnapshot().operation?.phase).toBe("accepted");
    const count = ports.job.mock.calls.length;
    await vi.advanceTimersByTimeAsync(10_000);
    expect(ports.job).toHaveBeenCalledTimes(count);
    await controller.retryObservation(referencePackTestJobId);
    expect(
      controller.getSnapshot().jobs[referencePackTestJobId]?.snapshot.status,
    ).toBe("running");
    admission.resolve({ kind: "accepted", job: job("queued", secondId) });
    read.resolve({ kind: "read", value: job("failed") });
    await vi.advanceTimersByTimeAsync(0);
    expect(controller.getSnapshot().jobs[secondId]).toBeUndefined();
    expect(
      controller.getSnapshot().jobs[referencePackTestJobId]?.snapshot.status,
    ).toBe("running");
  });

  it("keeps concealed observation unavailable until access is confirmed and clears confirmed loss", async () => {
    const { controller, ports } = setup();
    await ready(controller);
    ports.job.mockResolvedValue({
      kind: "failed",
      problem: { kind: "unavailable", status: 404, code: "job_not_found" },
    });
    ports.confirmAccess.mockResolvedValueOnce({ kind: "unavailable" });
    await controller.run({ kind: "refresh_all" });
    await waitFor(() =>
      expect(controller.getSnapshot().access).toBe("unavailable"),
    );
    expect(
      controller.getSnapshot().jobs[referencePackTestJobId]?.snapshot.status,
    ).toBe("queued");
    expect(ports.authorizationFailed).not.toHaveBeenCalled();
    ports.confirmAccess.mockResolvedValue({ kind: "authorized" });
    ports.job.mockResolvedValue({ kind: "access_failed", status: 403 });
    await controller.retryAccess();
    expect(ports.authorizationFailed).toHaveBeenCalledWith(403);
    expect(controller.getSnapshot().jobs).toEqual({});
    expect(controller.getSnapshot().catalog.rows).toEqual([]);
    expect(controller.getSnapshot().catalog.problem).toBeNull();
  });

  it("observes a completion race before cancellation and never sends a replacement action", async () => {
    const { controller, ports } = setup();
    await ready(controller);
    await controller.run({ kind: "refresh_all" });
    await observed(controller);
    ports.job.mockResolvedValue({ kind: "read", value: job("succeeded") });
    await controller.cancelJob(referencePackTestJobId);
    expect(ports.cancel).not.toHaveBeenCalled();
    expect(controller.getSnapshot().operation?.phase).toBe("committed");
    expect(ports.submit).toHaveBeenCalledTimes(1);
  });

  it("replays only the captured uncertain cancellation and observes authoritative completion", async () => {
    const { controller, ports } = setup();
    await ready(controller);
    await controller.run({ kind: "refresh_all" });
    await observed(controller);
    ports.cancel.mockResolvedValueOnce({
      kind: "uncertain",
      problem: referencePackTransportProblem,
    });
    await controller.cancelJob(referencePackTestJobId);
    await observed(controller);
    const capture = ports.cancel.mock.calls[0]?.[1];
    await controller.cancelJob(referencePackTestJobId);
    expect(ports.cancel).toHaveBeenCalledTimes(1);
    ports.cancel.mockResolvedValue({
      kind: "rejected",
      problem: { kind: "rejected", status: 409, code: "job_cancel_rejected" },
    });
    ports.job.mockResolvedValue({ kind: "read", value: job("succeeded") });
    await controller.cancelJob(referencePackTestJobId, true);
    await observed(controller, "succeeded");
    expect(ports.cancel.mock.calls[1]?.[1]).toBe(capture);
    expect(controller.getSnapshot().operation?.phase).toBe("committed");
  });

  it("rejects regressing observations and keeps terminal snapshots stable", () => {
    expect(acceptsReferencePackJob(job("running"), job("queued"))).toBe(false);
    expect(acceptsReferencePackJob(job("succeeded"), job("running"))).toBe(
      false,
    );
    expect(
      acceptsReferencePackJob(job("running"), job("running", secondId)),
    ).toBe(false);
    expect(acceptsReferencePackJob(job("queued"), job("canceled"))).toBe(true);
    expect(
      acceptsReferencePackJob(job("running"), job("cancel_requested")),
    ).toBe(true);
    expect(acceptsReferencePackJob(job("running"), job("failed"))).toBe(true);
    expect(acceptsReferencePackJob(job("succeeded"), job("succeeded"))).toBe(
      true,
    );
    expect(
      acceptsReferencePackJob(
        referencePackJobFixture("running", {
          progress: { completed: 3, total: 5 },
        }),
        referencePackJobFixture("running", {
          progress: { completed: 2, total: 5 },
        }),
      ),
    ).toBe(false);
  });
});
