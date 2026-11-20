import { useCallback, useEffect, useMemo } from "react";
import type { SheetRef } from "../../../shared/sheetRef";
import { useWorkbookCollaborationCoordinator } from "../../collaboration/useWorkbookCollaborationCoordinator";
import type {
  WorkbookCollaborationCoordinator,
  WorkbookCollaborationSnapshot,
} from "../../collaboration/WorkbookCollaborationCoordinator";
import type { WorkbookPresenceDraft } from "../../collaboration/workbookCollaborationMessages";
import type { WorkbookActiveSurfacePort } from "../../collaboration/workbookSurfacePort";
import type { WorkbookQueryInvalidationReason } from "../../lifecycle/workbookInvalidation";
import type { WorkbookQueryState } from "../../models/workbookQuery";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { useWorkbookBrowsingRegistry } from "../../query/WorkbookQueryBrowsingContext";
import type {
  WorkbookCanonicalQuery,
  WorkbookViewQueryPort,
} from "../../query/WorkbookViewQueryPort";
import { acceptTimelineLivePatchObservation } from "../adapters/acceptTimelineLivePatchObservation";
import type { TimelineRowStoreCommands } from "../models/timelineControllerPorts";
import {
  applyViewRowPatch,
  normalizeTimelinePatchCells,
  rowFromApi,
  type TimelinePatchCells,
  type WorkbookRow,
} from "../models/timelineRowModel";

export type TimelineCollaborationProjection = Pick<
  WorkbookCollaborationCoordinator,
  | "getSnapshot"
  | "publishPresence"
  | "registerActiveSurface"
  | "registerClientTxnResolver"
  | "requestAuthorizationRecovery"
  | "subscribe"
>;

type TimelineRowAdmission = {
  readonly acceptCommittedRow: (row: WorkbookRow) => {
    readonly row: WorkbookRow;
  };
  readonly acceptRecordVersion: (
    recordId: string,
    rowVersion: number,
  ) => unknown;
  readonly isStaleRecordVersion: (
    recordId: string,
    rowVersion: number,
  ) => boolean;
};

/**
 * Whether a live patch can change the record's placement under the accepted
 * or requested query. Unknown placement fails closed to a refresh.
 */
function livePatchCanMovePlacement(
  query: {
    readonly canonicalQuery: WorkbookCanonicalQuery | null;
    readonly requested: WorkbookQueryState | null;
  },
  changedFieldKeys: readonly string[],
  patch: TimelinePatchCells,
): boolean {
  if (query.canonicalQuery === null) return true;
  const keys = new Set([
    ...changedFieldKeys,
    ...Object.keys(patch.cells),
    ...Object.keys(patch.group_values ?? {}),
  ]);
  return [query.canonicalQuery, query.requested].some(
    (candidate) =>
      candidate !== null &&
      (candidate.filters.some(
        (filter) => filter.op === "full_text" || keys.has(filter.fieldKey),
      ) ||
        candidate.sort.some((entry) => keys.has(entry.fieldKey)) ||
        (typeof candidate.groupBy === "string" && keys.has(candidate.groupBy))),
  );
}

type TimelineCollaborationBinding = {
  readonly commands: {
    readonly publishPresence: (presence: WorkbookPresenceDraft) => void;
    readonly requestAuthorizationRecovery: () => void;
  };
  readonly snapshot: WorkbookCollaborationSnapshot;
};

/**
 * Timeline-owned reconciliation binding for the shell collaboration
 * coordinator. The coordinator owns stream and authorization policy; this
 * binding owns active-surface lifetime and Timeline row interpretation.
 */
