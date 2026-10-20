import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../testing/fetchMockTestSupport";
import {
  importActorID,
  importedIncidentID,
  importJob,
  importJobID,
} from "../testing/incidentImportTestSupport";
import type {
  ImportAdmissionOutcome,
  ImportCancellationOutcome,
  ImportObservationOutcome,
  IncidentImportJob,
} from "./api/incidentImportClient";
import {
  IncidentImportController,
  type IncidentImportPorts,
} from "./incidentImportModel";

function response(job: IncidentImportJob, status: 202): ImportAdmissionOutcome;
function response(job?: IncidentImportJob): ImportObservationOutcome;
function response(
  job = importJob(),
  status = 200,
): ImportAdmissionOutcome | ImportObservationOutcome {
  return status === 202 ? { kind: "accepted", job } : { kind: "observed", job };
}
const acknowledged = (job: IncidentImportJob): ImportCancellationOutcome => ({
  kind: "acknowledged",
  job,
});
const controllers: IncidentImportController[] = [];
const nextJobId = "00000000-0000-4000-8000-000000005004";
function setup(overrides: Partial<IncidentImportPorts> = {}) {
  const ports = {
    admit: vi
      .fn<IncidentImportPorts["admit"]>()
      .mockResolvedValue(response(importJob(), 202)),
    read: vi
      .fn<IncidentImportPorts["read"]>()
      .mockResolvedValue(response(importJob("running"))),
    cancel: vi
      .fn<IncidentImportPorts["cancel"]>()
      .mockResolvedValue(acknowledged(importJob("cancel_requested"))),
    openIncident: vi
      .fn<IncidentImportPorts["openIncident"]>()
      .mockResolvedValue("opened"),
    isCurrent: vi.fn().mockReturnValue(true),
    confirmAccess: async () => ({ kind: "unavailable" as const }),
    authorizationFailed: vi.fn(),
    transactionId: vi
      .fn()
      .mockReturnValueOnce("txn-one")
      .mockReturnValue("txn-two"),
    ...overrides,
  };
  const controller = new IncidentImportController(ports);
  controllers.push(controller);
  controller.setAuthority({ lifetime: "session-one", actorId: importActorID });
  controller.setActive(true);
  const file = new File(["exact bytes"], "archive.zip", {
    type: "application/zip",
  });
  controller.selectFile(file);
  return { controller, ports, file };
}
const settle = async () => {
  for (let i = 0; i < 10; i++) await Promise.resolve();
};
const tick = async (ms = 0) => {
  await settle();
  await vi.advanceTimersByTimeAsync(ms);
  await settle();
};
beforeEach(() => vi.useFakeTimers());
afterEach(() => {
  for (const c of controllers.splice(0)) c.dispose();
  vi.useRealTimers();
});

