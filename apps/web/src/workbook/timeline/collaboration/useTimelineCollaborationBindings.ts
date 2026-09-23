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
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { useWorkbookBrowsingRegistry } from "../../query/WorkbookQueryBrowsingContext";
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
  refreshRows,
  resolveClientTxn,
  rowsRef,
  rowStoreCommands,
}: {
  readonly activeSheetRef: SheetRef;
  readonly admission: TimelineRowAdmission;
  readonly clearCommittedRows: () => void;
  readonly beginRowsLoad: () => unknown;
  readonly collaborationProjection: TimelineCollaborationProjection;
  readonly refreshRows: (options?: {
    readonly requireAcceptance?: boolean;
  }) => Promise<void>;
  readonly resolveClientTxn: (
    clientTxnId: string | null | undefined,
  ) => boolean;
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
      const nextRows = rowsRef.current.map((row) => {
        if (row.recordId !== patch.record_id || row.rawRow === null) {
          return row;
        }
        patched = true;
        const committed = rowFromApi(applyViewRowPatch(row.rawRow, patch));
        const accepted = admission.acceptCommittedRow(committed);
        return {
          ...accepted.row,
          collectionDrafts: row.collectionDrafts,
          pendingSignature: row.pendingSignature,
        };
      });
      if (!patched) {
        return { kind: "refresh_required" };
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
    [admission, browsing, replaceRows, rowsRef],
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
