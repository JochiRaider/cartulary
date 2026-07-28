import { afterEach, describe, expect, it, vi } from "vitest";
import type { IncidentCollaborationEvent } from "../../collaboration/IncidentCollaborationSession";
import { WorkbookCollaborationProjection } from "./WorkbookCollaborationProjection";
import { WorkbookMutationRuntime } from "./WorkbookMutationRuntime";
import type { WorkbookActiveSurfacePort } from "./workbookSurfacePort";

function projectionFixture(
  initialSheetRef:
    | { readonly kind: "view_schema"; readonly id: string }
    | { readonly kind: "saved_view"; readonly id: string } = {
    kind: "view_schema",
    id: "cartulary.view.timeline.v2",
  },
) {
  let listener: ((event: IncidentCollaborationEvent) => void) | null = null;
  const published: unknown[] = [];
  const mutationRuntime = new WorkbookMutationRuntime({
    clientInstanceId: "client-1",
    incidentId: "incident-1",
  });
  const projection = new WorkbookCollaborationProjection({
    apiBase: undefined,
    initialSheetRef,
    mutationRuntime,
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
    mutationRuntime,
    projection,
    published,
    session,
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

afterEach(() => {
  vi.useRealTimers();
});

describe("WorkbookCollaborationProjection", () => {
  it("keeps base and saved-view presence exact, keyed, sorted, and self-free", () => {
    const fixture = projectionFixture();
    fixture.emit({
      kind: "message",
      message: {
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
      clearAuthorizedRows: vi.fn(),
      refresh: vi.fn(async () => undefined),
    };
    fixture.projection.registerActiveSurface(port);
    fixture.emit({
      kind: "message",
      message: {
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
      clearAuthorizedRows: vi.fn(),
      refresh,
    });
    fixture.emit({
      kind: "message",
      message: {
        type: "record_changed",
        stream_seq: 2,
        payload: {
          record_id: "record-1",
          row_version: 2,
          change_set_id: "change-2",
          client_txn_id: "remote-2",
          actor_user_id: "user-other",
          changed_field_keys: [],
          affected_views: [],
        },
      },
    } as IncidentCollaborationEvent);
    fixture.emit({
      kind: "reset_required",
      generation: 3,
      reason: "sequence_gap",
    });
    await vi.waitFor(() => {
      expect(refresh).toHaveBeenCalledTimes(2);
      expect(fixture.session.completeReset).toHaveBeenCalledWith(3);
    });
    fixture.projection.dispose();
  });

  it("clears protected rows and pauses replay on authorization loss and closure", () => {
    vi.useFakeTimers();
    const fixture = projectionFixture();
    const clearAuthorizedRows = vi.fn();
    fixture.projection.registerActiveSurface({
      identity: {
        sheetRef: {
          kind: "view_schema",
          id: "cartulary.view.timeline.v2",
        },
        viewSchemaId: "cartulary.view.timeline.v2",
      },
      applyRecordChanged: () => ({ kind: "applied" }),
      clearAuthorizedRows,
      refresh: vi.fn(async () => undefined),
    });

    fixture.emit({ kind: "authorization_lost" });
    expect(clearAuthorizedRows).toHaveBeenCalledOnce();
    expect(fixture.mutationRuntime.getSnapshot().authPaused).toBe(true);
    fixture.emit({ kind: "incident_closed" });
    expect(clearAuthorizedRows).toHaveBeenCalledTimes(2);
    fixture.projection.dispose();
  });
});
