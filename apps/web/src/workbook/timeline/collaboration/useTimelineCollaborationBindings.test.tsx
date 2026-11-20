import { requireViewContract } from "@cartulary/view-contracts";
import { act } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import {
  acceptedQueryMetadata,
  fullWorkbookViewRow,
} from "../../../testing/workbookFixtureData";
import { renderHookWithWorkbookQueryBrowsing as renderHook } from "../../../testing/workbookQueryTestSupport";
import type { RecordChangedPayload } from "../../collaboration/workbookCollaborationMessages";
import { emptyWorkbookPresence } from "../../collaboration/workbookPresencePresentation";
import type { WorkbookActiveSurfacePort } from "../../collaboration/workbookSurfacePort";
import {
  emptyWorkbookQueryState,
  type WorkbookQueryState,
} from "../../models/workbookQuery";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { useWorkbookQueryBrowser } from "../../query/WorkbookQueryBrowsingContext";
import type { WorkbookReadScope } from "../../query/WorkbookQueryRow";
import type { WorkbookViewQueryPort } from "../../query/WorkbookViewQueryPort";
import { workbookRowIsAdmissible } from "../../query/workbookRowObservation";
import {
  createDraftRow,
  normalizeTimelineFullRow,
  rowFromApi,
  type WorkbookRow,
} from "../models/timelineRowModel";
import {
  type TimelineCollaborationProjection,
  useTimelineCollaborationBindings,
} from "./useTimelineCollaborationBindings";

const timelineContract = requireViewContract(timelineViewSchemaId);
const acceptedScope: WorkbookReadScope = {
  actorId: "actor-1",
  sessionIdentity: "session-1",
  incidentId: "incident-1",
  epoch: 1,
};

