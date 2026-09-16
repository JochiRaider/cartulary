import { requireViewContract } from "@cartulary/view-contracts";
import { cleanup, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../../../testing/fetchMockTestSupport";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { createCoordinationCreateTransport } from "../../adapters/createCoordinationCreateTransport";
import type { RecordChangedMessage } from "../../collaboration/workbookCollaborationMessages";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookAuthoringReadPort } from "../../ports/WorkbookAuthoringReadPort";
import type { WorkbookSourceWriteSettlement } from "../../ports/WorkbookSourceWriteCoordination";
import {
  type CoordinationVariant,
  coordinationFeature,
  coordinationSourceViews,
  coordinationTarget,
  coordinationVariants,
} from "./coordinationCreateModel";
import type {
  CoordinationOutcome,
  CoordinationReceipt,
  CoordinationTransport,
} from "./coordinationCreateOperation";
import { WorkbookCoordinationCreateOwner } from "./WorkbookCoordinationCreateOwner";

afterEach(() => {
  cleanup();
  vi.useRealTimers();
  vi.unstubAllGlobals();
});
const authority: WorkbookMutationAuthority = {
  actorId: "10000000-0000-4000-8000-000000000001",
  incidentId: "40000000-0000-4000-8000-000000000004",
  role: "editor",
  closed: false,
  sessionIdentity: "session",
};
const sourceId = "20000000-0000-4000-8000-000000000002",
  artifactId = "30000000-0000-4000-8000-000000000003";
