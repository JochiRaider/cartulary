import {
  evidenceViewSchemaId,
  requireViewContract,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { deferred } from "../../../testing/fetchMockTestSupport";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { createEvidenceFileTransport } from "../../adapters/createEvidenceFileTransport";
import { createTimelineFileLinkTransport } from "../../adapters/createTimelineFileLinkTransport";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookAuthoringReadPort } from "../../ports/WorkbookAuthoringReadPort";
import { EvidenceFileFinalization } from "./EvidenceFileFinalization";
import { EvidenceUploadSession } from "./EvidenceUploadSession";
import {
  admitEvidenceFile,
  type EvidenceFileTransport,
} from "./evidenceFileOperation";
import { WorkbookEvidenceAttachmentOwner } from "./WorkbookEvidenceAttachmentOwner";
import { WorkbookTimelineFileOwner } from "./WorkbookTimelineFileOwner";

const incidentId = "40000000-0000-4000-8000-000000000004",
  evidenceId = "30000000-0000-4000-8000-000000000003",
  sourceId = "20000000-0000-4000-8000-000000000002",
  blobId = "50000000-0000-4000-8000-000000000005";
const authority: WorkbookMutationAuthority = {
  actorId: "10000000-0000-4000-8000-000000000001",
  incidentId,
  sessionIdentity: "session-one",
  role: "editor",
  closed: false,
};

function fixture() {
  let seq = 0;
  const ids = { create: vi.fn((prefix: string) => `${prefix}-${++seq}`) };
  const file = new File(["abc"], "capture.txt", { type: "text/plain" });
  const slot = {
    data: {
      incident_id: incidentId,
      object_blob_id: blobId,
      upload_state: "pending" as const,
      target_expires_at: new Date(Date.now() + 3_600_000).toISOString(),
      pending_expires_at: new Date(Date.now() + 86_400_000).toISOString(),
      upload_target: {
        method: "PUT" as const,
        href: "/api/v1/object-uploads/private-capability",
        expires_at: new Date(Date.now() + 3_600_000).toISOString(),
        headers: { "Content-Type": "text/plain" },
      },
      accepted_contract: {
        incident_id: incidentId,
        byte_size: 3,
        filename_hint: "capture.txt",
        content_type_hint: "text/plain",
        sha256_hex: null,
      },
    },
    meta: { request_id: "slot-request" },
  };
  slot.data.upload_target.expires_at = slot.data.target_expires_at;
  const evidence = fullWorkbookViewRow(
    requireViewContract(evidenceViewSchemaId),
    evidenceId,
    2,
    {
      "evidence.title": "capture.txt",
      "evidence.lifecycle_state": "requested",
      "evidence.storage_ref": `object://${blobId}`,
    },
  );
  const source = fullWorkbookViewRow(
    requireViewContract(timelineViewSchemaId),
    sourceId,
    1,
    { "timeline.raw_activity_text": "Original source" },
  );
  const receipt = {
    data: {
      view_schema_id: evidenceViewSchemaId,
      row: evidence,
      change_set_id: "60000000-0000-4000-8000-000000000006",
      object_blob_id: blobId,
    },
    meta: { request_id: "finalize-request" },
  };
  const transport = {
    ...createEvidenceFileTransport(undefined),
    slot: vi.fn<EvidenceFileTransport["slot"]>(async () => ({
      kind: "accepted",
      receipt: slot,
    })),
    transfer: vi.fn<EvidenceFileTransport["transfer"]>(async () => ({
      kind: "accepted",
    })),
    finalize: vi.fn<EvidenceFileTransport["finalize"]>(async () => ({
      kind: "accepted",
      receipt,
    })),
  };
  const reader: WorkbookAuthoringReadPort = {
    availableViews: async () => ({
      kind: "accepted" as const,
      value: [evidenceViewSchemaId, timelineViewSchemaId],
    }),
    verify: async () => {},
    page: async ({ viewSchemaId }) => {
      const row =
        viewSchemaId === evidenceViewSchemaId
          ? { ...evidence, row_version: 1 }
          : source;
      return {
        kind: "accepted",
        value: {
          candidates: [
            {
              recordId: row.record_id,
              viewSchemaId,
              displayText: "Original record",
              row,
            },
          ],
          hasMore: false,
          nextCursor: null,
        },
      };
    },
  };
  const effects = {
    coordinate: vi.fn(async () => ({
      kind: "settled" as const,
      minimumRowVersion: 0,
    })),
    accepted: vi.fn(),
    refresh: vi.fn(async () => {}),
  };
  return {
    ids,
    file,
    slot,
    evidence,
    source,
    receipt,
    transport,
    reader,
    effects,
  };
}