function committedRow(
  recordId: string,
  rowVersion: number,
  synopsis: string,
): WorkbookRow {
  return rowFromApi(
    normalizeTimelineFullRow(
      {
        ...fullWorkbookViewRow(timelineContract, recordId, rowVersion, {
          "timeline.activity_synopsis_text": synopsis,
        }),
        observation: { recordId, rowVersion, scope: acceptedScope },
      },
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
    presence: emptyWorkbookPresence,
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

const outsideRecordId = "record-9";

function outsidePatch(
  rowVersion: number,
  cells: Record<string, { value: unknown }> = {
    "timeline.data_source_text": { value: `Remote source ${rowVersion}` },
  },
  options: {
    readonly changeKind?: "patch" | "invalidate";
    readonly groupValues?: Record<string, unknown>;
  } = {},
): RecordChangedPayload {
  return {
    record_id: outsideRecordId,
    row_version: rowVersion,
    change_set_id: `change-outside-${rowVersion}`,
    client_txn_id: `remote-outside-${rowVersion}`,
    actor_user_id: "user-2",
    changed_field_keys: Object.keys(cells).sort(),
    affected_views: [
      options.changeKind === "invalidate"
        ? { view_schema_id: timelineViewSchemaId, change_kind: "invalidate" }
        : {
            view_schema_id: timelineViewSchemaId,
            change_kind: "patch",
            patch_cells: {
              record_id: outsideRecordId,
              row_version: rowVersion,
              cells,
              ...(options.groupValues === undefined
                ? {}
                : { group_values: options.groupValues }),
            },
          },
    ],
  };
}

/** One accepted Timeline window that materializes only record-1. */
function liveWindowFixture(
  retainsLivePlacementRecord: (recordId: string) => boolean = () => false,
) {
  const fixture = projectionFixture();
  const draft = createDraftRow(1);
  const rowsRef = {
    current: [committedRow("record-1", 1, "Visible"), draft],
  };
  const replaceRows = vi.fn();
  const known = new Map<string, number>();
  const acceptRecordVersion = vi.fn((recordId: string, rowVersion: number) => {
    known.set(recordId, Math.max(known.get(recordId) ?? 0, rowVersion));
  });
  const port: WorkbookViewQueryPort = {
    query: vi.fn<WorkbookViewQueryPort["query"]>(async (request) => ({
      kind: "accepted",
      value: {
        incidentId: "incident-1",
        viewSchemaId: timelineViewSchemaId,
        rows: [
          fullWorkbookViewRow(timelineContract, "record-1", 1, {
            "timeline.activity_synopsis_text": "Visible",
          }),
        ],
        ...acceptedQueryMetadata(timelineViewSchemaId, request.queryState),
      },
    })),
  };
  const rendered = renderHook(() => ({
    browsing: useWorkbookQueryBrowser(port, timelineViewSchemaId),
    binding: useTimelineCollaborationBindings({
      clearCommittedRows: () => {},
      activeSheetRef: { kind: "view_schema", id: timelineViewSchemaId },
      admission: {
        acceptCommittedRow: (row) => ({ row }),
        acceptRecordVersion,
        isStaleRecordVersion: (recordId, rowVersion) =>
          rowVersion < (known.get(recordId) ?? 0),
      },
      beginRowsLoad: () => undefined,
      collaborationProjection: fixture.projection,
      readScope: () => acceptedScope,
      refreshRows: async () => undefined,
      resolveClientTxn: () => false,
      retainsLivePlacementRecord,
      rowsRef,
      rowStoreCommands: { replaceRows, updateRows: vi.fn() },
    }),
  }));
  const read = async (
    queryState: WorkbookQueryState = emptyWorkbookQueryState(),
  ) => {
    await act(async () => {
      const browser = rendered.result.current.browsing.browser;
      if (!browser) throw new Error("Missing Timeline query browser");
      const result = await browser.query({
        contract: timelineContract,
        queryState,
        signal: new AbortController().signal,
      });
      if (result.kind !== "accepted")
        throw new Error("Expected an accepted Timeline window");
      browser.accept(result.value);
    });
  };
  const apply = (payload: RecordChangedPayload) => {
    const surface = fixture.activeSurface();
    if (!surface) throw new Error("Missing active Timeline surface");
    let result: ReturnType<WorkbookActiveSurfacePort["applyRecordChanged"]> = {
      kind: "stale",
    };
    act(() => {
      result = surface.applyRecordChanged(payload);
    });
    return result;
  };
  return {
    acceptRecordVersion,
    apply,
    browsing: () => rendered.result.current.browsing.browser,
    fixture,
    read,
    rendered,
    replaceRows,
    rowsRef,
  };
}

describe("useTimelineCollaborationBindings", () => {
  it("owns Timeline live admission, gap refresh, access invalidation, surface switches, and teardown", async () => {
    const fixture = projectionFixture();
    const draft = createDraftRow(1);
    const rowsRef = {
      current: [committedRow("record-1", 1, "Before"), draft],
    };
    const replaceRows = vi.fn();
    const beginRowsLoad = vi.fn();
    const refreshRows = vi.fn(async () => undefined);
    let currentScope: WorkbookReadScope | null = acceptedScope;
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
          clearCommittedRows: () => {},
          activeSheetRef: { kind: "saved_view", id: sheetId },
          admission: {
            acceptCommittedRow,
            acceptRecordVersion,
            isStaleRecordVersion: (_recordId, rowVersion) =>
              rowVersion < knownVersion,
          },
          beginRowsLoad,
          collaborationProjection: fixture.projection,
          readScope: () => currentScope,
          refreshRows,
          resolveClientTxn,
          rowsRef,
          rowStoreCommands: {
            replaceRows,
            updateRows: vi.fn(),
          },
        }),
      { initialProps: { sheetId: "saved-view-1" } },
    );

    expect(fixture.activeSurface()?.identity).toEqual({
      sheetRef: { kind: "saved_view", id: "saved-view-1" },
      viewSchemaId: timelineViewSchemaId,
    });
    expect(fixture.clientTxnResolver()?.("local-1")).toBe(true);

    const change: RecordChangedPayload = {
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
    };
    act(() => {
      expect(fixture.activeSurface()?.applyRecordChanged(change)).toEqual({
        kind: "applied",
      });
    });
    expect(rowsRef.current[0]?.values.activitySynopsisText).toBe("After");
    expect(rowsRef.current[1]).toBe(draft);
    expect(rowsRef.current[0]?.rawRow?.observation).toEqual({
      recordId: "record-1",
      rowVersion: 2,
      scope: acceptedScope,
    });
    const firstRawRow = () => {
      const raw = rowsRef.current[0]?.rawRow;
      if (!raw) throw new Error("Expected the first row to carry its raw row");
      return raw;
    };
    expect(
      workbookRowIsAdmissible(firstRawRow(), "record-1", acceptedScope),
    ).toBe(true);
    for (const scope of [
      null,
      { ...acceptedScope, epoch: 2 },
      { ...acceptedScope, actorId: "actor-2" },
      { ...acceptedScope, sessionIdentity: "session-2" },
      { ...acceptedScope, incidentId: "incident-2" },
    ]) {
      currentScope = scope;
      rowsRef.current = [committedRow("record-1", 1, "Before"), draft];
      act(() => {
        expect(fixture.activeSurface()?.applyRecordChanged(change)).toEqual({
          kind: "applied",
        });
      });
      expect(rowsRef.current[0]?.rawRow?.observation).toEqual({
        recordId: "record-1",
        rowVersion: 1,
        scope: acceptedScope,
      });
      expect(
        workbookRowIsAdmissible(firstRawRow(), "record-1", currentScope),
      ).toBe(false);
    }
    currentScope = acceptedScope;
    const unobserved = committedRow("record-1", 1, "Before");
    const unobservedRaw = unobserved.rawRow;
    if (!unobservedRaw)
      throw new Error("Expected the committed row to carry its raw row");
    const { observation: _observation, ...raw } = unobservedRaw;
    rowsRef.current = [{ ...unobserved, rawRow: raw }, draft];
    act(() => {
      expect(fixture.activeSurface()?.applyRecordChanged(change)).toEqual({
        kind: "applied",
      });
    });
    expect(rowsRef.current[0]?.rawRow?.observation).toBeUndefined();

    expect(
      fixture.activeSurface()?.applyRecordChanged({
        record_id: "record-1",
        row_version: 1,
        change_set_id: "change-stale",
        client_txn_id: "remote-stale",
        actor_user_id: "user-2",
        changed_field_keys: [],
        affected_views: [
          {
            view_schema_id: timelineViewSchemaId,
            change_kind: "invalidate",
          },
        ],
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
    expect(replaceRows).toHaveBeenLastCalledWith([draft]);

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
        clearCommittedRows: () => {},
        activeSheetRef: { kind: "view_schema", id: timelineViewSchemaId },
        admission: {
          acceptCommittedRow: (row) => ({ row }),
          acceptRecordVersion: () => undefined,
          isStaleRecordVersion: () => false,
        },
        beginRowsLoad: () => undefined,
        collaborationProjection: fixture.projection,
        readScope: () => acceptedScope,
        refreshRows: async () => undefined,
        resolveClientTxn: () => false,
        rowsRef,
        rowStoreCommands: {
          replaceRows: () => undefined,
          updateRows: () => undefined,
        },
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

    const sparse = (patchCells?: unknown): RecordChangedPayload =>
      ({
        record_id: "record-1",
        row_version: 2,
        change_set_id: "change-malformed",
        client_txn_id: "remote-malformed",
        actor_user_id: "user-2",
        changed_field_keys: ["timeline.activity_synopsis_text"],
        affected_views: [
          {
            view_schema_id: timelineViewSchemaId,
            change_kind: "patch",
            ...(patchCells === undefined ? {} : { patch_cells: patchCells }),
          },
        ],
      }) as unknown as RecordChangedPayload;
    for (const payload of [
      sparse({
        record_id: "record-1",
        row_version: "two",
        cells: { "timeline.activity_synopsis_text": { value: "Malformed" } },
      }),
      sparse(),
    ]) {
      expect(fixture.activeSurface()?.applyRecordChanged(payload)).toEqual({
        kind: "refresh_required",
      });
      expect(rowsRef.current[0]?.values.activitySynopsisText).toBe("Before");
    }
    unmount();
  });
  it("applies a placement-neutral patch for a reconciled out-of-window record without refresh or row replacement", async () => {
    const live = liveWindowFixture();
    await live.read();
    // A first sighting may be a remote create, so it still reconciles the window.
    expect(live.apply(outsidePatch(2))).toEqual({ kind: "refresh_required" });
    await live.read();
    live.replaceRows.mockClear();
    const rows = live.rowsRef.current;

    expect(live.apply(outsidePatch(3))).toEqual({ kind: "applied" });
    expect(live.apply(outsidePatch(4))).toEqual({ kind: "applied" });
    expect(live.replaceRows).not.toHaveBeenCalled();
    expect(live.rowsRef.current).toBe(rows);
    expect(live.acceptRecordVersion).toHaveBeenCalledWith(outsideRecordId, 4);
    expect(live.apply(outsidePatch(3))).toEqual({ kind: "stale" });
    live.rendered.unmount();
  });

  it("keeps refresh for out-of-window placement changes, retained records, invalidation, and lost placement evidence", async () => {
    const neutral = liveWindowFixture();
    await neutral.read();
    neutral.apply(outsidePatch(2));
    await neutral.read();
    // A read whose query sorts by the changed field can move the record.
    await neutral.read({
      ...emptyWorkbookQueryState(),
      sort: [{ fieldKey: "timeline.data_source_text", direction: "asc" }],
    });
    expect(neutral.apply(outsidePatch(3))).toEqual({
      kind: "refresh_required",
    });
    await neutral.read({
      ...emptyWorkbookQueryState(),
      filters: [
        {
          fieldKey: "timeline.capture_state",
          op: "eq",
          arg: { value: "rough" },
        },
      ],
    });
    expect(
      neutral.apply(
        outsidePatch(4, { "timeline.capture_state": { value: "rough" } }),
      ),
    ).toEqual({ kind: "refresh_required" });
    await neutral.read({
      ...emptyWorkbookQueryState(),
      filters: [
        {
          fieldKey: "timeline.full_text",
          op: "full_text",
          arg: { query: "remote" },
        },
      ],
    });
    expect(neutral.apply(outsidePatch(5))).toEqual({
      kind: "refresh_required",
    });
    await neutral.read({
      ...emptyWorkbookQueryState(),
      groupBy: "timeline.capture_state",
    });
    expect(
      neutral.apply(
        outsidePatch(
          6,
          { "timeline.capture_state": { value: "rough" } },
          { groupValues: { "timeline.capture_state": "rough" } },
        ),
      ),
    ).toEqual({ kind: "refresh_required" });
    await neutral.read();
    expect(
      neutral.apply(outsidePatch(7, {}, { changeKind: "invalidate" })),
    ).toEqual({ kind: "refresh_required" });
    // Access loss forgets placement evidence with the protected rows.
    neutral.fixture
      .activeSurface()
      ?.invalidate({ kind: "incident_access_lost" });
    await neutral.read();
    expect(neutral.apply(outsidePatch(8))).toEqual({
      kind: "refresh_required",
    });
    expect(neutral.replaceRows).not.toHaveBeenCalledWith(
      expect.arrayContaining([
        expect.objectContaining({ recordId: outsideRecordId }),
      ]),
    );
    neutral.rendered.unmount();

    const retained = liveWindowFixture(
      (recordId) => recordId === outsideRecordId,
    );
    await retained.read();
    retained.apply(outsidePatch(2));
    await retained.read();
    expect(retained.apply(outsidePatch(3))).toEqual({
      kind: "refresh_required",
    });
    retained.rendered.unmount();
  });
});
