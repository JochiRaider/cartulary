import { afterEach, describe, expect, it, vi } from "vitest";
import { importTiming } from "../imports/importCoordinator";
import type { ImportWriteAttempt } from "../imports/importRequests";
import {
  type ImportReadResult,
  type ImportWriteResult,
  importInterruptedFailure,
} from "../services/importClient";
import type {
  DiscoveredImportUnit,
  ExtensionMappingPreviewResource,
  ImportJobResource,
  ImportSelectionReceipt,
} from "../services/importContractAdapter";
import {
  decodeNetworkFlowImportPreviewResult,
  networkFlowMappingMetadata,
} from "../services/networkFlowContractAdapter";
import {
  importTestJob,
  importTestScope,
  importTestSession,
  importTestUnit,
} from "../testing/workbookImportTestSupport";
import {
  type NetworkFlowImportBinding,
  NetworkFlowImportController,
  type NetworkFlowImportPort,
} from "./NetworkFlowImportController";
import {
  buildNetworkFlowMappingCandidate,
  createNetworkFlowMappingDraft,
  networkFlowApprovalRequest,
} from "./networkFlowImportModel";
import { networkFlowImportMappingState } from "./networkFlowImportState";

const received = <T>(value: T): ImportReadResult<T> => ({
  kind: "received",
  value,
});
const interrupted = {
  kind: "failed",
  failure: importInterruptedFailure(),
} as const;
const rejected = {
  kind: "rejected",
  failure: {
    kind: "public",
    status: 409,
    code: "import_apply_blocked",
    reason: "unit_not_ready",
    field: null,
    retryable: false,
  },
} as const;
const fingerprint = "b".repeat(64);
const deferred = <T>() => {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
};
const controllers: NetworkFlowImportController[] = [];
const tableId = "00000000-0000-4000-8000-000000000009";
function applyJob(status: ImportJobResource["status"] = "queued") {
  const id = "00000000-0000-4000-8000-000000000008";
  const base = importTestJob(status);
  return {
    ...base,
    job_id: id,
    status_route: `/api/v1/jobs/${id}`,
    result_summary:
      status === "succeeded"
        ? {
            code: "import_session_applied",
            message: "Applied",
            resource_refs: [
              ...(base.result_summary?.resource_refs ?? []),
              {
                kind: "network_flow_table",
                id: tableId,
                route: `/api/v1/incidents/${importTestScope.incidentId}/network-flow/tables/${tableId}`,
              },
            ],
          }
        : base.result_summary,
  };
}
afterEach(() => {
  for (const controller of controllers.splice(0)) controller.retire();
  vi.useRealTimers();
});

