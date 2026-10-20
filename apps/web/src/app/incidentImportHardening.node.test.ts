import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { deferred, jsonResponse } from "../testing/fetchMockTestSupport";
import {
  importActorID,
  importedIncidentID,
  importJob,
  importJobID,
  jobEnvelope,
} from "../testing/incidentImportTestSupport";
import {
  cancelImportJob,
  captureImport,
  importIncidentBundle,
  importJobAdvances,
  readImportJob,
} from "./api/incidentImportClient";
import {
  IncidentImportController,
  openableImport,
} from "./incidentImportModel";

const controllers: IncidentImportController[] = [];
function selected(controller: IncidentImportController) {
  const entry = controller.getSnapshot().jobs[importJobID];
  if (!entry) throw new Error("Expected admitted job");
  return entry;
}
const signal = () => new AbortController().signal;
const attempt = () =>
  captureImport(new File(["exact bytes"], "bundle.tar"), "fixed-admission");
const tick = async (ms = 0) => {
  for (let i = 0; i < 12; i++) await Promise.resolve();
  await vi.advanceTimersByTimeAsync(ms);
};
function setup(status: "running" | "succeeded" = "succeeded") {
  const fetch = vi
    .fn()
    .mockImplementation(async (path: string) =>
      jsonResponse(
        jobEnvelope(importJob(path.endsWith("/import") ? "queued" : status)),
        path.endsWith("/import") ? 202 : 200,
      ),
    );
  vi.stubGlobal("fetch", fetch);
  const openIncident = vi.fn().mockResolvedValue("opened");
  const read = vi.fn(readImportJob);
  const cancel = vi.fn(cancelImportJob);
  const confirmAccess = vi.fn().mockResolvedValue({ kind: "unavailable" });
  const controller = new IncidentImportController({
    admit: importIncidentBundle,
    read,
    cancel,
    openIncident,
    isCurrent: () => true,
    confirmAccess,
    authorizationFailed: vi.fn(),
    transactionId: () => "fixed-cancel",
  });
  controllers.push(controller);
  controller.setAuthority({ lifetime: "session-one", actorId: importActorID });
  controller.setActive(true);
  controller.selectFile(new File(["exact bytes"], "bundle.tar"));
  controller.submit();
  return { controller, fetch, openIncident, read, cancel, confirmAccess };
}
beforeEach(() => vi.useFakeTimers());
afterEach(() => {
  for (const controller of controllers.splice(0)) controller.dispose();
  vi.useRealTimers();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

it("normalizes malformed admission bodies without losing received rejection or authorization status", async () => {
  for (const status of [400, 401, 403, 409, 413, 429]) {
    for (const body of [
      "null",
      "7",
      "[]",
      "false",
      "{",
      "<html>private</html>",
      "",
    ]) {
      vi.stubGlobal(
        "fetch",
        vi.fn().mockResolvedValue(
          new Response(body, {
            status,
            headers: { "Content-Type": "application/json" },
          }),
        ),
      );
      await expect(
        importIncidentBundle(attempt(), signal()),
      ).resolves.toMatchObject({
        kind: status === 401 || status === 403 ? "access_failed" : "rejected",
      });
    }
  }
});

it("settles malformed read and cancellation bodies with operation-specific outcomes", async () => {
  for (const status of [400, 401, 403, 404, 408, 409, 429, 500]) {
    for (const body of [
      "null",
      "7",
      "[]",
      "false",
      "{",
      "<html>private</html>",
      "",
    ]) {
      vi.stubGlobal(
        "fetch",
        vi.fn().mockImplementation(
          async () =>
            new Response(body, {
              status,
              headers: { "Content-Type": "application/json" },
            }),
        ),
      );
      const access = status === 401 || status === 403;
      const read = await readImportJob(importJobID, signal());
      expect(read.kind).toBe(
        access ? "access_failed" : status === 404 ? "unavailable" : "failed",
      );
      const cancel = await cancelImportJob(
        importJobID,
        { client_txn_id: "cancel" },
        signal(),
      );
      expect(cancel.kind).toBe(
        access
          ? "access_failed"
          : status === 404
            ? "unavailable"
            : status === 408 || status >= 500
              ? "uncertain"
              : "rejected",
      );
      expect(JSON.stringify([read, cancel])).not.toContain("private");
    }
  }
  for (const broken of [
    () => {
      throw new TypeError("private");
    },
    () => Promise.reject(new TypeError("private")),
    async () =>
      new Response("{", {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    async () => jsonResponse({ data: {} }, 202),
  ]) {
    vi.stubGlobal("fetch", vi.fn(broken));
    expect((await importIncidentBundle(attempt(), signal())).kind).toBe(
      "uncertain",
    );
    expect((await readImportJob(importJobID, signal())).kind).toBe("failed");
    expect(
      (
        await cancelImportJob(
          importJobID,
          { client_txn_id: "cancel" },
          signal(),
        )
      ).kind,
    ).toBe("uncertain");
  }
});

it("retains a supported result action throughout an unchanged refresh", async () => {
  const { controller, fetch } = setup();
  await tick();
  const gate = deferred<Response>();
  fetch.mockReturnValueOnce(gate.promise);
  controller.refresh();
  await tick();
  expect(openableImport(selected(controller))).toBe(importedIncidentID);
  gate.resolve(jsonResponse(jobEnvelope(importJob("succeeded"))));
  await tick();
  expect(openableImport(selected(controller))).toBe(importedIncidentID);
});

it("requires a newly started job read before opening and suppresses duplicate intent", async () => {
  const { controller, fetch, openIncident, read } = setup();
  await tick();
  const gate = deferred<Response>();
  fetch.mockReturnValueOnce(gate.promise);
  controller.open();
  controller.open();
  await tick();
  expect(read).toHaveBeenCalledTimes(2);
  expect(openIncident).not.toHaveBeenCalled();
  gate.resolve(jsonResponse(jobEnvelope(importJob("succeeded"))));
  await tick();
  expect(openIncident).toHaveBeenCalledOnce();
});

it("rechecks cancellation eligibility and performs no mutation when completion wins preflight", async () => {
  const { controller, fetch, cancel } = setup("running");
  await tick();
  fetch.mockResolvedValueOnce(
    jsonResponse(jobEnvelope(importJob("succeeded"))),
  );
  controller.cancel();
  await tick();
  expect(cancel).not.toHaveBeenCalled();
  expect(controller.getSnapshot().jobs[importJobID]?.job.status).toBe(
    "succeeded",
  );
});

it("does not let browser clock skew establish terminal job unavailability", async () => {
  const { controller, openIncident } = setup();
  await tick();
  vi.setSystemTime(new Date("2100-01-01T00:00:00Z"));
  expect(openableImport(selected(controller))).toBe(importedIncidentID);
  controller.open();
  await tick();
  expect(openIncident).toHaveBeenCalledOnce();
  vi.setSystemTime(new Date("1900-01-01T00:00:00Z"));
  expect(openableImport(selected(controller))).toBe(importedIncidentID);
  controller.open();
  await tick();
  expect(openIncident).toHaveBeenCalledTimes(2);
});

it("rejects conflicting committed targets while accepting unknown additive references", () => {
  const previous = importJob("succeeded");
  const summary = previous.result_summary;
  const refs = summary?.resource_refs;
  const target = refs?.[0];
  if (!summary || !refs || !target)
    throw new Error("Expected committed target");
  const changed = importJob("succeeded", {
    result_summary: {
      ...summary,
      resource_refs: [{ ...target, id: importJobID }],
    },
  });
  expect(importJobAdvances(previous, changed)).toBe(false);
  const additive = importJob("succeeded", {
    result_summary: {
      ...summary,
      resource_refs: [...refs, { kind: "future_output", id: "opaque" }],
    },
  });
  expect(importJobAdvances(previous, additive)).toBe(true);
});

it("prioritizes a fresh action read over queued refresh and includes queue time in its deadline", async () => {
  const { controller, read, openIncident } = setup();
  await tick();
  const older = deferred<Awaited<ReturnType<typeof readImportJob>>>();
  const preflight = deferred<Awaited<ReturnType<typeof readImportJob>>>();
  read
    .mockReturnValueOnce(older.promise)
    .mockReturnValueOnce(preflight.promise);
  controller.refresh();
  await tick();
  controller.refresh();
  controller.open();
  await tick(20_000);
  expect(read).toHaveBeenCalledTimes(2);
  older.resolve({ kind: "observed", job: importJob("succeeded") });
  await tick();
  expect(read).toHaveBeenCalledTimes(3);
  expect(controller.getSnapshot().action.kind).toBe("checking");
  await tick(10_000);
  expect(controller.getSnapshot().action.kind).toBe("failed");
  expect(openIncident).not.toHaveBeenCalled();
  preflight.resolve({ kind: "observed", job: importJob("succeeded") });
  await tick();
  expect(openIncident).not.toHaveBeenCalled();
});

it("invalidates undispatched intent on departure and ignores late preflight responses", async () => {
  const { controller, read, openIncident } = setup();
  await tick();
  const gate = deferred<Awaited<ReturnType<typeof readImportJob>>>();
  read.mockReturnValueOnce(gate.promise);
  controller.open();
  await tick();
  controller.setActive(false);
  gate.resolve({ kind: "observed", job: importJob("succeeded") });
  await tick();
  expect(openIncident).not.toHaveBeenCalled();
  expect(controller.getSnapshot().action.kind).toBe("idle");
  controller.setActive(true);
  await tick();
  controller.open();
  controller.retire();
  await tick();
  expect(openIncident).not.toHaveBeenCalled();
  expect(controller.getSnapshot().order).toEqual([]);
});

it("retains the exact cancellation attempt after dispatch loss across panel changes", async () => {
  const { controller, cancel } = setup("running");
  await tick();
  const gate = deferred<Awaited<ReturnType<typeof cancelImportJob>>>();
  cancel
    .mockReturnValueOnce(gate.promise)
    .mockResolvedValueOnce({ kind: "acknowledged", job: importJob("running") });
  controller.cancel();
  await tick();
  expect(cancel).toHaveBeenCalledOnce();
  controller.setActive(false);
  await tick(30_000);
  expect(controller.getSnapshot().jobs[importJobID]?.cancellation.kind).toBe(
    "uncertain",
  );
  controller.setActive(true);
  await tick();
  controller.cancel();
  await tick();
  expect(cancel).toHaveBeenCalledTimes(2);
  expect(cancel.mock.calls[1]?.[1]).toBe(cancel.mock.calls[0]?.[1]);
  gate.resolve({ kind: "acknowledged", job: importJob("canceled") });
  await tick();
  expect(controller.getSnapshot().jobs[importJobID]?.job.status).toBe(
    "running",
  );
});

it("keeps an unchanged workbook handoff valid while another refresh is requested", async () => {
  const { controller, openIncident } = setup();
  await tick();
  const gate = deferred<"opened">();
  openIncident.mockReturnValueOnce(gate.promise);
  controller.open();
  await tick();
  controller.refresh();
  await tick();
  const current = openIncident.mock.calls[0]?.[2] as () => boolean;
  expect(current()).toBe(true);
  gate.resolve("opened");
  await tick();
  expect(controller.getSnapshot().navigation.kind).toBe("opened");
});

it("confirms ambiguous job unavailability without inferring global access loss", async () => {
  const { controller, read, confirmAccess, cancel } = setup("running");
  await tick();
  const gate = deferred<{
    kind: "authorized";
    authority: { lifetime: string; actorId: string };
  }>();
  confirmAccess.mockReturnValueOnce(gate.promise);
  read.mockResolvedValueOnce({ kind: "unavailable" });
  controller.refresh();
  await tick();
  expect(controller.getSnapshot().access).toBe("checking");
  expect(controller.getSnapshot().order).toEqual([importJobID]);
  controller.cancel();
  controller.selectFile(new File(["different"], "different.tar"));
  expect(cancel).not.toHaveBeenCalled();
  expect(controller.getSnapshot().selectedFile).toBeNull();
  gate.resolve({
    kind: "authorized",
    authority: { lifetime: "session-one", actorId: importActorID },
  });
  await tick();
  expect(controller.getSnapshot().access).toBe("ready");
  expect(controller.getSnapshot().jobs[importJobID]?.observation.kind).toBe(
    "unavailable",
  );
});

it("clears protected imports before confirmed loss and fences obsolete access confirmation", async () => {
  const { controller, read, confirmAccess } = setup();
  await tick();
  const gate = deferred<{
    kind: "authorized";
    authority: { lifetime: string; actorId: string };
  }>();
  confirmAccess.mockReturnValueOnce(gate.promise);
  read.mockResolvedValueOnce({ kind: "access_failed", status: 403 });
  controller.refresh();
  await tick();
  expect(controller.getSnapshot().order).toEqual([]);
  expect(controller.getSnapshot().access).toBe("checking");
  controller.setAuthority({ lifetime: "replacement", actorId: importActorID });
  gate.resolve({
    kind: "authorized",
    authority: { lifetime: "session-one", actorId: importActorID },
  });
  await tick();
  expect(controller.getSnapshot().order).toEqual([]);
  expect(controller.getSnapshot().access).toBe("ready");
  controller.selectFile(new File(["new"], "new.tar"));
  controller.submit();
  await tick();
  confirmAccess.mockResolvedValueOnce({ kind: "access_lost" });
  read.mockResolvedValueOnce({ kind: "unavailable" });
  controller.refresh();
  await tick();
  expect(controller.getSnapshot().order).toEqual([]);
  expect(controller.getSnapshot().selectedFile).toBeNull();
  expect(controller.getSnapshot().access).toBe("lost");
});
