import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import type { WorkbookActiveSurfacePort } from "../../collaboration/workbookSurfacePort";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  createDraftRow,
  normalizeTimelineFullRow,
  rowFromApi,
  type WorkbookRow,
} from "../models/workbookTimelineModel";
import {
  type TimelineCollaborationProjection,
  useTimelineCollaborationBindings,
} from "./useTimelineCollaborationBindings";

const timelineContract = requireViewContract(timelineViewSchemaId);

function committedRow(
  recordId: string,
  rowVersion: number,
  synopsis: string,
): WorkbookRow {
  return rowFromApi(
    normalizeTimelineFullRow(
      fullWorkbookViewRow(timelineContract, recordId, rowVersion, {
        "timeline.activity_synopsis_text": synopsis,
      }),
      "timeline collaboration fixture",
    ),
  );
}

function projectionFixture() {
  let activeSurface: WorkbookActiveSurfacePort | null = null;
  let clientTxnResolver:
    | ((clientTxnId: string | null | undefined) => boolean)
    | null = null;
  const activeSurfaceReleases = vi.fn();
  const resolverReleases = vi.fn();
  const publishPresence = vi.fn();
  const requestAuthorizationRecovery = vi.fn();
  const snapshot = {
    activeSheetPresenceRecords: [],
    connectionId: "connection-1",
    status: "connected" as const,
  };
  const projection: TimelineCollaborationProjection = {
    getSnapshot: () => snapshot,
    publishPresence,
    registerActiveSurface: (port) => {
      activeSurface = port;
      return () => {
        if (activeSurface === port) activeSurface = null;
        activeSurfaceReleases();
      };
    },
    registerClientTxnResolver: (resolver) => {
      clientTxnResolver = resolver;
      return () => {
        if (clientTxnResolver === resolver) clientTxnResolver = null;
        resolverReleases();
      };
    },
    requestAuthorizationRecovery,
    subscribe: () => () => undefined,
  };
  return {
    activeSurface: () => activeSurface,
    activeSurfaceReleases,
    clientTxnResolver: () => clientTxnResolver,
    projection,
    publishPresence,
    requestAuthorizationRecovery,
    resolverReleases,
  };
}

