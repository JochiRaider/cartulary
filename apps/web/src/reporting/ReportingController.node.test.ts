import { afterEach, expect, it, vi } from "vitest";
import type {
  ReportingClient,
  ReportingJob,
  ReportingReceipt,
} from "../services/reportingClient";
import {
  type ReportingBinding,
  ReportingController,
} from "./ReportingController";

const incidentId = "10000000-0000-4000-8000-000000000001",
  actorId = "10000000-0000-4000-8000-000000000002";
const jobId = "10000000-0000-4000-8000-000000000003",
  snapshotId = "10000000-0000-4000-8000-000000000004";
function job(overrides: Partial<ReportingJob> = {}): ReportingJob {
  return {
    job_id: jobId,
    scope: { kind: "incident", incident_id: incidentId },
    status: "queued",
    status_route: `/api/v1/jobs/${jobId}`,
    submitted_by_user_id: actorId,
    submitted_at: "2026-04-05T00:00:00Z",
    updated_at: "2026-04-05T00:00:00Z",
    started_at: null,
    finished_at: null,
    retained_until: null,
    progress: { completed: 0, total: 1 },
    cancelable: true,
    result_summary: null,
    error_summary: null,
    ...overrides,
  };
}
const receipt = (value: ReportingJob, status = 202): ReportingReceipt => ({
  ok: true,
  status,
  payload: { data: value, meta: { request_id: "reporting-test" } },
});
function succeeded(): ReportingJob {
  return job({
    status: "succeeded",
    cancelable: false,
    started_at: "2026-04-05T00:00:00Z",
    finished_at: "2026-04-05T00:00:01Z",
    updated_at: "2026-04-05T00:00:01Z",
    retained_until: "2026-04-12T00:00:01Z",
    progress: { completed: 1, total: 1 },
    result_summary: {
      code: "snapshot_created",
      message: "Snapshot ready",
      resource_refs: [
        {
          kind: "snapshot",
          id: snapshotId,
          route: `/api/v1/snapshots/${snapshotId}`,
        },
      ],
    },
  });
}
const controllers: ReportingController[] = [];
afterEach(() => {
  for (const controller of controllers.splice(0)) controller.dispose();
  vi.useRealTimers();
});
function setup() {
  let sequence = 0;
  const client: ReportingClient = {
    submit: vi.fn(async () => receipt(job())),
    read: vi.fn(async () => receipt(job(), 200)),
    cancel: vi.fn(async () =>
      receipt(job({ status: "cancel_requested", cancelable: false }), 200),
    ),
    snapshot: vi.fn(async () => ({
      ok: true as const,
      status: 200,
      payload: {
        meta: { request_id: "snapshot" },
        data: {
          snapshot_id: snapshotId,
          incident_id: incidentId,
          created_by_user_id: actorId,
          created_at: "2026-04-05T00:00:00Z",
          snapshot_at: "2026-04-05T00:00:00Z",
          source_change_set_high_watermark: "1",
          derivation_version: "test",
          source_capture_sha256: "0".repeat(64),
        },
      },
    })),
  };
  const controller = new ReportingController(
    undefined,
    () => `txn-${++sequence}`,
  );
  controllers.push(controller);
  const binding: ReportingBinding = {
    incidentId,
    actorId,
    lifetime: "lifetime",
    role: "editor",
    client,
    current: () => true,
    accessFailure: vi.fn(),
  };
  controller.bind(binding);
  controller.setPresented(true);
  return { controller, client, binding };
}