describe("incident import lifecycle", () => {
  it("captures one immutable admission and replays exactly after loss without file replacement", async () => {
    const pending = deferred<ImportAdmissionOutcome>();
    const admit = vi
      .fn<IncidentImportPorts["admit"]>()
      .mockReturnValueOnce(pending.promise)
      .mockResolvedValue(response(importJob("succeeded"), 202));
    const read = vi
      .fn<IncidentImportPorts["read"]>()
      .mockResolvedValue(response(importJob("succeeded")));
    const { controller, file, ports } = setup({ admit, read });
    controller.submit();
    controller.submit();
    controller.retryAdmission();
    controller.selectFile(new File(["replacement"], "replacement.zip"));
    expect(admit).toHaveBeenCalledTimes(1);
    expect(controller.getSnapshot().selectedFile).toBe(file);
    pending.reject(new TypeError("lost"));
    await tick();
    expect(controller.getSnapshot().admission.kind).toBe("uncertain");
    controller.setActive(false);
    controller.setActive(true);
    controller.retryAdmission();
    controller.retryAdmission();
    await tick();
    expect(admit).toHaveBeenCalledTimes(2);
    expect(admit.mock.calls[0]?.[2]).toBe(false);
    expect(admit.mock.calls[1]?.[2]).toBe(true);
    expect(admit.mock.calls[0]?.[0]).toBe(admit.mock.calls[1]?.[0]);
    expect(admit.mock.calls[1]?.[0].file).toBe(file);
    expect(controller.getSnapshot().order).toEqual([importJobID]);
    expect(controller.getSnapshot().selectedFile).toBeNull();
    controller.refresh();
    await tick();
    expect(admit).toHaveBeenCalledTimes(2);
    expect(ports.read).toHaveBeenCalled();
  });
  it("keeps timeout and malformed admission uncertain while definitive rejection unlocks input", async () => {
    const gate = deferred<ImportAdmissionOutcome>();
    const admit = vi
      .fn<IncidentImportPorts["admit"]>()
      .mockReturnValueOnce(gate.promise)
      .mockResolvedValueOnce({ kind: "uncertain", problem: "contract" })
      .mockResolvedValueOnce({ kind: "rejected", problem: "rejected" });
    const { controller } = setup({ admit });
    controller.submit();
    await tick(120_000);
    expect(controller.getSnapshot().admission.kind).toBe("uncertain");
    controller.retryAdmission();
    await tick();
    gate.resolve(response(importJob(), 202));
    await tick();
    expect(controller.getSnapshot().admission.kind).toBe("uncertain");
    expect(controller.getSnapshot().order).toEqual([]);
    controller.retryAdmission();
    await tick();
    expect(controller.getSnapshot().admission.kind).toBe("rejected");
    const next = new File(["next"], "next.zip");
    controller.selectFile(next);
    expect(controller.getSnapshot().selectedFile).toBe(next);
  });
  it("recovers a failed or timed out read and retains the last validated status", async () => {
    const read = vi
      .fn<IncidentImportPorts["read"]>()
      .mockResolvedValueOnce({ kind: "failed", problem: "transport" })
      .mockResolvedValueOnce(response(importJob("running")))
      .mockReturnValueOnce(deferred<ImportObservationOutcome>().promise)
      .mockResolvedValue(response(importJob("succeeded")));
    const { controller } = setup({ read });
    controller.submit();
    await tick();
    expect(controller.getSnapshot().jobs[importJobID]).toMatchObject({
      job: { status: "queued" },
      observation: { kind: "failed" },
    });
    await tick(10_000);
    expect(read).toHaveBeenCalledTimes(1);
    controller.refresh();
    await tick();
    expect(controller.getSnapshot().jobs[importJobID]?.job.status).toBe(
      "running",
    );
    await tick(31_000);
    expect(controller.getSnapshot().jobs[importJobID]?.observation.kind).toBe(
      "failed",
    );
    controller.refresh();
    await tick();
    expect(controller.getSnapshot().jobs[importJobID]?.job.status).toBe(
      "succeeded",
    );
  });
  it("observes multiple jobs fairly without overlapping reads or replacing selection", async () => {
    const firstRead = deferred<ImportObservationOutcome>();
    const job2 = importJob("queued", {
      job_id: nextJobId,
      status_route: `/api/v1/jobs/${nextJobId}`,
    });
    const admit = vi
      .fn<IncidentImportPorts["admit"]>()
      .mockResolvedValueOnce(response(importJob(), 202))
      .mockResolvedValueOnce(response(job2, 202));
    const read = vi
      .fn<IncidentImportPorts["read"]>()
      .mockReturnValueOnce(firstRead.promise)
      .mockImplementation(async (id) =>
        response(
          importJob("running", {
            job_id: id,
            status_route: `/api/v1/jobs/${id}`,
          }),
        ),
      );
    const { controller } = setup({ admit, read });
    controller.submit();
    await tick();
    controller.selectFile(new File(["second"], "second.zip"));
    controller.submit();
    await tick();
    expect(read).toHaveBeenCalledTimes(1);
    expect(controller.getSnapshot().selectedJobId).toBe(nextJobId);
    firstRead.resolve(response(importJob("running")));
    await tick(1000);
    expect(read.mock.calls[1]?.[0]).toBe(nextJobId);
    expect(controller.getSnapshot().selectedJobId).toBe(nextJobId);
    await tick(4000);
    expect(read.mock.calls.slice(2).map(([id]) => id)).toEqual([
      importJobID,
      nextJobId,
      importJobID,
      nextJobId,
    ]);
  });
  it("pauses reads across hidden panels and rejects late continuations after retirement", async () => {
    const oldRead = deferred<ImportObservationOutcome>();
    const read = vi
      .fn<IncidentImportPorts["read"]>()
      .mockReturnValueOnce(oldRead.promise)
      .mockResolvedValue(response(importJob("running")));
    const { controller } = setup({ read });
    controller.submit();
    await tick();
    controller.setActive(false);
    expect(read.mock.calls[0]?.[1].aborted).toBe(true);
    oldRead.resolve(response(importJob("succeeded")));
    await tick(5000);
    expect(controller.getSnapshot().jobs[importJobID]?.job.status).toBe(
      "queued",
    );
    controller.setActive(true);
    await tick();
    expect(controller.getSnapshot().jobs[importJobID]?.job.status).toBe(
      "running",
    );
    controller.pause();
    await tick(5000);
    expect(read).toHaveBeenCalledTimes(2);
    controller.resume();
    await tick();
    expect(read).toHaveBeenCalledTimes(3);
    controller.retire();
    await tick(5000);
    expect(controller.getSnapshot().order).toEqual([]);
    expect(controller.getSnapshot().selectedFile).toBeNull();
  });
  it("retains in-flight admission across navigation but excludes lost capability and new sessions", async () => {
    for (const retirement of ["navigation", "capability", "session"] as const) {
      const gate = deferred<ImportAdmissionOutcome>();
      const current = vi.fn().mockReturnValue(true);
      const { controller } = setup({
        admit: () => gate.promise,
        isCurrent: current,
      });
      controller.submit();
      controller.setActive(false);
      if (retirement === "capability") {
        current.mockReturnValue(false);
        controller.retire();
      }
      if (retirement === "session")
        controller.setAuthority({
          lifetime: "session-two",
          actorId: importActorID,
        });
      gate.resolve(response(importJob(), 202));
      await tick();
      expect(controller.getSnapshot().order).toEqual(
        retirement === "navigation" ? [importJobID] : [],
      );
      if (retirement !== "navigation")
        expect(controller.getSnapshot().selectedFile).toBeNull();
    }
  });
  it("observes cancellation uncertainty and reuses its exact id only while cancellation remains allowed", async () => {
    const cancel = vi
      .fn<IncidentImportPorts["cancel"]>()
      .mockRejectedValueOnce(new TypeError("lost cancel"))
      .mockResolvedValue(acknowledged(importJob("cancel_requested")));
    const { controller } = setup({ cancel });
    controller.submit();
    await tick();
    controller.cancel();
    controller.cancel();
    await tick();
    expect(cancel).toHaveBeenCalledTimes(1);
    expect(controller.getSnapshot().jobs[importJobID]).toMatchObject({
      job: { status: "running" },
      cancellation: { kind: "uncertain" },
    });
    controller.cancel();
    await tick();
    expect(cancel.mock.calls[0]?.[1]).toBe(cancel.mock.calls[1]?.[1]);
    expect(controller.getSnapshot().jobs[importJobID]?.job.status).toBe(
      "cancel_requested",
    );
    controller.cancel();
    expect(cancel).toHaveBeenCalledTimes(2);
  });
  it("handles rejected cancellation and a completion race through authoritative reads", async () => {
    for (const result of [
      { kind: "rejected", problem: "rejected" } as ImportCancellationOutcome,
      acknowledged(importJob("succeeded")),
    ]) {
      const read = vi
        .fn<IncidentImportPorts["read"]>()
        .mockResolvedValueOnce(response(importJob("running")))
        .mockResolvedValue(response(importJob("succeeded")));
      const { controller } = setup({ read, cancel: async () => result });
      controller.submit();
      await tick();
      controller.cancel();
      await tick();
      expect(controller.getSnapshot().jobs[importJobID]?.job.status).toBe(
        "succeeded",
      );
    }
    const pending = deferred<ImportCancellationOutcome>();
    const read = vi
      .fn<IncidentImportPorts["read"]>()
      .mockResolvedValueOnce(response(importJob("running")))
      .mockResolvedValueOnce(response(importJob("running")))
      .mockResolvedValue(response(importJob("succeeded")));
    const { controller, ports } = setup({
      read,
      cancel: () => pending.promise,
    });
    controller.submit();
    await tick();
    controller.cancel();
    await tick();
    controller.setActive(false);
    pending.resolve(acknowledged(importJob("succeeded")));
    await tick();
    expect(controller.getSnapshot().jobs[importJobID]?.observation.kind).toBe(
      "stale",
    );
    controller.setActive(true);
    controller.open();
    expect(ports.openIncident).not.toHaveBeenCalled();
    await tick();
    expect(ports.openIncident).toHaveBeenCalledOnce();
  });
  it("keeps terminal state across malformed or regressive reads and marks missing jobs unavailable", async () => {
    const read = vi
      .fn<IncidentImportPorts["read"]>()
      .mockResolvedValueOnce(response(importJob("succeeded")))
      .mockResolvedValueOnce(response(importJob("running")))
      .mockResolvedValueOnce({ kind: "unavailable" });
    const { controller, ports } = setup({ read });
    controller.submit();
    await tick();
    controller.refresh();
    await tick();
    expect(controller.getSnapshot().jobs[importJobID]).toMatchObject({
      job: { status: "succeeded" },
      observation: { kind: "failed" },
    });
    controller.open();
    expect(ports.openIncident).not.toHaveBeenCalled();
    controller.refresh();
    await tick();
    expect(controller.getSnapshot().jobs[importJobID]?.observation.kind).toBe(
      "unavailable",
    );
  });
  it("requires explicit opening and retries only the handoff after workbook failure", async () => {
    const openIncident = vi
      .fn<IncidentImportPorts["openIncident"]>()
      .mockResolvedValueOnce("unavailable")
      .mockResolvedValue("opened");
    const { controller, ports } = setup({
      read: async () => response(importJob("succeeded")),
      openIncident,
    });
    controller.submit();
    await tick();
    expect(openIncident).not.toHaveBeenCalled();
    controller.open();
    controller.open();
    await tick();
    expect(controller.getSnapshot().navigation.kind).toBe("unavailable");
    expect(controller.getSnapshot().jobs[importJobID]?.job.status).toBe(
      "succeeded",
    );
    controller.open();
    await tick();
    expect(openIncident.mock.calls[0]?.[0]).toBe(importedIncidentID);
    expect(openIncident).toHaveBeenCalledTimes(2);
    expect(ports.admit).toHaveBeenCalledTimes(1);
  });
  it("guards delayed handoff and authorization failures without stale protected publication", async () => {
    const gate = deferred<"opened">();
    const { controller, ports } = setup({
      read: async () => response(importJob("succeeded")),
      openIncident: () => gate.promise,
    });
    controller.submit();
    await tick();
    controller.open();
    await tick();
    controller.setActive(false);
    gate.resolve("opened");
    await tick();
    expect(controller.getSnapshot().navigation.kind).toBe("idle");
    controller.setActive(true);
    await tick();
    controller.retire();
    controller.open();
    expect(ports.admit).toHaveBeenCalledTimes(1);
    const second = setup({
      admit: async () => ({ kind: "access_failed", status: 403 }),
    });
    second.controller.submit();
    await tick();
    expect(second.controller.getSnapshot().order).toEqual([]);
    expect(second.ports.authorizationFailed).toHaveBeenCalledWith(403);
  });
  it("renders no inferred navigation for nonsuccess and uses server reads for job unavailability", async () => {
    for (const status of [
      "queued",
      "running",
      "cancel_requested",
      "failed",
      "canceled",
    ] as const) {
      const { controller, ports } = setup({
        read: async () => response(importJob(status)),
      });
      controller.submit();
      await tick();
      controller.open();
      expect(ports.openIncident).not.toHaveBeenCalled();
      expect(controller.getSnapshot().jobs[importJobID]?.job.status).toBe(
        status,
      );
      controller.dispose();
    }
    const now = Date.now();
    const terminal: IncidentImportJob = importJob("succeeded", {
      submitted_at: new Date(now - 9 * 86_400_000).toISOString(),
      updated_at: new Date(now - 8 * 86_400_000).toISOString(),
      started_at: new Date(now - 8 * 86_400_000).toISOString(),
      finished_at: new Date(now - 8 * 86_400_000).toISOString(),
      retained_until: new Date(now + 1000).toISOString(),
    });
    const { controller, ports } = setup({
      admit: async () =>
        response(
          importJob("queued", {
            submitted_at: terminal.submitted_at,
            updated_at: terminal.updated_at,
          }),
          202,
        ),
      read: vi
        .fn<IncidentImportPorts["read"]>()
        .mockResolvedValueOnce(response(terminal))
        .mockResolvedValue({ kind: "unavailable" }),
    });
    controller.submit();
    await tick(1000);
    expect(controller.getSnapshot().jobs[importJobID]?.observation.kind).toBe(
      "ready",
    );
    controller.open();
    await tick();
    expect(controller.getSnapshot().jobs[importJobID]?.observation.kind).toBe(
      "unavailable",
    );
    expect(ports.openIncident).not.toHaveBeenCalled();
  });
});
