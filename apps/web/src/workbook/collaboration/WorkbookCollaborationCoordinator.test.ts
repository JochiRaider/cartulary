import { afterEach, describe, expect, it, vi } from "vitest";
import type { IncidentCollaborationEvent } from "../../collaboration/IncidentCollaborationSession";
import type { AuthorizationRecoveryPort } from "../../shared/authorizationRecovery";
import { createWorkbookPendingMutationAdapter } from "../adapters/createWorkbookPendingMutationAdapter";
import { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import { createWorkbookCollaborationCoordinator } from "./WorkbookCollaborationCoordinator";
import type { WorkbookActiveSurfacePort } from "./workbookSurfacePort";

function manualTiming() {
  let nowMs = 0;
  let nextId = 0;
  const tasks = new Map<
    number,
    { readonly dueAtMs: number; readonly run: () => void }
  >();
  return {
    clock: { nowMs: () => nowMs },
    scheduler: {
      schedule(delayMs: number, run: () => void) {
        const id = ++nextId;
        tasks.set(id, { dueAtMs: nowMs + Math.max(0, delayMs), run });
        return { cancel: () => tasks.delete(id) };
      },
    },
    async advanceBy(durationMs: number) {
      const targetMs = nowMs + durationMs;
      while (true) {
        const next = Array.from(tasks.entries())
          .filter(([, task]) => task.dueAtMs <= targetMs)
          .sort(
            ([leftId, left], [rightId, right]) =>
              left.dueAtMs - right.dueAtMs || leftId - rightId,
          )[0];
        if (next === undefined) break;
        const [id, task] = next;
        tasks.delete(id);
        nowMs = task.dueAtMs;
        task.run();
        await Promise.resolve();
        await Promise.resolve();
      }
      nowMs = targetMs;
      await Promise.resolve();
    },
    pendingTaskCount: () => tasks.size,
  };
}

const transactionIds = {
  create: (prefix: string) => `${prefix}-txn`,
};

const serverEnvelope = {
  emitted_at: "2026-07-13T12:00:00Z",
  event_id: "event-1",
  incident_id: "incident-1",
} as const;

function projectionFixture(
  initialSheetRef:
    | { readonly kind: "view_schema"; readonly id: string }
    | { readonly kind: "saved_view"; readonly id: string } = {
    kind: "view_schema",
    id: "cartulary.view.timeline.v2",
  },
  options: {
    readonly recover?: AuthorizationRecoveryPort["recover"];
  } = {},
) {
  const timing = manualTiming();
  let listener: ((event: IncidentCollaborationEvent) => void) | null = null;
  const cleanupOrder: string[] = [];
  const published: unknown[] = [];
  const mutationRuntime = new WorkbookMutationRuntime(
    {
      clientInstanceId: "client-1",
      incidentId: "incident-1",
    },
    transactionIds,
    createWorkbookPendingMutationAdapter({
      apiBase: undefined,
      incidentId: "incident-1",
    }),
  );
  const invalidateMutation = mutationRuntime.invalidate.bind(mutationRuntime);
  const mutationInvalidation = vi.spyOn(mutationRuntime, "invalidate");
  mutationInvalidation.mockImplementation((reason) => {
    cleanupOrder.push("mutation");
    invalidateMutation(reason);
  });
  const continuityInvalidation = vi.fn(() => {
    cleanupOrder.push("continuity");
  });
  const evidenceInvalidation = vi.fn(() => {
    cleanupOrder.push("evidence");
  });
  const extensionInvalidation = vi.fn(() => {
    cleanupOrder.push("extension");
  });
  const inspectorInvalidation = vi.fn(() => {
    cleanupOrder.push("inspector");
  });
  const onAuthorizationRecovered = vi.fn();
  const onIncidentAccessLost = vi.fn();
  const queryInvalidation = vi.fn(() => {
    cleanupOrder.push("query");
  });
  const projection = createWorkbookCollaborationCoordinator({
    authorizationRecovery: {
      recover:
        options.recover ??
        (async () => ({
          kind: "authorized",
          role: "admin",
          userId: "user-1",
        })),
    },
    clock: timing.clock,
    continuityInvalidation,
    evidenceInvalidation,
    extensionInvalidation,
    incidentId: "incident-1",
    initialSheetRef,
    inspectorInvalidation,
    mutationRuntime,
    onAuthorizationRecovered,
    onIncidentAccessLost,
    queryInvalidation,
    scheduler: timing.scheduler,
  });
  const session = {
    completeReset: vi.fn(() => true),
    connectionId: "self-connection",
    publishPresence: vi.fn((presence: unknown) => {
      published.push(presence);
    }),
    reconnect: vi.fn(),
    status: "connected" as const,
    subscribe: vi.fn((next: (event: IncidentCollaborationEvent) => void) => {
      listener = next;
      return () => {
        listener = null;
      };
    }),
  };
  projection.attachSession(session);
  return {
    emit: (event: IncidentCollaborationEvent) => listener?.(event),
    cleanupOrder,
    continuityInvalidation,
    evidenceInvalidation,
    extensionInvalidation,
    inspectorInvalidation,
    mutationRuntime,
    mutationInvalidation,
    onAuthorizationRecovered,
    onIncidentAccessLost,
    projection,
    published,
    queryInvalidation,
    session,
    timing,
  };
}

function presence(
  connectionId: string,
  sheetRef:
    | { readonly kind: "view_schema"; readonly id: string }
    | { readonly kind: "saved_view"; readonly id: string },
) {
  return {
    connection_id: connectionId,
    user_id: `user-${connectionId}`,
    display_name: `Analyst ${connectionId}`,
    sheet_ref: sheetRef,
    record_id: "record-1",
    mode: "editing",
    field_key: "timeline.activity_synopsis_text",
    observed_at: "2026-07-28T12:00:00Z",
    expires_at: "2026-07-28T12:01:00Z",
  } as const;
}

afterEach(() => vi.restoreAllMocks());

describe("WorkbookCollaborationCoordinator", () => {
  it("keeps base and saved-view presence exact, keyed, sorted, and self-free", () => {
    const fixture = projectionFixture();
    fixture.emit({
      kind: "message",
      message: {
        ...serverEnvelope,
        type: "presence_snapshot",
        payload: {
          presences: [
            presence("self-connection", {
              kind: "view_schema",
              id: "cartulary.view.timeline.v2",
            }),
            presence("z-connection", {
              kind: "view_schema",
              id: "cartulary.view.timeline.v2",
            }),
            presence("a-connection", {
              kind: "view_schema",
              id: "cartulary.view.timeline.v2",
            }),
            presence("saved-connection", {
              kind: "saved_view",
              id: "saved-view-1",
            }),
          ],
        },
      },
    } as IncidentCollaborationEvent);

    expect(
      fixture.projection
        .getSnapshot()
        .activeSheetPresenceRecords.map((entry) => entry.connection_id),
    ).toEqual(["a-connection", "z-connection"]);

    fixture.projection.setActiveSheet({
      kind: "saved_view",
      id: "saved-view-1",
    });
    expect(
      fixture.projection
        .getSnapshot()
        .activeSheetPresenceRecords.map((entry) => entry.connection_id),
    ).toEqual(["saved-connection"]);
    fixture.projection.dispose();
  });

  it.each([
    "cartulary.view.timeline.v2",
    "cartulary.view.hosts.v1",
    "cartulary.view.assessments.v1",
    "cartulary.view.notes.v1",
  ])("routes record changes through the active %s renderer port", (viewId) => {
    const fixture = projectionFixture({ kind: "view_schema", id: viewId });
    const applyRecordChanged = vi.fn(() => ({ kind: "applied" as const }));
    const port: WorkbookActiveSurfacePort = {
      identity: {
        sheetRef: { kind: "view_schema", id: viewId },
        viewSchemaId: viewId,
      },
      applyRecordChanged,
      invalidate: vi.fn(),
      refresh: vi.fn(async () => undefined),
    };
    fixture.projection.registerActiveSurface(port);
    fixture.emit({
      kind: "message",
      message: {
        ...serverEnvelope,
        type: "record_changed",
        stream_seq: 1,
        payload: {
          record_id: "record-1",
          row_version: 2,
          change_set_id: "change-1",
          client_txn_id: "remote-1",
          actor_user_id: "user-other",
          changed_field_keys: ["field.one"],
          affected_views: [
            {
              view_schema_id: viewId,
              change_kind: "patch",
              patch_cells: {
                record_id: "record-1",
                row_version: 2,
                cells: { "field.one": { value: "remote" } },
              },
            },
          ],
        },
      },
    } as IncidentCollaborationEvent);

    expect(applyRecordChanged).toHaveBeenCalledOnce();
    fixture.projection.dispose();
  });

  it("refreshes invalidations and resets before releasing replay", async () => {
    const fixture = projectionFixture();
    const refresh = vi.fn(async () => undefined);
    fixture.projection.registerActiveSurface({
      identity: {
        sheetRef: {
          kind: "view_schema",
          id: "cartulary.view.timeline.v2",
        },
        viewSchemaId: "cartulary.view.timeline.v2",
      },
      applyRecordChanged: () => ({ kind: "refresh_required" }),
      invalidate: vi.fn(),
      refresh,
    });
    fixture.emit({
      kind: "message",
      message: {
        ...serverEnvelope,
        type: "record_changed",
        stream_seq: 2,
        payload: {
          record_id: "record-1",
          row_version: 2,
          change_set_id: "change-2",
          client_txn_id: "remote-2",
          actor_user_id: "user-other",
          changed_field_keys: [],
          affected_views: [
            {
              view_schema_id: "cartulary.view.timeline.v2",
              change_kind: "invalidate",
            },
          ],
        },
      },
    } as IncidentCollaborationEvent);
    fixture.emit({
      kind: "reset_required",
      generation: 3,
      reason: "sequence_gap",
    });
    await Promise.resolve();
    await Promise.resolve();
    expect(refresh).toHaveBeenCalledTimes(2);
    expect(fixture.session.completeReset).toHaveBeenCalledWith(3);
    fixture.projection.dispose();
  });

  it("routes typed invalidation and pauses replay on authorization loss and closure", () => {
    const fixture = projectionFixture();
    const invalidate = vi.fn(() => {
      fixture.cleanupOrder.push("active-query");
    });
    const applyRecordChanged = vi.fn(() => ({ kind: "applied" as const }));
    fixture.projection.registerActiveSurface({
      identity: {
        sheetRef: {
          kind: "view_schema",
          id: "cartulary.view.timeline.v2",
        },
        viewSchemaId: "cartulary.view.timeline.v2",
      },
      applyRecordChanged,
      invalidate,
      refresh: vi.fn(async () => undefined),
    });

    fixture.emit({ kind: "authorization_lost" });
    expect(invalidate).toHaveBeenNthCalledWith(1, {
      kind: "session_unavailable",
    });
    expect(fixture.cleanupOrder).toEqual([
      "mutation",
      "query",
      "active-query",
      "extension",
      "inspector",
      "continuity",
      "evidence",
    ]);
    expect(fixture.mutationRuntime.getSnapshot().authPaused).toBe(true);
    fixture.emit({
      kind: "message",
      message: {
        ...serverEnvelope,
        payload: {
          actor_user_id: "user-other",
          affected_views: [
            {
              change_kind: "invalidate",
              view_schema_id: "cartulary.view.timeline.v2",
            },
          ],
          change_set_id: "change-after-loss",
          changed_field_keys: [],
          client_txn_id: "remote-after-loss",
          record_id: "record-1",
          row_version: 2,
        },
        stream_seq: 2,
        type: "record_changed",
      },
    });
    expect(applyRecordChanged).not.toHaveBeenCalled();
    fixture.cleanupOrder.length = 0;
    fixture.emit({ kind: "incident_closed" });
    expect(invalidate).toHaveBeenNthCalledWith(2, {
      kind: "incident_closed",
    });
    expect(fixture.cleanupOrder).toEqual([
      "mutation",
      "query",
      "active-query",
      "inspector",
      "continuity",
      "evidence",
    ]);
    fixture.projection.dispose();
    fixture.projection.dispose();
  });

  it("recovers authorization through the injected port and keeps role downgrades read-only", async () => {
    const fixture = projectionFixture(
      { kind: "view_schema", id: "cartulary.view.timeline.v2" },
      {
        recover: async () => ({
          kind: "authorized",
          role: "viewer",
          userId: "user-viewer",
        }),
      },
    );
    const refresh = vi.fn(async () => undefined);
    fixture.projection.registerActiveSurface({
      identity: {
        sheetRef: {
          kind: "view_schema",
          id: "cartulary.view.timeline.v2",
        },
        viewSchemaId: "cartulary.view.timeline.v2",
      },
      applyRecordChanged: () => ({ kind: "applied" }),
      invalidate: vi.fn(),
      refresh,
    });

    fixture.emit({ kind: "session_revoked" });
    await fixture.timing.advanceBy(1_000);
    await Promise.resolve();

    expect(fixture.onAuthorizationRecovered).toHaveBeenCalledWith({
      kind: "authorized",
      role: "viewer",
      userId: "user-viewer",
    });
    expect(fixture.inspectorInvalidation).toHaveBeenCalledWith({
      kind: "incident_role_changed",
      role: "viewer",
    });
    expect(refresh).toHaveBeenCalledWith({
      reason: "authorization_recovered",
    });
    expect(fixture.mutationRuntime.getSnapshot().authPaused).toBe(true);
    expect(fixture.session.reconnect).toHaveBeenCalledOnce();
    fixture.projection.dispose();
  });

  it("disposes shell projections without disposing the borrowed mutation runtime", () => {
    const fixture = projectionFixture();
    fixture.projection.dispose();
    expect(fixture.mutationInvalidation).not.toHaveBeenCalledWith({
      kind: "runtime_disposed",
    });
  });

  it("confirms access loss without replay and rejects late recovery after disposal", async () => {
    let resolveRecovery:
      | ((result: { readonly kind: "access_lost" }) => void)
      | undefined;
    const recovery = new Promise<{ readonly kind: "access_lost" }>(
      (resolve) => {
        resolveRecovery = resolve;
      },
    );
    const fixture = projectionFixture(
      { kind: "view_schema", id: "cartulary.view.timeline.v2" },
      {
        recover: () => recovery,
      },
    );

    fixture.emit({ kind: "authorization_lost" });
    await fixture.timing.advanceBy(1_000);
    fixture.projection.dispose();
    fixture.projection.dispose();
    resolveRecovery?.({ kind: "access_lost" });
    await Promise.resolve();
    await Promise.resolve();

    expect(fixture.onIncidentAccessLost).not.toHaveBeenCalled();
    expect(fixture.onAuthorizationRecovered).not.toHaveBeenCalled();
    expect(fixture.session.reconnect).not.toHaveBeenCalled();
  });

  it("clears protected owners again when recovery confirms incident access loss", async () => {
    const fixture = projectionFixture(
      { kind: "view_schema", id: "cartulary.view.timeline.v2" },
      {
        recover: async () => ({ kind: "access_lost" }),
      },
    );

    fixture.emit({ kind: "authorization_lost" });
    await fixture.timing.advanceBy(1_000);
    await Promise.resolve();

    expect(fixture.queryInvalidation).toHaveBeenLastCalledWith({
      kind: "incident_access_lost",
    });
    expect(fixture.extensionInvalidation).toHaveBeenLastCalledWith({
      kind: "incident_access_lost",
    });
    expect(fixture.evidenceInvalidation).toHaveBeenLastCalledWith({
      kind: "incident_access_lost",
    });
    expect(fixture.onIncidentAccessLost).toHaveBeenCalledOnce();
    expect(fixture.session.reconnect).not.toHaveBeenCalled();
    fixture.projection.dispose();
  });

  it("rejects duplicate transactions and reconciles an inactive surface on registration", async () => {
    const fixture = projectionFixture();
    const resolved = vi.fn(
      (clientTxnId: string | null | undefined) =>
        clientTxnId === "resolved-txn",
    );
    fixture.projection.registerClientTxnResolver(resolved);
    const event: IncidentCollaborationEvent = {
      kind: "message",
      message: {
        ...serverEnvelope,
        payload: {
          actor_user_id: "user-other",
          affected_views: [
            {
              change_kind: "invalidate",
              view_schema_id: "cartulary.view.timeline.v2",
            },
          ],
          change_set_id: "change-1",
          changed_field_keys: [],
          client_txn_id: "resolved-txn",
          record_id: "record-1",
          row_version: 2,
        },
        stream_seq: 1,
        type: "record_changed",
      },
    };
    fixture.emit(event);
    expect(resolved).toHaveBeenCalledWith("resolved-txn");

    fixture.projection.registerClientTxnResolver(() => false);
    if (event.kind !== "message" || event.message.type !== "record_changed") {
      throw new Error("invalid record-changed fixture");
    }
    fixture.emit({
      ...event,
      message: {
        ...event.message,
        payload: { ...event.message.payload, client_txn_id: "remote-txn" },
      },
    });
    const refresh = vi.fn(async () => undefined);
    fixture.projection.registerActiveSurface({
      applyRecordChanged: vi.fn(() => ({ kind: "applied" as const })),
      identity: {
        sheetRef: {
          id: "cartulary.view.timeline.v2",
          kind: "view_schema",
        },
        viewSchemaId: "cartulary.view.timeline.v2",
      },
      invalidate: vi.fn(),
      refresh,
    });
    await Promise.resolve();
    expect(refresh).toHaveBeenCalledWith({
      reason: "inactive_surface_reconciliation",
    });
    fixture.projection.dispose();
  });

  it("retries reset refresh and settles only after synchronization succeeds", async () => {
    const fixture = projectionFixture();
    const refresh = vi
      .fn<() => Promise<void>>()
      .mockRejectedValueOnce(new Error("offline"))
      .mockResolvedValueOnce(undefined);
    fixture.projection.registerActiveSurface({
      applyRecordChanged: () => ({ kind: "applied" }),
      identity: {
        sheetRef: {
          id: "cartulary.view.timeline.v2",
          kind: "view_schema",
        },
        viewSchemaId: "cartulary.view.timeline.v2",
      },
      invalidate: vi.fn(),
      refresh,
    });
    fixture.emit({
      generation: 8,
      kind: "reset_required",
      reason: "sequence_gap",
    });
    await Promise.resolve();
    expect(fixture.session.completeReset).not.toHaveBeenCalled();
    expect(fixture.timing.pendingTaskCount()).toBe(1);
    await fixture.timing.advanceBy(1_000);
    expect(refresh).toHaveBeenCalledTimes(2);
    expect(fixture.session.completeReset).toHaveBeenCalledWith(8);
    fixture.projection.dispose();
  });

  it("cancels late authorization recovery when the incident closes", async () => {
    let resolveRecovery:
      | ((result: {
          readonly kind: "authorized";
          readonly role: "admin";
          readonly userId: string;
        }) => void)
      | undefined;
    const recovery = new Promise<{
      readonly kind: "authorized";
      readonly role: "admin";
      readonly userId: string;
    }>((resolve) => {
      resolveRecovery = resolve;
    });
    const fixture = projectionFixture(undefined, { recover: () => recovery });
    fixture.emit({ kind: "authorization_lost" });
    await fixture.timing.advanceBy(1_000);
    fixture.emit({ kind: "incident_closed" });
    resolveRecovery?.({ kind: "authorized", role: "admin", userId: "user-1" });
    await Promise.resolve();
    await Promise.resolve();
    expect(fixture.onAuthorizationRecovered).not.toHaveBeenCalled();
    expect(fixture.session.reconnect).not.toHaveBeenCalled();
    expect(fixture.timing.pendingTaskCount()).toBe(0);
    fixture.projection.dispose();
  });
});
