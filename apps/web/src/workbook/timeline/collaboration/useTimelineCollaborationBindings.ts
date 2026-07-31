import type { Dispatch, SetStateAction } from "react";
import { useCallback, useEffect, useMemo } from "react";
import type { WorkbookSheetRef } from "../../../shared/workbookSheetRef";
import { useWorkbookCollaborationCoordinator } from "../../collaboration/useWorkbookCollaborationCoordinator";
import type {
  WorkbookCollaborationCoordinator,
  WorkbookCollaborationSnapshot,
} from "../../collaboration/WorkbookCollaborationCoordinator";
import type { WorkbookPresenceDraft } from "../../collaboration/workbookCollaborationMessages";
import type { WorkbookActiveSurfacePort } from "../../collaboration/workbookSurfacePort";
import type { WorkbookQueryInvalidationReason } from "../../lifecycle/workbookInvalidation";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  applyViewRowPatch,
  normalizeTimelinePatchCells,
  rowFromApi,
  type TimelinePatchCells,
  type WorkbookRow,
} from "../models/workbookTimelineModel";

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

export type TimelineCollaborationBinding = {
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
  collaborationProjection,
  refreshRows,
  resolveClientTxn,
  rowsRef,
  setRows,
}: {
  readonly activeSheetRef: WorkbookSheetRef;
  readonly admission: TimelineRowAdmission;
  readonly beginRowsLoad: () => unknown;
  readonly collaborationProjection: TimelineCollaborationProjection;
  readonly refreshRows: () => Promise<void>;
  readonly resolveClientTxn: (
    clientTxnId: string | null | undefined,
  ) => boolean;
  readonly rowsRef: { current: WorkbookRow[] };
  readonly setRows: Dispatch<SetStateAction<WorkbookRow[]>>;
}): TimelineCollaborationBinding {
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
      setRows(nextRows);
      return { kind: "applied" };
    },
    [admission, rowsRef, setRows],
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
        const localDrafts = rowsRef.current.filter(
          (row) => row.recordId === null,
        );
        rowsRef.current = localDrafts;
        setRows(localDrafts);
      },
      refresh: async () => {
        await refreshRows();
      },
    }),
    [
      activeSheetRef,
      applyRecordChanged,
      beginRowsLoad,
      refreshRows,
      rowsRef,
      setRows,
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
