import {
  evidenceViewSchemaId,
  requireViewContract,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../../../testing/fetchMockTestSupport";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { createTimelineRelatedEvidenceTransport } from "../../adapters/createTimelineRelatedEvidenceTransport";
import { readWorkbookAuthoringRecord } from "../../adapters/readWorkbookAuthoringRecord";
import type { RecordChangedMessage } from "../../collaboration/workbookCollaborationMessages";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookAuthoringReadPort } from "../../ports/WorkbookAuthoringReadPort";
import type { WorkbookSourceWriteSettlement } from "../../ports/WorkbookSourceWriteCoordination";
import { TimelineRelatedEvidenceContext } from "./TimelineRelatedEvidenceContext";
import type {
  RelatedEvidenceOutcome,
  RelatedEvidenceReceipt,
  RelatedEvidenceTransport,
} from "./timelineRelatedEvidenceOperation";
import { useTimelineRelatedEvidenceAttachment } from "./useTimelineRelatedEvidenceAttachment";
import { WorkbookTimelineRelatedEvidenceOwner } from "./WorkbookTimelineRelatedEvidenceOwner";

afterEach(() => {
  cleanup();
  vi.useRealTimers();
  vi.unstubAllGlobals();
});
const actor = "10000000-0000-4000-8000-000000000001",
  sourceId = "20000000-0000-4000-8000-000000000002",
  targetId = "30000000-0000-4000-8000-000000000003";