describe("useTimelineCollaborationBindings", () => {
  it("owns Timeline live admission, gap refresh, access invalidation, surface switches, and teardown", async () => {
    const fixture = projectionFixture();
    const draft = createDraftRow(1);
    const rowsRef = {
      current: [committedRow("record-1", 1, "Before"), draft],
    };
    const setRows = vi.fn();
    const beginRowsLoad = vi.fn();
    const refreshRows = vi.fn(async () => undefined);
    let knownVersion = 1;
    const acceptRecordVersion = vi.fn(
      (_recordId: string, rowVersion: number) => {
        knownVersion = Math.max(knownVersion, rowVersion);
      },
    );
    const acceptCommittedRow = vi.fn((row: WorkbookRow) => ({ row }));
    const resolveClientTxn = vi.fn((clientTxnId) => clientTxnId === "local-1");
    const { result, rerender, unmount } = renderHook(
      ({ sheetId }) =>
        useTimelineCollaborationBindings({
          activeSheetRef: { kind: "saved_view", id: sheetId },
          admission: {
            acceptCommittedRow,
            acceptRecordVersion,
            isStaleRecordVersion: (_recordId, rowVersion) =>
              rowVersion < knownVersion,
          },
          beginRowsLoad,
          collaborationProjection: fixture.projection,
          refreshRows,
          resolveClientTxn,
          rowsRef,
          setRows,
        }),
      { initialProps: { sheetId: "saved-view-1" } },
    );

    expect(fixture.activeSurface()?.identity).toEqual({
      sheetRef: { kind: "saved_view", id: "saved-view-1" },
      viewSchemaId: timelineViewSchemaId,
    });
    expect(fixture.clientTxnResolver()?.("local-1")).toBe(true);

    act(() => {
      expect(
        fixture.activeSurface()?.applyRecordChanged({
          record_id: "record-1",
          row_version: 2,
          change_set_id: "change-1",
          client_txn_id: "remote-1",
          actor_user_id: "user-2",
          changed_field_keys: ["timeline.activity_synopsis_text"],
          affected_views: [
            {
              view_schema_id: timelineViewSchemaId,
              change_kind: "patch",
              patch_cells: {
                record_id: "record-1",
                row_version: 2,
                cells: {
                  "timeline.activity_synopsis_text": { value: "After" },
                },
              },
            },
          ],
        }),
      ).toEqual({ kind: "applied" });
    });
    expect(rowsRef.current[0]?.values.activitySynopsisText).toBe("After");
    expect(rowsRef.current[1]).toBe(draft);

    expect(
      fixture.activeSurface()?.applyRecordChanged({
        record_id: "record-1",
        row_version: 1,
        change_set_id: "change-stale",
        client_txn_id: "remote-stale",
        actor_user_id: "user-2",
        changed_field_keys: [],
        affected_views: [],
      }),
    ).toEqual({ kind: "stale" });

    fixture
      .activeSurface()
      ?.invalidate({ kind: "collaboration_reset_required" });
    expect(beginRowsLoad).toHaveBeenCalledOnce();
    expect(rowsRef.current).toHaveLength(2);
    await fixture.activeSurface()?.refresh({ reason: "sequence_gap" });
    expect(refreshRows).toHaveBeenCalledOnce();

    fixture.activeSurface()?.invalidate({ kind: "incident_access_lost" });
    expect(rowsRef.current).toEqual([draft]);
    expect(setRows).toHaveBeenLastCalledWith([draft]);

    act(() => {
      result.current.commands.publishPresence({
        fieldKey: null,
        mode: "viewing",
        recordId: "record-1",
      });
      result.current.commands.requestAuthorizationRecovery();
    });
    expect(fixture.publishPresence).toHaveBeenCalledOnce();
    expect(fixture.requestAuthorizationRecovery).toHaveBeenCalledOnce();

    rerender({ sheetId: "saved-view-2" });
    expect(fixture.activeSurface()?.identity.sheetRef).toEqual({
      kind: "saved_view",
      id: "saved-view-2",
    });
    expect(fixture.activeSurfaceReleases).toHaveBeenCalledOnce();

    unmount();
    expect(fixture.activeSurface()).toBeNull();
    expect(fixture.clientTxnResolver()).toBeNull();
    expect(fixture.activeSurfaceReleases).toHaveBeenCalledTimes(2);
    expect(fixture.resolverReleases).toHaveBeenCalledOnce();
  });

  it("fails closed to refresh for malformed or target-inconsistent sparse patches", () => {
    const fixture = projectionFixture();
    const rowsRef = { current: [committedRow("record-1", 1, "Before")] };
    const { unmount } = renderHook(() =>
      useTimelineCollaborationBindings({
        activeSheetRef: { kind: "view_schema", id: timelineViewSchemaId },
        admission: {
          acceptCommittedRow: (row) => ({ row }),
          acceptRecordVersion: () => undefined,
          isStaleRecordVersion: () => false,
        },
        beginRowsLoad: () => undefined,
        collaborationProjection: fixture.projection,
        refreshRows: async () => undefined,
        resolveClientTxn: () => false,
        rowsRef,
        setRows: () => undefined,
      }),
    );

    expect(
      fixture.activeSurface()?.applyRecordChanged({
        record_id: "record-1",
        row_version: 2,
        change_set_id: "change-invalid",
        client_txn_id: "remote-invalid",
        actor_user_id: "user-2",
        changed_field_keys: ["timeline.activity_synopsis_text"],
        affected_views: [
          {
            view_schema_id: timelineViewSchemaId,
            change_kind: "patch",
            patch_cells: {
              record_id: "record-other",
              row_version: 2,
              cells: {
                "timeline.activity_synopsis_text": { value: "Wrong row" },
              },
            },
          },
        ],
      }),
    ).toEqual({ kind: "refresh_required" });
    expect(rowsRef.current[0]?.values.activitySynopsisText).toBe("Before");
    unmount();
  });
});