describe("retained Evidence file recovery", () => {
  it("rejects multiple files without choosing or discarding one", () => {
    const f = fixture();
    expect(admitEvidenceFile([f.file, f.file])).toEqual({
      kind: "rejected",
      message: "Choose one file at a time.",
    });
    expect(admitEvidenceFile([])).toEqual({ kind: "empty" });
  });
  it("retains an exact uncertain slot and never repeats an uncertain transfer", async () => {
    const f = fixture();
    f.transport.slot.mockResolvedValueOnce({ kind: "uncertain" });
    f.transport.transfer.mockResolvedValueOnce({ kind: "uncertain" });
    const upload = new EvidenceUploadSession(
      f.file,
      f.transport,
      f.ids,
      () => {},
      () => {},
    );
    expect(await upload.prepare(authority)).toBe(false);
    expect(await upload.prepare(authority)).toBe(false);
    expect(f.transport.slot.mock.calls[0]?.[0]).toBe(
      f.transport.slot.mock.calls[1]?.[0],
    );
    expect(upload.isFinalizable).toBe(true);
    expect(await upload.prepare(authority)).toBe(false);
    expect(f.transport.transfer).toHaveBeenCalledTimes(1);
    expect(f.transport.transfer.mock.calls[0]?.[1]).toBe(f.file);
    expect(upload.hasFile).toBe(true);
    upload.finalized();
    expect(upload.hasFile).toBe(false);
  });
  it("retires an unused capability after issuing-session replacement", async () => {
    const f = fixture();
    f.transport.transfer.mockResolvedValueOnce({ kind: "not_dispatched" });
    const upload = new EvidenceUploadSession(
      f.file,
      f.transport,
      f.ids,
      () => {},
      () => {},
    );
    await upload.prepare(authority);
    upload.setAuthority({ ...authority, sessionIdentity: "session-two" });
    expect(
      await upload.prepare({ ...authority, sessionIdentity: "session-two" }),
    ).toBe(false);
    expect(upload.status.phase).toBe("fresh_required");
    expect(f.transport.transfer).toHaveBeenCalledTimes(1);
    expect(upload.hasFile).toBe(true);
    upload.retire();
    expect(upload.hasFile).toBe(false);
  });
  it("retains existing-Evidence acceptance across detachment and failed refresh", async () => {
    const f = fixture();
    const pending =
      deferred<Awaited<ReturnType<EvidenceFileTransport["finalize"]>>>();
    f.transport.finalize.mockReturnValueOnce(pending.promise);
    f.effects.refresh.mockRejectedValueOnce(new Error("read failed"));
    const owner = new WorkbookEvidenceAttachmentOwner(
      incidentId,
      f.ids,
      f.effects,
    );
    owner.configure(f.reader, async () => authority, f.transport);
    owner.setAuthority(authority);
    const token = Symbol();
    owner.attach(token, evidenceId);
    owner.begin({ ...f.evidence, row_version: 1 }, [f.file]);
    owner.begin({ ...f.evidence, row_version: 1 }, [f.file]);
    await waitFor(() => expect(f.transport.finalize).toHaveBeenCalledTimes(1));
    owner.detach(token);
    pending.resolve({ kind: "accepted", receipt: f.receipt });
    await waitFor(() =>
      expect(owner.getSnapshot()[0]?.refreshRequired).toBe(true),
    );
    expect(owner.getSnapshot()[0]?.accepted).toBe(true);
    await owner.refresh(evidenceId);
    expect(f.transport.slot).toHaveBeenCalledTimes(1);
    expect(f.transport.transfer).toHaveBeenCalledTimes(1);
    expect(f.transport.finalize).toHaveBeenCalledTimes(1);
    expect(JSON.stringify(owner.getSnapshot())).not.toContain(
      "private-capability",
    );
  });
  it("finalizes an uncertain transfer and exactly replays uncertain finalization", async () => {
    const f = fixture();
    f.transport.transfer.mockResolvedValueOnce({ kind: "uncertain" });
    f.transport.finalize.mockResolvedValueOnce({ kind: "uncertain" });
    const owner = new WorkbookEvidenceAttachmentOwner(
      incidentId,
      f.ids,
      f.effects,
    );
    owner.configure(f.reader, async () => authority, f.transport);
    owner.setAuthority(authority);
    owner.begin({ ...f.evidence, row_version: 1 }, [f.file]);
    await waitFor(() => expect(owner.getSnapshot()[0]?.busy).toBe(false));
    expect(f.transport.finalize).not.toHaveBeenCalled();
    await owner.resume(evidenceId);
    await owner.resume(evidenceId);
    expect(f.transport.finalize.mock.calls[0]?.[0]).toBe(
      f.transport.finalize.mock.calls[1]?.[0],
    );
    expect(owner.getSnapshot()[0]?.accepted).toBe(true);
    expect(f.transport.transfer).toHaveBeenCalledTimes(1);
  });
  it("preserves created Evidence after Timeline detachment and recovers only the link", async () => {
    const f = fixture();
    const pending =
      deferred<Awaited<ReturnType<EvidenceFileTransport["finalize"]>>>();
    f.transport.finalize.mockReturnValueOnce(pending.promise);
    const linked = {
      data: {
        view_schema_id: timelineViewSchemaId,
        change_set_id: "70000000-0000-4000-8000-000000000007",
        row: { ...f.source, row_version: 2 },
      },
      meta: { request_id: "link-request" },
    };
    const links = {
      ...createTimelineFileLinkTransport(undefined),
      send: vi.fn(async () => ({ kind: "accepted" as const, receipt: linked })),
    };
    const owner = new WorkbookTimelineFileOwner(incidentId, f.ids, f.effects);
    owner.configure(f.reader, async () => authority, f.transport, links);
    owner.setAuthority(authority);
    owner.setPresentation("original");
    owner.begin({ key: sourceId, recordId: sourceId, rowVersion: 1 }, [f.file]);
    await waitFor(() => expect(f.transport.finalize).toHaveBeenCalledTimes(1));
    owner.setPresentation(null);
    pending.resolve({ kind: "accepted", receipt: f.receipt });
    await waitFor(() => expect(owner.getSnapshot()[0]?.busy).toBe(false));
    expect(links.send).not.toHaveBeenCalled();
    owner.setPresentation("restored");
    await owner.review(sourceId);
    owner.confirmReview(sourceId);
    await owner.resume(sourceId);
    expect(links.send).toHaveBeenCalledTimes(1);
    expect(f.transport.finalize).toHaveBeenCalledTimes(1);
    expect(f.transport.transfer).toHaveBeenCalledTimes(1);
    expect(owner.getSnapshot()[0]?.attached).toBe(true);
    owner.suspend();
    expect(owner.getSnapshot()).toEqual([]);
    owner.setAuthority({ ...authority, sessionIdentity: "new" });
    expect(owner.getSnapshot()[0]?.attached).toBe(true);
    owner.setAuthority({ ...authority, actorId: "other-account" });
    expect(owner.getSnapshot()).toEqual([]);
  });
  it("retains timed-out finalization through replay denial and accepts its late receipt", async () => {
    vi.useFakeTimers();
    try {
      const f = fixture(),
        pending =
          deferred<Awaited<ReturnType<EvidenceFileTransport["finalize"]>>>();
      f.transport.finalize
        .mockReturnValueOnce(pending.promise)
        .mockResolvedValueOnce({
          kind: "rejected",
          failure: { kind: "authorization_lost", message: "denied" },
        });
      const accepted = vi.fn(),
        finalization = new EvidenceFileFinalization(
          f.transport,
          () => {},
          accepted,
        );
      const attempt = f.transport.capture({
        stage: "attach",
        authority,
        clientTxnId: "same-finalize",
        recordId: evidenceId,
        baseRowVersion: 1,
        objectBlobId: blobId,
      });
      const first = finalization.send(attempt);
      await vi.advanceTimersByTimeAsync(30_001);
      await first;
      expect(finalization.state).toMatchObject({
        phase: "uncertain",
        pending: false,
      });
      await finalization.send(attempt);
      expect(finalization.state.phase).toBe("uncertain");
      expect(finalization.resetRejected()).toBe(false);
      pending.resolve({ kind: "accepted", receipt: f.receipt });
      await vi.advanceTimersByTimeAsync(0);
      expect(finalization.state.receipt).toBe(f.receipt);
      expect(accepted).toHaveBeenCalledTimes(1);
      expect(f.transport.finalize.mock.calls[0]?.[0]).toBe(
        f.transport.finalize.mock.calls[1]?.[0],
      );

      const stale =
        deferred<Awaited<ReturnType<EvidenceFileTransport["finalize"]>>>();
      const current =
        deferred<Awaited<ReturnType<EvidenceFileTransport["finalize"]>>>();
      f.transport.finalize
        .mockReturnValueOnce(stale.promise)
        .mockReturnValueOnce(current.promise);
      const later = new EvidenceFileFinalization(
        f.transport,
        () => {},
        accepted,
      );
      const original = later.send(attempt);
      await vi.advanceTimersByTimeAsync(30_001);
      await original;
      const replay = later.send(attempt);
      stale.resolve({
        kind: "rejected",
        failure: { kind: "authorization_lost", message: "denied" },
      });
      await vi.advanceTimersByTimeAsync(0);
      expect(later.state).toMatchObject({ phase: "submitting", pending: true });
      expect(later.resetRejected()).toBe(false);
      current.resolve({ kind: "accepted", receipt: f.receipt });
      await replay;
      expect(later.state.receipt).toBe(f.receipt);

      const oldSuccess =
        deferred<Awaited<ReturnType<EvidenceFileTransport["finalize"]>>>();
      f.transport.finalize
        .mockReturnValueOnce(oldSuccess.promise)
        .mockReturnValueOnce(new Promise(() => {}));
      const lateOwner = new EvidenceFileFinalization(
        f.transport,
        () => {},
        accepted,
      );
      const initial = lateOwner.send(attempt);
      await vi.advanceTimersByTimeAsync(30_001);
      await initial;
      const outstanding = lateOwner.send(attempt);
      oldSuccess.resolve({ kind: "accepted", receipt: f.receipt });
      await vi.advanceTimersByTimeAsync(0);
      expect(lateOwner.state).toMatchObject({
        phase: "accepted",
        pending: false,
        receipt: f.receipt,
      });
      await vi.advanceTimersByTimeAsync(30_001);
      await outstanding;
      expect(lateOwner.state.pending).toBe(false);
    } finally {
      vi.useRealTimers();
    }
  });
  it("discards a transferring file without creating Evidence from a late callback", async () => {
    const f = fixture(),
      pending =
        deferred<Awaited<ReturnType<EvidenceFileTransport["transfer"]>>>();
    f.transport.transfer.mockReturnValueOnce(pending.promise);
    const owner = new WorkbookTimelineFileOwner(incidentId, f.ids, f.effects);
    owner.configure(
      f.reader,
      async () => authority,
      f.transport,
      createTimelineFileLinkTransport(undefined),
    );
    owner.setAuthority(authority);
    owner.begin({ key: sourceId, recordId: sourceId, rowVersion: 1 }, [f.file]);
    await waitFor(() => expect(f.transport.transfer).toHaveBeenCalledTimes(1));
    expect(owner.blocksRecord(sourceId)).toBe(false);
    owner.discard(sourceId);
    expect(owner.getSnapshot()).toEqual([]);
    pending.resolve({ kind: "accepted" });
    await Promise.resolve();
    await Promise.resolve();
    expect(f.transport.finalize).not.toHaveBeenCalled();
    expect(f.transport.transfer.mock.calls[0]?.[2].aborted).toBe(true);
  });
  it("requires displayed source review after edits and never links a missing source", async () => {
    const f = fixture(),
      pending =
        deferred<Awaited<ReturnType<EvidenceFileTransport["finalize"]>>>();
    f.transport.finalize.mockReturnValueOnce(pending.promise);
    const links = {
      ...createTimelineFileLinkTransport(undefined),
      send: vi.fn(),
    };
    const owner = new WorkbookTimelineFileOwner(incidentId, f.ids, f.effects);
    owner.configure(f.reader, async () => authority, f.transport, links);
    owner.setAuthority(authority);
    owner.begin({ key: sourceId, recordId: sourceId, rowVersion: 1 }, [f.file]);
    await waitFor(() => expect(f.transport.finalize).toHaveBeenCalledTimes(1));
    owner.acceptVersion(sourceId, 2);
    pending.resolve({ kind: "accepted", receipt: f.receipt });
    await waitFor(() => expect(owner.getSnapshot()[0]?.busy).toBe(false));
    await owner.resume(sourceId);
    expect(links.send).not.toHaveBeenCalled();
    f.reader.page = async () => ({
      kind: "accepted",
      value: { candidates: [], hasMore: false, nextCursor: null },
    });
    await owner.review(sourceId);
    expect(owner.getSnapshot()[0]?.reviewText).toBe(null);
    owner.confirmReview(sourceId);
    await owner.resume(sourceId);
    expect(links.send).not.toHaveBeenCalled();
    expect(owner.getSnapshot()[0]?.evidenceAccepted).toBe(true);
    expect(f.transport.finalize).toHaveBeenCalledTimes(1);
  });
  it("recovers operation authorization denial only after current authority and explicit source review", async () => {
    for (const stage of ["slot", "finalize"] as const) {
      const f = fixture();
      f.transport[stage].mockResolvedValueOnce({
        kind: "rejected",
        failure: { kind: "authorization_lost", message: "denied" },
      });
      let current: WorkbookMutationAuthority | null = authority;
      const authorityReader = vi.fn(async () => {
        if (!current) throw new Error("uncertain session");
        return current;
      });
      const owner = new WorkbookEvidenceAttachmentOwner(
        incidentId,
        f.ids,
        f.effects,
      );
      owner.configure(f.reader, authorityReader, f.transport);
      owner.setAuthority(authority);
      owner.begin({ ...f.evidence, row_version: 1 }, [f.file]);
      await waitFor(() => expect(owner.getSnapshot()[0]?.busy).toBe(false));
      expect(owner.getSnapshot()).toHaveLength(1); // An operation denial does not revoke the incident.
      expect(owner.getSnapshot()[0]?.accepted).toBe(false);
      current = null;
      await owner.review(evidenceId);
      expect(owner.getSnapshot()).toEqual([]);
      const attempts = f.transport[stage].mock.calls.length;
      await owner.resume(evidenceId);
      expect(f.transport[stage]).toHaveBeenCalledTimes(attempts);
      current = { ...authority, sessionIdentity: "recovered-session" };
      owner.setAuthority(current);
      await owner.review(evidenceId);
      expect(owner.getSnapshot()[0]?.reviewText).toContain("capture.txt");
      owner.confirmReview(evidenceId);
      await owner.resume(evidenceId);
      if (owner.getSnapshot()[0]?.canFreshSlot) {
        owner.freshSlot(evidenceId);
        await owner.review(evidenceId);
        owner.confirmReview(evidenceId);
        await owner.resume(evidenceId);
      }
      expect(owner.getSnapshot()[0]?.accepted).toBe(true);
      expect(f.transport[stage].mock.calls[1]?.[0].clientTxnId).not.toBe(
        f.transport[stage].mock.calls[0]?.[0].clientTxnId,
      );
      owner.retire();
    }
  });
  it("conceals retained files and fences late creation on role loss closure and account retirement", async () => {
    for (const boundary of [
      "viewer",
      "closed",
      "suspend",
      "account",
      "retire",
    ] as const) {
      const f = fixture(),
        pending =
          deferred<Awaited<ReturnType<EvidenceFileTransport["finalize"]>>>();
      f.transport.finalize.mockReturnValueOnce(pending.promise);
      const links = {
        ...createTimelineFileLinkTransport(undefined),
        send: vi.fn(),
      };
      const owner = new WorkbookTimelineFileOwner(incidentId, f.ids, f.effects);
      owner.configure(f.reader, async () => authority, f.transport, links);
      owner.setAuthority(authority);
      owner.begin({ key: sourceId, recordId: sourceId, rowVersion: 1 }, [
        f.file,
      ]);
      await waitFor(() =>
        expect(f.transport.finalize).toHaveBeenCalledTimes(1),
      );
      if (boundary === "viewer")
        owner.setAuthority({ ...authority, role: "viewer" });
      if (boundary === "closed") owner.closeIncident();
      if (boundary === "suspend") owner.suspend();
      if (boundary === "account")
        owner.setAuthority({ ...authority, actorId: "replacement-account" });
      if (boundary === "retire") owner.retire();
      pending.resolve({ kind: "accepted", receipt: f.receipt });
      await Promise.resolve();
      await Promise.resolve();
      await Promise.resolve();
      expect(links.send).not.toHaveBeenCalled();
      if (["suspend", "account", "retire"].includes(boundary))
        expect(owner.getSnapshot()).toEqual([]);
      else expect(owner.getSnapshot()[0]?.canResume).toBe(false);
      if (["account", "retire"].includes(boundary))
        expect(f.effects.accepted).not.toHaveBeenCalled();
      else expect(f.effects.accepted).toHaveBeenCalledTimes(1);
      owner.retire();
    }
  });
  it("preserves selected files on replacement refusal and resolves off-page sources through the incident reader", async () => {
    const f = fixture();
    f.transport.slot.mockResolvedValueOnce({ kind: "uncertain" });
    const linked = {
      data: {
        view_schema_id: timelineViewSchemaId,
        change_set_id: "70000000-0000-4000-8000-000000000007",
        row: { ...f.source, row_version: 2 },
      },
      meta: { request_id: "link" },
    };
    const links = {
      ...createTimelineFileLinkTransport(undefined),
      send: vi.fn(async () => ({ kind: "accepted" as const, receipt: linked })),
    };
    const page = f.reader.page;
    f.reader.page = vi.fn<WorkbookAuthoringReadPort["page"]>(async (input) =>
      input.cursor === null && input.viewSchemaId === timelineViewSchemaId
        ? {
            kind: "accepted",
            value: { candidates: [], nextCursor: "off-page", hasMore: true },
          }
        : page(input),
    );
    const owner = new WorkbookTimelineFileOwner(incidentId, f.ids, f.effects);
    owner.configure(f.reader, async () => authority, f.transport, links);
    owner.setAuthority(authority);
    const source = { key: sourceId, recordId: sourceId, rowVersion: 1 };
    owner.begin(source, [f.file]);
    await waitFor(() => expect(owner.getSnapshot()[0]?.busy).toBe(false));
    expect(
      owner.begin(source, [new File(["replacement"], "replacement.txt")]),
    ).toContain("Resume or discard");
    expect(owner.begin(source, [f.file, f.file])).toBe(
      "Choose one file at a time.",
    );
    expect(owner.getAdmissionNotice()).toBe("Choose one file at a time.");
    expect(owner.getSnapshot()[0]?.filename).toBe(f.file.name);
    await owner.resume(sourceId);
    expect(f.transport.transfer.mock.calls[0]?.[1]).toBe(f.file);
    expect(links.send).toHaveBeenCalledTimes(1);
    expect(f.reader.page).toHaveBeenCalledWith(
      expect.objectContaining({
        cursor: "off-page",
        viewSchemaId: timelineViewSchemaId,
      }),
    );
    owner.discard(sourceId);
    expect(owner.getSnapshot()).toEqual([]);
  });
});