const token = Symbol("test");
function fixture(
  variant: CoordinationVariant = "lesson",
  view = "cartulary.view.timeline.v2",
) {
  let sequence = 0;
  const ids = { create: vi.fn(() => `coordination-create-${++sequence}`) };
  const effects = {
    coordinate: vi.fn<() => Promise<WorkbookSourceWriteSettlement>>(
      async () => ({ kind: "settled", minimumRowVersion: 0 }),
    ),
    accepted: vi.fn(),
    refresh: vi.fn(async () => {}),
  };
  const owner = new WorkbookCoordinationCreateOwner(
    authority.incidentId,
    ids,
    effects,
  );
  owner.setAuthority(authority);
  const reader: WorkbookAuthoringReadPort = {
    availableViews: vi.fn(async () => ({
      kind: "accepted" as const,
      value: coordinationSourceViews(variant),
    })),
    verify: vi.fn(async () => {}),
    page: vi.fn(async (input) => ({
      kind: "accepted" as const,
      value: {
        candidates:
          input.viewSchemaId === view
            ? [
                {
                  recordId: sourceId,
                  displayText: "Source",
                  viewSchemaId: view,
                  row: { record_id: sourceId, row_version: 1, cells: {} },
                },
              ]
            : input.viewSchemaId === "incident_members"
              ? [
                  {
                    recordId: authority.actorId,
                    displayText: "Member",
                    viewSchemaId: "incident_members",
                  },
                ]
              : [],
        hasMore: false,
        nextCursor: null,
      },
    })),
  };
  const authorityReader = vi.fn(
    async (): Promise<WorkbookMutationAuthority> => authority,
  );
  const transport = {
    ...createCoordinationCreateTransport("/original-api"),
    send: vi.fn<CoordinationTransport["send"]>(async () => ({
      kind: "uncertain",
    })),
  };
  owner.configure(reader, authorityReader, transport);
  const subject = {
    subject: {
      kind: "live" as const,
      recordId: sourceId,
      viewSchemaId: view,
      rowVersion: 1,
      label: "Source",
      surfaceLabel: requireViewContract(view).title,
    },
    cells: {},
  };
  const feature = required(coordinationFeature(view, variant));
  const sheet = { kind: "view_schema" as const, id: view };
  owner.begin(subject, feature, sheet, token);
  const target = coordinationTarget(variant);
  const fields: Record<string, string> = {};
  for (const key of target.minimumCreateFieldSets[0] ?? []) {
    const field = target.fieldMap[key];
    fields[key] = key.endsWith("_user_id")
      ? authority.actorId
      : (field?.enumValues?.[0] ?? "Authored text");
    owner.update(key, fields[key] as string);
  }
  const receipt: CoordinationReceipt = {
    data: {
      view_schema_id: target.viewSchemaId as "cartulary.view.lesson.v1",
      source_record_id: sourceId,
      link_type: "references_artifact",
      change_set_id: "50000000-0000-4000-8000-000000000005",
      row: fullWorkbookViewRow(target, artifactId, 1, fields),
    },
    meta: { request_id: "request-1" },
  };
  return {
    owner,
    reader,
    authorityReader,
    transport,
    ids,
    effects,
    receipt,
    subject,
    feature,
    sheet,
  };
}
describe("Coordination atomic recovery", () => {
  it("freezes all twelve routes and guards same-frame activation before preparation", async () => {
    let count = 0;
    for (const variant of coordinationVariants)
      for (const view of coordinationSourceViews(variant)) {
        count++;
        const { owner, effects, transport, ids } = fixture(variant, view);
        const gate = deferred<WorkbookSourceWriteSettlement>();
        effects.coordinate.mockReturnValue(gate.promise);
        const first = owner.submit(token),
          second = owner.submit(token);
        expect(owner.busy).toBe(true);
        expect(effects.coordinate).toHaveBeenCalledOnce();
        gate.resolve({ kind: "settled", minimumRowVersion: 0 });
        await Promise.all([first, second]);
        expect(transport.send).toHaveBeenCalledOnce();
        expect(ids.create).toHaveBeenCalledOnce();
        const attempt = required(transport.send.mock.calls[0]?.[0]);
        expect(attempt.operationID).toBe("createViewRow");
        expect(attempt.pathParameters).toEqual({
          incident_id: authority.incidentId,
          view_schema_id: coordinationTarget(variant).viewSchemaId,
        });
        expect(attempt.request).toHaveProperty(
          "coordination.source_record_id",
          sourceId,
        );
        expect(attempt.body).toBe(JSON.stringify(attempt.request));
        expect(Object.isFrozen(attempt.review.draft.values)).toBe(true);
        expect(attempt.apiBase).toBe("/original-api");
      }
    expect(count).toBe(12);
  });
  it("replays the unchanged attempt after response loss despite closure deletion and transport reconfiguration", async () => {
    const { owner, transport, receipt, reader, authorityReader, ids } =
      fixture();
    await owner.submit(token);
    const attempt = required(owner.getSnapshot().entries[0]).attempt;
    owner.detach(token);
    owner.update("lesson.summary", "Changed");
    owner.changeSource(null);
    owner.discard();
    expect(owner.getSnapshot().draft?.values["lesson.summary"]).toBe(
      "Authored text",
    );
    owner.closeIncident();
    authorityReader.mockResolvedValue({ ...authority, closed: true });
    vi.mocked(reader.page).mockResolvedValue({
      kind: "accepted",
      value: { candidates: [], hasMore: false, nextCursor: null },
    });
    const replacement = {
      ...createCoordinationCreateTransport("/new-api"),
      send: vi.fn<CoordinationTransport["send"]>(async () => ({
        kind: "accepted",
        receipt,
      })),
    };
    owner.configure(reader, authorityReader, replacement);
    const reads = vi.mocked(reader.verify).mock.calls.length;
    await owner.replay(attempt.clientTxnId);
    expect(replacement.send.mock.calls[0]?.[0]).toBe(attempt);
    expect(ids.create).toHaveBeenCalledOnce();
    expect(reader.verify).toHaveBeenCalledTimes(reads);
    expect(owner.getSnapshot().entries[0]?.receipt).toEqual(receipt);
    expect(transport.send).toHaveBeenCalledOnce();
  });
  it("retains late acceptance after detachment and ignores an obsolete replay rejection", async () => {
    vi.useFakeTimers();
    const { owner, transport, receipt, effects } = fixture();
    const first = deferred<CoordinationOutcome>(),
      replay = deferred<CoordinationOutcome>();
    transport.send
      .mockReturnValueOnce(first.promise)
      .mockReturnValueOnce(replay.promise);
    const submit = owner.submit(token);
    await vi.advanceTimersByTimeAsync(30_000);
    await submit;
    expect(owner.getSnapshot().entries[0]?.phase).toBe("uncertain");
    owner.detach(token);
    const recovering = owner.replay(
      required(owner.getSnapshot().entries[0]).attempt.clientTxnId,
    );
    await vi.advanceTimersByTimeAsync(0);
    first.resolve({ kind: "accepted", receipt });
    await vi.advanceTimersByTimeAsync(0);
    replay.resolve({
      kind: "rejected",
      failure: { kind: "authorization_lost", message: "Obsolete" },
    });
    await recovering;
    expect(owner.getSnapshot().entries[0]?.receipt).toEqual(receipt);
    expect(owner.getSnapshot().entries[0]?.phase).toBe("accepted");
    expect(effects.accepted).toHaveBeenCalledOnce();
  });
  it("keeps malformed or mismatched association receipts uncertain and replay rejection cannot disprove commit", async () => {
    for (const change of [
      "source",
      "link",
      "target",
      "row",
      "meta",
      "change",
    ]) {
      const { owner, transport, receipt, effects } = fixture();
      const broken = structuredClone(receipt);
      if (change === "source" && "source_record_id" in broken.data)
        broken.data.source_record_id = artifactId;
      if (change === "link") Reflect.deleteProperty(broken.data, "link_type");
      if (change === "target")
        Reflect.set(broken.data, "view_schema_id", "cartulary.view.notes.v1");
      if (change === "row") Reflect.deleteProperty(broken.data.row, "cells");
      if (change === "meta") broken.meta.request_id = "";
      if (change === "change") broken.data.change_set_id = "";
      transport.send
        .mockResolvedValueOnce({ kind: "accepted", receipt: broken })
        .mockResolvedValueOnce({
          kind: "rejected",
          failure: {
            kind: "validation",
            message: "Cannot validate fresh source",
          },
        });
      await owner.submit(token);
      expect(owner.getSnapshot().entries[0]?.phase).toBe("uncertain");
      await owner.replay(
        required(owner.getSnapshot().entries[0]).attempt.clientTxnId,
      );
      expect(owner.getSnapshot().entries[0]?.phase).toBe("uncertain");
      expect(effects.accepted).not.toHaveBeenCalled();
    }
  });
  it("separates accepted receipts from refresh debt and preserves a newer draft during read-only recovery", async () => {
    const {
      owner,
      transport,
      receipt,
      effects,
      reader,
      subject,
      feature,
      sheet,
    } = fixture();
    effects.refresh.mockRejectedValueOnce(new Error("refresh failure"));
    transport.send.mockResolvedValue({ kind: "accepted", receipt });
    await owner.submit(token);
    await waitFor(() =>
      expect(owner.getSnapshot().entries[0]?.refresh).toBe("required"),
    );
    const id = required(owner.getSnapshot().entries[0]).attempt.clientTxnId;
    expect(owner.getSnapshot().entries[0]?.receipt).toEqual(receipt);
    expect(owner.begin(subject, feature, sheet, token)).toBe(true);
    owner.update("lesson.summary", "New draft");
    const readCount = vi.mocked(reader.page).mock.calls.length;
    await owner.retryRefresh(id);
    expect(owner.getSnapshot().entries[0]?.refresh).toBe("complete");
    expect(reader.page).toHaveBeenCalledTimes(readCount + 2);
    expect(transport.send).toHaveBeenCalledOnce();
    expect(owner.getSnapshot().draft?.values["lesson.summary"]).toBe(
      "New draft",
    );
  });
  it("allocates a new identity for corrected fresh requests and validates explicit unlinked receipts", async () => {
    const { owner, transport, receipt, ids } = fixture();
    transport.send.mockResolvedValueOnce({
      kind: "rejected",
      failure: {
        kind: "validation",
        message: "Correct summary",
        fields: [
          { field: "lesson.summary", message: "invalid_value" },
          { field: "undeclared", message: "Invalid" },
        ],
      },
    });
    await owner.submit(token);
    expect(owner.getSnapshot().errors).toEqual({
      "lesson.summary":
        "This value could not be accepted. Review it and try again.",
    });
    owner.update("lesson.summary", "Corrected");
    owner.changeSource(null);
    const unlinked: CoordinationReceipt = {
      ...receipt,
      data: {
        view_schema_id: receipt.data.view_schema_id,
        change_set_id: receipt.data.change_set_id,
        row: receipt.data.row,
      },
    };
    transport.send.mockResolvedValueOnce({
      kind: "accepted",
      receipt: unlinked,
    });
    await owner.submit(token);
    expect(ids.create).toHaveBeenCalledTimes(2);
    expect(transport.send.mock.calls[1]?.[0].request).toHaveProperty(
      "coordination.source_record_id",
      null,
    );
    expect(owner.getSnapshot().entries[1]?.receipt).toEqual(unlinked);
    expect(transport.send.mock.calls[0]?.[0].request).toHaveProperty(
      "coordination.source_record_id",
      sourceId,
    );
  });
  it("blocks source conflicts and stale review without saving unrelated authoring", async () => {
    const { owner, transport, effects, reader } = fixture();
    effects.coordinate.mockResolvedValueOnce({
      kind: "blocked",
      reason: "pending_recovery",
    });
    await owner.submit(token);
    expect(transport.send).not.toHaveBeenCalled();
    owner.registerSourceCoordinator(async () => false);
    await owner.submit(token);
    expect(transport.send).not.toHaveBeenCalled();
    owner.registerSourceCoordinator(async () => true);
    vi.mocked(reader.page).mockResolvedValue({
      kind: "accepted",
      value: {
        candidates: [
          {
            recordId: sourceId,
            displayText: "Source",
            viewSchemaId: "cartulary.view.timeline.v2",
            row: { record_id: sourceId, row_version: 2, cells: {} },
          },
        ],
        hasMore: false,
        nextCursor: null,
      },
    });
    await owner.submit(token);
    expect(owner.getSnapshot().needsReview).toBe(true);
    expect(transport.send).not.toHaveBeenCalled();
    await owner.review();
    expect(owner.getSnapshot().needsReview).toBe(false);
    await owner.submit(token);
    expect(transport.send).toHaveBeenCalledOnce();
  });
  it("correlates socket and HTTP acceptance in either order and refreshes late affected views", async () => {
    for (const first of ["socket", "http"] as const) {
      const { owner, transport, receipt, effects } = fixture();
      const pending = deferred<CoordinationOutcome>();
      transport.send.mockReturnValue(pending.promise);
      const sending = owner.submit(token);
      await waitFor(() => expect(transport.send).toHaveBeenCalledOnce());
      const txn = required(owner.getSnapshot().entries[0]).attempt.clientTxnId;
      const event: RecordChangedMessage = {
        type: "record_changed",
        incident_id: authority.incidentId,
        event_id: "created",
        emitted_at: "2026-09-13T14:00:00Z",
        stream_seq: 1,
        payload: {
          record_id: artifactId,
          row_version: 1,
          client_txn_id: txn,
          actor_user_id: authority.actorId,
          change_set_id: receipt.data.change_set_id,
          changed_field_keys: [],
          affected_views: [
            {
              view_schema_id: "cartulary.view.lesson.v1",
              change_kind: "invalidate",
            },
          ],
        },
      };
      if (first === "socket") owner.observeSocket(event);
      pending.resolve({ kind: "accepted", receipt });
      await sending;
      if (first === "http") owner.observeSocket(event);
      owner.observeSocket(event);
      owner.observeSocket({
        ...event,
        event_id: "other-actor",
        payload: { ...event.payload, actor_user_id: sourceId },
      });
      expect(owner.getSnapshot().entries[0]?.observations).toEqual([event]);
      await waitFor(() =>
        expect(owner.getSnapshot().entries[0]?.refresh).toBe("complete"),
      );
      owner.observeSocket({
        ...event,
        event_id: "late",
        stream_seq: 2,
        payload: {
          ...event.payload,
          affected_views: [
            {
              view_schema_id: "cartulary.view.evidence.v1",
              change_kind: "invalidate",
            },
          ],
        },
      });
      await waitFor(() =>
        expect(effects.refresh).toHaveBeenLastCalledWith(
          expect.arrayContaining(["cartulary.view.evidence.v1"]),
          expect.anything(),
        ),
      );
      expect(effects.accepted).toHaveBeenCalledOnce();
      expect(transport.send).toHaveBeenCalledOnce();
    }
  });
  it("retains deletion observations against stale reads until explicit source clear or recovered review", async () => {
    const { owner, transport, reader } = fixture();
    const event: RecordChangedMessage = {
      type: "record_changed",
      incident_id: authority.incidentId,
      event_id: "removed",
      emitted_at: "2026-09-13T14:00:00Z",
      stream_seq: 1,
      payload: {
        record_id: sourceId,
        row_version: 1,
        client_txn_id: "source-delete",
        actor_user_id: authority.actorId,
        change_set_id: "50000000-0000-4000-8000-000000000005",
        changed_field_keys: [],
        affected_views: [
          {
            view_schema_id: "cartulary.view.timeline.v2",
            change_kind: "remove",
          },
        ],
      },
    };
    owner.observeSocket(event);
    expect(owner.getSnapshot().needsReview).toBe(true);
    await owner.review();
    expect(owner.getSnapshot().message).toContain("Source is unavailable");
    await owner.submit(token);
    expect(transport.send).not.toHaveBeenCalled();
    owner.observeSocket({
      ...event,
      event_id: "restored",
      stream_seq: 2,
      payload: {
        ...event.payload,
        row_version: 2,
        affected_views: [
          {
            view_schema_id: "cartulary.view.timeline.v2",
            change_kind: "invalidate",
          },
        ],
      },
    });
    await owner.review();
    expect(owner.getSnapshot().message).toContain("projection is behind");
    expect(owner.getSnapshot().draft?.source?.rowVersion).toBe(1);
    const original = vi.mocked(reader.page).getMockImplementation();
    vi.mocked(reader.page).mockImplementation(async (input) => {
      const result = await required(original)(input);
      if (result.kind !== "accepted") return result;
      return {
        ...result,
        value: {
          ...result.value,
          candidates: result.value.candidates.map((c) =>
            c.row ? { ...c, row: { ...c.row, row_version: 2 } } : c,
          ),
        },
      };
    });
    await owner.review();
    expect(owner.getSnapshot().needsReview).toBe(false);
    expect(owner.getSnapshot().draft?.source?.rowVersion).toBe(2);
    owner.observeSocket(event);
    expect(owner.getSnapshot().needsReview).toBe(false);
    owner.changeSource(null);
    await owner.submit(token);
    expect(transport.send.mock.calls[0]?.[0].request).toHaveProperty(
      "coordination.source_record_id",
      null,
    );
  });
  it("blocks uncertain authority and late preparation across closure reopen and role loss", async () => {
    const { owner, authorityReader, transport, reader } = fixture();
    authorityReader.mockRejectedValueOnce(
      new Error("Authorization read failed"),
    );
    await owner.submit(token);
    expect(transport.send).not.toHaveBeenCalled();
    owner.closeIncident();
    expect(owner.canSubmit()).toBe(false);
    owner.setAuthority(authority);
    expect(owner.getSnapshot().needsReview).toBe(true);
    await owner.submit(token);
    expect(transport.send).not.toHaveBeenCalled();
    await owner.review();
    const gate = deferred<void>();
    vi.mocked(reader.verify).mockReturnValue(gate.promise);
    const sending = owner.submit(token);
    await waitFor(() => expect(owner.busy).toBe(true));
    owner.setAuthority({ ...authority, role: "viewer" });
    gate.resolve();
    await sending;
    expect(transport.send).not.toHaveBeenCalled();
    expect(owner.getSnapshot().draft?.values["lesson.summary"]).toBe(
      "Authored text",
    );
  });
  it("fences account replacement and conceals late acceptance through incident-scoped suspension", async () => {
    const first = fixture();
    const access = deferred<WorkbookMutationAuthority>();
    first.authorityReader.mockReturnValue(access.promise);
    const preparing = first.owner.submit(token);
    first.owner.setAuthority({ ...authority, actorId: artifactId });
    access.resolve(authority);
    await preparing;
    expect(first.transport.send).not.toHaveBeenCalled();
    expect(first.owner.getSnapshot().draft).toBeNull();
    const { owner, transport, receipt, effects } = fixture();
    const pending = deferred<CoordinationOutcome>();
    transport.send.mockReturnValue(pending.promise);
    const submitting = owner.submit(token);
    await waitFor(() => expect(transport.send).toHaveBeenCalledOnce());
    owner.suspend();
    pending.resolve({ kind: "accepted", receipt });
    await submitting;
    expect(owner.getSnapshot().entries).toEqual([]);
    expect(effects.accepted).not.toHaveBeenCalled();
    owner.setAuthority({ ...authority, sessionIdentity: "renewed" });
    expect(owner.getSnapshot().entries[0]?.receipt).toEqual(receipt);
    expect(owner.getSnapshot().attachment).toBeNull();
  });
});
function required<T>(value: T | null | undefined): T {
  if (value == null) throw new Error("Required fixture");
  return value;
}