const authority: WorkbookMutationAuthority = {
  actorId: actor,
  incidentId: "40000000-0000-4000-8000-000000000004",
  sessionIdentity: "session",
  role: "editor",
  closed: false,
};
const attachment = Symbol("inspector");
function fixture() {
  let sequence = 0;
  const ids = { create: vi.fn((prefix: string) => `${prefix}-${++sequence}`) };
  const effects = {
    coordinate: vi.fn<() => Promise<WorkbookSourceWriteSettlement>>(
      async () => ({ kind: "settled", minimumRowVersion: 0 }),
    ),
    accepted: vi.fn(),
    refresh: vi.fn(async () => {}),
    conflict: vi.fn(),
  };
  const owner = new WorkbookTimelineRelatedEvidenceOwner(
    authority.incidentId,
    ids,
    effects,
  );
  owner.setAuthority(authority);
  const source = fullWorkbookViewRow(
    requireViewContract(timelineViewSchemaId),
    sourceId,
    1,
    {
      "timeline.raw_activity_text": "Raw preserved  Timeline text",
      "timeline.capture_state": "reviewed",
    },
  );
  const target = fullWorkbookViewRow(
    requireViewContract(evidenceViewSchemaId),
    targetId,
    1,
    {
      "evidence.title": "Evidence metadata",
      "evidence.lifecycle_state": "requested",
      "evidence.linked_record_count": 0,
    },
  );
  const feature = requireViewContract(
    timelineViewSchemaId,
  ).inspectorConfig.featureGroups.find(
    (item) => item.featureGroupKey === "create_related.evidence",
  );
  if (!feature) throw new Error("Missing feature");
  const subject = {
    subject: {
      kind: "live" as const,
      recordId: sourceId,
      rowVersion: 1,
      viewSchemaId: timelineViewSchemaId,
      label: "Original source",
      surfaceLabel: "Timeline",
    },
    cells: source.cells,
  };
  const current = {
    source: source as typeof source | null,
    target: target as typeof target | null,
  };
  const reader: WorkbookAuthoringReadPort = {
    availableViews: async () => ({
      kind: "accepted" as const,
      value: [evidenceViewSchemaId, timelineViewSchemaId],
    }),
    verify: vi.fn(async () => {}),
    page: vi.fn(async ({ viewSchemaId }) => {
      const row =
        viewSchemaId === timelineViewSchemaId ? current.source : current.target;
      return {
        kind: "accepted" as const,
        value: {
          candidates: row
            ? [
                {
                  recordId: row.record_id,
                  viewSchemaId,
                  displayText: "Record",
                  row,
                },
              ]
            : [],
          hasMore: false,
          nextCursor: null,
        },
      };
    }),
  };
  const createReceipt: RelatedEvidenceReceipt = {
    data: {
      view_schema_id: evidenceViewSchemaId,
      row: target,
      change_set_id: "50000000-0000-4000-8000-000000000005",
    },
    meta: { request_id: "create-request" },
  };
  const linkReceipt: RelatedEvidenceReceipt = {
    data: {
      view_schema_id: timelineViewSchemaId,
      change_set_id: "60000000-0000-4000-8000-000000000006",
      row: {
        ...source,
        row_version: 2,
        cells: {
          ...source.cells,
          "timeline.capture_state": { value: "enriched" },
          "timeline.attached_evidence_ids": {
            value: {
              kind: "collection_value_v1",
              ordered: false,
              items: [{ linked_record_id: targetId, item_ref: "link" }],
            },
          },
        },
      },
    },
    meta: { request_id: "link-request" },
  };
  const transport = {
    ...createTimelineRelatedEvidenceTransport("/original-api"),
    send: vi.fn<RelatedEvidenceTransport["send"]>(async (attempt) => {
      if (attempt.stage === "create")
        return { kind: "accepted", receipt: createReceipt };
      current.source = linkReceipt.data.row;
      current.target = {
        ...target,
        cells: {
          ...target.cells,
          "evidence.linked_record_count": { value: 1 },
        },
      };
      return { kind: "accepted", receipt: linkReceipt };
    }),
  };
  const authorityReader = vi.fn(
    async (): Promise<WorkbookMutationAuthority> => authority,
  );
  owner.configure(reader, authorityReader, transport);
  owner.begin(
    subject,
    feature,
    { kind: "view_schema", id: timelineViewSchemaId },
    attachment,
  );
  owner.update(
    "evidence.collector_party_text",
    "  Source-preserved Collector  ",
  );
  const checkpoint = () => {
    const value = owner.getSnapshot().checkpoints[0];
    if (!value) throw new Error("Missing checkpoint");
    return value;
  };
  return {
    owner,
    ids,
    effects,
    reader,
    current,
    transport,
    authorityReader,
    createReceipt,
    linkReceipt,
    source,
    target,
    subject,
    feature,
    checkpoint,
  };
}
describe("Timeline related Evidence recovery", () => {
  it("reserves duplicate activation and captures two immutable exact requests with independent full receipts", async () => {
    const f = fixture(),
      gate = deferred<WorkbookSourceWriteSettlement>();
    f.effects.coordinate.mockReturnValueOnce(gate.promise);
    const submit = f.owner.submit(attachment);
    void f.owner.submit(attachment);
    f.owner.update("evidence.title", "Ignored during preparation");
    expect(f.transport.send).not.toHaveBeenCalled();
    gate.resolve({ kind: "settled", minimumRowVersion: 0 });
    await submit;
    await waitFor(() =>
      expect(f.checkpoint().links[0]?.refresh).toBe("complete"),
    );
    expect(f.transport.send).toHaveBeenCalledTimes(2);
    const create = f.checkpoint().create.attempt,
      link = f.checkpoint().links[0]?.attempt;
    expect(JSON.parse(create.body)).toEqual({
      client_txn_id: create.clientTxnId,
      "evidence.collector_party_text": "Source-preserved Collector",
    });
    expect(link?.path).toBe(`/api/v1/records/${sourceId}`);
    expect(JSON.parse(link?.body ?? "{}")).toEqual({
      view_schema_id: timelineViewSchemaId,
      base_row_version: 1,
      client_txn_id: link?.clientTxnId,
      changes: [
        {
          field_key: "timeline.attached_evidence_ids",
          action_payload: {
            kind: "collection_actions_v1",
            actions: [{ op: "add_record_ref", linked_record_id: targetId }],
          },
        },
      ],
    });
    expect(create.clientTxnId).not.toBe(link?.clientTxnId);
    expect(Object.isFrozen(create.review.draft.values)).toBe(true);
    expect(f.checkpoint().create.receipt).toEqual(f.createReceipt);
    expect(f.checkpoint().links[0]?.receipt).toEqual(f.linkReceipt);
    expect(
      f.owner.latestRow(sourceId)?.cells["timeline.raw_activity_text"],
    ).toEqual(f.source.cells["timeline.raw_activity_text"]);
    expect(
      f.owner.latestRow(targetId)?.cells["evidence.linked_record_count"]?.value,
    ).toBe(1);
  });
  it("retains committed creation after link rejection and retries only a newly reviewed link", async () => {
    const f = fixture();
    f.transport.send
      .mockResolvedValueOnce({ kind: "accepted", receipt: f.createReceipt })
      .mockResolvedValueOnce({
        kind: "rejected",
        failure: { kind: "stale_target", message: "Source changed" },
      });
    await f.owner.submit(attachment);
    await waitFor(() =>
      expect(f.checkpoint().links[0]?.phase).toBe("rejected"),
    );
    expect(f.owner.getSnapshot().draft).toBeNull();
    expect(f.checkpoint().create.receipt).toEqual(f.createReceipt);
    expect(
      f.owner.begin(
        f.subject,
        f.feature,
        { kind: "view_schema", id: timelineViewSchemaId },
        attachment,
      ),
    ).toBe(false);
    expect(f.owner.getSnapshot().draft).toBeNull();
    await waitFor(() => expect(f.owner.busy).toBe(false));
    expect(await f.owner.reviewLink(f.checkpoint().id)).toBe(true);
    await f.owner.link(f.checkpoint().id);
    expect(
      f.transport.send.mock.calls.map(([attempt]) => attempt.stage),
    ).toEqual(["create", "link", "link"]);
    expect(f.checkpoint().links[1]?.receipt).toEqual(f.linkReceipt);
  });
  it("recovers uncertain creation with the original key and body then requires review of the original source", async () => {
    const f = fixture();
    f.transport.send
      .mockResolvedValueOnce({ kind: "uncertain" })
      .mockResolvedValueOnce({ kind: "accepted", receipt: f.createReceipt });
    await f.owner.submit(attachment);
    const original = f.checkpoint().create.attempt;
    f.owner.detach(attachment);
    f.owner.update("evidence.title", "Must not replace uncertain input");
    await f.owner.replay(f.checkpoint().id, original.clientTxnId);
    expect(f.transport.send.mock.calls[1]?.[0]).toBe(original);
    expect(f.transport.send).toHaveBeenCalledTimes(2);
    expect(f.ids.create).toHaveBeenCalledTimes(1);
    expect(f.checkpoint().create.receipt?.data.row.record_id).toBe(targetId);
    expect(f.checkpoint().links).toHaveLength(0);
    expect(await f.owner.reviewLink(f.checkpoint().id)).toBe(true);
    await f.owner.link(f.checkpoint().id);
    expect(f.checkpoint().links[0]?.attempt.review.source.record_id).toBe(
      sourceId,
    );
  });
  it("replays uncertain linking unchanged after source version changes without repeating Evidence creation", async () => {
    const f = fixture();
    f.transport.send
      .mockResolvedValueOnce({ kind: "accepted", receipt: f.createReceipt })
      .mockResolvedValueOnce({ kind: "uncertain" })
      .mockResolvedValueOnce({ kind: "accepted", receipt: f.linkReceipt });
    await f.owner.submit(attachment);
    await waitFor(() =>
      expect(f.checkpoint().links[0]?.phase).toBe("uncertain"),
    );
    const link = f.checkpoint().links[0]?.attempt;
    if (!link) throw new Error("Missing link");
    f.current.source = { ...f.linkReceipt.data.row, row_version: 7 };
    f.owner.observe(sourceId, 7);
    await f.owner.replay(f.checkpoint().id, link.clientTxnId);
    expect(f.transport.send.mock.calls[2]?.[0]).toBe(link);
    expect(JSON.parse(link.body).base_row_version).toBe(1);
    expect(f.ids.create).toHaveBeenCalledTimes(2);
    await waitFor(() =>
      expect(f.owner.latestRow(sourceId)?.row_version).toBe(7),
    );
    expect(f.checkpoint().links[0]?.receipt?.data.row.row_version).toBe(2);
  });
  it("pauses after navigation or remote changes between stages and cancels only undispatched preparation", async () => {
    for (const navigation of [true, false]) {
      const f = fixture(),
        response = deferred<RelatedEvidenceOutcome>();
      f.transport.send.mockReturnValueOnce(response.promise);
      const submit = f.owner.submit(attachment);
      await waitFor(() => expect(f.transport.send).toHaveBeenCalledTimes(1));
      if (navigation) f.owner.detach(attachment);
      else {
        f.current.source = { ...f.source, row_version: 3 };
        f.owner.observe(sourceId, 3);
      }
      response.resolve({ kind: "accepted", receipt: f.createReceipt });
      await submit;
      await waitFor(() => expect(f.owner.busy).toBe(false));
      expect(f.checkpoint().create.phase).toBe("accepted");
      expect(f.checkpoint().links).toHaveLength(0);
      expect(f.transport.send).toHaveBeenCalledTimes(1);
    }
    const f = fixture(),
      saves = deferred<WorkbookSourceWriteSettlement>();
    f.effects.coordinate.mockReturnValueOnce(saves.promise);
    const submit = f.owner.submit(attachment);
    f.owner.detach(attachment);
    saves.resolve({ kind: "settled", minimumRowVersion: 0 });
    await submit;
    expect(f.transport.send).not.toHaveBeenCalled();
    expect(f.owner.getSnapshot().draft?.source.recordId).toBe(sourceId);
  });
  it("keeps accepted receipts through projection failure and recovers with reads only", async () => {
    const f = fixture();
    f.effects.refresh.mockRejectedValueOnce(new Error("Refresh unavailable"));
    await f.owner.submit(attachment);
    await waitFor(() =>
      expect(f.checkpoint().links[0]?.refresh).toBe("required"),
    );
    expect(f.checkpoint().create.phase).toBe("accepted");
    expect(f.checkpoint().links[0]?.phase).toBe("accepted");
    await f.owner.retryRefresh(f.checkpoint().id);
    expect(f.checkpoint().links[0]?.refresh).toBe("complete");
    expect(f.transport.send).toHaveBeenCalledTimes(2);
    f.current.source = null;
    f.current.target = null;
    await f.owner.retryRefresh(f.checkpoint().id);
    expect(f.checkpoint()).toMatchObject({
      sourceUnavailable: true,
      evidenceUnavailable: true,
    });
    expect(f.owner.linkComplete(f.checkpoint())).toBe(true);
    expect(f.checkpoint().create.receipt).toEqual(f.createReceipt);
  });
  it("distinguishes definitive field rejection and explicit new-ID recovery from uncertainty", async () => {
    const f = fixture();
    f.transport.send.mockResolvedValueOnce({
      kind: "rejected",
      failure: {
        kind: "validation",
        message: "Invalid collector",
        fields: [
          {
            field: "evidence.collector_party_text",
            message: "Collector field detail",
          },
        ],
      },
    });
    await f.owner.submit(attachment);
    expect(f.owner.getSnapshot().errors["evidence.collector_party_text"]).toBe(
      "Collector field detail",
    );
    expect(
      f.owner.getSnapshot().draft?.values["evidence.collector_party_text"],
    ).toBe("  Source-preserved Collector  ");
    f.owner.update("evidence.collector_party_text", "Corrected");
    expect(await f.owner.review()).toBe(true);
    f.transport.send.mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "client_txn_conflict", message: "Request conflict" },
    });
    await f.owner.submit(attachment);
    const rejected = f.owner.getSnapshot().checkpoints[1];
    if (!rejected) throw new Error("Missing new attempt");
    await f.owner.review();
    await f.owner.submit(attachment);
    expect(f.transport.send).toHaveBeenCalledTimes(2);
    f.owner.reviewNewRequestId(rejected.id, rejected.id);
    expect(await f.owner.review()).toBe(true);
    await f.owner.submit(attachment);
    await waitFor(() => expect(f.transport.send).toHaveBeenCalledTimes(4));
    const uncertain = fixture();
    uncertain.transport.send
      .mockResolvedValueOnce({ kind: "uncertain" })
      .mockResolvedValueOnce({
        kind: "rejected",
        failure: {
          kind: "authorization_lost",
          message: "Local Evidence denied",
        },
      });
    await uncertain.owner.submit(attachment);
    await uncertain.owner.replay(
      uncertain.checkpoint().id,
      uncertain.checkpoint().id,
    );
    expect(uncertain.checkpoint().create.phase).toBe("uncertain");
    expect(uncertain.owner.getSnapshot().authority).not.toBeNull();
  });
  it("accepts late creation during suspension and retires protected state on account replacement", async () => {
    const f = fixture(),
      response = deferred<RelatedEvidenceOutcome>();
    f.transport.send.mockReturnValueOnce(response.promise);
    const submit = f.owner.submit(attachment);
    await waitFor(() => expect(f.transport.send).toHaveBeenCalledTimes(1));
    f.owner.suspend();
    response.resolve({ kind: "accepted", receipt: f.createReceipt });
    await submit;
    expect(f.owner.getSnapshot().checkpoints).toEqual([]);
    f.owner.setAuthority(authority);
    expect(f.checkpoint().create.receipt).toEqual(f.createReceipt);
    expect(f.checkpoint().links).toHaveLength(0);
    f.authorityReader.mockResolvedValue({ ...authority, role: "viewer" });
    expect(await f.owner.reviewLink(f.checkpoint().id)).toBe(false);
    expect(f.transport.send).toHaveBeenCalledTimes(1);
    f.owner.setAuthority({ ...authority, actorId: targetId });
    expect(f.owner.getSnapshot().checkpoints).toEqual([]);
  });
  it("retains timeout uncertainty across replay rejection and accepts a late receipt monotonically", async () => {
    vi.useFakeTimers();
    const f = fixture(),
      original = deferred<RelatedEvidenceOutcome>(),
      replay = deferred<RelatedEvidenceOutcome>();
    f.transport.send
      .mockReturnValueOnce(original.promise)
      .mockReturnValueOnce(replay.promise);
    const submit = f.owner.submit(attachment);
    await vi.advanceTimersByTimeAsync(0);
    await vi.advanceTimersByTimeAsync(30_000);
    await submit;
    expect(f.checkpoint().create.phase).toBe("uncertain");
    const recovering = f.owner.replay(f.checkpoint().id, f.checkpoint().id);
    await vi.advanceTimersByTimeAsync(0);
    original.resolve({ kind: "accepted", receipt: f.createReceipt });
    await vi.advanceTimersByTimeAsync(0);
    replay.resolve({
      kind: "rejected",
      failure: { kind: "stale_target", message: "Unconfirmed replay rejected" },
    });
    await recovering;
    expect(f.checkpoint().create.phase).toBe("accepted");
    expect(f.checkpoint().create.receipt).toEqual(f.createReceipt);
    expect(f.ids.create).toHaveBeenCalledTimes(1);
    expect(f.checkpoint().links).toHaveLength(0);
  });
  it("detaches presentation on row sheet and inspector changes without rebinding the draft", async () => {
    const f = fixture();
    f.owner.discard();
    let sheet: { kind: "view_schema"; id: string } = {
      kind: "view_schema",
      id: timelineViewSchemaId,
    };
    const hook = renderHook(
      ({ subject }) => useTimelineRelatedEvidenceAttachment(subject),
      {
        initialProps: { subject: f.subject },
        wrapper: ({ children }) => (
          <TimelineRelatedEvidenceContext.Provider
            value={{ owner: f.owner, sheetRef: sheet }}
          >
            {children}
          </TimelineRelatedEvidenceContext.Provider>
        ),
      },
    );
    act(() => {
      hook.result.current.begin(f.feature);
      f.owner.update("evidence.source_party_text", "retained text");
    });
    hook.rerender({
      subject: {
        ...f.subject,
        subject: { ...f.subject.subject, recordId: targetId },
      },
    });
    expect(f.owner.getSnapshot().attachment).toBeNull();
    expect(f.owner.getSnapshot().draft?.source.recordId).toBe(sourceId);
    act(() => f.owner.discard());
    hook.rerender({ subject: f.subject });
    act(() => {
      hook.result.current.begin(f.feature);
    });
    sheet = { kind: "view_schema", id: evidenceViewSchemaId };
    hook.rerender({ subject: f.subject });
    expect(f.owner.getSnapshot().attachment).toBeNull();
    hook.unmount();
    expect(f.owner.getSnapshot().draft).not.toBeNull();
  });
  it("validates lost malformed and uncorrelated responses for each exact stage transport", async () => {
    const f = fixture(),
      port = createTimelineRelatedEvidenceTransport("/api");
    const draft = f.owner.getSnapshot().draft;
    if (!draft) throw new Error("Missing draft");
    for (const stage of ["create", "link"] as const) {
      const attempt = port.capture(
        stage,
        { authority, draft, source: f.source, presentationRevision: 1 },
        `original-${stage}`,
        stage === "link" ? targetId : null,
      );
      const receipt = stage === "create" ? f.createReceipt : f.linkReceipt;
      const fetch = vi
        .fn()
        .mockRejectedValueOnce(new Error("lost"))
        .mockResolvedValueOnce(
          new Response(
            JSON.stringify({ ...receipt, meta: { request_id: 42 } }),
            { status: 200, headers: { "Content-Type": "application/json" } },
          ),
        )
        .mockResolvedValueOnce(
          new Response(
            JSON.stringify({ data: { row: { record_id: targetId } } }),
            { status: 200, headers: { "Content-Type": "application/json" } },
          ),
        )
        .mockResolvedValueOnce(
          new Response(JSON.stringify(receipt), {
            status: 200,
            headers: {
              "Content-Type": "application/json",
              "X-Request-ID": "unrelated",
            },
          }),
        )
        .mockResolvedValueOnce(
          new Response(JSON.stringify(receipt), {
            status: 200,
            headers: {
              "Content-Type": "application/json",
              "X-Request-ID": receipt.meta.request_id,
            },
          }),
        );
      vi.stubGlobal("fetch", fetch);
      for (let count = 0; count < 4; count++)
        expect(await port.send(attempt, new AbortController().signal)).toEqual({
          kind: "uncertain",
        });
      expect(await port.send(attempt, new AbortController().signal)).toEqual({
        kind: "accepted",
        receipt,
        status: 200,
      });
      expect(
        fetch.mock.calls.every(([, init]) => init.body === attempt.body),
      ).toBe(true);
    }
  });
  it("routes collection conflicts and preserves the created checkpoint when the saved collection is kept", async () => {
    const f = fixture();
    const conflict = {
      conflict_token: "token",
      record_id: sourceId,
      field_key: "timeline.attached_evidence_ids",
      conflict_resolution_class: "collection_review" as const,
      base_row_version: 1,
      current_row_version: 2,
      client_value: {},
      server_value: {},
    };
    f.transport.send
      .mockResolvedValueOnce({ kind: "accepted", receipt: f.createReceipt })
      .mockResolvedValueOnce({
        kind: "rejected",
        failure: {
          kind: "same_field_conflict",
          message: "Review collection",
          conflict,
        },
      });
    await f.owner.submit(attachment);
    await waitFor(() => expect(f.effects.conflict).toHaveBeenCalledOnce());
    expect(f.effects.conflict.mock.calls[0]?.[1]).toEqual(conflict);
    f.owner.conflictResolved("token", "keep_saved", {
      data: { view_schema_id: timelineViewSchemaId, row: f.source },
      meta: { request_id: "kept" },
    });
    expect(f.checkpoint().associationPresent).toBe(false);
    expect(f.owner.linkComplete(f.checkpoint())).toBe(false);
    expect(f.checkpoint().resolution?.receipt.meta.request_id).toBe("kept");
    f.owner.conflictChanged("token", {
      ...conflict,
      conflict_token: "renewed-token",
    });
    f.owner.conflictResolved("renewed-token", "use_unsaved", f.linkReceipt);
    expect(f.owner.linkComplete(f.checkpoint())).toBe(true);
    expect(f.transport.send).toHaveBeenCalledTimes(2);
  });
  it("correlates socket and HTTP ordering and rereads newer source and equal-version derived Evidence counts", async () => {
    for (const order of ["socket", "http"] as const) {
      const f = fixture(),
        pending = deferred<RelatedEvidenceOutcome>();
      f.transport.send.mockReturnValueOnce(pending.promise);
      const sending = f.owner.submit(attachment);
      await waitFor(() => expect(f.transport.send).toHaveBeenCalledOnce());
      f.owner.detach(attachment);
      const event: RecordChangedMessage = {
        type: "record_changed",
        incident_id: authority.incidentId,
        event_id: "created-event",
        emitted_at: "2026-09-12T20:00:00Z",
        stream_seq: 1,
        payload: {
          record_id: targetId,
          row_version: 1,
          client_txn_id: f.checkpoint().create.attempt.clientTxnId,
          actor_user_id: actor,
          change_set_id: f.createReceipt.data.change_set_id,
          changed_field_keys: [],
          affected_views: [
            { view_schema_id: evidenceViewSchemaId, change_kind: "invalidate" },
          ],
        },
      };
      if (order === "socket") f.owner.observeSocket(event);
      pending.resolve({ kind: "accepted", receipt: f.createReceipt });
      await sending;
      if (order === "http") f.owner.observeSocket(event);
      f.owner.observeSocket(event);
      await waitFor(() =>
        expect(f.checkpoint().create.refresh).toBe("complete"),
      );
      expect(f.checkpoint().create.observations).toHaveLength(1);
      expect(f.checkpoint().create.receipt).toEqual(f.createReceipt);
      const refresh = deferred<void>();
      f.effects.refresh.mockReturnValueOnce(refresh.promise);
      const reading = f.owner.retryRefresh(f.checkpoint().id);
      await waitFor(() =>
        expect(f.checkpoint().create.refresh).toBe("refreshing"),
      );
      f.current.source = { ...f.source, row_version: 4 };
      f.owner.observe(sourceId, 4);
      f.current.target = {
        ...f.target,
        cells: {
          ...f.target.cells,
          "evidence.linked_record_count": { value: 2 },
        },
      };
      f.owner.observeSocket({
        ...event,
        event_id: "remote-link",
        stream_seq: 2,
        payload: {
          ...event.payload,
          record_id: sourceId,
          row_version: 4,
          client_txn_id: "remote",
          change_set_id: "remote-change",
          affected_views: [
            { view_schema_id: timelineViewSchemaId, change_kind: "invalidate" },
          ],
        },
      });
      refresh.resolve();
      await reading;
      await waitFor(() =>
        expect(f.checkpoint().create.refresh).toBe("complete"),
      );
      expect(f.owner.latestRow(sourceId)?.row_version).toBe(4);
      expect(
        f.owner.latestRow(targetId)?.cells["evidence.linked_record_count"]
          ?.value,
      ).toBe(2);
      expect(
        f.checkpoint().create.receipt?.data.row.cells[
          "evidence.linked_record_count"
        ]?.value,
      ).toBe(0);
      expect(f.transport.send).toHaveBeenCalledOnce();
      f.current.source = null;
      f.owner.observeSocket({
        ...event,
        event_id: "removed",
        stream_seq: 3,
        payload: {
          ...event.payload,
          record_id: sourceId,
          row_version: 5,
          client_txn_id: "delete",
          change_set_id: "delete-change",
          affected_views: [
            { view_schema_id: timelineViewSchemaId, change_kind: "remove" },
          ],
        },
      });
      await waitFor(() => expect(f.checkpoint().sourceUnavailable).toBe(true));
      expect(f.owner.latestRow(sourceId)).toBeNull();
      expect(await f.owner.reviewLink(f.checkpoint().id)).toBe(false);
      f.current.source = { ...f.source, row_version: 6 };
      expect(await f.owner.reviewLink(f.checkpoint().id)).toBe(true);
      f.current.target = null;
      await f.owner.retryRefresh(f.checkpoint().id);
      expect(f.owner.latestRow(targetId)).toBeNull();
      expect(f.checkpoint().evidenceUnavailable).toBe(true);
      expect(f.checkpoint().create.receipt).toEqual(f.createReceipt);
    }
  });
  it("reads the original source through authorized pages and never treats failed paging as absence", async () => {
    const f = fixture(),
      signal = new AbortController().signal;
    const page = vi
      .fn<WorkbookAuthoringReadPort["page"]>()
      .mockResolvedValueOnce({
        kind: "accepted",
        value: { candidates: [], hasMore: true, nextCursor: "next" },
      })
      .mockResolvedValueOnce({
        kind: "accepted",
        value: {
          candidates: [
            {
              recordId: sourceId,
              viewSchemaId: timelineViewSchemaId,
              displayText: "Original",
              row: f.source,
            },
          ],
          hasMore: false,
          nextCursor: null,
        },
      });
    expect(
      await readWorkbookAuthoringRecord(
        { ...f.reader, page },
        timelineViewSchemaId,
        sourceId,
        signal,
      ),
    ).toEqual(f.source);
    expect(page.mock.calls.map(([input]) => input.cursor)).toEqual([
      null,
      "next",
    ]);
    expect(page.mock.calls[0]?.[0].queryState.filters).toEqual([]);
    page.mockResolvedValue({
      kind: "accepted",
      value: { candidates: [], hasMore: true, nextCursor: "loop" },
    });
    await expect(
      readWorkbookAuthoringRecord(
        { ...f.reader, page },
        timelineViewSchemaId,
        sourceId,
        signal,
      ),
    ).rejects.toThrow("paging changed");
    page.mockRejectedValueOnce(new Error("Read access failed"));
    await expect(
      readWorkbookAuthoringRecord(
        { ...f.reader, page },
        timelineViewSchemaId,
        sourceId,
        signal,
      ),
    ).rejects.toThrow("Read access failed");
    expect(f.owner.getSnapshot().authority).toEqual(authority);
  });
});
