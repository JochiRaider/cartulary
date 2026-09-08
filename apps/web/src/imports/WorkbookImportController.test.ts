import { afterEach, describe, expect, it, vi } from "vitest";
import {
  type ImportReadResult,
  type ImportWriteResult,
  importInterruptedFailure,
} from "../services/importClient";
import type {
  DiscoveredImportUnit,
  ImportJobResource,
} from "../services/importContractAdapter";
import {
  importTestIds as ids,
  importTestApprovedUnit,
  importTestJob,
  importTestMapping,
  importTestPreview,
  importTestScope,
  importTestSession,
  importTestUnit,
} from "../testing/workbookImportTestSupport";
import type { ImportWriteAttempt } from "./importRequests";
import {
  type WorkbookImportBinding,
  WorkbookImportController,
  type WorkbookImportPort,
} from "./WorkbookImportController";
import {
  createWorkbookMappingDraft,
  workbookMappingErrors,
  workbookMappingRequest,
} from "./workbookImportMapping";
import { importApplyBlocker, importOutcomeViews } from "./workbookImportState";

const received = <T>(value: T): ImportReadResult<T> => ({
  kind: "received",
  value,
});
const interrupted = {
  kind: "failed",
  failure: importInterruptedFailure(),
} as const;
const deferred = <T>() => {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
};
const settle = async () => {
  for (let i = 0; i < 60; i++) await Promise.resolve();
};
const controllers: WorkbookImportController[] = [];
afterEach(() => {
  for (const c of controllers.splice(0)) c.dispose();
  vi.useRealTimers();
});

function harness() {
  let session = importTestSession();
  let units = [importTestUnit()];
  let job = importTestJob("succeeded");
  let txn = 0;
  const send = vi.fn(
    async (attempt: ImportWriteAttempt): Promise<ImportWriteResult> => {
      if (attempt.kind === "upload")
        return {
          kind: "accepted",
          receipt: { kind: "job", job: importTestJob() },
        };
      if (attempt.kind === "mapping") {
        const unit = importTestApprovedUnit();
        units = [unit];
        return { kind: "accepted", receipt: { kind: "unit", unit } };
      }
      if (attempt.kind === "select" || attempt.kind === "skip") {
        const selected = attempt.kind === "select";
        session = {
          ...session,
          session_status: selected ? "ready_to_apply" : "mapped",
          selected_unit_ids: selected ? [attempt.unitId] : [],
        };
        const unit = importTestApprovedUnit({
          unit_status: selected ? "ready" : "skipped",
        });
        units = [unit];
        return {
          kind: "accepted",
          receipt: {
            kind: "selection",
            selection: {
              import_session_id: ids.session,
              session_status: session.session_status,
              selected_unit_ids: session.selected_unit_ids,
              unit,
            },
          },
        };
      }
      if (attempt.kind === "region") {
        const unit = importTestUnit({
          import_unit_id: ids.secondUnit,
          locator_kind: "operator_region",
          locator: { sheet_name: "Sheet1" },
        });
        units = [...units, unit];
        return { kind: "accepted", receipt: { kind: "unit", unit } };
      }
      if (attempt.kind === "cancel") {
        job = importTestJob("cancel_requested", {
          job_id: job.job_id,
          status_route: job.status_route,
          cancelable: false,
        });
        return { kind: "accepted", receipt: { kind: "job", job } };
      }
      job = importTestJob("running", {
        job_id: "00000000-0000-4000-8000-000000000007",
        status_route: "/api/v1/jobs/00000000-0000-4000-8000-000000000007",
      });
      return { kind: "accepted", receipt: { kind: "job", job } };
    },
  );
  const client = {
    send,
    readJob: vi.fn(async () => received(job)),
    readSession: vi.fn(async () => received(session)),
    listUnits: vi.fn(async () => received(units)),
    readUnit: vi.fn(async () => received(units[0] as DiscoveredImportUnit)),
    preview: vi.fn(async (unit: DiscoveredImportUnit) =>
      received(importTestPreview(unit)),
    ),
  } satisfies WorkbookImportPort;
  const controller = new WorkbookImportController({
    transactionId: () => `txn-${++txn}`,
    clock: {
      now: () => Date.now(),
      schedule: (fn, ms) => {
        const id = setTimeout(fn, ms);
        return () => clearTimeout(id);
      },
    },
  });
  controllers.push(controller);
  let binding: WorkbookImportBinding = {
    scope: importTestScope,
    role: "editor",
    closed: false,
    available: true,
    client,
    current: () => true,
    accessFailure: vi.fn(),
  };
  controller.bind(binding);
  controller.setPresented(true);
  return {
    controller,
    client,
    setSession: (value: typeof session) => {
      session = value;
    },
    setUnits: (value: typeof units) => {
      units = value;
    },
    setJob: (value: ImportJobResource) => {
      job = value;
    },
    bind: (patch: Partial<WorkbookImportBinding>) => {
      binding = { ...binding, ...patch };
      controller.bind(binding);
    },
    async discover() {
      controller.chooseFile(
        new File(["Activity Synopsis\nObservation"], "source.csv"),
      );
      await controller.upload();
      await settle();
    },
  };
}