function harness() {
  const session = importTestSession();
  const unit = importTestUnit({
    inferred_column_count: 9,
    source_rect_a1: "A1:I2",
  });
  const columns = networkFlowMappingMetadata.source_profiles[0].fields
    .filter((f) => f.requirement === "required")
    .map((f, index) => ({
      source_column_ordinal: index + 1,
      source_header_text: f.aliases[0],
    }));
  const source = { ...unit, columns, preview_rows: [], truncated: false };
  const discovery = {
    sessionId: session.import_session_id,
    session,
    unit,
    preview: source,
  };
  const candidate = buildNetworkFlowMappingCandidate(
    createNetworkFlowMappingDraft(columns),
    columns,
  );
  const sourceColumns = columns.map((c) => ({
    source_column_ordinal: c.source_column_ordinal,
    raw_header_text: c.source_header_text,
    normalized_header_for_suggestion: c.source_header_text.toLowerCase(),
    raw_header_sha256: "a".repeat(64),
    sample_values: [{ safe_sample: null, raw_value_sha256: "a".repeat(64) }],
    detected_empty_count: 0,
  }));
  const ownerPreview = decodeNetworkFlowImportPreviewResult({
    schema_id: "cartulary.network_flow.import_preview_result.v1",
    source_content_sha256: session.source_content_sha256,
    source_columns: sourceColumns,
    materialized_mapping: { ...candidate, source_columns: sourceColumns },
    mapping_fingerprint: fingerprint,
    preview_record_count: 1,
    preview_accepted_count: 1,
    preview_rejected_count: 0,
    diagnostics: [],
    diagnostics_truncated: false,
  });
  const wrapper: ExtensionMappingPreviewResource<unknown> = {
    schema_id: "cartulary.imports.extension_mapping_preview_result.v1",
    import_session_id: session.import_session_id,
    import_unit_id: unit.import_unit_id,
    target_kind: "network_flow_table",
    extension_profile_id: "network_flow_activity",
    owner_result_schema_id: "cartulary.network_flow.import_preview_result.v1",
    owner_result: ownerPreview,
  };
  const request = networkFlowApprovalRequest(
    discovery,
    {
      target_kind: "network_flow_table",
      extension_profile_id: "network_flow_activity",
      owner_mapping_schema_id: "cartulary.network_flow.mapping_candidate.v1",
      owner_mapping: { ...candidate },
    },
    "mapping-fixture",
  );
  const approved: DiscoveredImportUnit = {
    ...unit,
    unit_status: "ready",
    mapping_fingerprint: fingerprint,
    approved_mapping: {
      target_kind: "network_flow_table",
      extension_profile_id: "network_flow_activity",
      owner_mapping_schema_id: "cartulary.network_flow.mapping_candidate.v1",
      owner_mapping: { ...candidate },
      source_columns: [...request.source_columns],
    },
  };
  const selection: ImportSelectionReceipt = {
    import_session_id: session.import_session_id,
    session_status: "ready_to_apply",
    selected_unit_ids: [unit.import_unit_id],
    unit: approved,
  };
  const client = {
    send: vi.fn(
      async (
        attempt: ImportWriteAttempt,
        _signal?: AbortSignal,
        _replay?: boolean,
      ): Promise<ImportWriteResult> =>
        attempt.kind === "upload" || attempt.kind === "apply"
          ? {
              kind: "accepted",
              receipt: {
                kind: "job",
                job: attempt.kind === "apply" ? applyJob() : importTestJob(),
              },
            }
          : attempt.kind === "mapping"
            ? { kind: "accepted", receipt: { kind: "unit", unit: approved } }
            : { kind: "accepted", receipt: { kind: "selection", selection } },
    ),
    readJob: vi.fn(
      async (id: string): Promise<ImportReadResult<ImportJobResource>> =>
        received(
          id === applyJob().job_id
            ? applyJob("succeeded")
            : importTestJob("succeeded"),
        ),
    ),
    readSession: vi.fn(
      async (): Promise<ImportReadResult<typeof session>> => received(session),
    ),
    listUnits: vi.fn(
      async (): Promise<ImportReadResult<DiscoveredImportUnit[]>> =>
        received([unit]),
    ),
    readUnit: vi.fn(async () => received(approved)),
    preview: vi.fn(
      async (): Promise<ImportReadResult<typeof source>> => received(source),
    ),
    previewMapping: vi.fn(async () => received(wrapper)),
  } satisfies NetworkFlowImportPort;
  let txn = 0;
  const controller = new NetworkFlowImportController({
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
  let binding: NetworkFlowImportBinding = {
    scope: importTestScope,
    role: "editor",
    closed: false,
    available: true,
    current: () => true,
    accessFailure: vi.fn(),
    client,
  };
  controller.bind(binding);
  controller.setWorkspaceActive(true);
  const upload = () =>
    controller.upload(new File(["source"], "flows.csv", { type: "text/csv" }));
  return {
    controller,
    client,
    upload,
    approved,
    wrapper,
    ownerPreview,
    session,
    unit,
    binding: (change: Partial<NetworkFlowImportBinding>) => {
      binding = { ...binding, ...change };
      controller.bind(binding);
    },
    getBinding: () => binding,
  };
}

describe("Network Flow import ownership", () => {
  it("admits one upload synchronously and keeps preview separate from durable approval", async () => {
    const h = harness();
    await Promise.all([h.upload(), h.upload()]);
    expect(h.client.send).toHaveBeenCalledTimes(1);
    await h.controller.requestPreview();
    expect(h.client.send).toHaveBeenCalledTimes(1);
    expect(h.controller.canContinue()).toBe(true);
    expect(h.controller.getSnapshot().approval).toBeNull();
    await h.controller.approve();
    expect(h.controller.getSnapshot().approval?.fingerprint).toBe(fingerprint);
    expect(h.controller.getSnapshot().selection).toBeNull();
    const attempt = h.client.send.mock.calls[1]?.[0];
    expect(attempt?.kind).toBe("mapping");
    if (attempt?.kind === "mapping")
      expect(
        attempt.body.source_columns.map((c) => c.source_column_ordinal),
      ).toEqual([1, 2, 3, 4, 5, 6, 7, 8, 9]);
    const rejectedUpload = harness();
    rejectedUpload.client.send.mockResolvedValueOnce(rejected);
    await rejectedUpload.upload();
    expect(rejectedUpload.controller.canDiscardDraft()).toBe(true);
    rejectedUpload.controller.discardDraft();
    expect(rejectedUpload.controller.getSnapshot().stage).toBe("idle");
  });
  it("discards stale preview success and failure after a candidate changes", async () => {
    for (const departure of [false, true]) {
      for (const lateFailure of [false, true]) {
        const h = harness();
        await h.upload();
        const pending = deferred<ImportReadResult<typeof h.wrapper>>();
        h.client.previewMapping.mockImplementationOnce(() => pending.promise);
        const first = h.controller.requestPreview();
        const draft = h.controller.getSnapshot().draft;
        if (!draft) throw new Error("Missing draft");
        if (departure) {
          h.controller.setWorkspaceActive(false);
          h.controller.setWorkspaceActive(true);
        } else {
          h.controller.updateDraft({
            ...draft,
            displayNameOverride: "new intent",
          });
        }
        await h.controller.requestPreview();
        const current = h.controller.getSnapshot().preview;
        pending.resolve(lateFailure ? interrupted : received(h.wrapper));
        await first;
        expect(h.controller.getSnapshot().preview).toBe(current);
        expect(h.controller.getSnapshot().previewFailure).toBeNull();
      }
    }
  });
  it("invalidates approval applicability on fingerprint mismatch without selecting or applying", async () => {
    const h = harness();
    await h.upload();
    await h.controller.requestPreview();
    h.client.send.mockResolvedValueOnce({
      kind: "accepted",
      receipt: {
        kind: "unit",
        unit: { ...h.approved, mapping_fingerprint: "c".repeat(64) },
      },
    });
    expect(await h.controller.approve()).toBe(false);
    expect(networkFlowImportMappingState(h.controller.getSnapshot())).toBe(
      "validation_preview_pending",
    );
    expect(h.controller.getSnapshot().approval?.fingerprint).toBe(
      "c".repeat(64),
    );
    expect(h.client.send.mock.calls.map(([a]) => a.kind)).toEqual([
      "upload",
      "mapping",
    ]);
  });
  it("preserves copyable drafts on write loss and closure and purges replaced authority", async () => {
    for (const change of [{ role: "viewer" as const }, { closed: true }]) {
      const h = harness();
      await h.upload();
      await h.controller.requestPreview();
      const draft = h.controller.getSnapshot().draft;
      h.binding(change);
      expect(h.controller.getSnapshot().draft).toBe(draft);
      expect(h.controller.canContinue()).toBe(false);
      await h.controller.approve();
      expect(h.client.send).toHaveBeenCalledTimes(1);
      h.binding({ scope: { ...importTestScope, lifetime: "replacement" } });
      expect(h.controller.getSnapshot().draft).toBeNull();
    }
  });
  it("retains the immutable submitted candidate and typed definite rejection", async () => {
    const h = harness();
    await h.upload();
    await h.controller.requestPreview();
    const pending = deferred<ImportWriteResult>();
    h.client.send.mockImplementationOnce(() => pending.promise);
    const approving = h.controller.approve();
    const before = h.controller.getSnapshot().draft;
    if (!before) throw new Error("Missing draft");
    h.controller.updateDraft({ ...before, displayNameOverride: "replacement" });
    expect(h.controller.getSnapshot().draft).toBe(before);
    const request = h.controller.getSnapshot().write?.request;
    expect(Object.isFrozen(request)).toBe(true);
    pending.resolve(rejected);
    await approving;
    expect(h.controller.getSnapshot().write).toMatchObject({
      request,
      disposition: "rejected",
      failure: rejected.failure,
    });
  });
  it("blocks changed profiles and incomplete mappings while keeping preview rejection counts separate", async () => {
    const h = harness();
    await h.upload();
    await h.controller.requestPreview();
    const draft = h.controller.getSnapshot().draft;
    if (!draft) throw new Error("Missing draft");
    h.controller.updateDraft({
      ...draft,
      sourceProfileId: "ipfix_csv_projection_v1",
    });
    expect(networkFlowImportMappingState(h.controller.getSnapshot())).toBe(
      "mapping_required",
    );
    await h.controller.requestPreview();
    expect(h.client.previewMapping).toHaveBeenCalledTimes(1);
    h.controller.updateDraft({
      ...draft,
      columnChoices: { ...draft.columnChoices, 1: null },
    });
    expect(h.controller.canContinue()).toBe(false);
    h.controller.updateDraft(draft);
    h.client.previewMapping.mockResolvedValueOnce(
      received({
        ...h.wrapper,
        owner_result: {
          ...h.ownerPreview,
          preview_accepted_count: 0,
          preview_rejected_count: 1,
        },
      }),
    );
    await h.controller.requestPreview();
    expect(h.controller.canContinue()).toBe(true);
  });
});

describe("Network Flow import recovery", () => {
  it("revalidates retained intent after authority changes before replaying selection", async () => {
    const h = harness();
    await h.upload();
    await h.controller.requestPreview();
    await h.controller.approve();
    h.client.send.mockResolvedValueOnce({
      kind: "uncertain",
      failure: importInterruptedFailure(),
    });
    await h.controller.continueApply();
    const request = h.controller.getSnapshot().write?.request;
    h.binding({ role: "viewer" });
    h.binding({ role: "editor" });
    await h.controller.retryWrite();
    expect(h.client.send).toHaveBeenCalledTimes(3);
    expect(h.controller.canPreview()).toBe(true);
    await h.controller.requestPreview();
    await h.controller.retryWrite();
    expect(h.client.send.mock.calls[3]?.[0]).toBe(request);
    expect(h.controller.getSnapshot().applyJob?.resource.status).toBe(
      "succeeded",
    );
  });
  it("retains resource-specific denials without declaring lost incident access", async () => {
    const h = harness();
    await h.upload();
    await h.controller.requestPreview();
    const failure = {
      kind: "public" as const,
      status: 403,
      code: "authorization_denied",
      reason: null,
      field: "unit_id",
      retryable: false,
    };
    h.client.send.mockResolvedValueOnce({ kind: "rejected", failure });
    await h.controller.approve();
    expect(h.controller.getSnapshot()).toMatchObject({
      access: "active",
      canWrite: false,
      write: { failure },
    });
    expect(h.controller.getSnapshot().draft).not.toBeNull();
    expect(h.getBinding().accessFailure).toHaveBeenCalledWith(failure);
    const closed = harness();
    await closed.upload();
    await closed.controller.requestPreview();
    const closedFailure = { ...failure, status: 409, code: "incident_closed" };
    closed.client.send.mockResolvedValueOnce({
      kind: "rejected",
      failure: closedFailure,
    });
    await closed.controller.approve();
    expect(closed.controller.getSnapshot()).toMatchObject({
      closed: true,
      canWrite: false,
    });
    expect(closed.controller.getSnapshot().draft).not.toBeNull();
    expect(closed.getBinding().accessFailure).toHaveBeenCalledWith(
      closedFailure,
    );
    for (const status of [401, 403, 404]) {
      const handoff = harness();
      await handoff.upload();
      await handoff.controller.requestPreview();
      const tableFailure = { ...failure, status };
      handoff.controller.setHandoff(async () => ({
        kind: "failed",
        failure: tableFailure,
      }));
      await handoff.controller.continueApply();
      expect(handoff.controller.getSnapshot().access).toBe(
        status === 401 ? "paused" : "active",
      );
      if (status === 401) handoff.binding({});
      expect(handoff.controller.getSnapshot().applyJob?.resource.status).toBe(
        "succeeded",
      );
      expect(handoff.controller.getSnapshot().handoff).toMatchObject({
        status: "failed",
        failure: tableFailure,
      });
      if (status === 404)
        expect(handoff.getBinding().accessFailure).not.toHaveBeenCalled();
      else
        expect(handoff.getBinding().accessFailure).toHaveBeenCalledWith(
          tableFailure,
        );
    }
  });
  it("retains terminal job success when a delayed cancellation acknowledgement arrives", async () => {
    vi.useFakeTimers();
    const h = harness();
    h.client.readJob.mockResolvedValueOnce(interrupted);
    await h.upload();
    h.client.readJob.mockResolvedValueOnce(received(importTestJob("running")));
    const pending = deferred<ImportWriteResult>();
    h.client.send.mockImplementationOnce(() => pending.promise);
    const canceling = h.controller.cancel();
    await vi.advanceTimersByTimeAsync(importTiming.request + 1);
    await canceling;
    await h.controller.resumeObservation();
    expect(h.controller.getSnapshot().discoveryJob?.resource.status).toBe(
      "succeeded",
    );
    h.controller.startNew();
    expect(h.controller.getSnapshot().cancellation?.disposition).toBe(
      "uncertain",
    );
    pending.resolve({
      kind: "accepted",
      receipt: { kind: "job", job: importTestJob("cancel_requested") },
    });
    await vi.advanceTimersByTimeAsync(0);
    expect(h.controller.getSnapshot().discoveryJob?.resource.status).toBe(
      "succeeded",
    );
    expect(h.controller.getSnapshot().cancellation?.disposition).toBe(
      "accepted",
    );
  });
  it("continues selection after acknowledged approval and exactly replays uncertain apply", async () => {
    const h = harness();
    await h.upload();
    await h.controller.requestPreview();
    await h.controller.approve();
    h.client.send.mockResolvedValueOnce(rejected);
    await h.controller.continueApply();
    expect(h.controller.getSnapshot().approval?.fingerprint).toBe(fingerprint);
    expect(h.controller.getSnapshot().write?.request.kind).toBe("select");
    const selectAttempt = h.controller.getSnapshot().write?.request;
    const send = h.client.send.getMockImplementation();
    h.client.send.mockImplementation(async (request, signal, replay) =>
      request.kind === "apply" && !replay
        ? { kind: "uncertain", failure: importInterruptedFailure() }
        : send
          ? send(request, signal, replay)
          : rejected,
    );
    await h.controller.retryWrite();
    expect(h.controller.getSnapshot().selection?.unit.mapping_fingerprint).toBe(
      fingerprint,
    );
    const attempt = h.controller.getSnapshot().write?.request;
    expect(attempt?.kind).toBe("apply");
    await Promise.all([
      h.controller.retryWrite(),
      h.controller.retryWrite(),
      h.controller.continueApply(),
    ]);
    const calls = h.client.send.mock.calls;
    expect(calls.map(([a]) => a.kind)).toEqual([
      "upload",
      "mapping",
      "select",
      "select",
      "apply",
      "apply",
    ]);
    expect(calls[3]?.[0]).not.toEqual(selectAttempt);
    expect(calls[5]?.[0]).toBe(attempt);
    expect(calls[5]?.[2]).toBe(true);
    expect(h.controller.getSnapshot().applyJob?.resource.status).toBe(
      "succeeded",
    );
    expect(h.controller.getSnapshot().handoff?.tableId).toBe(tableId);
  });
  it("retains discovery receipts through each source read failure without uploading again", async () => {
    for (const step of ["readSession", "listUnits", "preview"] as const) {
      const h = harness();
      h.client[step].mockResolvedValueOnce(interrupted);
      await h.upload();
      expect(h.controller.getSnapshot().discoveryJob?.resource.status).toBe(
        "succeeded",
      );
      expect(h.controller.getSnapshot().sourceFailure).toEqual(
        interrupted.failure,
      );
      await h.controller.refreshSource();
      expect(h.controller.getSnapshot().draft).not.toBeNull();
      expect(h.client.send).toHaveBeenCalledTimes(1);
    }
  });
  it("resumes known apply jobs and recovers only handoff after confirmed success", async () => {
    const h = harness();
    await h.upload();
    await h.controller.requestPreview();
    h.client.readJob.mockResolvedValueOnce(interrupted);
    const handoff = vi.fn(async () => ({
      kind: "failed" as const,
      failure: importInterruptedFailure(),
    }));
    h.controller.setHandoff(handoff);
    await h.controller.continueApply();
    expect(h.controller.getSnapshot().applyJob).toMatchObject({
      resource: { status: "queued" },
      failure: interrupted.failure,
    });
    await h.controller.resumeObservation();
    expect(h.controller.getSnapshot().applyJob?.resource.status).toBe(
      "succeeded",
    );
    expect(h.controller.getSnapshot().handoff?.status).toBe("failed");
    await h.controller.recoverHandoff();
    expect(handoff).toHaveBeenCalledTimes(2);
    expect(h.client.send.mock.calls.map(([a]) => a.kind)).toEqual([
      "upload",
      "mapping",
      "select",
      "apply",
    ]);
  });
  it("expires bounded observation and pauses workspace departure without canceling server work", async () => {
    vi.useFakeTimers();
    const h = harness();
    h.client.readJob.mockResolvedValue(received(importTestJob("running")));
    const uploading = h.upload();
    await vi.advanceTimersByTimeAsync(importTiming.observation + 1);
    await uploading;
    expect(h.controller.getSnapshot().discoveryJob).toMatchObject({
      resource: { status: "running" },
      observing: false,
    });
    const resuming = h.controller.resumeObservation();
    await vi.advanceTimersByTimeAsync(0);
    h.controller.setPresented(false);
    expect(h.controller.getSnapshot().discoveryJob?.observing).toBe(true);
    h.controller.setWorkspaceActive(false);
    await resuming;
    expect(h.controller.getSnapshot().discoveryJob?.observing).toBe(false);
    h.controller.discardDraft();
    expect(h.controller.getSnapshot().discoveryJob).not.toBeNull();
    h.controller.setWorkspaceActive(true);
    h.client.readJob.mockResolvedValue(received(importTestJob("succeeded")));
    await h.controller.resumeObservation();
    expect(h.controller.getSnapshot().draft).not.toBeNull();
    expect(h.client.send).toHaveBeenCalledTimes(1);
  });
  it("accepts a late receipt only for the same unresolved workflow", async () => {
    for (const retire of [false, true]) {
      vi.useFakeTimers();
      const h = harness();
      const pending = deferred<ImportWriteResult>();
      h.client.send.mockImplementationOnce(() => pending.promise);
      const uploading = h.upload();
      await vi.advanceTimersByTimeAsync(importTiming.upload + 1);
      await uploading;
      expect(h.controller.getSnapshot().write?.disposition).toBe("uncertain");
      if (retire) h.controller.retire();
      pending.resolve({
        kind: "accepted",
        receipt: { kind: "job", job: importTestJob() },
      });
      await vi.advanceTimersByTimeAsync(0);
      expect(
        h.controller.getSnapshot().discoveryJob?.resource.status ?? null,
      ).toBe(retire ? null : "succeeded");
    }
  });
  it("hides uncertain authority and retires incident actor and independent claim replacement", async () => {
    for (const change of [
      { scope: { ...importTestScope, incidentId: "replacement" } },
      { scope: { ...importTestScope, actorId: "replacement" } },
      { available: false },
      { role: "" as const },
    ]) {
      const h = harness();
      await h.upload();
      h.controller.pause();
      expect(h.controller.getSnapshot().draft).toBeNull();
      h.binding({});
      expect(h.controller.getSnapshot().draft).not.toBeNull();
      h.binding(change);
      expect(h.controller.getSnapshot().draft).toBeNull();
    }
  });
  it("replays uncertain cancellation and reconciles cancel requested with completion races", async () => {
    for (const race of [false, true]) {
      const h = harness();
      h.client.readJob.mockResolvedValueOnce(interrupted);
      await h.upload();
      h.client.readJob.mockResolvedValueOnce(
        received(importTestJob("running")),
      );
      h.client.send.mockResolvedValueOnce({
        kind: "uncertain",
        failure: importInterruptedFailure(),
      });
      await Promise.all([h.controller.cancel(), h.controller.cancel()]);
      const attempt = h.controller.getSnapshot().cancellation?.request;
      expect(attempt?.kind).toBe("cancel");
      h.client.readJob.mockResolvedValueOnce(
        received(importTestJob(race ? "succeeded" : "cancel_requested")),
      );
      h.client.send.mockResolvedValueOnce({
        kind: "accepted",
        receipt: {
          kind: "job",
          job: importTestJob(race ? "succeeded" : "cancel_requested"),
        },
      });
      h.client.readJob.mockResolvedValueOnce(
        received(importTestJob(race ? "succeeded" : "canceled")),
      );
      await h.controller.cancel();
      expect(h.client.send.mock.calls[2]?.[0]).toBe(attempt);
      expect(h.client.send.mock.calls[2]?.[2]).toBe(true);
      expect(h.controller.getSnapshot().discoveryJob?.resource.status).toBe(
        race ? "succeeded" : "canceled",
      );
    }
  });
  it("fences delayed handoff selection across workspace departure and authority replacement", async () => {
    for (const leave of [
      () => {},
      (h: ReturnType<typeof harness>) => h.controller.setWorkspaceActive(false),
      (h: ReturnType<typeof harness>) => h.binding({ role: "viewer" }),
    ]) {
      const h = harness();
      await h.upload();
      await h.controller.requestPreview();
      const pending =
        deferred<
          import("./networkFlowImportState").NetworkFlowImportHandoffResult
        >();
      let request:
        | import("./networkFlowImportState").NetworkFlowImportHandoffRequest
        | undefined;
      h.controller.setHandoff(async (value) => {
        request = value;
        return pending.promise;
      });
      const applying = h.controller.continueApply();
      await vi.waitFor(() => expect(request).toBeDefined());
      leave(h);
      if (h.controller.getSnapshot().handoff?.status === "loading")
        expect(request?.current()).toBe(true);
      else expect(request?.current()).toBe(false);
      pending.resolve({ kind: "unavailable" });
      await applying;
      expect(h.controller.getSnapshot().applyJob?.resource.status).toBe(
        "succeeded",
      );
      expect(h.client.send).toHaveBeenCalledTimes(4);
    }
  });
});
