import {
  assessmentsViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  hostsViewSchemaId,
  requireViewContract,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../../../testing/fetchMockTestSupport";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { createContextualCreateTransport } from "../../adapters/createContextualCreateTransport";
import type { RecordChangedMessage } from "../../collaboration/workbookCollaborationMessages";
import { useInspectorCreateRelatedWorkflow } from "../../inspector/useInspectorCreateRelatedWorkflow";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookSourceWriteSettlement } from "../../ports/WorkbookSourceWriteCoordination";
import { createWorkbookMutationRuntime } from "../../runtime/createWorkbookMutationRuntime";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import { ContextualCreateContext } from "./ContextualCreateContext";
import type {
  ContextualCreateOutcome,
  ContextualCreateReader,
  ContextualCreateReceipt,
  ContextualCreateTransport,
} from "./contextualCreateOperation";
import { WorkbookContextualTaskDecisionCreateOwner } from "./WorkbookContextualTaskDecisionCreateOwner";

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  vi.useRealTimers();
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
function fixture(
  decision = false,
  runtime?: WorkbookMutationRuntime,
  view: string = timelineViewSchemaId,
) {
  let sequence = 0;
  const ids = { create: vi.fn((prefix: string) => `${prefix}-${++sequence}`) };
  const effects = {
    coordinate: vi.fn<() => Promise<WorkbookSourceWriteSettlement>>(
      async () => ({ kind: "settled", minimumRowVersion: 0 }),
    ),
    accepted: vi.fn(),
    refresh: vi.fn(async () => {}),
    observed: vi.fn(),
  };
  const owner =
    runtime?.contextualCreate ??
    new WorkbookContextualTaskDecisionCreateOwner(
      authority.incidentId,
      ids,
      effects,
    );
  owner.setAuthority(authority);
  const contract = requireViewContract(view);
  const feature = contract.inspectorConfig.featureGroups.find(
    (item) =>
      item.featureGroupKey ===
      (decision ? "create_related.decision" : "create_related.task_request"),
  );
  if (!feature) throw new Error("Expected declared feature");
  const subject = {
    cells: {},
    subject: {
      kind: "live" as const,
      recordId: sourceId,
      viewSchemaId: view,
      rowVersion: 1,
      label: "Source",
      surfaceLabel: "Timeline",
    },
  };
  const reader: ContextualCreateReader = {
    availableViews: async () => ({ kind: "accepted" as const, value: [view] }),
    verify: vi.fn(async () => {}),
    page: vi.fn(async () => ({
      kind: "accepted" as const,
      value: {
        candidates: [
          {
            recordId: sourceId,
            displayText: "Source",
            viewSchemaId: contract.viewSchemaId,
            row: fullWorkbookViewRow(contract, sourceId, 1, {}),
          },
        ],
        hasMore: false,
        nextCursor: null,
      },
    })),
  };
  const transport = {
    ...createContextualCreateTransport("/original-api"),
    send: vi.fn<ContextualCreateTransport["send"]>(
      async (): Promise<ContextualCreateOutcome> => ({ kind: "uncertain" }),
    ),
  };
  const authorityReader = vi.fn(
    async (): Promise<WorkbookMutationAuthority> => authority,
  );
  owner.configure(reader, authorityReader, transport);
  owner.begin(subject, feature, { kind: "view_schema", id: view }, attachment);
  if (decision) {
    owner.update("decision.summary", "Reviewed summary");
    owner.update("decision.rationale", "Reviewed rationale");
    owner.update("decision.decision_type", "containment");
  } else {
    owner.update("task.title", "Reviewed task");
    owner.update("task.task_kind", "follow_up");
  }
  const receipt: ContextualCreateReceipt = {
    data: {
      view_schema_id: decision
        ? decisionsViewSchemaId
        : taskRequestsViewSchemaId,
      change_set_id: "50000000-0000-4000-8000-000000000005",
      row: fullWorkbookViewRow(
        requireViewContract(
          decision ? decisionsViewSchemaId : taskRequestsViewSchemaId,
        ),
        targetId,
        1,
        {},
      ),
    },
    meta: { request_id: "request-1" },
  };
  return {
    owner,
    ids,
    effects,
    reader,
    transport,
    authorityReader,
    receipt,
    feature,
    subject,
  };
}
describe("contextual create recovery", () => {
  it("creates from a detached Assessment with real coordination and retains refresh debt and acceptance", async () => {
    for (const debt of [false, true]) {
      const writes = vi.fn(async () => {
        throw new Error("No source write is permitted");
      });
      const runtime = createWorkbookMutationRuntime(
        {
          incidentId: authority.incidentId,
          clientInstanceId: "retained-assessment",
        },
        { create: () => "retained-decision" },
        { execute: writes },
      );
      const { owner, reader, transport, receipt } = fixture(
        true,
        runtime,
        assessmentsViewSchemaId,
      );
      runtime.history.acceptVersion(sourceId, 1);
      // This materialization mismatch alone triggered the old mounted-surface reconciliation.
      expect(runtime.explicitPatches.latestRow(sourceId)).toBeNull();
      if (debt) runtime.retainSurfaceRefreshDebt(assessmentsViewSchemaId);
      const hostsRefresh = vi.fn();
      runtime.registerSurface(hostsViewSchemaId, hostsRefresh);
      owner.detach(attachment);
      const resumed = Symbol("Hosts retained form");
      owner.resume(resumed);
      let failRefresh = false;
      vi.mocked(reader.page).mockImplementation(async ({ viewSchemaId }) => {
        if (failRefresh)
          return {
            kind: "rejected",
            failure: { kind: "retryable", message: "read unavailable" },
          };
        const row =
          viewSchemaId === assessmentsViewSchemaId
            ? fullWorkbookViewRow(
                requireViewContract(viewSchemaId),
                sourceId,
                1,
                {},
              )
            : receipt.data.row;
        return {
          kind: "accepted",
          value: {
            candidates: [
              {
                recordId: row.record_id,
                displayText: "Record",
                viewSchemaId,
                row,
              },
            ],
            hasMore: false,
            nextCursor: null,
          },
        };
      });
      transport.send.mockImplementation(async () => {
        failRefresh = true;
        return { kind: "accepted", receipt };
      });
      await owner.submit(resumed);
      await waitFor(() =>
        expect(owner.getSnapshot().entries[0]?.refresh).toBe("required"),
      );
      expect(transport.send).toHaveBeenCalledOnce();
      expect(owner.getSnapshot().entries[0]?.receipt).toEqual(receipt);
      expect(
        owner.getSnapshot().entries[0]?.attempt.review.draft.source.recordId,
      ).toBe(sourceId);
      failRefresh = false;
      await owner.retryRefresh("retained-decision");
      expect(owner.getSnapshot().entries[0]?.refresh).toBe("complete");
      const sourceRefresh = vi.fn();
      runtime.registerSurface(assessmentsViewSchemaId, sourceRefresh);
      await waitFor(() => expect(sourceRefresh).toHaveBeenCalledOnce());
      expect(hostsRefresh).not.toHaveBeenCalled();
      expect(writes).not.toHaveBeenCalled();
      expect(transport.send).toHaveBeenCalledOnce();
      runtime.invalidate({ kind: "runtime_disposed" });
    }
  });
  it("withdraws contextual readiness for unavailable or changed source reads while retaining values", async () => {
    for (const kind of [
      "missing",
      "incomplete",
      "wrong_identity",
      "changed",
      "stale",
    ] as const) {
      const { owner, reader, transport } = fixture(true);
      vi.mocked(reader.page).mockResolvedValue(
        kind === "incomplete"
          ? {
              kind: "rejected",
              failure: { kind: "retryable", message: "read unavailable" },
            }
          : {
              kind: "accepted",
              value: {
                candidates:
                  kind === "missing"
                    ? []
                    : [
                        {
                          recordId: sourceId,
                          displayText: "Source",
                          viewSchemaId: timelineViewSchemaId,
                          row: fullWorkbookViewRow(
                            requireViewContract(timelineViewSchemaId),
                            kind === "wrong_identity" ? targetId : sourceId,
                            kind === "changed" ? 2 : kind === "stale" ? 0 : 1,
                            {},
                          ),
                        },
                      ],
                hasMore: false,
                nextCursor: null,
              },
            },
      );
      await owner.submit(attachment);
      expect(transport.send).not.toHaveBeenCalled();
      expect(owner.getSnapshot().draft?.values["decision.summary"]).toBe(
        "Reviewed summary",
      );
      expect(owner.getSnapshot().message).toContain("retained");
    }
  });
  it("recovers a timed-out observation with the same attempt and accepts a late original receipt monotonically", async () => {
    vi.useFakeTimers();
    const { owner, transport, receipt, ids } = fixture();
    const original = deferred<ContextualCreateOutcome>(),
      replay = deferred<ContextualCreateOutcome>();
    transport.send
      .mockReturnValueOnce(original.promise)
      .mockReturnValueOnce(replay.promise);
    const submission = owner.submit(attachment);
    await vi.advanceTimersByTimeAsync(0);
    expect(transport.send).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(30_000);
    await submission;
    const id = required(owner.getSnapshot().entries[0]).attempt.clientTxnId;
    expect(owner.getSnapshot().entries[0]).toMatchObject({
      phase: "uncertain",
      transportPending: false,
    });
    const recovering = owner.replay(id);
    await vi.advanceTimersByTimeAsync(0);
    expect(transport.send).toHaveBeenCalledTimes(2);
    expect(ids.create).toHaveBeenCalledTimes(1);
    original.resolve({ kind: "accepted", receipt });
    await vi.advanceTimersByTimeAsync(0);
    replay.resolve({
      kind: "rejected",
      failure: { kind: "authorization_lost", message: "Access changed" },
    });
    await recovering;
    expect(owner.getSnapshot().entries[0]?.receipt).toEqual(receipt);
    expect(owner.getSnapshot().entries[0]?.phase).toBe("accepted");
  });
  it("reserves activation synchronously, waits for prior saves, and captures immutable reviewed requests for both targets", async () => {
    for (const decision of [false, true]) {
      const { owner, effects, ids, transport } = fixture(decision);
      const save = deferred<WorkbookSourceWriteSettlement>();
      effects.coordinate.mockReturnValue(save.promise);
      const first = owner.submit(attachment),
        second = owner.submit(attachment);
      expect(effects.coordinate).toHaveBeenCalledTimes(1);
      expect(transport.send).not.toHaveBeenCalled();
      owner.update(
        decision ? "decision.summary" : "task.title",
        "Unreviewed edit",
      );
      save.resolve({ kind: "settled", minimumRowVersion: 0 });
      await Promise.all([first, second]);
      expect(ids.create).toHaveBeenCalledTimes(1);
      expect(transport.send).toHaveBeenCalledTimes(1);
      const attempt = required(owner.getSnapshot().entries[0]).attempt;
      expect(Object.isFrozen(attempt.request)).toBe(true);
      expect(attempt.body).not.toContain("Unreviewed edit");
      expect(attempt.request).not.toHaveProperty("base_row_version");
      expect(attempt.apiBase).toBe("/original-api");
      expect(attempt.request).toHaveProperty(
        decision ? "decision.support_refs" : "task.linked_record_ids",
      );
    }
  });
  it("invalidates reviewed readiness when source changes while earlier saves are pending", async () => {
    const { owner, effects, transport } = fixture();
    const save = deferred<WorkbookSourceWriteSettlement>();
    effects.coordinate.mockReturnValue(save.promise);
    const submitting = owner.submit(attachment);
    owner.observe(sourceId, 2);
    save.resolve({ kind: "settled", minimumRowVersion: 0 });
    await submitting;
    expect(transport.send).not.toHaveBeenCalled();
    expect(owner.getSnapshot().draft?.values["task.title"]).toBe(
      "Reviewed task",
    );
    expect(owner.getSnapshot().needsReview).toBe(true);
  });
  it("replays the original body and key after loss without accepting edited inputs or regenerating defaults", async () => {
    const { owner, ids, transport, receipt } = fixture();
    await owner.submit(attachment);
    const entry = required(owner.getSnapshot().entries[0]);
    owner.update("task.title", "Another draft");
    owner.discard();
    expect(owner.getSnapshot().draft?.values["task.title"]).toBe(
      "Reviewed task",
    );
    transport.send.mockResolvedValue({ kind: "accepted", receipt });
    await owner.replay(entry.attempt.clientTxnId);
    expect(ids.create).toHaveBeenCalledTimes(1);
    expect(transport.send.mock.calls[1]?.[0]).toBe(
      transport.send.mock.calls[0]?.[0],
    );
    expect(owner.getSnapshot().entries[0]?.receipt).toEqual(receipt);
  });
  it("retains acceptance through detachment and failed refresh, then recovers with reads only", async () => {
    const { owner, transport, reader, effects, receipt, subject, feature } =
      fixture(true);
    const pending = deferred<ContextualCreateOutcome>();
    transport.send.mockReturnValue(pending.promise);
    const submitting = owner.submit(attachment);
    await waitFor(() => expect(transport.send).toHaveBeenCalledTimes(1));
    vi.mocked(reader.page).mockResolvedValue({
      kind: "rejected",
      failure: { kind: "retryable", message: "offline" },
    });
    owner.detach(attachment);
    pending.resolve({ kind: "accepted", receipt });
    await submitting;
    await waitFor(() =>
      expect(owner.getSnapshot().entries[0]?.refresh).toBe("required"),
    );
    expect(effects.accepted).toHaveBeenCalledWith(receipt, expect.any(String));
    expect(
      owner.begin(
        subject,
        feature,
        { kind: "view_schema", id: timelineViewSchemaId },
        attachment,
      ),
    ).toBe(true);
    owner.update("decision.summary", "Replacement draft");
    vi.mocked(reader.page).mockResolvedValue({
      kind: "accepted",
      value: { candidates: [], hasMore: false, nextCursor: null },
    });
    await owner.retryRefresh(
      required(owner.getSnapshot().entries[0]).attempt.clientTxnId,
    );
    expect(transport.send).toHaveBeenCalledTimes(1);
    expect(owner.getSnapshot().entries[0]?.refresh).toBe("complete");
    expect(owner.getSnapshot().draft?.values["decision.summary"]).toBe(
      "Replacement draft",
    );
    expect(reader.page).toHaveBeenCalledWith(
      expect.objectContaining({ viewSchemaId: decisionsViewSchemaId }),
    );
    expect(reader.page).toHaveBeenCalledWith(
      expect.objectContaining({ viewSchemaId: timelineViewSchemaId }),
    );
  });
  it("preserves field rejection and requires deliberate correction and a new attempt", async () => {
    const { owner, transport, ids } = fixture();
    transport.send.mockResolvedValueOnce({
      kind: "rejected",
      failure: {
        kind: "validation",
        message: "Reference unavailable",
        fields: [
          { field: "task.linked_record_ids", message: "Reference unavailable" },
        ],
      },
    });
    await owner.submit(attachment);
    expect(owner.getSnapshot().errors).toEqual({
      "task.linked_record_ids": "Reference unavailable",
    });
    owner.update("task.linked_record_ids", "");
    expect(await owner.review()).toBe(true);
    await owner.submit(attachment);
    expect(ids.create).toHaveBeenCalledTimes(2);
    expect(transport.send.mock.calls[0]?.[0].clientTxnId).not.toBe(
      transport.send.mock.calls[1]?.[0].clientTxnId,
    );
  });
  it("checks current authority on recovery and never treats a replay rejection as proof of no commit", async () => {
    const { owner, transport, authorityReader, reader } = fixture();
    await owner.submit(attachment);
    const id = required(owner.getSnapshot().entries[0]).attempt.clientTxnId;
    authorityReader.mockResolvedValue({ ...authority, role: "viewer" });
    await owner.replay(id);
    expect(transport.send).toHaveBeenCalledTimes(1);
    owner.setAuthority(authority);
    authorityReader.mockResolvedValue(authority);
    vi.mocked(reader.verify).mockRejectedValueOnce(
      new Error("Capability removed"),
    );
    await owner.replay(id);
    expect(transport.send).toHaveBeenCalledTimes(1);
    expect(owner.getSnapshot().entries[0]?.phase).toBe("uncertain");
    transport.send.mockResolvedValue({
      kind: "rejected",
      failure: { kind: "authorization_lost", message: "Unavailable" },
    });
    await owner.replay(id);
    expect(owner.getSnapshot().entries[0]?.phase).toBe("uncertain");
  });
  it("retains late receipt during suspension and ignores completion after incident retirement", async () => {
    for (const retire of [false, true]) {
      const { owner, transport, receipt } = fixture();
      const pending = deferred<ContextualCreateOutcome>();
      transport.send.mockReturnValue(pending.promise);
      const submitting = owner.submit(attachment);
      await waitFor(() => expect(transport.send).toHaveBeenCalledTimes(1));
      if (retire) owner.retire();
      else owner.suspend();
      pending.resolve({ kind: "accepted", receipt });
      await submitting;
      expect(owner.getSnapshot().entries).toEqual([]);
      owner.setAuthority(authority);
      expect(owner.getSnapshot().entries).toHaveLength(retire ? 0 : 1);
    }
  });
  it("accepts receipts outside query pages and never regresses newer HTTP or socket observations", async () => {
    for (const socketFirst of [false, true]) {
      const { owner, transport, receipt } = fixture();
      if (socketFirst) owner.observe(targetId, 3);
      transport.send.mockResolvedValue({ kind: "accepted", receipt });
      await owner.submit(attachment);
      if (!socketFirst) owner.observe(targetId, 3);
      expect(owner.latestVersion(targetId)).toBe(3);
      expect(owner.acceptRow(receipt.data.row)).toBeNull();
      expect(
        owner.getSnapshot().entries[0]?.receipt?.data.row.row_version,
      ).toBe(1);
      owner.acceptRow({ ...receipt.data.row, row_version: 4 });
      expect(owner.latestRow(targetId)?.row_version).toBe(4);
    }
  });
  it("correlates full socket envelopes in either order and preserves removal observations without inferring deletion from pages", async () => {
    for (const first of ["socket", "http"] as const) {
      const { owner, transport, receipt, effects } = fixture();
      const pending = deferred<ContextualCreateOutcome>();
      transport.send.mockReturnValue(pending.promise);
      const sending = owner.submit(attachment);
      await waitFor(() => expect(transport.send).toHaveBeenCalledTimes(1));
      const txn = required(owner.getSnapshot().entries[0]).attempt.clientTxnId;
      const event: RecordChangedMessage = {
        type: "record_changed",
        incident_id: authority.incidentId,
        event_id: "event-1",
        emitted_at: "2026-09-12T20:00:00Z",
        stream_seq: 1,
        payload: {
          record_id: targetId,
          row_version: 1,
          client_txn_id: txn,
          actor_user_id: actor,
          change_set_id: receipt.data.change_set_id,
          changed_field_keys: [],
          affected_views: [
            {
              view_schema_id: taskRequestsViewSchemaId,
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
      expect(owner.getSnapshot().entries[0]?.observations).toEqual([event]);
      owner.observeSocket({
        ...event,
        event_id: "foreign",
        payload: { ...event.payload, actor_user_id: sourceId },
      });
      expect(owner.getSnapshot().entries[0]?.observations).toHaveLength(1);
      await waitFor(() =>
        expect(owner.getSnapshot().entries[0]?.refresh).toBe("complete"),
      );
      owner.observeSocket({
        ...event,
        event_id: "linked-source",
        stream_seq: 2,
        payload: {
          ...event.payload,
          record_id: sourceId,
          row_version: 2,
          affected_views: [
            { view_schema_id: evidenceViewSchemaId, change_kind: "invalidate" },
          ],
        },
      });
      await waitFor(() =>
        expect(effects.refresh).toHaveBeenLastCalledWith(
          expect.anything(),
          expect.arrayContaining([
            evidenceViewSchemaId,
            taskRequestsViewSchemaId,
            timelineViewSchemaId,
          ]),
        ),
      );
      owner.observeSocket({
        ...event,
        event_id: "removed",
        stream_seq: 2,
        payload: {
          ...event.payload,
          row_version: 2,
          client_txn_id: "another-operation",
          affected_views: [
            { view_schema_id: taskRequestsViewSchemaId, change_kind: "remove" },
          ],
        },
      });
      expect(
        owner.acceptRow(
          { ...receipt.data.row, row_version: 2 },
          taskRequestsViewSchemaId,
        ),
      ).toBeNull();
      expect(owner.getSnapshot().entries[0]?.receipt).toEqual(receipt);
      expect(
        owner.acceptRow(
          { ...receipt.data.row, row_version: 3 },
          taskRequestsViewSchemaId,
        )?.row_version,
      ).toBe(3);
    }
    const { owner, transport, receipt, effects } = fixture();
    const pending = deferred<ContextualCreateOutcome>();
    transport.send.mockReturnValue(pending.promise);
    const sending = owner.submit(attachment);
    await waitFor(() => expect(transport.send).toHaveBeenCalledTimes(1));
    const txn = required(owner.getSnapshot().entries[0]).attempt.clientTxnId;
    owner.observeSocket({
      type: "record_changed",
      incident_id: authority.incidentId,
      event_id: "mismatch",
      emitted_at: "2026-09-12T20:00:00Z",
      stream_seq: 1,
      payload: {
        record_id: targetId,
        row_version: 1,
        client_txn_id: txn,
        actor_user_id: actor,
        change_set_id: "unrelated-change-set",
        changed_field_keys: [],
        affected_views: [
          {
            view_schema_id: taskRequestsViewSchemaId,
            change_kind: "invalidate",
          },
        ],
      },
    });
    pending.resolve({ kind: "accepted", receipt });
    await sending;
    expect(owner.getSnapshot().entries[0]?.phase).toBe("uncertain");
    expect(effects.accepted).not.toHaveBeenCalled();
  });
  it("attaches shared inspector forms without using the legacy Timeline create port", async () => {
    const { owner, subject, feature } = fixture();
    owner.discard();
    const { result, rerender, unmount } = renderHook(
      ({ selected }) =>
        useInspectorCreateRelatedWorkflow({
          selectedSubject: selected,
        }),
      {
        initialProps: { selected: subject },
        wrapper: ({ children }) => (
          <ContextualCreateContext.Provider
            value={{
              owner,
              sheetRef: { kind: "view_schema", id: timelineViewSchemaId },
            }}
          >
            {children}
          </ContextualCreateContext.Provider>
        ),
      },
    );
    act(() => {
      result.current.commands.begin(feature);
      owner.update("task.title", "Draft");
    });
    expect(result.current.snapshot.workflow?.targetContract.viewSchemaId).toBe(
      taskRequestsViewSchemaId,
    );
    rerender({
      selected: {
        ...subject,
        subject: { ...subject.subject, recordId: targetId },
      },
    });
    expect(result.current.snapshot.workflow).toBeNull();
    expect(owner.getSnapshot().draft?.source.recordId).toBe(sourceId);
    unmount();
    expect(owner.getSnapshot().draft).not.toBeNull();
  });
  it("classifies lost and malformed responses as uncertain and validates complete receipt correlation", async () => {
    for (const decision of [false, true]) {
      const { owner, receipt } = fixture(decision),
        port = createContextualCreateTransport("/api");
      const attempt = port.capture(
        { authority, draft: required(owner.getSnapshot().draft) },
        "original-key",
      );
      const fetch = vi
        .fn()
        .mockRejectedValueOnce(new Error("lost"))
        .mockResolvedValueOnce(
          new Response(
            JSON.stringify({ data: { row: { record_id: targetId } } }),
            { status: 201, headers: { "Content-Type": "application/json" } },
          ),
        )
        .mockResolvedValueOnce(
          new Response(JSON.stringify(receipt), {
            status: 201,
            headers: {
              "Content-Type": "application/json",
              "X-Request-ID": "other",
            },
          }),
        )
        .mockResolvedValueOnce(
          new Response(JSON.stringify(receipt), {
            status: 200,
            headers: { "Content-Type": "application/json" },
          }),
        );
      vi.stubGlobal("fetch", fetch);
      for (let count = 0; count < 3; count++)
        expect(await port.send(attempt, new AbortController().signal)).toEqual({
          kind: "uncertain",
        });
      expect(await port.send(attempt, new AbortController().signal)).toEqual({
        kind: "accepted",
        receipt,
        status: 200,
      });
      const field = decision
        ? "decision.support_refs"
        : "task.linked_record_ids";
      const rejection = {
        error: {
          code: "invalid_mutation_payload",
          status: 400,
          request_id: "request-error",
          retryable: false,
          message: "invalid mutation payload",
          details: { field, reason_code: "invalid_value" },
        },
      };
      fetch.mockResolvedValueOnce(
        new Response(JSON.stringify(rejection), {
          status: 400,
          headers: {
            "Content-Type": "application/json",
            "X-Request-ID": "unrelated-response",
          },
        }),
      );
      expect(await port.send(attempt, new AbortController().signal)).toEqual({
        kind: "uncertain",
      });
      fetch.mockResolvedValueOnce(
        new Response(JSON.stringify(rejection), {
          status: 400,
          headers: {
            "Content-Type": "application/json",
            "X-Request-ID": "request-error",
          },
        }),
      );
      expect(
        await port.send(attempt, new AbortController().signal),
      ).toMatchObject({
        kind: "rejected",
        failure: {
          kind: "validation",
          fields: [{ field, message: "invalid_value" }],
        },
      });
      expect(
        fetch.mock.calls.every(
          ([, init]) => init.body === attempt.body && init.method === "POST",
        ),
      ).toBe(true);
    }
  });
});

function required<T>(value: T | null | undefined): T {
  if (value == null) throw new Error("Required test value missing");
  return value;
}