export function useTimelineCollaborationBindings({
  activeSheetRef,
  admission,
  beginRowsLoad,
  clearCommittedRows,
  collaborationProjection,
  readScope,
  refreshRows,
  resolveClientTxn,
  retainsLivePlacementRecord,
  rowsRef,
  rowStoreCommands,
}: {
  readonly activeSheetRef: SheetRef;
  readonly admission: TimelineRowAdmission;
  readonly clearCommittedRows: () => void;
  readonly beginRowsLoad: () => unknown;
  readonly collaborationProjection: TimelineCollaborationProjection;
  readonly readScope: WorkbookViewQueryPort["readScope"];
  readonly refreshRows: (options?: {
    readonly requireAcceptance?: boolean;
  }) => Promise<void>;
  readonly resolveClientTxn: (
    clientTxnId: string | null | undefined,
  ) => boolean;
  /** Records whose client-local work keeps every live change on the refresh path. */
  readonly retainsLivePlacementRecord?: (recordId: string) => boolean;
  readonly rowsRef: { current: WorkbookRow[] };
  readonly rowStoreCommands: TimelineRowStoreCommands;
}): TimelineCollaborationBinding {
  const browsing = useWorkbookBrowsingRegistry();
  const { replaceRows } = rowStoreCommands;
  const snapshot = useWorkbookCollaborationCoordinator(collaborationProjection);

  useEffect(
    () => collaborationProjection.registerClientTxnResolver(resolveClientTxn),
    [collaborationProjection, resolveClientTxn],
  );

  const applyRecordChanged = useCallback<
    WorkbookActiveSurfacePort["applyRecordChanged"]
  >(
    (payload) => {
      if (
        admission.isStaleRecordVersion(payload.record_id, payload.row_version)
      ) {
        return { kind: "stale" };
      }
      admission.acceptRecordVersion(payload.record_id, payload.row_version);
      const affectedView = payload.affected_views.find(
        (view) => view.view_schema_id === timelineViewSchemaId,
      );
      if (
        affectedView?.change_kind !== "patch" ||
        affectedView.patch_cells === undefined
      ) {
        return { kind: "refresh_required" };
      }

      let patch: TimelinePatchCells;
      try {
        patch = normalizeTimelinePatchCells(
          affectedView.patch_cells,
          "record_changed patch_cells",
        );
      } catch {
        return { kind: "refresh_required" };
      }
      if (patch.record_id !== payload.record_id) {
        return { kind: "refresh_required" };
      }
      if (admission.isStaleRecordVersion(patch.record_id, patch.row_version)) {
        return { kind: "stale" };
      }
      admission.acceptRecordVersion(patch.record_id, patch.row_version);

      let patched = false;
      const scope = readScope?.() ?? null;
      const nextRows = rowsRef.current.map((row) => {
        if (row.recordId !== patch.record_id || row.rawRow === null) {
          return row;
        }
        patched = true;
        const committed = rowFromApi(
          acceptTimelineLivePatchObservation(
            row.rawRow,
            applyViewRowPatch(row.rawRow, patch),
            patch.record_id,
            scope,
          ),
        );
        const accepted = admission.acceptCommittedRow(committed);
        return {
          ...accepted.row,
          collectionDrafts: row.collectionDrafts,
          pendingSignature: row.pendingSignature,
        };
      });
      if (!patched) {
        // REQ-03-097 has no cells to patch outside the materialized window. A
        // known record whose placement cannot change stays outside it; anything
        // else reconciles the window through the HTTP view query.
        const browser = browsing.find(timelineViewSchemaId);
        return browser?.observeLivePlacement(patch.record_id) === true &&
          !livePatchCanMovePlacement(
            browser.getSnapshot(),
            payload.changed_field_keys,
            patch,
          ) &&
          retainsLivePlacementRecord?.(patch.record_id) !== true
          ? { kind: "applied" }
          : { kind: "refresh_required" };
      }
      rowsRef.current = nextRows;
      replaceRows(nextRows);
      const browser = browsing.find(timelineViewSchemaId);
      browser?.observeRows(
        nextRows.flatMap((row) => (row.rawRow ? [row.rawRow] : [])),
      );
      const placement = browser?.getSnapshot().canonicalQuery;
      return payload.changed_field_keys.some(
        (key) =>
          placement?.sort.some((sort) => sort.fieldKey === key) ||
          placement?.filters.some(
            (filter) => filter.fieldKey === key || filter.op === "full_text",
          ) ||
          placement?.groupBy === key,
      )
        ? { kind: "refresh_required" }
        : { kind: "applied" };
    },
    [
      admission,
      browsing,
      readScope,
      replaceRows,
      retainsLivePlacementRecord,
      rowsRef,
    ],
  );

  const activeSurface = useMemo<WorkbookActiveSurfacePort>(
    () => ({
      identity: {
        sheetRef: activeSheetRef,
        viewSchemaId: timelineViewSchemaId,
      },
      applyRecordChanged,
      invalidate: (reason: WorkbookQueryInvalidationReason) => {
        beginRowsLoad();
        if (
          reason.kind === "collaboration_reset_required" ||
          reason.kind === "incident_closed"
        ) {
          return;
        }
        clearCommittedRows();
        browsing.find(timelineViewSchemaId)?.invalidate();
        const localDrafts = rowsRef.current.filter(
          (row) => row.recordId === null,
        );
        rowsRef.current = localDrafts;
        replaceRows(localDrafts);
      },
      refresh: async (options) => {
        const read = () =>
          refreshRows({
            requireAcceptance: options?.reason === "authorization_recovered",
          });
        const browser = browsing.find(timelineViewSchemaId);
        if (options?.reason === "record_changed" && browser)
          await browser.reconcile(read);
        else await read();
      },
    }),
    [
      activeSheetRef,
      applyRecordChanged,
      beginRowsLoad,
      clearCommittedRows,
      refreshRows,
      rowsRef,
      replaceRows,
      browsing,
    ],
  );

  useEffect(
    () => collaborationProjection.registerActiveSurface(activeSurface),
    [activeSurface, collaborationProjection],
  );

  const publishPresence = useCallback(
    (presence: WorkbookPresenceDraft) => {
      collaborationProjection.publishPresence(presence);
    },
    [collaborationProjection],
  );
  const requestAuthorizationRecovery = useCallback(() => {
    collaborationProjection.requestAuthorizationRecovery();
  }, [collaborationProjection]);

  return {
    commands: {
      publishPresence,
      requestAuthorizationRecovery,
    },
    snapshot,
  };
}