describe("Workbook import lifecycle", () => {
  it("fences late writes at every command stage across role closure claim and incident boundaries", async () => {
    for (const stage of [
      "upload",
      "mapping",
      "select",
      "region",
      "apply",
      "cancel",
    ] as const) {
      for (const boundary of ["role", "closed", "claim", "incident"] as const) {
        const h = harness();
        if (stage !== "upload") {
          if (stage === "region")
            h.setUnits([
              importTestUnit({
                locator_kind: "xlsx_used_range",
                locator: { sheet_name: "Sheet1" },
              }),
            ]);
          await h.discover();
        }
        if (["select", "apply", "cancel"].includes(stage))
          await h.controller.approve(ids.unit, stage !== "select");
        if (stage === "cancel") {
          await h.controller.apply();
          await settle();
          h.controller.stopObservation();
        }
        const pending = deferred<ImportWriteResult>();
        h.client.send.mockReturnValueOnce(pending.promise);
        if (stage === "upload")
          h.controller.chooseFile(new File(["source"], "source.csv"));
        const operation =
          stage === "upload"
            ? h.controller.upload()
            : stage === "mapping"
              ? h.controller.approve(ids.unit)
              : stage === "select"
                ? h.controller.select(ids.unit, true)
                : stage === "region"
                  ? h.controller.createRegion(ids.unit, {
                      startRow: 1,
                      endRow: 2,
                      startColumn: 1,
                      endColumn: 1,
                    })
                  : stage === "apply"
                    ? h.controller.apply()
                    : h.controller.cancel();
        await settle();
        const admitted = h.client.send.mock.calls.length;
        expect(
          h.controller.getSnapshot()[
            stage === "cancel" ? "cancellation" : "operation"
          ]?.phase,
        ).toBe("pending");
        h.bind(
          boundary === "role"
            ? { role: "viewer" }
            : boundary === "closed"
              ? { closed: true }
              : boundary === "claim"
                ? { available: false }
                : { scope: { ...importTestScope, incidentId: ids.secondUnit } },
        );
        pending.resolve({
          kind: "accepted",
          receipt: { kind: "unit", unit: importTestApprovedUnit() },
        });
        await operation;
        await settle();
        expect(h.client.send).toHaveBeenCalledTimes(admitted);
        expect(
          h.controller.getSnapshot()[
            stage === "cancel" ? "cancellation" : "operation"
          ]?.phase,
        ).not.toBe("pending");
        if (boundary === "claim" || boundary === "incident") {
          expect(h.controller.getSnapshot().file).toBeNull();
          expect(h.controller.getSnapshot().units).toEqual([]);
        } else expect(h.controller.getSnapshot().canWrite).toBe(false);
        h.controller.dispose();
      }
    }
  });
  it("keeps header suggestions advisory and associates duplicate destinations with both columns", () => {
    const unit = importTestUnit({
      inferred_column_count: 2,
      source_rect_a1: "A1:B2",
    });
    const preview = {
      ...importTestPreview(unit),
      columns: [
        { source_column_ordinal: 1, source_header_text: "Activity Synopsis" },
        { source_column_ordinal: 2, source_header_text: "Activity Synopsis" },
      ],
    };
    const draft = createWorkbookMappingDraft(preview, unit);
    expect(draft.dirty).toBe(true);
    expect(Object.values(draft.fields).filter(Boolean)).toHaveLength(1);
    const duplicate = {
      ...draft,
      fields: {
        1: "timeline.activity_synopsis_text",
        2: "timeline.activity_synopsis_text",
      },
    };
    expect(Object.keys(workbookMappingErrors(duplicate, preview))).toEqual([
      "1",
      "2",
    ]);
    expect(workbookMappingRequest(unit, preview, duplicate, "txn")).toBeNull();
    const retained = importTestApprovedUnit({
      approved_mapping: {
        ...importTestMapping,
        source_columns: [
          {
            ...importTestMapping.source_columns[0],
            empty_value_policy: "write_null",
          },
        ],
      },
    });
    expect(
      workbookMappingRequest(
        retained,
        importTestPreview(),
        createWorkbookMappingDraft(importTestPreview(), retained),
        "txn",
      ),
    ).toEqual(
      expect.objectContaining({
        source_columns: [
          expect.objectContaining({ empty_value_policy: "write_null" }),
        ],
      }),
    );
  });
  it("retains approval after selection failure and recovers only selection across skip and reselect", async () => {
    const h = harness();
    await h.discover();
    const original = h.client.send.getMockImplementation();
    h.client.send.mockImplementation(async (attempt) =>
      attempt.kind === "select"
        ? {
            kind: "rejected",
            failure: {
              ...importInterruptedFailure(),
              kind: "public",
              status: 409,
              reason: "selection_conflict",
            },
          }
        : (original as NonNullable<typeof original>)(attempt),
    );
    await h.controller.approve(ids.unit);
    expect(h.controller.getSnapshot().units[0]?.unit.mapping_fingerprint).toBe(
      "a".repeat(64),
    );
    expect(h.controller.getSnapshot().session?.selected_unit_ids).toEqual([]);
    expect(h.controller.getSnapshot().operation?.attempt.kind).toBe("select");
    h.client.send.mockImplementation(original as NonNullable<typeof original>);
    await h.controller.retryWrite();
    await h.controller.select(ids.unit, false);
    await h.controller.select(ids.unit, true);
    expect(h.client.send.mock.calls.map(([a]) => a.kind)).toEqual([
      "upload",
      "mapping",
      "select",
      "select",
      "skip",
      "select",
    ]);
    expect(h.controller.getSnapshot().session?.selected_unit_ids).toEqual([
      ids.unit,
    ]);
  });

  it("admits once synchronously and retains the exact uncertain upload for replay", async () => {
    vi.useFakeTimers();
    const h = harness();
    const pending = deferred<ImportWriteResult>();
    h.client.send.mockReturnValueOnce(pending.promise);
    h.controller.chooseFile(new File(["original"], "original.csv"));
    const first = h.controller.upload();
    void h.controller.upload();
    h.controller.chooseFile(new File(["replacement"], "replacement.csv"));
    await vi.advanceTimersByTimeAsync(120_001);
    await first;
    const attempt = h.controller.getSnapshot().operation?.attempt;
    expect(attempt?.kind).toBe("upload");
    expect(h.client.send).toHaveBeenCalledTimes(1);
    expect(h.controller.getSnapshot().file?.name).toBe("original.csv");
    await h.controller.retryWrite();
    await settle();
    expect(h.client.send.mock.calls[1]?.[0]).toBe(attempt);
    expect(h.controller.getSnapshot().session?.import_session_id).toBe(
      ids.session,
    );
    await h.controller.approve(ids.unit);
    h.client.send.mockResolvedValueOnce({
      kind: "uncertain",
      failure: importInterruptedFailure(),
    });
    await h.controller.apply();
    const applyAttempt = h.controller.getSnapshot().operation?.attempt;
    expect(applyAttempt?.kind).toBe("apply");
    await h.controller.select(ids.unit, false);
    expect(h.controller.getSnapshot().session?.selected_unit_ids).toEqual([
      ids.unit,
    ]);
    await h.controller.retryWrite();
    const applies = h.client.send.mock.calls.filter(
      ([request]) => request.kind === "apply",
    );
    expect(applies).toHaveLength(2);
    expect(applies[0]?.[0]).toBe(applies[1]?.[0]);
  });

  it("reconciles current late acknowledgement and fences a replaced lifetime", async () => {
    vi.useFakeTimers();
    const h = harness();
    const pending = deferred<ImportWriteResult>();
    h.client.send.mockReturnValueOnce(pending.promise);
    h.controller.chooseFile(new File(["bytes"], "source.csv"));
    const first = h.controller.upload();
    await vi.advanceTimersByTimeAsync(120_001);
    await first;
    expect(h.controller.getSnapshot().operation?.phase).toBe("uncertain");
    pending.resolve({
      kind: "accepted",
      receipt: { kind: "job", job: importTestJob() },
    });
    await settle();
    expect(h.controller.getSnapshot().session?.import_session_id).toBe(
      ids.session,
    );
    const late = deferred<ImportWriteResult>();
    h.client.send.mockReturnValueOnce(late.promise);
    const approve = h.controller.approve(ids.unit);
    h.bind({ scope: { ...importTestScope, lifetime: "replacement" } });
    late.resolve({
      kind: "accepted",
      receipt: { kind: "unit", unit: importTestApprovedUnit() },
    });
    await approve;
    await settle();
    expect(h.controller.getSnapshot().units).toEqual([]);
    expect(h.controller.getSnapshot().file).toBeNull();
    expect(
      h.client.send.mock.calls.filter(([a]) => a.kind === "select"),
    ).toHaveLength(0);
  });

  it("retains accepted discovery through read failure and resumes the same job after drawer closure", async () => {
    const h = harness();
    h.client.readJob.mockResolvedValueOnce(interrupted);
    await h.discover();
    const jobId = h.controller.getSnapshot().job?.job_id;
    expect(h.controller.getSnapshot().observationFailure).not.toBeNull();
    h.controller.setPresented(false);
    await h.controller.resumeObservation();
    await settle();
    h.controller.setPresented(true);
    expect(h.controller.getSnapshot().job?.job_id).toBe(jobId);
    expect(h.controller.getSnapshot().session?.import_session_id).toBe(
      ids.session,
    );
    expect(h.client.send).toHaveBeenCalledTimes(1);
  });

  it("keeps empty discovery and preview failure distinct and bounds previews to one request", async () => {
    const h = harness();
    h.setUnits([]);
    await h.discover();
    expect(h.controller.getSnapshot().units).toEqual([]);
    expect(h.controller.getSnapshot().loadFailure).toBeNull();
    h.setUnits([
      importTestUnit(),
      importTestUnit({ import_unit_id: ids.secondUnit }),
    ]);
    const pending =
      deferred<ImportReadResult<ReturnType<typeof importTestPreview>>>();
    h.client.preview.mockReturnValueOnce(pending.promise);
    await h.controller.refresh();
    h.controller.loadPreview(ids.secondUnit);
    expect(h.client.preview).toHaveBeenCalledTimes(1);
    pending.resolve(interrupted);
    await settle();
    expect(h.controller.getSnapshot().units[0]?.previewFailure).not.toBeNull();
    expect(h.controller.getSnapshot().units[1]?.preview).not.toBeNull();
    h.controller.loadPreview(ids.unit);
    await settle();
    expect(h.controller.getSnapshot().units[0]?.previewFailure).toBeNull();
  });

  it("creates only the returned durable region without approving or selecting it", async () => {
    const h = harness();
    h.setUnits([
      importTestUnit({
        locator_kind: "xlsx_used_range",
        locator: { sheet_name: "Sheet1" },
      }),
    ]);
    await h.discover();
    await h.controller.createRegion(ids.unit, {
      startRow: 1,
      endRow: 2,
      startColumn: 1,
      endColumn: 1,
    });
    await settle();
    const region = h.controller
      .getSnapshot()
      .units.find((u) => u.unit.import_unit_id === ids.secondUnit);
    expect(region?.unit.locator_kind).toBe("operator_region");
    expect(region?.unit.approved_mapping).toBeUndefined();
    expect(h.controller.getSnapshot().session?.selected_unit_ids).toEqual([]);
  });

  it("reconciles changed selection before apply and blocks overlap and terminal sessions", async () => {
    const h = harness();
    await h.discover();
    await h.controller.approve(ids.unit);
    h.setSession(
      importTestSession({ session_status: "mapped", selected_unit_ids: [] }),
    );
    await h.controller.apply();
    expect(h.client.send.mock.calls.some(([a]) => a.kind === "apply")).toBe(
      false,
    );
    expect(h.controller.getSnapshot().message).toContain("selection changed");
    h.setUnits([
      importTestApprovedUnit(),
      importTestApprovedUnit({ import_unit_id: ids.secondUnit }),
    ]);
    h.setSession(
      importTestSession({
        session_status: "ready_to_apply",
        selected_unit_ids: [ids.unit, ids.secondUnit],
      }),
    );
    await h.controller.refresh();
    expect(importApplyBlocker(h.controller.getSnapshot())).toContain("overlap");
    h.setSession(importTestSession({ source_content_sha256: "c".repeat(64) }));
    expect(await h.controller.refresh()).toBe(false);
    expect(h.controller.getSnapshot().loadFailure?.kind).toBe("contract");
    expect(h.controller.getSnapshot().session?.selected_unit_ids).toEqual([
      ids.unit,
      ids.secondUnit,
    ]);
    h.setSession(importTestSession());
    h.setUnits([
      importTestApprovedUnit({ locator: { file: "contradictory-source" } }),
    ]);
    expect(await h.controller.refresh()).toBe(false);
    h.setUnits([
      importTestApprovedUnit(),
      importTestApprovedUnit({ import_unit_id: ids.secondUnit }),
    ]);
    h.setSession(importTestSession({ session_status: "applied" }));
    await h.controller.refresh();
    await h.controller.select(ids.unit, true);
    expect(h.client.send.mock.calls.map(([a]) => a.kind)).toEqual([
      "upload",
      "mapping",
      "select",
    ]);
  });

  it("fences stale resource loads after a mutation and gates every write on current authority", async () => {
    const h = harness();
    await h.discover();
    const stale =
      deferred<ImportReadResult<ReturnType<typeof importTestSession>>>();
    h.client.readSession.mockReturnValueOnce(stale.promise);
    const refresh = h.controller.refresh();
    await h.controller.approve(ids.unit);
    stale.resolve(received(importTestSession()));
    await refresh;
    expect(h.controller.getSnapshot().session?.selected_unit_ids).toEqual([
      ids.unit,
    ]);
    await h.controller.select(ids.unit, false);
    h.controller.updateMapping(ids.unit, {
      ordinal: 1,
      fieldKey: "timeline.activity_synopsis_text",
    });
    const draft = h.controller.getSnapshot().units[0]?.draft;
    const count = h.client.send.mock.calls.length;
    h.bind({ role: "viewer" });
    await h.controller.approve(ids.unit);
    await h.controller.select(ids.unit, true);
    await h.controller.apply();
    await h.controller.createRegion(ids.unit, {
      startRow: 1,
      endRow: 2,
      startColumn: 1,
      endColumn: 1,
    });
    expect(h.client.send).toHaveBeenCalledTimes(count);
    expect(h.controller.getSnapshot().units[0]?.draft).toEqual(draft);
    h.bind({ role: "editor", closed: true });
    await h.controller.approve(ids.unit);
    expect(h.client.send).toHaveBeenCalledTimes(count);
    h.bind({ available: false });
    expect(h.controller.getSnapshot().units).toEqual([]);
    const local = harness();
    await local.discover();
    const unavailable = {
      ...importInterruptedFailure(),
      kind: "public" as const,
      status: 404,
      code: "import_unit_not_found",
    };
    local.client.send.mockResolvedValueOnce({
      kind: "rejected",
      failure: unavailable,
    });
    await local.controller.approve(ids.unit);
    expect(local.controller.getSnapshot().units).toEqual([]);
    expect(local.controller.getSnapshot().operation).toBeNull();
    expect(local.controller.getSnapshot().access).toBe("active");
    expect(local.controller.getSnapshot().file).not.toBeNull();
    local.client.readSession.mockResolvedValueOnce({
      kind: "failed",
      failure: { ...unavailable, code: "import_session_not_found" },
    });
    await local.controller.refresh();
    expect(local.controller.getSnapshot().file).toBeNull();
    expect(local.controller.getSnapshot().access).toBe("active");
  });

  it("pauses sensitive presentation on unresolved authentication without replaying restored writes", async () => {
    const h = harness();
    await h.discover();
    const pending = deferred<ImportWriteResult>();
    h.client.send.mockReturnValueOnce(pending.promise);
    const approve = h.controller.approve(ids.unit);
    h.controller.pause();
    expect(h.controller.getSnapshot().units).toEqual([]);
    pending.resolve({
      kind: "accepted",
      receipt: { kind: "unit", unit: importTestApprovedUnit() },
    });
    await approve;
    h.bind({});
    await settle();
    expect(h.controller.getSnapshot().operation?.phase).toBe("uncertain");
    expect(h.client.send).toHaveBeenCalledTimes(2);
  });

  it("requires current cancelability, admits cancellation once, and handles completion races", async () => {
    const h = harness();
    h.setJob(importTestJob("running"));
    await h.discover();
    h.controller.stopObservation();
    const pending = deferred<ImportReadResult<ImportJobResource>>();
    h.client.readJob.mockReturnValueOnce(pending.promise);
    const cancel = h.controller.cancel();
    void h.controller.cancel();
    expect(h.controller.getSnapshot().cancellation?.phase).toBe("pending");
    pending.resolve(received(importTestJob("succeeded")));
    await cancel;
    await settle();
    expect(
      h.client.send.mock.calls.filter(([a]) => a.kind === "cancel"),
    ).toHaveLength(0);
    const second = harness();
    second.setJob(importTestJob("running", { cancelable: false }));
    await second.discover();
    await second.controller.cancel();
    expect(second.client.send).toHaveBeenCalledTimes(1);
    const replay = harness();
    replay.setJob(importTestJob("running"));
    await replay.discover();
    replay.controller.stopObservation();
    replay.client.send.mockResolvedValueOnce({
      kind: "uncertain",
      failure: importInterruptedFailure(),
    });
    await replay.controller.cancel();
    replay.controller.stopObservation();
    expect(replay.controller.getSnapshot().cancellation?.phase).toBe(
      "uncertain",
    );
    await replay.controller.cancel();
    const cancels = replay.client.send.mock.calls.filter(
      ([attempt]) => attempt.kind === "cancel",
    );
    expect(cancels).toHaveLength(2);
    expect(cancels[0]?.[0]).toBe(cancels[1]?.[0]);
  });

  it("separates cancel requested from canceled and retains committed partial outcomes and navigation", async () => {
    const h = harness();
    await h.discover();
    await h.controller.approve(ids.unit);
    await h.controller.apply();
    await settle();
    h.controller.stopObservation();
    h.bind({ closed: true });
    await settle();
    h.controller.stopObservation();
    await h.controller.cancel();
    await settle();
    expect(h.controller.getSnapshot().job?.status).toBe("cancel_requested");
    expect(h.controller.canCancel()).toBe(false);
    const applied = importTestApprovedUnit({ unit_status: "applied" });
    h.setUnits([
      applied,
      importTestUnit({ import_unit_id: ids.secondUnit, unit_status: "failed" }),
    ]);
    h.setSession(
      importTestSession({
        session_status: "partially_applied",
        selected_unit_ids: [ids.unit, ids.secondUnit],
      }),
    );
    const currentJob = h.controller.getSnapshot().job as ImportJobResource;
    h.setJob(
      importTestJob("canceled", {
        job_id: currentJob.job_id,
        status_route: currentJob.status_route,
      }),
    );
    await h.controller.resumeObservation();
    expect(h.controller.getSnapshot().job?.status).toBe("canceled");
    expect(
      importOutcomeViews(h.controller.getSnapshot()).map((v) => v.id),
    ).toEqual(["cartulary.view.timeline.v2"]);
    const navigate = vi.fn();
    await h.controller.openResult("cartulary.view.timeline.v2", navigate);
    expect(navigate).toHaveBeenCalledWith("cartulary.view.timeline.v2");
    h.controller.setPresented(false);
    await h.controller.openResult("cartulary.view.timeline.v2", navigate);
    expect(navigate).toHaveBeenCalledTimes(1);
  });
});