it("retains exact uncertain reporting commands and prevents duplicate admission", async () => {
  const { controller, client } = setup();
  let settle!: (v: ReportingReceipt) => void;
  vi.mocked(client.submit).mockImplementationOnce(
    () =>
      new Promise((resolve) => {
        settle = resolve;
      }),
  );
  const submitting = controller.createSnapshot();
  await controller.createSnapshot();
  expect(client.submit).toHaveBeenCalledTimes(1);
  const original = vi.mocked(client.submit).mock.calls[0]?.[0];
  settle({ ok: false, status: 503, payload: {} });
  await submitting;
  await controller.createSnapshot();
  expect(client.submit).toHaveBeenCalledTimes(1);
  await controller.replay();
  expect(vi.mocked(client.submit).mock.calls[1]?.[0]).toBe(original);
  expect(controller.getSnapshot().job?.job_id).toBe(jobId);
  expect(controller.getSnapshot().attempt).toBeNull();
});
it("rejects foreign jobs wrong result identities and regressing observations", async () => {
  const { controller, client } = setup();
  vi.mocked(client.submit).mockResolvedValueOnce(
    receipt(job({ scope: { kind: "incident", incident_id: actorId } })),
  );
  await controller.createSnapshot();
  expect(controller.getSnapshot().job).toBeNull();
  vi.mocked(client.submit).mockResolvedValueOnce(receipt(succeeded()));
  await controller.replay();
  expect(controller.getSnapshot().snapshotId).toBe(snapshotId);
  await controller.refresh();
  expect(controller.getSnapshot().job?.status).toBe("succeeded");
  expect(controller.getSnapshot().paused).toBe(true);
  controller.retire();
  const h = setup();
  const malformed = succeeded();
  const ref = malformed.result_summary?.resource_refs?.[0];
  if (!ref) throw new Error("Fixture must have a snapshot result");
  ref.route = `/api/v1/releases/${snapshotId}`;
  vi.mocked(h.client.submit).mockResolvedValueOnce(receipt(malformed));
  await h.controller.createSnapshot();
  expect(h.controller.getSnapshot().snapshotId).toBeNull();
});
it("pauses detached reporting without replay and retires replaced authority", async () => {
  const { controller, client, binding } = setup();
  let settle!: (v: ReportingReceipt) => void;
  vi.mocked(client.submit).mockImplementationOnce(
    () =>
      new Promise((resolve) => {
        settle = resolve;
      }),
  );
  const pending = controller.createSnapshot();
  controller.setPresented(false);
  settle(receipt(job()));
  await pending;
  expect(controller.getSnapshot().job).toBeNull();
  expect(controller.getSnapshot().attempt).not.toBeNull();
  controller.bind(binding);
  controller.setPresented(true);
  expect(client.submit).toHaveBeenCalledTimes(1);
  expect(client.read).not.toHaveBeenCalled();
  controller.bind({ ...binding, role: "viewer" });
  await controller.replay();
  expect(client.submit).toHaveBeenCalledTimes(1);
  controller.bind({ ...binding, actorId: snapshotId, lifetime: "replacement" });
  expect(controller.getSnapshot().attempt).toBeNull();
  expect(controller.getSnapshot().snapshotId).toBeNull();
});
it("cancels the exact reporting job and replays an uncertain cancellation unchanged", async () => {
  const { controller, client } = setup();
  await controller.createSnapshot();
  vi.mocked(client.cancel).mockRejectedValueOnce(new Error("lost response"));
  await controller.cancel();
  const captured = vi.mocked(client.cancel).mock.calls[0];
  expect(captured?.[0]).toBe(jobId);
  expect(controller.getSnapshot().cancellation).not.toBeNull();
  await controller.cancel();
  expect(vi.mocked(client.cancel).mock.calls[1]?.[1]).toBe(captured?.[1]);
  expect(controller.getSnapshot().job?.status).toBe("cancel_requested");
  expect(controller.getSnapshot().cancellation).toBeNull();
});
it("renders only the verified same-incident snapshot with the exact internal default template", async () => {
  const { controller, client } = setup();
  vi.mocked(client.submit).mockResolvedValueOnce(receipt(succeeded()));
  await controller.createSnapshot();
  await controller.createReport();
  expect(client.snapshot).toHaveBeenCalledWith(
    snapshotId,
    expect.any(AbortSignal),
  );
  expect(vi.mocked(client.submit).mock.calls[1]?.[0]).toEqual({
    kind: "report",
    request: {
      snapshot_id: snapshotId,
      client_txn_id: "txn-2",
      template_id: "cartulary.report.default",
      template_version: "4",
      output_kind: "slidev",
      release_scope: "internal_draft",
    },
  });
});
it("bounds reporting observation without retrying the product action or blocking cancellation behind a read", async () => {
  vi.useFakeTimers();
  const { controller, client } = setup();
  await controller.createSnapshot();
  vi.mocked(client.read).mockImplementation(() => new Promise(() => {}));
  const reading = controller.refresh();
  await controller.cancel();
  await reading;
  expect(client.cancel).toHaveBeenCalledTimes(1);
  await vi.advanceTimersByTimeAsync(31_000);
  expect(controller.getSnapshot().paused).toBe(true);
  expect(client.submit).toHaveBeenCalledTimes(1);
});
